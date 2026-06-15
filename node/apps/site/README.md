# @ting/site

Public Next.js site (App Router, SSR). Browsers hit **Gateway** (`/`); Gateway
reverse-proxies non-API paths to this app (`SITE_SERVICE_URL`, default `:3006`).

## Local dev

```bash
# Terminal 1 — Gateway + business (from repo root)
make run-gateway
make run-business

# Terminal 2 — Next dev server
make run-site

# Browse via Gateway (not :3006 directly in production-like flow)
open http://127.0.0.1:8080/
```

Server components call Gateway with `GATEWAY_URL` (default `http://127.0.0.1:8080`).
