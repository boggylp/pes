[CmdletBinding(DefaultParameterSetName = 'ExplicitOutput')]
param(
    [Parameter(Mandatory, Position = 0)]
    [string]$Path,

    [Parameter(Position = 1, ParameterSetName = 'ExplicitOutput')]
    [string]$OutPath,

    [Parameter(ParameterSetName = 'TeamNaming')]
    [int]$TeamId,

    [Parameter(ParameterSetName = 'TeamNaming')]
    [string]$TeamName,

    [Parameter(ParameterSetName = 'TeamNaming')]
    [ValidateRange(1, 9)]
    [int]$Slot = 1,

    [Parameter(ParameterSetName = 'TeamNaming')]
    [ValidateSet('p', 'g')]
    [string]$KitType = 'p',

    [Parameter(ParameterSetName = 'TeamNaming')]
    [string]$TeamsFile = (Join-Path ${env:ProgramFiles(x86)} 'SP Football Life 2026\FL26_teams.txt'),

    [Parameter(ParameterSetName = 'TeamNaming')]
    [string]$OutDir
)

$ErrorActionPreference = 'Stop'

$FtexToolVersion = 'v0.4.0'
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

function Resolve-TeamId {
    param([string]$TeamsFile, [string]$Name)

    if (-not (Test-Path -LiteralPath $TeamsFile)) {
        throw "Teams file not found: $TeamsFile"
    }

    $matches = Get-Content -LiteralPath $TeamsFile | Where-Object {
        $_ -match "^\s*(\d+)\s*-\s*(.+?)\s*$" -and $Matches[2] -like "*$Name*"
    } | ForEach-Object {
        if ($_ -match "^\s*(\d+)\s*-\s*(.+?)\s*$") {
            [PSCustomObject]@{ Id = [int]$Matches[1]; Name = $Matches[2] }
        }
    }

    if (-not $matches) {
        throw "No team matched '$Name' in $TeamsFile"
    }
    if (@($matches).Count -gt 1) {
        $list = ($matches | ForEach-Object { "$($_.Id) - $($_.Name)" }) -join "`n  "
        throw "Multiple teams matched '$Name':`n  $list`nRefine the name or pass -TeamId."
    }
    return $matches[0]
}

if (-not (Get-Command magick -ErrorAction SilentlyContinue)) {
    throw "magick not in PATH. Install ImageMagick."
}

$ftexTool = Get-FtexTool
$src = Get-Item -LiteralPath $Path

if ($PSCmdlet.ParameterSetName -eq 'TeamNaming') {
    if (-not $TeamId) {
        if (-not $TeamName) {
            throw "Provide -TeamId or -TeamName."
        }
        $team = Resolve-TeamId -TeamsFile $TeamsFile -Name $TeamName
        $TeamId = $team.Id
        Write-Verbose "Resolved team: $($team.Id) - $($team.Name)"
    }
    $baseName = "u${TeamId}${KitType}${Slot}"
    if (-not $OutDir) { $OutDir = $src.Directory.FullName }
    $OutPath = Join-Path $OutDir ($baseName + '.ftex')
}
elseif (-not $OutPath) {
    $OutPath = Join-Path $src.Directory.FullName ($src.BaseName + '.ftex')
}

$outFolder = [IO.Path]::GetDirectoryName([IO.Path]::GetFullPath($OutPath))
if (-not (Test-Path -LiteralPath $outFolder)) {
    New-Item -ItemType Directory -Path $outFolder -Force | Out-Null
}
$outBase = [IO.Path]::GetFileNameWithoutExtension($OutPath)

$tmp = Join-Path ([IO.Path]::GetTempPath()) ('kit-to-ftex-' + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
$dds = Join-Path $tmp ($outBase + '.dds')

try {
    & magick $src.FullName -define dds:compression=dxt5 -define dds:mipmaps=-1 $dds
    if ($LASTEXITCODE) { throw "magick failed ($LASTEXITCODE)" }

    & $ftexTool -f 0 -i $dds -o $tmp
    if ($LASTEXITCODE) { throw "FtexTool failed ($LASTEXITCODE)" }

    $produced = @()
    Get-ChildItem -LiteralPath $tmp -Filter "$outBase.*" -File | Where-Object {
        $_.Extension -eq '.ftex' -or $_.Name -match '\.\d+\.ftexs$'
    } | ForEach-Object {
        $dest = Join-Path $outFolder $_.Name
        Move-Item -Force -LiteralPath $_.FullName -Destination $dest
        $produced += $dest
    }

    $produced | ForEach-Object { Write-Output $_ }
}
finally {
    Remove-Item -Recurse -Force -LiteralPath $tmp -ErrorAction SilentlyContinue
}
