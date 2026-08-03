# Generates self-signed TLS certs for local hysteria-test relay.
$ErrorActionPreference = "Stop"
$Dir = Join-Path $PSScriptRoot "..\infrastructure\hysteria-test\certs"
New-Item -ItemType Directory -Force -Path $Dir | Out-Null

$openssl = Get-Command openssl -ErrorAction SilentlyContinue
if (-not $openssl) {
    $gitOpenssl = "C:\Program Files\Git\usr\bin\openssl.exe"
    if (Test-Path $gitOpenssl) { $openssl = $gitOpenssl } else { throw "openssl not found" }
} else {
    $openssl = $openssl.Source
}

& $openssl req -x509 -nodes -newkey rsa:2048 `
    -keyout (Join-Path $Dir "key.pem") `
    -out (Join-Path $Dir "cert.pem") `
    -days 3650 -subj "/CN=localhost" 2>&1 | Out-Null

Write-Host "Certs written to $Dir" -ForegroundColor Green
