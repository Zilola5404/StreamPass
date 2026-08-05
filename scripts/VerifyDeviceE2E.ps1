# Device / API E2E for current StreamPass APK.
# Usage:
#   .\scripts\VerifyDeviceE2E.ps1
#   .\scripts\VerifyDeviceE2E.ps1 -ApkPath "...\StreamPass-v0.1.1+25-signed-arm64.apk"

param(
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$ApkPath = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

function Find-LatestApk {
    $dir = Join-Path $Root "client\build\app\outputs\flutter-apk"
    if (-not (Test-Path $dir)) { return $null }
    $apk = Get-ChildItem $dir -Filter "StreamPass-v*-signed-arm64.apk" |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if ($apk) { return $apk.FullName }
    return $null
}

if (-not $ApkPath) {
    $ApkPath = Find-LatestApk
    if (-not $ApkPath) {
        $ApkPath = Join-Path $Root "client\build\app\outputs\flutter-apk\StreamPass-v0.1.1+25-signed-arm64.apk"
    }
}

Write-Host "=== StreamPass Device E2E ===" -ForegroundColor Cyan
Write-Host "API: $BaseUrl"
Write-Host "APK: $ApkPath"
if (Test-Path $ApkPath) {
    $apkInfo = Get-Item $ApkPath
    Write-Host ("APK size: {0:N1} MB, mtime: {1}" -f ($apkInfo.Length / 1MB), $apkInfo.LastWriteTime)
} else {
    Write-Host "APK missing on disk (download URL check still runs)" -ForegroundColor Yellow
}

$failures = 0
function Step-Pass([string]$msg) { Write-Host "[PASS] $msg" -ForegroundColor Green }
function Step-Fail([string]$msg, [string]$detail = "") {
    $script:failures++
    if ($detail) { Write-Host "[FAIL] $msg - $detail" -ForegroundColor Red }
    else { Write-Host "[FAIL] $msg" -ForegroundColor Red }
}

# --- API smoke ---
& "$Root\scripts\SmokeTest.ps1" -BaseUrl $BaseUrl
if (-not $?) { throw "smoke failed" }

# --- Regions ---
try {
    $regions = Invoke-RestMethod "$BaseUrl/api/v1/regions"
    $codes = @($regions | ForEach-Object { $_.code })
    if ($codes.Count -lt 1) { Step-Fail "regions" "empty list" }
    else { Step-Pass ("regions: " + ($codes -join ", ")) }
} catch {
    Step-Fail "regions" $_.Exception.Message
}

# --- Published config / rules / download ---
try {
    $cfg = Invoke-RestMethod "$BaseUrl/api/v1/config"
    $ver = $cfg.version
    $min = $cfg.min_supported_client_version
    $url = $cfg.client_download_url
    Step-Pass "config version=$ver min=$min"
    if ($min -and ([version]$min -gt [version]"0.1.1")) {
        Step-Fail "min_supported_client_version" "blocks 0.1.1 ($min)"
    }
    if ($url) {
        try {
            $head = Invoke-WebRequest -Uri $url -Method Head -UseBasicParsing
            Step-Pass "download URL HTTP $($head.StatusCode) ($url)"
        } catch {
            Step-Fail "download URL" "$url - $($_.Exception.Message)"
        }
    } else {
        Step-Fail "client_download_url" "empty in published config"
    }
} catch {
    Step-Fail "config" $_.Exception.Message
}

try {
    $rules = Invoke-RestMethod "$BaseUrl/api/v1/rules"
    $n = 0
    if ($rules -is [System.Array]) { $n = $rules.Count }
    elseif ($rules.rules) { $n = @($rules.rules).Count }
    elseif ($rules.Count) { $n = $rules.Count }
    if ($n -lt 1) { Step-Fail "rules" "empty" }
    else { Step-Pass "rules count=$n" }
} catch {
    Step-Fail "rules" $_.Exception.Message
}

# --- TCP underlay ports reachable ---
$tcpPorts = @(8443, 24443)
foreach ($port in $tcpPorts) {
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $iar = $tcp.BeginConnect("212.43.156.33", $port, $null, $null)
        $ok = $iar.AsyncWaitHandle.WaitOne(3000, $false)
        if (-not $ok) {
            $tcp.Close()
            Step-Fail "TCP underlay :$port" "connect timeout (bridge not deployed?)"
            continue
        }
        $tcp.EndConnect($iar)
        $tcp.Close()
        Step-Pass "TCP underlay :$port open"
    } catch {
        Step-Fail "TCP underlay :$port" $_.Exception.Message
    }
}

# --- adb device ---
$adb = Get-Command adb -ErrorAction SilentlyContinue
$deviceLine = $null
if ($adb) {
    $deviceLine = (adb devices | Select-String "`tdevice$")
}

if (-not $deviceLine) {
    Write-Host ""
    Write-Host "[MANUAL] No adb device attached. On phone:" -ForegroundColor Yellow
    Write-Host "  1. adb install -r `"$ApkPath`""
    Write-Host "  2. Login -> Home shows region (NL/DE/PL/FI)"
    Write-Host "  3. Connect VPN -> YouTube/Instagram via relay; .ru DIRECT"
    Write-Host "  4. Settings -> App bypass -> pick bank/gov app"
    Write-Host "  5. Disconnect cleanly (no crash)"
    Write-Host "  6. Optional UDP-block test: logcat shows tcp/8443 candidate"
} else {
    Write-Host "Device detected: $deviceLine" -ForegroundColor Green
    if (Test-Path $ApkPath) {
        adb install -r $ApkPath
        if ($LASTEXITCODE -ne 0) { throw "adb install failed" }
        Step-Pass "APK installed - complete connect flow on device (VPN permission)"
    } else {
        Step-Fail "APK install" "file missing: $ApkPath"
    }
}

Write-Host ""
if ($failures -gt 0) {
    Write-Host "Device E2E finished with $failures failure(s)" -ForegroundColor Red
    exit 1
}
Write-Host "Device E2E API checks: PASS" -ForegroundColor Green
exit 0
