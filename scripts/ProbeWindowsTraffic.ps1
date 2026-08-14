# Probe Windows TUN traffic: start sidecar, connect, hit sites, dump logs.
$ErrorActionPreference = 'Stop'
$Native = 'C:\01_Projects\StreamPass\client\windows\native'
$Exe = Join-Path $Native 'streampasscore.exe'
$Probe = "$env:TEMP\sp-win-probe.json"
$LogOut = 'C:\01_Projects\StreamPass\reports\QA\win-traffic-probe-log.txt'
$IsAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
Write-Host "Admin=$IsAdmin"
if (-not (Test-Path $Exe)) { throw "missing $Exe" }
if (-not (Test-Path $Probe)) { throw "missing $Probe - fetch API first" }

$cfg = Get-Content $Probe -Raw | ConvertFrom-Json
$portFile = Join-Path $env:TEMP ("sp-probe-port-{0}.json" -f [guid]::NewGuid().ToString('N'))
$logLines = New-Object System.Collections.Generic.List[string]
function L([string]$m) { $logLines.Add(("[{0}] {1}" -f (Get-Date -Format 'HH:mm:ss.fff'), $m)); Write-Host $m }

# kill stale
Get-Process streampasscore -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 500

# Elevate core (UAC) so Wintun CreateAdapter works
L 'launching elevated streampasscore (UAC)...'
try {
  $p = Start-Process -FilePath $Exe -ArgumentList @('--port-file', $portFile) -WorkingDirectory $Native -Verb RunAs -PassThru -WindowStyle Hidden
} catch {
  throw "UAC/elevation failed: $_"
}
if ($null -eq $p) { throw 'UAC cancelled' }
$meta = $null
$deadline = (Get-Date).AddSeconds(12)
while ((Get-Date) -lt $deadline) {
  if (Test-Path $portFile) {
    try { $meta = Get-Content $portFile -Raw | ConvertFrom-Json } catch {}
    if ($meta -and $meta.port) { break }
  }
  Start-Sleep -Milliseconds 40
}
if (-not $meta) { throw 'no port-file from core' }
L ("core port={0}" -f $meta.port)

$tcp = New-Object System.Net.Sockets.TcpClient
$tcp.Connect('127.0.0.1', [int]$meta.port)
$stream = $tcp.GetStream()
$reader = New-Object System.IO.StreamReader($stream)
$writer = New-Object System.IO.StreamWriter($stream)
$writer.AutoFlush = $true

$recvJob = {
  param($r, $list)
  while ($true) {
    $line = $r.ReadLine()
    if ($null -eq $line) { break }
    [void]$list.Add($line)
    Write-Host ("CORE {0}" -f $line)
  }
}
# sync read loop in background via runspace
$pool = [runspacefactory]::CreateRunspace()
$pool.Open()
$ps = [powershell]::Create()
$ps.Runspace = $pool
$null = $ps.AddScript({
  param($stream)
  $r = New-Object System.IO.StreamReader($stream)
  $buf = New-Object System.Collections.Concurrent.ConcurrentQueue[string]
  while ($true) {
    try {
      $line = $r.ReadLine()
      if ($null -eq $line) { break }
      [void]$buf.Enqueue($line)
      [Console]::WriteLine("CORE $line")
    } catch { break }
  }
  return $buf
}).AddArgument($stream)
# Simpler: poll DataAvailable in main loop instead

function Send-Cmd($obj) {
  $json = ($obj | ConvertTo-Json -Compress -Depth 6)
  $bytes = [Text.Encoding]::UTF8.GetBytes($json + "`n")
  $stream.Write($bytes, 0, $bytes.Length)
}

$start = @{
  id = 1
  cmd = 'start'
  token = $meta.token
  relayHost = $cfg.relayHost
  relayPort = [int]$cfg.relayPort
  connectionConfig = $cfg.connectionConfig
  rulesJson = $cfg.rulesJson
  exclusionsJson = '[]'
  networkMode = 'split'
  mtu = 1400
  blockUdp443 = $false
}
L ("start relay={0}:{1} mode=split" -f $cfg.relayHost, $cfg.relayPort)
Send-Cmd $start

$gotReply = $false
$ok = $false
$err = $null
$until = (Get-Date).AddSeconds(45)
$coreLog = New-Object System.Collections.Generic.List[string]
while ((Get-Date) -lt $until) {
  while ($stream.DataAvailable) {
    $line = $reader.ReadLine()
    if ($null -eq $line) { break }
    $coreLog.Add($line)
    L ("CORE {0}" -f $line)
    if ($line -match '"id"\s*:\s*1') {
      $gotReply = $true
      try {
        $j = $line | ConvertFrom-Json
        if ($j.ok) { $ok = $true } else { $err = $j.error }
      } catch {}
    }
  }
  if ($gotReply) { break }
  Start-Sleep -Milliseconds 100
}

if (-not $ok) {
  L ("START FAILED: {0}" -f $err)
  Get-NetAdapter | Format-Table Name, Status | Out-String | ForEach-Object { L $_ }
} else {
  L 'START OK - probing sites'
  Start-Sleep -Seconds 2
  Get-NetAdapter | Where-Object { $_.InterfaceDescription -match 'Wintun|StreamPass' -or $_.Name -match 'StreamPass' } | ForEach-Object {
    L ("adapter {0} status={1} idx={2}" -f $_.Name, $_.Status, $_.ifIndex)
  }
  Get-NetRoute -AddressFamily IPv4 -DestinationPrefix '0.0.0.0/0' -ErrorAction SilentlyContinue | ForEach-Object {
    L ("route0 via={0} if={1} metric={2}" -f $_.NextHop, $_.InterfaceAlias, $_.RouteMetric)
  }
  Get-DnsClientServerAddress -AddressFamily IPv4 -ErrorAction SilentlyContinue | Where-Object { $_.ServerAddresses } | ForEach-Object {
    L ("dns if={0} servers={1}" -f $_.InterfaceAlias, ($_.ServerAddresses -join ','))
  }

  foreach ($u in @('https://ya.ru','https://2ip.ru','https://ifconfig.me/ip','https://www.google.com','https://example.com','https://www.youtube.com')) {
    try {
      $sw = [Diagnostics.Stopwatch]::StartNew()
      $r = Invoke-WebRequest -Uri $u -UseBasicParsing -TimeoutSec 15
      $sw.Stop()
      $snip = ($r.Content.ToString() -replace '\s+',' ').Substring(0, [Math]::Min(60, $r.Content.Length))
      L ("SITE OK {0} code={1} ms={2} body={3}" -f $u, $r.StatusCode, $sw.ElapsedMilliseconds, $snip)
    } catch {
      L ("SITE FAIL {0} err={1}" -f $u, $_.Exception.Message)
    }
  }

  # drain more core logs briefly
  $drainUntil = (Get-Date).AddSeconds(3)
  while ((Get-Date) -lt $drainUntil) {
    while ($stream.DataAvailable) {
      $line = $reader.ReadLine()
      if ($null -eq $line) { break }
      $coreLog.Add($line)
      L ("CORE {0}" -f $line)
    }
    Start-Sleep -Milliseconds 50
  }
}

Send-Cmd @{ id = 2; cmd = 'stop'; token = $meta.token }
Start-Sleep -Seconds 1
try { $tcp.Close() } catch {}
Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
Remove-Item $portFile -Force -ErrorAction SilentlyContinue

$logLines | Set-Content $LogOut -Encoding UTF8
L ("wrote {0}" -f $LogOut)
