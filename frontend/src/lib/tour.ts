/** Шаг тура: цель ищется по атрибуту `data-tour`. */
export type TourStep = {
  id: string;
  /** Значение `data-tour` у подсвечиваемого элемента. */
  target: string;
  title: string;
  text: string;
  /** Маршрут, на котором живёт шаг. Тур перейдёт туда сам. */
  href: string;
};

export const TOUR_STEPS: TourStep[] = [
  {
    id: "nav",
    target: "nav",
    href: "/overview",
    title: "Разделы проекта",
    text: "Слева — четыре рабочих экрана: диаграмма Ганта, обзор с метриками, карточка задачи и мастер создания проекта.",
  },
  {
    id: "project",
    target: "project-switcher",
    href: "/overview",
    title: "Текущий проект",
    text: "Здесь видно здоровье проекта и критические риски. Если проектов несколько, карточка становится переключателем — выбор запоминается между переходами.",
  },
  {
    id: "metrics",
    target: "kpi",
    href: "/overview",
    title: "Показатели",
    text: "Прогресс, соблюдение сроков и разбивка задач по статусам. Соблюдение сроков считает сервер по методу критического пути.",
  },
  {
    id: "milestones",
    target: "milestones",
    href: "/overview",
    title: "Контрольные точки",
    text: "Обязательства проекта по датам. Здесь их создают, отмечают достигнутыми и привязывают к ним задачи — по этим задачам считается прогресс вехи.",
  },
  {
    id: "gantt",
    target: "gantt",
    href: "/gantt",
    title: "Реестр и таймлайн",
    text: "Слева задачи, справа их отрезки и связи. Задачи сгруппированы под своими вехами, группу можно свернуть. Клик по названию открывает карточку задачи.",
  },
  {
    id: "create",
    target: "create-task",
    href: "/gantt",
    title: "Создание задачи",
    text: "Задачу можно создать из шапки с любого экрана: указать сроки, ответственного и веху, к которой она ведёт.",
  },
];

const STORAGE_KEY = "taygant.tourCompleted";

/** Тур показывается один раз; отметка живёт в браузере пользователя. */
export function isTourCompleted(): boolean {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    // Приватный режим может запрещать доступ — тогда просто покажем тур снова.
    return false;
  }
}

export function markTourCompleted(): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, "1");
  } catch {
    // Не критично: тур появится в следующий раз, данные при этом не теряются.
  }
}
