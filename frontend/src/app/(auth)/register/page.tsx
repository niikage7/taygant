import type { Metadata } from "next";
import Link from "next/link";

import { AuthHeading } from "@/components/auth/auth-heading";
import { RegisterForm } from "@/components/auth/register-form";

export const metadata: Metadata = {
  title: "Регистрация",
  description: "Создание аккаунта в системе управления проектами и диаграммой Ганта",
};

export default function RegisterPage() {
  return (
    <>
      <AuthHeading title="Регистрация" subtitle="Управление проектами и диаграмма Ганта" />
      <RegisterForm />

      <p className="mt-6 text-center text-sm text-ink-muted">
        Уже есть аккаунт?{" "}
        <Link
          href="/login"
          className="rounded-control font-medium text-brand underline underline-offset-2 hover:text-brand-hover focus-visible:focus-ring"
        >
          Войдите
        </Link>
      </p>

      <p className="mt-6 border-t border-line pt-5 text-center text-xs leading-relaxed text-ink-faint">
        Нажимая на кнопку «Зарегистрироваться», вы принимаете{" "}
        {/*
          Открываем в новой вкладке: пользователь читает соглашение посреди
          заполнения формы, и уход со страницы стёр бы уже введённые данные.
        */}
        <Link
          href="/terms"
          target="_blank"
          rel="noopener noreferrer"
          className="rounded-control font-semibold text-brand hover:text-brand-hover focus-visible:focus-ring"
        >
          пользовательское соглашение
        </Link>
      </p>
    </>
  );
}
