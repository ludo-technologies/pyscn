"""
pyqol - A next-generation Python static analysis tool.

This package provides a Python wrapper for the pyqol binary,
which is implemented in Go for high performance.
"""

__version__ = "0.1.0"
__author__ = "pyqol team"
__email__ = "team@pyqol.dev"

from .main import main

__all__ = ["main"]