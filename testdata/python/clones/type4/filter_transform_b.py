# Type-4 Clone Example: Filter and transform (version B)
# Uses list comprehension style (unrolled)


def filter_and_double(values, limit):
    """Filter values above limit and double them."""
    filtered = []
    transformed = []

    # First pass: filter
    for v in values:
        if v > limit:
            filtered.append(v)

    # Second pass: transform
    for f in filtered:
        transformed.append(f * 2)

    return transformed


def process_data(elements, cutoff):
    """Process elements: filter by cutoff and square."""
    above_cutoff = []
    squared_values = []

    for e in elements:
        if e >= cutoff:
            above_cutoff.append(e)

    for a in above_cutoff:
        squared_values.append(a * a)

    return squared_values


def count_matching(collection, target):
    """Count items matching target."""
    matches = []
    for c in collection:
        if c == target:
            matches.append(c)
    return len(matches)
