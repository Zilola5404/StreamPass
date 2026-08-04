# StreamPass PostgreSQL Backup (BL-033)
# Usage:
#   .\scripts\Backup.ps1
#   .\scripts\Backup.ps1 -OutputDir C:\backups\streampass -RetentionDays 30

param(
    [string]$OutputDir = "",
    [int]$RetentionDays = 30
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if (-not $OutputDir) {
    $OutputDir = Join-Path $Root "backups"
}

$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$BackupFile = Join-Path $OutputDir "streampass_$Timestamp.sql.gz"
$PlainTmp = Join-Path $OutputDir "streampass_$Timestamp.sql"

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

Write-Host "=== StreamPass Backup ===" -ForegroundColor Cyan
Write-Host "Output: $BackupFile"

Push-Location $Root
try {
    $cid = (docker compose ps -q postgres | Select-Object -First 1)
    if (-not $cid) { throw "postgres container not running (start compose locally or use scripts/backup-postgres.sh on the VPS)" }

    # Write SQL via docker redirect into a temp file (binary-safe), then gzip.
    docker exec -i $cid pg_dump -U streampass -d streampass --no-owner --no-acl | Set-Content -Path $PlainTmp -Encoding utf8NoBOM
    if ($LASTEXITCODE -ne 0) { throw "pg_dump failed" }

    if (Get-Command gzip -ErrorAction SilentlyContinue) {
        & gzip -f -c $PlainTmp | Set-Content -Path $BackupFile -AsByteStream
    } else {
        # .NET gzip fallback when gzip.exe is unavailable
        $bytes = [System.IO.File]::ReadAllBytes($PlainTmp)
        $ms = New-Object System.IO.MemoryStream
        $gz = New-Object System.IO.Compression.GzipStream($ms, [System.IO.Compression.CompressionMode]::Compress)
        $gz.Write($bytes, 0, $bytes.Length)
        $gz.Dispose()
        [System.IO.File]::WriteAllBytes($BackupFile, $ms.ToArray())
        $ms.Dispose()
    }
    Remove-Item $PlainTmp -Force -ErrorAction SilentlyContinue
}
finally {
    Pop-Location
}

if (-not (Test-Path $BackupFile) -or (Get-Item $BackupFile).Length -lt 32) {
    throw "backup file missing or too small: $BackupFile"
}

$SizeKb = [math]::Round((Get-Item $BackupFile).Length / 1KB, 1)
Write-Host "Backup complete: $BackupFile ($SizeKb KB)" -ForegroundColor Green

Get-ChildItem $OutputDir -Filter "streampass_*.sql.gz" |
    Where-Object { $_.LastWriteTime -lt (Get-Date).AddDays(-$RetentionDays) } |
    ForEach-Object {
        Remove-Item $_.FullName -Force
        Write-Host "Removed old: $($_.Name)"
    }

$latest = Join-Path $OutputDir "streampass_latest.sql.gz"
Copy-Item $BackupFile $latest -Force
Write-Host "Latest: $latest"
