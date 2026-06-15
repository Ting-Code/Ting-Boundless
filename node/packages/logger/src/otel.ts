import { diag, DiagConsoleLogger, DiagLogLevel } from '@opentelemetry/api';
import { logs, SeverityNumber } from '@opentelemetry/api-logs';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPLogExporter as GrpcLogExporter } from '@opentelemetry/exporter-logs-otlp-grpc';
import { OTLPLogExporter as HttpLogExporter } from '@opentelemetry/exporter-logs-otlp-http';
import { OTLPTraceExporter as GrpcTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc';
import { OTLPTraceExporter as HttpTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { NodeSDK } from '@opentelemetry/sdk-node';
import { BatchLogRecordProcessor, LoggerProvider } from '@opentelemetry/sdk-logs';
import { SEMRESATTRS_SERVICE_NAME } from '@opentelemetry/semantic-conventions';
import type { LogLevel } from './logger';
import { registerOtelLogEmitter } from './otel-logs';

function isPlaceholder(endpoint: string): boolean {
  const s = endpoint.toLowerCase();
  return (
    s.includes('placeholder') ||
    s.includes('example.invalid') ||
    s.includes('rm-xxx')
  );
}

function grpcURL(endpoint: string): string {
  if (endpoint.includes('://')) {
    return endpoint;
  }
  return `http://${endpoint.replace(/\/$/, '')}`;
}

function logsExportEnabled(): boolean {
  const v = (process.env.OTEL_LOGS_EXPORTER ?? 'otlp').trim().toLowerCase();
  return v !== 'none' && v !== 'false' && v !== 'off';
}

function severity(level: LogLevel): SeverityNumber {
  switch (level) {
    case 'debug':
      return SeverityNumber.DEBUG;
    case 'warn':
      return SeverityNumber.WARN;
    case 'error':
      return SeverityNumber.ERROR;
    default:
      return SeverityNumber.INFO;
  }
}

let sdk: NodeSDK | null = null;
let logProvider: LoggerProvider | null = null;

/**
 * Initializes OpenTelemetry when OTEL_EXPORTER_OTLP_ENDPOINT is set.
 * Import from a dedicated `instrument.ts` before other application modules.
 */
export function initOtelFromEnv(serviceName: string): (() => Promise<void>) | null {
  const endpoint = process.env.OTEL_EXPORTER_OTLP_ENDPOINT?.trim() ?? '';
  if (!endpoint || isPlaceholder(endpoint)) {
    return null;
  }

  const svc =
    process.env.OTEL_SERVICE_NAME?.trim() && process.env.OTEL_SERVICE_NAME !== 'unset'
      ? process.env.OTEL_SERVICE_NAME.trim()
      : serviceName;

  if (process.env.OTEL_LOG_LEVEL === 'debug') {
    diag.setLogger(new DiagConsoleLogger(), DiagLogLevel.DEBUG);
  }

  const protocol = (process.env.OTEL_EXPORTER_OTLP_PROTOCOL ?? '').toLowerCase();
  const useGrpc = protocol === 'grpc' || (protocol === '' && endpoint.includes(':4317'));

  const traceExporter = useGrpc
    ? new GrpcTraceExporter({ url: grpcURL(endpoint) })
    : new HttpTraceExporter({ url: endpoint });

  const resource = new Resource({
    [SEMRESATTRS_SERVICE_NAME]: svc,
  });

  if (logsExportEnabled()) {
    const logExporter = useGrpc
      ? new GrpcLogExporter({ url: grpcURL(endpoint) })
      : new HttpLogExporter({ url: endpoint });
    logProvider = new LoggerProvider({ resource });
    logProvider.addLogRecordProcessor(new BatchLogRecordProcessor(logExporter));
    logs.setGlobalLoggerProvider(logProvider);
    const otelLogger = logs.getLogger(svc);
    registerOtelLogEmitter((level, message, fields) => {
      otelLogger.emit({
        severityNumber: severity(level),
        body: message,
        attributes: fields as Record<string, string | number | boolean>,
      });
    });
  }

  sdk = new NodeSDK({
    resource,
    traceExporter,
    instrumentations: [
      getNodeAutoInstrumentations({
        '@opentelemetry/instrumentation-fs': { enabled: false },
      }),
    ],
  });

  sdk.start();
  return async () => {
    if (logProvider) {
      await logProvider.shutdown();
      logProvider = null;
      registerOtelLogEmitter(() => undefined);
    }
    if (sdk) {
      await sdk.shutdown();
      sdk = null;
    }
  };
}
