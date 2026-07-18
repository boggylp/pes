[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('status', 'switch-dt13-vanilla', 'switch-dt18-vanilla')]
    [string]$Action = 'status',

    [string]$GameRoot = (Join-Path ${env:ProgramFiles(x86)} 'SP Football Life 2026'),

    [string]$GameplayRoot = (Join-Path $env:USERPROFILE 'MEGA\gaming\pes\gameplay'),

    [string]$VanillaSubdir = 'fl26-v2.2'
)

$ErrorActionPreference = 'Stop'

function Require-Path {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        throw "Path not found: $Path"
    }
}

function Get-Sha256 {
    param([string]$Path)

    (Get-FileHash $Path -Algorithm SHA256).Hash
}

function Get-SystemFile {
    Join-Path $env:USERPROFILE 'Documents\KONAMI\eFootball PES 2021 SEASON UPDATE\2026\save\SYSTEM00000000'
}

function Get-TrackingLines {
    param([string]$SiderPath)

    Get-Content $SiderPath | Where-Object { $_ -match '^\s*; gameplay:' }
}

function Get-GameplaySiderLines {
    param([string]$SiderPath)

    $pattern = '(?i)(gameplay|sse|ai_tweaks|anti-cheat|speedserver2|attacking_mentality|attack_mentality|stamina|noscript|matchset|crow|anima)'

    Get-Content $SiderPath | Where-Object {
        $_ -match '^\s*(cpk\.root|lua\.module)\s*=' -and $_ -match $pattern
    }
}

function Update-TrackingComment {
    param(
        [string]$SiderPath,
        [ValidateSet('dt13', 'dt18')]
        [string]$Component,
        [string]$Value
    )

    $stamp = '| applied {0} {1}' -f (Get-Date -Format 'yyyy-MM-dd'), [System.Net.Dns]::GetHostName()
    $lines = [System.Collections.Generic.List[string]]::new()
    $lines.AddRange([string[]](Get-Content $SiderPath))

    $updated = $false
    for ($i = 0; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -notmatch '^\s*; gameplay:') {
            continue
        }

        $line = $lines[$i] -replace '\s*\| applied .*$', ''
        if ($line -match "\b$Component=\S+") {
            $line = $line -replace "\b$Component=\S+", "$Component=$Value"
        }
        else {
            $line = "$line $Component=$Value"
        }
        $lines[$i] = "$line $stamp"
        $updated = $true
    }

    if (-not $updated) {
        $replacement = "; gameplay: $Component=$Value $stamp"
        $insertAt = -1
        for ($i = 0; $i -lt $lines.Count; $i++) {
            if ($lines[$i] -match '^\s*overlay\.enabled') {
                $insertAt = $i
                break
            }
        }

        if ($insertAt -lt 0) {
            $lines.Add($replacement)
        }
        else {
            $lines.Insert($insertAt, $replacement)
        }
    }

    Set-Content $SiderPath -Value $lines -Encoding ascii
}

