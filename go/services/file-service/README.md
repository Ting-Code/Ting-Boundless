# file-service

File upload/download over S3-compatible storage (MinIO / Aliyun OSS).

| Method | Path | Notes |
|--------|------|-------|
| POST | `/v1/files/` | Multipart field `file` → S3 + `files` metadata row |
| GET | `/v1/files/` | List current user's files (`?limit=50`) |
| GET | `/v1/files/{id}` | File metadata (owner only) |
| GET | `/v1/files/{id}/download` | Stream object bytes from S3 (service proxy) |
| GET | `/v1/files/{id}/url` | Presigned GET URL (`?expires=` seconds, default `FILE_PRESIGN_SECONDS`) |
| DELETE | `/v1/files/{id}` | Delete object + metadata (owner only) |

Identity from Gateway headers only (`identity.Middleware`).

## Data

`files` table in `app_db` (`go/migrations/file-service/`) — metadata for objects
stored in S3. Applied on startup and via `make migrate`.

## Environment

| Variable | Purpose |
|----------|---------|
| `HTTP_ADDR` | Listen (default `:8083`) |
| `POSTGRES_*` | `app_db` for file metadata |
| `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET` | Object storage (MinIO/OSS) |
| `FILE_MAX_BYTES` | Max upload size (default 20971520 = 20 MiB) |
| `FILE_PRESIGN_SECONDS` | Presigned URL TTL (default 3600) |
| `INTERNAL_API_TOKEN` | Gateway trust |

## Local

```bash
make migrate
make run-file-service
# MinIO: make up-infra (includes minio) or local MinIO on :9000
curl -F "file=@./README.md" -H "X-User-Id: dev" -H "X-Internal-Token: $INTERNAL_API_TOKEN" \
  http://127.0.0.1:8083/v1/files/
# Through Gateway (preferred):
curl -F "file=@./README.md" -H "Authorization: Bearer $(make dev-jwt)" http://127.0.0.1:8080/v1/files/
```
