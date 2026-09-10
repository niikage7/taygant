"use client";

import {
  BarChart3,
  CirclePlus,
  GitBranch,
  LineChart,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ComponentType } from "react";

import { LogoMark } from "@/components/brand/logo";
import { ProjectSwitcher } from "@/components/app/project-switcher";
import type { ProjectSummary } from "@/types";
import { cn } from "@/lib/utils";

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
    { href: "/projects/new", label: "Новый проект", icon: CirclePlus },
  ];
}

export function Sidebar({
  projects,
  currentProjectId,
  onSelectProject,
  detailsHref = null,
}: {
  projects: ProjectSummary[];
  currentProjectId: string | undefined;
  onSelectProject: (projectId: string) => void;
  /** Адрес первой задачи проекта для пункта «Детали и Зависимости». */
  detailsHref?: string | null;
}) {
  const pathname = usePathname();
  const items = navItems(detailsHref);

  return (
    <aside className="flex w-70 shrink-0 flex-col border-r border-line bg-surface">
      <Link
        href="/overview"
        className="flex items-center gap-2.5 px-4 py-3.5 focus-visible:focus-ring"
      >
        <LogoMark className="size-8" />
        <span className="min-w-0 truncate text-sm font-bold text-ink">taygant</span>
      </Link>

      <nav className="flex-1 px-2 pt-3" aria-label="Разделы проекта">
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
        <ProjectSwitcher
          projects={projects}
          currentProjectId={currentProjectId}
          onSelect={onSelectProject}
        />

        <div className="mt-4 flex items-center justify-between gap-2">
          <span className="text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
            Экспорт
          </span>
          <span className="flex gap-1.5">
            {["PDF", "XLSX"].map((format) => (
              <button
                key={format}
                type="button"
                className="rounded-control border border-line px-2 py-1 text-[11px] font-medium text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink focus-visible:focus-ring"
              >
                {format}
              </button>
            ))}
          </span>
        </div>
      </div>
    </aside>
  );
}
