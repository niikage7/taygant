import { CalendarDays, CircleCheck, Waypoints } from "lucide-react";
import type { Metadata } from "next";

import { GanttBoard } from "@/components/gantt/gantt-board";
import { Badge } from "@/components/ui/badge";
import { demoDependencies, demoProject, demoTasks, demoUsers } from "@/data/demo";
import { formatDate } from "@/lib/format";

export const metadata: Metadata = {
  title: "Диаграмма Ганта",
  description: "Календарно-сетевой график проекта с расчётом критического пути",
};

export default function GanttPage() {
  const project = demoProject;
  const criticalStages = demoTasks.filter((task) => task.isCriticalPath).length;

  return (
    <div className="mx-auto max-w-[1400px] space-y-4">
      <div>
        <h1 className="flex flex-wrap items-center gap-2 text-lg font-bold tracking-tight text-ink">
          <span className="font-mono text-sm text-brand">{project.code}</span>
          <span className="text-ink-faint">·</span>
          {project.name}
          <Badge tone="neutral" size="sm" className="font-mono">v2.4</Badge>
        </h1>
        <p className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-[13px] text-ink-muted">
          <span className="flex items-center gap-1.5">
            <CalendarDays className="size-3.5" />
            {formatDate(project.startDate)} — {formatDate(project.deadline)} (
            {project.durationCalendarDays} дней)
          </span>
          <span className="flex items-center gap-1.5">
            <CircleCheck className="size-3.5" />
            Прогресс: {project.progressPercent}%
          </span>
          <span className="flex items-center gap-1.5 text-danger">
            <Waypoints className="size-3.5" />
            Критический путь: {criticalStages} этапа
          </span>
        </p>
      </div>

      <GanttBoard
        project={project}
        tasks={demoTasks}
        dependencies={demoDependencies}
        currentUser={demoUsers.mikhail}
        today="2025-11-11"
      />
    </div>
  );
}
