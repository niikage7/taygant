/**
 * Демонстрационные данные внутренних экранов.
 *
 * ⚠️ ВРЕМЕННЫЙ ИСТОЧНИК. В бэкенде из 15 групп эндпоинтов реализованы только
 * `/auth/*` и `/users/*` — остальные 40 хендлеров пока возвращают 501
 * (`notImplemented` в `backend/internal/api/*.go`), а seed наполняет лишь таблицу
 * пользователей. Подключать экраны к API до этого момента бессмысленно: они
 * показывали бы только ошибку.
 *
 * Содержимое повторяет макеты и типизировано боевыми интерфейсами из `@/types`,
 * поэтому переход на реальные данные — это замена вызовов в `src/data/queries.ts`
 * на соответствующие сервисы, без правок компонентов.
 */
import type {
  ChecklistItem,
  MemberWorkload,
  Milestone,
  Project,
  ProjectDashboard,
  ProjectMember,
  ShiftSimulation,
  Task,
  TaskDependency,
  TaskDetail,
  User,
} from "@/types";

const PROJECT_ID = "prj-2025-084";

export const demoUsers = {
  anna: {
    id: "u-anna",
    fullName: "Анна Соколова",
    email: "sokolova@taygant.dev",
    department: "Институт кибернетики",
    position: "Бизнес-аналитик",
    avatarUrl: null,
  },
  dmitry: {
    id: "u-dmitry",
    fullName: "Дмитрий Кузнецов",
    email: "kuznetsov@taygant.dev",
    department: "Отдел разработки",
    position: "Backend & DB",
    avatarUrl: null,
  },
  mikhail: {
    id: "u-mikhail",
    fullName: "Михаил Тарасов",
    email: "tarasov@taygant.dev",
    department: "Отдел разработки",
    position: "Lead Fullstack",
    avatarUrl: null,
  },
  elena: {
    id: "u-elena",
    fullName: "Елена Васильева",
    email: "vasileva@taygant.dev",
    department: "Отдел разработки",
    position: "Frontend Tech Lead (ТПУ Core)",
    avatarUrl: null,
  },
  olga: {
    id: "u-olga",
    fullName: "Ольга Романова",
    email: "romanova@taygant.dev",
    department: "Отдел качества",
    position: "QA & Docs",
    avatarUrl: null,
  },
  alex: {
    id: "u-alex",
    fullName: "Алекс Волков",
    email: "volkov@taygant.dev",
    department: "Институт кибернетики",
    position: "Руководитель проекта",
    avatarUrl: null,
  },
} satisfies Record<string, User>;

export const demoProject: Project = {
  id: PROJECT_ID,
  code: "TPU-PROJ-884",
  name: "Внедрение корпоративной системы управления проектами ТПУ",
  status: "active",
  phase: "Фаза 3: Внедрение & тестирование",
  startDate: "2025-10-01",
  deadline: "2025-12-05",
  progressPercent: 64,
  healthIndex: 64,
  criticalRisksCount: 1,
  description:
    "Единая система планирования и контроля проектов университета: диаграмма Ганта, расчёт критического пути и мониторинг рисков.",
  customerOrg: "Институт кибернетики ТПУ",
  durationCalendarDays: 66,
  workingCalendarType: "5/2",
  includePublicHolidays: true,
  autoRecalculateDependents: true,
  highlightCriticalPath: true,
  createdBy: demoUsers.alex,
  createdAt: "2025-09-20T09:00:00Z",
  updatedAt: "2025-11-05T14:30:00Z",
};

