#Requires -Version 5.1
# install.ps1 - Install promptc on Windows
# Usage: irm https://raw.githubusercontent.com/Geogboe/promptc/main/install.ps1 | iex
$ErrorActionPreference = 'Stop'

$Repo = if ($env:PROMPTC_REPO) { $env:PROMPTC_REPO } else { 'Geogboe/promptc' }
$Binary = 'promptc.exe'
$ApiBase = if ($env:PROMPTC_RELEASES_API_BASE) { $env:PROMPTC_RELEASES_API_BASE.TrimEnd('/') } else { 'https://api.github.com/repos' }
$ReleaseTag = $env:PROMPTC_RELEASE_TAG
$Headers = @{
    Accept = 'application/vnd.github+json'
    'User-Agent' = 'promptc-install'
}

# ── Detect arch ──────────────────────────────────────────────────────────────

$Arch = if ([System.Environment]::Is64BitOperatingSystem) { 'amd64' } else {
    throw "32-bit Windows is not supported."
}
$OS = 'windows'

function Get-ReleaseMetadata {
    if ($ReleaseTag) {
        $ReleaseUrl = "$ApiBase/$Repo/releases/tags/$ReleaseTag"
    } else {
        $ReleaseUrl = "$ApiBase/$Repo/releases/latest"
    }

    Write-Host "Fetching release metadata..."
    try {
        return Invoke-RestMethod -Uri $ReleaseUrl -Headers $Headers
    } catch {
        throw "Failed to fetch release metadata from ${ReleaseUrl}: $($_.Exception.Message)"
    }
}

function Select-ReleaseAsset {
    param(
        [object[]]$Assets,
        [string]$Pattern,
        [string]$Description
    )

    $Asset = $Assets | Where-Object { $_.name -match $Pattern } | Select-Object -First 1
    if (-not $Asset) {
        $Available = ($Assets | ForEach-Object { $_.name }) -join ', '
        if (-not $Available) {
            $Available = '<none>'
        }
        throw "Could not find a $Description asset matching '$Pattern'. Available assets: $Available"
    }

    return $Asset
}

function Get-SHA256 {
    param([string]$Path)

    if (Get-Command Get-FileHash -ErrorAction SilentlyContinue) {
        return (Get-FileHash -Path $Path -Algorithm SHA256).Hash.ToLower()
    }

    $Stream = [System.IO.File]::OpenRead($Path)
    try {
        $Hasher = [System.Security.Cryptography.SHA256]::Create()
        try {
            $HashBytes = $Hasher.ComputeHash($Stream)
        } finally {
            $Hasher.Dispose()
        }
    } finally {
        $Stream.Dispose()
    }

    return ([BitConverter]::ToString($HashBytes)).Replace('-', '').ToLower()
}

$Release = Get-ReleaseMetadata
$Tag = $Release.tag_name
if (-not $Tag) { throw "Could not determine release tag." }
Write-Host "Release: $Tag"

if (-not $Release.assets -or @($Release.assets).Count -eq 0) {
    throw "Release '$Tag' has no published assets."
}

$Assets = @($Release.assets)
$ArchivePattern = "^promptc_.+_${OS}_${Arch}\.zip$"
$ArchiveAsset = Select-ReleaseAsset -Assets $Assets -Pattern $ArchivePattern -Description 'Windows archive'
$ChecksumsAsset = Select-ReleaseAsset -Assets $Assets -Pattern '^checksums\.txt$' -Description 'checksum file'

# ── Work in a temp dir ────────────────────────────────────────────────────────

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
    # ── Download archive + checksums ─────────────────────────────────────────

    Write-Host "Downloading $($ArchiveAsset.name)..."
    $ArchivePath   = Join-Path $Tmp $ArchiveAsset.name
    $ChecksumsPath = Join-Path $Tmp 'checksums.txt'

    Invoke-WebRequest -Uri $ArchiveAsset.browser_download_url -OutFile $ArchivePath -UseBasicParsing
    Invoke-WebRequest -Uri $ChecksumsAsset.browser_download_url -OutFile $ChecksumsPath -UseBasicParsing

    # ── Verify SHA256 ─────────────────────────────────────────────────────────

    Write-Host "Verifying checksum..."
    $ChecksumPattern = "^(?<hash>[A-Fa-f0-9]{64})\s+$([regex]::Escape($ArchiveAsset.name))$"
    $ChecksumLine = Get-Content $ChecksumsPath | Where-Object { $_ -match $ChecksumPattern } | Select-Object -First 1
    if (-not $ChecksumLine) { throw "Checksum entry for '$($ArchiveAsset.name)' not found in checksums.txt." }
    $Expected = ([regex]::Match($ChecksumLine, '^(?<hash>[A-Fa-f0-9]{64})\s+')).Groups['hash'].Value

    $Actual = Get-SHA256 -Path $ArchivePath
    if ($Actual -ne $Expected.ToLower()) {
        throw "Checksum mismatch!`n  Expected: $Expected`n  Actual:   $Actual"
    }
    Write-Host "Checksum OK."

    # ── Extract binary ────────────────────────────────────────────────────────

    Expand-Archive -Path $ArchivePath -DestinationPath $Tmp -Force

    $ExtractedBinary = Join-Path $Tmp 'promptc.exe'
    if (-not (Test-Path $ExtractedBinary)) {
        throw "Binary 'promptc.exe' not found in archive."
    }

    # ── Determine install directory ───────────────────────────────────────────

    if ($env:PROMPTC_INSTALL_DIR) {
        $InstallDir = $env:PROMPTC_INSTALL_DIR
    } else {
        $InstallDir = Join-Path $env:USERPROFILE '.local\bin'
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    # ── Install ───────────────────────────────────────────────────────────────

    $Destination = Join-Path $InstallDir $Binary
    Copy-Item -Path $ExtractedBinary -Destination $Destination -Force
    Write-Host "Installed promptc to $Destination"

    # ── Add to user PATH if not already present ───────────────────────────────

    $UserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $SkipPathUpdate = $env:PROMPTC_SKIP_PATH_UPDATE
    if (-not $SkipPathUpdate -and $UserPath -notlike "*$InstallDir*") {
        $NewPath = "$UserPath;$InstallDir"
        [Environment]::SetEnvironmentVariable('Path', $NewPath, 'User')
        Write-Host "Added $InstallDir to your user PATH."
        Write-Host "Restart your terminal for the PATH change to take effect."
        # Also update current session so --version works immediately
        $env:PATH = "$env:PATH;$InstallDir"
    } elseif ($SkipPathUpdate) {
        Write-Host "Skipping PATH update because PROMPTC_SKIP_PATH_UPDATE is set."
    }

    # ── Confirm ───────────────────────────────────────────────────────────────

    & $Destination --version

} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}
