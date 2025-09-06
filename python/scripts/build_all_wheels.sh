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
    
    echo -e "${GREEN}Building wheels for all platforms...${NC}"
    echo "Project dir: $project_dir"
    echo "Python dir: $python_dir"
    
    # Create directories
    local bin_dir="$python_dir/src/pyqol/bin"
    local dist_dir="$python_dir/dist"
    
    mkdir -p "$bin_dir" "$dist_dir"
    
    # Clean previous builds
    rm -rf "$bin_dir"/* "$dist_dir"/*
    
    # Build binaries and wheels for each platform
    for platform_config in "${PLATFORMS[@]}"; do
        IFS=':' read -r goos goarch wheel_platform <<< "$platform_config"
        
        local binary_name="pyqol-${goos}-${goarch}"
        if [[ "$goos" == "windows" ]]; then
            binary_name="${binary_name}.exe"
        fi
        
        local binary_path="$bin_dir/$binary_name"
        
        echo -e "${GREEN}Building for ${goos}/${goarch}...${NC}"
        
        # Build Go binary
        if ! GOOS="$goos" GOARCH="$goarch" go build \
            -ldflags="-s -w" \
            -o "$binary_path" \
            "$project_dir/cmd/pyqol"; then
            echo -e "${YELLOW}Warning: Failed to build binary for ${goos}/${goarch}${NC}"
            continue
        fi
        
        echo "  Binary created: $binary_path"
        
        # Create wheel
        if ! "$script_dir/create_wheel.sh" \
            --platform "$wheel_platform" \
            --binary "$binary_path" \
            --output "$dist_dir"; then
            echo -e "${YELLOW}Warning: Failed to create wheel for ${wheel_platform}${NC}"
            continue
        fi
    done
    
    echo -e "${GREEN}Build completed!${NC}"
    echo "Wheels created:"
    ls -la "$dist_dir"/*.whl 2>/dev/null || echo "No wheels found"
    
    echo ""
    echo "To test a wheel:"
    echo "  pip install python/dist/pyqol-*.whl"
    echo "  pyqol --version"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi