# E2E: WeChat mini-program login (mock) -> Bearer JWT -> /v1/users/me.
# Requires: Postgres app_db, auth-service, user-service, gateway with AUTH_JWKS_URL.
$ErrorActionPreference = "Stop"
$gateway = if ($env:GATEWAY_URL) { $env:GATEWAY_URL } else { "http://127.0.0.1:8080" }

function Invoke-Login($code) {
  $bodyFile = Join-Path $env:TEMP "ting-e2e-mp-login.json"
  $json = @{ code = $code } | ConvertTo-Json -Compress
  Set-Content -Path $bodyFile -Value $json -NoNewline -Encoding utf8
  $raw = curl.exe -sS -H "Content-Type: application/json" --data-binary "@$bodyFile" -X POST "$gateway/v1/auth/miniprogram/login"
  if ($raw -match '"code"\s*:\s*"') {
    throw "miniprogram login failed: $raw"
  }
  return $raw | ConvertFrom-Json
}

Write-Host "0. Gateway JWKS (auth-service via anon /v1/auth/)..."
$jwks = curl.exe -sS "$gateway/v1/auth/jwks"
Write-Host $jwks
if ($jwks -notmatch '"keys"') {
  throw @"
Gateway /v1/auth/jwks failed.
Ensure auth-service is running (:8084) and .env has:
  WECHAT_MOCK_MODE=true
  AUTH_OIDC_ISSUER=http://127.0.0.1:8084/oidc
  AUTH_JWKS_URL=http://127.0.0.1:8084/v1/auth/jwks
Restart gateway after updating .env.
"@
}

Write-Host "1. Mini-program login (mock openid A + unionid)..."
$loginA = Invoke-Login "e2e_mp_a|e2e_union_1"
Write-Host ($loginA | ConvertTo-Json -Compress)
if (-not $loginA.access_token) {
  throw "expected access_token"
}

Write-Host "2. Second mini-program (mock openid B, same unionid)..."
$loginB = Invoke-Login "e2e_mp_b|e2e_union_1"
Write-Host ($loginB | ConvertTo-Json -Compress)
if ($loginA.user_id -ne $loginB.user_id) {
  throw "unionid binding failed: user_id $($loginA.user_id) vs $($loginB.user_id)"
}

Write-Host "3. GET /v1/users/me with Bearer token..."
$me = curl.exe -sS -H "Authorization: Bearer $($loginA.access_token)" "$gateway/v1/users/me"
Write-Host $me
$expected = '"' + $loginA.user_id + '"'
if ($me -notmatch $expected) {
  throw "expected user_id $($loginA.user_id) in /v1/users/me"
}

Write-Host "OK: Mini-program mock login + unionid bind + Gateway JWT E2E passed."
