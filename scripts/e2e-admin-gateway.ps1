# E2E: Gateway dev cookie login -> /v1/business/items (no browser).
# Requires: Redis, make run-gateway, make run-business, Postgres with app_db.
$ErrorActionPreference = "Stop"
$gateway = if ($env:GATEWAY_URL) { $env:GATEWAY_URL } else { "http://127.0.0.1:8080" }
$jar = Join-Path $env:TEMP "ting-e2e-cookies.txt"

if (Test-Path $jar) { Remove-Item $jar -Force }

Write-Host "0. Gateway -> business-service ping..."
$ping = curl.exe -sS "$gateway/v1/business/ping"
Write-Host $ping
if ($ping -notmatch 'business-service') {
  throw @"
Gateway /v1/business/ping failed ($ping).
Restart gateway after setting in .env:
  BUSINESS_SERVICE_URL=http://127.0.0.1:3005
(Logto uses :3001 — do not point business upstream there.)
"@
}

Write-Host "1. Dev sign-in (sets tb_session cookie)..."
curl.exe -sS -c $jar -L "$gateway/sign-in/dev?return_to=/admin/items&user_id=e2e-user&tenant_id=e2e-tenant" -o NUL

Write-Host "2. GET /v1/business/me"
$me = curl.exe -sS -b $jar "$gateway/v1/business/me"
Write-Host $me
if ($me -notmatch '"user_id":"e2e-user"') {
  throw "expected user_id e2e-user in /v1/business/me"
}

Write-Host "3. POST /v1/business/items"
$bodyFile = Join-Path $env:TEMP "ting-e2e-item.json"
Set-Content -Path $bodyFile -Value '{"title":"e2e-item","body":"from script"}' -NoNewline -Encoding utf8
$created = curl.exe -sS -b $jar -H "Content-Type: application/json" --data-binary "@$bodyFile" -X POST "$gateway/v1/business/items"
Write-Host $created
if ($created -notmatch '"title":"e2e-item"') {
  throw "create item failed"
}

Write-Host "4. GET /v1/business/items"
$list = curl.exe -sS -b $jar "$gateway/v1/business/items"
Write-Host $list
if ($list -notmatch 'e2e-item') {
  throw "list should contain e2e-item"
}

Write-Host "OK: Gateway cookie -> business-service E2E passed."
