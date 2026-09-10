import { httpClient } from "./http-client";
import type { Comment, CommentCreateRequest } from "@/types";

export const commentsService = {
  /** Возвращает ленту комментариев задачи в хронологическом порядке. */
  list(taskId: string): Promise<Comment[]> {
    return httpClient.get<Comment[]>(`/tasks/${taskId}/comments`);
  },

  /** Публикует новый комментарий от имени текущего пользователя. */
  add(taskId: string, payload: CommentCreateRequest): Promise<Comment> {
    return httpClient.post<Comment>(`/tasks/${taskId}/comments`, payload);
  },
};
