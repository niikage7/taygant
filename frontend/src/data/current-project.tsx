"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useSyncExternalStore,
  type ReactNode,
} from "react";

import { useProjects } from "@/data/queries";
import type { ProjectSummary } from "@/types";

const STORAGE_KEY = "taygant.currentProjectId";

/**
 * Мини-хранилище поверх localStorage с оповещением подписчиков. Обычный
 * `useState` + `useEffect` для первого чтения запрещён линтером
 * (react-hooks/set-state-in-effect) и всё равно расходился бы на сервере
 * (localStorage там нет) — useSyncExternalStore решает оба вопроса разом:
 * на сервере и при первом клиентском рендере отдаёт `null` (совпадает с SSR),
 * а сразу после гидратации подставляет настоящее значение без ручного setState.
 */
const listeners = new Set<() => void>();

/** Выбор, сделанный в этой вкладке: переживает запрет записи в localStorage. */
let selectedInSession: string | null = null;

function readStoredProjectId(): string | null {
  if (selectedInSession) return selectedInSession;
  try {
    return window.localStorage.getItem(STORAGE_KEY);
  } catch {
    // Приватный режим может запрещать доступ — тогда просто нет запомненного выбора.
    return null;
  }
}

function writeStoredProjectId(id: string): void {
  selectedInSession = id;
  try {
    window.localStorage.setItem(STORAGE_KEY, id);
  } catch {
    // Приватный режим может запрещать запись — выбор тогда живёт до перезагрузки.
  }
  listeners.forEach((listener) => listener());
}

function subscribeStoredProjectId(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function useStoredProjectId(): string | null {
  return useSyncExternalStore(subscribeStoredProjectId, readStoredProjectId, () => null);
}

type CurrentProjectValue = {
  projectId: string | undefined;
  projects: ProjectSummary[];
  selectProject: (projectId: string) => void;
  isLoading: boolean;
  isEmpty: boolean;
  error: Error | null;
};

const CurrentProjectContext = createContext<CurrentProjectValue | null>(null);

/**
 * Выбранный проект, общий для всех внутренних экранов.
 *
 * Хранится в localStorage, чтобы выбор переживал переходы между разделами и
 * перезагрузку: иначе пользователь с несколькими проектами возвращался бы к
 * первому после каждого клика по навигации.
 *
 * Сохранённый идентификатор проверяется по списку с сервера — проект могли
 * удалить или отобрать к нему доступ, и тогда все экраны молча получали бы 404.
 */
export function CurrentProjectProvider({ children }: { children: ReactNode }) {
  const projects = useProjects();
  const storedId = useStoredProjectId();

  const list = useMemo(() => projects.data ?? [], [projects.data]);
  const projectId =
    list.find((project) => project.id === storedId)?.id ?? list[0]?.id;

  const selectProject = useCallback((id: string) => {
    writeStoredProjectId(id);
  }, []);

  const value = useMemo<CurrentProjectValue>(
    () => ({
      projectId,
      projects: list,
      selectProject,
      isLoading: projects.isPending,
      isEmpty: projects.isSuccess && list.length === 0,
      error: projects.error,
    }),
    [projectId, list, selectProject, projects.isPending, projects.isSuccess, projects.error],
  );

  return (
    <CurrentProjectContext.Provider value={value}>
      {children}
    </CurrentProjectContext.Provider>
  );
}

export function useCurrentProject(): CurrentProjectValue {
  const value = useContext(CurrentProjectContext);
  if (!value) {
    throw new Error("useCurrentProject доступен только внутри CurrentProjectProvider");
  }
  return value;
}
