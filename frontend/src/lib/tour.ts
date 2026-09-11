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

export type TourRect = { top: number; left: number; width: number; height: number };

/**
 * Куда поставить карточку тура относительно подсвеченной цели (координаты
 * вьюпорта). Пробуем снизу, сверху, справа, слева (высокой цели — сначала
 * сбоку) — первое место, где карточка
 * целиком влезает в экран. Цель бывает высокой (навигация сайдбара тянется
 * почти на весь экран) или широкой (блок Ганта), и «снизу или сверху» не
 * хватает: карточка уезжала за верхний край. Если не влез ни один вариант,
 * прижимаем карточку к нижнему краю экрана поверх цели — лучше перекрыть часть
 * цели, чем спрятать кнопки тура.
 */
export function placeTourCard(
  target: TourRect,
  card: { width: number; height: number },
  viewport: { width: number; height: number },
  gap: number,
): { top: number; left: number } {
  const clamp = (value: number, min: number, max: number) =>
    Math.min(Math.max(value, min), Math.max(max, min));
  const horizontal = clamp(target.left, gap, viewport.width - card.width - gap);
  const vertical = clamp(target.top, gap, viewport.height - card.height - gap);

  const below = { top: target.top + target.height + gap, left: horizontal };
  const above = { top: target.top - gap - card.height, left: horizontal };
  const right = { top: vertical, left: target.left + target.width + gap };
  const left = { top: vertical, left: target.left - gap - card.width };
  // У высокой узкой цели (сайдбар) карточка сбоку читается как подпись к ней,
  // а снизу оказалась бы у края экрана, далеко от начала блока.
  const candidates =
    target.height > target.width ? [right, left, below, above] : [below, above, right, left];
  const fits = ({ top, left }: { top: number; left: number }) =>
    top >= gap &&
    left >= gap &&
    top + card.height <= viewport.height - gap &&
    left + card.width <= viewport.width - gap;

  return (
    candidates.find(fits) ?? {
      top: Math.max(gap, viewport.height - card.height - gap),
      left: horizontal,
    }
  );
}

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
