# Install Erlang OTP 27.x for RabbitMQ 4.3 (Erlang 29 is unsupported).
# Usage: powershell -ExecutionPolicy Bypass -File scripts/install-erlang27.ps1

$ErrorActionPreference = "Stop"
$Target = "D:\app\Erlang-OTP-27"
$ZipUrl = "https://github.com/erlang/otp/releases/download/OTP-27.3.4/otp_win64_27.3.4.zip"
$ZipPath = Join-Path $Target "otp_win64_27.3.4.zip"

if (Test-Path "$Target\bin\erl.exe") {
    Write-Host "Erlang 27 already installed at $Target"
    & "$Target\bin\erl.exe" -eval 'erlang:display(erlang:system_info(otp_release)), halt().' -noshell
    exit 0
}

New-Item -ItemType Directory -Force -Path $Target | Out-Null
Write-Host "Downloading Erlang OTP 27.3.4..."
Invoke-WebRequest -Uri $ZipUrl -OutFile $ZipPath
Write-Host "Extracting to $Target..."
Expand-Archive -Path $ZipPath -DestinationPath $Target -Force
& "$Target\bin\erl.exe" -eval 'erlang:display(erlang:system_info(otp_release)), halt().' -noshell
Write-Host "Done. Use scripts\start-rabbitmq.bat to start RabbitMQ."
