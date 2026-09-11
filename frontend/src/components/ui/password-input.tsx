"use client";

import { Eye, EyeOff, Lock } from "lucide-react";
import { useState } from "react";

import { Input, type InputProps } from "@/components/ui/input";

/** Поле пароля с переключателем видимости — как в макетах входа и регистрации. */
export function PasswordInput({ icon, ...props }: InputProps) {
  const [visible, setVisible] = useState(false);

  return (
    <Input
      type={visible ? "text" : "password"}
      icon={icon ?? <Lock />}
      trailing={
        <button
          type="button"
          onClick={() => setVisible((value) => !value)}
          aria-label={visible ? "Скрыть пароль" : "Показать пароль"}
          aria-pressed={visible}
          className="-mr-1 flex size-6 cursor-pointer items-center justify-center rounded-control text-ink-faint transition-colors hover:text-ink-muted focus-visible:focus-ring"
        >
          {visible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
        </button>
      }
      {...props}
    />
  );
}
