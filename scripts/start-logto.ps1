# Start Logto (native, Windows) against local Postgres logto_db.
# Uses @logto/cli to install the prebuilt release (recommended on Windows).
#
# Usage (repo root):
#   powershell -ExecutionPolicy Bypass -File scripts/start-logto.ps1
#   powershell -ExecutionPolicy Bypass -File scripts/start-logto.ps1 -SeedOnly
#
# Requires: Node 22.x (nvm), Postgres logto_db, ports 3001/3002 free.
# See docs/LOGTO_SETUP.md for Admin Console app configuration.

param(
    [switch]$SeedOnly,
    [switch]$Reinstall
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$LogtoDir = Join-Path $RepoRoot "deploy\logto"
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

function Resolve-Node22 {
    $candidates = @(
        "$env:APPDATA\nvm\v22.14.0\node.exe",
        "$env:APPDATA\nvm\v22.13.0\node.exe",
        "$env:APPDATA\nvm\v22.12.0\node.exe"
    )
    foreach ($c in $candidates) {
        if (Test-Path $c) { return $c }
    }
    throw "Node 22 not found. Run: nvm install 22.14.0 && nvm use 22.14.0"
}

$NodeExe = Resolve-Node22
$NodeDir = Split-Path -Parent $NodeExe
$env:PATH = "$NodeDir;$env:PATH"
Write-Host "Using Node: $(& $NodeExe --version) at $NodeExe"

$dotenv = Load-DotEnv $EnvFile
$pgUser = if ($dotenv["POSTGRES_USER"]) { $dotenv["POSTGRES_USER"] } else { "ting" }
$pgPass = if ($dotenv["POSTGRES_PASSWORD"]) { $dotenv["POSTGRES_PASSWORD"] } else { "change-me" }
$pgHost = if ($dotenv["POSTGRES_HOST"]) { $dotenv["POSTGRES_HOST"] } else { "127.0.0.1" }
$pgPort = if ($dotenv["POSTGRES_PORT"]) { $dotenv["POSTGRES_PORT"] } else { "5432" }
$logtoDb = if ($dotenv["LOGTO_DB"]) { $dotenv["LOGTO_DB"] } else { "logto_db" }
$dbUrl = "postgresql://${pgUser}:${pgPass}@${pgHost}:${pgPort}/${logtoDb}"

if ($Reinstall -and (Test-Path $LogtoDir)) {
    Write-Host "Removing existing deploy/logto ..."
    Remove-Item -Recurse -Force $LogtoDir
}

if (-not (Test-Path (Join-Path $LogtoDir "package.json"))) {
    Write-Host "Installing Logto prebuilt release (downloads ~257MB; may take a while)..."
    Push-Location $RepoRoot
    try {
        & "$NodeDir\npx.cmd" @logto/cli init -p deploy/logto --db-url $dbUrl --dapc
    } finally {
        Pop-Location
    }
}

$logtoEnv = @"
DB_URL=$dbUrl
ENDPOINT=http://127.0.0.1:3001
ADMIN_ENDPOINT=http://127.0.0.1:3002
TRUST_PROXY_HEADER=1
"@
Set-Content -Path (Join-Path $LogtoDir ".env") -Value $logtoEnv -Encoding UTF8

Push-Location $LogtoDir
try {
    $env:CI = "true"
    Write-Host "Deploying database alterations..."
    & "$NodeDir\npm.cmd" run alteration deploy latest

    Write-Host "Seeding Logto database (if empty)..."
    & "$NodeDir\npm.cmd" run cli db seed -- --swe --disable-admin-pwned-password-check

    if ($SeedOnly) {
        Write-Host "Seed complete. Open http://127.0.0.1:3002 — see docs/LOGTO_SETUP.md"
        return
    }

    Write-Host "Starting Logto: OIDC :3001 | Admin :3002"
    & "$NodeDir\npm.cmd" start
}
finally {
    Pop-Location
}
