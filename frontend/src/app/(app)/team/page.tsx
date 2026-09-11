"use client";

import { PageHeader } from "@/components/app/page-header";
import { EmptyProjects, PageError, PageLoading } from "@/components/app/page-state";
import { TeamManager } from "@/components/project/team-manager";
import { useCurrentProject } from "@/data/current-project";
import { useProjectAccess } from "@/data/project-access";
import { useProject } from "@/data/queries";

/**
 * Команда проекта: состав, роли и уровни доступа.
 *
 * Добавление и исключение участников вынесены в отдельный экран, а не оставлены
 * только шагу мастера при создании: состав меняется на всём протяжении проекта.
 */
export default function TeamPage() {
  const { projectId, isEmpty, error } = useCurrentProject();
  const project = useProject(projectId);
  const access = useProjectAccess();

  if (error) return <PageError error={error} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} onRetry={project.refetch} />;
  if (!projectId || !project.data) return <PageLoading />;

  return (
    <div className="mx-auto max-w-[1000px] space-y-4">
      <PageHeader eyebrow={`Проект #${project.data.code}`} title="Состав команды и права доступа" />

      <TeamManager
        projectId={projectId}
        projectName={project.data.name}
        canManage={access.isFull}
      />
    </div>
  );
}
