import type { ReactNode } from "react";

import { Sidebar } from "@/components/app/sidebar";
import { TopBar } from "@/components/app/top-bar";
import { demoProject } from "@/data/demo";

/**
 * Оболочка внутренних экранов: слева навигация по проекту, сверху поиск.
 *
 * Сводка проекта в сайдбаре берётся из демо-данных — см. предупреждение
 * в `src/data/demo.ts`.
 */
export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen flex-1">
      <Sidebar
        projectName="ИС мониторинга ТПУ"
        healthIndex={demoProject.healthIndex}
        criticalRisksCount={demoProject.criticalRisksCount}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar />
        <main className="min-w-0 flex-1 bg-page p-6">{children}</main>
      </div>
    </div>
  );
}
