# StreamPass PostgreSQL Backup
# Usage: .\scripts\Backup.ps1 [-OutputDir .\backups]

param(
    [string]$OutputDir = ".\backups"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$BackupFile = Join-Path $OutputDir "streampass_$Timestamp.sql"

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "=== StreamPass Backup ===" -ForegroundColor Cyan
Write-Host "Output: $BackupFile"

Push-Location $Root
docker compose exec -T postgres pg_dump -U streampass streampass | Out-File -FilePath $BackupFile -Encoding utf8
if ($LASTEXITCODE -ne 0) { throw "pg_dump failed" }
Pop-Location

$Size = (Get-Item $BackupFile).Length / 1KB
Write-Host "Backup complete: $BackupFile ($([math]::Round($Size, 1)) KB)" -ForegroundColor Green
