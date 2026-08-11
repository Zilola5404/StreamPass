# MeasureDeviceSLA.ps1 - BL-053 device / API SLA smoke measurement
# Targets (TZ section 22 / FS section 5):
#   cold start (session check) <= 2s
#   connect to Connected       <= 5s
#   recover after network loss <= 10s
#
# Usage:
#   powershell -File scripts/MeasureDeviceSLA.ps1
#   powershell -File scripts/MeasureDeviceSLA.ps1 -ApiBase https://212-43-156-33.nip.io
#   powershell -File scripts/MeasureDeviceSLA.ps1 -SkipAdb
#   powershell -File scripts/MeasureDeviceSLA.ps1 -LogFile connect-export.txt

param(
    [string]$ApiBase = "https://212-43-156-33.nip.io",
    [switch]$SkipAdb,
    [string]$LogFile = ""
)

$ErrorActionPreference = "Continue"
$ApiV1 = if ($ApiBase.EndsWith("/api/v1")) { $ApiBase } else { "$ApiBase/api/v1" }

Write-Host "=== StreamPass SLA measurement (BL-053) ===" -ForegroundColor Cyan
Write-Host "API: $ApiV1"
Write-Host "Targets: cold<=2s connect<=5s recover<=10s"
Write-Host ""

function Measure-Http {
    param([string]$Name, [scriptblock]$Action, [double]$BudgetSec)
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $ok = $false
    $detail = ""
    try {
        $result = & $Action
        $ok = $true
        $detail = "$result"
    } catch {
        $detail = $_.Exception.Message
    }
    $sw.Stop()
    $sec = [math]::Round($sw.Elapsed.TotalSeconds, 3)
    $pass = $ok -and ($sec -le $BudgetSec)
    $mark = if ($pass) { "PASS" } else { "FAIL" }
    $color = if ($pass) { "Green" } else { "Red" }
    Write-Host ("[{0}] {1}: {2}s (budget {3}s) - {4}" -f $mark, $Name, $sec, $BudgetSec, $detail) -ForegroundColor $color
    return [pscustomobject]@{ Name = $Name; Seconds = $sec; Budget = $BudgetSec; Pass = $pass; Detail = $detail }
}

function Parse-ConnectLog {
    param([string]$Path)
    if (-not (Test-Path $Path)) {
        Write-Host "[SKIP] log file not found: $Path" -ForegroundColor Yellow
        return @()
    }
    $text = Get-Content -Raw $Path
    $results = @()

    # Flutter log: event=connecting -> event=connected
    $connecting = [regex]::Match($text, '(\d{4}-\d{2}-\d{2}T[\d:.]+Z).*event=connecting')
    $connected = [regex]::Match($text, '(\d{4}-\d{2}-\d{2}T[\d:.]+Z).*event=connected')
    if ($connecting.Success -and $connected.Success) {
        $t0 = [datetimeoffset]::Parse($connecting.Groups[1].Value)
        $t1 = [datetimeoffset]::Parse($connected.Groups[1].Value)
        $sec = [math]::Round(($t1 - $t0).TotalSeconds, 3)
        $pass = $sec -le 5
        $mark = if ($pass) { "PASS" } else { "FAIL" }
        $color = if ($pass) { "Green" } else { "Red" }
        Write-Host ("[{0}] device connect (log): {1}s (budget 5s)" -f $mark, $sec) -ForegroundColor $color
        $results += [pscustomobject]@{ Name = "device connect (log)"; Seconds = $sec; Budget = 5; Pass = $pass; Detail = "connecting->connected" }
    }

    # Native: connect session begin -> tunnel event=connected
    $begin = [regex]::Match($text, '(\d{4}-\d{2}-\d{2}T[\d:.]+Z).*connect session begin')
    $tunnel = [regex]::Match($text, '(\d{4}-\d{2}-\d{2}T[\d:.]+Z).*tunnel event=connected')
    if ($begin.Success -and $tunnel.Success) {
        $t0 = [datetimeoffset]::Parse($begin.Groups[1].Value)
        $t1 = [datetimeoffset]::Parse($tunnel.Groups[1].Value)
        $sec = [math]::Round(($t1 - $t0).TotalSeconds, 3)
        $pass = $sec -le 5
        $mark = if ($pass) { "PASS" } else { "FAIL" }
        $color = if ($pass) { "Green" } else { "Red" }
        Write-Host ("[{0}] device connect (native): {1}s (budget 5s)" -f $mark, $sec) -ForegroundColor $color
        $results += [pscustomobject]@{ Name = "device connect (native)"; Seconds = $sec; Budget = 5; Pass = $pass; Detail = "session begin->tunnel connected" }
    }

    # Hysteria pingMs
    $ping = [regex]::Match($text, 'pingMs=(\d+)')
    if ($ping.Success) {
        $ms = [int]$ping.Groups[1].Value
        $sec = [math]::Round($ms / 1000.0, 3)
        $pass = $sec -le 5
        $mark = if ($pass) { "PASS" } else { "FAIL" }
        $color = if ($pass) { "Green" } else { "Red" }
        Write-Host ("[{0}] hysteria handshake: {1}s / {2}ms (budget 5s)" -f $mark, $sec, $ms) -ForegroundColor $color
        $results += [pscustomobject]@{ Name = "hysteria handshake"; Seconds = $sec; Budget = 5; Pass = $pass; Detail = "${ms}ms" }
    }

    return $results
}

