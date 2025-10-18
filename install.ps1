# Installation script for Kick Assembler Language Server (Windows)
# Usage: iwr -useb https://raw.githubusercontent.com/cybersorcerer/kickass_ls/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$REPO = "cybersorcerer/kickass_ls"
$BINARY_NAME = "kickass_ls.exe"
$INSTALL_DIR = "$env:LOCALAPPDATA\kickass_ls\bin"
$CONFIG_DIR = "$env:LOCALAPPDATA\kickass_ls\config"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Success { Write-ColorOutput Green $args }
function Write-Info { Write-ColorOutput Cyan $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Error { Write-ColorOutput Red $args }

# Get latest release version from GitHub
function Get-LatestVersion {
    Write-Info "Fetching latest release..."

    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$REPO/releases/latest"
        $version = $release.tag_name
        Write-Success "Latest version: $version"
        return $version
    }
    catch {
        Write-Warning "Could not fetch latest release, using 'latest'"
        return "latest"
    }
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE

    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default {
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Download and extract release
function Download-Release {
    param($Version, $Arch)

    $platform = "windows-$Arch"
    $downloadUrl = "https://github.com/$REPO/releases/download/$Version/kickass_ls-$Version-$platform.zip"
    $tempDir = New-TemporaryFile | ForEach-Object { Remove-Item $_; New-Item -ItemType Directory -Path $_ }
    $zipFile = Join-Path $tempDir "kickass_ls.zip"

    Write-Info "Downloading $BINARY_NAME for $platform..."
    Write-Info "URL: $downloadUrl"

    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile -UseBasicParsing
    }
    catch {
        Write-Error "Failed to download release"
        Write-Warning "Please check if a release exists for $platform"
        Remove-Item -Recurse -Force $tempDir
        exit 1
    }

    Write-Info "Extracting archive..."
    Expand-Archive -Path $zipFile -DestinationPath $tempDir -Force

    return Join-Path $tempDir "kickass_ls-$Version-$platform"
}

# Install binary and configuration files
function Install-Files {
    param($ExtractDir)

    Write-Info "Installing files..."

    # Create directories
    New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $CONFIG_DIR | Out-Null

    # Install binary
    $binaryPath = Join-Path $ExtractDir $BINARY_NAME
    if (Test-Path $binaryPath) {
        Copy-Item $binaryPath -Destination $INSTALL_DIR -Force
        Write-Success "✓ Binary installed to $INSTALL_DIR\$BINARY_NAME"
    }
    else {
        Write-Error "Binary not found in archive"
        exit 1
    }

    # Install configuration files
    $configFiles = @("kickass.json", "mnemonic.json", "c64memory.json")
    foreach ($configFile in $configFiles) {
        $configPath = Join-Path $ExtractDir $configFile
        if (Test-Path $configPath) {
            Copy-Item $configPath -Destination $CONFIG_DIR -Force
            Write-Success "✓ Installed $configFile to $CONFIG_DIR"
        }
        else {
            Write-Warning "⚠ Warning: $configFile not found in archive"
        }
    }
}

# Check and update PATH
function Update-Path {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -notlike "*$INSTALL_DIR*") {
        Write-Warning "⚠ $INSTALL_DIR is not in your PATH"
        Write-Info "Adding to PATH..."

        $newPath = "$INSTALL_DIR;$currentPath"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

        # Update current session
        $env:Path = "$INSTALL_DIR;$env:Path"

        Write-Success "✓ Added to PATH (restart your terminal for system-wide effect)"
    }
    else {
        Write-Success "✓ $INSTALL_DIR is already in your PATH"
    }
}

# Verify installation
function Test-Installation {
    Write-Info ""
    Write-Info "Verifying installation..."

    $binaryPath = Join-Path $INSTALL_DIR $BINARY_NAME
    if (Test-Path $binaryPath) {
        try {
            $versionOutput = & $binaryPath --version 2>&1
        }
        catch {
            $versionOutput = "unknown"
        }

        Write-Success "✓ Installation successful!"
        Write-Success "  Version: $versionOutput"
        Write-Success "  Binary: $binaryPath"
        Write-Success "  Config: $CONFIG_DIR"
    }
    else {
        Write-Error "Installation verification failed"
        exit 1
    }
}

# Main installation flow
function Main {
    Write-Host ""
    Write-ColorOutput Cyan "========================================"
    Write-ColorOutput Cyan " Kick Assembler Language Server Setup "
    Write-ColorOutput Cyan "========================================"
    Write-Host ""

    $arch = Get-Architecture
    Write-Info "Platform detected: windows-$arch"
    Write-Host ""

    $version = Get-LatestVersion
    $extractDir = Download-Release -Version $version -Arch $arch

    Install-Files -ExtractDir $extractDir
    Update-Path
    Test-Installation

    # Cleanup
    Remove-Item -Recurse -Force (Split-Path $extractDir -Parent)

    Write-Host ""
    Write-ColorOutput Green "Installation complete!"
    Write-Host ""
    Write-Info "Next steps:"
    Write-Info "  1. Restart your terminal to update PATH"
    Write-Info "  2. Configure your editor/LSP client to use: $INSTALL_DIR\$BINARY_NAME"
    Write-Info "  3. For setup instructions, see: https://github.com/$REPO#editor-configuration"
    Write-Host ""
}

# Run main function
Main
