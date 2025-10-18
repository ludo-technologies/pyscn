#!/bin/bash

# build_platform_wheel.sh - Build wheel for a specific platform
# Usage: build_platform_wheel.sh <platform> <wheel_platform>

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

main() {
    local platform="$1"
    local wheel_platform="$2"
    
    if [[ -z "$platform" || -z "$wheel_platform" ]]; then
        echo -e "${RED}Usage: build_platform_wheel.sh <platform> <wheel_platform>${NC}"
        echo "Example: build_platform_wheel.sh darwin-arm64 macosx_11_0_arm64"
        exit 1
    fi
    
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local python_dir="$(dirname "$script_dir")"
    local project_dir="$(dirname "$python_dir")"
    
    # Function to convert git describe output to PEP 440 compliant version
    normalize_version() {
        local git_describe="$1"
        
        # Remove v prefix if present
        git_describe="${git_describe#v}"
        
        if [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$ ]]; then
            # After tag: 0.1.0-3-g278cb14[-dirty] -> 0.1.0.post3+g278cb14
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
        elif [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-beta\.([0-9]+)(-dirty)?$ ]]; then
            # Beta release: 0.1.0-beta.6[-dirty] -> 0.1.0b6[.dev0]
            local base_version="${BASH_REMATCH[1]}"
            local beta_number="${BASH_REMATCH[2]}"
            local is_dirty="${BASH_REMATCH[3]}"
            
            if [[ -n "$is_dirty" ]]; then
                echo "${base_version}b${beta_number}.dev0"
            else
                echo "${base_version}b${beta_number}"
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
        elif [[ "$git_describe" =~ ^[0-9a-f]+(-dirty)?$ ]]; then
            # No tags: 278cb14[-dirty] -> 0.0.0.dev0+g278cb14
            local commit_hash="${git_describe%-dirty}"
            echo "0.0.0.dev0+g${commit_hash}"
        else
            # Fallback for unexpected format
            echo "0.0.0.dev0"
        fi
    }
    
    # Auto-detect version from git tags and normalize to PEP 440
    # In CI, use the git tag directly if available to ensure consistency across platforms
    local git_tag="${GITHUB_REF_NAME:-${GITHUB_REF#refs/tags/}}"
    local version
    
    # Try multiple sources for the version tag
    if [[ -z "$git_tag" && -n "${GITHUB_REF:-}" ]]; then
        git_tag="${GITHUB_REF#refs/tags/}"
    fi
    
    # Additional fallback: check if we're at a specific tag with git command
    if [[ -z "$git_tag" || "$git_tag" == "0.0.0.dev0" ]]; then
        local current_tag=$(git tag --points-at HEAD 2>/dev/null | grep "^v[0-9]" | head -1)
        if [[ -n "$current_tag" ]]; then
            git_tag="$current_tag"
        fi
    fi
    
    if [[ -n "$git_tag" && "$git_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-beta\.[0-9]+)?$ ]]; then
        # Running in CI with a version tag - use tag directly
        version=$(normalize_version "${git_tag}")
        echo -e "${GREEN}Using CI tag version: $version${NC}"
    else
        # Local or non-tag build - use git describe
        version=$(normalize_version "$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0.dev0")")
        echo -e "${YELLOW}Using git describe version: $version${NC}"
    fi
    
    # Get build information for version injection
    local go_module=$(go list -m)
    local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local date=$(date +%Y-%m-%d)
    
    echo -e "${GREEN}Building wheel for platform: $platform${NC}"
    echo "Project dir: $project_dir"
    echo "Python dir: $python_dir"
    echo "Version: $version"
    echo "GOOS: ${GOOS:-$(go env GOOS)}"
    echo "GOARCH: ${GOARCH:-$(go env GOARCH)}"
    
    # Create directories
    local bin_dir="$python_dir/src/pyscn/bin"
    local dist_dir="$project_dir/dist"
    
    mkdir -p "$bin_dir" "$dist_dir"
    
    local binary_suffix=""
    if [[ "$platform" == *"windows"* ]]; then
        binary_suffix=".exe"
    fi
    
    local binaries=("pyscn" "pyscn-mcp")
    local built_binary_paths=()
    
    echo -e "${GREEN}Building Go binaries for $platform...${NC}"
    
    # Build Go binary with platform-specific settings
    local build_cmd="go build"
    local ldflags="-s -w -X '${go_module}/internal/version.Version=${version}' -X '${go_module}/internal/version.Commit=${commit}' -X '${go_module}/internal/version.Date=${date}' -X '${go_module}/internal/version.BuiltBy=build_platform_wheel.sh'"
    
    for binary in "${binaries[@]}"; do
        local binary_filename="${binary}-${platform}${binary_suffix}"
        local binary_path="$bin_dir/$binary_filename"
        local cmd_path="$project_dir/cmd/$binary"
        
        if [[ ! -d "$cmd_path" ]]; then
            echo -e "${RED}Error: Command directory not found for ${binary}: $cmd_path${NC}"
            exit 1
        fi
        
        echo "Building ${binary} -> ${binary_filename}"
        
        # Platform-specific build configuration
        case "$platform" in
            *"windows"*)
                echo "  Using MinGW-w64 for Windows build..."
                CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
                ;;
            *"darwin"*)
                echo "  Building for macOS..."
                # Set SDKROOT if not already set
                if [[ -z "${SDKROOT}" && -n "$(command -v xcrun)" ]]; then
                    export SDKROOT="$(xcrun --show-sdk-path)"
                    echo "  Set SDKROOT to: $SDKROOT"
                fi
                CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
                ;;
            *"linux"*)
                echo "  Building for Linux..."
                CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
                ;;
            *)
                echo -e "${RED}Error: Unknown platform $platform${NC}"
                exit 1
                ;;
        esac
        
        if [[ ! -f "$binary_path" ]]; then
            echo -e "${RED}Error: Failed to build ${binary} for $platform${NC}"
            echo "Expected binary at: $binary_path"
            exit 1
        fi
        
        echo -e "${GREEN}Binary created successfully: $binary_path${NC}"
        local size
        size=$(du -h "$binary_path" | cut -f1)
        echo "  Binary size: $size"
        
        built_binary_paths+=("$binary_path")
    done
    
    # Create wheel using the existing create_wheel.sh script
    echo -e "${GREEN}Creating wheel...${NC}"
    local create_args=(
        --platform "$wheel_platform"
        --version "$version"
        --output "$dist_dir"
    )
    
    for binary_path in "${built_binary_paths[@]}"; do
        create_args+=(--binary "$binary_path")
    done
    
    if ! "$script_dir/create_wheel.sh" "${create_args[@]}"; then
        echo -e "${RED}Error: Failed to create wheel for $wheel_platform${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Platform-specific wheel build completed successfully!${NC}"
    echo "Wheels created:"
    ls -la "$dist_dir"/*.whl 2>/dev/null || echo "No wheels found"

    # Build pyscn-mcp wheel using the same pyscn-mcp binary
    echo -e "${GREEN}Building pyscn-mcp wheel...${NC}"

    # Copy pyscn-mcp binary to python/src/pyscn_mcp/bin/
    local mcp_bin_dir="$python_dir/src/pyscn_mcp/bin"
    mkdir -p "$mcp_bin_dir"

    local mcp_binary_filename="pyscn-mcp-${platform}${binary_suffix}"
    local mcp_binary_path="$bin_dir/$mcp_binary_filename"
    cp "$mcp_binary_path" "$mcp_bin_dir/"

    # Build pyscn-mcp wheel from python/ directory
    cd "$python_dir"

    # Temporarily use pyproject-mcp.toml
    ln -sf pyproject-mcp.toml pyproject.toml

    # Build wheel
    pip install --user build >/dev/null 2>&1 || true
    # Use python3 for better cross-platform compatibility (especially Windows MSYS2)
    SETUPTOOLS_SCM_PRETEND_VERSION="$version" python3 -m build --wheel 2>/dev/null || \
    SETUPTOOLS_SCM_PRETEND_VERSION="$version" python -m build --wheel

    # Remove symlink
    rm -f pyproject.toml

    # Rename and move wheel
    local mcp_built_wheel=$(ls -t dist/pyscn_mcp-*.whl 2>/dev/null | head -1)
    if [[ -f "$mcp_built_wheel" ]]; then
        local mcp_wheel_name="pyscn_mcp-${version}-py3-none-${wheel_platform}.whl"

        # Extract wheel, modify WHEEL file, update RECORD, and repack
        local temp_dir=$(mktemp -d)
        cd "$temp_dir"
        unzip -q "$python_dir/$mcp_built_wheel"

        # Update platform tag in WHEEL file
        local wheel_file=$(find . -name "WHEEL" -type f)
        if [[ -f "$wheel_file" ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s/Tag: py3-none-any/Tag: py3-none-${wheel_platform}/" "$wheel_file"
            else
                sed -i "s/Tag: py3-none-any/Tag: py3-none-${wheel_platform}/" "$wheel_file"
            fi
        fi

        # Update RECORD file with new hash for WHEEL
        local record_file=$(find . -name "RECORD" -type f)
        if [[ -f "$record_file" && -f "$wheel_file" ]]; then
            local wheel_hash=$(python -c "import hashlib,base64; print('sha256=' + base64.urlsafe_b64encode(hashlib.sha256(open('$wheel_file', 'rb').read()).digest()).decode().rstrip('='))")
            local wheel_size=$(wc -c < "$wheel_file" | tr -d ' ')
            local wheel_path="${wheel_file#./}"
            if [[ "$OSTYPE" == "darwin"* ]]; then
                sed -i '' "s|^${wheel_path},.*|${wheel_path},${wheel_hash},${wheel_size}|" "$record_file"
            else
                sed -i "s|^${wheel_path},.*|${wheel_path},${wheel_hash},${wheel_size}|" "$record_file"
            fi
        fi

        # Repack wheel
        rm -f "$python_dir/$mcp_built_wheel"
        zip -q -r "$dist_dir/$mcp_wheel_name" .

        cd "$python_dir"
        rm -rf "$temp_dir"

        echo -e "${GREEN}pyscn-mcp wheel created: $dist_dir/$mcp_wheel_name${NC}"
    else
        echo -e "${YELLOW}Warning: pyscn-mcp wheel not created${NC}"
    fi

    # Cleanup (keep dist/ with wheels, only remove build artifacts)
    rm -rf "$mcp_bin_dir"
    rm -rf "$python_dir/build"
    rm -rf "$python_dir/dist"
    rm -rf "$python_dir/src/pyscn_mcp.egg-info"

    echo -e "${GREEN}All wheels created successfully!${NC}"
    echo "Final wheels:"
    ls -la "$dist_dir"/*.whl 2>/dev/null || echo "No wheels found"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
