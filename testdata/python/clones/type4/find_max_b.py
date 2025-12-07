# Type-4 Clone Example: Find maximum value (version B)
# Uses enumerate and ternary expressions


def find_maximum(data):
    """Find maximum value using enumerate."""
    if len(data) == 0:
        return None

    result = data[0]

    for i, item in enumerate(data):
        result = item if item > result else result

    return result


def find_min_max(nums):
    """Find both min and max using reduce-like pattern."""
    if not nums:
        return None, None

    lo = nums[0]
    hi = nums[0]

    for n in nums[1:]:
        lo = n if n < lo else lo
        hi = n if n > hi else hi

    return lo, hi
