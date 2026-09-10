import type { Metadata } from "next";
import Link from "next/link";
import { readFile } from "node:fs/promises";
import path from "node:path";

import { LogoMark } from "@/components/brand/logo";
import { Markdown } from "@/components/content/markdown";

export const metadata: Metadata = {
  title: "Пользовательское соглашение",
  description: "Условия использования сервиса Taygant",
};

/**
 * Текст соглашения лежит в `src/content/terms.md` — править нужно его, вёрстку
 * трогать не придётся. Страница статическая, поэтому файл читается один раз
 * во время сборки, а не на каждый запрос.
 */
async function readTerms(): Promise<string> {
  const file = path.join(process.cwd(), "src/content/terms.md");
  return readFile(file, "utf8");
}

export default async function TermsPage() {
  const terms = await readTerms();

  return (
    <div className="flex flex-1 flex-col bg-page">
      <header className="border-b border-line bg-surface">
        <div className="mx-auto flex w-full max-w-3xl items-center justify-between gap-4 px-6 py-4">
          <Link
            href="/login"
            className="flex items-center gap-2.5 rounded-control focus-visible:focus-ring"
          >
            <LogoMark className="size-8" />
            <span className="text-sm font-bold text-ink">ПроектКонтроль</span>
          </Link>
          <Link
            href="/register"
            className="rounded-control text-[13px] font-medium text-brand hover:text-brand-hover focus-visible:focus-ring"
          >
            К регистрации
          </Link>
        </div>
      </header>

      <main className="mx-auto w-full max-w-3xl flex-1 px-6 py-10">
        <article className="rounded-card bg-surface p-8 shadow-card sm:p-10">
          <Markdown>{terms}</Markdown>
        </article>
      </main>

      <footer className="pb-10 text-center text-xs text-ink-faint">
        © 2026 taygantt
      </footer>
    </div>
  );
}
