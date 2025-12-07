# Type-4 Clone Example: Filter and transform (version A)
# Uses traditional loop with accumulator


def filter_and_double(numbers, threshold):
    """Filter numbers above threshold and double them."""
    result = []
    for num in numbers:
        if num > threshold:
            doubled = num * 2
            result.append(doubled)
    return result


def process_data(items, min_value):
    """Process items: filter by min_value and square."""
    output = []
    for item in items:
        if item >= min_value:
            squared = item * item
            output.append(squared)
    return output


def count_matching(data, predicate_value):
    """Count items matching a condition."""
    count = 0
    for d in data:
        if d == predicate_value:
            count = count + 1
    return count
