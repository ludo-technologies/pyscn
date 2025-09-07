"""
Main entry point for pyscn command-line interface.

This module provides a Python wrapper for the Go-implemented pyscn binary.
It automatically detects the platform and executes the appropriate binary.
"""

import os
import sys
import subprocess
import platform
from pathlib import Path


def get_binary_path() -> str:
    """
    Get the path to the pyscn binary for the current platform.
    
    Returns:
        str: Path to the pyscn binary.
        
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
    binary_name = f"pyscn-{system}-{machine}"
    if system == "windows":
        binary_name += ".exe"
    
    # Binary path within the package
    binary_path = Path(__file__).parent / "bin" / binary_name
    
    if not binary_path.exists():
        raise FileNotFoundError(
            f"pyscn binary not found for platform {system}-{machine}.\n"
            f"Expected location: {binary_path}\n"
            f"Please check that the package was installed correctly."
        )
    
    return str(binary_path)


def main():
    """
    Main entry point for pyscn CLI.
    
    Executes the Go-implemented pyscn binary with the provided arguments.
    """
    try:
        binary_path = get_binary_path()
        
        # Execute the binary with all arguments
        result = subprocess.run(
            [binary_path] + sys.argv[1:],
            capture_output=False  # Pass through stdout/stderr
        )
        
        # Exit with the same code as the binary
        sys.exit(result.returncode)
        
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
        
    except KeyboardInterrupt:
        # Handle Ctrl+C gracefully
        sys.exit(130)
        
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()