$ErrorActionPreference = 'Stop'

$repo = Join-Path $env:USERPROFILE 'dev\priv\pes'
Set-Location $repo

$out = Join-Path $repo 'presentmon-current.csv'
$log = Join-Path $repo 'presentmon-current.log'

Remove-Item $out -Force -ErrorAction SilentlyContinue
Remove-Item $log -Force -ErrorAction SilentlyContinue

Start-Transcript -Path $log -Force | Out-Null

presentmon `
  --process_name FL_2026.exe `
  --process_name brave.exe `
  --process_name dwm.exe `
  --track_gpu_video `
  --timed 30 `
  --terminate_after_timed `
  --stop_existing_session `
  --output_file $out `
  --v2_metrics `
  --date_time `
  --no_console_stats

Get-Item $out | Select-Object FullName, Length, LastWriteTime | Format-List

Stop-Transcript | Out-Null
