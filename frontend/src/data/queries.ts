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
  usersService,
  workloadService,
} from "@/services";
import type {
  Comment,
  GanttChart,
  GanttScale,
  HistoryEntry,
  ApplyShiftRequest,
  MemberWorkload,
  MilestoneCreateRequest,
  Project,
  ProjectDashboard,
  ProjectMember,
  ProjectMemberCreateRequest,
  ProjectMemberUpdateRequest,
  ProjectSummary,
  ShiftSimulation,
  SimulateShiftRequest,
  Task,
  TaskCreateRequest,
  TaskDependencyCreateRequest,
  TaskDetail,
  TaskUpdateRequest,
  User,
} from "@/types";

/**
 * Слой доступа к данным экранов.
 *
 * Состояние бэкенда на момент написания (см. `backend/internal/api/`):
 * реализованы projects, tasks (включая `/gantt`), dependencies, milestones,
 * members, sprints, checklist, comments, history, users, auth, workload и export.
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
  workload: (projectId: string) => ["projects", projectId, "workload"] as const,
  members: (projectId: string) => ["projects", projectId, "members"] as const,
  users: (search: string) => ["users", search] as const,
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

/**
 * Массовая привязка задач к вехе.
 *
 * Отдельного эндпоинта для этого нет — каждой задаче проставляется `milestoneId`
 * через PATCH. Запросы идут параллельно; результат сообщает, что не прошло,
 * а успешные привязки остаются: откатывать их хуже, чем показать частичный
 * результат.
 */
export function useAssignTasksToMilestone() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({
      taskIds,
      milestoneId,
    }: {
      taskIds: string[];
      milestoneId: string | null;
    }) => {
      const results = await Promise.allSettled(
        taskIds.map((taskId) => tasksService.update(taskId, { milestoneId })),
      );
      return results.flatMap((result, index) =>
        result.status === "rejected"
          ? [{ taskId: taskIds[index], error: result.reason as unknown }]
          : [],
      );
    },
  }); // <-- Добавьте эту строку (закрывает useMutation)
}
/**
 * Перенос задачи на диаграмме Ганта: меняются только сроки.
 *
 * Отдельный хук, а не useUpdateTask: перетаскивание идёт под курсором и должно
 * одним invalidate обновить весь экран Ганта, а не только карточку задачи.
 */
export function useMoveTask() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      taskId,
      startDate,
      endDate,
    }: {
      taskId: string;
      startDate: string;
      endDate: string;
    }) => tasksService.update(taskId, { startDate, endDate }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["projects"] }),
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

/** Загрузка команды по текущему спринту (`GET /projects/{id}/workload`). */
export function useWorkload(projectId: string | undefined): UseQueryResult<MemberWorkload[]> {
  return useQuery({
    queryKey: queryKeys.workload(projectId ?? ""),
    queryFn: () => workloadService.list(projectId as string),
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

/** Пользователи платформы — источник выбора при добавлении в команду проекта. */
export function useUsers(search?: string): UseQueryResult<User[]> {
  return useQuery({
    queryKey: queryKeys.users(search ?? ""),
    queryFn: () => usersService.list(search),
  });
}

/** Инвалидируем список команды: от него зависит уровень доступа на всех экранах. */
function useMemberInvalidation(projectId: string | undefined) {
  const queryClient = useQueryClient();
  return () =>
    queryClient.invalidateQueries({ queryKey: queryKeys.members(projectId ?? "") });
}

export function useAddMember(projectId: string | undefined) {
  const invalidate = useMemberInvalidation(projectId);
  return useMutation({
    mutationFn: (payload: ProjectMemberCreateRequest) =>
      membersService.add(projectId as string, payload),
    onSuccess: invalidate,
  });
}

export function useUpdateMember(projectId: string | undefined) {
  const invalidate = useMemberInvalidation(projectId);
  return useMutation({
    mutationFn: ({
      memberId,
      payload,
    }: {
      memberId: string;
      payload: ProjectMemberUpdateRequest;
    }) => membersService.update(projectId as string, memberId, payload),
    onSuccess: invalidate,
  });
}

export function useRemoveMember(projectId: string | undefined) {
  const invalidate = useMemberInvalidation(projectId);
  return useMutation({
    mutationFn: (memberId: string) => membersService.remove(projectId as string, memberId),
    onSuccess: invalidate,
  });
}

/**
 * Вехи меняют и список вех, и дашборд (`milestonesClosed`, `nearestMilestones`),
 * и данные Ганта — поэтому инвалидируем всё поддерево проектов.
 */
function useMilestoneInvalidation() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: ["projects"] });
}

export function useCreateMilestone(projectId: string | undefined) {
  const invalidate = useMilestoneInvalidation();
  return useMutation({
    mutationFn: (payload: MilestoneCreateRequest) =>
      milestonesService.create(projectId as string, payload),
    onSuccess: invalidate,
  });
}

export function useUpdateMilestone() {
  const invalidate = useMilestoneInvalidation();
  return useMutation({
    mutationFn: ({
      milestoneId,
      payload,
    }: {
      milestoneId: string;
      payload: MilestoneCreateRequest;
    }) => milestonesService.update(milestoneId, payload),
    onSuccess: invalidate,
  });
}

export function useDeleteMilestone() {
  const invalidate = useMilestoneInvalidation();
  return useMutation({
    mutationFn: (milestoneId: string) => milestonesService.remove(milestoneId),
    onSuccess: invalidate,
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
