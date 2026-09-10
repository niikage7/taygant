import type { ProjectStatus, WorkingCalendarType } from "./common";
import type { ProjectMemberCreateRequest } from "./member";
import type { User } from "./user";

/** Краткая карточка проекта для списков и виджетов. */
export interface ProjectSummary {
  /** Уникальный идентификатор проекта. */
  id: string;
  /** Короткий читаемый код проекта. */
  code: string;
  /** Название проекта. */
  name: string;
  /** Текущий статус проекта. */
  status: ProjectStatus;
  /** Текущая фаза реализации проекта (произвольный текст). */
  phase?: string;
  /** Дата начала графика проекта. */
  startDate: string;
  /** Плановый дедлайн (срок сдачи) проекта. */
  deadline: string;
  /** Общий процент выполнения проекта (взвешенный по задачам). */
  progressPercent: number;
  /** Интегральный индекс здоровья проекта (0–100), отображаемый в виде индикатора мониторинга. */
  healthIndex: number;
  /** Количество открытых рисков с критической серьёзностью. */
  criticalRisksCount: number;
}

/** Полная карточка проекта с параметрами календаря и настройками расчёта графика. */
export interface Project extends ProjectSummary {
  /** Описание и бизнес-цель проекта. */
  description?: string;
  /** Заказчик или подразделение, для которого выполняется проект. */
  customerOrg?: string;
  /** Продолжительность проекта в календарных днях (от startDate до deadline). */
  durationCalendarDays: number;
  /** Рабочий календарь, используемый при расчёте сроков задач. */
  workingCalendarType: WorkingCalendarType;
  /** Учитывать ли государственные праздники РФ как нерабочие дни. */
  includePublicHolidays: boolean;
  /**
   * Если true — при сдвиге даты задачи все зависимые задачи по цепочке Finish-to-Start
   * автоматически пересчитываются с сохранением буфера.
   */
  autoRecalculateDependents: boolean;
  /** Если true — задачи с нулевым резервом времени подсвечиваются как критический путь на диаграмме Ганта. */
  highlightCriticalPath: boolean;
  /** Пользователь, создавший проект. */
  createdBy: User;
  /** Дата и время создания проекта. */
  createdAt: string;
  /** Дата и время последнего изменения проекта. */
  updatedAt: string;
}

/** Данные для создания проекта на первом шаге мастера инициации. */
export interface ProjectCreateRequest {
  /** Название проекта. */
  name: string;
  /** Описание и бизнес-цель проекта. */
  description?: string;
  /** Заказчик или подразделение, для которого выполняется проект. */
  customerOrg?: string;
  /** Дата старта графика проекта. */
  startDate: string;
  /** Плановый дедлайн проекта. */
  deadline: string;
  /** Рабочий календарь проекта. */
  workingCalendarType?: WorkingCalendarType;
  /** Учитывать ли государственные праздники РФ как нерабочие дни. */
  includePublicHolidays?: boolean;
  /** Включить автоматический пересчёт зависимых задач при сдвиге дат. */
  autoRecalculateDependents?: boolean;
  /** Подсвечивать критический путь на диаграмме Ганта. */
  highlightCriticalPath?: boolean;
  /** Начальный состав команды проекта, назначаемый на шаге 2 мастера. */
  members?: ProjectMemberCreateRequest[];
  /** Если true — сохранить проект как черновик без перехода к диаграмме Ганта. */
  saveAsDraft?: boolean;
}

/** Набор полей проекта, доступных для частичного обновления (PATCH). */
export interface ProjectUpdateRequest {
  /** Новое название проекта. */
  name?: string;
  /** Новое описание проекта. */
  description?: string;
  /** Новый статус проекта. */
  status?: ProjectStatus;
  /** Новая фаза реализации проекта. */
  phase?: string;
  /** Новая дата старта графика проекта. */
  startDate?: string;
  /** Новый плановый дедлайн проекта. */
  deadline?: string;
  /** Новый рабочий календарь проекта. */
  workingCalendarType?: WorkingCalendarType;
}

/** Параметры фильтрации списка проектов. */
export interface ProjectListParams {
  /** Фильтр по статусу проекта. */
  status?: ProjectStatus;
  /** Подстрока для поиска по названию или коду проекта. */
  search?: string;
}
