# StreamPass Smoke Test
# Usage:
#   .\scripts\SmokeTest.ps1
#   .\scripts\SmokeTest.ps1 -BaseUrl https://212-43-156-33.nip.io
#   .\scripts\SmokeTest.ps1 -BaseUrl https://212-43-156-33.nip.io -AdminKey $env:ADMIN_API_KEY

param(
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$AdminKey = ""
)

$ErrorActionPreference = "Stop"
$ApiBase = "$BaseUrl/api/v1"
$Passed = 0
$Failed = 0

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Url,
        [int]$ExpectedStatus = 200,
        [string]$Method = "GET",
        [hashtable]$Headers = @{}
    )
    $attempts = 0
    while ($true) {
        $attempts++
        try {
            $response = Invoke-WebRequest -Uri $Url -Method $Method -Headers $Headers -UseBasicParsing -ErrorAction Stop
            if ($response.StatusCode -eq $ExpectedStatus) {
                Write-Host "[PASS] $Name ($($response.StatusCode))" -ForegroundColor Green
                $script:Passed++
                return
            }
            Write-Host "[FAIL] $Name (expected $ExpectedStatus, got $($response.StatusCode))" -ForegroundColor Red
            $script:Failed++
            return
        } catch {
            $statusCode = $null
            try { $statusCode = [int]$_.Exception.Response.StatusCode } catch {}
            if ($null -ne $statusCode -and $statusCode -eq $ExpectedStatus) {
                Write-Host "[PASS] $Name ($statusCode)" -ForegroundColor Green
                $script:Passed++
                return
            }
            if ($attempts -lt 3 -and $null -eq $statusCode) {
                Start-Sleep -Seconds 1
                continue
            }
            Write-Host "[FAIL] $Name - $($_.Exception.Message)" -ForegroundColor Red
            $script:Failed++
            return
        }
    }
}

Write-Host "=== StreamPass Smoke Test ===" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl"
Write-Host ""

Test-Endpoint -Name "Health (bare)" -Url "$BaseUrl/health"
Test-Endpoint -Name "Health (v1)" -Url "$ApiBase/health"
Test-Endpoint -Name "Rules (public)" -Url "$ApiBase/rules"
Test-Endpoint -Name "Config (public)" -Url "$ApiBase/config"
Test-Endpoint -Name "Regions (public)" -Url "$ApiBase/regions"
Test-Endpoint -Name "Servers (no auth -> 401)" -Url "$ApiBase/servers" -ExpectedStatus 401
Test-Endpoint -Name "Metrics public blocked" -Url "$BaseUrl/metrics" -ExpectedStatus 404
Test-Endpoint -Name "Admin UI" -Url "$BaseUrl/admin/"

if ($AdminKey) {
    Test-Endpoint -Name "Servers/all (admin)" -Url "$ApiBase/servers/all" -Headers @{ "X-Admin-Key" = $AdminKey }
} else {
    Write-Host "[SKIP] Servers/all (admin) - pass -AdminKey to enable" -ForegroundColor Yellow
}

$color = if ($Failed -eq 0) { "Green" } else { "Red" }
Write-Host ""
Write-Host "=== Results: $Passed passed, $Failed failed ===" -ForegroundColor $color
if ($Failed -gt 0) { exit 1 }
