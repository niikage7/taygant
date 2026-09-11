import { clsx, type ClassValue } from "clsx";
import { extendTailwindMerge } from "tailwind-merge";

/**
 * tailwind-merge не читает globals.css и не знает наших размеров шрифта.
 * Без этой настройки `text-13` он принимает за цвет текста и выкидывает стоящий
 * рядом `text-white` — у Button size="sm" текст становился тёмным на синем.
 * Список должен совпадать с --text-* в globals.css.
 */
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      text: ["3xs", "2xs", "13", "15", "26"],
    },
  },
});

/** Склеивает классы и разрешает конфликты Tailwind-утилит (последний выигрывает). */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
