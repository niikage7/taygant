"use client";

import { PageHeader } from "@/components/app/page-header";
import {
  EmptyProjectTasks,
  EmptyProjects,
  PageError,
  PageLoading,
} from "@/components/app/page-state";
import { KanbanBoard } from "@/components/board/kanban-board";
import { useCurrentProject } from "@/data/current-project";
import { useProject, useTasks } from "@/data/queries";
import { pluralizeCount } from "@/lib/format";

/**
 * Реестр задач проекта вне Ганта — канбан-доска по статусам.
 *
 * Задачи берутся тем же `useTasks`, что и в оболочке и поиске: ключ общий,
 * лишнего запроса при переходе на доску нет. `live` добавляет к этому
 * регулярное обновление — на доске карточки двигает вся команда сразу.
 */
export default function BoardPage() {
  const { projectId, isEmpty, error: projectsError } = useCurrentProject();
  const project = useProject(projectId);
  const tasks = useTasks(projectId, { live: true });

  if (projectsError) return <PageError error={projectsError} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} onRetry={project.refetch} />;
  if (tasks.error) return <PageError error={tasks.error} onRetry={tasks.refetch} />;
  if (!projectId || !project.data || !tasks.data) return <PageLoading />;

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <PageHeader
        eyebrow={`Проект #${project.data.code}`}
        title="Канбан-доска"
        description={`${project.data.name} · ${pluralizeCount(tasks.data.length, ["задача", "задачи", "задач"])}`}
      />

      {tasks.data.length === 0 ? (
        <EmptyProjectTasks />
      ) : (
        <KanbanBoard projectId={projectId} tasks={tasks.data} />
      )}
    </div>
  );
}
