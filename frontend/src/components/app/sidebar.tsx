"use client";

import {
  BarChart3,
  CirclePlus,
  GitBranch,
  LineChart,
  Settings,
  TriangleAlert,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import type { ComponentType } from "react";

import { LogoMark } from "@/components/brand/logo";
import { Progress } from "@/components/ui/progress";
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
  projectName,
  healthIndex,
  criticalRisksCount,
  detailsHref = null,
}: {
  projectName: string;
  healthIndex: number;
  criticalRisksCount: number;
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
        <span className="min-w-0">
          <span className="block text-sm leading-tight font-bold text-ink">
            ПроектКонтроль
          </span>
          <span className="block text-[10px] font-semibold tracking-wider text-brand uppercase">
            ТПУ Enterprise
          </span>
        </span>
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
        <div className="rounded-control border border-line p-3">
          <div className="flex items-baseline justify-between gap-2">
            <span className="truncate text-[11px] font-semibold tracking-wider text-ink-faint uppercase">
              {projectName}
            </span>
            <span className="text-xs font-bold text-brand">{healthIndex}%</span>
          </div>
          <Progress
            value={healthIndex}
            className="mt-2 h-1.5"
            label="Индекс здоровья проекта"
          />
          {criticalRisksCount > 0 ? (
            <p className="mt-2 flex items-center gap-1.5 text-xs text-danger">
              <TriangleAlert className="size-3.5 shrink-0" />
              {criticalRisksCount} критический риск
            </p>
          ) : null}
        </div>

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

        <button
          type="button"
          className="mt-3 flex items-center gap-2.5 rounded-control text-sm text-ink-muted transition-colors hover:text-ink focus-visible:focus-ring"
        >
          <Settings className="size-4" />
          Параметры системы
        </button>
      </div>
    </aside>
  );
}
