@echo off
echo ========================================
echo WAF Local Deployment
echo ========================================
echo.

REM Check if .env exists
if not exist .env (
    echo ERROR: .env file not found!
    echo Please create .env file first.
    pause
    exit /b 1
)

echo Building and starting WAF services...
echo.

REM Stop any existing containers
docker compose down

REM Build and start all services
docker compose up -d --build

if %errorlevel% neq 0 (
    echo.
    echo ERROR: Docker build failed!
    echo Check the error messages above.
    pause
    exit /b 1
)

echo.
echo ========================================
echo WAF Services Started!
echo ========================================
echo.
echo WAF Frontend (Web UI):  http://localhost:3000
echo WAF API:                http://localhost:3001
echo WAF Proxy:              http://localhost:8080
echo.
echo Default Login:
echo   Username: admin
echo   Password: admin123
echo.
echo View logs:
echo   docker compose logs -f
echo.
echo Stop services:
echo   docker compose down
echo.
echo ========================================
pause
