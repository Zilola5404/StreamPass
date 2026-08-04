# StreamPass Load Test (BL-032)
# Usage:
#   .\scripts\LoadTest.ps1
#   .\scripts\LoadTest.ps1 -BaseUrl https://212-43-156-33.nip.io -Duration 20s -Rps 40
#   .\scripts\LoadTest.ps1 -Email user@example.com -Password secret

param(
    [string]$BaseUrl = "https://212-43-156-33.nip.io",
    [string]$Duration = "15s",
    [int]$Rps = 30,
    [string]$Email = "",
    [string]$Password = "",
    [string]$AdminKey = ""
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

$argsList = @(
    "run", "./scripts/loadtest",
    "-base", $BaseUrl,
    "-duration", $Duration,
    "-rps", "$Rps"
)
if ($Email) { $argsList += @("-email", $Email) }
if ($Password) { $argsList += @("-password", $Password) }
if ($AdminKey) { $argsList += @("-admin-key", $AdminKey) }

Write-Host "=== StreamPass LoadTest ===" -ForegroundColor Cyan
Push-Location $Root
try {
    & go @argsList
    if ($LASTEXITCODE -ne 0) { throw "loadtest failed with exit $LASTEXITCODE" }
}
finally {
    Pop-Location
}
