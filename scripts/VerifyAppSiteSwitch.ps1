# Interactive QA: switch between sites/apps and report problems per target.
#
# Prerequisites: phone via adb, VPN connected in StreamPass.
#
# Usage:
#   .\scripts\VerifyAppSiteSwitch.ps1                    # all scenarios, interactive
#   .\scripts\VerifyAppSiteSwitch.ps1 -AutoLaunch        # adb opens URL/app, 12s wait each
#   .\scripts\VerifyAppSiteSwitch.ps1 -Scenario site_youtube
#   .\scripts\VerifyAppSiteSwitch.ps1 -SkipManual          # log analysis only, no y/n prompts
#   .\scripts\VerifyAppSiteSwitch.ps1 -ReportPath reports\QA\traffic-switch-latest.md
#
param(
    [string[]]$Scenario = @(),
    [switch]$AutoLaunch,
    [switch]$SkipManual,
    [switch]$WithUnit,
    [int]$WaitSeconds = 12,
    [string]$ReportPath = ""
)

$ErrorActionPreference = "Stop"
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding = [System.Text.Encoding]::UTF8
} catch {}

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$ExpectationsFile = Join-Path $Root "scripts\traffic_expectations.json"
$results = @()
$failCount = 0
$warnCount = 0

function Write-Step([string]$level, [string]$msg, [string]$detail = "") {
    switch ($level) {
        "PASS" { Write-Host "[PASS] $msg" -ForegroundColor Green }
        "FAIL" { Write-Host "[FAIL] $msg" -ForegroundColor Red; if ($detail) { Write-Host "       $detail" -ForegroundColor DarkRed } }
        "WARN" { Write-Host "[WARN] $msg" -ForegroundColor Yellow; if ($detail) { Write-Host "       $detail" -ForegroundColor DarkYellow } }
        "INFO" { Write-Host "[INFO] $msg" -ForegroundColor Cyan }
        "SKIP" { Write-Host "[SKIP] $msg" -ForegroundColor DarkGray }
    }
}

function Invoke-AdbQuiet {
    param([string[]]$AdbArgs)
    if (-not $AdbArgs -or $AdbArgs.Count -eq 0) { return }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        & adb @AdbArgs 2>&1 | Out-Null
    } finally {
        $ErrorActionPreference = $prevEap
    }
}

function Test-PackageInstalled([string]$pkg) {
    if (-not $pkg) { return $false }
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $out = adb shell pm path $pkg 2>&1
    } finally {
        $ErrorActionPreference = $prevEap
    }
    return ($out -match "package:")
}

function Get-FreshLogs {
    param([string]$SinceMarker)
    $prevEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $logcat = adb logcat -d -s StreamPassConnect StreamPassVpn StreamPassTunnel 2>&1
        $appLog = adb exec-out run-as com.streampass.app cat files/connect.log 2>&1
    } finally {
        $ErrorActionPreference = $prevEap
    }
    $combined = @($logcat)
    if ($LASTEXITCODE -eq 0 -and $appLog -and ($appLog -notmatch "run-as")) {
        $combined += $appLog
    }
    $text = $combined -join "`n"
    if ($SinceMarker -and $text -match [regex]::Escape($SinceMarker)) {
        $idx = $text.LastIndexOf($SinceMarker)
        if ($idx -ge 0) { return $text.Substring($idx) }
    }
    return $text
}

