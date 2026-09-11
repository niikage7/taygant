package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"golang.org/x/image/font/gofont/goregular"
	"gorm.io/gorm"

	"taygant_backend/internal/models"
)

// Export — выгрузка сводки проекта во внешние форматы (реестр задач, Гант, сводка).
type Export struct {
	db         *gorm.DB
	tasks      *Tasks
	milestones *Milestones
}

// NewExport собирает сервис экспорта поверх уже готовых сервисов задач и вех:
// нужны их вычисляемые поля (эффективный статус, критический путь, буфер,
// статус вехи), а не сырые строки таблиц.
func NewExport(db *gorm.DB, tasks *Tasks, milestones *Milestones) *Export {
	return &Export{db: db, tasks: tasks, milestones: milestones}
}

// ExportFormat — формат файла выгрузки, см. GET /projects/{id}/export.
type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatXLSX ExportFormat = "xlsx"
)

// Valid сообщает, поддерживает ли сервис такой формат.
func (f ExportFormat) Valid() bool {
	return f == ExportFormatPDF || f == ExportFormatXLSX
}

var exportContentType = map[ExportFormat]string{
	ExportFormatPDF:  "application/pdf",
	ExportFormatXLSX: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
}

// File — готовый файл выгрузки.
type File struct {
	Content     []byte
	ContentType string
	FileName    string
}

// Generate строит файл выгрузки проекта: сводку, реестр задач (с критическим
// путём и буфером, как в /gantt) и вехи.
func (s *Export) Generate(ctx context.Context, projectID uuid.UUID, format ExportFormat) (File, error) {
	if !format.Valid() {
		return File{}, Invalid("недопустимый формат экспорта %q", format)
	}

	var project models.Project
	if err := s.db.WithContext(ctx).First(&project, "id = ?", projectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return File{}, ErrNotFound
		}
		return File{}, fmt.Errorf("найти проект: %w", err)
	}

	metrics, err := ComputeMetrics(ctx, s.db, project)
	if err != nil {
		return File{}, err
	}
	tasks, err := s.tasks.List(ctx, projectID, project, Filter{})
	if err != nil {
		return File{}, err
	}
	milestones, err := s.milestones.List(ctx, projectID)
	if err != nil {
		return File{}, err
	}

	var content []byte
	switch format {
	case ExportFormatXLSX:
		content, err = buildExportXLSX(project, metrics, tasks, milestones)
	case ExportFormatPDF:
		content, err = buildExportPDF(project, metrics, tasks, milestones)
	}
	if err != nil {
		return File{}, fmt.Errorf("сформировать файл выгрузки: %w", err)
	}

	return File{
		Content:     content,
		ContentType: exportContentType[format],
		FileName:    fmt.Sprintf("%s.%s", sanitizeFileName(project.Code), format),
	}, nil
}

// sanitizeFileName оставляет от кода проекта только безопасные для имени
// файла символы: Content-Disposition не экранирует произвольные байты.
func sanitizeFileName(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "project"
	}
	return b.String()
}

func boolLabel(b bool) string {
	if b {
		return "да"
	}
	return "нет"
}

// formatDateRu форматирует дату в принятом в интерфейсе виде ДД.ММ.ГГГГ.
// Не путать с models.Date.String() — тот отдаёт формат спецификации
// (YYYY-MM-DD) для сериализации в API и трогать его нельзя.
func formatDateRu(d models.Date) string {
	if d.IsZero() {
		return ""
	}
	return d.Format("02.01.2006")
}

// taskStatusLabelsRu — русские подписи статусов задач для экспорта, в точности
// как в интерфейсе (см. frontend/src/lib/task-status.ts, TASK_STATUS_META).
var taskStatusLabelsRu = map[models.TaskStatus]string{
	models.TaskStatusPlanned:    "План",
	models.TaskStatusInProgress: "В работе",
	models.TaskStatusDone:       "Готово",
	models.TaskStatusOverdue:    "Просрочена",
	models.TaskStatusBlocked:    "Заблокирована",
}

