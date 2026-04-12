$ErrorActionPreference = 'Stop'

wpr -status

$repo = Join-Path $env:USERPROFILE 'dev\priv\pes'

Get-ChildItem $repo -Filter 'pes-trace*.etl' -ErrorAction SilentlyContinue |
    Sort-Object LastWriteTime -Descending |
    Select-Object FullName, Length, LastWriteTime |
    Format-Table -AutoSize
