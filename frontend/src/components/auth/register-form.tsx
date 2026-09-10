"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, AtSign, CircleUserRound } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { FormField } from "@/components/ui/form-field";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { toUserMessage } from "@/lib/api-error-message";
import { startSession } from "@/lib/session";
import { authService } from "@/services";

const registerSchema = z.object({
  email: z.email("Введите корректный адрес почты"),
  fullName: z.string().trim().min(2, "Укажите имя и фамилию").max(120, "Не более 120 знаков"),
  password: z.string().min(8, "Пароль должен быть не короче 8 знаков"),
});

type RegisterValues = z.infer<typeof registerSchema>;

export function RegisterForm() {
  const router = useRouter();
  const [formError, setFormError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: { email: "", fullName: "", password: "" },
  });

  async function onSubmit(values: RegisterValues) {
    setFormError(null);
    try {
      const session = await authService.register(values);
      startSession(session, true);
      router.push("/");
      router.refresh();
    } catch (error) {
      setFormError(
        toUserMessage(
          error,
          {
            409: "Пользователь с такой почтой уже зарегистрирован",
            404: "Регистрация пока не подключена на сервере",
            501: "Регистрация пока не подключена на сервере",
          },
          "Не удалось зарегистрироваться. Попробуйте ещё раз",
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

      <FormField id="fullName" label="Имя" error={errors.fullName?.message}>
        <Input
          id="fullName"
          autoComplete="name"
          placeholder="Иванов Иван"
          icon={<CircleUserRound />}
          invalid={Boolean(errors.fullName)}
          aria-describedby={errors.fullName ? "fullName-error" : undefined}
          {...register("fullName")}
        />
      </FormField>

      <FormField id="password" label="Пароль" error={errors.password?.message}>
        <PasswordInput
          id="password"
          autoComplete="new-password"
          placeholder="••••••••"
          invalid={Boolean(errors.password)}
          aria-describedby={errors.password ? "password-error" : undefined}
          {...register("password")}
        />
      </FormField>

      <Button type="submit" size="lg" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? "Создаём аккаунт…" : "Зарегистрироваться"}
        {isSubmitting ? null : <ArrowRight />}
      </Button>
    </form>
  );
}
