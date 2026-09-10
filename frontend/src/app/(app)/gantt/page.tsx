"use client";

import { CalendarDays, CircleCheck, Waypoints } from "lucide-react";

import {
  EmptyProjects,
  PageError,
  PageLoading,
} from "@/components/app/page-state";
import { GanttBoard } from "@/components/gantt/gantt-board";
import { Badge } from "@/components/ui/badge";
import { useCurrentProject } from "@/data/current-project";
import {
  useProject,
  useProjectDependencies,
  useTasks,
} from "@/data/queries";
import { formatDate } from "@/lib/format";
import { getStoredUser } from "@/lib/session";

export default function GanttPage() {
  const { projectId, isEmpty, error: projectsError } = useCurrentProject();
  const project = useProject(projectId);
  const tasks = useTasks(projectId);
  const { dependencies, isLoading: dependenciesLoading } = useProjectDependencies(
    tasks.data,
  );

  if (projectsError) return <PageError error={projectsError} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} />;
  if (tasks.error) return <PageError error={tasks.error} />;
  if (!project.data || !tasks.data) return <PageLoading />;

  const currentUser = getStoredUser();
  const criticalStages = tasks.data.filter((task) => task.isCriticalPath).length;

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <div>
        <h1 className="flex flex-wrap items-center gap-2 text-lg font-bold tracking-tight text-ink">
          <span className="font-mono text-sm text-brand">{project.data.code}</span>
          <span className="text-ink-faint">·</span>
          {project.data.name}
          {project.data.phase ? (
            <Badge tone="neutral" size="sm">{project.data.phase}</Badge>
          ) : null}
        </h1>
        <p className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-[13px] text-ink-muted">
          <span className="flex items-center gap-1.5">
            <CalendarDays className="size-3.5" />
            {formatDate(project.data.startDate)} — {formatDate(project.data.deadline)} (
            {project.data.durationCalendarDays} дней)
          </span>
          <span className="flex items-center gap-1.5">
            <CircleCheck className="size-3.5" />
            Прогресс: {Math.round(project.data.progressPercent)}%
          </span>
          {criticalStages > 0 ? (
            <span className="flex items-center gap-1.5 text-danger">
              <Waypoints className="size-3.5" />
              Критический путь: {criticalStages} этапа
            </span>
          ) : null}
        </p>
      </div>

      {tasks.data.length === 0 ? (
        <PageLoading label="В проекте пока нет задач — добавьте первую." />
      ) : (
        <GanttBoard
          project={project.data}
          tasks={tasks.data}
          dependencies={dependencies}
          dependenciesLoading={dependenciesLoading}
          currentUserId={currentUser?.id}
          today={new Date().toISOString().slice(0, 10)}
        />
      )}
    </div>
  );
}
