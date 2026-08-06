# Verify StreamPass traffic behavior: unit matrix + device log patterns + manual checklist.
#
# Usage:
#   .\scripts\VerifyTrafficBehavior.ps1                         # unit tests + checklist
#   .\scripts\VerifyTrafficBehavior.ps1 -AfterConnect           # right after VPN connect
#   .\scripts\VerifyTrafficBehavior.ps1 -AfterTraffic           # after opening youtube.com (foreign DNS via Go)
#   .\scripts\VerifyTrafficBehavior.ps1 -AfterTrafficRu         # after connect (RU DNS via OS / Yandex)
#   .\scripts\VerifyTrafficBehavior.ps1 -AfterDisconnect        # after disconnect
#   .\scripts\VerifyTrafficBehavior.ps1 -DeviceOnly               # adb log patterns only
#   .\scripts\VerifyTrafficBehavior.ps1 -AfterConnect -WithUnit   # device + unit tests
#
param(
    [switch]$DeviceOnly,
    [switch]$AfterConnect,
    [switch]$AfterTraffic,
    [switch]$AfterTrafficRu,
    [switch]$AfterDisconnect,
    [switch]$SkipUnit,
    [switch]$WithUnit
)

$ErrorActionPreference = "Stop"
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding = [System.Text.Encoding]::UTF8
} catch {}

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$ExpectationsFile = Join-Path $Root "scripts\traffic_expectations.json"
$failures = 0
$unitRan = $false
$unitPass = $false
$deviceRan = $false
$devicePass = $false

# -AfterConnect / -AfterDisconnect imply device-only unless -WithUnit
$devicePhase = $AfterConnect -or $AfterTraffic -or $AfterTrafficRu -or $AfterDisconnect -or $DeviceOnly
if ($devicePhase -and -not $WithUnit) {
    $SkipUnit = $true
}

function Step-Pass([string]$msg) { Write-Host "[PASS] $msg" -ForegroundColor Green }
function Step-Fail([string]$msg, [string]$detail = "") {
    $script:failures++
    if ($detail) { Write-Host "[FAIL] $msg - $detail" -ForegroundColor Red }
    else { Write-Host "[FAIL] $msg" -ForegroundColor Red }
}
function Step-Warn([string]$msg) { Write-Host "[WARN] $msg" -ForegroundColor Yellow }
function Step-Info([string]$msg) { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
function Step-Skip([string]$msg) { Write-Host "[SKIP] $msg" -ForegroundColor DarkGray }

Write-Host "=== StreamPass Traffic Behavior Verification ===" -ForegroundColor Cyan
if ($AfterConnect) { Write-Host "Mode: AfterConnect (VPN connect logs)" -ForegroundColor Cyan }
elseif ($AfterTraffic) { Write-Host "Mode: AfterTraffic (open youtube.com in Chrome, then run)" -ForegroundColor Cyan }
elseif ($AfterTrafficRu) { Write-Host "Mode: AfterTrafficRu (RU DNS via OS Yandex - run while connected)" -ForegroundColor Cyan }
elseif ($AfterDisconnect) { Write-Host "Mode: AfterDisconnect (disconnect logs)" -ForegroundColor Cyan }
elseif ($DeviceOnly) { Write-Host "Mode: DeviceOnly" -ForegroundColor Cyan }
else { Write-Host "Mode: Unit + checklist (connect phone for device checks)" -ForegroundColor Cyan }
Write-Host ""

# --- Load expectations JSON (UTF-8) ---
if (-not (Test-Path $ExpectationsFile)) {
    Step-Fail "traffic_expectations.json" "missing at $ExpectationsFile"
    $expectations = $null
} else {
    $jsonText = [System.IO.File]::ReadAllText($ExpectationsFile, [System.Text.Encoding]::UTF8)
    $expectations = $jsonText | ConvertFrom-Json
    Step-Pass "loaded traffic_expectations.json (v$($expectations.version))"
}

# --- Unit: Go traffic matrix ---
if (-not $DeviceOnly -and -not $SkipUnit) {
    $unitRan = $true
    Step-Info "Running Go traffic matrix tests..."
    Push-Location (Join-Path $Root "client\go_core")
    go test ./internal/decision/ -run TrafficMatrix -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Step-Fail "go decision TrafficMatrix" } else { Step-Pass "go decision TrafficMatrix" }

    go test ./internal/dnscache/ -run TrafficDNSMatrix -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Step-Fail "go dnscache TrafficDNSMatrix" } else { Step-Pass "go dnscache TrafficDNSMatrix" }

    go test ./mobile/ -run TrafficMatrix -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Step-Fail "go mobile TrafficMatrix" } else { Step-Pass "go mobile TrafficMatrix public API" }
    Pop-Location

    Step-Info "Running Flutter traffic + lifecycle tests..."
    Push-Location (Join-Path $Root "client")
    flutter test test/traffic_behavior_test.dart test/vpn_lifecycle_test.dart 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Step-Fail "flutter traffic/lifecycle tests" } else { Step-Pass "flutter traffic/lifecycle tests" }
    Pop-Location
    $unitPass = ($failures -eq 0)
} elseif ($SkipUnit) {
    Step-Skip "unit tests (use default run or -WithUnit to include)"
}

