import { traceIdFromParent, HEADER_TRACEPARENT } from './trace';

export const IDENTITY_HEADERS = {
  requestId: 'x-request-id',
  userId: 'x-user-id',
  tenantId: 'x-tenant-id',
} as const;

export function requestContextFromHeaders(
  headers: Record<string, string | string[] | undefined>,
): {
  requestId?: string;
  traceId?: string;
  userId?: string;
  tenantId?: string;
} {
  const pick = (key: string): string => {
    const raw = headers[key];
    if (Array.isArray(raw)) {
      return raw[0] ?? '';
    }
    return raw ?? '';
  };

  const traceId = traceIdFromParent(pick(HEADER_TRACEPARENT));
  const requestId = pick(IDENTITY_HEADERS.requestId);
  const userId = pick(IDENTITY_HEADERS.userId);
  const tenantId = pick(IDENTITY_HEADERS.tenantId);

  return {
    ...(requestId ? { requestId } : {}),
    ...(traceId ? { traceId } : {}),
    ...(userId ? { userId } : {}),
    ...(tenantId ? { tenantId } : {}),
  };
}
