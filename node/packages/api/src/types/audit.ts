import type { components } from '../generated/audit.v1';
import type { AuditEventsQuery } from '../paths';

export type AuditEvent = components['schemas']['AuditEvent'];
export type ListAuditEventsResponse = components['schemas']['ListAuditEventsResponse'];

/** @alias AuditEventsQuery */
export type ListAuditEventsQuery = AuditEventsQuery;
export type { AuditEventsQuery };
