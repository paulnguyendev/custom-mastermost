# Mattermost Development Stop Script
# Usage: .\stop-dev.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Stopping Mattermost Development" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Kill processes on port 8065 (Backend)
Write-Host "[1/2] Stopping Backend Server (port 8065)..." -ForegroundColor Yellow
$port8065 = netstat -ano | findstr ":8065.*LISTENING"
if ($port8065) {
    $pids = ($port8065 | ForEach-Object { ($_ -split '\s+')[-1] }) | Select-Object -Unique
    foreach ($pid in $pids) {
        if ($pid -and $pid -ne "0") {
            taskkill /F /PID $pid 2>$null
            Write-Host "  Killed process $pid" -ForegroundColor Green
        }
    }
} else {
    Write-Host "  No process found on port 8065" -ForegroundColor Gray
}

# Kill processes on port 9005 (Webpack)
Write-Host "[2/2] Stopping Webpack Dev Server (port 9005)..." -ForegroundColor Yellow
$port9005 = netstat -ano | findstr ":9005.*LISTENING"
if ($port9005) {
    $pids = ($port9005 | ForEach-Object { ($_ -split '\s+')[-1] }) | Select-Object -Unique
    foreach ($pid in $pids) {
        if ($pid -and $pid -ne "0") {
            taskkill /F /PID $pid 2>$null
            Write-Host "  Killed process $pid" -ForegroundColor Green
        }
    }
} else {
    Write-Host "  No process found on port 9005" -ForegroundColor Gray
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Development Environment Stopped!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Optional: Stop Docker containers
$stopDocker = Read-Host "Stop Docker containers too? (y/N)"
if ($stopDocker -eq "y" -or $stopDocker -eq "Y") {
    Write-Host "Stopping Docker containers..." -ForegroundColor Yellow
    Push-Location "$PSScriptRoot\server"
    docker-compose -f docker-compose.makefile.yml down
    Pop-Location
    Write-Host "Docker containers stopped" -ForegroundColor Green
}

