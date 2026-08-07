# Diagnose why StreamPass VPN has no / broken traffic (layer-by-layer blockers).
#
# Usage:
#   .\scripts\DiagnoseTrafficBlock.ps1
#   .\scripts\DiagnoseTrafficBlock.ps1 -LiveProbe
#   .\scripts\DiagnoseTrafficBlock.ps1 -ReportPath reports\QA\traffic-block-diagnosis.md
#   .\scripts\DiagnoseTrafficBlock.ps1 -WithUnit
#   .\scripts\DiagnoseTrafficBlock.ps1 -AllowNoVpn   # skip VPN gate (infra checks only)
#
param(
    [switch]$LiveProbe,
    [switch]$WithUnit,
    [switch]$AllowNoVpn,
    [int]$ProbeWaitSeconds = 15,
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$ReportPath = ""
)

$ErrorActionPreference = "Stop"
try {
    [Console]::OutputEncoding = [System.Text.Encoding]::UTF8
    $OutputEncoding = [System.Text.Encoding]::UTF8
} catch {}

$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
. (Join-Path $Root "scripts\StreamPassDeviceLogs.ps1")
$findings = @()
$blockers = @()
$warns = @()

function Add-Finding {
    param(
        [string]$Layer,
        [string]$Check,
        [string]$Status,
        [string]$Detail = "",
        [switch]$IsBlocker
    )
    $script:findings += [pscustomobject]@{
        Layer  = $Layer
        Check  = $Check
        Status = $Status
        Detail = $Detail
    }
    if ($IsBlocker -and $Status -eq "FAIL") {
        $script:blockers += "$Layer / $Check : $Detail"
    } elseif ($Status -eq "WARN") {
        $script:warns += "$Layer / $Check : $Detail"
    }
}

function Write-FindingLine($f) {
    $color = switch ($f.Status) {
        "PASS" { "Green" }
        "FAIL" { "Red" }
        "WARN" { "Yellow" }
        default { "DarkGray" }
    }
    $suffix = if ($f.Detail) { " - $($f.Detail)" } else { "" }
    Write-Host ("[{0}] {1}/{2}{3}" -f $f.Status, $f.Layer, $f.Check, $suffix) -ForegroundColor $color
}

function Invoke-AdbQuiet {
    param([string[]]$AdbArgs)
    if (-not $AdbArgs) { return "" }
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        return (& adb @AdbArgs 2>&1 | Out-String)
    } finally {
        $ErrorActionPreference = $prev
    }
}

function Get-DeviceLogs {
    return Get-StreamPassDeviceLogs
}

function Count-Pattern([string]$Text, [string]$Pattern) {
    if (-not $Text) { return 0 }
    return ([regex]::Matches($Text, $Pattern)).Count
}

Write-Host "=== StreamPass Traffic Block Diagnosis ===" -ForegroundColor Cyan
Write-Host ""

if ($WithUnit) {
    Write-Host "--- Unit: traffic path ---" -ForegroundColor Cyan
    Push-Location (Join-Path $Root "client\go_core")
    go test ./internal/tunbridge/ -run TrafficPathDiagnosis -count=1 2>&1 | Out-Null
    if ($LASTEXITCODE -ne 0) {
        Add-Finding -Layer "unit" -Check "TrafficPathDiagnosis" -Status "FAIL" -Detail "go test failed" -IsBlocker
    } else {
        Add-Finding -Layer "unit" -Check "TrafficPathDiagnosis" -Status "PASS"
    }
    Pop-Location
}

Write-Host "--- Layer 0: Backend ---" -ForegroundColor Cyan
try {
    $health = Invoke-WebRequest -Uri "$BaseUrl/health" -UseBasicParsing -TimeoutSec 12
    if ($health.StatusCode -eq 200) {
        Add-Finding -Layer "backend" -Check "health" -Status "PASS" -Detail $health.Content
    } else {
        Add-Finding -Layer "backend" -Check "health" -Status "FAIL" -Detail "HTTP $($health.StatusCode)" -IsBlocker
    }
} catch {
    Add-Finding -Layer "backend" -Check "health" -Status "FAIL" -Detail $_.Exception.Message -IsBlocker
}

