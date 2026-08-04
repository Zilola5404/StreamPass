# Device / API E2E checklist for APK +17/+18 (manual steps when no adb device).
# Usage:
#   .\scripts\VerifyDeviceE2E.ps1
#   .\scripts\VerifyDeviceE2E.ps1 -ApkPath "...\StreamPass-v0.1.1+18-signed-arm64.apk"

param(
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$ApkPath = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if (-not $ApkPath) {
    $ApkPath = Join-Path $Root "client\build\app\outputs\flutter-apk\StreamPass-v0.1.1+18-signed-arm64.apk"
    if (-not (Test-Path $ApkPath)) {
        $ApkPath = Join-Path $Root "client\build\app\outputs\flutter-apk\StreamPass-v0.1.1+17-signed-arm64.apk"
    }
}

Write-Host "=== StreamPass Device E2E ===" -ForegroundColor Cyan
Write-Host "API: $BaseUrl"
Write-Host "APK: $ApkPath"

& "$Root\scripts\SmokeTest.ps1" -BaseUrl $BaseUrl
if (-not $?) { throw "smoke failed" }

$regions = Invoke-RestMethod "$BaseUrl/api/v1/regions"
Write-Host ("Regions: " + (($regions | ForEach-Object { $_.code }) -join ", "))

$adb = Get-Command adb -ErrorAction SilentlyContinue
$deviceLine = ""
if ($adb) {
    $deviceLine = (adb devices | Select-String "`tdevice$")
}
if (-not $deviceLine) {
    Write-Host ""
    Write-Host "[MANUAL] No adb device attached. Install APK and verify:" -ForegroundColor Yellow
    Write-Host "  1. adb install -r `"$ApkPath`""
    Write-Host "  2. Login → Home shows Amsterdam (NL)"
    Write-Host "  3. Open Regions → see DE/PL/FI if registered"
    Write-Host "  4. Connect VPN → browse → check foreign IP"
    Write-Host "  5. Disconnect cleanly"
    Write-Host "API-side automated checks: PASS"
    exit 0
}

Write-Host "Device detected: $deviceLine" -ForegroundColor Green
if (Test-Path $ApkPath) {
    adb install -r $ApkPath
    if ($LASTEXITCODE -ne 0) { throw "adb install failed" }
    Write-Host "APK installed. Complete connect flow on device (VPN permission dialog)." -ForegroundColor Green
} else {
    Write-Host "APK missing: $ApkPath" -ForegroundColor Yellow
}
