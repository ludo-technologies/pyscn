"""Two functions, one of them with unreachable code."""


def with_dead_code(x):
    return x
    print("unreachable")


def without_dead_code(x):
    return x * 2
