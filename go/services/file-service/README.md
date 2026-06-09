# file-service

File upload/download over S3-compatible storage (MinIO / Aliyun OSS).

Endpoints: `POST /v1/files/`.

Env: `HTTP_ADDR`, `S3_ENDPOINT`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_BUCKET`,
`POSTGRES_*`.

TODO: streaming upload to S3, metadata in app_db, signed download URLs.
