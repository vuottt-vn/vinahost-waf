# WAF Application Startup Script
# This script builds and runs the WAF application with SQLite

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Web Application Firewall - Startup" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
$goPath = Get-Command go -ErrorAction SilentlyContinue
if (-not $goPath) {
    # Try common locations
    $possiblePaths = @(
        "C:\Program Files\Go\bin\go.exe",
        "C:\Go\bin\go.exe",
        "$env:LOCALAPPDATA\go\bin\go.exe"
    )
    
    foreach ($path in $possiblePaths) {
        if (Test-Path $path) {
            $goPath = $path
            break
        }
    }
}

if (-not $goPath) {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "After installation, restart your terminal and run this script again." -ForegroundColor Yellow
    Write-Host ""
    pause
    exit 1
}

Write-Host "✓ Go found: $goPath" -ForegroundColor Green
Write-Host ""

# Set environment variables for SQLite
$env:DB_PATH = "waf.db"
$env:WAF_CRS_ENABLED = "false"  # Disable CRS until it's downloaded
$env:RATE_LIMIT_ENABLED = "true"
$env:RATE_LIMIT_MAX_REQUESTS = "100"
$env:RATE_LIMIT_WINDOW_SECONDS = "60"

Write-Host "Configuration:" -ForegroundColor Cyan
Write-Host "  Database: SQLite ($env:DB_PATH)" -ForegroundColor White
Write-Host "  Rate Limiting: Enabled (100 req/min)" -ForegroundColor White
Write-Host "  OWASP CRS: Disabled (run scripts\setup-crs.bat to enable)" -ForegroundColor White
Write-Host ""

# Build WAF Proxy
Write-Host "Building WAF Proxy..." -ForegroundColor Yellow
& go build -o waf.exe ./cmd/waf
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to build WAF Proxy" -ForegroundColor Red
    pause
    exit 1
}
Write-Host "✓ WAF Proxy built successfully" -ForegroundColor Green
Write-Host ""

# Build WAF API
Write-Host "Building WAF API Server..." -ForegroundColor Yellow
& go build -o waf-api.exe ./cmd/api
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to build WAF API Server" -ForegroundColor Red
    pause
    exit 1
}
Write-Host "✓ WAF API Server built successfully" -ForegroundColor Green
Write-Host ""

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "Starting WAF Services" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# Start WAF Proxy in background
Write-Host "Starting WAF Proxy on port 8080..." -ForegroundColor Yellow
$wafProcess = Start-Process -FilePath ".\waf.exe" -PassThru -WindowStyle Normal
Write-Host "✓ WAF Proxy started (PID: $($wafProcess.Id))" -ForegroundColor Green
Write-Host ""

# Wait a moment for proxy to start
Start-Sleep -Seconds 2

# Start WAF API in background
Write-Host "Starting WAF API on port 3001..." -ForegroundColor Yellow
$apiProcess = Start-Process -FilePath ".\waf-api.exe" -PassThru -WindowStyle Normal
Write-Host "✓ WAF API Server started (PID: $($apiProcess.Id))" -ForegroundColor Green
Write-Host ""

Write-Host "======================================" -ForegroundColor Green
Write-Host "WAF Services Running!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host ""
Write-Host "WAF Proxy:  http://localhost:8080" -ForegroundColor White
Write-Host "WAF API:    http://localhost:3001" -ForegroundColor White
Write-Host "Dashboard:  http://localhost:5173  (if frontend is running)" -ForegroundColor White
Write-Host ""
Write-Host "Default Login:" -ForegroundColor Yellow
Write-Host "  Username: admin" -ForegroundColor White
Write-Host "  Password: admin123" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to stop all services..." -ForegroundColor Cyan
pause

# Stop services
Write-Host ""
Write-Host "Stopping WAF services..." -ForegroundColor Yellow
Stop-Process -Id $wafProcess.Id -Force -ErrorAction SilentlyContinue
Stop-Process -Id $apiProcess.Id -Force -ErrorAction SilentlyContinue
Write-Host "✓ All services stopped" -ForegroundColor Green
