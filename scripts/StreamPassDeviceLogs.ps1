# Shared adb log helpers for StreamPass QA scripts (dot-source from scripts/*.ps1).

function Initialize-StreamPassLogCapture {
    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        adb logcat -G 16M 2>&1 | Out-Null
    } finally {
        $ErrorActionPreference = $prev
    }
    $script:StreamPassPersistLogAvailable = $false
    $probe = adb exec-out run-as com.streampass.app id 2>&1 | Out-String
    if ($probe -and ($probe -notmatch "run-as|not debuggable|Unknown package")) {
        $script:StreamPassPersistLogAvailable = $true
    }
}

function Get-StreamPassDeviceLogs {
    param([string]$SinceMarker = "")

    $prev = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $logcat = adb logcat -d -s StreamPassConnect StreamPassVpn StreamPassTunnel 2>&1
        $appLog = ""
        if ($script:StreamPassPersistLogAvailable) {
            $appLog = adb exec-out run-as com.streampass.app cat files/connect.log 2>&1 | Out-String
        }
    } finally {
        $ErrorActionPreference = $prev
    }

    $parts = @(@($logcat) -join "`n")
    if ($appLog -and ($appLog -notmatch "run-as|not debuggable")) {
        $parts += $appLog
    }
    $text = ($parts | Where-Object { $_ }) -join "`n"

    if ($SinceMarker -and $text -match [regex]::Escape($SinceMarker)) {
        $idx = $text.LastIndexOf($SinceMarker)
        if ($idx -ge 0) { return $text.Substring($idx) }
    }
    return $text
}

function Test-StreamPassVpnConnected {
    param([string]$Logs)
    if (-not $Logs) { return $false }
    return ($Logs -match "tunnel event=connected")
}

function Write-StreamPassLogSourceNote {
    if ($script:StreamPassPersistLogAvailable) {
        Write-Host "[INFO] log source: logcat + connect.log (debuggable build)" -ForegroundColor Cyan
    } else {
        Write-Host "[WARN] log source: logcat ring-buffer only (release APK: run-as connect.log unavailable)" -ForegroundColor Yellow
        Write-Host "       Connect VPN before the run; avoid long idle gaps so connect lines stay in buffer." -ForegroundColor DarkYellow
    }
}

function Assert-StreamPassVpnConnected {
    param(
        [string]$Logs,
        [switch]$AllowNoVpn
    )
    if (Test-StreamPassVpnConnected $Logs) {
        return $true
    }
    if ($AllowNoVpn) {
        return $false
    }
    Write-Host "[FAIL] VPN not connected - open StreamPass, tap Connect, wait for Connected, then re-run." -ForegroundColor Red
    Write-Host "       tunnel event=connected not found in captured logs." -ForegroundColor DarkRed
    exit 1
}
