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
# Use WebRequest + ConvertFrom-Json: Invoke-RestMethod on PS5.1 sometimes
# collapses JSON arrays into a single object with .value/.Count.
$resp = Invoke-WebRequest -Method GET -Uri $url -Headers @{ "X-Admin-Key" = $AdminKey } -UseBasicParsing
$parsed = $resp.Content | ConvertFrom-Json
if ($null -eq $parsed) {
    $events = @()
} elseif ($parsed -is [System.Array]) {
    $events = @($parsed)
} elseif ($parsed.PSObject.Properties.Name -contains 'value' -and $parsed.value -is [System.Array]) {
    # Tolerate accidental PS5 wrapper if someone re-uploaded a bad export
    $events = @($parsed.value)
} else {
    $events = @($parsed)
}

if ($FailsOnly) {
    $events = @($events | Where-Object { $_.result -ne 'ok' })
}

$count = $events.Count
Write-Host "Fetched $count event(s)"

if ($count -gt 0) {
    $events |
        Select-Object recorded_at, site, host, dest_ip, dest_port, mode, result, slow, latency_ms, reason, error_code, client_version |
        Format-Table -AutoSize |
        Out-Host
}

if ($OutFile) {
    $json = if ($count -eq 0) {
        '[]'
    } else {
        # Compress avoids huge indented dumps; Depth keeps nested fields.
        ConvertTo-Json -InputObject ([object[]]$events) -Depth 6
    }
    # If PS still wraps, unwrap once more to a raw JSON array string.
    if ($json -match '^\s*\{\s*"value"\s*:') {
        $wrap = $json | ConvertFrom-Json
        $json = ConvertTo-Json -InputObject ([object[]]$wrap.value) -Depth 6
    }
    Set-Content -Path $OutFile -Value $json -Encoding UTF8
    Write-Host "Wrote $count events to $OutFile"
} else {
    Write-Host "Total: $count events (pass -OutFile to save JSON)"
}
