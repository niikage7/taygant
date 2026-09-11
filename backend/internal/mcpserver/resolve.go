package mcpserver

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"taygant_backend/internal/access"
	"taygant_backend/internal/models"
)

// Ассистент ссылается на проекты и задачи так же, как человек в разговоре:
// кодом, частью названия или id из прошлых ответов. Функции ниже превращают
// такую ссылку в конкретную запись — и только среди проектов, где человек
// состоит в команде: чужие проекты для ассистента не существуют.

// memberships возвращает проекты пользователя вместе с его записью участника,
// новые проекты первыми. Удалённые проекты Preload не подгружает.
func (s *Server) memberships(ctx context.Context, userID uuid.UUID) ([]access.Membership, error) {
	var members []models.ProjectMember
	err := s.db.WithContext(ctx).Preload("Project").Where("user_id = ?", userID).Find(&members).Error
	if err != nil {
		return nil, fmt.Errorf("выбрать проекты пользователя: %w", err)
	}

	out := make([]access.Membership, 0, len(members))
	for _, m := range members {
		if m.Project == nil {
			continue
		}
		project := *m.Project
		m.Project = nil
		out = append(out, access.Membership{Project: project, Member: m})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Project.CreatedAt.After(out[j].Project.CreatedAt)
	})
	return out, nil
}

// pickProject выбирает проект по ссылке: id, код («PRJ-2026-A1B2»), название
// или его часть. Без ссылки подходит единственный неархивный проект — чтобы
// человеку с одним проектом не приходилось его называть.
func pickProject(all []access.Membership, ref string) (access.Membership, error) {
	ref = strings.TrimSpace(ref)
	if len(all) == 0 {
		return access.Membership{}, failf("вы пока не состоите ни в одном проекте — попросите менеджера добавить вас в команду")
	}

	if ref == "" {
		active := make([]access.Membership, 0, len(all))
		for _, m := range all {
			if m.Project.Status != models.ProjectStatusArchived {
				active = append(active, m)
			}
		}
		switch len(active) {
		case 0:
			return access.Membership{}, failf("все ваши проекты в архиве — укажите нужный параметром project: %s", projectList(all))
		case 1:
			return active[0], nil
		default:
			return access.Membership{}, failf("у вас несколько проектов, уточните какой (параметр project): %s", projectList(active))
		}
	}

	if id, err := uuid.Parse(ref); err == nil {
		for _, m := range all {
			if m.Project.ID == id {
				return m, nil
			}
		}
		return access.Membership{}, failf("проект %s не найден среди ваших. Ваши проекты: %s", ref, projectList(all))
	}

	for _, m := range all {
		if strings.EqualFold(m.Project.Code, ref) || strings.EqualFold(m.Project.Name, ref) {
			return m, nil
		}
	}

	needle := strings.ToLower(ref)
	var matches []access.Membership
	for _, m := range all {
		if strings.Contains(strings.ToLower(m.Project.Name), needle) {
			matches = append(matches, m)
		}
	}
	switch len(matches) {
	case 0:
		return access.Membership{}, failf("проект «%s» не найден среди ваших. Ваши проекты: %s", ref, projectList(all))
	case 1:
		return matches[0], nil
	default:
		return access.Membership{}, failf("под «%s» подходят несколько проектов: %s — уточните", ref, projectList(matches))
	}
}

// projectList перечисляет проекты для сообщения ассистенту.
func projectList(ms []access.Membership) string {
	parts := make([]string, 0, len(ms))
	for _, m := range ms {
		parts = append(parts, fmt.Sprintf("«%s» (%s)", m.Project.Name, m.Project.Code))
	}
	return strings.Join(parts, ", ")
}

// taskCodePattern узнаёт код задачи в том виде, в каком его пишут люди:
// «TASK-005», «task-5», «task 5», «5».
var taskCodePattern = regexp.MustCompile(`(?i)^(?:task)?[\s_-]*0*(\d{1,6})$`)

