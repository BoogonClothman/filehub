@echo off
chcp 65001 >nul
:: FileHub — Register as Windows Service
:: Requires Administrator privileges
:: Usage: 右键 → 以管理员身份运行

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo 请以管理员身份运行此脚本。
    echo 右键 install-service.bat → 以管理员身份运行
    pause
    exit /b 1
)

set SERVICE_NAME=FileHub
set BIN_PATH=%~dp0..\filehub.exe
set DISPLAY_NAME=FileHub — LAN File Sharing

echo ============================================
echo   FileHub Windows Service 安装
echo ============================================
echo.
echo 服务名称: %SERVICE_NAME%
echo 可执行文件: %BIN_PATH%
echo.

if not exist "%BIN_PATH%" (
    echo [错误] 找不到 filehub.exe
    echo 请先在项目根目录执行: go build -o filehub.exe .
    pause
    exit /b 1
)

sc query %SERVICE_NAME% >nul 2>&1
if %errorlevel% equ 0 (
    echo [警告] 服务已存在，正在删除旧服务...
    sc stop %SERVICE_NAME% >nul 2>&1
    sc delete %SERVICE_NAME% >nul 2>&1
    timeout /t 2 /nobreak >nul
)

sc create %SERVICE_NAME% ^
    binPath= "\"%BIN_PATH%\"" ^
    start= auto ^
    DisplayName= "%DISPLAY_NAME%"

if %errorlevel% equ 0 (
    echo [成功] 服务已注册

    sc description %SERVICE_NAME% "LAN file sharing hub — browse, upload, download, manage files"
    sc failure %SERVICE_NAME% reset= 86400 actions= restart/5000/restart/5000/restart/5000

    echo.
    echo 正在启动服务...
    sc start %SERVICE_NAME%
    if %errorlevel% equ 0 (
        echo [成功] 服务已启动
        echo.
        echo 浏览器打开 http://localhost:5000 即可使用
    ) else (
        echo [失败] 服务启动失败，请检查日志
    )
) else (
    echo [失败] 服务注册失败
)

echo.
pause
