# StreamPass Smoke Test
# Usage: .\scripts\SmokeTest.ps1 [-BaseUrl http://localhost:8080]

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$ApiBase = "$BaseUrl/api/v1"
$Passed = 0
$Failed = 0

function Test-Endpoint {
    param([string]$Name, [string]$Url, [int]$ExpectedStatus = 200, [string]$Method = "GET")
    try {
        $response = Invoke-WebRequest -Uri $Url -Method $Method -UseBasicParsing -ErrorAction Stop
        if ($response.StatusCode -eq $ExpectedStatus) {
            Write-Host "[PASS] $Name ($($response.StatusCode))" -ForegroundColor Green
            $script:Passed++
        } else {
            Write-Host "[FAIL] $Name (expected $ExpectedStatus, got $($response.StatusCode))" -ForegroundColor Red
            $script:Failed++
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq $ExpectedStatus) {
            Write-Host "[PASS] $Name ($statusCode)" -ForegroundColor Green
            $script:Passed++
        } else {
            Write-Host "[FAIL] $Name - $($_.Exception.Message)" -ForegroundColor Red
            $script:Failed++
        }
    }
}

Write-Host "=== StreamPass Smoke Test ===" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl`n"

Test-Endpoint "Health (bare)" "$BaseUrl/health"
Test-Endpoint "Health (v1)" "$ApiBase/health"
Test-Endpoint "Rules (public)" "$ApiBase/rules"
Test-Endpoint "Config (public)" "$ApiBase/config"
Test-Endpoint "Servers (no auth → 401)" "$ApiBase/servers" -ExpectedStatus 401

Write-Host "`n=== Results: $Passed passed, $Failed failed ===" -ForegroundColor $(if ($Failed -eq 0) { "Green" } else { "Red" })
if ($Failed -gt 0) { exit 1 }
