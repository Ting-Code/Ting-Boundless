# platform-contracts

The cross-language **source of truth**. Define shared behavior here first;
language SDKs (e.g. Go `pkg/`) are helpers, not the source of truth.

## Contents

| Path | Purpose |
|------|---------|
| `proto/` | Protobuf API + shared message contracts (error, identity, audit) |
| `buf.yaml` / `buf.gen.yaml` | buf lint, breaking-change, multi-language codegen |
| `schemas/` | Language-neutral JSON Schemas (logging, audit event, error, identity) |
| `docs/` | Tracing and metrics conventions |

## Rules

- External APIs are versioned with a `/v1` path prefix; proto packages are
  versioned (`ting.common.v1`).
- Breaking changes must be caught by `buf breaking` in CI before deploy.
- When a shared shape changes, update the proto/schema here first, then
  regenerate and adapt SDKs (`make proto`).

## Generate

```bash
make proto            # buf lint && buf generate  -> gen/ (gitignored)
make proto-breaking   # buf breaking against main
```

> `gen/` is gitignored; code is regenerated from these contracts.
