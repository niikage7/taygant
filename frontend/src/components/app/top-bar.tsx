import { Search } from "lucide-react";

/** Верхняя панель с глобальным поиском по задачам. */
export function TopBar() {
  return (
    <header className="flex h-14 shrink-0 items-center justify-center border-b border-line bg-surface px-6">
      <label className="flex h-8 w-full max-w-xl items-center gap-2 rounded-control border border-line bg-surface-subtle px-3">
        <Search className="size-4 shrink-0 text-ink-faint" />
        <input
          type="search"
          placeholder="Поиск по задачам..."
          className="min-w-0 flex-1 bg-transparent text-sm text-ink outline-none placeholder:text-ink-faint"
        />
        <kbd className="shrink-0 rounded-control border border-line bg-surface px-1.5 py-0.5 font-mono text-[10px] text-ink-faint">
          ⌘K
        </kbd>
      </label>
    </header>
  );
}
