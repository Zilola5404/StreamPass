# Collect StreamPass connection logs from a connected Android device via adb.
#
# Usage:
#   .\scripts\CollectConnectLogs.ps1
#   .\scripts\CollectConnectLogs.ps1 -OutFile connect-log.txt
#   .\scripts\CollectConnectLogs.ps1 -Clear

param(
    [string]$OutFile = "",
    [switch]$Clear
)

$ErrorActionPreference = "Stop"
$tags = "StreamPassConnect|StreamPassVpn|StreamPassTunnel|flutter"

if (-not (Get-Command adb -ErrorAction SilentlyContinue)) {
    Write-Host "[FAIL] adb not found in PATH. Install Android platform-tools." -ForegroundColor Red
    exit 1
}

$devices = (& adb devices) | Select-Object -Skip 1 | Where-Object { $_ -match "device$" }
if (-not $devices) {
    Write-Host "[FAIL] No Android device connected." -ForegroundColor Red
    exit 1
}

Write-Host "=== StreamPass connect logs (adb) ===" -ForegroundColor Cyan
Write-Host "Device: $($devices[0])`n"

if ($Clear) {
    & adb logcat -c | Out-Null
    Write-Host "Logcat cleared." -ForegroundColor Yellow
    exit 0
}

$lines = & adb logcat -d -s StreamPassConnect StreamPassVpn StreamPassTunnel flutter 2>&1
$filtered = $lines | Where-Object { $_ -match $tags }

if ($OutFile) {
    $filtered | Set-Content -Path $OutFile -Encoding UTF8
    Write-Host "Saved $($filtered.Count) lines to $OutFile" -ForegroundColor Green
} else {
    $filtered | Select-Object -Last 80 | ForEach-Object { Write-Host $_ }
    Write-Host "`nTip: reproduce Connect on phone, then re-run with -OutFile connect-log.txt" -ForegroundColor Gray
}

# Pull native connect.log from app storage (requires debug/release with run-as or rooted)
Write-Host "`n--- connect.log from app (if accessible) ---" -ForegroundColor Cyan
$appLog = & adb exec-out run-as com.streampass.app cat files/connect.log 2>&1
if ($LASTEXITCODE -eq 0 -and $appLog) {
    if ($OutFile) {
        Add-Content -Path $OutFile -Value "`n--- files/connect.log ---`n$appLog"
    } else {
        $appLog | Select-Object -Last 30 | ForEach-Object { Write-Host $_ }
    }
} else {
    Write-Host "(run-as unavailable on release build — use Diagnostics screen Copy in app)" -ForegroundColor Yellow
}