func taskStatusLabelRu(s models.TaskStatus) string {
	if label, ok := taskStatusLabelsRu[s]; ok {
		return label
	}
	return string(s)
}

// milestoneStatusLabelsRu — русские подписи статусов вех для экспорта, как в
// интерфейсе (см. frontend/src/lib/task-status.ts, MILESTONE_STATUS_META).
var milestoneStatusLabelsRu = map[models.MilestoneStatus]string{
	models.MilestoneStatusPlanned: "План",
	models.MilestoneStatusCurrent: "Текущая",
	models.MilestoneStatusDone:    "Сдано",
	models.MilestoneStatusFinal:   "Финал",
}

func milestoneStatusLabelRu(s models.MilestoneStatus) string {
	if label, ok := milestoneStatusLabelsRu[s]; ok {
		return label
	}
	return string(s)
}

// projectStatusLabelsRu — русские подписи статусов проекта для шапки экспорта.
var projectStatusLabelsRu = map[models.ProjectStatus]string{
	models.ProjectStatusDraft:     "Черновик",
	models.ProjectStatusActive:    "Активен",
	models.ProjectStatusOnHold:    "Приостановлен",
	models.ProjectStatusCompleted: "Завершён",
	models.ProjectStatusArchived:  "Архивирован",
}

func projectStatusLabelRu(s models.ProjectStatus) string {
	if label, ok := projectStatusLabelsRu[s]; ok {
		return label
	}
	return string(s)
}

// formatProgressRu округляет процент прогресса до целого — как везде в
// интерфейсе ("45%", без десятичных).
func formatProgressRu(percent float64) string {
	return fmt.Sprintf("%d", int(math.Round(percent)))
}

// ---- XLSX ----

func buildExportXLSX(project models.Project, metrics Metrics, tasks []models.Task, milestones []models.Milestone) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const summarySheet = "Проект"
	if err := f.SetSheetName("Sheet1", summarySheet); err != nil {
		return nil, err
	}
	writeSummarySheet(f, summarySheet, project, metrics)

	const tasksSheet = "Задачи"
	if _, err := f.NewSheet(tasksSheet); err != nil {
		return nil, err
	}
	writeTasksSheet(f, tasksSheet, tasks)

	const milestonesSheet = "Вехи"
	if _, err := f.NewSheet(milestonesSheet); err != nil {
		return nil, err
	}
	writeMilestonesSheet(f, milestonesSheet, milestones)

	f.SetActiveSheet(0)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("собрать xlsx: %w", err)
	}
	return buf.Bytes(), nil
}

func writeSummarySheet(f *excelize.File, sheet string, project models.Project, metrics Metrics) {
	rows := [][2]string{
		{"Код проекта", project.Code},
		{"Название", project.Name},
		{"Заказчик", project.CustomerOrg},
		{"Статус", projectStatusLabelRu(project.Status)},
		{"Фаза", project.Phase},
		{"Дата начала", formatDateRu(project.StartDate)},
		{"Дедлайн", formatDateRu(project.Deadline)},
		{"Прогресс, %", formatProgressRu(metrics.ProgressPercent)},
		{"Индекс здоровья", fmt.Sprintf("%d", metrics.HealthIndex)},
		{"Критических рисков", fmt.Sprintf("%d", metrics.CriticalRisksCount)},
		{"Рабочий календарь", string(project.WorkingCalendarType)},
	}
	for i, row := range rows {
		r := i + 1
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), row[0])
		f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row[1])
	}
	f.SetColWidth(sheet, "A", "A", 22)
	f.SetColWidth(sheet, "B", "B", 40)
}

var taskSheetHeader = []string{
	"Код", "WBS", "Название", "Ответственный", "Статус", "Начало", "Окончание",
	"Дней (раб.)", "% выполнения", "Критический путь", "Буфер, дн.",
}

