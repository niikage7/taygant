import type { Components } from "react-markdown";
import ReactMarkdown from "react-markdown";
import remarkBreaks from "remark-breaks";

/**
 * Оформление markdown токенами дизайн-системы.
 *
 * Плагин typography сознательно не подключаем: он приносит собственную шкалу
 * цветов и отступов, которая расходится с макетом. Здесь ровно те элементы,
 * что встречаются в правовых документах.
 */
const components: Components = {
  h1: ({ children }) => (
    <h1 className="mb-2 text-[26px] leading-tight font-bold tracking-tight text-ink">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="mt-10 mb-3 border-t border-line pt-6 text-lg font-semibold text-ink first:mt-0 first:border-0 first:pt-0">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="mt-6 mb-2 text-base font-semibold text-ink">{children}</h3>
  ),
  p: ({ children }) => (
    <p className="mt-3 text-sm leading-relaxed text-ink-muted">{children}</p>
  ),
  // Маркеры задаём на самих списках, а не на <li>: иначе кастомная точка
  // маркированного списка накладывалась бы на нумерацию упорядоченного.
  ul: ({ children }) => (
    <ul className="mt-3 space-y-2 text-sm leading-relaxed text-ink-muted [&>li]:relative [&>li]:pl-4 [&>li]:before:absolute [&>li]:before:top-2.5 [&>li]:before:left-0 [&>li]:before:size-1 [&>li]:before:rounded-full [&>li]:before:bg-brand">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="mt-3 list-decimal space-y-2 pl-5 text-sm leading-relaxed text-ink-muted marker:text-ink-faint">
      {children}
    </ol>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold text-ink">{children}</strong>
  ),
  em: ({ children }) => <em className="text-ink-faint italic">{children}</em>,
  a: ({ href, children }) => (
    <a
      href={href}
      className="rounded-control font-medium text-brand underline underline-offset-2 hover:text-brand-hover focus-visible:focus-ring"
    >
      {children}
    </a>
  ),
  hr: () => <hr className="mt-10 border-line" />,
  code: ({ children }) => (
    <code className="rounded-control bg-surface-muted px-1.5 py-0.5 font-mono text-[13px] text-ink">
      {children}
    </code>
  ),
};

/** Рендерит markdown-документ в оформлении дизайн-системы. */
export function Markdown({ children }: { children: string }) {
  return (
    <ReactMarkdown
      // remark-breaks: одиночные переносы строк в исходнике становятся <br>.
      // В правовых текстах строки вроде «Дата публикации» и «Редакция» набраны
      // отдельными строками и без этого склеились бы в одну.
      remarkPlugins={[remarkBreaks]}
      components={components}
    >
      {children}
    </ReactMarkdown>
  );
}