function Invoke-ScenarioLaunch {
    param($sc)
    if ($sc.type -eq "site" -and $sc.url) {
        Invoke-AdbQuiet @('shell', 'am', 'start', '-a', 'android.intent.action.VIEW', '-d', $sc.url)
        return "launched url=$($sc.url)"
    }
    if ($sc.type -eq "app" -and $sc.package) {
        $pkg = $sc.package
        $prevEap = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            $component = adb shell cmd package resolve-activity --brief -c android.intent.category.LAUNCHER $pkg 2>&1 |
                Select-Object -Last 1
        } finally {
            $ErrorActionPreference = $prevEap
        }
        if ($component -and ($component -match "/")) {
            Invoke-AdbQuiet @('shell', 'am', 'start', '-n', $component.Trim())
            return "launched pkg=$pkg via am start"
        }
        Invoke-AdbQuiet @('shell', 'monkey', '-p', $pkg, '-c', 'android.intent.category.LAUNCHER', '1')
        return "launched pkg=$pkg via monkey"
    }
    return "manual (no auto launch)"
}

function Analyze-Scenario {
    param(
        $sc,
        [string]$ConnectLogs,
        [string]$StepLogs,
        [hashtable]$ManualAnswers
    )
    $issues = @()
    $notes = @()
    $status = "PASS"

    foreach ($sign in @($sc.success_signs)) {
        if (-not $sign.pattern) { continue }
        $rx = [regex]$sign.pattern
        $checkStep = -not $sign.connect_only
        $inStep = $checkStep -and $rx.IsMatch($StepLogs)
        $inConnect = $rx.IsMatch($ConnectLogs)
        if ($inStep -or $inConnect) {
            $notes += "OK: $($sign.note)"
        } elseif ($sign.note -and -not $sign.optional) {
            $issues += "WARN: expected sign '$($sign.note)' not found in logs"
            if ($status -eq "PASS") { $status = "WARN" }
        } elseif ($sign.note) {
            $notes += "SKIP optional: $($sign.note)"
        }
    }

    foreach ($fail in @($sc.failure_signs)) {
        if (-not $fail.pattern) { continue }
        $rx = [regex]$fail.pattern
        if ($rx.IsMatch($StepLogs) -or $rx.IsMatch($ConnectLogs)) {
            $issues += "FAIL: $($fail.issue)"
            $status = "FAIL"
        }
    }

    foreach ($check in @($sc.manual_checks)) {
        if ($ManualAnswers.ContainsKey($check)) {
            if (-not $ManualAnswers[$check]) {
                $issues += "FAIL (manual): $check"
                $status = "FAIL"
            } else {
                $notes += "OK (manual): $check"
            }
        }
    }

    return @{
        Status = $status
        Issues = $issues
        Notes  = $notes
    }
}

Write-Host "=== StreamPass App/Site Switch QA ===" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $ExpectationsFile)) {
    Write-Step "FAIL" "traffic_expectations.json missing" $ExpectationsFile
    exit 1
}

$jsonText = [System.IO.File]::ReadAllText($ExpectationsFile, [System.Text.Encoding]::UTF8)
$expectations = $jsonText | ConvertFrom-Json
Write-Step "PASS" "loaded traffic_expectations.json (v$($expectations.version))"

if ($WithUnit) {
    Write-Step "INFO" "Running Go/Flutter traffic unit tests..."
    Push-Location (Join-Path $Root "client\go_core")
    go test ./internal/decision/ -run TrafficMatrix -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Step "FAIL" "go decision TrafficMatrix" }
    else { Write-Step "PASS" "go decision TrafficMatrix" }
    go test ./internal/dnscache/ -run TrafficDNSMatrix -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Step "FAIL" "go dnscache TrafficDNSMatrix" }
    else { Write-Step "PASS" "go dnscache TrafficDNSMatrix" }
    Pop-Location
    Push-Location (Join-Path $Root "client")
    flutter test test/traffic_behavior_test.dart 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) { Write-Step "FAIL" "flutter traffic_behavior_test" }
    else { Write-Step "PASS" "flutter traffic_behavior_test" }
    Pop-Location
}

$adb = Get-Command adb -ErrorAction SilentlyContinue
if (-not $adb) {
    Write-Step "FAIL" "adb not found in PATH"
    exit 1
}

