[CmdletBinding()]
param(
    [Parameter(Mandatory, Position = 0)]
    [string]$Path,

    [Parameter(Position = 1)]
    [string]$OutPath
)

$ErrorActionPreference = 'Stop'

$FtexToolVersion = 'v0.3.3'
$FtexToolUrl = "https://github.com/Atvaark/FtexTool/releases/download/$FtexToolVersion/FtexTool.$FtexToolVersion.zip"
$FtexToolDir = Join-Path $PSScriptRoot "bin\FtexTool-$FtexToolVersion"
$FtexToolExe = Join-Path $FtexToolDir 'FtexTool.exe'

function Get-FtexTool {
    if (Test-Path -LiteralPath $FtexToolExe) { return $FtexToolExe }

    Write-Verbose "Fetching FtexTool $FtexToolVersion from GitHub"
    $zip = Join-Path ([IO.Path]::GetTempPath()) 'FtexTool.zip'
    Invoke-WebRequest -Uri $FtexToolUrl -OutFile $zip -UseBasicParsing

    $tmpExtract = Join-Path ([IO.Path]::GetTempPath()) ('FtexTool-extract-' + [Guid]::NewGuid())
    Expand-Archive -Path $zip -DestinationPath $tmpExtract -Force

    New-Item -ItemType Directory -Path $FtexToolDir -Force | Out-Null
    Get-ChildItem -Path $tmpExtract -Recurse -File | ForEach-Object {
        Copy-Item -LiteralPath $_.FullName -Destination $FtexToolDir -Force
    }

    Remove-Item -LiteralPath $zip -Force -ErrorAction SilentlyContinue
    Remove-Item -Recurse -Force -LiteralPath $tmpExtract -ErrorAction SilentlyContinue

    if (-not (Test-Path -LiteralPath $FtexToolExe)) {
        throw "FtexTool download failed: $FtexToolExe missing after extraction"
    }
    return $FtexToolExe
}

if (-not (Get-Command magick -ErrorAction SilentlyContinue)) {
    throw "magick not in PATH. Install ImageMagick."
}

$ftexTool = Get-FtexTool

$src = Get-Item -LiteralPath $Path

if (-not $OutPath) {
    $OutPath = Join-Path $src.Directory.FullName ($src.BaseName + '.ftex')
}

$outDir = [IO.Path]::GetDirectoryName([IO.Path]::GetFullPath($OutPath))
if (-not (Test-Path -LiteralPath $outDir)) {
    New-Item -ItemType Directory -Path $outDir -Force | Out-Null
}

$tmp = Join-Path ([IO.Path]::GetTempPath()) ('kit-to-ftex-' + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
$dds = Join-Path $tmp ($src.BaseName + '.dds')

try {
    & magick $src.FullName -define dds:compression=dxt5 $dds
    if ($LASTEXITCODE) { throw "magick failed ($LASTEXITCODE)" }

    & $ftexTool -i $dds -o $outDir
    if ($LASTEXITCODE) { throw "FtexTool failed ($LASTEXITCODE)" }

    $produced = Join-Path $outDir ($src.BaseName + '.ftex')
    if ($produced -ne $OutPath -and (Test-Path -LiteralPath $produced)) {
        Move-Item -Force -LiteralPath $produced -Destination $OutPath
    }
    Write-Output $OutPath
}
finally {
    Remove-Item -Recurse -Force -LiteralPath $tmp -ErrorAction SilentlyContinue
}
