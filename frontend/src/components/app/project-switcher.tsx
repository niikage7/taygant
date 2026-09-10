"use client";

import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { Check, ChevronsUpDown, TriangleAlert } from "lucide-react";

import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import type { ProjectSummary } from "@/types";

/**
 * Карточка текущего проекта в сайдбаре.
 *
 * Когда проект один, это просто сводка — выпадающий список без выбора только
 * мешал бы. Как только проектов несколько, та же карточка становится кнопкой
 * переключения, чтобы не занимать место отдельным элементом навигации.
 */
export function ProjectSwitcher({
  projects,
  currentProjectId,
  onSelect,
}: {
  projects: ProjectSummary[];
  currentProjectId: string | undefined;
  onSelect: (projectId: string) => void;
}) {
  const current = projects.find((project) => project.id === currentProjectId);
  const progress = current?.progressPercent ?? 0;
  const health = current?.healthIndex ?? 0;

  const summary = (
    <>
      <div className="flex items-baseline justify-between gap-2">
        <span className="truncate text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
          {current?.name ?? "Проект не выбран"}
        </span>
        <span className="shrink-0 text-xs font-bold text-brand">
          {current ? `${Math.round(progress)}%` : "—"}
        </span>
      </div>
      <Progress
        value={progress}
        className="mt-2 h-1.5"
        label={`Прогресс проекта: ${Math.round(progress)}%`}
      />
      <div className="mt-1.5 flex items-baseline justify-between gap-2 text-[11px] text-ink-faint">
        <span className="tracking-wider uppercase">Прогресс</span>
        {current ? <span>Здоровье: {Math.round(health)}%</span> : null}
      </div>
      {current && current.criticalRisksCount > 0 ? (
        <p className="mt-2 flex items-center gap-1.5 text-xs text-danger">
          <TriangleAlert className="size-3.5 shrink-0" />
          {current.criticalRisksCount} критический риск
        </p>
      ) : null}
    </>
  );

  if (projects.length < 2) {
    return <div className="rounded-control border border-line p-3">{summary}</div>;
  }

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        className={cn(
          "w-full rounded-control border border-line p-3 text-left transition-colors",
          "hover:bg-surface-subtle focus-visible:focus-ring",
          "data-[state=open]:border-brand data-[state=open]:bg-surface-subtle",
        )}
        aria-label={`Текущий проект: ${current?.name ?? "не выбран"}. Сменить проект`}
      >
        <div className="flex items-start gap-1.5">
          <div className="min-w-0 flex-1">{summary}</div>
          <ChevronsUpDown className="mt-0.5 size-3.5 shrink-0 text-ink-faint" />
        </div>
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          side="top"
          align="start"
          sideOffset={6}
          className="z-50 max-h-80 w-64 overflow-y-auto rounded-card border border-line bg-surface p-1 shadow-popover"
        >
          <DropdownMenu.Label className="px-2 py-1.5 text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
            Проекты ({projects.length})
          </DropdownMenu.Label>
          {projects.map((project) => {
            const active = project.id === currentProjectId;
            return (
              <DropdownMenu.Item
                key={project.id}
                onSelect={() => onSelect(project.id)}
                className={cn(
                  "flex cursor-pointer items-start gap-2 rounded-control px-2 py-2 text-[13px] outline-none",
                  "data-[highlighted]:bg-surface-muted",
                  active && "text-brand",
                )}
              >
                <Check
                  className={cn("mt-0.5 size-3.5 shrink-0", !active && "opacity-0")}
                />
                <span className="min-w-0 flex-1">
                  <span className="block truncate font-medium">{project.name}</span>
                  <span className="mt-0.5 block truncate font-mono text-[11px] text-ink-faint">
                    {project.code} · {Math.round(project.progressPercent)}%
                  </span>
                </span>
              </DropdownMenu.Item>
            );
          })}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
