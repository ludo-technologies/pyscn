"""One heavily coupled class and one lightly coupled class.

Together with deps.py this gives a population where show_zeros, min_cbo and
max_cbo each drop a different subset from the displayed class list.
"""

from deps import DepA, DepB, DepC, DepD, DepE, DepF


class Hub:
    def build(self):
        return [DepA(), DepB(), DepC(), DepD(), DepE(), DepF()]


class Pair:
    def build(self):
        return DepA(), DepB()
