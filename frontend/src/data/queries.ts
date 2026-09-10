"use client";

import { useQueries, useQuery, type UseQueryResult } from "@tanstack/react-query";

import {
  dependenciesService,
  milestonesService,
  projectsService,
  tasksService,
} from "@/services";
import type { Project, ProjectSummary, Task, TaskDependency, TaskDetail } from "@/types";

/**
 * Слой доступа к данным экранов.
 *
 * Состояние бэкенда на момент написания (см. `backend/internal/api/`):
 * реализованы projects, tasks, dependencies, milestones, members, sprints,
 * checklist, comments, history, users и auth. Возвращают 501: dashboard,
 * workload, simulation и export. Маршрут `/projects/{id}/gantt` не
 * зарегистрирован вовсе — CPM-движка пока нет, поэтому `tasksService.getGantt`
 * вызывать нельзя, задачи и связи собираются здесь по отдельности.
 */

export const queryKeys = {
  projects: ["projects"] as const,
  project: (id: string) => ["projects", id] as const,
  tasks: (projectId: string) => ["projects", projectId, "tasks"] as const,
  milestones: (projectId: string) => ["projects", projectId, "milestones"] as const,
  taskDependencies: (taskId: string) => ["tasks", taskId, "dependencies"] as const,
  task: (taskId: string) => ["tasks", taskId] as const,
};

export function useProjects(): UseQueryResult<ProjectSummary[]> {
  return useQuery({
    queryKey: queryKeys.projects,
    queryFn: () => projectsService.list(),
  });
}

export function useProject(projectId: string | undefined): UseQueryResult<Project> {
  return useQuery({
    queryKey: queryKeys.project(projectId ?? ""),
    queryFn: () => projectsService.getById(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useTasks(projectId: string | undefined): UseQueryResult<Task[]> {
  return useQuery({
    queryKey: queryKeys.tasks(projectId ?? ""),
    queryFn: () => tasksService.list(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useMilestones(projectId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.milestones(projectId ?? ""),
    queryFn: () => milestonesService.list(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useTaskDetail(taskId: string): UseQueryResult<TaskDetail> {
  return useQuery({
    queryKey: queryKeys.task(taskId),
    queryFn: () => tasksService.getById(taskId),
  });
}

/**
 * Все связи проекта.
 *
 * Эндпоинта «связи проекта» нет — есть только связи конкретной задачи, поэтому
 * опрашиваем задачи параллельно и берём у каждой её предшественников: каждое
 * ребро принадлежит ровно одной задаче как входящее, так что дубликатов не
 * возникает и склеивать ничего не нужно.
 */
export function useProjectDependencies(tasks: Task[] | undefined): {
  dependencies: TaskDependency[];
  isLoading: boolean;
} {
  const results = useQueries({
    queries: (tasks ?? []).map((task) => ({
      queryKey: queryKeys.taskDependencies(task.id),
      queryFn: () => dependenciesService.list(task.id),
    })),
  });

  return {
    dependencies: results.flatMap((result) => result.data?.predecessors ?? []),
    isLoading: results.some((result) => result.isPending),
  };
}
