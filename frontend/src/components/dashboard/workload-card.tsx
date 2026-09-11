/**
 * Загрузка команды текущего спринта.
 *
 * Данные приходят из `GET /projects/{id}/workload` (по умолчанию — текущий спринт).
 */
import { Card, CardBody, CardHeader, CardTitle } from "@/components/ui/card";
import { Avatar } from "@/components/ui/avatar";
import { UserHoverCard } from "@/components/user/user-card";
import { Progress } from "@/components/ui/progress";
import { cn } from "@/lib/utils";
import type { MemberWorkload } from "@/types";

/** Загрузка команды текущего спринта: перегрузка подсвечивается красным. */
export function WorkloadCard({ workload }: { workload: MemberWorkload[] }) {
  return (
    <Card className="flex flex-col">
      <CardHeader className="items-start">
        <div>
          <CardTitle>Нагрузка команды</CardTitle>
          <p className="mt-0.5 text-13 text-ink-muted">Текущий спринт</p>
        </div>
        {workload.length > 0 ? (
          <span className="text-xs text-ink-faint">
            Лимиты — по участникам
          </span>
        ) : null}
      </CardHeader>
      <CardBody className="flex-1 space-y-2">
        {workload.map((item) => {
          const overloaded = item.utilizationPercent > 100;
          const assigned = Math.round(item.assignedHoursPerWeek);
          const free = item.weeklyHoursLimit - assigned;
          return (
            <div
              key={item.user.id}
              className="rounded-control bg-surface-subtle p-3"
            >
              <div className="flex items-start justify-between gap-3">
                <UserHoverCard user={item.user}>
                  <div className="flex min-w-0 items-center gap-2">
                    <Avatar fullName={item.user.fullName} className="size-8 text-xs" />
                    <div className="min-w-0">
                      <p className="truncate text-13 font-semibold text-ink">
                        {item.user.fullName}
                      </p>
                      <p className="truncate text-xs text-ink-faint">
                        {item.user.position}
                      </p>
                    </div>
                  </div>
                </UserHoverCard>
                <div className="shrink-0 text-right">
                  <p
                    className={cn(
                      "text-13 font-bold",
                      overloaded ? "text-danger" : "text-success",
                    )}
                  >
                    {item.utilizationPercent}%
                  </p>
                  <p className="text-xs text-ink-faint">
                    {assigned} ч / нед
                    {!overloaded && free > 0 ? " (есть слот)" : ""}
                  </p>
                </div>
              </div>
              <Progress
                value={item.utilizationPercent}
                className="mt-2.5 h-1.5"
                barClassName={cn(
                  overloaded && "bg-danger",
                  !overloaded && item.utilizationPercent >= 70 && "bg-brand",
                  !overloaded && item.utilizationPercent < 70 && "bg-success",
                )}
                label={`Загрузка ${item.user.fullName}`}
              />
            </div>
          );
        })}
      </CardBody>
    </Card>
  );
}
