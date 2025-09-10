"""
Complex dead code patterns for advanced testing.
"""

import sys
from typing import Optional, Union


def complex_control_flow():
    """Complex function with multiple dead code patterns."""
    state = "initial"
    result = []
    
    try:
        if state == "initial":
            state = "processing"
            
            for i in range(10):
                if i > 5:
                    raise StopIteration("Processing complete")
                result.append(i)
            
            # Dead code - loop always raises exception
            print("Loop completed normally")
            state = "completed"
            
    except StopIteration:
        return result
        # Dead code after return in exception handler
        print("Exception handled")
        state = "error"
    
    # Dead code - function always returns in exception handler
    print("Function end")
    return []


def generator_with_dead_code():
    """Generator function with dead code."""
    yield 1
    yield 2
    return  # Early return in generator
    
    # Dead code after return
    yield 3
    yield 4


async def async_dead_code():
    """Async function with dead code."""
    import asyncio
    
    await asyncio.sleep(0.1)
    return "async result"
    
    # Dead code after return
    await asyncio.sleep(1.0)
    return "never reached"


def context_manager_dead_code():
    """Function with dead code in context manager."""
    class CustomContext:
        def __enter__(self):
            return self
        
        def __exit__(self, exc_type, exc_val, exc_tb):
            if exc_type:
                return True  # Suppress exception
    
    with CustomContext():
        raise ValueError("Test exception")
        # Dead code after raise
        print("This won't execute")
    
    # This code is reachable due to exception suppression
    return "context completed"


def decorator_dead_code():
    """Function with dead code in decorator pattern."""
    def early_return_decorator(func):
        def wrapper(*args, **kwargs):
            return "decorator result"
            # Dead code in decorator
            result = func(*args, **kwargs)
            return f"decorated: {result}"
        return wrapper
    
    @early_return_decorator
    def decorated_function():
        print("Original function")
        return "original result"
    
    return decorated_function()


def class_method_dead_code():
    """Function demonstrating dead code in class methods."""
    
    class TestClass:
        def __init__(self):
            self.value = 10
            return  # Early return in __init__
            # Dead code
            self.other_value = 20
            self._setup()
        
        def _setup(self):
            """This method is never called."""
            self.configured = True
        
        def method_with_dead_code(self):
            if self.value > 0:
                return self.value * 2
            
            # Dead code - value is always > 0
            print("Negative value handling")
            return 0
    
    obj = TestClass()
    return obj.method_with_dead_code()


def lambda_dead_code():
    """Function with dead code involving lambdas."""
    always_true = True
    
    if always_true:
        func = lambda x: x * 2
        return func(5)
    
    # Dead code - condition always true
    func = lambda x: x * 3
    return func(5)


def comprehension_dead_code():
    """Function with dead code in comprehensions."""
    data = [1, 2, 3, 4, 5]
    
    if len(data) > 0:
        return [x * 2 for x in data if x > 0]
    
    # Dead code - data always has length > 0
    return [x for x in data if x < 0]


def exception_hierarchy_dead_code():
    """Function with dead code in exception hierarchy."""
    try:
        raise ValueError("Test error")
    except Exception:
        return "caught general exception"
    except ValueError:
        # Dead code - ValueError is already caught by Exception
        return "caught specific exception"
    
    # Dead code - exception always caught
    return "no exception"


def finally_block_patterns():
    """Function demonstrating dead code with finally blocks."""
    try:
        return "try block result"
        # Dead code after return in try block
        print("After return in try")
    except Exception:
        return "exception result"
        # Dead code after return in except block
        print("After return in except")
    finally:
        # This code is reachable - finally always executes
        print("Finally block executes")
        return "finally result"  # This overrides other returns
    
    # Dead code after try-except-finally
    return "end of function"


if __name__ == "__main__":
    print(complex_control_flow())
    print(class_method_dead_code())
    print(comprehension_dead_code())