func writeTasksSheet(f *excelize.File, sheet string, tasks []models.Task) {
	writeSheetRow(f, sheet, 1, toAnySlice(taskSheetHeader)...)
	for i, t := range tasks {
		assignee := ""
		if t.Assignee != nil {
			assignee = t.Assignee.FullName
		}
		writeSheetRow(f, sheet, i+2,
			t.Code, t.WBSNumber, t.Title, assignee, taskStatusLabelRu(t.Status),
			formatDateRu(t.StartDate), formatDateRu(t.EndDate), t.DurationWorkingDays,
			t.ProgressPercent, boolLabel(t.IsCriticalPath), t.BufferDays,
		)
	}
	setColWidths(f, sheet, 12, 10, 42, 24, 14, 12, 12, 12, 12, 16, 10)
}

var milestoneSheetHeader = []string{"Код", "Название", "Плановая дата", "Фактическая дата", "Статус", "Риск, дн."}

func writeMilestonesSheet(f *excelize.File, sheet string, milestones []models.Milestone) {
	writeSheetRow(f, sheet, 1, toAnySlice(milestoneSheetHeader)...)
	for i, m := range milestones {
		actual := ""
		if m.ActualDate != nil {
			actual = formatDateRu(*m.ActualDate)
		}
		risk := ""
		if m.RiskDays != nil {
			risk = fmt.Sprintf("%d", *m.RiskDays)
		}
		writeSheetRow(f, sheet, i+2, m.Code, m.Name, formatDateRu(m.PlannedDate), actual, milestoneStatusLabelRu(m.Status), risk)
	}
	setColWidths(f, sheet, 10, 40, 14, 16, 12, 10)
}

func writeSheetRow(f *excelize.File, sheet string, row int, values ...any) {
	for i, v := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			continue
		}
		f.SetCellValue(sheet, cell, v)
	}
}

func setColWidths(f *excelize.File, sheet string, widths ...float64) {
	for i, w := range widths {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			continue
		}
		f.SetColWidth(sheet, col, col, w)
	}
}

