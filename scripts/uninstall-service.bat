@echo off
chcp 65001 >nul
:: FileHub — Remove Windows Service
:: Requires Administrator privileges
:: Usage: 右键 → 以管理员身份运行

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo 请以管理员身份运行此脚本。
    echo 右键 uninstall-service.bat → 以管理员身份运行
    pause
    exit /b 1
)

set SERVICE_NAME=FileHub

echo ============================================
echo   FileHub Windows Service 卸载
echo ============================================
echo.

sc query %SERVICE_NAME% >nul 2>&1
if %errorlevel% neq 0 (
    echo 服务 %SERVICE_NAME% 不存在，无需卸载。
    pause
    exit /b 0
)

echo 正在停止服务...
sc stop %SERVICE_NAME% >nul 2>&1
timeout /t 2 /nobreak >nul

echo 正在删除服务...
sc delete %SERVICE_NAME%

if %errorlevel% equ 0 (
    echo [成功] 服务已卸载
) else (
    echo [失败] 卸载失败
)

echo.
pause
