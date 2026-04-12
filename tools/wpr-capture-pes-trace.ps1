$ErrorActionPreference = 'Stop'

$repo = Join-Path $env:USERPROFILE 'dev\priv\pes'
Set-Location $repo

$out = Join-Path $repo 'pes-trace.etl'
$log = Join-Path $repo 'pes-trace.log'

Remove-Item $out -Force -ErrorAction SilentlyContinue
Remove-Item $log -Force -ErrorAction SilentlyContinue

Start-Transcript -Path $log -Force | Out-Null

Write-Host "Starting WPR capture"
wpr -cancel *> $null
wpr -status
wpr -start GPU -start Video -start DesktopComposition -filemode
Start-Sleep -Seconds 45
Write-Host "Stopping WPR capture to $out"
wpr -stop $out
wpr -status

Get-Item $out | Select-Object FullName, Length, LastWriteTime | Format-List

Stop-Transcript | Out-Null
