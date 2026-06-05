# Check local infra + Ting Boundless service readiness.
# Usage: powershell -ExecutionPolicy Bypass -File scripts/verify-local.ps1

$ErrorActionPreference = "Continue"
$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
Set-Location $RepoRoot

function Test-TcpPort([string]$HostName, [int]$Port) {
    try {
        $c = New-Object System.Net.Sockets.TcpClient
        $iar = $c.BeginConnect($HostName, $Port, $null, $null)
        $ok = $iar.AsyncWaitHandle.WaitOne(2000, $false)
        if ($ok -and $c.Connected) { $c.Close(); return $true }
        $c.Close()
        return $false
    } catch { return $false }
}

Write-Host "=== Local infrastructure ===" -ForegroundColor Cyan

$redisCli = "D:\app\Redis-8.6.2\redis-cli.exe"
if (Test-Path $redisCli) {
    $pong = & $redisCli ping 2>$null
    if ($pong -eq "PONG") { Write-Host "[OK]   Redis (6379)" -ForegroundColor Green }
    else { Write-Host "[FAIL] Redis not responding" -ForegroundColor Red }
} elseif (Test-TcpPort "127.0.0.1" 6379) {
    Write-Host "[OK]   Port 6379 open (redis-cli not at default path)" -ForegroundColor Yellow
} else {
    Write-Host "[FAIL] Redis (6379) — start D:\app\Redis-8.6.2\start.bat" -ForegroundColor Red
}

if (Test-TcpPort "127.0.0.1" 5432) {
    $psql = "D:\app\PostgreSQL\bin\psql.exe"
    if (Test-Path $psql) {
        $env:PGPASSWORD = "change-me"
        & $psql -h 127.0.0.1 -U ting -d app_db -tAc "SELECT 1" 2>$null | Out-Null
        $pgOk = ($LASTEXITCODE -eq 0)
        Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
        if ($pgOk) { Write-Host "[OK]   PostgreSQL ting@app_db" -ForegroundColor Green }
        else { Write-Host "[FAIL] PostgreSQL up but ting/app_db not ready — run scripts/setup-local.bat" -ForegroundColor Red }
    } else {
        Write-Host "[WARN] PostgreSQL port open; psql not found for auth check" -ForegroundColor Yellow
    }
} else {
    Write-Host "[FAIL] PostgreSQL (5432)" -ForegroundColor Red
}

if (Test-TcpPort "127.0.0.1" 5672) {
    Write-Host "[OK]   RabbitMQ (5672)" -ForegroundColor Green
} else {
    Write-Host "[SKIP] RabbitMQ (5672) — Start-Service RabbitMQ (admin) or scripts/start-infra.bat" -ForegroundColor Yellow
}

if (Test-TcpPort "127.0.0.1" 9000) {
    Write-Host "[WARN] Port 9000 open (may not be MinIO — check S3_ENDPOINT)" -ForegroundColor Yellow
} else {
    Write-Host "[SKIP] MinIO (9000) — optional for file-service" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== Next steps ===" -ForegroundColor Cyan
Write-Host "1. If PostgreSQL failed: add POSTGRES_ADMIN_PASSWORD=... to .env, then scripts/setup-local.bat"
Write-Host "2. go run ./services/user-service  &&  curl http://127.0.0.1:8081/readyz"
Write-Host "3. go run ./services/gateway       &&  curl http://127.0.0.1:8080/readyz"
