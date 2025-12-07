# Type-4 Clone Example: Sum of numbers (iterative version)
# This is semantically equivalent to sum_recursive.py but uses iteration


def sum_numbers(numbers):
    """Calculate sum using a for loop."""
    total = 0
    for num in numbers:
        total = total + num
    return total


def sum_range(n):
    """Sum numbers from 1 to n using iteration."""
    result = 0
    i = 1
    while i <= n:
        result = result + i
        i = i + 1
    return result
