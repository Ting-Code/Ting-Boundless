# E2E: Bearer JWT (mobile API path) -> Gateway -> /v1/users/me + /v1/business/me
# Uses GATEWAY_DEV_JWT_SECRET dev token — NOT a substitute for Logto PKCE integration tests.
# See docs/MOBILE_AUTH.md for the full native OIDC flow.
$ErrorActionPreference = "Stop"
$gateway = if ($env:GATEWAY_URL) { $env:GATEWAY_URL } else { "http://127.0.0.1:8080" }
$repoRoot = Split-Path -Parent $PSScriptRoot
$goDir = Join-Path $repoRoot "go"

Write-Host "1. Issue dev Bearer JWT (GATEWAY_DEV_JWT_SECRET)..."
$token = & go -C $goDir run ./cmd/dev-jwt 2>&1
if ($LASTEXITCODE -ne 0 -or -not $token) {
  throw "dev-jwt failed: $token"
}
$token = $token.Trim()

Write-Host "2. GET /v1/users/me (Bearer)..."
$usersMe = curl.exe -sS "$gateway/v1/users/me" -H "Authorization: Bearer $token"
Write-Host $usersMe
if ($usersMe -notmatch '"user_id"') {
  throw "expected user_id in /v1/users/me response"
}

Write-Host "3. GET /v1/business/me (Bearer)..."
$bizMe = curl.exe -sS "$gateway/v1/business/me" -H "Authorization: Bearer $token"
Write-Host $bizMe
if ($bizMe -notmatch '"user_id"') {
  throw "expected user_id in /v1/business/me response"
}

Write-Host "OK: mobile Bearer path via Gateway"
