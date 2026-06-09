# Logging conventions

Every service emits **JSON to stdout** (12-Factor). Collection is the platform's job.

## Schema

Authoritative shape: `schemas/logging.schema.json`.

Required on every line:

```text
@timestamp
log.level
message
service.name
```

Access logs (`message=http_request`) must also include:

```text
request_id
http.request.method
http.response.status_code
url.path          # normalized route, not raw ids
```

When available:

```text
trace_id          # from traceparent
user_id
tenant_id
```

## SDKs

| Language | Package |
|----------|---------|
| Go | `go/pkg/logger`, `go/pkg/httpx.AccessLog` |
| TypeScript | `@ting/logger` (`LoggingInterceptor` for Nest) |

Business code logs domain events with stable `message` names (`item_created`, not sentences).

## Do not log

- Secrets, tokens, cookies, passwords
- Full request/response bodies with PII
- Audit-worthy actions (use CloudEvents → audit-service instead)

## Correlation

- Gateway generates `request_id` at the edge; propagates `X-Request-Id` internally.
- Gateway generates `traceparent` when absent; propagate on every hop.
- See `docs/tracing.md`.
