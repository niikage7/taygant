"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, AtSign } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PasswordInput } from "@/components/ui/password-input";
import { toUserMessage } from "@/lib/api-error-message";
import { startSession } from "@/lib/session";
import { authService } from "@/services";

const loginSchema = z.object({
  email: z.email("Введите корректный адрес почты"),
  password: z.string().min(1, "Введите пароль"),
  remember: z.boolean(),
});

type LoginValues = z.infer<typeof loginSchema>;

export function LoginForm() {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { email: "", password: "", remember: true },
  });

  const remember = watch("remember");

  async function onSubmit(values: LoginValues) {
    setFormError(null);
    try {
      const session = await authService.login({
        email: values.email,
        password: values.password,
      });
      startSession(session, values.remember);
      router.push("/overview");
      router.refresh();
    } catch (error) {
      setFormError(
        toUserMessage(
          error,
          { 401: "Неверная почта или пароль" },
          "Не удалось войти. Попробуйте ещё раз",
        ),
      );
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="mt-7 space-y-4" noValidate>
      {formError ? <Alert tone="danger">{formError}</Alert> : null}

      <FormField id="email" label="Почта" error={errors.email?.message}>
        <Input
          id="email"
          type="email"
          autoComplete="email"
          placeholder="username@domain.com"
          icon={<AtSign />}
          invalid={Boolean(errors.email)}
          aria-describedby={errors.email ? "email-error" : undefined}
          {...register("email")}
        />
      </FormField>

      <FormField id="password" label="Пароль" error={errors.password?.message}>
        <PasswordInput
          id="password"
          autoComplete="current-password"
          placeholder="••••••••"
          invalid={Boolean(errors.password)}
          aria-describedby={errors.password ? "password-error" : undefined}
          {...register("password")}
        />
      </FormField>

      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Checkbox
            id="remember"
            checked={remember}
            onCheckedChange={(checked) =>
              setValue("remember", checked === true, { shouldDirty: true })
            }
          />
          <Label htmlFor="remember" className="font-normal text-ink-muted">
            Запомнить меня
          </Label>
        </div>
        <Link
          href="/forgot-password"
          className="rounded-control text-[13px] font-medium text-brand hover:text-brand-hover focus-visible:focus-ring"
        >
          Забыли пароль?
        </Link>
      </div>

      <Button type="submit" size="lg" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "Входим…" : "Войти в систему"}
        {isSubmitting ? null : <ArrowRight />}
      </Button>
    </form>
  );
}
