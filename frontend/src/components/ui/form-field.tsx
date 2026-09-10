import type { ReactNode } from "react";

import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

type FormFieldProps = {
  /** id управляемого поля — связывает подпись, ошибку и input. */
  id: string;
  label: string;
  /** Текст ошибки; при наличии подписывается под полем и озвучивается скринридером. */
  error?: string;
  /** Подсказка справа от подписи (например, «Макс. 120 знаков»). */
  hint?: ReactNode;
  className?: string;
  children: ReactNode;
};

/** Обёртка «подпись + поле + ошибка» с единым вертикальным ритмом макета. */
export function FormField({ id, label, error, hint, className, children }: FormFieldProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      <div className="flex items-baseline justify-between gap-3">
        <Label htmlFor={id}>{label}</Label>
        {hint ? <span className="text-xs text-ink-faint">{hint}</span> : null}
      </div>
      {children}
      {error ? (
        <p id={`${id}-error`} role="alert" className="text-xs text-danger">
          {error}
        </p>
      ) : null}
    </div>
  );
}
