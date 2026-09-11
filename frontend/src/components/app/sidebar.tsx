"use client";

import * as Dialog from "@radix-ui/react-dialog";
import * as Tooltip from "@radix-ui/react-tooltip";
import {
  BarChart3,
  Bot,
  CircleAlert,
  CirclePlus,
  Download,
  GraduationCap,
  GitBranch,
  LineChart,
  LoaderCircle,
  Users,
  X,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState, type ComponentType } from "react";

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

const EXPORT_FORMATS: { format: ExportFormat; label: string }[] = [
  { format: "pdf", label: "Скачать PDF" },
  { format: "xlsx", label: "Скачать XLSX" },
];

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

type SidebarProps = {
  projects: ProjectSummary[];
  currentProjectId: string | undefined;
  onSelectProject: (projectId: string) => void;
  /** Адрес первой задачи проекта для пункта «Детали и Зависимости». */
  detailsHref?: string | null;
  onStartTour: () => void;
};

/**
 * Навигация проекта. От `lg` — постоянная панель слева; ниже `lg` содержимое
 * то же самое, но открывается выезжающей панелью (Radix Dialog даёт фокус-
 * ловушку, закрытие по Esc и оверлей бесплатно).
 */
export function Sidebar({
  mobileOpen,
  onMobileOpenChange,
  ...props
}: SidebarProps & { mobileOpen: boolean; onMobileOpenChange: (open: boolean) => void }) {
  const pathname = usePathname();

  // Страховка на случай перехода мимо клика по ссылке в панели (например,
  // кнопкой «назад» в браузере) — панель не должна остаться открытой поверх
  // нового экрана. onMobileOpenChange — это setState из (app)/layout.tsx и
  // как таковой имеет стабильную ссылку, поэтому в зависимостях безопасен.
  useEffect(() => {
    onMobileOpenChange(false);
  }, [pathname, onMobileOpenChange]);

  return (
    <>
      <aside className="hidden w-70 shrink-0 flex-col border-r border-line bg-surface lg:flex">
        <SidebarContent {...props} pathname={pathname} />
      </aside>

      <Dialog.Root open={mobileOpen} onOpenChange={onMobileOpenChange}>
        <Dialog.Portal>
          <Dialog.Overlay className="fixed inset-0 z-40 bg-ink/30 lg:hidden" />
          <Dialog.Content
            className="fixed inset-y-0 left-0 z-50 flex w-70 max-w-[85vw] flex-col bg-surface shadow-popover outline-none lg:hidden"
            aria-describedby={undefined}
          >
            <Dialog.Title className="sr-only">Меню навигации</Dialog.Title>
            <Dialog.Close asChild>
              <button
                type="button"
                aria-label="Закрыть меню"
                className="absolute top-3 right-3 z-10 flex size-8 cursor-pointer items-center justify-center rounded-control text-ink-muted hover:bg-surface-muted hover:text-ink focus-visible:focus-ring"
              >
                <X className="size-4" />
              </button>
            </Dialog.Close>
            <SidebarContent
              {...props}
              pathname={pathname}
              onNavigate={() => onMobileOpenChange(false)}
            />
          </Dialog.Content>
        </Dialog.Portal>
      </Dialog.Root>
    </>
  );
}

/**
 * Содержимое сайдбара — единственное определение, используется и в
 * постоянной панели, и в мобильной выезжающей: расхождение разметки между
 * ними означало бы два места для правки одного и того же меню.
 */
