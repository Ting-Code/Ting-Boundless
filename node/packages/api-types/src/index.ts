/**
 * Shared API types. Mirror platform-contracts/openapi/business.v1.yaml until
 * openapi-typescript generation is wired (pnpm generate).
 */

export type BusinessPingResponse = {
  ok: true;
  service: string;
};

export type BusinessMeResponse = {
  user_id: string;
  tenant_id: string;
  roles: string[];
  scopes: string[];
  subject: string;
  request_id: string;
};

export type BusinessItem = {
  id: string;
  tenant_id: string;
  title: string;
  body: string;
  created_by: string;
  created_at: string;
  updated_at: string;
};

export type ListItemsResponse = {
  items: BusinessItem[];
};

export type CreateItemRequest = {
  title: string;
  body?: string;
};

export type CreateItemResponse = {
  item: BusinessItem;
};

export type ErrorEnvelope = {
  error: {
    code: string;
    message: string;
    request_id?: string;
  };
};
