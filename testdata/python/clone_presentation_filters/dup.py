"""Two byte-identical functions plus one near-miss variant.

The pairs this produces land at different similarities, so min_similarity and
max_similarity each hide a different subset of them from the report.
"""


def process_alpha(items, factor):
    total = 0
    for item in items:
        if item is None:
            continue
        if item < 0:
            total -= item * factor
        else:
            total += item * factor
    if total > 100:
        return total / 2
    return total


def process_beta(items, factor):
    total = 0
    for item in items:
        if item is None:
            continue
        if item < 0:
            total -= item * factor
        else:
            total += item * factor
    if total > 100:
        return total / 2
    return total
