"""
Entry point for pyscn-mcp MCP server.

This module provides a Python wrapper for the Go-implemented pyscn-mcp binary.
It automatically detects the platform and executes the appropriate MCP server binary.
"""

import os
import sys
import platform
from pathlib import Path


def get_mcp_binary_path() -> str:
    """
    Get the path to the pyscn-mcp binary for the current platform.

    Returns:
        str: Path to the pyscn-mcp binary.

    Raises:
        FileNotFoundError: If the binary is not found for the current platform.
    """
    system = platform.system().lower()
    machine = platform.machine().lower()

    # Normalize architecture names
    if machine in ('x86_64', 'amd64'):
        machine = 'amd64'
    elif machine in ('aarch64', 'arm64'):
        machine = 'arm64'
    else:
        raise FileNotFoundError(
            f"Unsupported architecture: {machine}. "
            f"Supported architectures: amd64, arm64"
        )

    # Determine binary name
    binary_name = f"pyscn-mcp-{system}-{machine}"
    if system == "windows":
        binary_name += ".exe"

    # Binary path within the package
    binary_path = Path(__file__).parent / "bin" / binary_name

    if not binary_path.exists():
        raise FileNotFoundError(
            f"pyscn-mcp binary not found for platform {system}-{machine}.\n"
            f"Expected location: {binary_path}\n"
            f"Please check that the package was installed correctly."
        )

    return str(binary_path)


def main():
    """
    Main entry point for pyscn-mcp MCP server.

    Replaces the current process with the Go-implemented MCP server binary.
    This ensures proper stdio handling for MCP's JSON-RPC communication.
    """
    try:
        binary_path = get_mcp_binary_path()

        # Prepare arguments
        args = [binary_path] + sys.argv[1:]

        # Replace the current process with the MCP server binary
        # This is critical for MCP servers as they need direct stdio access
        # and proper signal handling without a Python wrapper layer
        if sys.platform == "win32":
            # Windows: use os.execv
            os.execv(binary_path, args)
        else:
            # Unix-like: use os.execv
            os.execv(binary_path, args)

    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        print(
            f"\nPlatform information:\n"
            f"  System: {platform.system()}\n"
            f"  Architecture: {platform.machine()}\n"
            f"  Python: {platform.python_version()}",
            file=sys.stderr
        )
        sys.exit(1)

    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
