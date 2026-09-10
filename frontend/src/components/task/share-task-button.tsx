"use client";

import { Check, Share2 } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";

/**
 * Копирует ссылку на задачу в буфер обмена.
 *
 * `navigator.clipboard` есть не везде (нужен https или localhost), поэтому при
 * отказе показываем сам адрес и выделяем его — пользователь скопирует вручную,
 * а не останется без результата.
 */
export function ShareTaskButton() {
  const [state, setState] = useState<"idle" | "copied" | "failed">("idle");
  const [url, setUrl] = useState("");

  useEffect(() => {
    if (state === "idle") return;
    const timer = setTimeout(() => setState("idle"), 3000);
    return () => clearTimeout(timer);
  }, [state]);

  async function share() {
    const link = window.location.href;
    setUrl(link);
    try {
      await navigator.clipboard.writeText(link);
      setState("copied");
    } catch {
      setState("failed");
    }
  }

  return (
    <span className="relative">
      <Button type="button" variant="secondary" onClick={share}>
        {state === "copied" ? <Check /> : <Share2 />}
        {state === "copied" ? "Ссылка скопирована" : "Поделиться"}
      </Button>

      {state === "failed" ? (
        <span
          role="status"
          className="absolute top-full right-0 z-10 mt-2 w-80 rounded-control bg-surface p-3 text-xs shadow-popover"
        >
          <span className="block text-ink-muted">
            Браузер не дал доступ к буферу — скопируйте ссылку вручную:
          </span>
          <input
            readOnly
            value={url}
            onFocus={(event) => event.target.select()}
            className="mt-2 w-full rounded-control bg-surface-muted px-2 py-1 font-mono text-[11px] text-ink outline-none"
          />
        </span>
      ) : null}
    </span>
  );
}
