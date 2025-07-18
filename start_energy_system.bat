@echo off
chcp 65001
title 能源監控系統 v2.0

echo ==================================================
echo           能源監控系統 v2.0
echo        Modbus TCP + SQLite3 + Chart.js
echo ==================================================
echo.

REM 檢查 Go 環境
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo ❌ 錯誤: 找不到 Go 執行環境
    echo 請先安裝 Go 語言：https://golang.org/dl/
    pause
    exit /b 1
)

echo ✅ 檢查 Go 環境... 完成

REM 檢查必要檔案
if not exist "energy_backend.go" (
    echo ❌ 錯誤: 找不到 energy_backend.go 檔案
    pause
    exit /b 1
)

if not exist "energy_dashboard.html" (
    echo ❌ 錯誤: 找不到 energy_dashboard.html 檔案
    pause
    exit /b 1
)

echo ✅ 檢查檔案... 完成

REM 下載相依套件
echo 📦 正在下載相依套件...
go mod tidy
if %ERRORLEVEL% neq 0 (
    echo ❌ 相依套件下載失敗
    pause
    exit /b 1
)

echo ✅ 相依套件... 完成

REM 編譯程式
echo 🔨 正在編譯 Go 程式...
set CGO_ENABLED=1
set GOOS=windows
set GOARCH=amd64

go build -ldflags="-s -w" -o energy_system.exe energy_backend.go
if %ERRORLEVEL% neq 0 (
    echo ❌ 編譯失敗
    pause
    exit /b 1
)

echo ✅ 編譯... 完成

echo.
echo ==================================================
echo 🚀 啟動能源監控系統
echo ==================================================
echo.
echo 系統功能:
echo • 🔄 每 5 秒自動收集電表資料 (Modbus TCP)
echo • 💾 資料儲存至 SQLite3 資料庫
echo • 📊 網頁即時監控儀表板
echo • 📈 多時間軸曲線圖 (每日/每月/每季/每年)
echo • 🌐 HTTP API 服務 (localhost:8080)
echo.
echo 注意事項:
echo 1. 確保電表 (192.168.1.9:502) 可正常連線
echo 2. 系統將自動開啟瀏覽器
echo 3. 按 Ctrl+C 可停止系統
echo.

REM 啟動系統
energy_system.exe

echo.
echo 系統已停止，按任意鍵退出...
pause >nul 