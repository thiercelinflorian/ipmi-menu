#!/bin/bash
# IPMI Menu - Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/thiercelinflorian/ipmi-menu/refs/heads/main/install.sh | bash
#        curl -fsSL ... | bash -s -- --upgrade
#        curl -fsSL ... | bash -s -- --uninstall

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PACKAGE_NAME="ipmi-menu"
MIN_PYTHON_VERSION="3.9"
MODE="install"

# -----------------------------------------------------------------------------
# Utility functions
# -----------------------------------------------------------------------------

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[OK]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

version_gte() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

get_python_version() {
    "$1" -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')" 2>/dev/null
}

usage() {
    echo ""
    echo -e "${BLUE}IPMI Menu - Installation Script${NC}"
    echo ""
    echo "Usage:"
    echo "  ./install.sh [OPTIONS]"
    echo "  curl -fsSL <url> | bash -s -- [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --install     Install ipmi-menu (default)"
    echo "  --upgrade     Upgrade ipmi-menu to latest version"
    echo "  --uninstall   Remove ipmi-menu"
    echo "  --help        Show this help message"
    echo ""
    exit 0
}

# -----------------------------------------------------------------------------
# Parse arguments
# -----------------------------------------------------------------------------

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --install)
                MODE="install"
                shift
                ;;
            --upgrade)
                MODE="upgrade"
                shift
                ;;
            --uninstall)
                MODE="uninstall"
                shift
                ;;
            --help|-h)
                usage
                ;;
            *)
                warn "Unknown option: $1"
                usage
                ;;
        esac
    done
}

# -----------------------------------------------------------------------------
# OS Detection
# -----------------------------------------------------------------------------

detect_os() {
    case "$(uname -s)" in
        Darwin*)    OS="macos" ;;
        Linux*)     OS="linux" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac

    if [[ "$OS" == "linux" ]]; then
        if command_exists apt-get; then
            DISTRO="debian"
        elif command_exists dnf; then
            DISTRO="fedora"
        elif command_exists yum; then
            DISTRO="rhel"
        elif command_exists pacman; then
            DISTRO="arch"
        elif command_exists apk; then
            DISTRO="alpine"
        else
            DISTRO="unknown"
        fi
    fi

    info "Detected OS: $OS${DISTRO:+ ($DISTRO)}"
}

# -----------------------------------------------------------------------------
# Homebrew Installation (macOS)
# -----------------------------------------------------------------------------

install_homebrew() {
    if command_exists brew; then
        success "Homebrew is already installed"
        return 0
    fi

    info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

    if [[ -f "/opt/homebrew/bin/brew" ]]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [[ -f "/usr/local/bin/brew" ]]; then
        eval "$(/usr/local/bin/brew shellenv)"
    fi

    success "Homebrew installed successfully"
}

# -----------------------------------------------------------------------------
# Python Installation
# -----------------------------------------------------------------------------

find_suitable_python() {
    local candidates=("python3" "python3.14" "python3.13" "python3.12" "python3.11" "python3.10" "python3.9" "python")

    for cmd in "${candidates[@]}"; do
        if command_exists "$cmd"; then
            local version
            version=$(get_python_version "$cmd")
            if [[ -n "$version" ]] && version_gte "$version" "$MIN_PYTHON_VERSION"; then
                PYTHON_CMD="$cmd"
                PYTHON_VERSION="$version"
                return 0
            fi
        fi
    done
    return 1
}

install_python() {
    if find_suitable_python; then
        success "Python $PYTHON_VERSION found ($PYTHON_CMD)"
        return 0
    fi

    info "Python >= $MIN_PYTHON_VERSION not found. Installing..."

    case "$OS" in
        macos)
            install_homebrew
            brew install python@3.12
            ;;
        linux)
            case "$DISTRO" in
                debian)
                    sudo apt-get update
                    sudo apt-get install -y python3 python3-pip python3-venv
                    ;;
                fedora)
                    sudo dnf install -y python3 python3-pip
                    ;;
                rhel)
                    sudo yum install -y python3 python3-pip
                    ;;
                arch)
                    sudo pacman -Sy --noconfirm python python-pip
                    ;;
                alpine)
                    sudo apk add --no-cache python3 py3-pip
                    ;;
                *)
                    error "Cannot auto-install Python. Please install Python >= $MIN_PYTHON_VERSION manually."
                    ;;
            esac
            ;;
    esac

    if ! find_suitable_python; then
        error "Failed to install Python >= $MIN_PYTHON_VERSION"
    fi

    success "Python $PYTHON_VERSION installed successfully"
}

# -----------------------------------------------------------------------------
# pipx Installation
# -----------------------------------------------------------------------------