func toAnySlice(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// ---- PDF ----

// pdfFontFamily — имя, под которым в документ вшивается шрифт Go (goregular):
// в отличие от встроенных Helvetica/Arial, он покрывает кириллицу, без которой
// весь текст отчёта (названия задач, статусы) превратился бы в кракозябры.
const pdfFontFamily = "Go"

func buildExportPDF(project models.Project, metrics Metrics, tasks []models.Task, milestones []models.Milestone) ([]byte, error) {
	pdf := fpdf.New("L", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes(pdfFontFamily, "", goregular.TTF)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont(pdfFontFamily, "", 16)
	pdf.CellFormat(0, 10, fmt.Sprintf("%s (%s)", project.Name, project.Code), "", 1, "L", false, 0, "")

	pdf.SetFont(pdfFontFamily, "", 11)
	summary := []string{
		fmt.Sprintf("Статус: %s   Фаза: %s", projectStatusLabelRu(project.Status), project.Phase),
		fmt.Sprintf("Сроки: %s - %s   Рабочий календарь: %s", formatDateRu(project.StartDate), formatDateRu(project.Deadline), project.WorkingCalendarType),
		fmt.Sprintf("Прогресс: %s%%   Индекс здоровья: %d   Критических рисков: %d", formatProgressRu(metrics.ProgressPercent), metrics.HealthIndex, metrics.CriticalRisksCount),
	}
	for _, line := range summary {
		pdf.CellFormat(0, 7, line, "", 1, "L", false, 0, "")
	}
	pdf.Ln(4)

	pdf.SetFont(pdfFontFamily, "", 14)
	pdf.CellFormat(0, 8, fmt.Sprintf("Задачи (%d)", len(tasks)), "", 1, "L", false, 0, "")

	taskCols := []pdfColumn{
		{"Код", 18}, {"Название", 68}, {"Ответственный", 38}, {"Статус", 22},
		{"Начало", 20}, {"Окончание", 22}, {"Дн. (раб.)", 18}, {"%", 12},
		{"Крит. путь", 18}, {"Буфер, дн.", 18},
	}
	writePDFTable(pdf, taskCols, len(tasks), func(i int) []string {
		t := tasks[i]
		assignee := ""
		if t.Assignee != nil {
			assignee = t.Assignee.FullName
		}
		return []string{
			t.Code, t.Title, assignee, taskStatusLabelRu(t.Status),
			formatDateRu(t.StartDate), formatDateRu(t.EndDate),
			fmt.Sprintf("%d", t.DurationWorkingDays), fmt.Sprintf("%d", t.ProgressPercent),
			boolLabel(t.IsCriticalPath), fmt.Sprintf("%d", t.BufferDays),
		}
	})

	pdf.Ln(6)
	pdf.SetFont(pdfFontFamily, "", 14)
	pdf.CellFormat(0, 8, fmt.Sprintf("Вехи (%d)", len(milestones)), "", 1, "L", false, 0, "")

	milestoneCols := []pdfColumn{
		{"Код", 18}, {"Название", 90}, {"Плановая дата", 30}, {"Фактическая дата", 32}, {"Статус", 24}, {"Риск, дн.", 20},
	}
	writePDFTable(pdf, milestoneCols, len(milestones), func(i int) []string {
		m := milestones[i]
		actual := ""
		if m.ActualDate != nil {
			actual = formatDateRu(*m.ActualDate)
		}
		risk := ""
		if m.RiskDays != nil {
			risk = fmt.Sprintf("%d", *m.RiskDays)
		}
		return []string{m.Code, m.Name, formatDateRu(m.PlannedDate), actual, milestoneStatusLabelRu(m.Status), risk}
	})

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("собрать pdf: %w", err)
	}
	return buf.Bytes(), nil
}

type pdfColumn struct {
	Header string
	Width  float64
}

// writePDFTable рисует таблицу с фиксированной шириной колонок и повторяющимся
// заголовком на каждой странице (fpdf.SetAutoPageBreak сам не переносит
// заголовок, поэтому проверяем остаток места и подставляем его вручную).
// Длинные значения обрезаются многоточием — полноценный перенос текста внутри
// ячейки усложнил бы синхронизацию высоты строки между колонками без выигрыша
// для отчёта, который читают, а не редактируют.
func writePDFTable(pdf *fpdf.Fpdf, cols []pdfColumn, rowCount int, row func(i int) []string) {
	const rowHeight = 7.0
	drawHeader := func() {
		pdf.SetFont(pdfFontFamily, "", 9)
		pdf.SetFillColor(230, 230, 230)
		for _, c := range cols {
			pdf.CellFormat(c.Width, rowHeight, c.Header, "1", 0, "L", true, 0, "")
		}
		pdf.Ln(-1)
	}
	drawHeader()

	if rowCount == 0 {
		pdf.SetFont(pdfFontFamily, "", 9)
		pdf.CellFormat(0, rowHeight, "Нет данных", "1", 1, "L", false, 0, "")
		return
	}

	pdf.SetFont(pdfFontFamily, "", 9)
	_, pageHeight := pdf.GetPageSize()
	_, _, _, bottomMargin := pdf.GetMargins()
	pageBreakTrigger := pageHeight - bottomMargin

	for i := 0; i < rowCount; i++ {
		if pdf.GetY()+rowHeight > pageBreakTrigger {
			pdf.AddPage()
			drawHeader()
			pdf.SetFont(pdfFontFamily, "", 9)
		}
		for c, v := range row(i) {
			pdf.CellFormat(cols[c].Width, rowHeight, truncatePDFCell(v, cols[c].Width), "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}
}

// truncatePDFCell грубо оценивает вмещающееся число символов (~2мм на символ
// при 9pt) — не точная метрика шрифта, но rune-safe и достаточная для отчёта.
func truncatePDFCell(s string, widthMM float64) string {
	maxRunes := int(widthMM / 2.0)
	if maxRunes < 1 {
		maxRunes = 1
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	if maxRunes == 1 {
		return string(runes[:1])
	}
	return string(runes[:maxRunes-1]) + "…"
}
