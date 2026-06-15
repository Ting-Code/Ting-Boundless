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
} as const;

export const filePaths = {
  list: '/v1/files/',
  item: (id: string) => `/v1/files/${id}`,
  download: (id: string) => `/v1/files/${id}/download`,
  url: (id: string) => `/v1/files/${id}/url`,
} as const;

export const auditPaths = {
  events: '/v1/audit/events',
} as const;