function Resolve-VanillaSource {
    param([string]$FileName)

    $vanillaRoot = Join-Path $GameplayRoot 'vanilla'
    Require-Path $vanillaRoot

    $candidates = @(
        (Join-Path $vanillaRoot $FileName),
        (Join-Path (Join-Path $vanillaRoot $VanillaSubdir) $FileName),
        (Join-Path (Join-Path $vanillaRoot 'dt13 & dt18 vanilla') $FileName)
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return @{
                Path = $candidate
                TempDir = $null
            }
        }
    }

    $archive = Join-Path $vanillaRoot 'dt13 & dt18 vanilla.rar'
    Require-Path $archive

    $tempDir = Join-Path $env:TEMP ("pes-vanilla-" + [guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tempDir | Out-Null

    tar -xf $archive -C $tempDir

    $extracted = Join-Path $tempDir (Join-Path 'dt13 & dt18 vanilla' $FileName)
    Require-Path $extracted

    return @{
        Path = $extracted
        TempDir = $tempDir
    }
}

function Show-Status {
    $siderPath = Join-Path $GameRoot 'SiderAddons\sider.ini'
    $dt13Path = Join-Path $GameRoot 'Data\dt13_all.cpk'
    $dt18Path = Join-Path $GameRoot 'Data\dt18_all.cpk'
    $exePath = Join-Path $GameRoot 'FL_2026.exe'
    $systemFile = Get-SystemFile

    Require-Path $GameRoot
    Require-Path $siderPath
    Require-Path $dt13Path
    Require-Path $dt18Path
    Require-Path $exePath

    Write-Host "Game root: $GameRoot"
    Write-Host "Gameplay root: $GameplayRoot"
    Write-Host ''

    @(
        [pscustomobject]@{ Component = 'dt13'; Hash = (Get-Sha256 $dt13Path); Size = (Get-Item $dt13Path).Length; Path = $dt13Path }
        [pscustomobject]@{ Component = 'dt18'; Hash = (Get-Sha256 $dt18Path); Size = (Get-Item $dt18Path).Length; Path = $dt18Path }
        [pscustomobject]@{ Component = 'exe'; Hash = (Get-Sha256 $exePath); Size = (Get-Item $exePath).Length; Path = $exePath }
    ) | Format-Table -AutoSize

    Write-Host ''
    Write-Host 'Tracking comments:'
    Get-TrackingLines $siderPath | ForEach-Object { Write-Host $_ }

    Write-Host ''
    Write-Host 'Active gameplay-related sider entries:'
    $entries = @(Get-GameplaySiderLines $siderPath)
    if ($entries.Count -eq 0) {
        Write-Host '(none matched the gameplay filter)'
    }
    else {
        $entries | ForEach-Object { Write-Host $_ }
    }

    Write-Host ''
    if (Test-Path $systemFile) {
        Write-Host "SYSTEM cache: present ($systemFile)"
    }
    else {
        Write-Host 'SYSTEM cache: missing'
    }
}

function Switch-VanillaComponent {
    param(
        [ValidateSet('dt13_all.cpk', 'dt18_all.cpk')]
        [string]$FileName
    )

    $component = if ($FileName -eq 'dt13_all.cpk') { 'dt13' } else { 'dt18' }
    $targetPath = Join-Path $GameRoot (Join-Path 'Data' $FileName)
    $backupDir = Join-Path $GameRoot '.backup\Data'
    $siderPath = Join-Path $GameRoot 'SiderAddons\sider.ini'
    $systemFile = Get-SystemFile

    Require-Path $GameRoot
    Require-Path $targetPath
    Require-Path $siderPath

    $source = Resolve-VanillaSource $FileName

    try {
        $beforeHash = Get-Sha256 $targetPath
        $sourceHash = Get-Sha256 $source.Path

        if ($beforeHash -eq $sourceHash) {
            Write-Host "$component already matches the vanilla source"
            return
        }

        New-Item -ItemType Directory -Path $backupDir -Force | Out-Null

        $timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
        $backupName = '{0}.{1}{2}' -f [System.IO.Path]::GetFileNameWithoutExtension($FileName), $timestamp, [System.IO.Path]::GetExtension($FileName)
        $backupPath = Join-Path $backupDir $backupName

        Copy-Item $targetPath $backupPath
        Copy-Item $source.Path $targetPath -Force
        Update-TrackingComment -SiderPath $siderPath -Component $component -Value ('vanilla({0})' -f $sourceHash.Substring(0, 8).ToLower())

        [pscustomobject]@{
            Component = $component
            Source = $source.Path
            Backup = $backupPath
            BeforeHash = $beforeHash
            AfterHash = (Get-Sha256 $targetPath)
            SystemCache = if (Test-Path $systemFile) { 'present' } else { 'missing' }
        } | Format-List
    }
    finally {
        if ($source.TempDir) {
            Remove-Item $source.TempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

switch ($Action) {
    'status' { Show-Status }
    'switch-dt13-vanilla' { Switch-VanillaComponent 'dt13_all.cpk' }
    'switch-dt18-vanilla' { Switch-VanillaComponent 'dt18_all.cpk' }
}