/** Задачи проекта в порядке WBS — общий источник для Ганта и реестра. */
export const demoTasks: Task[] = [
  {
    id: "t-1",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-001",
    wbsNumber: "1",
    title: "Анализ требований и ТЗ",
    assignee: demoUsers.anna,
    status: "done",
    isMilestone: false,
    startDate: "2025-10-01",
    endDate: "2025-10-12",
    durationCalendarDays: 12,
    durationWorkingDays: 9,
    progressPercent: 100,
    weightPercent: 12,
    isCriticalPath: false,
    bufferDays: 3,
    planVsActualDeviationDays: 0,
  },
  {
    id: "t-2",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-002",
    wbsNumber: "2",
    title: "UX/UI Проектирование",
    assignee: demoUsers.dmitry,
    status: "done",
    isMilestone: false,
    startDate: "2025-10-13",
    endDate: "2025-10-26",
    durationCalendarDays: 14,
    durationWorkingDays: 10,
    progressPercent: 100,
    weightPercent: 14,
    isCriticalPath: false,
    bufferDays: 2,
    planVsActualDeviationDays: 0,
  },
  {
    id: "t-3",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-003",
    wbsNumber: "3",
    title: "Архитектура БД и API",
    assignee: demoUsers.mikhail,
    status: "in_progress",
    isMilestone: false,
    startDate: "2025-10-27",
    endDate: "2025-11-10",
    durationCalendarDays: 15,
    durationWorkingDays: 11,
    progressPercent: 75,
    weightPercent: 18,
    isCriticalPath: false,
    bufferDays: 1,
    planVsActualDeviationDays: -3,
  },
  {
    id: "t-4",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-004",
    wbsNumber: "4",
    title: "Фронтенд: Модуль диаграммы Ганта",
    assignee: demoUsers.elena,
    status: "in_progress",
    isMilestone: false,
    startDate: "2025-11-02",
    endDate: "2025-11-18",
    durationCalendarDays: 17,
    durationWorkingDays: 12,
    progressPercent: 40,
    weightPercent: 14,
    isCriticalPath: true,
    bufferDays: 0,
    planVsActualDeviationDays: -1,
  },
  {
    id: "t-5",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-005",
    wbsNumber: "5",
    title: "Интеграция пересчёта сроков (CPM)",
    assignee: demoUsers.mikhail,
    status: "planned",
    isMilestone: false,
    startDate: "2025-11-15",
    endDate: "2025-11-25",
    durationCalendarDays: 11,
    durationWorkingDays: 8,
    progressPercent: 0,
    weightPercent: 16,
    isCriticalPath: true,
    bufferDays: 0,
    planVsActualDeviationDays: 0,
  },
  {
    id: "t-6",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-006",
    wbsNumber: "6",
    title: "Нагрузочное тестирование PostgreSQL",
    assignee: demoUsers.olga,
    status: "planned",
    isMilestone: false,
    startDate: "2025-11-24",
    endDate: "2025-12-04",
    durationCalendarDays: 11,
    durationWorkingDays: 8,
    progressPercent: 0,
    weightPercent: 12,
    isCriticalPath: false,
    bufferDays: 2,
    planVsActualDeviationDays: 0,
  },
  {
    id: "t-7",
    projectId: PROJECT_ID,
    sprintId: "s-4",
    parentTaskId: null,
    code: "TASK-007",
    wbsNumber: "7",
    title: "Релиз и защита в ТПУ",
    assignee: demoUsers.alex,
    status: "planned",
    isMilestone: true,
    startDate: "2025-12-05",
    endDate: "2025-12-05",
    durationCalendarDays: 1,
    durationWorkingDays: 1,
    progressPercent: 0,
    weightPercent: 14,
    isCriticalPath: true,
    bufferDays: 0,
    planVsActualDeviationDays: 0,
  },
];

/** Связи между задачами: цепочка FS плюс финиш-к-финишу перед релизом. */
export const demoDependencies: TaskDependency[] = [
  { id: "d-1", predecessorTaskId: "t-1", successorTaskId: "t-2", type: "FS", lagDays: 0, isCritical: false },
  { id: "d-2", predecessorTaskId: "t-2", successorTaskId: "t-3", type: "FS", lagDays: 0, isCritical: false },
  { id: "d-3", predecessorTaskId: "t-3", successorTaskId: "t-4", type: "FS", lagDays: 0, isCritical: true },
  { id: "d-4", predecessorTaskId: "t-4", successorTaskId: "t-5", type: "FS", lagDays: 0, isCritical: true },
  { id: "d-5", predecessorTaskId: "t-5", successorTaskId: "t-6", type: "FS", lagDays: 0, isCritical: false },
  { id: "d-6", predecessorTaskId: "t-6", successorTaskId: "t-7", type: "FS", lagDays: 0, isCritical: true },
];

