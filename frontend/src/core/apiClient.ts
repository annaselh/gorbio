/**
 * The backend mounts routes directly under /api - see the Register method on
 * each Go module - not under /api/v1. In development Vite proxies /api to
 * localhost:8080; in production the SPA is served same-origin, or its origin is
 * listed in CORS_ORIGINS on the server.
 */
const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? "/api";

export class ApiError extends Error {
  readonly status: number;
  readonly body: unknown;

  constructor(status: number, body: unknown, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.body = body;
  }
}

interface RequestOptions extends Omit<RequestInit, "body"> {
  body?: unknown;
}

export async function request<T>(
  path: string,
  { body, headers, ...init }: RequestOptions = {},
): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    // The session is an HttpOnly cookie the page cannot read, so it has to be
    // attached by the browser. Without this the cookie is never sent and every
    // authenticated call fails with 401.
    credentials: "include",
    headers: {
      ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
      ...headers,
    },
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  const payload = res.status === 204 ? null : await res.json().catch(() => null);

  if (!res.ok) {
    // The Go handlers answer with {"error": "..."}; keep `message` as a
    // fallback so a proxy or gateway error shape still surfaces something.
    const problem = payload as { error?: string; message?: string } | null;
    const message =
      problem?.error ?? problem?.message ?? `${res.status} ${res.statusText}`;
    throw new ApiError(res.status, payload, message);
  }

  return payload as T;
}

export const api = {
  get: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "GET" }),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "POST", body }),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "PUT", body }),
  del: <T>(path: string, opts?: RequestOptions) =>
    request<T>(path, { ...opts, method: "DELETE" }),
};
