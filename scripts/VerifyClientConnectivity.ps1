# StreamPass client connectivity diagnostic
# Checks API, auth, subscription, relay list, and hysteria2 handshake.
#
# Usage:
#   .\scripts\VerifyClientConnectivity.ps1
#   .\scripts\VerifyClientConnectivity.ps1 -Email you@example.com -Password 'secret'
#   .\scripts\VerifyClientConnectivity.ps1 -SkipRelayHandshake

param(
    [string]$ApiBase = "https://212-43-156-33.nip.io/api/v1",
    [string]$Email = "",
    [string]$Password = "",
    [switch]$SkipRelayHandshake
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$GoCore = Join-Path $Root "client\go_core"
$Passed = 0
$Failed = 0
$Warn = 0

function Step-Pass($Name) { Write-Host "[PASS] $Name" -ForegroundColor Green; $script:Passed++ }
function Step-Fail($Name, $Detail) { Write-Host "[FAIL] $Name - $Detail" -ForegroundColor Red; $script:Failed++ }
function Step-Warn($Name, $Detail) { Write-Host "[WARN] $Name - $Detail" -ForegroundColor Yellow; $script:Warn++ }

Write-Host "=== StreamPass Client Connectivity ===" -ForegroundColor Cyan
Write-Host "API: $ApiBase`n"

# 1. Health (no auth)
try {
    $health = Invoke-RestMethod -Uri "$ApiBase/health" -Method GET
    if ($health.status -eq "ok") { Step-Pass "GET /health" } else { Step-Fail "GET /health" "unexpected body" }
} catch {
    Step-Fail "GET /health" $_.Exception.Message
}

# 2. Auth
$token = $null
if (-not $Email) {
    $Email = "diag-$(Get-Date -Format 'yyyyMMddHHmmss')@streampass.local"
    $Password = "DiagTest123!"
    Write-Host "Using auto test user: $Email" -ForegroundColor Gray
}

try {
    $regBody = @{ email = $Email; password = $Password } | ConvertTo-Json
    $reg = Invoke-RestMethod -Uri "$ApiBase/register" -Method POST -ContentType "application/json" -Body $regBody
    $token = $reg.access_token
    Step-Pass "POST /register (or existing user login next)"
} catch {
    try {
        $loginBody = @{ email = $Email; password = $Password } | ConvertTo-Json
        $login = Invoke-RestMethod -Uri "$ApiBase/login" -Method POST -ContentType "application/json" -Body $loginBody
        $token = $login.access_token
        Step-Pass "POST /login"
    } catch {
        Step-Fail "Auth" $_.Exception.Message
    }
}

if (-not $token) {
    Write-Host "`nCannot continue without token." -ForegroundColor Red
    exit 1
}

$headers = @{ Authorization = "Bearer $token" }

# 3. Subscription — VPN connect is blocked when INACTIVE (home_screen.dart)
try {
    $sub = Invoke-RestMethod -Uri "$ApiBase/subscription" -Headers $headers
    $status = $sub.status
    if ($status -eq "ACTIVE") {
        Step-Pass "GET /subscription - ACTIVE (VPN allowed)"
    } else {
        Step-Warn "GET /subscription - $status" "Client blocks VPN until subscription is active. Activate in DB: UPDATE users SET subscription_active_until = NOW() + INTERVAL '30 days' WHERE email = '$Email';"
    }
} catch {
    Step-Fail "GET /subscription" $_.Exception.Message
}

# 4. Relay servers
$relayUri = $null
try {
    $servers = Invoke-RestMethod -Uri "$ApiBase/servers" -Headers $headers
    if ($servers -is [System.Array] -and $servers.Count -gt 0) {
        Step-Pass "GET /servers - $($servers.Count) relay(s)"
        foreach ($s in $servers) {
            $cfg = [string]$s.connection_config
            Write-Host "  • $($s.id) $($s.host):$($s.port) healthy=$($s.healthy)" -ForegroundColor Gray
            if ($cfg -match '^hysteria2://') {
                Step-Pass "  connection_config for $($s.id) is hysteria2://"
                if (-not $relayUri) { $relayUri = $cfg }
            } elseif ($cfg -match '^https?://') {
                Step-Fail "  connection_config for $($s.id)" "Hiddify subscription URL - go_core needs hysteria2://"
            } elseif ([string]::IsNullOrWhiteSpace($cfg)) {
                Step-Fail "  connection_config for $($s.id)" "empty - tunnel will fail without auth/obfs"
            } else {
                Step-Fail "  connection_config for $($s.id)" "unsupported scheme"
            }
        }
    } else {
        Step-Fail "GET /servers" "empty list - no healthy relays (health monitor marks UDP-only relays unhealthy via TCP probe)"
    }
} catch {
    Step-Fail "GET /servers" $_.Exception.Message
}

# 5. AAR
$aar = Join-Path $Root "client\android\app\libs\streampasscore.aar"
if (Test-Path $aar) {
    $mb = [math]::Round((Get-Item $aar).Length / 1MB, 1)
    Step-Pass "streampasscore.aar ($mb MB)"
} else {
    Step-Fail "streampasscore.aar" "missing - rebuild go_core AAR"
}

# 6. Hysteria handshake (same path as Android go_core)
if (-not $SkipRelayHandshake -and $relayUri) {
    $env:STREAMPASS_RELAY_URI = $relayUri
    Push-Location $GoCore
    try {
        go test -timeout 2m -run "TestIntegrationHysteriaConnect" ./internal/hyconfig/ 2>&1 | Out-Host
        if ($LASTEXITCODE -eq 0) {
            Step-Pass "Hysteria handshake (go_core integration)"
        } else {
            Step-Fail "Hysteria handshake" "exit $LASTEXITCODE - relay unreachable or bad connection_config"
        }
    } finally { Pop-Location }
} elseif ($SkipRelayHandshake) {
    Write-Host "[SKIP] Hysteria handshake -SkipRelayHandshake" -ForegroundColor Yellow
} else {
    Step-Warn "Hysteria handshake" "no relay URI to test"
}

Write-Host "`n=== Results: $Passed passed, $Failed failed, $Warn warnings ===" -ForegroundColor $(if ($Failed -eq 0) { "Green" } else { "Red" })
if ($Failed -gt 0) { exit 1 }
