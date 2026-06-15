import type { LogLevel } from './logger';

let emitFn: ((level: LogLevel, message: string, fields: Record<string, unknown>) => void) | null =
  null;

/** Registers OTLP log emission (called from initOtelFromEnv). */
export function registerOtelLogEmitter(
  fn: (level: LogLevel, message: string, fields: Record<string, unknown>) => void,
): void {
  emitFn = fn;
}

export function emitOtelLog(level: LogLevel, message: string, fields: Record<string, unknown>): void {
  emitFn?.(level, message, fields);
}
