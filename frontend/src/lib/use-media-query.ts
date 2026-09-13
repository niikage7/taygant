"use client";

import { useSyncExternalStore } from "react";

/**
 * Совпадает ли медиазапрос — для раскладок, которые нельзя выразить одним CSS:
 * например, ширина реестра Ганта нужна и таблице, и расчёту прокрутки таймлайна.
 *
 * useSyncExternalStore, а не useState+useEffect: на сервере и при гидратации
 * отдаётся `serverValue` (разметка совпадает), а настоящее значение подставляется
 * сразу после неё и обновляется при повороте экрана или смене ширины окна.
 */
export function useMediaQuery(query: string, serverValue = false): boolean {
  return useSyncExternalStore(
    (onChange) => {
      const list = window.matchMedia(query);
      list.addEventListener("change", onChange);
      return () => list.removeEventListener("change", onChange);
    },
    () => window.matchMedia(query).matches,
    () => serverValue,
  );
}
