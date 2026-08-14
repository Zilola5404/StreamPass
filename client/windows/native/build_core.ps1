# Build streampasscore.exe (Wintun sidecar) and fetch official wintun.dll.
$ErrorActionPreference = 'Stop'
$NativeDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$GoCore = Resolve-Path (Join-Path $NativeDir '..\..\go_core')
$ExeOut = Join-Path $NativeDir 'streampasscore.exe'
$WintunOut = Join-Path $NativeDir 'wintun.dll'
$WintunVersion = '0.14.1'
$WintunDllSha256 = 'E5DA8447DC2C320EDC0FC52FA01885C103DE8C118481F683643CACC3220DAFCE'
$WintunUrls = @(
    "https://github.com/tun2proxy/wintun-bindings/raw/master/wintun/bin/amd64/wintun.dll",
    "https://www.wintun.net/builds/wintun-$WintunVersion.zip"
)

function Needs-Rebuild {
    if (-not (Test-Path $ExeOut)) { return $true }
    $exeTime = (Get-Item $ExeOut).LastWriteTimeUtc
    $latest = Get-ChildItem -Path $GoCore -Recurse -Include *.go,go.mod |
        Sort-Object LastWriteTimeUtc -Descending |
        Select-Object -First 1
    if ($null -eq $latest) { return $true }
    return $latest.LastWriteTimeUtc -gt $exeTime
}

if (-not (Test-Path $WintunOut)) {
    Write-Host "Downloading wintun $WintunVersion..."
    $tmp = Join-Path $NativeDir "wintun-download.bin"
    $downloaded = $false
    foreach ($url in $WintunUrls) {
        Write-Host "GET $url"
        if (Test-Path $tmp) { Remove-Item -Force $tmp }
        & curl.exe -L --fail --connect-timeout 20 --max-time 120 -o $tmp $url
        if ($LASTEXITCODE -ne 0 -or -not (Test-Path $tmp)) { continue }
        if ($url -like '*.dll') {
            Copy-Item $tmp $WintunOut -Force
            $downloaded = $true
            break
        }
        $extract = Join-Path $NativeDir '_wintun_extract'
        if (Test-Path $extract) { Remove-Item -Recurse -Force $extract }
        Expand-Archive -Path $tmp -DestinationPath $extract -Force
        $dll = Get-ChildItem -Path $extract -Recurse -Filter wintun.dll |
            Where-Object { $_.FullName -match 'amd64' } |
            Select-Object -First 1
        if ($null -eq $dll) { continue }
        Copy-Item $dll.FullName $WintunOut -Force
        Remove-Item -Recurse -Force $extract
        $downloaded = $true
        break
    }
    if (Test-Path $tmp) { Remove-Item -Force $tmp }
    if (-not $downloaded) { throw "failed to download wintun.dll (tried $($WintunUrls -join ', '))" }
}

$hash = (Get-FileHash $WintunOut -Algorithm SHA256).Hash
if ($hash -ne $WintunDllSha256) {
    throw "wintun.dll SHA-256 mismatch: $hash (want $WintunDllSha256)"
}
Write-Host "wintun.dll OK sha256=$hash"

if (Needs-Rebuild) {
    Write-Host "Building streampasscore.exe (tags=with_gvisor)..."
    $env:CGO_ENABLED = '0'
    Push-Location $GoCore
    try {
        & go build -tags with_gvisor -ldflags "-H windowsgui" -o $ExeOut ./desktop
        if ($LASTEXITCODE -ne 0) { throw "go build failed: $LASTEXITCODE" }
    } finally {
        Pop-Location
    }
    Write-Host "streampasscore.exe -> $ExeOut"
} else {
    Write-Host "streampasscore.exe is up to date"
}

$FlutterDebug = Join-Path $NativeDir '..\..\build\windows\x64\runner\Debug'
if (Test-Path $FlutterDebug) {
    try {
        Copy-Item $ExeOut (Join-Path $FlutterDebug 'streampasscore.exe') -Force -ErrorAction Stop
        Copy-Item $WintunOut (Join-Path $FlutterDebug 'wintun.dll') -Force -ErrorAction Stop
    } catch {
        Write-Host "WARN: could not copy to Debug (exe locked?): $_"
    }
}
$FlutterRelease = Join-Path $NativeDir '..\..\build\windows\x64\runner\Release'
if (Test-Path $FlutterRelease) {
    try {
        Copy-Item $ExeOut (Join-Path $FlutterRelease 'streampasscore.exe') -Force -ErrorAction Stop
        Copy-Item $WintunOut (Join-Path $FlutterRelease 'wintun.dll') -Force -ErrorAction Stop
    } catch {
        Write-Host "WARN: could not copy to Release (exe locked?): $_"
    }
}
