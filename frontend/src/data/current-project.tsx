"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from "react";

import { useProjects } from "@/data/queries";
import type { ProjectSummary } from "@/types";

const STORAGE_KEY = "taygant.currentProjectId";

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
  const [storedId, setStoredId] = useState<string | null>(null);

  // Читаем localStorage в эффекте, а не при рендере: на сервере его нет,
  // и обращение во время рендера рассинхронизировало бы разметку.
  useEffect(() => {
    try {
      setStoredId(window.localStorage.getItem(STORAGE_KEY));
    } catch {
      setStoredId(null);
    }
  }, []);

  const list = useMemo(() => projects.data ?? [], [projects.data]);
  const projectId =
    list.find((project) => project.id === storedId)?.id ?? list[0]?.id;

  const selectProject = useCallback((id: string) => {
    setStoredId(id);
    try {
      window.localStorage.setItem(STORAGE_KEY, id);
    } catch {
      // Приватный режим может запрещать запись — выбор тогда живёт до перезагрузки.
    }
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
