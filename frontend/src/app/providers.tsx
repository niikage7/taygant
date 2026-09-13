"use client";

import * as Tooltip from "@radix-ui/react-tooltip";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useState, type ReactNode } from "react";

import { ApiError } from "@/services/http-client";

/**
 * QueryClient создаётся внутри состояния, а не в модуле: при рендере на сервере
 * модульная переменная была бы общей для всех запросов и данные одного
 * пользователя утекли бы другому.
 */
export function Providers({ children }: { children: ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            refetchOnWindowFocus: false,
            retry: (failureCount, error) => {
              // 4xx повторять бессмысленно: это не сбой сети, а отказ по существу
              // (нет прав, не найдено, эндпоинт ещё не реализован).
              if (error instanceof ApiError && error.status < 500) return false;
              return failureCount < 2;
            },
          },
        },
      }),
  );

  // Tooltip.Provider на всё приложение: всплывающие карточки пользователей
  // стоят в таблицах, комментариях и на доске, и заводить провайдер в каждом
  // из этих мест значило бы размножить одинаковую настройку задержки.
  return (
    <QueryClientProvider client={client}>
      <Tooltip.Provider delayDuration={300}>{children}</Tooltip.Provider>
    </QueryClientProvider>
  );
}
