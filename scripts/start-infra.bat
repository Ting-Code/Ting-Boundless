@echo off
REM Start local Redis + RabbitMQ without admin (manual processes).
REM Run from repo root: scripts\start-infra.bat

echo Starting Redis...
start "Redis" /D "D:\app\Redis-8.6.2" cmd /k redis-server.exe redis.conf

set ERLANG_HOME=D:\app\Erlang OTP
set RABBIT_HOME=D:\app\rabbitmq\rabbitmq_server-4.3.1
if exist "%RABBIT_HOME%\sbin\rabbitmq-server.bat" (
    echo Starting RabbitMQ (ERLANG_HOME=%ERLANG_HOME%)...
    start "RabbitMQ" /D "%RABBIT_HOME%\sbin" cmd /k "set ERLANG_HOME=D:\app\Erlang OTP&& rabbitmq-server.bat"
) else (
    echo RabbitMQ not found at %RABBIT_HOME%
    echo Try admin PowerShell: Start-Service RabbitMQ
)

echo.
echo Wait a few seconds, then run: powershell -File scripts\verify-local.ps1
pause
