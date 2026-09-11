import type { User } from "./user";

/** Комментарий к задаче, оставленный участником команды. */
export interface Comment {
  /** Идентификатор комментария. */
  id: string;
  /** Идентификатор задачи, к которой относится комментарий. */
  taskId: string;
  /** Автор комментария. */
  author: User;
  /** Текст комментария. */
  text: string;
  /** Дата и время публикации комментария. */
  createdAt: string;
}

/** Данные для добавления комментария. */
export interface CommentCreateRequest {
  /** Текст комментария. */
  text: string;
}
