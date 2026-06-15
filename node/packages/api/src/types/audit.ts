import type { components } from '../generated/audit.v1';

export type AuditEvent = components['schemas']['AuditEvent'];
export type ListAuditEventsResponse = components['schemas']['ListAuditEventsResponse'];

/** Query string params for listAuditEvents (not generated as a schema). */
export type ListAuditEventsQuery = {
  limit?: number;
  type?: string;
  source?: string;
};
