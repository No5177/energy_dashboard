@echo off
echo ===============================================
echo 台達電表 Modbus TCP 通訊測試
echo ===============================================
echo.
echo 電表資訊:
echo IP: 192.168.1.9
echo 子網路遮罩: 255.255.255.0
echo 閘道: 192.168.1.1
echo 通訊位址: 2
echo.
echo 開始編譯並執行測試程式...
echo.

REM 編譯 Modbus 測試程式
go build -o modbus_test.exe modbus_client.go

if %ERRORLEVEL% NEQ 0 (
    echo ❌ 編譯失敗！
    echo 請確認 Go 環境已正確安裝
    pause
    exit /b 1
)

echo ✅ 編譯成功
echo.
echo 執行測試程式...
echo.

REM 執行測試程式
modbus_test.exe

echo.
echo 測試完成，按任意鍵關閉視窗...
pause > nul

REM 清理編譯產生的檔案
del modbus_test.exe 