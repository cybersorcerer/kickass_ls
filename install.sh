#!/bin/bash
# Installation script for Kick Assembler Language Server
# Usage: curl -fsSL https://raw.githubusercontent.com/cybersorcerer/c64.nvim/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# Configuration
REPO="cybersorcerer/kickass_ls"
BINARY_NAME="kickass_ls"
INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/kickass_ls"

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            echo -e "${RED}Unsupported operating system: $OS${RESET}"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            echo -e "${RED}Unsupported architecture: $ARCH${RESET}"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
}

# Get latest release version from GitHub
get_latest_version() {
    echo -e "${CYAN}Fetching latest release...${RESET}"

    # Try to get latest release tag
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        echo -e "${YELLOW}Could not fetch latest release, using 'latest'${RESET}"
        VERSION="latest"
    else
        echo -e "${GREEN}Latest version: ${VERSION}${RESET}"
    fi
}

# Download and extract release
download_release() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/kickass_ls-${VERSION}-${PLATFORM}.tar.gz"
    TEMP_DIR=$(mktemp -d)

    echo -e "${CYAN}Downloading ${BINARY_NAME} for ${PLATFORM}...${RESET}"
    echo -e "${CYAN}URL: ${DOWNLOAD_URL}${RESET}"

    if ! curl -fsSL "$DOWNLOAD_URL" -o "${TEMP_DIR}/kickass_ls.tar.gz"; then
        echo -e "${RED}Failed to download release${RESET}"
        echo -e "${YELLOW}Please check if a release exists for ${PLATFORM}${RESET}"
        rm -rf "$TEMP_DIR"
        exit 1
    fi

    echo -e "${CYAN}Extracting archive...${RESET}"
    tar -xzf "${TEMP_DIR}/kickass_ls.tar.gz" -C "$TEMP_DIR"

    EXTRACT_DIR="${TEMP_DIR}/kickass_ls-${VERSION}-${PLATFORM}"
}

# Install binary and configuration files
install_files() {
    echo -e "${CYAN}Installing files...${RESET}"

    # Create directories
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"

    # Install binary
    if [ -f "${EXTRACT_DIR}/${BINARY_NAME}" ]; then
        cp "${EXTRACT_DIR}/${BINARY_NAME}" "$INSTALL_DIR/"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
        echo -e "${GREEN}✓ Binary installed to ${INSTALL_DIR}/${BINARY_NAME}${RESET}"
    else
        echo -e "${RED}Binary not found in archive${RESET}"
        exit 1
    fi

    # Install configuration files
    for config_file in kickass.json mnemonic.json c64memory.json; do
        if [ -f "${EXTRACT_DIR}/${config_file}" ]; then
            cp "${EXTRACT_DIR}/${config_file}" "$CONFIG_DIR/"
            echo -e "${GREEN}✓ Installed ${config_file} to ${CONFIG_DIR}/${RESET}"
        else
            echo -e "${YELLOW}⚠ Warning: ${config_file} not found in archive${RESET}"
        fi
    done

    # Cleanup
    rm -rf "$TEMP_DIR"
}

# Check if install directory is in PATH
check_path() {
    if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo ""
        echo -e "${YELLOW}⚠ Warning: ${INSTALL_DIR} is not in your PATH${RESET}"
        echo -e "${YELLOW}Add this line to your ~/.bashrc or ~/.zshrc:${RESET}"
        echo -e "${CYAN}  export PATH=\"\$HOME/.local/bin:\$PATH\"${RESET}"
        echo ""
        echo -e "${YELLOW}Then run: source ~/.bashrc  (or source ~/.zshrc)${RESET}"
    else
        echo -e "${GREEN}✓ ${INSTALL_DIR} is in your PATH${RESET}"
    fi
}

# Verify installation
verify_installation() {
    echo ""
    echo -e "${CYAN}Verifying installation...${RESET}"

    if [ -x "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        VERSION_OUTPUT=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>&1 || echo "unknown")
        echo -e "${GREEN}✓ Installation successful!${RESET}"
        echo -e "${GREEN}  Version: ${VERSION_OUTPUT}${RESET}"
        echo -e "${GREEN}  Binary: ${INSTALL_DIR}/${BINARY_NAME}${RESET}"
        echo -e "${GREEN}  Config: ${CONFIG_DIR}/${RESET}"
    else
        echo -e "${RED}Installation verification failed${RESET}"
        exit 1
    fi
}

# Main installation flow
main() {
    echo -e "${BOLD}${CYAN}========================================${RESET}"
    echo -e "${BOLD}${CYAN} Kick Assembler Language Server Setup ${RESET}"
    echo -e "${BOLD}${CYAN}========================================${RESET}"
    echo ""

    detect_platform
    echo -e "${CYAN}Platform detected: ${PLATFORM}${RESET}"
    echo ""

    get_latest_version
    download_release
    install_files
    verify_installation
    check_path

    echo ""
    echo -e "${BOLD}${GREEN}Installation complete!${RESET}"
    echo ""
    echo -e "${CYAN}Next steps:${RESET}"
    echo -e "  1. Make sure ${INSTALL_DIR} is in your PATH"
    echo -e "  2. Configure your editor/LSP client to use: ${INSTALL_DIR}/${BINARY_NAME}"
    echo -e "  3. For setup instructions, see: https://github.com/${REPO}#editor-configuration"
    echo ""
}

# Run main function
main
