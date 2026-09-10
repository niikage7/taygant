import { cn } from "@/lib/utils";

/**
 * Знак «ПроектКонтроль»: связанные полосы Ганта со стрелкой зависимости.
 * Воспроизведён из макета в векторе, чтобы масштабироваться без потери качества.
 */
export function LogoMark({ className, ...props }: React.ComponentProps<"svg">) {
  return (
    <svg
      viewBox="0 0 80 80"
      role="img"
      aria-label="ПроектКонтроль"
      className={cn("size-10", className)}
      {...props}
    >
      <rect width="80" height="80" rx="18" fill="var(--color-brand-logo)" />
      <rect x="16" y="21.5" width="27" height="7.5" rx="3.75" fill="#fff" />
      <path
        d="M43.5 24.75 49.5 29.5 43.5 34.5"
        fill="none"
        stroke="var(--color-brand-logo-soft)"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <rect x="32" y="36" width="31" height="7.5" rx="3.75" fill="var(--color-brand-logo-soft)" />
      <rect x="24" y="50" width="24" height="7.5" rx="3.75" fill="#fff" />
      <circle cx="56" cy="53.75" r="6" fill="var(--color-brand-logo-dot)" />
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
