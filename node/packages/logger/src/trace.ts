/** W3C Trace Context header name. */
export const HEADER_TRACEPARENT = 'traceparent';

/** Extract 32-hex trace id from a traceparent value. */
export function traceIdFromParent(traceparent: string | undefined): string {
  if (!traceparent) {
    return '';
  }
  const parts = traceparent.trim().split('-');
  if (parts.length < 2 || parts[1].length !== 32) {
    return '';
  }
  if (!/^[0-9a-fA-F]{32}$/.test(parts[1])) {
    return '';
  }
  return parts[1].toLowerCase();
}

function pickHeader(
  headers: Record<string, string | string[] | undefined>,
  key: string,
): string {
  const raw = headers[key];
  if (Array.isArray(raw)) {
    return raw[0] ?? '';
  }
  return raw ?? '';
}

/** Create a W3C traceparent for a new root span. */
export function newTraceparent(): string {
  const traceId = crypto.randomUUID().replace(/-/g, '');
  const spanId = crypto.randomUUID().replace(/-/g, '').slice(0, 16);
  return `00-${traceId}-${spanId}-01`;
}

/**
 * Ensure the request has traceparent (preserve inbound, generate if missing)
 * and mirror it on the response for correlation.
 */
export function ensureTraceparent(
  req: { headers: Record<string, string | string[] | undefined> },
  res: { setHeader(name: string, value: string): void },
): string {
  let tp = pickHeader(req.headers, HEADER_TRACEPARENT);
  if (!tp) {
    tp = newTraceparent();
    req.headers[HEADER_TRACEPARENT] = tp;
  }
  res.setHeader(HEADER_TRACEPARENT, tp);
  return tp;
}
