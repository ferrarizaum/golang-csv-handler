# Build script for AWS Lambda deployment

$ErrorActionPreference = "Stop"

Write-Host "Building Lambda function..." -ForegroundColor Green

# Set environment variables for Linux build
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Build the binary
Set-Location -Path "$PSScriptRoot\.."
go build -ldflags="-s -w" -o lambda/bootstrap cmd/lambda/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    
    # Create deployment package
    Set-Location -Path "$PSScriptRoot"
    
    if (Test-Path "function.zip") {
        Remove-Item "function.zip" -Force
    }
    
    Compress-Archive -Path "bootstrap" -DestinationPath "function.zip"
    
    Write-Host "Deployment package created: lambda/function.zip" -ForegroundColor Green
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
