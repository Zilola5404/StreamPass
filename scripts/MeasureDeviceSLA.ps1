# MeasureDeviceSLA.ps1 — BL-053 device / API SLA smoke measurement
# Targets (ТЗ §22 / FS §5):
#   cold start (session check) ≤ 2s
#   connect to Connected       ≤ 5s
#   recover after network loss ≤ 10s
#
# Usage:
#   pwsh scripts/MeasureDeviceSLA.ps1
#   pwsh scripts/MeasureDeviceSLA.ps1 -ApiBase https://212-43-156-33.nip.io
#   pwsh scripts/MeasureDeviceSLA.ps1 -SkipAdb

param(
    [string]$ApiBase = "https://212-43-156-33.nip.io",
    [switch]$SkipAdb
)

$ErrorActionPreference = "Continue"
$ApiV1 = if ($ApiBase.EndsWith("/api/v1")) { $ApiBase } else { "$ApiBase/api/v1" }

Write-Host "=== StreamPass SLA measurement (BL-053) ===" -ForegroundColor Cyan
Write-Host "API: $ApiV1"
Write-Host "Targets: cold≤2s connect≤5s recover≤10s"
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
    Write-Host ("[{0}] {1}: {2}s (budget {3}s) — {4}" -f $mark, $Name, $sec, $BudgetSec, $detail) -ForegroundColor $color
    return [pscustomobject]@{ Name = $Name; Seconds = $sec; Budget = $BudgetSec; Pass = $pass; Detail = $detail }
}

$results = @()

# Proxy for cold-start path: GET /health should be well under 2s (session check
# on device also hits /refresh when logged in — measured separately via adb).
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

if (-not $SkipAdb) {
    $adb = Get-Command adb -ErrorAction SilentlyContinue
    if ($null -eq $adb) {
        Write-Host "[SKIP] adb not in PATH — device connect/recover not measured" -ForegroundColor Yellow
    } else {
        $devices = & adb devices 2>$null | Select-String "device$"
        if (-not $devices) {
            Write-Host "[SKIP] no adb device attached" -ForegroundColor Yellow
        } else {
            Write-Host ""
            Write-Host "Device present. Manual steps for connect/recover:" -ForegroundColor Cyan
            Write-Host "  1. Open StreamPass, tap Connect, note time to Connected (budget 5s)."
            Write-Host "  2. Toggle airplane mode 3s, note recover to Connected (budget 10s)."
            Write-Host "  3. Optionally parse logcat:"
            Write-Host "       adb logcat -d | Select-String 'connect|Connected|failover'"
            Write-Host ""
            # Best-effort: time launching main activity cold start.
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
Write-Host "Summary: $passed / $total automated checks passed" -ForegroundColor $(if ($passed -eq $total) { "Green" } else { "Yellow" })
Write-Host "Connect ≤5s and recover ≤10s require a manual device pass (or instrumented UI test)."
exit $(if ($passed -eq $total) { 0 } else { 1 })
