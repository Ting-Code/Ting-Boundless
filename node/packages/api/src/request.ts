import type { ErrorEnvelope } from './types/business';

export class ApiError extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
    public readonly requestId?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/** Join gateway base URL with a relative API path (SSR / server-side). */
export function resolveApiUrl(path: string, baseUrl = ''): string {
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  const base = baseUrl.replace(/\/$/, '');
  return `${base}${path}`;
}

export type ApiFetchOptions = RequestInit & {
  /** Absolute gateway origin for server-side calls (e.g. http://127.0.0.1:8080). */
  baseUrl?: string;
};

export async function apiFetch<T>(path: string, init?: ApiFetchOptions): Promise<T> {
  const { baseUrl, ...requestInit } = init ?? {};
  const url = resolveApiUrl(path, baseUrl);
  const res = await fetch(url, {
    ...requestInit,
    credentials: requestInit.credentials ?? 'include',
    headers: {
      ...(requestInit.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      ...(requestInit.headers ?? {}),
    },
  });

  if (!res.ok) {
    let code = `http.${res.status}`;
    let message = res.statusText;
    let requestId: string | undefined;
    try {
      const body = (await res.json()) as ErrorEnvelope;
      code = body.error?.code ?? code;
      message = body.error?.message ?? message;
      requestId = body.error?.request_id;
    } catch {
      // non-json error body
    }
    throw new ApiError(code, message, res.status, requestId);
  }

  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export async function apiUpload<T>(path: string, file: File, init?: ApiFetchOptions): Promise<T> {
  const body = new FormData();
  body.append('file', file);
  return apiFetch<T>(path, {
    ...init,
    method: 'POST',
    body,
  });
}
