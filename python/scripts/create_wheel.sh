#!/bin/bash

# create_wheel.sh - Create Python wheel without Python build tools
# This script creates a wheel file directly using shell commands and zip

set -e

# Configuration
PACKAGE_NAME="pyqol"
VERSION="0.1.0"
PYTHON_TAG="py3"
ABI_TAG="none"

# Platform detection
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        darwin)
            case "$arch" in
                x86_64) echo "macosx_10_9_x86_64" ;;
                arm64) echo "macosx_11_0_arm64" ;;
                *) echo "unsupported"; exit 1 ;;
            esac
            ;;
        linux)
            case "$arch" in
                x86_64) echo "manylinux_2_17_x86_64" ;;
                aarch64) echo "manylinux_2_17_aarch64" ;;
                *) echo "unsupported"; exit 1 ;;
            esac
            ;;
        mingw*|msys*|cygwin*)
            case "$arch" in
                x86_64) echo "win_amd64" ;;
                *) echo "unsupported"; exit 1 ;;
            esac
            ;;
        *)
            echo "unsupported"
            exit 1
            ;;
    esac
}

# Create wheel for specified platform
create_wheel() {
    local platform_tag="$1"
    local binary_path="$2"
    local output_dir="$3"
    
    if [[ -z "$platform_tag" || -z "$binary_path" || -z "$output_dir" ]]; then
        echo "Usage: create_wheel <platform_tag> <binary_path> <output_dir>"
        exit 1
    fi
    
    if [[ ! -f "$binary_path" ]]; then
        echo "Error: Binary not found at $binary_path"
        exit 1
    fi
    
    # Wheel filename
    local wheel_name="${PACKAGE_NAME}-${VERSION}-${PYTHON_TAG}-${ABI_TAG}-${platform_tag}.whl"
    local wheel_path="${output_dir}/${wheel_name}"
    
    # Create temporary directory
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    echo "Creating wheel: $wheel_name"
    echo "  Platform: $platform_tag"
    echo "  Binary: $binary_path"
    echo "  Output: $wheel_path"
    
    # Create wheel directory structure
    local wheel_dir="$temp_dir/wheel"
    local pkg_dir="$wheel_dir/$PACKAGE_NAME"
    local bin_dir="$pkg_dir/bin"
    local metadata_dir="$wheel_dir/${PACKAGE_NAME}-${VERSION}.dist-info"
    
    mkdir -p "$bin_dir"
    mkdir -p "$metadata_dir"
    
    # Copy Python source files
    local src_dir="$(dirname "$0")/../src/pyqol"
    cp "$src_dir/__init__.py" "$pkg_dir/"
    cp "$src_dir/__main__.py" "$pkg_dir/"
    cp "$src_dir/main.py" "$pkg_dir/"
    
    # Copy binary
    cp "$binary_path" "$bin_dir/"
    
    # Create METADATA file
    cat > "$metadata_dir/METADATA" << EOF
Metadata-Version: 2.1
Name: $PACKAGE_NAME
Version: $VERSION
Summary: A next-generation Python static analysis tool using Control Flow Graph and tree edit distance algorithms
Home-page: https://github.com/pyqol/pyqol
Author: pyqol team
Author-email: team@pyqol.dev
License: MIT
Classifier: Development Status :: 4 - Beta
Classifier: Environment :: Console
Classifier: Intended Audience :: Developers
Classifier: License :: OSI Approved :: MIT License
Classifier: Operating System :: OS Independent
Classifier: Programming Language :: Python :: 3
Classifier: Programming Language :: Python :: 3.8
Classifier: Programming Language :: Python :: 3.9
Classifier: Programming Language :: Python :: 3.10
Classifier: Programming Language :: Python :: 3.11
Classifier: Programming Language :: Python :: 3.12
Classifier: Topic :: Software Development :: Quality Assurance
Requires-Python: >=3.8
Description-Content-Type: text/markdown

# pyqol - Python Quality of Life

A next-generation Python static analysis tool that uses Control Flow Graph (CFG) and tree edit distance algorithms to provide deep code quality insights beyond traditional linters.
EOF

    # Create WHEEL file
    cat > "$metadata_dir/WHEEL" << EOF
Wheel-Version: 1.0
Generator: pyqol-create-wheel
Root-Is-Purelib: false
Tag: $PYTHON_TAG-$ABI_TAG-$platform_tag
EOF

    # Create entry_points.txt
    cat > "$metadata_dir/entry_points.txt" << EOF
[console_scripts]
pyqol = pyqol.__main__:main
EOF

    # Create RECORD file (simplified - just list files)
    find "$wheel_dir" -type f -exec basename {} \; | sort > "$metadata_dir/RECORD"
    
    # Create the wheel (zip file)
    mkdir -p "$output_dir"
    (cd "$wheel_dir" && zip -r "$wheel_path" . -x "*.DS_Store")
    
    echo "✅ Wheel created successfully: $wheel_path"
    
    # Verify wheel
    if [[ -f "$wheel_path" ]]; then
        local size=$(du -h "$wheel_path" | cut -f1)
        echo "   Size: $size"
        echo "   Contents:"
        unzip -l "$wheel_path" | head -20
    fi
}

# Main execution
main() {
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local python_dir="$(dirname "$script_dir")"
    local project_dir="$(dirname "$python_dir")"
    
    # Default values
    local platform_tag=""
    local binary_path=""
    local output_dir="$python_dir/dist"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --platform)
                platform_tag="$2"
                shift 2
                ;;
            --binary)
                binary_path="$2"
                shift 2
                ;;
            --output)
                output_dir="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --platform TAG    Platform tag (e.g., macosx_11_0_arm64)"
                echo "  --binary PATH     Path to pyqol binary"
                echo "  --output DIR      Output directory (default: dist/)"
                echo "  --help           Show this help"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Auto-detect platform if not specified
    if [[ -z "$platform_tag" ]]; then
        platform_tag=$(detect_platform)
    fi
    
    # Auto-detect binary if not specified
    if [[ -z "$binary_path" ]]; then
        local os=$(uname -s | tr '[:upper:]' '[:lower:]')
        local arch=$(uname -m)
        
        # Normalize architecture to match Python wrapper expectations
        case "$arch" in
            x86_64) arch="amd64" ;;
            aarch64) arch="arm64" ;;
            # Keep arm64 as-is for macOS
        esac
        
        local binary_name="pyqol-${os}-${arch}"
        if [[ "$os" == *"mingw"* || "$os" == *"msys"* || "$os" == *"cygwin"* ]]; then
            binary_name="${binary_name}.exe"
        fi
        
        binary_path="$python_dir/src/pyqol/bin/$binary_name"
    fi
    
    create_wheel "$platform_tag" "$binary_path" "$output_dir"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi