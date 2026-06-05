# Tracing conventions

OpenTelemetry is the cross-language observability standard.

## Propagation

- Use **W3C Trace Context**: propagate the `traceparent` header on every hop
  (client → Gateway → service, service → service, worker → service).
- `baggage` is optional.
- Always propagate identity headers alongside trace headers:
  `X-Request-Id`, `X-User-Id`, `X-Tenant-Id`, `X-Roles`, `X-Scopes`, `X-Auth-Subject`.

## Correlation

- The Gateway generates `request_id` at the edge (client-supplied `X-Request-Id`
  is not trusted). Internally it is preserved as the correlation id.
- Include `trace_id` and `request_id` in every log line (see logging.schema.json).

## Spans

- Name spans by operation, not by URL with ids in it.
- Follow OpenTelemetry semantic conventions for attribute names
  (e.g. `http.request.method`, `server.address`).
