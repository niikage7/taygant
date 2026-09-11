"use client";

import { ArrowLeft, ArrowRight, X } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { TOUR_STEPS, markTourCompleted } from "@/lib/tour";
import { cn } from "@/lib/utils";

const CARD_WIDTH = 320;
const GAP = 12;

type Rect = { top: number; left: number; width: number; height: number };

/**
 * Сайдбар рендерит одно и то же содержимое дважды — в постоянной панели
 * (скрыта ниже `lg`) и в выезжающей панели на мобильных, — поэтому в DOM
 * может быть два элемента с одним `data-tour`. `querySelector` вернул бы
 * первый по порядку документа, а он как раз может быть скрытым; берём первый
 * действительно отрисованный (getClientRects().length > 0).
 */
function findVisibleTourTarget(target: string): HTMLElement | null {
  const candidates = document.querySelectorAll<HTMLElement>(`[data-tour="${target}"]`);
  for (const candidate of candidates) {
    if (candidate.getClientRects().length > 0) return candidate;
  }
  return null;
}

/**
 * Обучающий тур: подсвечивает элементы по атрибуту `data-tour` и объясняет их.
 *
 * Цель ищется в DOM, а не через ref: шаги живут на разных маршрутах, и передавать
 * ссылки через полприложения пришлось бы вручную через все промежуточные
 * компоненты. Если элемент не найден (экран ещё грузится или блок скрыт при
 * пустых данных), шаг пропускается — тур не должен упираться в пустоту.
 */
export function ProductTour({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const pathname = usePathname();
  const [index, setIndex] = useState(0);
  const [rect, setRect] = useState<Rect | null>(null);

  const step = TOUR_STEPS[index];

  const finish = useCallback(() => {
    markTourCompleted();
    onClose();
  }, [onClose]);

  // Шаг может жить на другом экране — переходим туда до поиска цели.
  useEffect(() => {
    if (step && pathname !== step.href) router.push(step.href);
  }, [step, pathname, router]);

  useEffect(() => {
    if (!step || pathname !== step.href) return;

    let cancelled = false;
    let attempts = 0;

    const locate = () => {
      if (cancelled) return;
      // Ищем видимый экземпляр: например, «nav» и «project-switcher» на узком
      // экране лежат в закрытой выезжающей панели (display: none), пока её
      // не открыли, — тур должен пропустить шаг, а не подсветить невидимый блок.
      const element = findVisibleTourTarget(step.target);

      if (!element) {
        // Экран мог ещё не отрисоваться: ждём около двух секунд, потом
        // пропускаем шаг, чтобы тур не завис.
        if (attempts++ < 20) {
          window.setTimeout(locate, 100);
          return;
        }
        setIndex((current) =>
          current + 1 < TOUR_STEPS.length ? current + 1 : current,
        );
        return;
      }

      element.scrollIntoView({ block: "center", behavior: "smooth" });
      const box = element.getBoundingClientRect();
      setRect({
        top: box.top,
        left: box.left,
        width: box.width,
        height: box.height,
      });
    };

    locate();
    return () => {
      cancelled = true;
    };
  }, [step, pathname]);

  // Пересчитываем при прокрутке и изменении размера: подсветка привязана к
  // экранным координатам и иначе «отклеится» от элемента.
  useEffect(() => {
    if (!step) return;
    const update = () => {
      const element = findVisibleTourTarget(step.target);
      if (!element) return;
      const box = element.getBoundingClientRect();
      setRect({ top: box.top, left: box.left, width: box.width, height: box.height });
    };
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    return () => {
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
    };
  }, [step]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") finish();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [finish]);

  const cardRef = useRef<HTMLDivElement>(null);
  const hasRect = rect !== null;
  // Фокус — на карточку тура при каждом новом шаге (а не только при монтировании):
  // без этого читалка экрана озвучивала бы то, что было в фокусе до тура, а
  // клавиатурный пользователь не понимал бы, что на экране появился диалог.
  useEffect(() => {
    cardRef.current?.focus();
  }, [index, hasRect]);

  if (!step || !rect) return null;

  const isLast = index === TOUR_STEPS.length - 1;
  const below = rect.top + rect.height + GAP + 180 < window.innerHeight;
  const cardTop = below ? rect.top + rect.height + GAP : rect.top - GAP;
  const cardLeft = Math.min(
    Math.max(rect.left, GAP),
    Math.max(window.innerWidth - CARD_WIDTH - GAP, GAP),
  );

  return (
    <div className="fixed inset-0 z-[100]">
      {/* Затемнение с «окном» вокруг цели — одной тенью, без четырёх блоков. */}
      <div
        aria-hidden
        className="pointer-events-none absolute rounded-control ring-2 ring-brand transition-all duration-200 motion-reduce:transition-none"
        style={{
          top: rect.top - 4,
          left: rect.left - 4,
          width: rect.width + 8,
          height: rect.height + 8,
          boxShadow: "0 0 0 9999px rgb(15 23 42 / 0.55)",
        }}
      />

      <div
        ref={cardRef}
        role="dialog"
        aria-modal="true"
        aria-label={`Обучающий тур: ${step.title}`}
        tabIndex={-1}
        className={cn(
          "absolute w-80 rounded-card bg-surface p-4 shadow-popover outline-none",
          !below && "-translate-y-full",
        )}
        style={{ top: cardTop, left: cardLeft }}
      >
        <div className="flex items-start justify-between gap-3">
          <p className="text-2xs font-semibold tracking-wider text-brand uppercase">
            Шаг {index + 1} из {TOUR_STEPS.length}
          </p>
          <button
            type="button"
            onClick={finish}
            aria-label="Закрыть тур"
            className="cursor-pointer rounded-control text-ink-faint hover:text-ink focus-visible:focus-ring"
          >
            <X className="size-4" aria-hidden />
          </button>
        </div>

        <h2 className="mt-2 text-15 font-semibold text-ink">{step.title}</h2>
        <p className="mt-1.5 text-13 leading-relaxed text-ink-muted">{step.text}</p>

        <div className="mt-4 flex items-center justify-between gap-2">
          <button
            type="button"
            onClick={finish}
            className="cursor-pointer rounded-control text-13 text-ink-faint hover:text-ink focus-visible:focus-ring"
          >
            Пропустить
          </button>

          <span className="flex items-center gap-2">
            {index > 0 ? (
              <Button
                variant="secondary"
                size="sm"
                onClick={() => setIndex((current) => current - 1)}
              >
                <ArrowLeft aria-hidden />
                Назад
              </Button>
            ) : null}
            <Button
              size="sm"
              onClick={() => (isLast ? finish() : setIndex((current) => current + 1))}
            >
              {isLast ? "Готово" : "Далее"}
              {isLast ? null : <ArrowRight aria-hidden />}
            </Button>
          </span>
        </div>
      </div>
    </div>
  );
}
