"""
pyqol - A next-generation Python static analysis tool.

This package provides a Python wrapper for the pyqol binary,
which is implemented in Go for high performance.
"""

from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version("pyqol")
except PackageNotFoundError:
    __version__ = "0.0.0"

__author__ = "pyqol team"
__email__ = "team@pyqol.dev"

from .main import main

__all__ = ["main"]