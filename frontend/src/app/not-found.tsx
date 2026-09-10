import { ArrowLeft, Compass } from "lucide-react";
import type { Metadata } from "next";
import Link from "next/link";

import { LogoMark } from "@/components/brand/logo";
import { Button } from "@/components/ui/button";

export const metadata: Metadata = {
  title: "Страница не найдена",
};

/** Общий экран 404: ловит и опечатки в адресе, и ссылки на удалённые задачи. */
export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-1 flex-col items-center justify-center bg-page px-4 py-12 text-center">
      <LogoMark className="size-10" />
      <p className="mt-6 flex items-center gap-2 font-mono text-sm text-ink-faint">
        <Compass className="size-4" />
        404
      </p>
      <h1 className="mt-2 text-2xl font-bold tracking-tight text-ink">
        Страница не найдена
      </h1>
      <p className="mt-2 max-w-sm text-[13px] leading-relaxed text-ink-muted">
        Адрес не существует или объект был удалён. Проверьте ссылку или вернитесь
        к обзору проекта.
      </p>
      <Button asChild className="mt-6">
        <Link href="/overview">
          <ArrowLeft />
          К обзору проекта
        </Link>
      </Button>
    </div>
  );
}