$relayHost = "212.43.156.33"
foreach ($port in @(443, 8443, 24443)) {
    try {
        $tcp = New-Object System.Net.Sockets.TcpClient
        $iar = $tcp.BeginConnect($relayHost, $port, $null, $null)
        $ok = $iar.AsyncWaitHandle.WaitOne(5000, $false)
        if ($ok -and $tcp.Connected) {
            Add-Finding -Layer "relay" -Check "TCP :$port" -Status "PASS"
            $tcp.Close()
        } else {
            Add-Finding -Layer "relay" -Check "TCP :$port" -Status "FAIL" -Detail "timeout" -IsBlocker
            $tcp.Close()
        }
    } catch {
        Add-Finding -Layer "relay" -Check "TCP :$port" -Status "FAIL" -Detail $_.Exception.Message -IsBlocker
    }
}

Write-Host "--- Layer 1: Device ---" -ForegroundColor Cyan
if (-not (Get-Command adb -ErrorAction SilentlyContinue)) {
    Add-Finding -Layer "device" -Check "adb" -Status "FAIL" -Detail "adb not in PATH" -IsBlocker
    foreach ($f in $findings) { Write-FindingLine $f }
    exit 1
}

$devOut = adb devices 2>&1 | Out-String
if ($devOut -match "`tunauthorized") {
    Add-Finding -Layer "device" -Check "adb authorized" -Status "FAIL" -Detail "Allow USB debugging" -IsBlocker
} elseif ($devOut -match "`tdevice") {
    Add-Finding -Layer "device" -Check "adb" -Status "PASS"
} else {
    Add-Finding -Layer "device" -Check "adb" -Status "FAIL" -Detail "no device - connect phone" -IsBlocker
}

$logs = ""
$hasDevice = @($findings | Where-Object { $_.Layer -eq "device" -and $_.Status -eq "PASS" }).Count -gt 0
if ($hasDevice) {
    Initialize-StreamPassLogCapture
    Write-StreamPassLogSourceNote
    $logs = Get-DeviceLogs
    if ([string]::IsNullOrWhiteSpace($logs)) {
        Add-Finding -Layer "device" -Check "connect logs" -Status "WARN" -Detail "empty - connect VPN in app first"
    } else {
        Add-Finding -Layer "device" -Check "connect logs" -Status "PASS" -Detail "$($logs.Length) chars"
    }
    if (-not $AllowNoVpn -and -not (Test-StreamPassVpnConnected $logs)) {
        Add-Finding -Layer "vpn" -Check "pre-check" -Status "FAIL" -Detail "VPN not connected - connect in StreamPass first" -IsBlocker
    }
}

Write-Host "--- Layer 2: VPN connect ---" -ForegroundColor Cyan
if ($logs) {
    if ($logs -match "tunnel event=connected|PrepareRelay OK") {
        Add-Finding -Layer "vpn" -Check "tunnel connected" -Status "PASS"
    } else {
        Add-Finding -Layer "vpn" -Check "tunnel connected" -Status "FAIL" -Detail "not connected" -IsBlocker
    }

    if ($logs -match "vpn dns=10\.10\.0\.1 \(Go dnscache\)") {
        Add-Finding -Layer "vpn" -Check "DNS via TUN" -Status "PASS" -Detail "10.10.0.1 Go dnscache (HostForIP works)"
    } elseif ($logs -match "vpn dns=77\.88\.8\.8") {
        Add-Finding -Layer "vpn" -Check "DNS via TUN" -Status "FAIL" -Detail "OLD APK: Yandex OS DNS bypasses Go -> host= empty -> foreign DIRECT" -IsBlocker
    } else {
        Add-Finding -Layer "vpn" -Check "DNS via TUN" -Status "WARN" -Detail "vpn dns= line missing - rebuild APK"
    }

    if ($logs -match "protect\(fd=\d+\)=false") {
        Add-Finding -Layer "vpn" -Check "socket protect" -Status "FAIL" -Detail "protect=false -> hysteria loop, no traffic" -IsBlocker
    } elseif ($logs -match "protect\(fd=\d+\)=true") {
        Add-Finding -Layer "vpn" -Check "socket protect" -Status "PASS"
    } else {
        Add-Finding -Layer "vpn" -Check "socket protect" -Status "WARN" -Detail "no protect lines"
    }

    if ($logs -match "split-tunnel mode=exclude-ru|split-tunnel mode=intl-only") {
        Add-Finding -Layer "vpn" -Check "split tunnel" -Status "PASS"
    } else {
        Add-Finding -Layer "vpn" -Check "split tunnel" -Status "WARN" -Detail "split-tunnel line missing"
    }
}

