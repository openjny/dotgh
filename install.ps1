# dotgh installer script for Windows (PowerShell)
# Usage: irm https://raw.githubusercontent.com/openjny/dotgh/main/install.ps1 | iex
#
# Environment variables:
#   DOTGH_INSTALL_DIR - Custom installation directory (default: $env:LOCALAPPDATA\dotgh)
#   DOTGH_VERSION     - Specific version to install (default: latest)

$ErrorActionPreference = "Stop"

# Configuration
$script:Repo = "openjny/dotgh"
$script:BinaryName = "dotgh"
$script:GitHubApi = "https://api.github.com"
$script:GitHubReleases = "https://github.com/$script:Repo/releases"

# ============================================================================
# Logging functions with colors
# ============================================================================

function Write-Info {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Blue -NoNewline
    Write-Host $Message -ForegroundColor White
}

function Write-Success {
    param([string]$Message)
    Write-Host "==> " -ForegroundColor Green -NoNewline
    Write-Host $Message -ForegroundColor White
}

function Write-Warn {
    param([string]$Message)
    Write-Host "Warning: " -ForegroundColor Yellow -NoNewline
    Write-Host $Message
}

function Write-Err {
    param([string]$Message)
    Write-Host "Error: " -ForegroundColor Red -NoNewline
    Write-Host $Message
    exit 1
}

# ============================================================================
# Helper functions
# ============================================================================

function Get-SystemArch {
    <#
    .SYNOPSIS
    Detect system architecture and return GoReleaser-compatible name
    #>
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { Write-Err "Unsupported architecture: $arch" }
    }
}

function Get-LatestVersion {
    <#
    .SYNOPSIS
    Get the latest release version from GitHub API
    #>
    $url = "$script:GitHubApi/repos/$script:Repo/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $url -UseBasicParsing
        return $response.tag_name
    }
    catch {
        Write-Err "Failed to get latest version from GitHub: $_"
    }
}

function Get-InstallDir {
    <#
    .SYNOPSIS
    Determine installation directory
    #>
    if ($env:DOTGH_INSTALL_DIR) {
        return $env:DOTGH_INSTALL_DIR
    }
    return "$env:LOCALAPPDATA\$script:BinaryName"
}

function Test-Checksum {
    <#
    .SYNOPSIS
    Verify SHA256 checksum of a file
    #>
    param(
        [string]$FilePath,
        [string]$ExpectedHash
    )
    
    $actualHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()
    $expectedLower = $ExpectedHash.ToLower()
    
    if ($actualHash -ne $expectedLower) {
        Write-Err "Checksum verification failed!`nExpected: $expectedLower`nActual: $actualHash"
    }
    Write-Success "Checksum verified"
}

function Add-ToUserPath {
    <#
    .SYNOPSIS
    Add directory to user PATH if not already present
    #>
    param([string]$Directory)
    
    $currentPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
    
    if ($currentPath -notlike "*$Directory*") {
        Write-Info "Adding $Directory to user PATH..."
        $newPath = "$Directory;$currentPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::User)
        # Also update current session
        $env:Path = "$Directory;$env:Path"
        Write-Success "Added to PATH. Restart your terminal for changes to take effect."
    }
}

# ============================================================================
# Main installation function
# ============================================================================

function Install-Dotgh {
    Write-Host ""
    Write-Info "Installing $script:BinaryName..."
    Write-Host ""

    # Detect platform
    $os = "windows"
    $arch = Get-SystemArch
    Write-Info "Detected platform: $os/$arch"

    # Get version
    $version = $env:DOTGH_VERSION
    if (-not $version) {
        Write-Info "Fetching latest version..."
        $version = Get-LatestVersion
    }
    Write-Info "Version: $version"

    # Determine filename and URLs
    $filename = "${script:BinaryName}_${os}_${arch}.zip"
    $downloadUrl = "$script:GitHubReleases/download/$version/$filename"
    $checksumsUrl = "$script:GitHubReleases/download/$version/checksums.txt"

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "dotgh-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download archive
        $archivePath = Join-Path $tempDir $filename
        Write-Info "Downloading from $downloadUrl"
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing

        # Download checksums
        $checksumsPath = Join-Path $tempDir "checksums.txt"
        Write-Info "Downloading checksums..."
        Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing

        # Verify checksum
        $checksumContent = Get-Content $checksumsPath
        $expectedChecksum = ($checksumContent | Where-Object { $_ -match $filename } | ForEach-Object { ($_ -split '\s+')[0] })
        if (-not $expectedChecksum) {
            Write-Err "Checksum not found for $filename"
        }
        Test-Checksum -FilePath $archivePath -ExpectedHash $expectedChecksum

        # Extract archive
        Write-Info "Extracting archive..."
        $extractPath = Join-Path $tempDir "extracted"
        Expand-Archive -Path $archivePath -DestinationPath $extractPath -Force

        # Get installation directory
        $installDir = Get-InstallDir
        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }

        # Install binary
        $binarySrc = Join-Path $extractPath "$script:BinaryName.exe"
        $binaryDest = Join-Path $installDir "$script:BinaryName.exe"

        if (-not (Test-Path $binarySrc)) {
            Write-Err "Binary not found in archive"
        }

        Write-Info "Installing to $binaryDest"
        Copy-Item -Path $binarySrc -Destination $binaryDest -Force

        # Add to PATH
        Add-ToUserPath -Directory $installDir

        # Verify installation
        Write-Host ""
        Write-Success "$script:BinaryName $version installed successfully!"
        Write-Host ""
        
        # Show version
        try {
            & $binaryDest version
        }
        catch {
            # Ignore errors from version command
            $null = $_
        }
        Write-Host ""
    }
    finally {
        # Cleanup temp directory
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run installation
Install-Dotgh
