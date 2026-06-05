@echo off
REM Bootstrap local PostgreSQL. Run from repo root:
REM   scripts\setup-local.bat
cd /d %~dp0..
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0setup-local.ps1"
pause
