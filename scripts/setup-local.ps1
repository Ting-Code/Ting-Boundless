# Bootstrap local PostgreSQL for Ting Boundless (Windows).
# Usage (from repo root):
#   powershell -ExecutionPolicy Bypass -File scripts/setup-local.ps1
#   powershell -ExecutionPolicy Bypass -File scripts/setup-local.ps1 -Password "your-postgres-password"
#
# Or set POSTGRES_ADMIN_PASSWORD in .env (gitignored) for non-interactive runs.

param(
    [string]$Password = ""
)

$ErrorActionPreference = "Stop"
$RepoRoot = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$SqlFile = Join-Path $RepoRoot "deploy\postgres\setup-local.sql"
$EnvFile = Join-Path $RepoRoot ".env"

function Load-DotEnv([string]$Path) {
    if (-not (Test-Path $Path)) { return @{} }
    $vars = @{}
    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if ($line -eq "" -or $line.StartsWith("#")) { return }
        $i = $line.IndexOf("=")
        if ($i -lt 1) { return }
        $k = $line.Substring(0, $i).Trim()
        $v = $line.Substring($i + 1).Trim()
        $vars[$k] = $v
    }
    return $vars
}

$PsqlCandidates = @(
    "D:\app\PostgreSQL\bin\psql.exe",
    "C:\Program Files\PostgreSQL\18\bin\psql.exe",
    "C:\Program Files\PostgreSQL\16\bin\psql.exe",
    "C:\Program Files\PostgreSQL\17\bin\psql.exe"
)

$Psql = $null
foreach ($c in $PsqlCandidates) {
    if (Test-Path $c) { $Psql = $c; break }
}
if (-not $Psql) {
    $found = Get-Command psql -ErrorAction SilentlyContinue
    if ($found) { $Psql = $found.Source }
}
if (-not $Psql) {
    Write-Error "psql not found. Install PostgreSQL or add bin to PATH."
}

if (-not $Password) {
    $Password = $env:POSTGRES_ADMIN_PASSWORD
}
if (-not $Password) {
    $dotenv = Load-DotEnv $EnvFile
    if ($dotenv.ContainsKey("POSTGRES_ADMIN_PASSWORD")) {
        $Password = $dotenv["POSTGRES_ADMIN_PASSWORD"]
    }
}

if (-not $Password) {
    Write-Host "Using: $Psql"
    Write-Host "SQL:   $SqlFile"
    Write-Host ""
    Write-Host "Tip: add POSTGRES_ADMIN_PASSWORD to .env to skip this prompt."
    Write-Host "Enter postgres superuser password (input hidden):"
    $secure = Read-Host -AsSecureString
    $bstr = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
    $Password = [Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
    [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
}

Write-Host "Using: $Psql"
Write-Host "SQL:   $SqlFile"

$env:PGPASSWORD = $Password
& $Psql -h 127.0.0.1 -U postgres -f $SqlFile
$code = $LASTEXITCODE
Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue

if ($code -ne 0) {
    Write-Error "setup-local.sql failed (exit $code). Check postgres superuser password."
}

Write-Host ""
Write-Host "Done. Ensure .env matches:"
Write-Host "  POSTGRES_USER=ting"
Write-Host "  POSTGRES_PASSWORD=change-me"
Write-Host "  POSTGRES_HOST=localhost"
Write-Host ""
Write-Host "Verify: powershell -File scripts/verify-local.ps1"
