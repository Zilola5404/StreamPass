# StreamPass Test Runner
# Usage: .\scripts\RunTests.ps1 [-Target Backend|Client|All]

param(
    [ValidateSet("Backend", "Client", "All", "Load")]
    [string]$Target = "All"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

function Test-Backend {
    Write-Host "=== Testing Backend ===" -ForegroundColor Cyan
    Push-Location $Root
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "go vet failed" }
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "go test failed" }
    Write-Host "Backend tests: OK" -ForegroundColor Green
    Pop-Location
}

function Test-Client {
    Write-Host "=== Testing Client ===" -ForegroundColor Cyan
    Push-Location "$Root\client"
    flutter analyze
    if ($LASTEXITCODE -ne 0) { throw "flutter analyze failed" }
    flutter test
    if ($LASTEXITCODE -ne 0) { throw "flutter test failed" }
    Write-Host "Client tests: OK" -ForegroundColor Green
    Pop-Location
}

function Test-Load {
    Write-Host "=== Load Test (BL-032, short) ===" -ForegroundColor Cyan
    Push-Location $Root
    go run ./scripts/loadtest -base https://212-43-156-33.nip.io -duration 5s -rps 15
    if ($LASTEXITCODE -ne 0) { throw "loadtest failed" }
    Write-Host "Load test: OK" -ForegroundColor Green
    Pop-Location
}

switch ($Target) {
    "Backend" { Test-Backend }
    "Client"  { Test-Client }
    "Load"    { Test-Load }
    "All"     { Test-Backend; Test-Client }
}

Write-Host "`nAll tests passed." -ForegroundColor Green
