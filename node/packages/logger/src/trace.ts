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
