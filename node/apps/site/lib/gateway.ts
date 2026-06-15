import {
  apiFetch,
  businessPaths,
  type BusinessPingResponse,
} from '@ting/api';

/** Server-side Gateway base URL (not exposed to the browser). */
export function gatewayBase(): string {
  return process.env.GATEWAY_URL ?? 'http://127.0.0.1:8080';
}

export async function fetchBusinessPing(): Promise<BusinessPingResponse | null> {
  try {
    return await apiFetch<BusinessPingResponse>(businessPaths.ping, {
      baseUrl: gatewayBase(),
      cache: 'no-store',
      credentials: 'omit',
    });
  } catch {
    return null;
  }
}
