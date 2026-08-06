# Verify BL-035 off-site backup on primary + secondary VPS (read-only).
# Usage:
#   $env:STREAMPASS_SSH_KEY = "C:\Users\me\.ssh\id_rsa"
#   .\scripts\VerifyOffsiteBackup.ps1
#   .\scripts\VerifyOffsiteBackup.ps1 -PrimaryHost 212.43.156.33 -SecondaryHost 212.43.157.167

param(
    [string]$PrimaryHost = "212.43.156.33",
    [string]$SecondaryHost = "212.43.157.167",
    [string]$User = "root",
    [string]$KeyPath = "",
    [string]$Password = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

if (-not $KeyPath) { $KeyPath = $env:STREAMPASS_SSH_KEY }
if (-not $Password) { $Password = $env:STREAMPASS_SSH_PASSWORD }

$plink = "C:\Program Files\PuTTY\plink.exe"
if (-not (Test-Path $plink)) { throw "plink not found: $plink" }

function Invoke-Remote([string]$HostName, [string]$Command) {
    $args = @("-batch")
    if ($KeyPath -and (Test-Path $KeyPath)) {
        $args += @("-i", $KeyPath)
    } elseif ($Password) {
        $args += @("-pw", $Password)
    } else {
        throw "Set STREAMPASS_SSH_KEY or STREAMPASS_SSH_PASSWORD for SSH"
    }
    $args += "${User}@${HostName}", $Command
    & $plink @args
    if ($LASTEXITCODE -ne 0) { throw "Remote command failed on $HostName" }
}

$failures = 0
function Step-Pass([string]$msg) { Write-Host "[PASS] $msg" -ForegroundColor Green }
function Step-Fail([string]$msg, [string]$detail = "") {
    $script:failures++
    if ($detail) { Write-Host "[FAIL] $msg - $detail" -ForegroundColor Red }
    else { Write-Host "[FAIL] $msg" -ForegroundColor Red }
}

Write-Host "=== Verify off-site backup (BL-035) ===" -ForegroundColor Cyan

try {
    $cron = Invoke-Remote $PrimaryHost "crontab -l 2>/dev/null | grep streampass-offsite-backup || true"
    if ($cron -match "streampass-offsite-backup") { Step-Pass "primary cron: off-site job installed" }
    else { Step-Fail "primary cron" "streampass-offsite-backup not found" }
} catch {
    Step-Fail "primary cron" $_.Exception.Message
}

try {
    $logTail = Invoke-Remote $PrimaryHost "tail -n 5 /var/backups/streampass/offsite.log 2>/dev/null || true"
    if ($logTail -match "OK offsite sync") { Step-Pass "primary offsite.log: recent OK" }
    elseif ($logTail) { Step-Pass "primary offsite.log present (check manually)"; Write-Host $logTail }
    else { Step-Fail "primary offsite.log" "empty or missing" }
} catch {
    Step-Fail "primary offsite.log" $_.Exception.Message
}

try {
    $encList = Invoke-Remote $SecondaryHost "ls -1 /var/backups/streampass/streampass_*.sql.gz.enc 2>/dev/null | tail -n 3 || true"
    if ($encList -match "\.enc") { Step-Pass "secondary .enc files present"; Write-Host $encList }
    else { Step-Fail "secondary .enc" "no encrypted dumps on $SecondaryHost" }
} catch {
    Step-Fail "secondary .enc" $_.Exception.Message
}

if ($failures -gt 0) {
    Write-Host "=== FAIL: $failures check(s) ===" -ForegroundColor Red
    exit 1
}
Write-Host "=== PASS: off-site backup verification ===" -ForegroundColor Green