# --- Device log validation ---
$adb = Get-Command adb -ErrorAction SilentlyContinue
$hasDevice = $false
$devLine = $null
if ($adb) {
    $devLine = (adb devices 2>&1) | Select-String "`tdevice$"
    $hasDevice = [bool]$devLine
}

function Get-ConnectLogs {
    $logcat = adb logcat -d -s StreamPassConnect StreamPassVpn StreamPassTunnel 2>&1
    $appLog = adb exec-out run-as com.streampass.app cat files/connect.log 2>&1
    $combined = @($logcat)
    if ($LASTEXITCODE -eq 0 -and $appLog -and ($appLog -notmatch "run-as")) {
        $combined += $appLog
    }
    return ($combined -join "`n")
}

function Test-LogPatterns {
    param(
        [array]$Patterns,
        [string]$LogText,
        [string]$Phase,
        [switch]$Strict
    )
    foreach ($p in $Patterns) {
        $regex = [regex]$p.pattern
        $found = $regex.IsMatch($LogText)
        if ($p.forbidden) {
            if ($found) { Step-Fail "$Phase forbidden: $($p.pattern)" $p.note }
            else { Step-Pass "$Phase no forbidden: $($p.pattern)" }
            continue
        }
        if ($p.required -and -not $found) {
            if ($Strict) {
                Step-Fail "$Phase missing: $($p.pattern)" $p.note
            } else {
                Step-Warn "$Phase not in log yet: $($p.pattern) ($($p.note))"
            }
        } elseif ($found) {
            Step-Pass "$Phase found: $($p.pattern)"
        }
    }
}

if ($hasDevice) {
    Step-Pass "adb device: $($devLine.Line.Trim())"

    if ($devicePhase) {
        $deviceRan = $true
        $beforeFails = $failures
        $logs = Get-ConnectLogs
        if ([string]::IsNullOrWhiteSpace($logs)) {
            Step-Fail "device logs" "empty - perform connect/disconnect in StreamPass app first"
        } else {
            Step-Pass "collected device logs ($($logs.Length) chars)"

            if ($AfterConnect -or $DeviceOnly) {
                Test-LogPatterns -Patterns $expectations.connect_log_patterns -LogText $logs -Phase "connect" -Strict:$AfterConnect
            }
            if ($AfterTraffic) {
                Test-LogPatterns -Patterns $expectations.traffic_log_patterns_foreign -LogText $logs -Phase "traffic-foreign" -Strict
            }
            if ($AfterTrafficRu) {
                Test-LogPatterns -Patterns $expectations.traffic_log_patterns_ru -LogText $logs -Phase "traffic-ru" -Strict
            }
            if ($AfterDisconnect -or $DeviceOnly) {
                Test-LogPatterns -Patterns $expectations.disconnect_log_patterns -LogText $logs -Phase "disconnect" -Strict
            }

            $installedBypass = @()
            foreach ($app in $expectations.apps) {
                if (-not $app.bypass) { continue }
                $pkg = $app.package
                $check = adb shell pm path $pkg 2>&1
                if ($check -match "package:") {
                    $installedBypass += $app.label
                    if ($logs -match ("VPN app-bypass:\s*" + [regex]::Escape($pkg)) -or $logs -match [regex]::Escape($pkg)) {
                        Step-Pass "bypass log mentions installed app: $($app.label) ($pkg)"
                    } else {
                        Step-Warn "installed bypass app $($app.label) - no log line for $pkg (check app-bypass applied=N)"
                    }
                }
            }
            if ($installedBypass.Count -eq 0) {
                Step-Warn "no known bypass apps installed - app-bypass applied=0 is expected"
            }
        }
        $devicePass = ($failures -eq $beforeFails)
    } else {
        Step-Info "Phone connected. Recommended flow:"
        Write-Host "  1. Connect VPN in StreamPass"
        Write-Host "  2. .\scripts\VerifyTrafficBehavior.ps1 -AfterConnect"
        Write-Host "  3. .\scripts\VerifyTrafficBehavior.ps1 -AfterTrafficRu   (RU DNS via OS)"
        Write-Host "  4. Open youtube.com in Chrome (while connected)"
        Write-Host "  5. .\scripts\VerifyTrafficBehavior.ps1 -AfterTraffic"
        Write-Host "  6. Disconnect VPN"
        Write-Host "  7. .\scripts\VerifyTrafficBehavior.ps1 -AfterDisconnect"
    }
} else {
    if ($devicePhase) {
        Step-Fail "adb device" "phone not detected - enable USB debugging and run: adb devices"
        Write-Host ""
        Write-Host "Setup:" -ForegroundColor Yellow
        Write-Host "  1. Phone: Settings -> Developer options -> USB debugging ON"
        Write-Host "  2. Connect USB cable, accept RSA prompt on phone"
        Write-Host "  3. adb devices   (must show 'device', not 'unauthorized')"
        Write-Host "  4. Connect VPN in StreamPass app"
        Write-Host "  5. Re-run: .\scripts\VerifyTrafficBehavior.ps1 -AfterConnect"
    } else {
        Step-Warn "No adb device - device log checks skipped (unit tests still ran)"
    }
}

