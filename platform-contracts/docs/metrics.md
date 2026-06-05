# Metrics conventions

Prometheus is the metrics standard; export via the OpenTelemetry SDK or a native
Prometheus client. Every service exposes `/metrics`.

## Naming

- Use OpenTelemetry semantic conventions where they exist.
- Prefix custom metrics with the domain, e.g. `user_registrations_total`.
- Counters end in `_total`; durations are seconds (`_seconds`).

## Baseline metrics per service

- `http_server_request_duration_seconds` (histogram, labels: method, route, status)
- `http_server_active_requests` (gauge)
- process/runtime metrics (Go runtime collector)

## Labels

- Keep cardinality bounded. Never use raw user ids, request ids, or full paths
  with ids as label values.
