import { cn } from "@/lib/utils";

/**
 * Знак «taygant»: три полосы Ганта в синем скруглённом квадрате.
 * Воспроизведён из макета в векторе, чтобы масштабироваться без потери качества.
 */
export function LogoMark({ className, ...props }: React.ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 44 44"
      role="img"
      aria-label="taygant"
      className={cn("size-10", className)}
      {...props}
    >
      <rect width="44" height="44" rx="12" fill="var(--color-brand-logo)" />
      <rect x="8" y="11.9" width="17" height="4.2" rx="2.1" fill="#fff" />
      <rect
        x="21"
        y="18.9"
        width="14.5"
        height="4.2"
        rx="2.1"
        fill="var(--color-brand-logo-soft)"
      />
      <rect x="12" y="25.9" width="16" height="4.2" rx="2.1" fill="#fff" />
    </svg>
  );
}

/** Логотип в плитке светло-синего фона — как в шапке экранов авторизации. */
export function LogoTile({ className }: { className?: string }) {
  return (
    <span
      className={cn(
        "flex size-14 items-center justify-center rounded-control bg-brand-tint",
        className,
      )}
    >
      <LogoMark className="size-10" />
    </span>
  );
}
