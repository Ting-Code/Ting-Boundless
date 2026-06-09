const UUID_SEGMENT =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const ID_SEGMENT = /^[0-9a-z]{12,}$/i;
const NUM_SEGMENT = /^[0-9]+$/;

/** Replace high-cardinality path segments with :id for access logs. */
export function normalizePath(path: string): string {
  if (!path || path === '/') {
    return path || '/';
  }
  const parts = path.split('/');
  const normalized = parts.map((part) => {
    if (!part) {
      return part;
    }
    if (UUID_SEGMENT.test(part) || ID_SEGMENT.test(part) || NUM_SEGMENT.test(part)) {
      return ':id';
    }
    return part;
  });
  const joined = normalized.join('/');
  return joined.startsWith('/') ? joined : `/${joined}`;
}
