import { AsyncLocalStorage } from 'node:async_hooks';

export type RequestLogContext = {
  requestId?: string;
  traceId?: string;
  userId?: string;
  tenantId?: string;
};

const storage = new AsyncLocalStorage<RequestLogContext>();

export function runWithRequestContext<T>(ctx: RequestLogContext, fn: () => T): T {
  return storage.run(ctx, fn);
}

export function getRequestContext(): RequestLogContext {
  return storage.getStore() ?? {};
}
