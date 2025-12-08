@echo off
REM Build script for Lambda function (Windows CMD)
REM This script builds the Lambda function for Linux (required for AWS Lambda)

echo Building Lambda function...

REM Ensure we're in the lambda directory (script should be run from lambda directory)
cd /d "%~dp0"

REM Set environment variables for cross-compilation
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

REM Build the Lambda function (build from current directory)
go build -o bootstrap .

if %ERRORLEVEL% EQU 0 (
    echo Build complete! Binary: bootstrap
) else (
    echo Build failed!
    exit /b 1
)
