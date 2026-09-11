"use client";

import {
  addDays,
  differenceInCalendarDays,
  eachDayOfInterval,
  format,
  getDay,
  parseISO,
} from "date-fns";
import {
  ArrowRight,
  CalendarCheck2,
  CircleCheck,
  Flag,
  Info,
  SlidersHorizontal,
  Users,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { TeamStep, type DraftMember } from "@/components/project/team-step";
import { WizardStep } from "@/components/project/wizard-step";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardBody } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { DateInput } from "@/components/ui/date-input";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Progress } from "@/components/ui/progress";
import { Textarea } from "@/components/ui/textarea";
import { Alert } from "@/components/ui/alert";
import { queryKeys } from "@/data/queries";
import { useCurrentProject } from "@/data/current-project";
import { toUserMessage } from "@/lib/api-error-message";
import { formatWeekday, plural } from "@/lib/format";
import { projectsService } from "@/services";
import { cn } from "@/lib/utils";
import type { ProjectCreateRequest, User, WorkingCalendarType } from "@/types";

const NAME_MAX_LENGTH = 120;

const toIsoDate = (date: Date) => format(date, "yyyy-MM-dd");
const SECTION = "text-2xs font-semibold tracking-wider text-ink-faint uppercase";

// date-fns не умеет считать рабочие дни по шестидневке (там всегда сб+вс — выходные),
// поэтому дни недели разбираем сами: для "6/1" выходной только воскресенье.
const countWorkingDays = (start: Date, end: Date, calendar: WorkingCalendarType) =>
  eachDayOfInterval({ start, end }).filter((day) => {
    const weekday = getDay(day);
    return calendar === "6/1" ? weekday !== 0 : weekday !== 0 && weekday !== 6;
  }).length;

/**
 * Мастер инициации проекта. Все три шага показаны на одной странице — так же,
 * как в макете: пользователь видит целиком, что настраивает, а шаги сверху
 * работают как оглавление.
 */
