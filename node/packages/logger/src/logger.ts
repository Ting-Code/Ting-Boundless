import { getRequestContext } from './context';

export type LogLevel = 'debug' | 'info' | 'warn' | 'error';

export type LogFields = Record<string, unknown>;

let serviceName = 'unknown';
let minLevel: LogLevel = 'info';

const levelRank: Record<LogLevel, number> = {
  debug: 10,
  info: 20,
  warn: 30,
  error: 40,
};

export function configureLogger(opts: { service: string; level?: string }): void {
  serviceName = opts.service;
  minLevel = parseLevel(opts.level ?? 'info');
}

function parseLevel(level: string): LogLevel {
  const v = level.trim().toLowerCase();
  if (v === 'debug' || v === 'warn' || v === 'warning' || v === 'error') {
    return v === 'warning' ? 'warn' : (v as LogLevel);
  }
  return 'info';
}

function shouldLog(level: LogLevel): boolean {
  return levelRank[level] >= levelRank[minLevel];
}

export function log(level: LogLevel, message: string, fields: LogFields = {}): void {
  if (!shouldLog(level)) {
    return;
  }
  const ctx = getRequestContext();
  const line = JSON.stringify({
    '@timestamp': new Date().toISOString(),
    'log.level': level,
    message,
    'service.name': serviceName,
    ...(ctx.requestId ? { request_id: ctx.requestId } : {}),
    ...(ctx.traceId ? { trace_id: ctx.traceId } : {}),
    ...(ctx.userId ? { user_id: ctx.userId } : {}),
    ...(ctx.tenantId ? { tenant_id: ctx.tenantId } : {}),
    ...fields,
  });
  if (level === 'error') {
    console.error(line);
  } else {
    console.log(line);
  }
}

export const logger = {
  debug: (message: string, fields?: LogFields) => log('debug', message, fields),
  info: (message: string, fields?: LogFields) => log('info', message, fields),
  warn: (message: string, fields?: LogFields) => log('warn', message, fields),
  error: (message: string, fields?: LogFields) => log('error', message, fields),
};

/** Nest Logger adapter that emits ECS JSON lines. */
export function createNestLoggerAdapter(): {
  log: (message: unknown) => void;
  error: (message: unknown, trace?: string) => void;
  warn: (message: unknown) => void;
  debug: () => void;
  verbose: () => void;
} {
  return {
    log: (message) => logger.info(String(message)),
    error: (message, trace) => logger.error(String(message), trace ? { trace } : {}),
    warn: (message) => logger.warn(String(message)),
    debug: () => undefined,
    verbose: () => undefined,
  };
}
