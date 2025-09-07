"""
pyscn - A next-generation Python static analysis tool.

This package provides a Python wrapper for the pyscn binary,
which is implemented in Go for high performance.
"""

from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("pyscn")
except PackageNotFoundError:
    __version__ = "0.0.0"

__author__ = "pyscn team"
__email__ = "team@pyscn.dev"

from .main import main

__all__ = ["main"]