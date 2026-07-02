@echo off
echo ====================================
echo OWASP CRS Setup Script for Windows
echo ====================================
echo.

set CRS_DIR=configs\crs

echo Checking if CRS directory exists...
if exist %CRS_DIR% (
    echo CRS directory already exists: %CRS_DIR%
    echo.
    echo To update CRS, delete the directory first:
    echo   rmdir /s /q %CRS_DIR%
    echo Then run this script again.
    exit /b 0
)

echo Creating CRS directory...
mkdir %CRS_DIR%
cd %CRS_DIR%

echo.
echo Downloading OWASP CRS from GitHub...
echo.

REM Check if git is available
where git >nul 2>nul
if %errorlevel% neq 0 (
    echo ERROR: git is not installed or not in PATH
    echo.
    echo Please install git from: https://git-scm.com/download/win
    echo Or manually download CRS from: https://github.com/coreruleset/coreruleset/releases
    echo Extract the contents to: %CD%
    exit /b 1
)

REM Clone the CRS repository
echo Cloning OWASP CRS repository...
git clone --depth 1 https://github.com/coreruleset/coreruleset.git .
if %errorlevel% neq 0 (
    echo ERROR: Failed to clone CRS repository
    exit /b 1
)

echo.
echo Setting up CRS configuration...
if exist crs-setup.conf.example (
    copy crs-setup.conf.example crs-setup.conf
    echo Created crs-setup.conf from example
) else (
    echo WARNING: crs-setup.conf.example not found
)

echo.
echo ====================================
echo OWASP CRS Setup Complete!
echo ====================================
echo.
echo CRS installed to: %CD%
echo.
echo Next steps:
echo 1. Review and edit configs\crs\crs-setup.conf
echo 2. Configure your WAF to use CRS by setting:
echo    WAF_CRS_ENABLED=true
echo    WAF_CRS_PATH=./configs/crs
echo.
echo 3. Restart the WAF proxy
echo.

cd ..\..
