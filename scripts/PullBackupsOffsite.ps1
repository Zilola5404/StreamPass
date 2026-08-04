# Pull encrypted/plain StreamPass DB backups from the VPS onto this PC (off-site relative to the server).
# Usage:
#   .\scripts\PullBackupsOffsite.ps1
#   .\scripts\PullBackupsOffsite.ps1 -RemoteHost 212.43.156.33 -LocalDir C:\Backups\StreamPass

param(
    [string]$RemoteHost = "212.43.156.33",
    [string]$User = "root",
    [string]$RemoteDir = "/var/backups/streampass",
    [string]$LocalDir = "",
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
$plink = "C:\Program Files\PuTTY\plink.exe"
if (-not (Test-Path $pscp)) { throw "pscp not found: $pscp" }

if (-not $Password) {
    # Prefer env to avoid embedding secrets in scripts/history when possible.
    $Password = $env:STREAMPASS_SSH_PASSWORD
}
if (-not $Password) {
    throw "Set -Password or env STREAMPASS_SSH_PASSWORD"
}

Write-Host "=== Pull off-site backups ===" -ForegroundColor Cyan
Write-Host "From ${User}@${RemoteHost}:${RemoteDir}"
Write-Host "To   $LocalDir"

& $pscp -pw $Password -batch "$User@${RemoteHost}:$RemoteDir/streampass_*.sql.gz" $LocalDir
if ($LASTEXITCODE -ne 0) { throw "pscp failed" }

Get-ChildItem $LocalDir -Filter "streampass_*.sql.gz" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 5 |
    ForEach-Object { Write-Host ("  " + $_.Name + "  " + $_.Length + " bytes") }

Write-Host "OK: off-site copy on this machine" -ForegroundColor Green
