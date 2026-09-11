import { LogoTile } from "@/components/brand/logo";

/** Шапка карточки: знак, заголовок и подзаголовок. */
export function AuthHeading({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <div className="flex flex-col items-center text-center">
      <LogoTile />
      <h1 className="mt-6 text-26 leading-tight font-bold tracking-tight text-ink">{title}</h1>
      <p className="mt-1 text-sm text-ink-muted">{subtitle}</p>
    </div>
  );
}
