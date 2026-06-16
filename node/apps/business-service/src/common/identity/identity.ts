export const IDENTITY_HEADERS = {
  requestId: 'x-request-id',
  userId: 'x-user-id',
  tenantId: 'x-tenant-id',
  roles: 'x-roles',
  scopes: 'x-scopes',
  subject: 'x-auth-subject',
} as const;

export type Identity = {
  requestId: string;
  userId: string;
  tenantId: string;
  roles: string[];
  scopes: string[];
  subject: string;
};

export function splitCsv(value: string | undefined): string[] {
  if (!value) {
    return [];
  }
  return value
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
}

export function identityFromHeaders(headers: Record<string, string | string[] | undefined>): Identity {
  const pick = (key: string): string => {
    const raw = headers[key];
    if (Array.isArray(raw)) {
      return raw[0] ?? '';
    }
    return raw ?? '';
  };

  return {
    requestId: pick(IDENTITY_HEADERS.requestId),
    userId: pick(IDENTITY_HEADERS.userId),
    tenantId: pick(IDENTITY_HEADERS.tenantId),
    roles: splitCsv(pick(IDENTITY_HEADERS.roles)),
    scopes: splitCsv(pick(IDENTITY_HEADERS.scopes)),
    subject: pick(IDENTITY_HEADERS.subject),
  };
}

export function isAuthenticated(id: Identity | undefined): boolean {
  return Boolean(id?.userId);
}

export function hasRole(id: Identity | undefined, role: string): boolean {
  return id?.roles.includes(role) ?? false;
}
