# StreamPass Build Script
# Usage: .\scripts\Build.ps1 [-Target Backend|Client|All]

param(
    [ValidateSet("Backend", "Client", "All")]
    [string]$Target = "All"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

function Build-Backend {
    Write-Host "=== Building Backend ===" -ForegroundColor Cyan
    Push-Location $Root
    go build ./...
    if ($LASTEXITCODE -ne 0) { throw "Backend build failed" }
    Write-Host "Backend: OK" -ForegroundColor Green
    Pop-Location
}

function Build-Client {
    Write-Host "=== Building Client ===" -ForegroundColor Cyan
    Push-Location "$Root\client"
    flutter pub get
    flutter analyze
    if ($LASTEXITCODE -ne 0) { throw "Flutter analyze failed" }
    Write-Host "Client analyze: OK" -ForegroundColor Green
    Pop-Location
}

function Build-Docker {
    Write-Host "=== Building Docker ===" -ForegroundColor Cyan
    Push-Location $Root
    docker compose build
    if ($LASTEXITCODE -ne 0) { throw "Docker build failed" }
    Write-Host "Docker: OK" -ForegroundColor Green
    Pop-Location
}

switch ($Target) {
    "Backend" { Build-Backend }
    "Client"  { Build-Client }
    "All"     { Build-Backend; Build-Client }
}

Write-Host "`nBuild complete." -ForegroundColor Green
