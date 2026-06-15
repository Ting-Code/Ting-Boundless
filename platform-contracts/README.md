# platform-contracts

The cross-language **source of truth**. Define shared behavior here first;
language SDKs (e.g. Go `go/pkg/`, TS `@ting/api`) are helpers, not the source of truth.

## OpenAPI (REST /v1)

| Spec | Service | Paths |
|------|---------|-------|
| `openapi/common.v1.yaml` | shared | `ErrorEnvelope` |
| `openapi/business.v1.yaml` | Nest business-service | `/v1/business/*` |
| `openapi/users.v1.yaml` | Go user-service | `/v1/users/me` |
| `openapi/files.v1.yaml` | Go file-service | `/v1/files/*` |
| `openapi/audit.v1.yaml` | Go audit-service | `/v1/audit/events` |

Generate TS types: `make generate-api` (openapi-typescript → `@ting/api`)

Lint: `make lint-openapi` · Breaking (PR): `make openapi-breaking` (requires [oasdiff](https://github.com/Tufin/oasdiff))

**V1 Web 后台** 消费上表全部 spec。小程序 / App 登录走 Go `auth-service`，暂不新增 `auth.v1.yaml` 或 Node 客户端。

## Contents

| Path | Purpose |
|------|---------|
| `proto/` | Protobuf API + shared message contracts (error, identity, audit) |
| `buf.yaml` / `buf.gen.yaml` | buf lint, breaking-change, multi-language codegen |
| `schemas/` | Language-neutral JSON Schemas (logging, audit event, error, identity) |
| `docs/` | Logging, tracing, and metrics conventions |

## Rules

- External APIs are versioned with a `/v1` path prefix; proto packages are
  versioned (`ting.common.v1`).
- Breaking changes must be caught by `buf breaking` in CI before deploy.
- When a shared shape changes, update the proto/schema here first, then
  regenerate and adapt SDKs (`make proto`).

## Generate

```bash
make proto            # buf lint && buf generate  -> go/gen/go/ (gitignored)
make proto-breaking   # buf breaking against main
make lint-openapi     # Redocly lint on openapi/*.yaml
```

> `go/gen/` is gitignored; stubs land at `go/gen/go/ting/common/v1/`.