$devLine = (adb devices 2>&1) | Select-String "`tdevice$"
$unauthLine = (adb devices 2>&1) | Select-String "`tunauthorized$"
if ($unauthLine) {
    Write-Step "FAIL" "adb device unauthorized - unlock phone and tap Allow USB debugging"
    Write-Host ""
    Write-Host "  Fix:" -ForegroundColor Yellow
    Write-Host "  1. Unplug and replug USB cable"
    Write-Host "  2. On phone: allow 'USB debugging' RSA prompt (Always allow from this computer)"
    Write-Host "  3. Run: adb devices   (must show 'device', not 'unauthorized')"
    exit 1
}
if (-not $devLine) {
    Write-Step "FAIL" "no adb device - enable USB debugging"
    Write-Host ""
    Write-Host "  Fix:" -ForegroundColor Yellow
    Write-Host "  1. Phone: Settings -> Developer options -> USB debugging ON"
    Write-Host "  2. Connect USB cable, accept RSA prompt on phone"
    Write-Host "  3. Run: adb devices   (must show serial + 'device')"
    exit 1
}
Write-Step "PASS" "adb device: $($devLine.Line.Trim())"

$connectLogs = Get-FreshLogs -SinceMarker ""
if ($connectLogs -notmatch "tunnel event=connected|connected") {
    Write-Step "WARN" "VPN may not be connected - connect in StreamPass first"
} else {
    Write-Step "PASS" "VPN connected (found in logs)"
}

$scenarios = @($expectations.switch_scenarios)
if ($Scenario.Count -gt 0) {
    $filter = @{}
    foreach ($id in $Scenario) { $filter[$id] = $true }
    $scenarios = $scenarios | Where-Object { $filter.ContainsKey($_.id) }
    if ($scenarios.Count -eq 0) {
        Write-Step "FAIL" "no scenarios matched: $($Scenario -join ', ')"
        exit 1
    }
}

Write-Host ""
Write-Step "INFO" "Scenarios to run: $($scenarios.Count)"
Write-Host ""

$scenarioIndex = 0
foreach ($sc in $scenarios) {
    $scenarioIndex++
    Write-Host "--- [$scenarioIndex/$($scenarios.Count)] $($sc.id): $($sc.label) ---" -ForegroundColor Cyan
    Write-Host "  type: $($sc.type) | expected: $($sc.decision)" -ForegroundColor DarkGray
    Write-Host "  $($sc.action)" -ForegroundColor White
    Write-Host ""

    if ($sc.skip_if_not_installed -and $sc.package -and -not (Test-PackageInstalled $sc.package)) {
        Write-Step "SKIP" "$($sc.label) - package $($sc.package) not installed"
        $results += [pscustomobject]@{
            Id       = $sc.id
            Label    = $sc.label
            Type     = $sc.type
            Decision = $sc.decision
            Status   = "SKIP"
            Problems = "not installed"
        }
        Write-Host ""
        continue
    }

    $marker = "SWITCHQA-$($sc.id)-$(Get-Date -Format 'HHmmss')"
    Invoke-AdbQuiet @('logcat', '-c')
    Invoke-AdbQuiet @('shell', 'log', '-t', $marker, 'marker')

    $launchInfo = ""
    if ($AutoLaunch -and ($sc.type -eq "site" -or ($sc.type -eq "app" -and $sc.package))) {
        $launchInfo = Invoke-ScenarioLaunch -sc $sc
        Write-Step "INFO" $launchInfo
        Write-Step "INFO" "Waiting ${WaitSeconds}s for traffic..."
        Start-Sleep -Seconds $WaitSeconds
    } else {
        Write-Host "  >> Do the action on the phone, then press Enter <<" -ForegroundColor Yellow
        if ($sc.type -eq "lifecycle") {
            Read-Host "  (Enter when task switch is done)"
        } else {
            Read-Host "  (Enter when site/app is open and loaded)"
        }
    }

    $stepLogs = Get-FreshLogs -SinceMarker $marker
    if ([string]::IsNullOrWhiteSpace($stepLogs)) {
        $stepLogs = Get-FreshLogs -SinceMarker ""
    }

    $manualAnswers = @{}
    if (-not $SkipManual -and $sc.manual_checks) {
        Write-Host ""
        Write-Host "  Manual checks:" -ForegroundColor Yellow
        foreach ($check in @($sc.manual_checks)) {
            $ans = Read-Host "    OK? [$check] (y/n/s=skip)"
            if ($ans -eq "s" -or $ans -eq "S") {
                $manualAnswers[$check] = $null
            } elseif ($ans -eq "y" -or $ans -eq "Y" -or $ans -eq "") {
                $manualAnswers[$check] = $true
            } else {
                $manualAnswers[$check] = $false
            }
        }
    }

    $analysis = Analyze-Scenario -sc $sc -ConnectLogs $connectLogs -StepLogs $stepLogs -ManualAnswers $manualAnswers

    foreach ($n in $analysis.Notes) { Write-Step "PASS" $n }
    foreach ($i in $analysis.Issues) {
        if ($i -like "WARN:*") { Write-Step "WARN" ($i -replace "^WARN:\s*", ""); $warnCount++ }
        else { Write-Step "FAIL" ($i -replace "^FAIL.*?:\s*", ""); $failCount++ }
    }

    if ($analysis.Issues.Count -eq 0) {
        Write-Step "PASS" "$($sc.id) - no problems detected"
    }

    $problemText = ($analysis.Issues -join "; ")
    if (-not $problemText) { $problemText = "-" }

    $results += [pscustomobject]@{
        Id       = $sc.id
        Label    = $sc.label
        Type     = $sc.type
        Decision = $sc.decision
        Status   = $analysis.Status
        Problems = $problemText
    }
    Write-Host ""
}

