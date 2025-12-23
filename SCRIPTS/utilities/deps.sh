#!/usr/bin/env bash
# deps.sh - Dependency Management Utilities
# Handles checking and installing module dependencies

set -euo pipefail

# Check if a command is available
deps.command_exists() {
    local cmd="$1"
    command -v "$cmd" &>/dev/null
}

# Detect package manager
deps.detect_pkg_manager() {
    if command -v pacman &>/dev/null; then
        echo "pacman"
    elif command -v apt &>/dev/null; then
        echo "apt"
    elif command -v dnf &>/dev/null; then
        echo "dnf"
    elif command -v brew &>/dev/null; then
        echo "brew"
    else
        echo "unknown"
    fi
}

# Parse dependency from module metadata
# Format: DEPENDS_PKG: arch:yt-dlp debian:yt-dlp pip:yt-dlp
deps.parse_package() {
    local depends_line="$1"
    local pkg_manager="$2"

    # Extract package name for current package manager
    case "$pkg_manager" in
        pacman)
            echo "$depends_line" | grep -o "arch:[^ ]*" | cut -d: -f2
            ;;
        apt)
            echo "$depends_line" | grep -o "debian:[^ ]*" | cut -d: -f2
            ;;
        dnf)
            echo "$depends_line" | grep -o "fedora:[^ ]*" | cut -d: -f2
            ;;
        brew)
            echo "$depends_line" | grep -o "brew:[^ ]*" | cut -d: -f2
            ;;
        *)
            # Fallback to pip
            echo "$depends_line" | grep -o "pip:[^ ]*" | cut -d: -f2
            ;;
    esac
}

# Check if a dependency is satisfied
deps.check() {
    local dep_name="$1"

    # Check if command exists
    if deps.command_exists "$dep_name"; then
        return 0
    fi

    # Check if it's a python package
    if python3 -c "import $dep_name" 2>/dev/null; then
        return 0
    fi

    return 1
}

# Install a package
deps.install() {
    local package="$1"
    local pkg_manager="$2"

    case "$pkg_manager" in
        pacman)
            sudo pacman -S --noconfirm "$package"
            ;;
        apt)
            sudo apt-get install -y "$package"
            ;;
        dnf)
            sudo dnf install -y "$package"
            ;;
        brew)
            brew install "$package"
            ;;
        pip)
            pip3 install --user "$package"
            ;;
        *)
            echo "Error: Unknown package manager: $pkg_manager" >&2
            return 1
            ;;
    esac
}

# Get module dependencies from file
deps.get_module_deps() {
    local module_file="$1"

    # Extract DEPENDS lines from module
    grep "^# DEPENDS:" "$module_file" 2>/dev/null | sed 's/^# DEPENDS: *//' || true
}

# Get module package dependencies
deps.get_module_packages() {
    local module_file="$1"

    # Extract DEPENDS_PKG lines
    grep "^# DEPENDS_PKG:" "$module_file" 2>/dev/null | sed 's/^# DEPENDS_PKG: *//' || true
}

# Get required config keys
deps.get_required_config() {
    local module_file="$1"

    grep "^# CONFIG_REQUIRED:" "$module_file" 2>/dev/null | sed 's/^# CONFIG_REQUIRED: *//' || true
}

# Get config prompts
deps.get_config_prompt() {
    local module_file="$1"
    local config_key="$2"

    # Find the CONFIG_PROMPT line that follows CONFIG_REQUIRED
    awk -v key="$config_key" '
        /^# CONFIG_REQUIRED:/ && $3 == key {
            found=1
            next
        }
        found && /^# CONFIG_PROMPT:/ {
            sub(/^# CONFIG_PROMPT: *"?/, "")
            sub(/"?$/, "")
            print
            exit
        }
        /^# CONFIG_REQUIRED:/ {
            found=0
        }
    ' "$module_file"
}

# Check all dependencies for a module
deps.check_module() {
    local module_file="$1"
    local missing_deps=()
    local missing_packages=()

    # Check command dependencies
    while IFS= read -r dep; do
        [[ -z "$dep" ]] && continue
        if ! deps.check "$dep"; then
            missing_deps+=("$dep")
        fi
    done < <(deps.get_module_deps "$module_file")

    # If we have missing deps, get package names
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        local pkg_manager
        pkg_manager=$(deps.detect_pkg_manager)

        while IFS= read -r pkg_line; do
            [[ -z "$pkg_line" ]] && continue
            local pkg
            pkg=$(deps.parse_package "$pkg_line" "$pkg_manager")
            [[ -n "$pkg" ]] && missing_packages+=("$pkg")
        done < <(deps.get_module_packages "$module_file")
    fi

    # Return arrays via stdout
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        echo "MISSING_DEPS:${missing_deps[*]}"
    fi
    if [[ ${#missing_packages[@]} -gt 0 ]]; then
        echo "MISSING_PACKAGES:${missing_packages[*]}"
    fi
}

# Install module dependencies
deps.install_module_deps() {
    local module_file="$1"
    local pkg_manager
    pkg_manager=$(deps.detect_pkg_manager)

    local packages
    while IFS= read -r pkg_line; do
        [[ -z "$pkg_line" ]] && continue
        local pkg
        pkg=$(deps.parse_package "$pkg_line" "$pkg_manager")
        if [[ -n "$pkg" ]]; then
            packages+=("$pkg")
        fi
    done < <(deps.get_module_packages "$module_file")

    if [[ ${#packages[@]} -gt 0 ]]; then
        for pkg in "${packages[@]}"; do
            echo "Installing $pkg..."
            if [[ "$pkg_manager" == "unknown" ]]; then
                # Fallback to pip
                deps.install "$pkg" "pip"
            else
                deps.install "$pkg" "$pkg_manager"
            fi
        done
        return 0
    fi

    return 1
}
