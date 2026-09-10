import { NextResponse, type NextRequest } from "next/server";

/**
 * Заголовки, из которых Fiber на бэкенде выводит IP клиента (ProxyHeader +
 * TrustedProxies) и по которым ключуется rate limit.
 */
const CLIENT_IP_HEADERS = [
  "x-forwarded-for",
  "x-real-ip",
  "x-client-ip",
  "forwarded",
];

/**
 * Срезает заголовки определения IP на пути к бэкенду.
 *
 * Rewrite прокидывает заголовки браузера в destination как есть — проверено:
 * запрос с `X-Forwarded-For: 203.0.113.9` доходит до бэкенда с этим значением.
 * Адрес Next при этом попадает в TRUSTED_PROXIES бэкенда (там по умолчанию
 * 127.0.0.1 и приватные подсети), поэтому Fiber такому заголовку верит. Значит
 * кто угодно мог бы слать произвольный X-Forwarded-For: крутить его от запроса
 * к запросу, обходя лимит, или наоборот выжечь квоту чужого IP.
 *
 * Сам Next реального адреса клиента не знает и подставить его не может: своего
 * X-Forwarded-For он не добавляет (тоже проверено — до бэкенда доходит только
 * x-forwarded-host). Поэтому единственный безопасный вариант — не передавать
 * ничего: бэкенд увидит адрес Next и будет знать, что IP недостоверен.
 *
 * ⚠️ Если перед Next появится настоящий обратный прокси (nginx, Traefik), его
 * X-Forwarded-For будет достоверным, и эту логику нужно пересмотреть: доверять
 * заголовку только от адреса того прокси.
 */
export function proxy(request: NextRequest) {
  const headers = new Headers(request.headers);
  for (const header of CLIENT_IP_HEADERS) {
    headers.delete(header);
  }
  return NextResponse.next({ request: { headers } });
}

export const config = {
  matcher: "/backend/:path*",
};
