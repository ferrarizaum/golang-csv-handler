@echo off
REM Build script for Lambda function (Windows)
REM This script builds the Lambda function for Linux (required for AWS Lambda)

echo Building Lambda function...

REM Change to lambda directory
cd lambda

REM Check if Makefile exists and use it, otherwise build manually
if exist Makefile (
    echo Using Makefile...
    make build
) else (
    echo Building manually...
    cd ..
    set GOOS=linux
    set GOARCH=amd64
    set CGO_ENABLED=0
    go build -o lambda\bootstrap ./lambda
)

echo Build complete! Binary: lambda\bootstrap

