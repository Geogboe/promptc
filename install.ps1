#Requires -Version 5.1
# install.ps1 - Install promptc on Windows
# Usage: irm https://raw.githubusercontent.com/Geogboe/promptc/main/install.ps1 | iex
$ErrorActionPreference = 'Stop'

$Repo   = 'Geogboe/promptc'
$Binary = 'promptc.exe'

# ── Detect arch ──────────────────────────────────────────────────────────────

$Arch = if ([System.Environment]::Is64BitOperatingSystem) { 'amd64' } else {
    throw "32-bit Windows is not supported."
}
$OS = 'windows'

# ── Fetch latest release tag ─────────────────────────────────────────────────

Write-Host "Fetching latest release..."
$Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$Tag = $Release.tag_name
if (-not $Tag) { throw "Could not determine latest release tag." }
Write-Host "Latest release: $Tag"

# ── Construct URLs ────────────────────────────────────────────────────────────

$Archive      = "promptc_${Tag}_${OS}_${Arch}.zip"
$BaseUrl      = "https://github.com/$Repo/releases/download/$Tag"
$ArchiveUrl   = "$BaseUrl/$Archive"
$ChecksumsUrl = "$BaseUrl/checksums.txt"

# ── Work in a temp dir ────────────────────────────────────────────────────────

$Tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
    # ── Download archive + checksums ─────────────────────────────────────────

    Write-Host "Downloading $Archive..."
    $ArchivePath   = Join-Path $Tmp $Archive
    $ChecksumsPath = Join-Path $Tmp 'checksums.txt'

    Invoke-WebRequest -Uri $ArchiveUrl   -OutFile $ArchivePath   -UseBasicParsing
    Invoke-WebRequest -Uri $ChecksumsUrl -OutFile $ChecksumsPath -UseBasicParsing

    # ── Verify SHA256 ─────────────────────────────────────────────────────────

    Write-Host "Verifying checksum..."
    $Expected = (Get-Content $ChecksumsPath | Where-Object { $_ -match [regex]::Escape($Archive) }) -split '\s+' | Select-Object -First 1
    if (-not $Expected) { throw "Checksum entry for '$Archive' not found in checksums.txt." }

    $Actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
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
    if ($UserPath -notlike "*$InstallDir*") {
        $NewPath = "$UserPath;$InstallDir"
        [Environment]::SetEnvironmentVariable('Path', $NewPath, 'User')
        Write-Host "Added $InstallDir to your user PATH."
        Write-Host "Restart your terminal for the PATH change to take effect."
        # Also update current session so --version works immediately
        $env:PATH = "$env:PATH;$InstallDir"
    }

    # ── Confirm ───────────────────────────────────────────────────────────────

    & $Destination --version

} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}
