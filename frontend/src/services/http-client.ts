/**
 * Все запросы идут через /backend, который Next.js проксирует на BACKEND_URL
 * (см. rewrites в next.config.ts). Бэкенд сам монтирует API под /api/v1.
 */
const API_BASE = "/backend/api/v1";

const ACCESS_TOKEN_STORAGE_KEY = "taygant.accessToken";

let accessToken: string | null = null;

/** Сохраняет access-токен в памяти и localStorage для последующих запросов. */
export function setAccessToken(token: string | null): void {
  accessToken = token;
  if (typeof window === "undefined") return;
  if (token) {
    window.localStorage.setItem(ACCESS_TOKEN_STORAGE_KEY, token);
  } else {
    window.localStorage.removeItem(ACCESS_TOKEN_STORAGE_KEY);
  }
}

/** Возвращает текущий access-токен, подхватывая его из localStorage при первом обращении. */
export function getAccessToken(): string | null {
  if (accessToken) return accessToken;
  if (typeof window !== "undefined") {
    accessToken = window.localStorage.getItem(ACCESS_TOKEN_STORAGE_KEY);
  }
  return accessToken;
}

/** Ошибка ответа API с сохранённым HTTP-статусом и телом ответа. */
export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

export type QueryValue = string | number | boolean | undefined | null;

export interface RequestOptions {
  /** Параметры query-строки; undefined/null значения опускаются. */
  query?: Record<string, QueryValue>;
  /** Сигнал для отмены запроса. */
  signal?: AbortSignal;
  /** Если false — не добавлять заголовок Authorization (по умолчанию true). */
  auth?: boolean;
}

function buildUrl(path: string, query?: Record<string, QueryValue>): string {
  const base =
    typeof window !== "undefined" ? window.location.origin : "http://localhost";
  const url = new URL(API_BASE + path, base);
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value === undefined || value === null) continue;
      url.searchParams.set(key, String(value));
    }
  }
  return url.pathname + url.search;
}

async function parseErrorMessage(res: Response, data: unknown): Promise<string> {
  if (data && typeof data === "object" && "message" in data) {
    const message = (data as { message?: unknown }).message;
    if (typeof message === "string" && message.length > 0) return message;
  }
  return res.statusText || `Request failed with status ${res.status}`;
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  options: RequestOptions = {},
): Promise<T> {
  const { query, signal, auth = true } = options;
  const headers: Record<string, string> = {};

  let payload: BodyInit | undefined;
  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
    payload = JSON.stringify(body);
  }

  if (auth) {
    const token = getAccessToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(buildUrl(path, query), {
    method,
    headers,
    body: payload,
    signal,
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const contentType = res.headers.get("content-type") ?? "";
  const isJson = contentType.includes("application/json");
  const data = isJson ? await res.json().catch(() => undefined) : undefined;

  if (!res.ok) {
    throw new ApiError(res.status, await parseErrorMessage(res, data), data);
  }

  return data as T;
}

export type BlobResponse = {
  blob: Blob;
  /** Значение Content-Disposition, если сервер его прислал (для имени скачиваемого файла). */
  contentDisposition: string | null;
};

async function requestBlob(
  method: string,
  path: string,
  options: RequestOptions = {},
): Promise<BlobResponse> {
  const { query, signal, auth = true } = options;
  const headers: Record<string, string> = {};

  if (auth) {
    const token = getAccessToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(buildUrl(path, query), { method, headers, signal });

  if (!res.ok) {
    const contentType = res.headers.get("content-type") ?? "";
    const data = contentType.includes("application/json")
      ? await res.json().catch(() => undefined)
      : undefined;
    throw new ApiError(res.status, await parseErrorMessage(res, data), data);
  }

  return {
    blob: await res.blob(),
    contentDisposition: res.headers.get("content-disposition"),
  };
}

export const httpClient = {
  get: <T>(path: string, options?: RequestOptions) =>
    request<T>("GET", path, undefined, options),
  post: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>("POST", path, body, options),
  patch: <T>(path: string, body?: unknown, options?: RequestOptions) =>
    request<T>("PATCH", path, body, options),
  delete: <T>(path: string, options?: RequestOptions) =>
    request<T>("DELETE", path, undefined, options),
  getBlob: (path: string, options?: RequestOptions) =>
    requestBlob("GET", path, options),
};
