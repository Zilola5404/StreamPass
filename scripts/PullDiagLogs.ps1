# Pull operator routing diagnostics from StreamPass backend.
# Usage:
#   $env:ADMIN_API_KEY = '...'
#   .\scripts\PullDiagLogs.ps1
#   .\scripts\PullDiagLogs.ps1 -UserId '<uuid>' -Limit 200 -OutFile .\diag.json
param(
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$AdminKey = $env:ADMIN_API_KEY,
    [string]$UserId = "",
    [int]$Limit = 100,
    [string]$OutFile = "",
    [switch]$FailsOnly
)

$ErrorActionPreference = "Stop"
if (-not $AdminKey) {
    throw "Set ADMIN_API_KEY or pass -AdminKey"
}

$api = $BaseUrl.TrimEnd('/')
if ($api -notmatch '/api/v1$') { $api = "$api/api/v1" }

$qs = "limit=$Limit"
if ($UserId) { $qs += "&user_id=$([uri]::EscapeDataString($UserId))" }
$url = "$api/admin/diag?$qs"

Write-Host "GET $url"
$headers = @{ "X-Admin-Key" = $AdminKey }
$events = Invoke-RestMethod -Method GET -Uri $url -Headers $headers

if ($FailsOnly) {
    $events = @($events | Where-Object { $_.result -ne 'ok' })
}

# Compact table for terminal review
$events | Select-Object recorded_at, user_id, proto, host, dest_ip, dest_port, mode, result, latency_ms, error_code |
    Format-Table -AutoSize

if ($OutFile) {
    $events | ConvertTo-Json -Depth 6 | Set-Content -Path $OutFile -Encoding UTF8
    Write-Host "Wrote $($events.Count) events to $OutFile"
} else {
    Write-Host "Total: $($events.Count) events (pass -OutFile to save JSON)"
}
