# Deploy to Tencent Cloud (GitHub Actions → TCR → CVM)

Pipeline: **GitHub Actions** builds images → **Tencent Container Registry (TCR)** → **CVM**
runs `docker compose` with managed **TencentDB / Redis / COS**.

## Architecture (production)

```text
Internet HTTPS :443
  → Nginx (CVM container, TLS cert)
  → Gateway
  → Go services + Nest business-service (same Docker network)
  → TencentDB PostgreSQL / Tencent Redis / COS (managed, not in compose)
```

Logto: self-host on CVM or use Logto Cloud; point `OIDC_*` in `.env`.

## 1. Tencent resources

| Resource | Product | Notes |
|----------|---------|-------|
| Compute | **CVM** or **Lighthouse** (2C4G+), Ubuntu 22.04 | Install Docker + Compose plugin |
| Registry | **TCR** 个人版/企业版 | e.g. `ccr.ccs.tencentyun.com` |
| Database | **TencentDB PostgreSQL** | `app_db`, `audit_db`, `logto_db` |
| Cache | **TencentDB Redis** | BFF sessions |
| Object storage | **COS** | S3-compatible; set `S3_ENDPOINT` |
| TLS | **SSL 证书** (免费/上传) | Mount to `deploy/nginx/certs/` or CLB |
| DNS | **DNSPod** | `api.example.com` → CVM 公网 IP |

Security group: open **22** (SSH, your IP only), **80**, **443**.

## 2. TCR

1. Console → 容器镜像服务 → 创建命名空间（如 `ting-prod`）
2. 创建镜像仓库（可选）或使用同一前缀
3. 访问凭证 → 设置固定密码

记下：

- `TCR_HOST` = `ccr.ccs.tencentyun.com`
- `TCR_IMAGE_PREFIX` = `ccr.ccs.tencentyun.com/ting-prod/ting-boundless`

## 3. CVM bootstrap (once)

```bash
# Docker
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker "$USER"

# Clone
sudo mkdir -p /opt/ting-boundless
sudo chown "$USER:$USER" /opt/ting-boundless
git clone https://github.com/<you>/Ting-Boundless.git /opt/ting-boundless
cd /opt/ting-boundless

# Production env (never commit)
cp .env.example .env
# Edit: POSTGRES_HOST, REDIS_ADDR, S3_*, OIDC_*, GATEWAY_PUBLIC_URL=https://api.example.com, INTERNAL_API_TOKEN, ...
```

Run DB migrations separately (`make migrate`, Nest drizzle on startup).

### HTTPS on CVM

```bash
# Option A: certbot
sudo apt install certbot
sudo certbot certonly --standalone -d api.example.com
sudo mkdir -p deploy/nginx/certs
sudo cp /etc/letsencrypt/live/api.example.com/fullchain.pem deploy/nginx/certs/
sudo cp /etc/letsencrypt/live/api.example.com/privkey.pem deploy/nginx/certs/
```

Add `listen 443 ssl` to `deploy/nginx/nginx.conf` (see `deploy/nginx/ssl.conf.example`).

## 4. GitHub Secrets

Repository → Settings → Secrets and variables → Actions:

| Secret | Example |
|--------|---------|
| `TCR_HOST` | `ccr.ccs.tencentyun.com` |
| `TCR_IMAGE_PREFIX` | `ccr.ccs.tencentyun.com/ting-prod/ting-boundless` |
| `TCR_USERNAME` | `100012345678` (腾讯云账号 ID) |
| `TCR_PASSWORD` | TCR 固定密码 |
| `DEPLOY_HOST` | CVM 公网 IP |
| `DEPLOY_USER` | `ubuntu` |
| `DEPLOY_SSH_KEY` | 部署用私钥（PEM 全文） |
| `DEPLOY_PATH` | `/opt/ting-boundless` |

## 5. Workflows

| File | Trigger | Action |
|------|---------|--------|
| `.github/workflows/ci.yml` | PR / push `main` | `go test`, `pnpm build` |
| `.github/workflows/deploy-tencent.yml` | push `main` / manual | build → TCR → SSH deploy |

Push to `main` 后自动：测试 → 构建 8 个镜像 → CVM 上 `docker compose pull && up -d`。

手动发布：Actions → **Deploy Tencent Cloud** → Run workflow。

## 6. Production `.env` highlights

```env
GATEWAY_PUBLIC_URL=https://api.example.com
OIDC_REDIRECT_URI=https://api.example.com/callback
GATEWAY_BFF_DEV_LOGIN=false
INTERNAL_API_TOKEN=<strong-random>

POSTGRES_HOST=<tencentdb-host>
POSTGRES_SSLMODE=require
REDIS_ADDR=<redis-host>:6379

S3_ENDPOINT=https://cos.<region>.myqcloud.com
```

## 7. WeChat mini-program

- Request 合法域名：`api.example.com`（仅域名，HTTPS）
- API：`https://api.example.com/v1/...`
- Login：`POST https://api.example.com/v1/auth/miniprogram/login`

## 8. Troubleshooting

| Issue | Check |
|-------|-------|
| TCR push 401 | `TCR_HOST` / 账号 ID / 固定密码 |
| CVM pull 慢 | TCR 与 CVM 同地域 |
| Gateway 502 business | `BUSINESS_SERVICE_URL=http://business-service:3005` in prod compose |
| OIDC redirect mismatch | Logto 应用回调与 `OIDC_REDIRECT_URI` 完全一致（https） |
| 直连 Nest 伪造身份 | 生产必须设 `INTERNAL_API_TOKEN` |

## 9. Optional upgrades

- **CLB** 做 HTTPS 终止，后端 CVM 只开 80
- **TKE** 替代单机 compose
- **CODING DevOps** 替代 GitHub Actions（同腾讯云账号）