# --- Manual checklist (only in default mode or when device missing) ---
if (-not $devicePhase -or -not $hasDevice) {
    if ($null -ne $expectations) {
        Write-Host ""
        Write-Host "=== Manual traffic checklist (on phone while connected) ===" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Sites:" -ForegroundColor Yellow
        foreach ($site in $expectations.sites) {
            Write-Host ("  [{0}] {1} (dns={2})" -f $site.decision, $site.host, $site.dns)
            Write-Host ("         {0}" -f $site.manual)
        }

        Write-Host ""
        Write-Host "Apps:" -ForegroundColor Yellow
        foreach ($app in $expectations.apps) {
            $tag = if ($app.bypass) { "BYPASS" } else { "IN-VPN" }
            Write-Host ("  [{0}] {1} ({2})" -f $tag, $app.label, $app.package)
            Write-Host ("         {0}" -f $app.manual)
        }

        Write-Host ""
        Write-Host "Lifecycle [manual on device]:" -ForegroundColor Yellow
        foreach ($step in $expectations.lifecycle) {
            if ($step.manual) {
                Write-Host ("  {0}: {1}" -f $step.step, $step.expect)
            }
        }
    }
}

Write-Host ""
Write-Host "=== Summary ===" -ForegroundColor Cyan
if ($unitRan) {
    if ($unitPass -or ($failures -eq 0 -and -not $devicePhase)) {
        Write-Host "  Unit tests (Go + Flutter): PASS" -ForegroundColor Green
    } else {
        Write-Host "  Unit tests (Go + Flutter): FAIL" -ForegroundColor Red
    }
} else {
    Write-Host "  Unit tests: SKIPPED" -ForegroundColor DarkGray
}

if ($deviceRan) {
    if ($devicePass) {
        Write-Host "  Device log checks: PASS" -ForegroundColor Green
    } else {
        Write-Host "  Device log checks: FAIL" -ForegroundColor Red
    }
} elseif ($devicePhase) {
    Write-Host "  Device log checks: NOT RUN (no adb)" -ForegroundColor Red
} else {
    Write-Host "  Device log checks: SKIPPED (no adb or use -AfterConnect)" -ForegroundColor DarkGray
}

Write-Host ""
if ($failures -gt 0) {
    Write-Host "Traffic behavior verification: FAIL ($failures issue(s))" -ForegroundColor Red
    exit 1
}

if ($devicePhase -and $hasDevice -and $deviceRan) {
    Write-Host "Traffic behavior verification: PASS (device logs validated)" -ForegroundColor Green
} elseif (-not $devicePhase) {
    Write-Host "Traffic behavior verification: PASS (unit tests only)" -ForegroundColor Green
    Write-Host "  -> Connect phone via adb and run -AfterConnect for full device validation" -ForegroundColor Yellow
} else {
    Write-Host "Traffic behavior verification: PASS" -ForegroundColor Green
}
exit 0
