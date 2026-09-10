import { LogoTile } from "@/components/brand/logo";

/** Шапка карточки: знак с плашкой «ТПУ MVP», заголовок и подзаголовок. */
export function AuthHeading({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="flex flex-col items-center text-center">
      <div className="relative">
        <LogoTile />
        <span className="absolute -bottom-2 left-1/2 -translate-x-1/2 rounded-[2px] bg-brand-chip px-1.5 py-1 text-[10px] leading-none font-bold tracking-wide whitespace-nowrap text-brand">
          ТПУ MVP
        </span>
      </div>
      <h1 className="mt-6 text-[26px] leading-tight font-bold tracking-tight text-ink">{title}</h1>
      <p className="mt-1 text-sm text-ink-muted">{subtitle}</p>
    </div>
  );
}
