# PowerShell build script for Lambda function (Windows)
# This script builds the Lambda function for Linux (required for AWS Lambda)

Write-Host "Building Lambda function..." -ForegroundColor Green

# Ensure we're in the lambda directory (script should be run from lambda directory)
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

# Set environment variables for cross-compilation
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Build the Lambda function (build from current directory)
go build -o bootstrap .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build complete! Binary: bootstrap" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
