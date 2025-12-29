@echo off
echo.
echo ========================================
echo   CODEX Build Script
echo ========================================
echo.

echo Building CODEX executable...
go build -o codex.exe main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo   Build successful!
    echo ========================================
    echo.
    echo Run CODEX with: codex.exe
    echo.
) else (
    echo.
    echo ========================================
    echo   Build failed!
    echo ========================================
    echo.
    exit /b 1
)
