#!/bin/bash

# create_wheel.sh - Create Python wheel without Python build tools
# This script creates a wheel file directly using shell commands and zip

set -e

# Configuration
PACKAGE_NAME="pyscn"  # Default, can be overridden with --package
binary_paths=()

# Function to convert git describe output to PEP 440 compliant version
normalize_version() {
    local git_describe="$1"
    
    # Remove v prefix if present
    git_describe="${git_describe#v}"
    
    # Check for SemVer prerelease versions (e.g., 0.1.0-beta.1, 0.1.0-alpha.1)
    if [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-(alpha|beta|rc)\.([0-9]+)(-([0-9]+)-g([0-9a-f]+))?(-dirty)?$ ]]; then
        # SemVer format: 0.1.0-beta.1 -> 0.1.0b1 (PEP 440)
        local base_version="${BASH_REMATCH[1]}"
        local prerelease_type="${BASH_REMATCH[2]}"
        local prerelease_num="${BASH_REMATCH[3]}"
        local commits_ahead="${BASH_REMATCH[5]}"
        local commit_hash="${BASH_REMATCH[6]}"
        local is_dirty="${BASH_REMATCH[7]}"
        
        # Convert to PEP 440 format
        case "$prerelease_type" in
            alpha) prerelease_type="a" ;;
            beta) prerelease_type="b" ;;
            rc) prerelease_type="rc" ;;
        esac
        
        if [[ -n "$commits_ahead" ]]; then
            # After prerelease tag
            if [[ -n "$is_dirty" ]]; then
                echo "${base_version}${prerelease_type}${prerelease_num}.post${commits_ahead}.dev0+g${commit_hash}"
            else
                echo "${base_version}${prerelease_type}${prerelease_num}.post${commits_ahead}+g${commit_hash}"
            fi
        elif [[ -n "$is_dirty" ]]; then
            # Dirty prerelease tag
            echo "${base_version}${prerelease_type}${prerelease_num}.dev0"
        else
            # Clean prerelease tag
            echo "${base_version}${prerelease_type}${prerelease_num}"
        fi
    # Check for beta/alpha/rc versions first (e.g., 0.1.0b1, 0.1.0a1, 0.1.0rc1)
    elif [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)(a|b|rc)([0-9]+)(-([0-9]+)-g([0-9a-f]+))?(-dirty)?$ ]]; then
        # Beta/alpha/rc version: 0.1.0b1[-3-g278cb14][-dirty]
        local base_version="${BASH_REMATCH[1]}"
        local prerelease_type="${BASH_REMATCH[2]}"
        local prerelease_num="${BASH_REMATCH[3]}"
        local commits_ahead="${BASH_REMATCH[5]}"
        local commit_hash="${BASH_REMATCH[6]}"
        local is_dirty="${BASH_REMATCH[7]}"
        
        if [[ -n "$commits_ahead" ]]; then
            # After prerelease tag
            if [[ -n "$is_dirty" ]]; then
                echo "${base_version}${prerelease_type}${prerelease_num}.post${commits_ahead}.dev0+g${commit_hash}"
            else
                echo "${base_version}${prerelease_type}${prerelease_num}.post${commits_ahead}+g${commit_hash}"
            fi
        elif [[ -n "$is_dirty" ]]; then
            # Dirty prerelease tag
            echo "${base_version}${prerelease_type}${prerelease_num}.dev0"
        else
            # Clean prerelease tag
            echo "${base_version}${prerelease_type}${prerelease_num}"
        fi
    elif [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$ ]]; then
        # After tag: 0.1.0-3-g278cb14[-dirty] -> 0.1.0.post3+g278cb14
        # Capture matches immediately to avoid interference
        local base_version="${BASH_REMATCH[1]}"
        local commits_ahead="${BASH_REMATCH[2]}"
        local commit_hash="${BASH_REMATCH[3]}"
        local is_dirty="${BASH_REMATCH[4]}"
        
        if [[ -n "$is_dirty" ]]; then
            # For dirty workspace, use dev version to avoid local version (PyPI rejection)
            echo "${base_version}.post${commits_ahead}.dev0+g${commit_hash}"
        else
            echo "${base_version}.post${commits_ahead}+g${commit_hash}"
        fi
    elif [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)(-dirty)?$ ]]; then
        # Clean or dirty tag: 0.1.0[-dirty] 
        local base_version="${BASH_REMATCH[1]}"
        local is_dirty="${BASH_REMATCH[2]}"
        
        if [[ -n "$is_dirty" ]]; then
            # For dirty workspace, append .dev0 instead of local version
            echo "${base_version}.dev0"
        else
            echo "$base_version"
        fi
    elif [[ "$git_describe" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        # Clean tag: 0.1.0 -> 0.1.0
        echo "$git_describe"
    elif [[ "$git_describe" =~ ^[0-9a-f]+(-dirty)?$ ]]; then
        # No tags: 278cb14[-dirty] -> 0.0.0.dev0+g278cb14
        local commit_hash="${git_describe%-dirty}"
        echo "0.0.0.dev0+g${commit_hash}"
    else
        # Fallback for unexpected format
        echo "0.0.0.dev0"
    fi
}

# Version will be set from command line or auto-detected
VERSION=""
PYTHON_TAG="py3"
ABI_TAG="none"

# Platform detection
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    # Normalize architecture for consistent naming
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64) arch="arm64" ;;
        # Keep arm64 as-is for macOS
    esac
    
    case "$os" in
        darwin)
            case "$arch" in
                amd64) echo "macosx_10_9_x86_64" ;;
                arm64) echo "macosx_11_0_arm64" ;;
                *) echo "unsupported"; exit 1 ;;
            esac
            ;;
        linux)
            case "$arch" in
                amd64) echo "manylinux_2_17_x86_64" ;;
                arm64) echo "manylinux_2_17_aarch64" ;;
                *) echo "unsupported"; exit 1 ;;
            esac
            ;;
        mingw*|msys*|cygwin*)
            case "$arch" in
                amd64) echo "win_amd64" ;;
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
    local output_dir="$2"
    
    if [[ -z "$platform_tag" || -z "$output_dir" ]]; then
        echo "Usage: create_wheel <platform_tag> <output_dir>"
        exit 1
    fi
    
    if [[ ${#binary_paths[@]} -eq 0 ]]; then
        echo "Error: No binaries specified for wheel creation"
        exit 1
    fi
    
    for binary_path in "${binary_paths[@]}"; do
        if [[ ! -f "$binary_path" ]]; then
            echo "Error: Binary not found at $binary_path"
            exit 1
        fi
    done
    
    # Ensure binaries have unique basenames and track required binaries
    local basenames=()
    local has_pyscn_binary=0
    local has_mcp_binary=0
    for binary_path in "${binary_paths[@]}"; do
        local basename="${binary_path##*/}"
        case "$basename" in
            pyscn-mcp-*)
                has_mcp_binary=1
                ;;
            pyscn-*)
                has_pyscn_binary=1
                ;;
        esac
        for seen_basename in "${basenames[@]}"; do
            if [[ "$seen_basename" == "$basename" ]]; then
                echo "Error: Duplicate binary name detected: $basename"
                exit 1
            fi
        done
        basenames+=("$basename")
    done

    # Validate required binaries based on package type
    if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
        if [[ "$has_pyscn_binary" -ne 1 ]]; then
            echo "Error: Missing pyscn CLI binary in provided --binary arguments"
            exit 1
        fi
        if [[ "$has_mcp_binary" -ne 1 ]]; then
            echo "Error: Missing pyscn-mcp MCP server binary in provided --binary arguments"
            exit 1
        fi
    elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
        if [[ "$has_mcp_binary" -ne 1 ]]; then
            echo "Error: Missing pyscn-mcp binary in provided --binary arguments"
            exit 1
        fi
    fi
    
    # Wheel filename
    local wheel_name="${PACKAGE_NAME}-${VERSION}-${PYTHON_TAG}-${ABI_TAG}-${platform_tag}.whl"
    local wheel_path="${output_dir}/${wheel_name}"
    
    # Create temporary directory
    local temp_dir
    temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    echo "Creating wheel: $wheel_name"
    echo "  Platform: $platform_tag"
    echo "  Output: $wheel_path"
    echo "  Binaries:"
    for binary_path in "${binary_paths[@]}"; do
        echo "    - $binary_path"
    done
    
    # Create wheel directory structure
    local wheel_dir="$temp_dir/wheel"
    local pkg_dir="$wheel_dir/$PACKAGE_NAME"
    local bin_dir="$pkg_dir/bin"
    local metadata_dir="$wheel_dir/${PACKAGE_NAME}-${VERSION}.dist-info"
    
    mkdir -p "$bin_dir"
    mkdir -p "$metadata_dir"
    
    # Copy Python source files based on package type
    local src_dir
    if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
        src_dir="$(dirname "$0")/../src/pyscn"
        cp "$src_dir/__init__.py" "$pkg_dir/"
        cp "$src_dir/__main__.py" "$pkg_dir/"
        cp "$src_dir/main.py" "$pkg_dir/"
        if [[ "$has_mcp_binary" -eq 1 ]]; then
            cp "$src_dir/mcp_main.py" "$pkg_dir/"
        fi
    elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
        src_dir="$(dirname "$0")/../src/pyscn_mcp"
        cp "$src_dir/__init__.py" "$pkg_dir/"
        cp "$src_dir/__main__.py" "$pkg_dir/"
    fi
    
    # Copy binaries
    for binary_path in "${binary_paths[@]}"; do
        cp "$binary_path" "$bin_dir/"
    done
    
    # Check if README exists
    if [[ ! -f "$readme_path" ]]; then
        echo "Error: README not found at $readme_path"
        exit 1
    fi

    # Create METADATA file based on package type
    if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
        cat > "$metadata_dir/METADATA" << EOF
Metadata-Version: 2.1
Name: $PACKAGE_NAME
Version: $VERSION
Summary: An intelligent Python code quality analyzer with architectural guidance
Home-page: https://github.com/ludo-technologies/pyscn
Author: DaisukeYoda
Author-email: daisukeyoda@users.noreply.github.com
License: MIT
Classifier: Development Status :: 5 - Production/Stable
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
Classifier: Programming Language :: Python :: 3.13
Classifier: Topic :: Software Development :: Quality Assurance
Classifier: Typing :: Typed
Requires-Python: >=3.8
Description-Content-Type: text/markdown

EOF
    elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
        cat > "$metadata_dir/METADATA" << EOF
Metadata-Version: 2.1
Name: pyscn-mcp
Version: $VERSION
Summary: MCP (Model Context Protocol) server for pyscn Python code analyzer
Home-page: https://github.com/ludo-technologies/pyscn
Author: DaisukeYoda
Author-email: daisukeyoda@users.noreply.github.com
License: MIT
Classifier: Development Status :: 5 - Production/Stable
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
Classifier: Programming Language :: Python :: 3.13
Classifier: Topic :: Software Development :: Quality Assurance
Classifier: Typing :: Typed
Requires-Python: >=3.8
Description-Content-Type: text/markdown

EOF
    fi

    # Append README content to METADATA
    cat "$readme_path" >> "$metadata_dir/METADATA"

    # Create WHEEL file
    cat > "$metadata_dir/WHEEL" << EOF
Wheel-Version: 1.0
Generator: pyscn-create-wheel
Root-Is-Purelib: false
Tag: $PYTHON_TAG-$ABI_TAG-$platform_tag
EOF

    # Create entry_points.txt based on package type
    {
        echo "[console_scripts]"
        if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
            if [[ "$has_pyscn_binary" -eq 1 ]]; then
                echo "pyscn = pyscn.__main__:main"
            fi
            if [[ "$has_mcp_binary" -eq 1 ]]; then
                echo "pyscn-mcp = pyscn.mcp_main:main"
            fi
        elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
            echo "pyscn-mcp = pyscn_mcp.__main__:main"
        fi
    } > "$metadata_dir/entry_points.txt"

    # Create RECORD file (proper CSV format)
    > "$metadata_dir/RECORD"  # Clear the file first
    (
        cd "$wheel_dir"

        # Helper to compute PEP 427-compliant urlsafe base64 sha256 digest
        compute_hash() {
            local f="$1"
            if command -v python3 >/dev/null 2>&1; then
                python3 - "$f" << 'PY'
import sys,hashlib,base64
p=sys.argv[1]
with open(p,'rb') as fh:
    d=hashlib.sha256(fh.read()).digest()
print('sha256=' + base64.urlsafe_b64encode(d).decode().rstrip('='))
PY
                return
            elif command -v python >/dev/null 2>&1; then
                python - "$f" << 'PY'
import sys,hashlib,base64
p=sys.argv[1]
with open(p,'rb') as fh:
    d=hashlib.sha256(fh.read()).digest()
print('sha256=' + base64.urlsafe_b64encode(d).decode().rstrip('='))
PY
                return
            elif command -v openssl >/dev/null 2>&1; then
                # openssl + base64; convert to urlsafe and strip '=' padding
                local b64
                b64=$(openssl dgst -sha256 -binary "$f" | base64 | tr '+/' '-_' | tr -d '=')
                echo "sha256=$b64"
                return
            fi
            echo ""
        }

        # Generate entries for all files except RECORD itself
        find . -type f ! -name "RECORD" | while IFS= read -r file; do
            # Remove leading ./
            file_path="${file#./}"

            # Calculate SHA256 hash (urlsafe base64, per PEP 427)
            file_hash=$(compute_hash "$file")

            # Get file size
            if [[ "$OSTYPE" == "darwin"* ]]; then
                file_size=$(stat -f%z "$file")
            else
                file_size=$(stat -c%s "$file")
            fi

            # Output CSV format: filepath,hash,size
            if [[ -n "$file_hash" ]]; then
                echo "$file_path,$file_hash,$file_size"
            else
                echo "$file_path,,$file_size"
            fi
        done | sort

        # Add RECORD file entry (no hash, no size for RECORD itself)
        echo "${PACKAGE_NAME}-${VERSION}.dist-info/RECORD,,"
    ) >> "$metadata_dir/RECORD"
    
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
    local output_dir="$project_dir/dist"
    local readme_path="$project_dir/README.md"
    binary_paths=()

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --platform)
                platform_tag="$2"
                shift 2
                ;;
            --binary)
                binary_paths+=("$2")
                shift 2
                ;;
            --output)
                output_dir="$2"
                shift 2
                ;;
            --version)
                VERSION="$2"
                shift 2
                ;;
            --package)
                PACKAGE_NAME="$2"
                shift 2
                ;;
            --readme)
                readme_path="$2"
                shift 2
                ;;
            --help)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --platform TAG    Platform tag (e.g., macosx_11_0_arm64)"
                echo "  --binary PATH     Path to binary to include (repeatable)"
                echo "  --output DIR      Output directory (default: dist/)"
                echo "  --version VER     Version string (default: auto-detect)"
                echo "  --package NAME    Package name: pyscn or pyscn_mcp (default: pyscn)"
                echo "  --readme PATH     Path to README file (default: README.md)"
                echo "  --help            Show this help"
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    # Set README path based on package if not specified
    if [[ "$PACKAGE_NAME" == "pyscn_mcp" && "$readme_path" == "$project_dir/README.md" ]]; then
        readme_path="$python_dir/README-mcp.md"
    fi
    
    # Auto-detect version if not specified
    if [[ -z "$VERSION" ]]; then
        VERSION=$(normalize_version "$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0.dev0")")
        echo "Auto-detected version: $VERSION"
    else
        echo "Using provided version: $VERSION"
    fi

    # Auto-detect platform if not specified
    if [[ -z "$platform_tag" ]]; then
        platform_tag=$(detect_platform)
    fi
    
    # Auto-detect binaries if not specified
    if [[ ${#binary_paths[@]} -eq 0 ]]; then
        local os=$(uname -s | tr '[:upper:]' '[:lower:]')
        local arch=$(uname -m)
        
        # Normalize architecture to match Python wrapper expectations
        case "$arch" in
            x86_64) arch="amd64" ;;
            aarch64) arch="arm64" ;;
            # Keep arm64 as-is for macOS
        esac
        
        local suffix=""
        case "$os" in
            mingw*|msys*|cygwin*)
                os="windows"
                suffix=".exe"
                ;;
            darwin|linux)
                # keep as-is
                ;;
            *)
                echo "Error: Unsupported OS for automatic binary detection: $os"
                exit 1
                ;;
        esac
        
        # Set bin_base based on package type
        local bin_base
        if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
            bin_base="$python_dir/src/pyscn/bin"
        elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
            bin_base="$python_dir/src/pyscn_mcp/bin"
        else
            echo "Error: Unknown package name: $PACKAGE_NAME"
            exit 1
        fi

        # Build expected_binaries based on package type
        local expected_binaries=()
        if [[ "$PACKAGE_NAME" == "pyscn" ]]; then
            # pyscn package needs both binaries
            expected_binaries=(
                "$bin_base/pyscn-${os}-${arch}${suffix}"
                "$bin_base/pyscn-mcp-${os}-${arch}${suffix}"
            )
        elif [[ "$PACKAGE_NAME" == "pyscn_mcp" ]]; then
            # pyscn_mcp package only needs the MCP server binary
            expected_binaries=(
                "$bin_base/pyscn-mcp-${os}-${arch}${suffix}"
            )
        fi
        
        local missing=()
        for candidate in "${expected_binaries[@]}"; do
            if [[ -f "$candidate" ]]; then
                binary_paths+=("$candidate")
            else
                missing+=("$candidate")
            fi
        done
        
        if [[ ${#missing[@]} -gt 0 ]]; then
            echo "Error: Unable to locate required binaries:"
            for candidate in "${missing[@]}"; do
                echo "  - $candidate"
            done
            exit 1
        fi
    fi
    
    create_wheel "$platform_tag" "$output_dir"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
