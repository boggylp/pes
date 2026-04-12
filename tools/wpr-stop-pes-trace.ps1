$ErrorActionPreference = 'Stop'

$repo = Join-Path $env:USERPROFILE 'dev\priv\pes'
Set-Location $repo

$out = Join-Path $repo 'pes-trace.etl'
$log = Join-Path $repo 'pes-trace-stop.log'

Remove-Item $log -Force -ErrorAction SilentlyContinue
Start-Transcript -Path $log -Force | Out-Null

wpr -stop $out
wpr -status
Get-Item $out | Select-Object FullName, Length, LastWriteTime | Format-List

Stop-Transcript | Out-Null
