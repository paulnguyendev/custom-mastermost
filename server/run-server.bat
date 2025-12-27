@echo off
set PATH=C:\Program Files\Go\bin;%PATH%
cd /d %~dp0

echo Checking Go version...
go version
if errorlevel 1 (
    echo ERROR: Go is not installed or not in PATH
    pause
    exit /b 1
)
echo.

echo Checking Docker...
docker info >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not running. Please start Docker Desktop first.
    pause
    exit /b 1
)
echo Docker is running.
echo.

echo Starting Docker dependencies (PostgreSQL, etc.)...
docker-compose -f docker-compose.makefile.yml up -d
echo.

echo Waiting for database to be ready...
timeout /t 5 /nobreak >nul

echo Starting Mattermost Server...
go run ./cmd/mattermost server

