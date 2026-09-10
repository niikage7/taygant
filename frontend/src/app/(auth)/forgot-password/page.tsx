import { ArrowLeft, Mail } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";

import { AuthHeading } from "@/components/auth/auth-heading";
import { Alert } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";

export const metadata: Metadata = {
  title: "Восстановление пароля",
  description: "Восстановление доступа к системе управления проектами",
};

/**
 * Заглушка восстановления пароля.
 *
 * Ссылка «Забыли пароль?» есть в макете входа, но сброса пароля нет ни в
 * `api-spec.yml`, ни в хендлерах — есть только login, register и refresh.
 * Страница нужна, чтобы ссылка не вела в 404 и человек понимал, что делать.
 */
export default function ForgotPasswordPage() {
  return (
    <>
      <AuthHeading
        title="Восстановление пароля"
        subtitle="Управление проектами и диаграмма Ганта"
      />

      <div className="mt-7 space-y-4">
        <Alert tone="info">
          Автоматический сброс пароля пока не подключён. Чтобы восстановить
          доступ, напишите администратору системы — он назначит новый пароль.
        </Alert>

        <a
          href="mailto:niikage7@gmail.com?subject=Восстановление%20доступа%20к%20taygant"
          className="flex items-center gap-2.5 rounded-control bg-surface-muted px-3 py-2.5 text-[13px] text-ink transition-colors hover:bg-brand-tint hover:text-brand focus-visible:focus-ring"
        >
          <Mail className="size-4 shrink-0 text-ink-faint" />
          Написать администратору
        </a>

        <Button asChild variant="secondary" size="lg" className="w-full">
          <Link href="/login">
            <ArrowLeft />
            Вернуться ко входу
          </Link>
        </Button>
      </div>
    </>
  );
}
