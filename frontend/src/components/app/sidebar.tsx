"use client";

"use client";

import {
  BarChart3,
  Bot,
  CircleAlert,
  CirclePlus,
  Download,
  GraduationCap,
  GitBranch,
  LineChart,
  Users,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, type ComponentType } from "react";

import { LogoMark } from "@/components/brand/logo";
import { ProjectSwitcher } from "@/components/app/project-switcher";
import { useCurrentProject } from "@/data/current-project";
import { toUserMessage } from "@/lib/api-error-message";
import { cn } from "@/lib/utils";
import { exportService } from "@/services";
import type { ExportFormat, ProjectSummary } from "@/types";

type NavItem = {
  href: string;
  label: string;
  icon: ComponentType<{ className?: string }>;
};

/**
 * «Детали и Зависимости» ведут на конкретную задачу, поэтому адрес известен
 * только когда задачи проекта загружены. Пока их нет, пункт остаётся видимым,
 * но неактивным — исчезающий и появляющийся пункт меню читался бы как сбой.
 */
function navItems(detailsHref: string | null): (NavItem & { disabled?: boolean })[] {
  return [
    { href: "/gantt", label: "Диаграмма Ганта", icon: BarChart3 },
    { href: "/overview", label: "Обзор и Аналитика", icon: LineChart },
    {
      href: detailsHref ?? "/gantt",
      label: "Детали и Зависимости",
      icon: GitBranch,
      disabled: !detailsHref,
    },
    { href: "/team", label: "Команда проекта", icon: Users },
    { href: "/assistant", label: "Ассистент (MCP)", icon: Bot },
    { href: "/projects/new", label: "Новый проект", icon: CirclePlus },
  ];
}

export function Sidebar({
  projects,
  currentProjectId,
  onSelectProject,
  detailsHref = null,
  onStartTour,
}: {
  projects: ProjectSummary[];
  currentProjectId: string | undefined;
  onSelectProject: (projectId: string) => void;
  /** Адрес первой задачи проекта для пункта «Детали и Зависимости». */
  detailsHref?: string | null;
  onStartTour: () => void;
}) {
  const pathname = usePathname();
  const items = navItems(detailsHref);
  const { projectId } = useCurrentProject();
  const [exporting, setExporting] = useState<ExportFormat | null>(null);
  const [exportError, setExportError] = useState<string | null>(null);

  const download = async (format: ExportFormat) => {
    if (!projectId || exporting) return;
    setExporting(format);
    setExportError(null);
    try {
      const { blob, filename } = await exportService.export(projectId, format);
      const url = URL.createObjectURL(blob);
      const anchor = document.createElement("a");
      anchor.href = url;
      anchor.download = filename;
      anchor.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      setExportError(toUserMessage(error, {}, "Не удалось подготовить файл"));
    } finally {
      setExporting(null);
    }
  };

  return (
    <aside className="flex w-70 shrink-0 flex-col border-r border-line bg-surface">
      <Link
        href="/overview"
        className="flex items-center gap-2.5 px-4 py-3.5 focus-visible:focus-ring"
      >
        <LogoMark className="size-8" />
        <span className="min-w-0 truncate text-sm font-bold text-ink">taygant</span>
      </Link>

      <nav className="flex-1 px-2 pt-3" aria-label="Разделы проекта" data-tour="nav">
        <p className="px-2 pb-2 text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
          Рабочее пространство
        </p>
        <ul className="space-y-0.5">
          {items.map((item) => {
            const active =
              pathname === item.href || pathname.startsWith(`${item.href}/`);
            const Icon = item.icon;
            const className = cn(
              "flex items-center gap-2.5 rounded-control px-2.5 py-2.5 text-sm transition-colors",
              active
                ? "bg-brand-tint font-semibold text-brand"
                : "text-ink-muted",
              item.disabled
                ? "cursor-not-allowed opacity-50"
                : "hover:bg-surface-muted hover:text-ink focus-visible:focus-ring",
            );

            return (
              <li key={item.label}>
                {item.disabled ? (
                  <span
                    aria-disabled
                    title="Появится, когда в проекте будут задачи"
                    className={className}
                  >
                    <Icon className="size-4 shrink-0" />
                    {item.label}
                  </span>
                ) : (
                  <Link
                    href={item.href}
                    aria-current={active ? "page" : undefined}
                    className={className}
                  >
                    <Icon className="size-4 shrink-0" />
                    {item.label}
                  </Link>
                )}
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="border-t border-line p-4">
        <div data-tour="project-switcher">
          <ProjectSwitcher
            projects={projects}
            currentProjectId={currentProjectId}
            onSelect={onSelectProject}
          />
        </div>

        <button
          type="button"
          onClick={onStartTour}
          className="mt-4 flex items-center gap-2 rounded-control text-[13px] text-ink-muted transition-colors hover:text-brand focus-visible:focus-ring"
        >
          <GraduationCap className="size-4" />
          Как это работает
        </button>

        <div className="mt-4 flex items-center justify-between gap-2">
          <span className="flex items-center gap-1.5 text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
            <Download className="size-3.5 shrink-0" />
            Экспорт
          </span>
          <span className="flex gap-1.5">
            {(["pdf", "xlsx"] as ExportFormat[]).map((format) => (
              <button
                key={format}
                type="button"
                onClick={() => void download(format)}
                disabled={exporting !== null}
                title={projectId ? undefined : "Сначала выберите проект"}
                className="rounded-control border border-line px-2 py-1 text-[11px] font-medium uppercase text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink disabled:opacity-50 disabled:cursor-not-allowed focus-visible:focus-ring"
              >
                {exporting === format ? "…" : format}
              </button>
            ))}
          </span>
        </div>
        {exportError ? (
          <p className="mt-1.5 flex items-center gap-1.5 text-[11px] text-danger">
            <CircleAlert className="size-3.5 shrink-0" />
            {exportError}
          </p>
        ) : null}
      </div>
    </aside>
  );
}
