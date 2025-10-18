#!/bin/bash

# build_mcp_wheel.sh - Build pyscn-mcp wheel for a specific platform
# Usage: build_mcp_wheel.sh <platform> <wheel_platform>

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

main() {
    local platform="$1"
    local wheel_platform="$2"

    if [[ -z "$platform" || -z "$wheel_platform" ]]; then
        echo -e "${RED}Usage: build_mcp_wheel.sh <platform> <wheel_platform>${NC}"
        exit 1
    fi

    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local python_dir="$(dirname "$script_dir")"
    local project_dir="$(dirname "$python_dir")"

    # Version detection (same as build_platform_wheel.sh)
    normalize_version() {
        local git_describe="$1"
        git_describe="${git_describe#v}"

        if [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)-([0-9]+)-g([0-9a-f]+)(-dirty)?$ ]]; then
            local base_version="${BASH_REMATCH[1]}"
            local commits_ahead="${BASH_REMATCH[2]}"
            local commit_hash="${BASH_REMATCH[3]}"
            local is_dirty="${BASH_REMATCH[4]}"

            if [[ -n "$is_dirty" ]]; then
                echo "${base_version}.post${commits_ahead}.dev0+g${commit_hash}"
            else
                echo "${base_version}.post${commits_ahead}+g${commit_hash}"
            fi
        elif [[ "$git_describe" =~ ^([0-9]+\.[0-9]+\.[0-9]+)(-dirty)?$ ]]; then
            local base_version="${BASH_REMATCH[1]}"
            local is_dirty="${BASH_REMATCH[2]}"

            if [[ -n "$is_dirty" ]]; then
                echo "${base_version}.dev0"
            else
                echo "$base_version"
            fi
        else
            echo "0.0.0.dev0"
        fi
    }

    local git_tag="${GITHUB_REF_NAME:-${GITHUB_REF#refs/tags/}}"
    local version

    if [[ -z "$git_tag" && -n "${GITHUB_REF:-}" ]]; then
        git_tag="${GITHUB_REF#refs/tags/}"
    fi

    if [[ -z "$git_tag" || "$git_tag" == "0.0.0.dev0" ]]; then
        local current_tag=$(git tag --points-at HEAD 2>/dev/null | grep "^v[0-9]" | head -1)
        if [[ -n "$current_tag" ]]; then
            git_tag="$current_tag"
        fi
    fi

    if [[ -n "$git_tag" && "$git_tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        version=$(normalize_version "${git_tag}")
        echo -e "${GREEN}Using CI tag version: $version${NC}"
    else
        version=$(normalize_version "$(git describe --tags --always --dirty 2>/dev/null || echo "0.0.0.dev0")")
        echo -e "${YELLOW}Using git describe version: $version${NC}"
    fi

    local go_module=$(go list -m)
    local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local date=$(date +%Y-%m-%d)

    echo -e "${GREEN}Building pyscn-mcp wheel for platform: $platform${NC}"
    echo "Version: $version"
    echo "GOOS: ${GOOS:-$(go env GOOS)}"
    echo "GOARCH: ${GOARCH:-$(go env GOARCH)}"

    local bin_dir="$python_dir/src/pyscn_mcp/bin"
    local dist_dir="$project_dir/dist"

    mkdir -p "$bin_dir" "$dist_dir"

    local binary_suffix=""
    if [[ "$platform" == *"windows"* ]]; then
        binary_suffix=".exe"
    fi

    local binary_filename="pyscn-mcp-${platform}${binary_suffix}"
    local binary_path="$bin_dir/$binary_filename"
    local cmd_path="$project_dir/cmd/pyscn-mcp"

    echo -e "${GREEN}Building pyscn-mcp binary...${NC}"

    local ldflags="-s -w -X '${go_module}/internal/version.Version=${version}' -X '${go_module}/internal/version.Commit=${commit}' -X '${go_module}/internal/version.Date=${date}'"

    case "$platform" in
        *"windows"*)
            CGO_ENABLED=1 go build -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
            ;;
        *"darwin"*)
            if [[ -z "${SDKROOT}" && -n "$(command -v xcrun)" ]]; then
                export SDKROOT="$(xcrun --show-sdk-path)"
            fi
            CGO_ENABLED=1 go build -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
            ;;
        *"linux"*)
            CGO_ENABLED=1 go build -ldflags="$ldflags" -o "$binary_path" "$cmd_path"
            ;;
    esac

    if [[ ! -f "$binary_path" ]]; then
        echo -e "${RED}Error: Failed to build pyscn-mcp${NC}"
        exit 1
    fi

    echo -e "${GREEN}Binary created: $binary_path${NC}"
    local size=$(du -h "$binary_path" | cut -f1)
    echo "  Binary size: $size"

    # Build wheel
    echo -e "${GREEN}Building wheel...${NC}"
    cd "$python_dir"

    # Use pyproject-mcp.toml via symlink
    ln -sf pyproject-mcp.toml pyproject.toml

    # Build wheel
    SETUPTOOLS_SCM_PRETEND_VERSION="$version" python -m build --wheel

    # Remove symlink
    rm -f pyproject.toml

    # Rename wheel to include correct platform tag
    local built_wheel=$(ls -t dist/pyscn_mcp-*.whl | head -1)
    if [[ -f "$built_wheel" ]]; then
        local wheel_name="pyscn_mcp-${version}-py3-none-${wheel_platform}.whl"
        local final_wheel="$dist_dir/$wheel_name"

        # Extract wheel, modify WHEEL file, update RECORD, and repack
        local temp_dir=$(mktemp -d)
        cd "$temp_dir"
        unzip -q "$python_dir/$built_wheel"

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
        rm -f "$python_dir/$built_wheel"
        zip -q -r "$final_wheel" .

        cd "$python_dir"
        rm -rf "$temp_dir"

        echo -e "${GREEN}Wheel created: $final_wheel${NC}"
    else
        echo -e "${RED}Error: Wheel not found${NC}"
        exit 1
    fi

    # Cleanup
    rm -rf "$bin_dir"
    rm -rf "$python_dir/dist"
    rm -rf "$python_dir/build"
    rm -rf "$python_dir/src/pyscn_mcp.egg-info"

    echo -e "${GREEN}pyscn-mcp wheel build completed!${NC}"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
