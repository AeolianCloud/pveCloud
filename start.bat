@echo off
chcp 65001 >nul 2>&1
title PveCloud Dev Launcher

echo ========================================
echo   PveCloud Development Launcher
echo ========================================
echo.

:: Start Go server (3 services in one terminal)
echo [1/3] Starting Go backend...
start "pvecloud-server" cmd /k "cd /d D:\UGit\pveCloud\server && echo === Public API (:8080) === && go run ./cmd/public-api"
timeout /t 1 /nobreak >nul
start "pvecloud-admin-api" cmd /k "cd /d D:\UGit\pveCloud\server && echo === Admin API (:8081) === && go run ./cmd/admin-api"
timeout /t 1 /nobreak >nul
start "pvecloud-worker" cmd /k "cd /d D:\UGit\pveCloud\server && echo === Worker (:8082) === && go run ./cmd/worker"

:: Start web frontend
echo [2/3] Starting web frontend...
start "pvecloud-web" cmd /k "cd /d D:\UGit\pveCloud\web && npm run dev"

:: Start admin frontend
echo [3/3] Starting admin frontend...
start "pvecloud-admin" cmd /k "cd /d D:\UGit\pveCloud\admin && npm run dev"

echo.
echo ========================================
echo   All 5 services launched!
echo   - Public API  : http://localhost:8080
echo   - Admin API   : http://localhost:8081
echo   - Worker      : http://localhost:8082
echo   - Web         : http://localhost:5173
echo   - Admin       : http://localhost:5174
echo ========================================
echo.
echo Press any key to exit this window (services will keep running)...
pause >nul
