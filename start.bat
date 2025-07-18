@echo off
chcp 65001
title 能源看板系統

echo ==================================================
echo           能源看板系統 v1.0
echo ==================================================
echo.

REM 檢查檔案是否存在
if not exist "web_server.exe" (
    echo 錯誤: 找不到 web_server.exe 檔案
    echo 請先執行 build.bat 編譯程式
    pause
    exit /b 1
)

if not exist "energy_dashboard.html" (
    echo 錯誤: 找不到 energy_dashboard.html 檔案
    pause
    exit /b 1
)

echo 正在啟動能源看板系統...
echo.
echo 注意事項:
echo 1. 請確保 LabVIEW 已在 port 8888 啟動
echo 2. 系統將在 port 5177 提供網頁服務
echo 3. 按 Ctrl+C 可停止系統
echo.

REM 啟動系統
web_server.exe 