Write-Host "--- Layer 3: Decision / traffic path ---" -ForegroundColor Cyan
if ($logs) {
    $mustRelayFail = Count-Pattern $logs '\[tun\] must-relay fail'
    $relayFail = Count-Pattern $logs '\[tun\] relay-tcp fail|\[tun\].*relay.*fail'
    $relayBlackhole = Count-Pattern $logs 'relay_blackhole'
    $fallback = Count-Pattern $logs 'FALLBACK after relay fail|fallback_after_relay_fail'
    $directFail = Count-Pattern $logs '\[tun\] direct-tcp fail'

    if ($mustRelayFail -gt 0) {
        Add-Finding -Layer "traffic" -Check "must-relay fail" -Status "FAIL" -Detail "count=$mustRelayFail relay path dead" -IsBlocker
    } else {
        Add-Finding -Layer "traffic" -Check "must-relay fail" -Status "PASS" -Detail "count=0"
    }

    if ($relayBlackhole -gt 0) {
        Add-Finding -Layer "traffic" -Check "relay blackhole" -Status "FAIL" -Detail "count=$relayBlackhole dial ok but no bytes" -IsBlocker
    }

    if ($relayFail -gt 0) {
        Add-Finding -Layer "traffic" -Check "relay dial errors" -Status "WARN" -Detail "count=$relayFail"
    }

    if ($fallback -gt 0) {
        Add-Finding -Layer "traffic" -Check "relay fallback to DIRECT" -Status "WARN" -Detail "count=$fallback -> RU IP on FALLBACK rules"
    }

    # [decision] host= ip=142.250.x action=DIRECT reason=default_direct
    $foreignDirect = [regex]::Matches($logs, '\[decision\]\s+host=\s+ip=(157\.240|142\.250|172\.217|104\.16)[^\s]+\s+rule=[^\s]+\s+action=DIRECT')
    if ($foreignDirect.Count -gt 0) {
        Add-Finding -Layer "traffic" -Check "foreign CDN -> DIRECT" -Status "FAIL" -Detail "count=$($foreignDirect.Count) IP-only, host empty (DNS/HostForIP broken)" -IsBlocker
    } else {
        $relayDecisions = Count-Pattern $logs '\[decision\][^\n]*action=RELAY'
        $directDecisions = Count-Pattern $logs '\[decision\][^\n]*action=DIRECT'
        if ($relayDecisions -eq 0 -and $directDecisions -eq 0) {
            Add-Finding -Layer "traffic" -Check "decision log" -Status "WARN" -Detail "no [decision] lines - open foreign site or -LiveProbe"
        } elseif ($relayDecisions -gt 0) {
            Add-Finding -Layer "traffic" -Check "decision log" -Status "PASS" -Detail "RELAY=$relayDecisions DIRECT=$directDecisions"
        } else {
            Add-Finding -Layer "traffic" -Check "decision log" -Status "WARN" -Detail "only DIRECT decisions ($directDecisions)"
        }
    }

    # Named foreign hosts going DIRECT (should be RELAY)
    $badHosts = @()
    foreach ($m in [regex]::Matches($logs, '\[decision\]\s+host=([\w\.-]+)\s+ip=[^\s]+\s+rule=[^\s]+\s+action=DIRECT')) {
        $h = $m.Groups[1].Value.ToLower()
        if ($h -match 'youtube|instagram|google|gemini|facebook|cdninstagram|googlevideo|chatgpt|openai') {
            $badHosts += $h
        }
    }
    if ($badHosts.Count -gt 0) {
        $uniq = ($badHosts | Select-Object -Unique) -join ', '
        Add-Finding -Layer "traffic" -Check "must-relay host -> DIRECT" -Status "FAIL" -Detail $uniq -IsBlocker
    }

    if ($directFail -gt 3) {
        Add-Finding -Layer "traffic" -Check "direct-tcp fail" -Status "WARN" -Detail "count=$directFail"
    }
}

