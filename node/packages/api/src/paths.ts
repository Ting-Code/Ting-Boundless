/** Gateway-relative API paths (single source for frontends and SSR). */

export const businessPaths = {
  ping: '/v1/business/ping',
  me: '/v1/business/me',
  items: '/v1/business/items',
  item: (id: string) => `/v1/business/items/${id}`,
} as const;

export const userPaths = {
  list: '/v1/users/',
  me: '/v1/users/me',
  listQuery: (limit = 50) => `/v1/users/?limit=${limit}`,
} as const;

export const filePaths = {
  list: '/v1/files/',
  listQuery: (limit = 50) => `/v1/files/?limit=${limit}`,
  item: (id: string) => `/v1/files/${id}`,
  download: (id: string) => `/v1/files/${id}/download`,
  url: (id: string) => `/v1/files/${id}/url`,
} as const;

export type AuditEventsQuery = {
  limit?: number;
  type?: string;
  source?: string;
};

export const auditPaths = {
  events: '/v1/audit/events',
  eventsQuery: ({ limit = 50, type, source }: AuditEventsQuery = {}) => {
    const params = new URLSearchParams({ limit: String(limit) });
    if (type) {
      params.set('type', type);
    }
    if (source) {
      params.set('source', source);
    }
    return `/v1/audit/events?${params}`;
  },
} as const;
