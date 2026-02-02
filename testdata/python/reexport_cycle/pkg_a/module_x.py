# This module imports from pkg_b, creating a potential cycle
from pkg_b.file import OtherClass


class SomeClass:
    """A class that is re-exported from __init__.py"""

    def __init__(self):
        self.other = OtherClass()
