#!/bin/bash
# dotgh installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/openjny/dotgh/main/install.sh | bash
#
# Environment variables:
#   DOTGH_INSTALL_DIR - Custom installation directory (default: ~/.local/bin or /usr/local/bin)
#   DOTGH_VERSION     - Specific version to install (default: latest)

set -euo pipefail

# Configuration
REPO="openjny/dotgh"
BINARY_NAME="dotgh"
GITHUB_API="https://api.github.com"
GITHUB_RELEASES="https://github.com/${REPO}/releases"

# Colors for output
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    BOLD=''
    NC=''
fi

# Logging functions
info() {
    printf "${BLUE}==>${NC} ${BOLD}%s${NC}\n" "$*"
}

success() {
    printf "${GREEN}==>${NC} ${BOLD}%s${NC}\n" "$*"
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$*" >&2
}

error() {
    printf "${RED}Error:${NC} %s\n" "$*" >&2
    exit 1
}

# Detect OS
detect_os() {
    local os
    os="$(uname -s)"
    case "${os}" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)       error "Unsupported operating system: ${os}" ;;
    esac
}

# Detect architecture
detect_arch() {
    local arch
    arch="$(uname -m)"
    case "${arch}" in
        x86_64|amd64)  echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *)             error "Unsupported architecture: ${arch}" ;;
    esac
}

# Check if a command exists
has_command() {
    command -v "$1" >/dev/null 2>&1
}

# Get the latest version from GitHub
get_latest_version() {
    local version
    if has_command curl; then
        version=$(curl -fsSL "${GITHUB_API}/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    elif has_command wget; then
        version=$(wget -qO- "${GITHUB_API}/repos/${REPO}/releases/latest" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    if [[ -z "${version}" ]]; then
        error "Failed to get latest version from GitHub"
    fi
    echo "${version}"
}

# Download file
download() {
    local url="$1"
    local dest="$2"

    info "Downloading from ${url}"
    if has_command curl; then
        curl -fsSL -o "${dest}" "${url}"
    elif has_command wget; then
        wget -qO "${dest}" "${url}"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi
}

# Verify checksum
verify_checksum() {
    local file="$1"
    local expected_checksum="$2"
    local actual_checksum

    if has_command sha256sum; then
        actual_checksum=$(sha256sum "${file}" | awk '{print $1}')
    elif has_command shasum; then
        actual_checksum=$(shasum -a 256 "${file}" | awk '{print $1}')
    else
        warn "Neither sha256sum nor shasum found. Skipping checksum verification."
        return 0
    fi

    if [[ "${actual_checksum}" != "${expected_checksum}" ]]; then
        error "Checksum verification failed!\nExpected: ${expected_checksum}\nActual: ${actual_checksum}"
    fi
    success "Checksum verified"
}

# Get installation directory
get_install_dir() {
    if [[ -n "${DOTGH_INSTALL_DIR:-}" ]]; then
        echo "${DOTGH_INSTALL_DIR}"
        return
    fi

    # Default installation directories
    if [[ -d "/usr/local/bin" ]] && [[ -w "/usr/local/bin" ]]; then
        echo "/usr/local/bin"
    else
        local local_bin="${HOME}/.local/bin"
        mkdir -p "${local_bin}"
        echo "${local_bin}"
    fi
}

# Check if directory is in PATH
check_path() {
    local dir="$1"
    if [[ ":${PATH}:" != *":${dir}:"* ]]; then
        warn "${dir} is not in your PATH"
        echo ""
        echo "Add the following to your shell configuration file:"
        case "${SHELL}" in
            */bash)
                echo "  echo 'export PATH=\"${dir}:\$PATH\"' >> ~/.bashrc"
                ;;
            */zsh)
                echo "  echo 'export PATH=\"${dir}:\$PATH\"' >> ~/.zshrc"
                ;;
            */fish)
                echo "  fish_add_path ${dir}"
                ;;
            *)
                echo "  export PATH=\"${dir}:\$PATH\""
                ;;
        esac
        echo ""
    fi
}

# Main installation function
main() {
    echo ""
    info "Installing ${BINARY_NAME}..."
    echo ""

    # Detect platform
    local os arch
    os=$(detect_os)
    arch=$(detect_arch)
    info "Detected platform: ${os}/${arch}"

    # Get version
    local version="${DOTGH_VERSION:-}"
    if [[ -z "${version}" ]]; then
        info "Fetching latest version..."
        version=$(get_latest_version)
    fi
    info "Version: ${version}"

    # Determine archive format and filename
    local ext="tar.gz"
    if [[ "${os}" == "windows" ]]; then
        ext="zip"
    fi
    local filename="${BINARY_NAME}_${os}_${arch}.${ext}"
    local download_url="${GITHUB_RELEASES}/download/${version}/${filename}"
    local checksums_url="${GITHUB_RELEASES}/download/${version}/checksums.txt"

    # Create temp directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # Download archive
    local archive_path="${tmp_dir}/${filename}"
    download "${download_url}" "${archive_path}"

    # Download and verify checksum
    local checksums_path="${tmp_dir}/checksums.txt"
    download "${checksums_url}" "${checksums_path}"

    local expected_checksum
    expected_checksum=$(grep "${filename}" "${checksums_path}" | awk '{print $1}')
    if [[ -z "${expected_checksum}" ]]; then
        error "Checksum not found for ${filename}"
    fi
    verify_checksum "${archive_path}" "${expected_checksum}"

    # Extract archive
    info "Extracting archive..."
    cd "${tmp_dir}"
    if [[ "${ext}" == "tar.gz" ]]; then
        tar -xzf "${archive_path}"
    else
        if has_command unzip; then
            unzip -q "${archive_path}"
        else
            error "unzip command not found. Please install it."
        fi
    fi

    # Install binary
    local install_dir
    install_dir=$(get_install_dir)
    local binary_src="${tmp_dir}/${BINARY_NAME}"
    local binary_dest="${install_dir}/${BINARY_NAME}"

    if [[ "${os}" == "windows" ]]; then
        binary_src="${binary_src}.exe"
        binary_dest="${binary_dest}.exe"
    fi

    if [[ ! -f "${binary_src}" ]]; then
        error "Binary not found in archive"
    fi

    info "Installing to ${binary_dest}"
    if [[ -w "${install_dir}" ]]; then
        mv "${binary_src}" "${binary_dest}"
        chmod +x "${binary_dest}"
    else
        info "Requesting sudo permission to install..."
        sudo mv "${binary_src}" "${binary_dest}"
        sudo chmod +x "${binary_dest}"
    fi

    # Verify installation
    if [[ -x "${binary_dest}" ]]; then
        echo ""
        success "${BINARY_NAME} ${version} installed successfully!"
        echo ""
        "${binary_dest}" version 2>/dev/null || true
        echo ""
        check_path "${install_dir}"
    else
        error "Installation failed"
    fi
}

# Run main function
main "$@"