function SidebarContent({
  projects,
  currentProjectId,
  onSelectProject,
  detailsHref = null,
  onStartTour,
  pathname,
  onNavigate,
}: SidebarProps & { pathname: string; onNavigate?: () => void }) {
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
      // toUserMessage сам отличает 501 (раздел ещё не реализован на бэкенде)
      // от настоящей ошибки — экспорт пока возвращает 501 для обоих форматов.
      setExportError(toUserMessage(error, {}, "Не удалось подготовить файл"));
    } finally {
      setExporting(null);
    }
  };

  return (
    <Tooltip.Provider delayDuration={300}>
      <Link
        href="/overview"
        onClick={onNavigate}
        className="flex items-center gap-2.5 px-4 py-3.5 focus-visible:focus-ring"
      >
        <LogoMark className="size-8" />
        <span className="min-w-0 truncate text-sm font-bold text-ink">taygant</span>
      </Link>

      <nav className="flex-1 overflow-y-auto px-2 pt-3" aria-label="Разделы проекта" data-tour="nav">
        <p className="px-2 pb-2 text-2xs font-semibold tracking-wider text-ink-faint uppercase">
          Рабочее пространство
        </p>
        <ul className="space-y-0.5">
          {items.map((item) => {
            const active = pathname === item.href || pathname.startsWith(`${item.href}/`);
            const Icon = item.icon;
            const className = cn(
              "flex items-center gap-2.5 rounded-control px-2.5 py-2.5 text-sm transition-colors",
              active ? "bg-brand-tint font-semibold text-brand" : "text-ink-muted",
              item.disabled
                ? "cursor-not-allowed opacity-50"
                : "cursor-pointer hover:bg-surface-muted hover:text-ink focus-visible:focus-ring",
            );

            return (
              <li key={item.label}>
                {item.disabled ? (
                  <Tooltip.Root>
                    <Tooltip.Trigger asChild>
                      <span tabIndex={0} aria-disabled className={className}>
                        <Icon className="size-4 shrink-0" />
                        {item.label}
                      </span>
                    </Tooltip.Trigger>
                    <Tooltip.Portal>
                      <Tooltip.Content
                        side="right"
                        sideOffset={6}
                        className="z-50 max-w-56 rounded-control bg-ink px-2.5 py-1.5 text-2xs text-white shadow-popover"
                      >
                        Появится, когда в проекте будут задачи
                        <Tooltip.Arrow className="fill-ink" />
                      </Tooltip.Content>
                    </Tooltip.Portal>
                  </Tooltip.Root>
                ) : (
                  <Link
                    href={item.href}
                    onClick={onNavigate}
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
          className="mt-4 flex cursor-pointer items-center gap-2 rounded-control text-13 text-ink-muted transition-colors hover:text-brand focus-visible:focus-ring"
        >
          <GraduationCap className="size-4" />
          Как это работает
        </button>

        <div className="mt-4 flex items-center justify-between gap-2">
          <span className="flex items-center gap-1.5 text-2xs font-semibold tracking-wider text-ink-faint uppercase">
            <Download className="size-3.5 shrink-0" aria-hidden />
            Экспорт
          </span>
          <span className="flex gap-1.5">
            {EXPORT_FORMATS.map(({ format, label }) => (
              <button
                key={format}
                type="button"
                onClick={() => void download(format)}
                disabled={exporting !== null || !projectId}
                aria-label={label}
                aria-busy={exporting === format}
                title={projectId ? undefined : "Сначала выберите проект"}
                className="flex min-w-9 cursor-pointer items-center justify-center rounded-control border border-line px-2 py-1 text-2xs font-medium uppercase text-ink-muted transition-colors hover:bg-surface-muted hover:text-ink disabled:cursor-not-allowed disabled:opacity-50 focus-visible:focus-ring"
              >
                {exporting === format ? (
                  <LoaderCircle className="size-3.5 animate-spin motion-reduce:animate-none" aria-hidden />
                ) : (
                  format
                )}
              </button>
            ))}
          </span>
        </div>
        {exportError ? (
          <p role="alert" className="mt-1.5 flex items-center gap-1.5 text-2xs text-danger-ink">
            <CircleAlert className="size-3.5 shrink-0" aria-hidden />
            {exportError}
          </p>
        ) : null}
      </div>
    </Tooltip.Provider>
  );
}

export type { SidebarProps };