// normalizeTaskCode приводит код к виду, в котором его выдаёт приложение
// (service.nextTaskCode): «TASK-005», «TASK-1234».
func normalizeTaskCode(ref string) (string, bool) {
	m := taskCodePattern.FindStringSubmatch(strings.TrimSpace(ref))
	if m == nil {
		return "", false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return "", false
	}
	return fmt.Sprintf("TASK-%03d", n), true
}

// maxTaskCandidates — сколько вариантов показать, если ссылка на задачу
// неоднозначна. Больше ассистенту не нужно: он всё равно переспросит.
const maxTaskCandidates = 5

// findTask находит задачу по ссылке: id, код или часть названия. projectRef
// сужает поиск до одного проекта — нужен, когда код встречается в нескольких.
// Возвращается задача в том виде, в каком хранится (статус — сохранённый, без
// вычисленных overdue/blocked), и запись участника для проверки прав.
func (s *Server) findTask(ctx context.Context, userID uuid.UUID, taskRef, projectRef string) (models.Task, access.Membership, error) {
	taskRef = strings.TrimSpace(taskRef)
	if taskRef == "" {
		return models.Task{}, access.Membership{}, failf("укажите задачу: код вроде TASK-005, id или часть названия")
	}

	all, err := s.memberships(ctx, userID)
	if err != nil {
		return models.Task{}, access.Membership{}, err
	}
	scope := all
	if strings.TrimSpace(projectRef) != "" {
		m, err := pickProject(all, projectRef)
		if err != nil {
			return models.Task{}, access.Membership{}, err
		}
		scope = []access.Membership{m}
	}
	if len(scope) == 0 {
		return models.Task{}, access.Membership{}, failf("вы пока не состоите ни в одном проекте — задач для вас нет")
	}

	byProject := make(map[uuid.UUID]access.Membership, len(scope))
	projectIDs := make([]uuid.UUID, 0, len(scope))
	for _, m := range scope {
		byProject[m.Project.ID] = m
		projectIDs = append(projectIDs, m.Project.ID)
	}

	find := func(condition string, value any) ([]models.Task, error) {
		var tasks []models.Task
		err := s.db.WithContext(ctx).
			Where("project_id IN ?", projectIDs).
			Where(condition, value).
			Order("code").
			Limit(maxTaskCandidates + 1).
			Find(&tasks).Error
		if err != nil {
			return nil, fmt.Errorf("найти задачу: %w", err)
		}
		return tasks, nil
	}

	var candidates []models.Task
	if id, parseErr := uuid.Parse(taskRef); parseErr == nil {
		candidates, err = find("id = ?", id)
	} else if code, ok := normalizeTaskCode(taskRef); ok {
		candidates, err = find("code = ?", code)
	} else {
		// Точное название важнее частичного: «Дизайн» не должен становиться
		// неоднозначным из-за соседней задачи «Дизайн лендинга».
		pattern := escapeLike(taskRef)
		candidates, err = find("title ILIKE ?", pattern)
		if err == nil && len(candidates) != 1 {
			candidates, err = find("title ILIKE ?", "%"+pattern+"%")
		}
	}
	if err != nil {
		return models.Task{}, access.Membership{}, err
	}

	switch len(candidates) {
	case 0:
		where := "в ваших проектах"
		if len(scope) == 1 {
			where = fmt.Sprintf("в проекте «%s»", scope[0].Project.Name)
		}
		return models.Task{}, access.Membership{}, failf("задача «%s» не найдена %s", taskRef, where)
	case 1:
		return candidates[0], byProject[candidates[0].ProjectID], nil
	}

	shown := candidates
	if len(shown) > maxTaskCandidates {
		shown = shown[:maxTaskCandidates]
	}
	parts := make([]string, 0, len(shown))
	for _, t := range shown {
		parts = append(parts, fmt.Sprintf("%s «%s» (проект «%s»)", t.Code, t.Title, byProject[t.ProjectID].Project.Name))
	}
	return models.Task{}, access.Membership{}, failf("под «%s» подходят несколько задач: %s — уточните код или проект", taskRef, strings.Join(parts, "; "))
}

// escapeLike обезвреживает спецсимволы шаблона LIKE во вводе: иначе «%» в
// названии превратил бы поиск в «любая задача».
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
