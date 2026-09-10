import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    default: "ПроектКонтроль",
    template: "%s · ПроектКонтроль",
  },
  description: "Управление проектами и диаграмма Ганта",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="ru" className="h-full antialiased">
      <head>
        {/*
          Шрифты объявлены в globals.css (@font-face поверх /public/fonts).
          Кириллические подмножества грузятся по unicode-range сами, а латинские
          нужны на каждой странице — их подгружаем заранее, чтобы не ловить FOUT.
        */}
        <link
          rel="preload"
          href="/fonts/inter-latin.woff2"
          as="font"
          type="font/woff2"
          crossOrigin="anonymous"
        />
        <link
          rel="preload"
          href="/fonts/inter-cyrillic.woff2"
          as="font"
          type="font/woff2"
          crossOrigin="anonymous"
        />
      </head>
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