install_pipx() {
    if command_exists pipx; then
        success "pipx is already installed"
        return 0
    fi

    info "Installing pipx..."

    case "$OS" in
        macos)
            if command_exists brew; then
                brew install pipx
            else
                "$PYTHON_CMD" -m pip install --user pipx
            fi
            ;;
        linux)
            case "$DISTRO" in
                debian)
                    sudo apt-get install -y pipx 2>/dev/null || "$PYTHON_CMD" -m pip install --user pipx
                    ;;
                fedora)
                    sudo dnf install -y pipx 2>/dev/null || "$PYTHON_CMD" -m pip install --user pipx
                    ;;
                arch)
                    sudo pacman -Sy --noconfirm python-pipx 2>/dev/null || "$PYTHON_CMD" -m pip install --user pipx
                    ;;
                *)
                    "$PYTHON_CMD" -m pip install --user pipx
                    ;;
            esac
            ;;
    esac

    if ! command_exists pipx; then
        "$PYTHON_CMD" -m pipx ensurepath 2>/dev/null || true
        export PATH="$HOME/.local/bin:$PATH"
    fi

    if ! command_exists pipx; then
        error "Failed to install pipx"
    fi

    success "pipx installed successfully"
}

# -----------------------------------------------------------------------------
# ipmitool Check
# -----------------------------------------------------------------------------

check_ipmitool() {
    if command_exists ipmitool; then
        success "ipmitool is available"
        return 0
    fi

    warn "ipmitool is not installed"
    info "Installing ipmitool..."

    case "$OS" in
        macos)
            if command_exists brew; then
                brew install ipmitool
            else
                warn "Please install ipmitool manually: brew install ipmitool"
                return 1
            fi
            ;;
        linux)
            case "$DISTRO" in
                debian)
                    sudo apt-get install -y ipmitool
                    ;;
                fedora)
                    sudo dnf install -y ipmitool
                    ;;
                rhel)
                    sudo yum install -y ipmitool
                    ;;
                arch)
                    sudo pacman -Sy --noconfirm ipmitool
                    ;;
                alpine)
                    sudo apk add --no-cache ipmitool
                    ;;
                *)
                    warn "Please install ipmitool manually for your distribution"
                    return 1
                    ;;
            esac
            ;;
    esac

    if command_exists ipmitool; then
        success "ipmitool installed successfully"
    else
        warn "Could not install ipmitool automatically. Please install it manually."
    fi
}

# -----------------------------------------------------------------------------
# ipmi-menu Install / Upgrade / Uninstall
# -----------------------------------------------------------------------------

do_install() {
    info "Installing $PACKAGE_NAME from PyPI..."

    if pipx list 2>/dev/null | grep -q "$PACKAGE_NAME"; then
        warn "$PACKAGE_NAME is already installed. Use --upgrade to update."
        return 0
    fi

    pipx install "$PACKAGE_NAME"
    success "$PACKAGE_NAME installed successfully"
}

do_upgrade() {
    info "Upgrading $PACKAGE_NAME..."

    if ! pipx list 2>/dev/null | grep -q "$PACKAGE_NAME"; then
        warn "$PACKAGE_NAME is not installed. Installing instead..."
        pipx install "$PACKAGE_NAME"
    else
        pipx upgrade "$PACKAGE_NAME"
    fi

    success "$PACKAGE_NAME upgraded successfully"
}

do_uninstall() {
    info "Uninstalling $PACKAGE_NAME..."

    if ! pipx list 2>/dev/null | grep -q "$PACKAGE_NAME"; then
        warn "$PACKAGE_NAME is not installed"
        return 0
    fi

    pipx uninstall "$PACKAGE_NAME"
    success "$PACKAGE_NAME uninstalled successfully"
}

# -----------------------------------------------------------------------------
# Main
# -----------------------------------------------------------------------------

main() {
    parse_args "$@"

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         IPMI Menu - Installer            ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"
    echo ""

    case "$MODE" in
        uninstall)
            if command_exists pipx; then
                do_uninstall
            else
                error "pipx not found. Cannot uninstall."
            fi
            echo ""
            echo -e "${GREEN}$PACKAGE_NAME has been removed.${NC}"
            echo ""
            exit 0
            ;;
        *)
            detect_os
            install_python
            install_pipx
            check_ipmitool

            if [[ "$MODE" == "upgrade" ]]; then
                do_upgrade
            else
                do_install
            fi
            ;;
    esac

    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════╗${NC}"
    if [[ "$MODE" == "upgrade" ]]; then
        echo -e "${GREEN}║         Upgrade complete!                ║${NC}"
    else
        echo -e "${GREEN}║       Installation complete!             ║${NC}"
    fi
    echo -e "${GREEN}╚══════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "Run ${BLUE}ipmi-menu${NC} to start the application."
    echo ""
    info "Reloading shell to update PATH..."
    exec $SHELL
}

main "$@"
