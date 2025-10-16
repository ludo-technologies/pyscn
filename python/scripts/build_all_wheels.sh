#!/bin/bash

# build_all_wheels.sh - Build wheels for all supported platforms
# This script builds Go binaries for all platforms and creates wheels

set -e

# Platform configurations (GOOS, GOARCH, wheel_platform_tag)
PLATFORMS=(
    "darwin:amd64:macosx_10_9_x86_64"
    "darwin:arm64:macosx_11_0_arm64" 
    "linux:amd64:manylinux_2_17_x86_64"
    "linux:arm64:manylinux_2_17_aarch64"
    "windows:amd64:win_amd64"
)

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

main() {
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

    # Auto-detect version from git tags and normalize to PEP 440
    local version=$(normalize_version "$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0.dev0")")
    
    # Get build information for version injection
    local go_module=$(go list -m)
    local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local date=$(date +%Y-%m-%d)
    
    echo -e "${GREEN}Building wheels for all platforms...${NC}"
    echo "Project dir: $project_dir"
    echo "Python dir: $python_dir"
    echo "Version: $version"
    
    # Create directories
    local bin_dir="$python_dir/src/pyscn/bin"
    local dist_dir="$project_dir/dist"
    
    mkdir -p "$bin_dir" "$dist_dir"
    
    # Clean previous builds
    rm -rf "$bin_dir"/* "$dist_dir"/*
    
    # Build binaries and wheels for each platform
    for platform_config in "${PLATFORMS[@]}"; do
        IFS=':' read -r goos goarch wheel_platform <<< "$platform_config"
        
        local binary_suffix=""
        if [[ "$goos" == "windows" ]]; then
            binary_suffix=".exe"
        fi
        
        echo -e "${GREEN}Building for ${goos}/${goarch}...${NC}"
        local binaries=("pyscn" "pyscn-mcp")
        local built_binary_paths=()
        local build_failed=0
        
        for binary in "${binaries[@]}"; do
            local binary_name="${binary}-${goos}-${goarch}${binary_suffix}"
            local binary_path="$bin_dir/$binary_name"
            local cmd_path="$project_dir/cmd/$binary"
            
            if [[ ! -d "$cmd_path" ]]; then
                echo -e "${YELLOW}Warning: Command directory not found (${cmd_path}); skipping ${goos}/${goarch}${NC}"
                build_failed=1
                break
            fi
            
            echo "  Building ${binary} -> ${binary_name}"
            
            if ! GOOS="$goos" GOARCH="$goarch" go build \
                -ldflags="-s -w \
                    -X '${go_module}/internal/version.Version=${version}' \
                    -X '${go_module}/internal/version.Commit=${commit}' \
                    -X '${go_module}/internal/version.Date=${date}' \
                    -X '${go_module}/internal/version.BuiltBy=build_all_wheels.sh'" \
                -o "$binary_path" \
                "$cmd_path"; then
                echo -e "${YELLOW}Warning: Failed to build ${binary} for ${goos}/${goarch}${NC}"
                build_failed=1
                break
            fi
            
            echo "    Binary created: $binary_path"
            built_binary_paths+=("$binary_path")
        done
        
        if [[ "$build_failed" -eq 1 ]]; then
            echo -e "${RED}Error: Failed to build required binaries for ${goos}/${goarch}${NC}"
            return 1
        fi
        
        local create_args=(
            --platform "$wheel_platform"
            --output "$dist_dir"
            --version "$version"
        )
        for binary_path in "${built_binary_paths[@]}"; do
            create_args+=(--binary "$binary_path")
        done
        
        if ! "$script_dir/create_wheel.sh" "${create_args[@]}"; then
            echo -e "${RED}Error: Failed to create wheel for ${wheel_platform}${NC}"
            return 1
        fi
    done
    
    echo -e "${GREEN}Build completed!${NC}"
    echo "Wheels created:"
    ls -la "$dist_dir"/*.whl 2>/dev/null || echo "No wheels found"
    
    echo ""
    echo "To test a wheel:"
    echo "  pip install dist/pyscn-*.whl"
    echo "  pyscn --version"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
