# @ting/admin

Vite + React **Web 后台** SPA (`basename` `/admin`). TanStack Query → Gateway `/v1` with
`credentials: include`（Gateway BFF HttpOnly cookie）。Paths、types、`apiFetch` 来自 `@ting/api`。

> V1 只做 Web 后台；小程序 / App 客户端不在 `node/` 范围内（见 `node/README.md` § V1 scope）。

## Run (dev)

Gateway、Redis、business-service 需已启动。Vite 将 `/v1`、`/sign-in` 代理到 Gateway。

```bash
# repo root
make run-gateway
make run-business
make run-admin
# http://localhost:5173/admin/items
```

**开发登录（无 Logto）：**「开发环境登录」或 `/sign-in/dev?return_to=/admin/items`。  
默认角色 `user,admin`（可访问 **审计** 页）；`.env` 需 `GATEWAY_BFF_DEV_LOGIN=true` 与 Redis。

**Logto 生产路径：**见 [`docs/BFF_LOGTO.md`](../../docs/BFF_LOGTO.md)（调用 Gateway `/sign-in`，无需在 Admin 写 OIDC）。

完整联调见 [`docs/E2E_ADMIN.md`](../../docs/E2E_ADMIN.md)、`make e2e-admin`。

## Pages (V1)

| Route | API | 说明 |
|-------|-----|------|
| `/admin/` | `businessPaths.ping` + 统计 | 概览：服务状态、条目/文件/用户计数 |
| `/admin/items` | `businessPaths.*` | 业务条目 CRUD + 详情抽屉 |
| `/admin/files` | `filePaths.*` | 上传 / 列表 / 下载 / 删除 / 详情抽屉 |
| `/admin/account` | `businessPaths.me`, `userPaths.me` | 身份 + 显示名称 |
| `/admin/audit` | `auditPaths.events` | 审计事件列表 + 详情抽屉（需 admin 角色） |
| `/admin/users` | `userPaths.list` | 当前租户用户列表（需 admin 角色） |

未登录时跳转 Gateway `/sign-in`（401 由 `useApiQuery` + 全局 QueryCache 自动跳转）。

## 开发说明

- `src/hooks/useApiQuery.ts` — 带 `authReturnTo` 的 TanStack Query 封装
- `src/hooks/useAuthMutation.ts` — mutation 401 自动跳转 sign-in
- `src/lib/queryClient.ts` — 401 全局重定向至 BFF sign-in

## OpenAPI 域

后台使用的契约：`business.v1`、`users.v1`、`files.v1`、`audit.v1`（见 `platform-contracts/openapi/`）。  
改 API 后执行 `make generate-api`。
