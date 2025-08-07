# Import statements

# Simple imports
import os
import sys
import math

# Multiple imports
import json, csv, sqlite3

# From imports
from datetime import datetime
from collections import defaultdict
from typing import List, Dict, Optional

# From import with alias
from collections import OrderedDict as OD
from numpy import array as np_array

# From import all
from math import *

# Relative imports
from . import module
from .. import parent_module
from ...package import submodule

from .sibling import function
from ..parent import ParentClass
from ...grandparent import CONSTANT

# Complex imports
from package.subpackage.module import (
    ClassA,
    ClassB,
    function_a,
    function_b as fb,
    CONSTANT_1,
    CONSTANT_2 as C2
)

# Import with __all__
__all__ = ['public_function', 'PublicClass']

def public_function():
    pass

class PublicClass:
    pass

def _private_function():
    pass

# Conditional imports
if sys.platform == "win32":
    import winreg
else:
    import pwd

# Try/except imports
try:
    import numpy as np
    HAS_NUMPY = True
except ImportError:
    HAS_NUMPY = False

# Function-level imports
def function_with_import():
    import tempfile
    return tempfile.mktemp()

# Class-level imports
class ClassWithImport:
    def method_with_import(self):
        from pathlib import Path
        return Path.home()