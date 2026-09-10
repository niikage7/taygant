"use client";

import { useMutation, useQuery, useQueryClient, type UseQueryResult } from "@tanstack/react-query";

import {
  checklistService,
  commentsService,
  dashboardService,
  dependenciesService,
  historyService,
  membersService,
  simulationService,
  milestonesService,
  projectsService,
  tasksService,
} from "@/services";
import type {
  Comment,
  GanttChart,
  GanttScale,
  HistoryEntry,
  ApplyShiftRequest,
  Project,
  ProjectDashboard,
  ProjectMember,
  ProjectSummary,
  ShiftSimulation,
  SimulateShiftRequest,
  Task,
  TaskCreateRequest,
  TaskDependencyCreateRequest,
  TaskDetail,
  TaskUpdateRequest,
} from "@/types";

/**
 * Слой доступа к данным экранов.
 *
 * Состояние бэкенда на момент написания (см. `backend/internal/api/`):
 * реализованы projects, tasks (включая `/gantt`), dependencies, milestones,
 * members, sprints, checklist, comments, history, users и auth.
 * Возвращают 501 только `workload` и `export`.
 */

export const queryKeys = {
  projects: ["projects"] as const,
  project: (id: string) => ["projects", id] as const,
  tasks: (projectId: string) => ["projects", projectId, "tasks"] as const,
  milestones: (projectId: string) => ["projects", projectId, "milestones"] as const,
  gantt: (projectId: string, scale: string) => ["projects", projectId, "gantt", scale] as const,
  taskDependencies: (taskId: string) => ["tasks", taskId, "dependencies"] as const,
  taskHistory: (taskId: string) => ["tasks", taskId, "history"] as const,
  taskComments: (taskId: string) => ["tasks", taskId, "comments"] as const,
  dashboard: (projectId: string) => ["projects", projectId, "dashboard"] as const,
  members: (projectId: string) => ["projects", projectId, "members"] as const,
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
 * Данные диаграммы Ганта одним запросом: задачи, связи и вехи.
 *
 * Раньше связи собирались обходом задач — эндпоинта «связи проекта» не было.
 * Теперь `/projects/{id}/gantt` отдаёт всё сразу, поэтому вместо N+1 запросов
 * достаточно одного.
 */
export function useGantt(
  projectId: string | undefined,
  scale: GanttScale,
): UseQueryResult<GanttChart> {
  return useQuery({
    queryKey: queryKeys.gantt(projectId ?? "", scale),
    queryFn: () => tasksService.getGantt(projectId as string, { scale }),
    enabled: Boolean(projectId),
  });
}

export function useTaskHistory(taskId: string): UseQueryResult<HistoryEntry[]> {
  return useQuery({
    queryKey: queryKeys.taskHistory(taskId),
    queryFn: () => historyService.list(taskId),
  });
}

export function useTaskComments(taskId: string): UseQueryResult<Comment[]> {
  return useQuery({
    queryKey: queryKeys.taskComments(taskId),
    queryFn: () => commentsService.list(taskId),
  });
}

/**
 * После любой правки задачи инвалидируем и саму задачу, и списки, где она
 * показана: реестр Ганта и обзор считают прогресс и отставания по этим данным,
 * иначе экран остался бы со старыми цифрами до перезагрузки.
 */
function useTaskInvalidation(taskId: string) {
  const queryClient = useQueryClient();
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.task(taskId) }),
      queryClient.invalidateQueries({ queryKey: ["projects"] }),
    ]);
}

export function useUpdateTask(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (payload: TaskUpdateRequest) => tasksService.update(taskId, payload),
    onSuccess: invalidate,
  });
}

export function useAddComment(taskId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (text: string) => commentsService.add(taskId, { text }),
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.taskComments(taskId) }),
        // commentsCount живёт в карточке задачи — счётчик на вкладке иначе отстанет.
        queryClient.invalidateQueries({ queryKey: queryKeys.task(taskId) }),
      ]),
  });
}

export function useCreateDependency(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (payload: TaskDependencyCreateRequest) =>
      dependenciesService.create(taskId, payload),
    onSuccess: invalidate,
  });
}

export function useDeleteDependency(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (dependencyId: string) => dependenciesService.remove(dependencyId),
    onSuccess: invalidate,
  });
}

export function useAddChecklistItem(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (text: string) => checklistService.add(taskId, { text }),
    onSuccess: invalidate,
  });
}

export function useDeleteChecklistItem(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (itemId: string) => checklistService.remove(itemId),
    onSuccess: invalidate,
  });
}

export function useToggleChecklistItem(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: ({ itemId, isDone }: { itemId: string; isDone: boolean }) =>
      checklistService.update(itemId, { isDone }),
    onSuccess: invalidate,
  });
}

export function useDashboard(projectId: string | undefined): UseQueryResult<ProjectDashboard> {
  return useQuery({
    queryKey: queryKeys.dashboard(projectId ?? ""),
    queryFn: () => dashboardService.get(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useProjectMembers(projectId: string | undefined): UseQueryResult<ProjectMember[]> {
  return useQuery({
    queryKey: queryKeys.members(projectId ?? ""),
    queryFn: () => membersService.list(projectId as string),
    enabled: Boolean(projectId),
  });
}

export function useCreateTask(projectId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (payload: TaskCreateRequest) => tasksService.create(projectId as string, payload),
    // Новая задача меняет и реестр Ганта, и метрики обзора.
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["projects"] }),
  });
}

/**
 * What-if расчёт сдвига. Это чтение, а не изменение плана, но выполняется по
 * запросу пользователя, поэтому мутация, а не query: иначе React Query
 * пересчитывал бы сценарий сам при каждом фокусе окна.
 */
export function useSimulateShift(taskId: string) {
  return useMutation({
    mutationFn: (payload: SimulateShiftRequest): Promise<ShiftSimulation> =>
      simulationService.simulateShift(taskId, payload),
  });
}

export function useApplyShift(taskId: string) {
  const invalidate = useTaskInvalidation(taskId);
  return useMutation({
    mutationFn: (payload: ApplyShiftRequest): Promise<ShiftSimulation> =>
      simulationService.applyShift(taskId, payload),
    onSuccess: invalidate,
  });
}
