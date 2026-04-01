# This imports SomeClass from pkg_a (re-exported from pkg_a.module_x)
# This creates a cycle: pkg_a.module_x -> pkg_b.file -> pkg_a -> pkg_a.module_x
from pkg_a import SomeClass


class OtherClass:
    """A class that depends on SomeClass"""

    def use_some_class(self):
        return SomeClass()
