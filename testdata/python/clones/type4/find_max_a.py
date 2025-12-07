# Type-4 Clone Example: Find maximum value (version A)
# Uses explicit loop with index tracking


def find_maximum(items):
    """Find maximum value using explicit index tracking."""
    if not items:
        return None

    max_val = items[0]
    max_idx = 0

    for idx in range(1, len(items)):
        if items[idx] > max_val:
            max_val = items[idx]
            max_idx = idx

    return max_val


def find_min_max(values):
    """Find both min and max in one pass."""
    if len(values) == 0:
        return None, None

    minimum = values[0]
    maximum = values[0]

    for val in values:
        if val < minimum:
            minimum = val
        if val > maximum:
            maximum = val

    return minimum, maximum
