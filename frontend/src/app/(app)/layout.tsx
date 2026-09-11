"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState, useSyncExternalStore, type ReactNode } from "react";

import { SessionCheckLoading } from "@/components/app/page-state";
import { Sidebar } from "@/components/app/sidebar";
import { ProductTour } from "@/components/tour/product-tour";
import { TopBar } from "@/components/app/top-bar";
import { CurrentProjectProvider, useCurrentProject } from "@/data/current-project";
import { useTasks } from "@/data/queries";
import { isTourCompleted } from "@/lib/tour";
import { getAccessToken } from "@/services/http-client";

type AuthState = "checking" | "authorized" | "anonymous";

const noSubscription = () => () => {};

/**
 * Есть ли Bearer-токен. Токен живёт в localStorage, которого нет на сервере —
 * useSyncExternalStore, а не useState+useEffect, отдаёт на сервере и при
 * первом клиентском рендере одинаковое «checking» (иначе разметка разошлась
 * бы при гидратации) и подставляет настоящее значение сразу после неё, без
 * ручного setState в эффекте.
 */
function useAuthState(): AuthState {
  return useSyncExternalStore(
    noSubscription,
    () => (getAccessToken() ? "authorized" : "anonymous"),
    () => "checking",
  );
}

/** Оболочка внутренних экранов. Все эндпоинты требуют Bearer-токен. */
export default function AppLayout({ children }: { children: ReactNode }) {
  return (
    <CurrentProjectProvider>
      <AppShell>{children}</AppShell>
    </CurrentProjectProvider>
  );
}

function AppShell({ children }: { children: ReactNode }) {
  const router = useRouter();
  const authState = useAuthState();
  const [tourOpen, setTourOpen] = useState(false);
  const [tourAutoStarted, setTourAutoStarted] = useState(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Без токена сразу отправляем на вход, не дожидаясь 401 от первого запроса.
  useEffect(() => {
    if (authState === "anonymous") router.replace("/login");
  }, [authState, router]);

  // Тур запускаем один раз, сразу после того как сессия подтверждена: до
  // этого экраны ещё не отрисованы, и подсвечивать было бы нечего. Условие
  // выполняется прямо при рендере (а не в эффекте): react-hooks/set-state-in-effect
  // запрещает setState в эффектах, а обновление во время рендера, guarded
  // флагом «уже запускали», — санкционированный React способ реагировать на
  // переход authState в новое значение без лишнего useEffect.
  if (authState === "authorized" && !tourAutoStarted) {
    setTourAutoStarted(true);
    if (!isTourCompleted()) setTourOpen(true);
  }

  const { projects, projectId, selectProject } = useCurrentProject();
  // Тот же ключ, что и на экране Ганта, — React Query переиспользует ответ,
  // лишнего запроса не будет.
  const tasks = useTasks(projectId);
  const firstTaskId = tasks.data?.[0]?.id;

  if (authState === "checking") {
    return (
      <div className="flex min-h-screen flex-1 items-center justify-center bg-page p-6">
        <SessionCheckLoading />
      </div>
    );
  }

  if (authState === "anonymous") return null;

  return (
    <div className="flex min-h-screen flex-1">
      <Sidebar
        projects={projects}
        currentProjectId={projectId}
        onSelectProject={selectProject}
        detailsHref={firstTaskId ? `/tasks/${firstTaskId}` : null}
        onStartTour={() => setTourOpen(true)}
        mobileOpen={mobileMenuOpen}
        onMobileOpenChange={setMobileMenuOpen}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar onOpenMenu={() => setMobileMenuOpen(true)} menuOpen={mobileMenuOpen} />
        <main className="min-w-0 flex-1 overflow-x-hidden bg-page p-4 sm:p-6">{children}</main>
      </div>

      {tourOpen ? <ProductTour onClose={() => setTourOpen(false)} /> : null}
    </div>
  );
}