export function NewProjectForm({
  users,
  currentUser,
}: {
  /** Пользователи платформы для назначения в команду (GET /users). */
  users: User[];
  currentUser: User | null;
}) {
  // Форма создаёт настоящий проект, поэтому поля пустые: предзаполненный текст
  // из макета попадал бы в базу у тех, кто не заметил его и нажал «Создать».
  // Даты — единственное исключение: срок по умолчанию удобнее готового, чем пустой.
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [startDate, setStartDate] = useState(() => toIsoDate(new Date()));
  const [deadline, setDeadline] = useState(() => toIsoDate(addDays(new Date(), 30)));
  const [customerOrg, setCustomerOrg] = useState("");
  const [calendar, setCalendar] = useState<WorkingCalendarType>("5/2");
  const [autoRecalculate, setAutoRecalculate] = useState(true);
  const [highlightCriticalPath, setHighlightCriticalPath] = useState(true);
  const [members, setMembers] = useState<DraftMember[]>([]);

  const router = useRouter();
  const queryClient = useQueryClient();
  const { selectProject } = useCurrentProject();

  const createProject = useMutation({
    mutationFn: (payload: ProjectCreateRequest) => projectsService.create(payload),
    onSuccess: async (project) => {
      // Созданный проект сразу становится текущим, иначе на экранах остался бы
      // выбранным прежний проект из localStorage.
      selectProject(project.id);
      // Список проектов формирует «текущий проект» для всех внутренних экранов,
      // поэтому его нужно перечитать до перехода на диаграмму.
      await queryClient.invalidateQueries({ queryKey: queryKeys.projects });
      router.push("/gantt");
    },
  });

  const submit = () => {
    if (!requiredFilled) return;
    createProject.mutate({
      name: name.trim(),
      description: description.trim() || undefined,
      customerOrg: customerOrg || undefined,
      startDate,
      deadline,
      workingCalendarType: calendar,
      autoRecalculateDependents: autoRecalculate,
      highlightCriticalPath,
      members: members.map((member) => ({
        userId: member.userId,
        projectRole: member.projectRole,
        accessLevel: member.accessLevel,
      })),
    });
  };

  const start = parseISO(startDate);
  const end = parseISO(deadline);
  const validRange = end > start;
  const calendarDays = validRange ? differenceInCalendarDays(end, start) + 1 : 0;
  const workingDays = validRange ? countWorkingDays(start, end, calendar) : 0;
  const weekendDays = calendarDays - workingDays;

  const requiredFilled = name.trim().length > 0 && validRange;

  return (
    <form
      className="space-y-4"
      onSubmit={(event) => {
        event.preventDefault();
        submit();
      }}
    >
      <div className="min-w-0">
        {/*
          Счётчик шагов убран вместе с меню: все три секции показаны на одной
          странице, и «Шаг 1 из 3» противоречил бы тому, что видит пользователь.
        */}
        <p className="flex items-center gap-2 text-2xs font-semibold tracking-wider text-brand uppercase">
          <span className="size-1.5 rounded-full bg-brand" />
          Мастер инициации
        </p>
        <h1 className="mt-1.5 text-2xl leading-tight font-bold tracking-tight text-ink">
          Создание нового проекта
        </h1>
        <p className="mt-1.5 max-w-xl text-13 leading-relaxed text-ink-muted">
          Задайте базовые параметры, плановые сроки и структуру задач. Конфигурация
          сформирует календарно-сетевой график и WBS-матрицу.
        </p>
      </div>

      <WizardStep
        icon={Info}
        title="Шаг 1. Основная информация"
        subtitle="Паспортные данные и ключевые вехи реализации"
        aside={
          <Badge tone="neutral" className="font-mono">
            ID: PRJ-2025-084
          </Badge>
        }
      >
        <div className="grid gap-4 lg:grid-cols-[1.6fr_1fr]">
          <div>
            <div className="flex items-baseline justify-between gap-3">
              <Label htmlFor="project-name">
                Название проекта <span className="text-danger">*</span>
              </Label>
              <span className="text-xs text-ink-faint">
                Макс. {NAME_MAX_LENGTH} знаков
              </span>
            </div>
            <Input
              id="project-name"
              value={name}
              maxLength={NAME_MAX_LENGTH}
              onChange={(event) => setName(event.target.value)}
              className="mt-1.5"
              required
            />
          </div>

          <div>
            <Label htmlFor="project-customer">
              Заказчик / Подразделение
            </Label>
            <Input
              id="project-customer"
              value={customerOrg}
              onChange={(event) => setCustomerOrg(event.target.value)}
              placeholder="Например, Отдел разработки"
              className="mt-1.5"
            />
          </div>
        </div>

        <div>
          <Label htmlFor="project-description">Описание и бизнес-цель проекта</Label>
          <Textarea
            id="project-description"
            rows={3}
            value={description}
            onChange={(event) => setDescription(event.target.value)}
            className="mt-1.5"
          />
        </div>

        <div className="grid gap-4 rounded-control bg-surface-subtle p-4 lg:grid-cols-[1fr_auto_1fr_1.2fr]">
          <div>
            <Label htmlFor="project-start" className={SECTION}>
              Дата старта графика
            </Label>
            <DateInput
              id="project-start"
              value={startDate}
              onChange={setStartDate}
              className="mt-1.5 bg-surface"
              weekdayHint={formatWeekday(startDate)}
            />
          </div>

          <span className="hidden items-start pt-9 text-ink-faint lg:flex">
            <ArrowRight className="size-4" />
          </span>

          <div>
            <Label htmlFor="project-deadline" className={SECTION}>
              Плановый дедлайн
            </Label>
            <DateInput
              id="project-deadline"
              value={deadline}
              onChange={setDeadline}
              className="mt-1.5 bg-surface"
              invalid={!validRange}
              weekdayHint={validRange ? formatWeekday(deadline) : undefined}
              minDate={startDate}
            />
          </div>

          <div className="rounded-control bg-surface p-3">
            <p className={SECTION}>Доступный бюджет времени</p>
            {validRange ? (
              <>
                <p className="mt-1 flex items-baseline gap-1.5">
                  <span className="text-2xl font-bold text-brand">{calendarDays}</span>
                  <span className="text-13 text-ink-muted">
                    {plural(calendarDays, ["календарный день", "календарных дня", "календарных дней"])}
                  </span>
                </p>
                <p className="mt-2 flex items-baseline justify-between gap-2 text-xs">
                  <span className="font-mono font-semibold text-success">
                    {workingDays} раб. дн.
                  </span>
                  <span className="text-ink-faint">/ {weekendDays} вых.</span>
                </p>
                <Progress
                  value={(workingDays / calendarDays) * 100}
                  className="mt-2 h-1.5"
                  label="Доля рабочих дней"
                />
              </>
            ) : (
              <p className="mt-2 text-13 text-danger">
                Дедлайн должен быть позже даты старта
              </p>
            )}
          </div>
        </div>
      </WizardStep>

      <WizardStep
        icon={Users}
        title="Шаг 2. Команда проекта и роли"
        subtitle="Назначение ответственных и распределение прав управления"
      >
        <TeamStep
          users={users}
          currentUser={currentUser}
          members={members}
          onChange={setMembers}
        />
      </WizardStep>

      <WizardStep
        icon={SlidersHorizontal}
        title="Шаг 3. Начальные параметры диаграммы Ганта"
        subtitle="Вычислительное ядро расписания и логика сдвига дат"
        aside={<span className="font-mono text-xs text-ink-faint">CPM Engine v4.2</span>}
      >
        <div className="grid gap-3 lg:grid-cols-2">
          <ToggleCard
            checked={autoRecalculate}
            onChange={setAutoRecalculate}
            title="Включить автоматический пересчёт зависимых задач при сдвиге дат"
            description="При переносе дедлайна предшественника все зависимые вехи цепочки Finish-to-Start каскадно сдвигаются вперёд с сохранением буфера."
          />
          <ToggleCard
            checked={highlightCriticalPath}
            onChange={setHighlightCriticalPath}
            title="Подсвечивать критический путь на таймлайне"
            titleAccent
            description="Задачи с нулевым резервом времени окрашиваются в контрастный бордово-красный цвет с жирными соединительными векторами."
          />
        </div>

        <div className="flex flex-wrap items-center justify-between gap-4 rounded-control bg-brand-tint p-4">
          <div className="flex min-w-0 items-start gap-3">
            <CalendarCheck2 className="mt-0.5 size-5 shrink-0 text-brand" />
            <div className="min-w-0">
              <p className="text-13 font-semibold text-ink">Рабочий календарь проекта</p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {(["5/2", "6/1"] as const).map((option) => (
              <button
                key={option}
                type="button"
                onClick={() => setCalendar(option)}
                aria-pressed={calendar === option}
                className={cn(
                  "rounded-control px-3 py-2 text-13 font-medium transition-colors focus-visible:focus-ring",
                  calendar === option
                    ? "bg-brand text-white"
                    : "border border-line bg-surface text-ink-muted hover:bg-surface-muted",
                )}
              >
                {option === "5/2" ? "Стандарт: 5/2 (40 ч/нед)" : "Шестидневка 6/1"}
              </button>
            ))}
          </div>
        </div>
      </WizardStep>

      {createProject.isError ? (
        <Alert tone="danger">
          {toUserMessage(
            createProject.error,
            { 403: "Недостаточно прав для создания проекта" },
            "Не удалось создать проект. Попробуйте ещё раз",
          )}
        </Alert>
      ) : null}

      <Card>
        <CardBody className="flex flex-wrap items-center justify-between gap-4">
          <p
            className={cn(
              "flex items-start gap-2 text-13",
              requiredFilled ? "text-ink-muted" : "text-danger",
            )}
          >
            <CircleCheck
              className={cn(
                "mt-px size-4 shrink-0",
                requiredFilled ? "text-success" : "text-danger",
              )}
            />
            {requiredFilled
              ? "Все обязательные поля заполнены. Сформирована структура из 5 этапов."
              : "Заполните название проекта и корректные сроки."}
          </p>

          <div className="flex flex-wrap items-center gap-2">
            <Button type="submit" disabled={!requiredFilled || createProject.isPending}>
              <Flag />
              {createProject.isPending
                ? "Создаём проект…"
                : "Создать проект и перейти к диаграмме Ганта"}
              <ArrowRight />
            </Button>
          </div>
        </CardBody>
      </Card>
    </form>
  );
}

function ToggleCard({
  checked,
  onChange,
  title,
  description,
  titleAccent,
}: {
  checked: boolean;
  onChange: (value: boolean) => void;
  title: string;
  description: string;
  titleAccent?: boolean;
}) {
  return (
    <label className="flex cursor-pointer items-start gap-3 rounded-control bg-surface-subtle p-4 transition-colors hover:bg-surface-muted">
      <Checkbox
        checked={checked}
        onCheckedChange={(value) => onChange(value === true)}
        className="mt-0.5"
      />
      <span className="min-w-0">
        <span className="block text-13 leading-snug font-semibold text-ink">
          {title}
          {titleAccent ? <span className="ml-1.5 text-danger">●</span> : null}
        </span>
        <span className="mt-1.5 block text-xs leading-relaxed text-ink-muted">
          {description}
        </span>
      </span>
    </label>
  );
}