$results = @()

$results += Measure-Http -Name "API health (cold proxy)" -BudgetSec 2 -Action {
    $r = Invoke-WebRequest -Uri "$ApiV1/health" -UseBasicParsing -TimeoutSec 5
    if ($r.StatusCode -ne 200) { throw "status $($r.StatusCode)" }
    return "HTTP $($r.StatusCode)"
}

$results += Measure-Http -Name "API config (startup path)" -BudgetSec 2 -Action {
    $r = Invoke-WebRequest -Uri "$ApiV1/config" -UseBasicParsing -TimeoutSec 5
    if ($r.StatusCode -ne 200) { throw "status $($r.StatusCode)" }
    return "HTTP $($r.StatusCode)"
}

if ($LogFile) {
    Write-Host ""
    Write-Host "Parsing device log: $LogFile" -ForegroundColor Cyan
    $results += Parse-ConnectLog -Path $LogFile
}

if (-not $SkipAdb) {
    $adb = Get-Command adb -ErrorAction SilentlyContinue
    if ($null -eq $adb) {
        Write-Host "[SKIP] adb not in PATH - device connect/recover not measured" -ForegroundColor Yellow
    } else {
        $devices = & adb devices 2>$null | Select-String "device$"
        if (-not $devices) {
            Write-Host "[SKIP] no adb device attached" -ForegroundColor Yellow
        } else {
            Write-Host ""
            Write-Host "Device present. Manual steps for recover:" -ForegroundColor Cyan
            Write-Host "  1. Toggle airplane mode 3s, note recover to Connected (budget 10s)."
            Write-Host "  2. adb logcat -d | Select-String 'failover|recover|connected'"
            Write-Host ""
            $results += Measure-Http -Name "adb am start (activity launch)" -BudgetSec 2 -Action {
                & adb shell am force-stop com.streampass.app 2>$null | Out-Null
                Start-Sleep -Milliseconds 300
                $out = & adb shell am start -W -n com.streampass.app/.MainActivity 2>&1 | Out-String
                if ($out -match "TotalTime:\s*(\d+)") {
                    $ms = [int]$Matches[1]
                    if ($ms -gt 2000) { throw "TotalTime ${ms}ms > 2000ms" }
                    return "TotalTime ${ms}ms"
                }
                return ($out.Trim() -replace "`r|`n", " ").Substring(0, [Math]::Min(120, $out.Length))
            }
        }
    }
}

Write-Host ""
$passed = @($results | Where-Object { $_.Pass }).Count
$total = $results.Count
Write-Host "Summary: $passed / $total checks passed" -ForegroundColor $(if ($passed -eq $total) { "Green" } else { "Yellow" })
if (-not $LogFile) {
    Write-Host "Tip: export Diagnostics log to file and pass -LogFile for connect timing."
}
Write-Host "Recover <=10s requires manual device pass (airplane toggle)."
exit $(if ($passed -eq $total) { 0 } else { 1 })
