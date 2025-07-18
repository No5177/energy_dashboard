@echo off
chcp 65001
echo 正在編譯 Go 程式...

REM 設定環境變數
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

REM 編譯程式
go build -ldflags="-s -w" -o web_server.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo 編譯成功！產生檔案: web_server.exe
    echo 檔案大小:
    dir web_server.exe | find "web_server.exe"
) else (
    echo 編譯失敗！
    pause
    exit /b 1
)

echo.
echo 編譯完成！
pause 