export const demoMilestones: Milestone[] = [
  { id: "m-1", projectId: PROJECT_ID, code: "КТ-1", name: "Архитектура и ТЗ", plannedDate: "2025-10-25", actualDate: "2025-10-25", status: "done", riskDays: null },
  { id: "m-2", projectId: PROJECT_ID, code: "КТ-2", name: "Интеграция Ганта", plannedDate: "2025-11-07", actualDate: null, status: "current", riskDays: 2 },
  { id: "m-3", projectId: PROJECT_ID, code: "КТ-3", name: "Пилотное тестирование", plannedDate: "2025-11-11", actualDate: null, status: "planned", riskDays: null },
  { id: "m-4", projectId: PROJECT_ID, code: "КТ-4", name: "Защита перед жюри", plannedDate: "2025-12-05", actualDate: null, status: "final", riskDays: null },
];

export const demoWorkload: MemberWorkload[] = [
  { user: demoUsers.mikhail, sprintId: "s-4", assignedHoursPerWeek: 44, weeklyHoursLimit: 40, utilizationPercent: 110 },
  { user: demoUsers.elena, sprintId: "s-4", assignedHoursPerWeek: 34, weeklyHoursLimit: 40, utilizationPercent: 85 },
  { user: demoUsers.dmitry, sprintId: "s-4", assignedHoursPerWeek: 28, weeklyHoursLimit: 40, utilizationPercent: 70 },
  { user: demoUsers.olga, sprintId: "s-4", assignedHoursPerWeek: 24, weeklyHoursLimit: 40, utilizationPercent: 60 },
];

const taskById = (id: string): Task => {
  const task = demoTasks.find((item) => item.id === id);
  if (!task) throw new Error(`Демо-данные: задача ${id} не найдена`);
  return task;
};

export const demoDashboard: ProjectDashboard = {
  project: demoProject,
  overallProgressPercent: 64.3,
  progressDeltaPercent: 4.2,
  milestonesClosed: 18,
  milestonesTotal: 28,
  scheduleAdherence: { status: "at_risk", forecastDeviationDays: 2, projectBufferDays: 0.5 },
  taskStatusBreakdown: { total: 28, done: 18, inProgress: 4, planned: 5, overdue: 1 },
  teamVelocity: { tasksPerWeek: 4.8 },
  nearestMilestones: demoMilestones,
  risks: [
    {
      severity: "critical",
      title: "Угроза срыва дедлайна этапа «Фронтенд Ганта»",
      description:
        "Тимлид направления Михаил Т. уходит в утверждённый учебный отпуск через 2 дня. Риск сдвига критического пути проекта на +2 рабочих дня.",
      relatedTaskId: "t-4",
      impactDays: 2,
    },
  ],
  attentionTasks: [
    { task: taskById("t-3"), deviationDays: -3 },
    { task: taskById("t-4"), deviationDays: -1 },
    { task: taskById("t-6"), deviationDays: 0 },
    { task: taskById("t-5"), deviationDays: -0.5 },
  ],
  teamWorkload: demoWorkload,
};

const demoChecklist: ChecklistItem[] = [
  { id: "c-1", taskId: "t-4", text: "Виртуализация строк таблицы и таймлайна", isDone: true },
  { id: "c-2", taskId: "t-4", text: "Связывание задач перетаскиванием коннекторов", isDone: true },
  { id: "c-3", taskId: "t-4", text: "Поддержка типов связей: FS, SS, FF, SF", isDone: true },
  { id: "c-4", taskId: "t-4", text: "Симулятор автоматического каскадного сдвига дат", isDone: false },
  { id: "c-5", taskId: "t-4", text: "Оптимизация рендера SVG-кривых связей до 60fps", isDone: false },
];

