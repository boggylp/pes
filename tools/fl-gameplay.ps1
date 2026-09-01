[CmdletBinding()]
param(
    [Parameter(Position = 0)]
    [ValidateSet('status', 'switch-dt13-vanilla', 'switch-dt18-vanilla')]
    [string]$Action = 'status',

    [string]$GameRoot = $(if ($env:PES_GAME_ROOT) { $env:PES_GAME_ROOT } else { Join-Path ${env:ProgramFiles(x86)} 'SP Football Life 2026' }),

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

$ConstantBins = @(
    'constant_match.bin'
    'constant_player.bin'
    'constant_positionCK.bin'
    'constant_positionPK.bin'
    'constant_shootAging.bin'
    'constant_stadium.bin'
    'constant_team.bin'
    'constant_tutorial.bin'
    'constant_tutorialConsole.bin'
)

$Dt18Releases = @{
    '3CFF08EE' = 'vanilla'
    'AE2D1537' = 'F4L-5.1'
    'B001E8DB' = 'NES2027-preview'
    '5CC62416' = 'Ivar-07-17'
    '5D714979' = 'FL27-beta'
    '8868077C' = 'TFSE-fouls'
    '77B48794' = 'Liberty-7.0b'
}

$BinReleases = @{
    'constant_team.bin/94CE2EC7' = 'SSE-5.2'
    'constant_player.bin/E0D81F53' = 'SSE-5.2'
}

function Get-ActiveCpkRoots {
    param([string]$SiderPath)

    $siderDir = Split-Path $SiderPath -Parent
    Get-Content $SiderPath | Where-Object { $_ -match '^\s*cpk\.root\s*=' } | ForEach-Object {
        $value = ($_ -split '=', 2)[1].Trim().Trim('"')
        $path = if ([System.IO.Path]::IsPathRooted($value)) { $value } else { Join-Path $siderDir $value }
        [pscustomobject]@{ Value = $value; Path = $path }
    }
}

function Get-EffectiveBins {
    param(
        [string]$SiderPath,
        [string]$Dt18Hash8
    )

    $roots = @(Get-ActiveCpkRoots $SiderPath)

    foreach ($bin in $ConstantBins) {
        $winner = $null
        foreach ($root in $roots) {
            $candidate = Join-Path $root.Path "common\match\constant\$bin"
            if (Test-Path $candidate) {
                $winner = [pscustomobject]@{
                    Bin = $bin
                    Source = $root.Value
                    Hash = (Get-Sha256 $candidate).Substring(0, 8)
                }
                break
            }
        }
        if (-not $winner) {
            $winner = [pscustomobject]@{ Bin = $bin; Source = 'dt18_all.cpk'; Hash = $Dt18Hash8 }
        }
        $winner
    }
}

function Get-EffectiveSummary {
    param(
        [object[]]$Bins,
        [string]$Dt18Hash8
    )

    $sources = @($Bins | Select-Object -ExpandProperty Source -Unique)
    $dt18Name = $Dt18Releases[$Dt18Hash8]
    if (-not $dt18Name) { $dt18Name = 'unknown' }
    $dt18Desc = "dt18_all.cpk $Dt18Hash8 ($dt18Name)"

    if ($sources.Count -gt 1) {
        return "mixed sources ($($sources -join ' + ')), see the bin table"
    }

    if ($sources[0] -eq 'dt18_all.cpk') {
        return "all 9 bins from $dt18Desc"
    }

    $labels = @($Bins | ForEach-Object { $BinReleases["$($_.Bin)/$($_.Hash)"] } | Where-Object { $_ } | Select-Object -Unique)
    $release = 'unrecognized bins'
    if ($labels.Count -eq 1) {
        $expected = @($BinReleases.GetEnumerator() | Where-Object { $_.Value -eq $labels[0] }).Count
        $matched = @($Bins | Where-Object { $BinReleases["$($_.Bin)/$($_.Hash)"] -eq $labels[0] }).Count
        if ($matched -eq $expected) {
            $release = $labels[0]
        }
    }

    "all 9 bins from $($sources[0]) ($release), $dt18Desc masked"
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

    $dt18Hash = Get-Sha256 $dt18Path

    @(
        [pscustomobject]@{ Component = 'dt13'; Hash = (Get-Sha256 $dt13Path); Size = (Get-Item $dt13Path).Length; Path = $dt13Path }
        [pscustomobject]@{ Component = 'dt18'; Hash = $dt18Hash; Size = (Get-Item $dt18Path).Length; Path = $dt18Path }
        [pscustomobject]@{ Component = 'exe'; Hash = (Get-Sha256 $exePath); Size = (Get-Item $exePath).Length; Path = $exePath }
    ) | Format-Table -AutoSize

    Write-Host ''
    Write-Host 'Effective constant bins (first cpk.root wins, then dt18_all.cpk):'
    $bins = @(Get-EffectiveBins -SiderPath $siderPath -Dt18Hash8 $dt18Hash.Substring(0, 8))
    $bins | Format-Table -AutoSize
    Write-Host ('Effective gameplay: ' + (Get-EffectiveSummary -Bins $bins -Dt18Hash8 $dt18Hash.Substring(0, 8)))

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
