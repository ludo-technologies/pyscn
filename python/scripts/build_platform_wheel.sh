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
    local version=$(normalize_version "$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0.dev0")")
    
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
    
    # Determine binary name based on platform
    local binary_name="pyscn-$platform"
    if [[ "$platform" == *"windows"* ]]; then
        binary_name="${binary_name}.exe"
    fi
    
    local binary_path="$bin_dir/$binary_name"
    
    echo -e "${GREEN}Building Go binary for $platform...${NC}"
    
    # Build Go binary with platform-specific settings
    local build_cmd="go build"
    local ldflags="-s -w -X '${go_module}/internal/version.Version=${version}' -X '${go_module}/internal/version.Commit=${commit}' -X '${go_module}/internal/version.Date=${date}' -X '${go_module}/internal/version.BuiltBy=build_platform_wheel.sh'"
    
    # Platform-specific build configuration
    case "$platform" in
        *"windows"*)
            echo "Building for Windows with MinGW-w64..."
            CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$project_dir/cmd/pyscn"
            ;;
        *"darwin"*)
            echo "Building for macOS..."
            # Set SDKROOT if not already set
            if [[ -z "${SDKROOT}" && -n "$(command -v xcrun)" ]]; then
                export SDKROOT="$(xcrun --show-sdk-path)"
                echo "Set SDKROOT to: $SDKROOT"
            fi
            CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$project_dir/cmd/pyscn"
            ;;
        *"linux"*)
            echo "Building for Linux..."
            CGO_ENABLED=1 $build_cmd -ldflags="$ldflags" -o "$binary_path" "$project_dir/cmd/pyscn"
            ;;
        *)
            echo -e "${RED}Error: Unknown platform $platform${NC}"
            exit 1
            ;;
    esac
    
    if [[ ! -f "$binary_path" ]]; then
        echo -e "${RED}Error: Failed to build binary for $platform${NC}"
        echo "Expected binary at: $binary_path"
        exit 1
    fi
    
    echo -e "${GREEN}Binary created successfully: $binary_path${NC}"
    local size=$(du -h "$binary_path" | cut -f1)
    echo "Binary size: $size"
    
    # Create wheel using the existing create_wheel.sh script
    echo -e "${GREEN}Creating wheel...${NC}"
    if ! "$script_dir/create_wheel.sh" \
        --platform "$wheel_platform" \
        --binary "$binary_path" \
        --output "$dist_dir"; then
        echo -e "${RED}Error: Failed to create wheel for $wheel_platform${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}Platform-specific wheel build completed successfully!${NC}"
    echo "Wheels created:"
    ls -la "$dist_dir"/*.whl 2>/dev/null || echo "No wheels found"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi