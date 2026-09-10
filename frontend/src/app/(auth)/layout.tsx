import type { ReactNode } from "react";

/**
 * Оболочка экранов входа/регистрации: светлый фон #F8F9FF с двумя размытыми
 * брендовыми пятнами, по центру — карточка 420px, снизу копирайт (как в макете).
 */
export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="relative flex min-h-full flex-1 flex-col items-center justify-center overflow-hidden bg-page px-4 py-12">
      <div aria-hidden className="pointer-events-none absolute inset-0">
        <div className="absolute top-[9%] left-[29%] size-56 rounded-card bg-brand/10 blur-[64px]" />
        <div className="absolute top-[58%] left-[53%] size-64 rounded-card bg-accent/5 blur-[64px]" />
      </div>

      <main className="relative w-full max-w-[420px] rounded-card bg-surface p-8 shadow-card">
        {children}
      </main>

      <p className="relative mt-6 text-xs text-ink-faint">© 2026 taygantt</p>
    </div>
  );
}