if ($LiveProbe -and $hasDevice) {
    Write-Host "--- Layer 4: Live probe ---" -ForegroundColor Cyan
    $probeMarker = "DIAGPROBE-$(Get-Date -Format 'HHmmss')"
    Invoke-AdbQuiet @('shell', 'log', '-t', $probeMarker, 'marker')
    foreach ($url in @('https://www.youtube.com', 'https://gemini.google.com')) {
        Invoke-AdbQuiet @('shell', 'am', 'start', '-a', 'android.intent.action.VIEW', '-d', $url)
        Start-Sleep -Seconds ([math]::Max(5, $ProbeWaitSeconds / 2))
    }
    Start-Sleep -Seconds 3
    $probe = Get-StreamPassDeviceLogs -SinceMarker $probeMarker
    if ([string]::IsNullOrWhiteSpace($probe)) {
        $probe = Get-DeviceLogs
    }
    $probe = "$logs`n$probe"

    $probeRelay = Count-Pattern $probe '\[decision\][^\n]*action=RELAY'
    $probeMustFail = Count-Pattern $probe '\[tun\] must-relay fail'
    $probeForeignDirect = Count-Pattern $probe '\[decision\]\s+host=\s+ip=(157\.240|142\.250)'

    if ($probeMustFail -gt 0) {
        Add-Finding -Layer "probe" -Check "live relay" -Status "FAIL" -Detail "must-relay fail during probe" -IsBlocker
    } elseif ($probeRelay -gt 0) {
        Add-Finding -Layer "probe" -Check "live relay" -Status "PASS" -Detail "RELAY decisions=$probeRelay"
    } elseif ($probeForeignDirect -gt 0) {
        Add-Finding -Layer "probe" -Check "live relay" -Status "FAIL" -Detail "foreign CDN still DIRECT (rebuild APK + reconnect)" -IsBlocker
    } else {
        Add-Finding -Layer "probe" -Check "live relay" -Status "WARN" -Detail "no RELAY decisions captured"
    }
}

Write-Host ""
Write-Host "=== Findings ===" -ForegroundColor Cyan
foreach ($f in $findings) { Write-FindingLine $f }

Write-Host ""
Write-Host "=== Blockers ===" -ForegroundColor Red
if ($blockers.Count -eq 0) {
    Write-Host "  (none)" -ForegroundColor Green
} else {
    foreach ($b in $blockers) { Write-Host "  * $b" -ForegroundColor Red }
}

Write-Host ""
Write-Host "=== Root cause cheat sheet ===" -ForegroundColor Cyan
Write-Host "  A) vpn dns=77.88.8.8 (old APK) -> rebuild + reconnect"
Write-Host "  B) [decision] host= empty + foreign IP + DIRECT -> DNS not through Go dnscache"
Write-Host "  C) [tun] must-relay fail -> relay/Hysteria broken (check relay TCP 443/8443)"
Write-Host "  D) protect(fd)=false -> routing loop, zero traffic"
Write-Host "  E) must-relay host -> DIRECT -> rules not loaded or wrong host"

if ($ReportPath) {
    $dir = Split-Path -Parent (Join-Path $Root $ReportPath)
    if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
    $full = Join-Path $Root $ReportPath
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm"
    $md = @("# Traffic block diagnosis", "", "Generated: $ts", "", "| Layer | Check | Status | Detail |", "|-------|-------|--------|--------|")
    foreach ($f in $findings) {
        $d = ($f.Detail -replace '\|', '/') -replace "`n", " "
        $md += "| $($f.Layer) | $($f.Check) | $($f.Status) | $d |"
    }
    $md += "", "## Blockers"
    if ($blockers.Count -eq 0) { $md += "- none" } else { foreach ($b in $blockers) { $md += "- $b" } }
    [System.IO.File]::WriteAllText($full, ($md -join "`n"), [System.Text.Encoding]::UTF8)
    Write-Host ""
    Write-Host "Report: $full" -ForegroundColor Green
}

Write-Host ""
if ($blockers.Count -gt 0) {
    Write-Host "VERDICT: TRAFFIC BLOCKED ($($blockers.Count) blocker(s))" -ForegroundColor Red
    exit 1
}
$vpnConnected = Test-StreamPassVpnConnected $logs
if (-not $vpnConnected) {
    Write-Host "VERDICT: INCONCLUSIVE - VPN not connected or connect lines missing from logcat (not a traffic PASS)" -ForegroundColor Yellow
    exit 2
}
if ($LiveProbe -and ($findings | Where-Object { $_.Layer -eq "probe" -and $_.Status -eq "WARN" }).Count -gt 0) {
    Write-Host "VERDICT: VPN up, no hard blockers, but live traffic proof incomplete (check foreign sites manually)" -ForegroundColor Yellow
    exit 2
}
Write-Host "VERDICT: no critical blockers in captured logs" -ForegroundColor Green
exit 0
