# Type-4 Clone Example: Sum of numbers (recursive version)
# This is semantically equivalent to sum_iterative.py but uses recursion


def sum_numbers(numbers):
    """Calculate sum using recursion."""
    if len(numbers) == 0:
        return 0
    return numbers[0] + sum_numbers(numbers[1:])


def sum_range(n):
    """Sum numbers from 1 to n using recursion."""
    if n <= 0:
        return 0
    return n + sum_range(n - 1)