Write-Host "=== Results table ===" -ForegroundColor Cyan
$results | Format-Table -AutoSize Id, Label, Type, Decision, Status, Problems

Write-Host "=== Summary ===" -ForegroundColor Cyan
$passed = ($results | Where-Object { $_.Status -eq "PASS" }).Count
$failed = ($results | Where-Object { $_.Status -eq "FAIL" }).Count
$warned = ($results | Where-Object { $_.Status -eq "WARN" }).Count
$skipped = ($results | Where-Object { $_.Status -eq "SKIP" }).Count
Write-Host "  PASS: $passed | WARN: $warned | FAIL: $failed | SKIP: $skipped" -ForegroundColor $(if ($failed -gt 0) { "Red" } elseif ($warned -gt 0) { "Yellow" } else { "Green" })

if ($ReportPath) {
    $dir = Split-Path -Parent (Join-Path $Root $ReportPath)
    if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
    $fullPath = Join-Path $Root $ReportPath
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm"
    $md = @(
        "# Traffic switch QA report",
        "",
        "Generated: $ts",
        "",
        "| ID | Label | Type | Expected | Status | Problems |",
        "|----|-------|------|----------|--------|----------|"
    )
    foreach ($r in $results) {
        $prob = $r.Problems -replace [char]0x007C, '/'
        $prob = $prob -replace "`n", " "
        $md += "| $($r.Id) | $($r.Label) | $($r.Type) | $($r.Decision) | $($r.Status) | $prob |"
    }
    $md += ""
    $md += "## Legend"
    $md += "- **site** - Chrome URL, check DIRECT/RELAY + DNS in logs"
    $md += "- **app** - native app, check VPN bypass + no VPN block dialog"
    $md += "- **lifecycle** - task switch stability (StreamPass must stay alive)"
    [System.IO.File]::WriteAllText($fullPath, ($md -join "`n"), [System.Text.Encoding]::UTF8)
    Write-Step "PASS" "Report saved: $fullPath"
}

if ($failed -gt 0 -or $failCount -gt 0) { exit 1 }
exit 0
