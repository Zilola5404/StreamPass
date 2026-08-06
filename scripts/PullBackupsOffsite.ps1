# Pull encrypted/plain StreamPass DB backups from the VPS onto this PC (off-site relative to the server).
# Usage:
#   $env:STREAMPASS_SSH_KEY = "C:\Users\me\.ssh\id_rsa"
#   .\scripts\PullBackupsOffsite.ps1
#   .\scripts\PullBackupsOffsite.ps1 -RemoteHost 212.43.156.33 -LocalDir C:\Backups\StreamPass

param(
    [string]$RemoteHost = "212.43.156.33",
    [string]$User = "root",
    [string]$RemoteDir = "/var/backups/streampass",
    [string]$LocalDir = "",
    [string]$KeyPath = "",
    [string]$Password = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if (-not $LocalDir) {
    $LocalDir = Join-Path $Root "backups\offsite"
}
if (-not (Test-Path $LocalDir)) {
    New-Item -ItemType Directory -Path $LocalDir -Force | Out-Null
}

$pscp = "C:\Program Files\PuTTY\pscp.exe"
if (-not (Test-Path $pscp)) { throw "pscp not found: $pscp" }

if (-not $KeyPath) { $KeyPath = $env:STREAMPASS_SSH_KEY }
if (-not $Password) { $Password = $env:STREAMPASS_SSH_PASSWORD }

function Invoke-Pscp([string]$RemoteSpec, [string]$LocalPath) {
    $args = @("-batch")
    if ($KeyPath -and (Test-Path $KeyPath)) {
        $args += @("-i", $KeyPath)
    } elseif ($Password) {
        $args += @("-pw", $Password)
    } else {
        throw "Set STREAMPASS_SSH_KEY or STREAMPASS_SSH_PASSWORD"
    }
    $args += $RemoteSpec, $LocalPath
    & $pscp @args
    return $LASTEXITCODE
}

Write-Host "=== Pull off-site backups ===" -ForegroundColor Cyan
Write-Host "From ${User}@${RemoteHost}:${RemoteDir}"
Write-Host "To   $LocalDir"

$rc = Invoke-Pscp "$User@${RemoteHost}:$RemoteDir/streampass_*.sql.gz" $LocalDir
if ($rc -ne 0) {
    Write-Host "WARN: plain .sql.gz pull failed (may be empty)" -ForegroundColor Yellow
}

# Encrypted off-site mirror on the VPS (and/or pulled from primary)
$encRemote = "/var/backups/streampass-offsite"
$rc = Invoke-Pscp "$User@${RemoteHost}:${encRemote}/streampass_*.sql.gz.enc" $LocalDir
if ($rc -ne 0) {
    Write-Host "WARN: encrypted .enc pull failed" -ForegroundColor Yellow
}

Get-ChildItem $LocalDir -Filter "streampass_*" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 8 |
    ForEach-Object { Write-Host ("  " + $_.Name + "  " + $_.Length + " bytes") }

Write-Host "OK: off-site copy on this machine" -ForegroundColor Green
