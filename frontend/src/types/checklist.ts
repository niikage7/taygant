/** Пункт чек-листа критериев приёмки (Definition of Done) задачи. */
export interface ChecklistItem {
  /** Идентификатор пункта чек-листа. */
  id: string;
  /** Идентификатор задачи, которой принадлежит пункт. */
  taskId: string;
  /** Формулировка критерия приёмки. */
  text: string;
  /** Признак выполнения критерия. */
  isDone: boolean;
}

/** Данные для добавления критерия приёмки. */
export interface ChecklistItemCreateRequest {
  /** Формулировка критерия приёмки. */
  text: string;
}

/** Данные для изменения критерия приёмки. */
export interface ChecklistItemUpdateRequest {
  /** Новая формулировка критерия приёмки. */
  text?: string;
  /** Признак выполнения критерия. */
  isDone?: boolean;
}
