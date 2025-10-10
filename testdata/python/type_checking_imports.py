# Test file for TYPE_CHECKING imports
import os
import sys
from typing import TYPE_CHECKING

# Regular imports that should be included
from collections import defaultdict
import json

# TYPE_CHECKING imports that should be ignored for circular dependency detection
if TYPE_CHECKING:
    from typing import List, Dict, Optional
    from some.circular.dependency import CircularClass
    import circular_module

# Another TYPE_CHECKING block
if TYPE_CHECKING:
    from another.circular import AnotherCircular

# Regular code after TYPE_CHECKING
def some_function():
    pass

# Nested TYPE_CHECKING (should also be detected)
def some_other_function():
    if TYPE_CHECKING:
        from nested.circular import NestedCircular
    pass

# Not TYPE_CHECKING - should be included
if sys.version_info >= (3, 8):
    from new_feature import something

# Complex TYPE_CHECKING condition - should still be detected
if TYPE_CHECKING and sys.version_info >= (3, 9):
    from complex.circular import ComplexCircular