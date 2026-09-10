import type { Metadata } from "next";
import Link from "next/link";

import { AuthHeading } from "@/components/auth/auth-heading";
import { LoginForm } from "@/components/auth/login-form";

export const metadata: Metadata = {
  title: "Вход в систему",
  description: "Вход в систему управления проектами и диаграммой Ганта",
};

export default function LoginPage() {
  return (
    <>
      <AuthHeading title="Вход в систему" subtitle="Управление проектами и диаграмма Ганта" />
      <LoginForm />
      <p className="mt-6 text-center text-sm text-ink-muted">
        Нет аккаунта?{" "}
        <Link
          href="/register"
          className="rounded-control font-medium text-brand underline underline-offset-2 hover:text-brand-hover focus-visible:focus-ring"
        >
          Зарегистрируйтесь
        </Link>
      </p>
    </>
  );
}
