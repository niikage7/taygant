/**
 * Статус проекта: draft — черновик (не запущен), active — выполняется,
 * on_hold — приостановлен, completed — завершён, archived — архивирован (удалён логически).
 */
export type ProjectStatus =
  | "draft"
  | "active"
  | "on_hold"
  | "completed"
  | "archived";

/**
 * Роль участника в проекте: project_manager — руководитель проекта,
 * analyst — бизнес-аналитик, developer — разработчик,
 * curator — куратор со стороны заказчика, other — прочая роль.
 */
export type ProjectRole =
  | "project_manager"
  | "analyst"
  | "developer"
  | "curator"
  | "other";

/**
 * Уровень доступа участника к проекту: full — полный доступ (управление проектом и командой),
 * edit — может редактировать задачи, view — доступ только для просмотра.
 */
export type AccessLevel = "full" | "edit" | "view";

/**
 * Статус задачи: planned — предстоит выполнить, in_progress — выполняется,
 * done — выполнена, overdue — просрочена, blocked — заблокирована зависимостью.
 */
export type TaskStatus =
  | "planned"
  | "in_progress"
  | "done"
  | "overdue"
  | "blocked";

/**
 * Тип связи между задачами по методу критического пути (CPM):
 * FS (Finish-to-Start) — предшественник должен завершиться до начала последователя;
 * SS (Start-to-Start) — задачи должны начаться одновременно;
 * FF (Finish-to-Finish) — задачи должны завершиться одновременно;
 * SF (Start-to-Finish) — последователь не может завершиться, пока не начался предшественник.
 */
export type DependencyType = "FS" | "SS" | "FF" | "SF";

/**
 * Тип рабочего календаря проекта: "5/2" — стандартная пятидневная неделя (40 ч/нед),
 * "6/1" — шестидневная рабочая неделя.
 */
export type WorkingCalendarType = "5/2" | "6/1";

/**
 * Статус контрольной точки: done — достигнута, current — ближайшая/текущая,
 * planned — запланирована, final — финальная веха проекта.
 */
export type MilestoneStatus = "done" | "current" | "planned" | "final";