export const demoTaskDetail: TaskDetail = {
  ...taskById("t-4"),
  title: "Разработка модуля интерактивной диаграммы Ганта",
  // В карточке задачи прогресс детальнее, чем в сводке реестра: 45% при плане 40%.
  progressPercent: 45,
  description:
    "Реализация Canvas/SVG движка визуализации диаграммы Ганта для кейс-чемпионата ТПУ. Включает отрисовку сетки дней/недель, виртуализацию рендеринга 500+ узлов, интерактивное перетаскивание дат и расчёт критического пути в реальном времени.",
  externalLink: {
    provider: "GitLab",
    url: "https://gitlab.example.org/tpu/taygant/-/issues/342",
    referenceId: "342",
    synced: true,
  },
  checklist: demoChecklist,
  predecessors: [
    {
      id: "d-3",
      predecessorTaskId: "t-3",
      successorTaskId: "t-4",
      predecessorTitle: "Архитектура БД и API",
      type: "FS",
      lagDays: 0,
      isCritical: true,
    },
  ],
  successors: [
    {
      id: "d-4",
      predecessorTaskId: "t-4",
      successorTaskId: "t-5",
      successorTitle: "Интеграция связей и пересчёта сроков",
      type: "FS",
      lagDays: 0,
      isCritical: true,
    },
    {
      id: "d-7",
      predecessorTaskId: "t-4",
      successorTaskId: "t-6",
      successorTitle: "Тестирование и валидация ТПУ",
      type: "FF",
      lagDays: 0,
      isCritical: false,
    },
  ],
  commentsCount: 3,
};

/** Результат what-if моделирования сдвига задачи #4 на +5 дней. */
export const demoSimulation: ShiftSimulation = {
  sourceTaskId: "t-4",
  shiftDays: 5,
  algorithm: "Forward-Pass CPM",
  affectedTasks: [
    {
      taskId: "t-4",
      title: "Модуль интерактивной диаграммы Ганта (Текущая)",
      isCriticalPath: true,
      originalStartDate: "2025-11-02",
      originalEndDate: "2025-11-18",
      newStartDate: "2025-11-07",
      newEndDate: "2025-11-23",
      dependencyType: "FS",
    },
    {
      taskId: "t-5",
      title: "Интеграция связей и пересчёта сроков",
      isCriticalPath: true,
      originalStartDate: "2025-11-19",
      originalEndDate: "2025-11-26",
      newStartDate: "2025-11-24",
      newEndDate: "2025-12-01",
      dependencyType: "FS",
    },
    {
      taskId: "t-6",
      title: "Тестирование и сдача в ТПУ Enterprise",
      isCriticalPath: false,
      originalStartDate: "2025-11-27",
      originalEndDate: "2025-12-05",
      newStartDate: "2025-12-02",
      newEndDate: "2025-12-10",
      dependencyType: "FS",
    },
  ],
  projectDeadlineImpact: {
    originalDeadline: "2025-12-05",
    newDeadline: "2025-12-10",
    deltaDays: 5,
  },
  bufferAvailableDays: 0.5,
  applied: false,
};

export const demoMembers: ProjectMember[] = [
  { id: "pm-1", user: demoUsers.alex, projectRole: "project_manager", sprintRole: "Руководитель проекта", accessLevel: "full", weeklyHoursLimit: 40 },
  { id: "pm-2", user: demoUsers.anna, projectRole: "analyst", sprintRole: "Аналитик", accessLevel: "edit", weeklyHoursLimit: 40 },
  { id: "pm-3", user: demoUsers.mikhail, projectRole: "developer", sprintRole: "Разработчик", accessLevel: "edit", weeklyHoursLimit: 40 },
  { id: "pm-4", user: demoUsers.olga, projectRole: "curator", sprintRole: "Куратор от ТПУ", accessLevel: "view", weeklyHoursLimit: 20 },
];
