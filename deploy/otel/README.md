# Local observability stack (V1)

OpenTelemetry Collector fans out to **Tempo** (traces), **Prometheus** (metrics), and **Loki** (logs).

## Start

```bash
# Observability only (no app services)
make up-obs

# Full local stack (infra + apps + observability)
make up
```

## Endpoints

| Service | URL | Purpose |
|---------|-----|---------|
| OTLP gRPC | `localhost:4317` | Send traces/metrics/logs from services |
| OTLP HTTP | `localhost:4318` | Same, HTTP/protobuf |
| Prometheus | `http://localhost:8889/metrics` | OTel-exported metrics |
| Tempo | `http://localhost:3200` | Trace query API |
| Loki | `http://localhost:3100` | Log storage |
| Grafana | `http://localhost:3003` | Explore traces, metrics, logs (anonymous Viewer) |

## Service configuration

Set in `.env` (already in `.env.example`):

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
```

Go services export traces + logs; Nest `@ting/logger` mirrors the same (stdout JSON + OTLP when endpoint set).

**Logs:** Go (`go/pkg/otel/logs.go`) and Nest (`@ting/logger`) fan out JSON stdout **and** OTLP logs
to Collector → Loki. Disable OTLP logs with `OTEL_LOGS_EXPORTER=none`.

## Grafana

Provisioned datasources: **Tempo** (default), **OTel-Prometheus**, **Loki**.

Dashboard **Platform overview** links Explore views for traces, metrics, and logs.
