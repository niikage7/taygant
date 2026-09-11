"use client";

import { ArrowLeft, ArrowRight, X } from "lucide-react";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import { TOUR_STEPS, markTourCompleted, placeTourCard, type TourRect } from "@/lib/tour";

const CARD_WIDTH = 320;
const GAP = 12;
/** Высота карточки до первого замера — примерно карточка с двумя строками текста. */
const CARD_HEIGHT_ESTIMATE = 200;
/** Сколько раз по 100 мс ждём появления цели, прежде чем пропустить шаг (≈5 с). */
const LOCATE_ATTEMPTS = 50;

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

function measure(element: HTMLElement): TourRect {
  const box = element.getBoundingClientRect();
  return { top: box.top, left: box.left, width: box.width, height: box.height };
}

/**
 * Обучающий тур: подсвечивает элементы по атрибуту `data-tour` и объясняет их.
 *
 * Цель ищется в DOM, а не через ref: шаги живут на разных маршрутах, и передавать
 * ссылки через полприложения пришлось бы вручную через все промежуточные
 * компоненты. Если элемент не найден (экран ещё грузится или блок скрыт при
 * пустых данных или без полного доступа), шаг пропускается в сторону, куда
 * листал пользователь, — тур не должен упираться в пустоту.
 */
export function ProductTour({ onClose }: { onClose: () => void }) {
  const router = useRouter();
  const pathname = usePathname();
  const [index, setIndex] = useState(0);
  // Подсветка помнит, к какому шагу относится. Раньше хранился голый rect:
  // пока цель следующего шага искалась (или не находилась вовсе), на экране
  // оставалась рамка предыдущей цели — и все дальнейшие шаги «указывали» на
  // переключатель проектов.
  const [located, setLocated] = useState<{ index: number; rect: TourRect } | null>(null);
  const [cardHeight, setCardHeight] = useState(CARD_HEIGHT_ESTIMATE);
  // Направление листания: пропуская шаг без цели, идём туда же, куда шёл
  // пользователь. Иначе «Назад» на пропущенный шаг тут же возвращал бы вперёд.
  const directionRef = useRef<1 | -1>(1);
  const cardRef = useRef<HTMLDivElement>(null);

  const step = TOUR_STEPS[index];

  const finish = useCallback(() => {
    markTourCompleted();
    onClose();
  }, [onClose]);

  const go = (delta: 1 | -1) => {
    directionRef.current = delta;
    setIndex((current) => current + delta);
  };

  // Шаг может жить на другом экране — переходим туда до поиска цели.
  useEffect(() => {
    if (step && pathname !== step.href) router.push(step.href);
  }, [step, pathname, router]);

  useEffect(() => {
    if (!step || pathname !== step.href) return;

    let cancelled = false;
    let attempts = 0;

    const skip = () => {
      const next = index + directionRef.current;
      if (next >= TOUR_STEPS.length) {
        finish();
      } else if (next < 0) {
        directionRef.current = 1;
        setIndex(index + 1);
      } else {
        setIndex(next);
      }
    };

    const locate = () => {
      if (cancelled) return;
      // Ищем видимый экземпляр: например, «nav» и «project-switcher» на узком
      // экране лежат в закрытой выезжающей панели (display: none), пока её
      // не открыли, — тур должен пропустить шаг, а не подсветить невидимый блок.
      const element = findVisibleTourTarget(step.target);

      if (!element) {
        // Экран мог ещё не загрузить данные: ждём, потом пропускаем шаг.
        if (attempts++ < LOCATE_ATTEMPTS) {
          window.setTimeout(locate, 100);
          return;
        }
        skip();
        return;
      }

      element.scrollIntoView({ block: "center", behavior: "smooth" });
      setLocated({ index, rect: measure(element) });
    };

    locate();
    return () => {
      cancelled = true;
    };
  }, [step, index, pathname, finish]);

  // Пересчитываем при прокрутке, изменении размера окна и сдвигах раскладки:
  // подсветка привязана к экранным координатам и иначе «отклеится» от элемента.
  // Сдвиги ловим ResizeObserver'ом — ни scroll, ни resize их не порождают, а
  // они случаются постоянно: например, навигация сайдбара (flex-1) сжимается,
  // когда под ней догружается карточка проекта, и рамка оставалась выше цели.
  useEffect(() => {
    if (!step) return;
    const update = () => {
      const element = findVisibleTourTarget(step.target);
      if (element) setLocated({ index, rect: measure(element) });
    };
    const observer = new ResizeObserver(update);
    const element = findVisibleTourTarget(step.target);
    if (element) observer.observe(element);
    // Тело страницы меняет размер, когда догружается контент над целью, —
    // тогда цель сдвигается, не меняя собственного размера.
    observer.observe(document.body);
    window.addEventListener("scroll", update, true);
    window.addEventListener("resize", update);
    return () => {
      observer.disconnect();
      window.removeEventListener("scroll", update, true);
      window.removeEventListener("resize", update);
    };
  }, [step, index, located?.index]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") finish();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [finish]);

  // Фокус — на карточку тура при каждом новом шаге (а не только при монтировании):
  // без этого читалка экрана озвучивала бы то, что было в фокусе до тура, а
  // клавиатурный пользователь не понимал бы, что на экране появился диалог.
  useEffect(() => {
    cardRef.current?.focus();
  }, [index]);

  // Высота карточки зависит от текста шага — меряем настоящую, иначе расчёт
  // «влезает ли снизу» ошибается на длинных описаниях.
  useEffect(() => {
    const card = cardRef.current;
    if (!card) return;
    const observer = new ResizeObserver(() => setCardHeight(card.offsetHeight));
    observer.observe(card);
    return () => observer.disconnect();
  }, []);

  if (!step) return null;

  const rect = located?.index === index ? located.rect : null;
  const isLast = index === TOUR_STEPS.length - 1;
  const viewport = { width: window.innerWidth, height: window.innerHeight };
  // Пока цель шага ищется, карточка стоит по центру поверх затемнения — без
  // рамки, чтобы не подсвечивать чужой элемент.
  const cardPosition = rect
    ? placeTourCard(rect, { width: CARD_WIDTH, height: cardHeight }, viewport, GAP)
    : {
        top: Math.max(GAP, (viewport.height - cardHeight) / 2),
        left: Math.max(GAP, (viewport.width - CARD_WIDTH) / 2),
      };

  return (
    <div className="fixed inset-0 z-[100]">
      {/* Разные key у рамки и сплошного затемнения: без них React переиспользует
          тот же div, и переход анимирует тёмный фон inset-0 в рамку — на экране
          мелькает сжимающийся тёмный прямоугольник. Анимируем только геометрию,
          чтобы рамка плавно переезжала между целями соседних шагов. */}
      {rect ? (
        // Затемнение с «окном» вокруг цели — одной тенью, без четырёх блоков.
        <div
          key="spotlight"
          aria-hidden
          className="pointer-events-none absolute rounded-control ring-2 ring-brand transition-[top,left,width,height] duration-200 motion-reduce:transition-none"
          style={{
            top: rect.top - 4,
            left: rect.left - 4,
            width: rect.width + 8,
            height: rect.height + 8,
            boxShadow: "0 0 0 9999px rgb(15 23 42 / 0.55)",
          }}
        />
      ) : (
        <div key="dim" aria-hidden className="absolute inset-0 bg-[rgb(15_23_42/0.55)]" />
      )}

      <div
        ref={cardRef}
        role="dialog"
        aria-modal="true"
        aria-label={`Обучающий тур: ${step.title}`}
        tabIndex={-1}
        className="absolute w-80 rounded-card bg-surface p-4 shadow-popover outline-none"
        style={cardPosition}
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
              <Button variant="secondary" size="sm" onClick={() => go(-1)}>
                <ArrowLeft aria-hidden />
                Назад
              </Button>
            ) : null}
            <Button size="sm" onClick={() => (isLast ? finish() : go(1))}>
              {isLast ? "Готово" : "Далее"}
              {isLast ? null : <ArrowRight aria-hidden />}
            </Button>
          </span>
        </div>
      </div>
    </div>
  );
}
