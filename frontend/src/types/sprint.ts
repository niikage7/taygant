/** Спринт (итерация) или релиз внутри проекта, используемый для группировки задач. */
export interface Sprint {
  /** Идентификатор спринта. */
  id: string;
  /** Идентификатор проекта, которому принадлежит спринт. */
  projectId: string;
  /** Название спринта. */
  name: string;
  /** Версия релиза, связанная со спринтом. */
  releaseVersion?: string;
  /** Дата начала спринта. */
  startDate: string;
  /** Дата окончания спринта. */
  endDate: string;
}

/** Данные для создания или изменения спринта. */
export interface SprintCreateRequest {
  /** Название спринта. */
  name: string;
  /** Версия релиза, связанная со спринтом. */
  releaseVersion?: string;
  /** Дата начала спринта. */
  startDate: string;
  /** Дата окончания спринта. */
  endDate: string;
}
