import type { ErrorEnvelope } from '@ting/api-types';

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

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
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
