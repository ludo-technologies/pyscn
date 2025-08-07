# Simple function definitions

def simple_function():
    """A simple function with no parameters."""
    pass

def function_with_params(a, b, c):
    """Function with positional parameters."""
    return a + b + c

def function_with_defaults(x, y=10, z=20):
    """Function with default parameters."""
    return x + y + z

def function_with_args(*args):
    """Function with variable positional arguments."""
    return sum(args)

def function_with_kwargs(**kwargs):
    """Function with variable keyword arguments."""
    return kwargs

def function_with_all(a, b=5, *args, **kwargs):
    """Function with all parameter types."""
    result = a + b
    result += sum(args)
    result += sum(kwargs.values())
    return result

def function_with_annotations(x: int, y: float) -> float:
    """Function with type annotations."""
    return x + y

def nested_function():
    """Function with nested function definition."""
    def inner():
        return "inner"
    return inner()

def recursive_function(n):
    """Recursive function example."""
    if n <= 1:
        return n
    return recursive_function(n - 1) + recursive_function(n - 2)

# Lambda functions
simple_lambda = lambda: 42
param_lambda = lambda x, y: x + y
complex_lambda = lambda x, y=10: x * y if x > 0 else -y

# Async functions
async def async_function():
    """Async function example."""
    return "async result"

async def async_with_await(url):
    """Async function with await."""
    result = await fetch_data(url)
    return result

# Generator functions
def simple_generator():
    """Simple generator function."""
    yield 1
    yield 2
    yield 3

def generator_with_loop(n):
    """Generator with loop."""
    for i in range(n):
        yield i * i

def generator_with_condition(items):
    """Generator with condition."""
    for item in items:
        if item > 0:
            yield item