# BL-003 E2E VPN verification (Android + relay)

param(
    [string]$RelayURI = "",
    [switch]$SkipGoIntegration,
    [switch]$SkipApkBuild
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$GoCore = Join-Path $Root "client\go_core"
$Passed = 0
$Failed = 0
$Skipped = 0

function Step-Pass($Name) { Write-Host "[PASS] $Name" -ForegroundColor Green; $script:Passed++ }
function Step-Fail($Name, $Detail) { Write-Host "[FAIL] $Name - $Detail" -ForegroundColor Red; $script:Failed++ }
function Step-Skip($Name, $Detail) { Write-Host "[SKIP] $Name - $Detail" -ForegroundColor Yellow; $script:Skipped++ }

if (-not $RelayURI) {
    if (-not $env:STREAMPASS_RELAY_URI) {
        Step-Skip "Go integration" "Set -RelayURI or `$env:STREAMPASS_RELAY_URI"
        $SkipGoIntegration = $true
    } else {
        $RelayURI = $env:STREAMPASS_RELAY_URI
    }
} else {
    $env:STREAMPASS_RELAY_URI = $RelayURI
}

Write-Host "=== BL-003 Verification ===" -ForegroundColor Cyan

# 1. AAR present + Go core class
$AarPath = Join-Path $Root "client\android\app\libs\streampasscore.aar"
if (Test-Path $AarPath) {
    $sizeMB = [math]::Round((Get-Item $AarPath).Length / 1MB, 1)
    Step-Pass "streampasscore.aar exists ($sizeMB MB)"
    $checkDir = Join-Path $env:TEMP "streampass-bl003-aar"
    New-Item -ItemType Directory -Force -Path $checkDir | Out-Null
    Copy-Item $AarPath (Join-Path $checkDir "core.aar") -Force
    Push-Location $checkDir
    jar xf core.aar classes.jar 2>$null
    $classes = jar tf classes.jar 2>$null
    Pop-Location
    if ($classes -match "mobile/Mobile\.class") {
        Step-Pass "AAR contains mobile.Mobile (gomobile entry)"
    } else {
        Step-Fail "AAR class check" "mobile.Mobile not found in classes.jar"
    }
} else {
    Step-Fail "streampasscore.aar" "missing at $AarPath"
}

# 2. Go unit tests
Push-Location $GoCore
try {
    go test -short ./... 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) { Step-Pass "go test -short ./..." } else { Step-Fail "go test -short" "exit $LASTEXITCODE" }
} finally { Pop-Location }

# 3. Go integration (live relay - start local: docker compose -f docker-compose.hysteria-test.yml up -d)
if (-not $SkipGoIntegration) {
    Push-Location $GoCore
    try {
        Write-Host "Running hysteria integration tests (STREAMPASS_RELAY_URI)..." -ForegroundColor Gray
        go test -v -timeout 3m -run "TestIntegrationHysteria" ./internal/hyconfig/ 2>&1
        if ($LASTEXITCODE -eq 0) { Step-Pass "hysteria integration (connect + foreign IP)" } else { Step-Fail "hysteria integration" "exit $LASTEXITCODE" }
    } finally { Pop-Location }
} else {
    Step-Skip "hysteria integration" "-SkipGoIntegration"
}

# 4. Android device / emulator
$adb = Get-Command adb -ErrorAction SilentlyContinue
if ($adb) {
    $devices = (& adb devices) | Select-Object -Skip 1 | Where-Object { $_ -match "device$" }
    if ($devices) {
        Step-Pass "Android device connected ($($devices.Count))"
    } else {
        Step-Skip "Android device E2E" "no device/emulator - manual Connect + ifconfig.me required"
    }
} else {
    Step-Skip "adb" "not in PATH"
}

# 5. APK build
if (-not $SkipApkBuild) {
    Push-Location (Join-Path $Root "client")
    try {
        flutter build apk --debug 2>&1 | Out-Null
        if ($LASTEXITCODE -eq 0) {
            Step-Pass "flutter build apk --debug"
        } else {
            Step-Fail "flutter build apk" "exit $LASTEXITCODE"
        }
    } finally { Pop-Location }
} else {
    Step-Skip "APK build" "-SkipApkBuild"
}

Write-Host "`n=== BL-003 Results: $Passed passed, $Failed failed, $Skipped skipped ===" -ForegroundColor $(if ($Failed -eq 0) { "Green" } else { "Red" })
if ($Failed -gt 0) { exit 1 }
