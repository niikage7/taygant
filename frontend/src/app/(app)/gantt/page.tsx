"use client";

import { format } from "date-fns";
import { CalendarDays, CircleCheck, Waypoints } from "lucide-react";

import { EmptyProjectTasks, EmptyProjects, PageError, PageLoading } from "@/components/app/page-state";
import { GanttBoard } from "@/components/gantt/gantt-board";
import { Badge } from "@/components/ui/badge";
import { useCurrentProject } from "@/data/current-project";
import { useGantt, useProject } from "@/data/queries";
import { formatDate } from "@/lib/format";
import { getStoredUser } from "@/lib/session";

export default function GanttPage() {
  const { projectId, isEmpty, error: projectsError } = useCurrentProject();
  const project = useProject(projectId);
  // Масштаб влияет только на отрисовку, поэтому запрашиваем один раз в weeks:
  // задачи, связи и вехи от него не зависят.
  const gantt = useGantt(projectId, "weeks");

  if (projectsError) return <PageError error={projectsError} />;
  if (isEmpty) return <EmptyProjects />;
  if (project.error) return <PageError error={project.error} />;
  if (gantt.error) return <PageError error={gantt.error} />;
  if (!project.data || !gantt.data) return <PageLoading />;

  const currentUser = getStoredUser();
  const { tasks, dependencies, milestones } = gantt.data;
  const criticalStages = tasks.filter((task) => task.isCriticalPath).length;

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <div>
        <h1 className="flex flex-wrap items-center gap-2 text-lg font-bold tracking-tight text-ink">
          <span className="font-mono text-sm text-brand">{project.data.code}</span>
          <span className="text-ink-faint">·</span>
          {project.data.name}
          {project.data.phase ? (
            <Badge tone="neutral" size="sm">
              {project.data.phase}
            </Badge>
          ) : null}
        </h1>
        <p className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-13 text-ink-muted">
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

      {tasks.length === 0 ? (
        <EmptyProjectTasks />
      ) : (
        <GanttBoard
          project={project.data}
          tasks={tasks}
          dependencies={dependencies}
          milestones={milestones}
          currentUserId={currentUser?.id}
          // Местная дата: toISOString даёт UTC, и восточнее Гринвича после
          // полуночи «сегодня» оставалось бы вчерашним днём.
          today={format(new Date(), "yyyy-MM-dd")}
        />
      )}
    </div>
  );
}
