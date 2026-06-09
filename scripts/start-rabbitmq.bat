@echo off
REM Start RabbitMQ with Erlang 27 (RabbitMQ 4.3.x does NOT support Erlang 29).
REM Run from repo root: scripts\start-rabbitmq.bat

set ERLANG_HOME=D:\app\Erlang-OTP-27
set RABBIT_HOME=D:\app\rabbitmq\rabbitmq_server-4.3.1

if not exist "%ERLANG_HOME%\bin\erl.exe" (
    echo Erlang 27 not found at %ERLANG_HOME%
    echo Run scripts\install-erlang27.ps1 or see README.
    exit /b 1
)

echo ERLANG_HOME=%ERLANG_HOME%
cd /d "%RABBIT_HOME%\sbin"
call rabbitmq-server.bat
