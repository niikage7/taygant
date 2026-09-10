"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, type ReactNode } from "react";

import { PageLoading } from "@/components/app/page-state";
import { Sidebar } from "@/components/app/sidebar";
import { TopBar } from "@/components/app/top-bar";
import { useProjects, useTasks } from "@/data/queries";
import { getAccessToken } from "@/services/http-client";

/**
 * Оболочка внутренних экранов. Все эндпоинты требуют Bearer-токен, поэтому без
 * него сразу отправляем на вход, не дожидаясь 401 от первого запроса.
 *
 * Проверка выполняется в эффекте, а не при рендере: токен лежит в localStorage,
 * которого на сервере нет, и обращение к нему во время рендера дало бы
 * расхождение серверной и клиентской разметки.
 */
export default function AppLayout({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [authorized, setAuthorized] = useState<boolean | null>(null);

  useEffect(() => {
    if (getAccessToken()) {
      setAuthorized(true);
      return;
    }
    setAuthorized(false);
    router.replace("/login");
  }, [router]);

  const projects = useProjects();
  const project = projects.data?.[0];
  // Тот же ключ, что и на экране Ганта, — React Query переиспользует ответ,
  // лишнего запроса не будет.
  const tasks = useTasks(project?.id);
  const firstTaskId = tasks.data?.[0]?.id;

  if (authorized === null) {
    return (
      <div className="flex min-h-screen flex-1 items-center justify-center bg-page p-6">
        <PageLoading label="Проверяем сессию…" />
      </div>
    );
  }

  if (!authorized) return null;

  return (
    <div className="flex min-h-screen flex-1">
      <Sidebar
        projectName={project?.name ?? "Проект не выбран"}
        healthIndex={project?.healthIndex ?? 0}
        criticalRisksCount={project?.criticalRisksCount ?? 0}
        detailsHref={firstTaskId ? `/tasks/${firstTaskId}` : null}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar />
        <main className="min-w-0 flex-1 bg-page p-6">{children}</main>
      </div>
    </div>
  );
}
