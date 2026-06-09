# Configure Logto default tenant for Gateway BFF (local dev).
# Creates API resource, Traditional Web App, and optional test user via Management API.
#
# Prerequisites: Logto running (deploy/logto-src), logto_db seeded, ting user has CREATEROLE.
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts/configure-logto-local.ps1
#   powershell -ExecutionPolicy Bypass -File scripts/configure-logto-local.ps1 -UpdateEnv

param(
    [switch]$UpdateEnv,
    [string]$ApiIdentifier = "https://api.ting-boundless.local",
    [string]$RedirectUri = "http://127.0.0.1:8080/callback"
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$EnvFile = Join-Path $RepoRoot ".env"

function Load-DotEnv([string]$Path) {
    if (-not (Test-Path $Path)) { return @{} }
    $vars = @{}
    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }
        $i = $line.IndexOf("=")
        if ($i -lt 1) { return }
        $vars[$line.Substring(0, $i).Trim()] = $line.Substring($i + 1).Trim()
    }
    return $vars
}

$dotenv = Load-DotEnv $EnvFile
$pgUser = if ($dotenv["POSTGRES_USER"]) { $dotenv["POSTGRES_USER"] } else { "ting" }
$pgPass = if ($dotenv["POSTGRES_PASSWORD"]) { $dotenv["POSTGRES_PASSWORD"] } else { "change-me" }
$pgHost = if ($dotenv["POSTGRES_HOST"]) { $dotenv["POSTGRES_HOST"] } else { "127.0.0.1" }
$pgPort = if ($dotenv["POSTGRES_PORT"]) { $dotenv["POSTGRES_PORT"] } else { "5432" }
$logtoDb = if ($dotenv["LOGTO_DB"]) { $dotenv["LOGTO_DB"] } else { "logto_db" }

$psql = "D:\app\PostgreSQL\bin\psql.exe"
if (-not (Test-Path $psql)) { throw "psql not found at $psql" }

$env:PGPASSWORD = $pgPass
$mDefault = & $psql -h $pgHost -U $pgUser -d $logtoDb -tAc "SELECT secret FROM applications WHERE id='m-default';"
if (-not $mDefault) { throw "m-default app not found; run Logto seed first." }

$tokenBody = @{
    grant_type    = "client_credentials"
    client_id     = "m-default"
    client_secret = $mDefault.Trim()
    resource      = "https://default.logto.app/api"
    scope         = "all"
}
$token = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:3002/oidc/token" `
    -ContentType "application/x-www-form-urlencoded" `
    -Body $tokenBody
$headers = @{ Authorization = "Bearer $($token.access_token)" }

$existing = Invoke-RestMethod -Uri "http://127.0.0.1:3001/api/resources" -Headers $headers
$resource = $existing | Where-Object { $_.indicator -eq $ApiIdentifier } | Select-Object -First 1
if (-not $resource) {
    $resource = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:3001/api/resources" `
        -Headers $headers -ContentType "application/json" `
        -Body (@{ name = "Ting Boundless API"; indicator = $ApiIdentifier } | ConvertTo-Json)
    Write-Host "Created API resource: $ApiIdentifier"
} else {
    Write-Host "API resource exists: $ApiIdentifier"
}

$apps = Invoke-RestMethod -Uri "http://127.0.0.1:3001/api/applications" -Headers $headers
$app = $apps | Where-Object { $_.name -eq "Gateway BFF" -and $_.type -eq "Traditional" } | Select-Object -First 1
if (-not $app) {
    $app = Invoke-RestMethod -Method Post -Uri "http://127.0.0.1:3001/api/applications" `
        -Headers $headers -ContentType "application/json" `
        -Body (@{
            name               = "Gateway BFF"
            type               = "Traditional"
            oidcClientMetadata = @{
                redirectUris         = @($RedirectUri)
                postLogoutRedirectUris = @()
            }
        } | ConvertTo-Json -Depth 5)
    Write-Host "Created Gateway BFF application"
} else {
    Write-Host "Gateway BFF application exists: $($app.id)"
}

$secrets = Invoke-RestMethod -Uri "http://127.0.0.1:3001/api/applications/$($app.id)/secrets" -Headers $headers
$secret = $secrets[0].value

Write-Host ""
Write-Host "OIDC_CLIENT_ID=$($app.id)"
Write-Host "OIDC_CLIENT_SECRET=$secret"
Write-Host "OIDC_RESOURCE=$ApiIdentifier"
Write-Host "GATEWAY_BFF_DEV_LOGIN=false"

if ($UpdateEnv -and (Test-Path $EnvFile)) {
    $content = Get-Content $EnvFile -Raw
    $content = $content -replace 'OIDC_CLIENT_ID=.*', "OIDC_CLIENT_ID=$($app.id)"
    $content = $content -replace 'OIDC_CLIENT_SECRET=.*', "OIDC_CLIENT_SECRET=$secret"
    if ($content -notmatch 'OIDC_RESOURCE=') {
        $content = $content -replace '(OIDC_AUDIENCE=.*)', "`$1`nOIDC_RESOURCE=$ApiIdentifier"
    } else {
        $content = $content -replace 'OIDC_RESOURCE=.*', "OIDC_RESOURCE=$ApiIdentifier"
    }
    $content = $content -replace 'GATEWAY_BFF_DEV_LOGIN=.*', 'GATEWAY_BFF_DEV_LOGIN=false'
    Set-Content -Path $EnvFile -Value $content -Encoding UTF8
    Write-Host "Updated $EnvFile"
}
