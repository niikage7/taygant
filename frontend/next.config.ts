import type { NextConfig } from "next";

/**
 * Адрес Go-бэкенда для server-side rewrite. Без запасного значения незаданная
 * переменная давала бы destination вида `undefined/api/v1/...`: rewrite молча
 * ломался бы на рантайме вместо понятной ошибки.
 *
 * Значение по умолчанию совпадает с APP_PORT в docker-compose.yml и с
 * .env.example, так что локальный запуск работает и без .env.
 */
const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:4000";

if (!process.env.BACKEND_URL) {
  console.warn(
    `[next.config] BACKEND_URL не задан — используется ${BACKEND_URL}. ` +
      "Для стенда задайте переменную явно.",
  );
}

const nextConfig: NextConfig = {
  reactCompiler: true,
  // Для Docker: сборка кладёт в .next/standalone готовый server.js и только
  // нужные ему зависимости — образ не тащит весь node_modules. На `pnpm dev`
  // не влияет.
  output: "standalone",
  rewrites: async () => {
    return [
      {
        source: "/backend/:path*",
        destination: `${BACKEND_URL}/:path*`,
      },
    ];
  },
};

export default nextConfig;
