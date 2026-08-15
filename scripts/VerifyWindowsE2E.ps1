# VerifyWindowsE2E.ps1 - AUDIT-WIN-001 physical Windows E2E checklist
#
# Separates test levels (audit section 10) - never merge into a single PASS.
#
# Usage:
#   powershell -NoProfile -ExecutionPolicy Bypass -File scripts\VerifyWindowsE2E.ps1
#   powershell -NoProfile -ExecutionPolicy Bypass -File scripts\VerifyWindowsE2E.ps1 -ReportPath reports\QA\win-e2e.md
#
param(
    [string]$ReportPath = "",
    [string]$LogPath = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Stamp = Get-Date -Format "yyyy-MM-dd HH:mm"
$IsAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole(
    [Security.Principal.WindowsBuiltInRole]::Administrator)

function Write-Level([string]$level, [string]$name, [string]$status, [string]$detail = "") {
    $color = switch ($status) {
        "PASS" { "Green" }
        "FAIL" { "Red" }
        "NOT RUN" { "DarkGray" }
        "MANUAL" { "Yellow" }
        default { "Cyan" }
    }
    $suffix = if ($detail) { " - $detail" } else { "" }
    Write-Host ("[{0}] {1}/{2}{3}" -f $status, $level, $name, $suffix) -ForegroundColor $color
}

Write-Host "=== AUDIT-WIN-001 Windows E2E ===" -ForegroundColor Cyan
Write-Host "Host: $env:COMPUTERNAME | Admin: $IsAdmin | $Stamp"
Write-Host ""

# --- Automated sub-layers (delegate to VerifyWindowsTUN where possible) ---
Write-Host "--- Automated sub-layers ---" -ForegroundColor Cyan
$unitPass = $false
$ipcPass = $false
$liveTun = "NOT RUN"

Push-Location (Join-Path $Root "client\go_core")
try {
    $env:CGO_ENABLED = "0"
    & go test ./internal/decision/ -run "WindowsSplit|TestDecide" -count=1 2>&1 | Out-Null
    $unitPass = ($LASTEXITCODE -eq 0)
} finally { Pop-Location }

$exe = Join-Path $Root "client\windows\native\streampasscore.exe"
if (-not (Test-Path $exe)) {
    & powershell -NoProfile -ExecutionPolicy Bypass -File (Join-Path $Root "client\windows\native\build_core.ps1") | Out-Null
}
if (Test-Path $exe) {
    $portFile = Join-Path $env:TEMP ("sp-e2e-{0}.json" -f [guid]::NewGuid().ToString("N"))
    $p = Start-Process -FilePath $exe -ArgumentList @("--port-file", $portFile) -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 2
    if (Test-Path $portFile) {
        try {
            $meta = Get-Content $portFile -Raw | ConvertFrom-Json
            $tcp = New-Object System.Net.Sockets.TcpClient
            $tcp.Connect("127.0.0.1", [int]$meta.port)
            $s = $tcp.GetStream()
            $msg = (@{ id = 1; cmd = "ping"; token = $meta.token } | ConvertTo-Json -Compress) + "`n"
            $b = [Text.Encoding]::UTF8.GetBytes($msg)
            $s.Write($b, 0, $b.Length)
            $buf = New-Object byte[] 2048
            $n = $s.Read($buf, 0, $buf.Length)
            $resp = [Text.Encoding]::UTF8.GetString($buf, 0, $n)
            $ipcPass = ($resp -match '"ok"\s*:\s*true')
            $tcp.Close()
        } catch {}
    }
    Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
    Remove-Item $portFile -Force -ErrorAction SilentlyContinue
}

if ($IsAdmin) {
    $liveTun = "MANUAL (run VerifyWindowsTUN.ps1 -Admin live section)"
} else {
    $liveTun = "SKIPPED (need Administrator for Wintun create)"
}

Write-Level "Unit" "Decision/DNS matrix" $(if ($unitPass) { "PASS" } else { "FAIL" })
Write-Level "Integration" "Sidecar IPC ping" $(if ($ipcPass) { "PASS" } else { "FAIL" })
Write-Level "Live Device" "Wintun create/stop" $liveTun

Write-Host ""
Write-Host "--- Browser E2E (MANUAL - required for audit close) ---" -ForegroundColor Cyan
$manual = @(
    @{ Level = "Browser E2E"; Name = "DIRECT ya.ru / 2ip.ru"; Expect = "Decision=DIRECT, page loads, RU IP on 2ip" },
    @{ Level = "Browser E2E"; Name = "RELAY youtube / github"; Expect = "Decision=RELAY, first_byte>0, bytes>0" },
    @{ Level = "Browser E2E"; Name = "DNS HostForIP"; Expect = "[dns] + host= non-empty on foreign flows" },
    @{ Level = "Browser E2E"; Name = "Split routing"; Expect = "2ip=ISP, foreign=relay egress" },
    @{ Level = "Relay E2E"; Name = "Relay down"; Expect = "DIRECT RU still works" },
    @{ Level = "Lifecycle"; Name = "Disconnect/reconnect"; Expect = "clean stop, no stale 0.0.0.0/0 route" }
)
foreach ($m in $manual) {
    Write-Level $m.Level $m.Name "NOT RUN" $m.Expect
}

Write-Host ""
Write-Host "Steps:" -ForegroundColor Yellow
Write-Host "  1. Run StreamPass as Administrator, connect VPN"
Write-Host "  2. Open sites in Chrome; confirm load + geo"
Write-Host "  3. Export logs: Diagnostics -> copy OR sidecar log file"
Write-Host "  4. Mark each row PASS/FAIL in the report below"

if (-not $LogPath) {
    $LogPath = Join-Path $Root ("reports\QA\win-e2e-log-{0}.txt" -f (Get-Date -Format "yyyyMMdd-HHmm"))
}
Write-Host ""
Write-Host "Optional log capture (after connect):" -ForegroundColor Cyan
Write-Host "  Get-Content `$env:LOCALAPPDATA\...\connect.log  # or sidecar stdout"

$reportFull = if ($ReportPath) { Join-Path $Root $ReportPath } else {
    Join-Path $Root ("reports\QA\win-e2e-{0}.md" -f (Get-Date -Format "yyyy-MM-dd"))
}
$dir = Split-Path -Parent $reportFull
if ($dir -and -not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }

$md = @(
    "# Windows E2E report (AUDIT-WIN-001)",
    "",
    "Generated: $Stamp",
    "",
    "Host: $env:COMPUTERNAME | Admin: $IsAdmin",
    "",
    "## Test levels (do not merge)",
    "",
    "| Level | Check | Status | Notes |",
    "|-------|-------|--------|-------|",
    "| Unit | Decision/DNS matrix | $(if ($unitPass) { 'PASS' } else { 'FAIL' }) | go test |",
    "| Integration | Sidecar IPC | $(if ($ipcPass) { 'PASS' } else { 'FAIL' }) | ping cmd |",
    "| Live Device | Wintun create | $liveTun | VerifyWindowsTUN.ps1 Admin |",
    "| Browser E2E | DIRECT ya.ru/2ip | NOT RUN | manual |",
    "| Browser E2E | RELAY youtube/github | NOT RUN | manual |",
    "| Browser E2E | DNS HostForIP | NOT RUN | manual |",
    "| Browser E2E | Split routing | NOT RUN | manual |",
    "| Relay E2E | Relay down | NOT RUN | manual |",
    "| Lifecycle | Disconnect/reconnect | NOT RUN | manual |",
    "",
    "## Verdict",
    "",
    "**Engineering Validation** - browser E2E NOT RUN until manual rows filled.",
    "",
    "Reference: reports/Audit/AUDIT-WIN-001-Windows-Client.md"
)
[System.IO.File]::WriteAllText($reportFull, ($md -join "`n"), [System.Text.Encoding]::UTF8)
Write-Host "Report: $reportFull" -ForegroundColor Green

if (-not $unitPass -or -not $ipcPass) { exit 1 }
exit 2
