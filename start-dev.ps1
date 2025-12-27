# Mattermost Development Start Script
# Usage: .\start-dev.ps1

$ErrorActionPreference = "Continue"
$ROOT_DIR = $PSScriptRoot

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Mattermost Development Environment" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker
Write-Host "[1/4] Checking Docker..." -ForegroundColor Yellow
$dockerStatus = docker info 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker is not running. Please start Docker Desktop first." -ForegroundColor Red
    exit 1
}
Write-Host "  Docker is running" -ForegroundColor Green

# Check if ports are free
Write-Host "[2/4] Checking ports..." -ForegroundColor Yellow

$port8065 = netstat -ano | findstr ":8065.*LISTENING"
if ($port8065) {
    Write-Host "  Port 8065 is in use. Killing existing process..." -ForegroundColor Yellow
    $pid8065 = ($port8065 -split '\s+')[-1]
    taskkill /F /PID $pid8065 2>$null
    Start-Sleep -Seconds 2
}

$port9005 = netstat -ano | findstr ":9005.*LISTENING"
if ($port9005) {
    Write-Host "  Port 9005 is in use. Killing existing process..." -ForegroundColor Yellow
    $pid9005 = ($port9005 -split '\s+')[-1]
    taskkill /F /PID $pid9005 2>$null
    Start-Sleep -Seconds 2
}
Write-Host "  Ports 8065 and 9005 are available" -ForegroundColor Green

# Start Backend Server
Write-Host "[3/4] Starting Backend Server (port 8065)..." -ForegroundColor Yellow
$serverPath = Join-Path $ROOT_DIR "server"
Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "run-server.bat" -WorkingDirectory $serverPath -WindowStyle Normal
Write-Host "  Backend server starting in new window..." -ForegroundColor Green

# Wait for server to be ready
Write-Host "  Waiting for server to be ready..." -ForegroundColor Gray
$maxRetries = 30
$retries = 0
do {
    Start-Sleep -Seconds 2
    $retries++
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8065/api/v4/system/ping" -TimeoutSec 2 -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            Write-Host "  Backend server is ready!" -ForegroundColor Green
            break
        }
    } catch {
        Write-Host "  Still waiting... ($retries/$maxRetries)" -ForegroundColor Gray
    }
} while ($retries -lt $maxRetries)

if ($retries -ge $maxRetries) {
    Write-Host "  Warning: Server may not be fully ready yet" -ForegroundColor Yellow
}

# Start Webpack Dev Server
Write-Host "[4/4] Starting Webpack Dev Server (port 9005)..." -ForegroundColor Yellow
$webappPath = Join-Path $ROOT_DIR "webapp\channels"
Start-Process -FilePath "cmd.exe" -ArgumentList "/c", "npm run dev-server" -WorkingDirectory $webappPath -WindowStyle Normal
Write-Host "  Webpack dev server starting in new window..." -ForegroundColor Green

# Done
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Development Environment Started!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Frontend:  http://localhost:9005" -ForegroundColor White
Write-Host "  Backend:   http://localhost:8065" -ForegroundColor White
Write-Host "  Console:   http://localhost:9005/admin_console" -ForegroundColor White
Write-Host ""
Write-Host "  To stop: .\stop-dev.ps1" -ForegroundColor Gray
Write-Host ""

# Open browser
$openBrowser = Read-Host "Open browser now? (Y/n)"
if ($openBrowser -ne "n" -and $openBrowser -ne "N") {
    Start-Sleep -Seconds 5
    Start-Process "http://localhost:9005"
}

