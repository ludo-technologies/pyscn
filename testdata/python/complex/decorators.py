# Decorator patterns

import functools
from typing import Any, Callable

# Simple decorator
def simple_decorator(func):
    def wrapper(*args, **kwargs):
        print(f"Before {func.__name__}")
        result = func(*args, **kwargs)
        print(f"After {func.__name__}")
        return result
    return wrapper

@simple_decorator
def decorated_function():
    print("Function body")

# Decorator with arguments
def repeat(times):
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            for _ in range(times):
                result = func(*args, **kwargs)
            return result
        return wrapper
    return decorator

@repeat(3)
def say_hello():
    print("Hello!")

# Multiple decorators
def bold(func):
    def wrapper(*args, **kwargs):
        return f"<b>{func(*args, **kwargs)}</b>"
    return wrapper

def italic(func):
    def wrapper(*args, **kwargs):
        return f"<i>{func(*args, **kwargs)}</i>"
    return wrapper

@bold
@italic
def formatted_text(text):
    return text

# Class decorator
def add_repr(cls):
    def __repr__(self):
        return f"{cls.__name__}({self.__dict__})"
    cls.__repr__ = __repr__
    return cls

@add_repr
class DecoratedClass:
    def __init__(self, x, y):
        self.x = x
        self.y = y

# Property decorator
class PropertyExample:
    def __init__(self):
        self._value = 0
    
    @property
    def value(self):
        return self._value
    
    @value.setter
    def value(self, val):
        self._value = val
    
    @value.deleter
    def value(self):
        self._value = 0

# Cached property
class CachedPropertyExample:
    def __init__(self):
        self._cache = {}
    
    @functools.cached_property
    def expensive_computation(self):
        print("Computing...")
        return sum(range(1000000))

# Method decorators
class MethodDecorators:
    class_var = "class"
    
    @classmethod
    def class_method(cls):
        return cls.class_var
    
    @staticmethod
    def static_method(x, y):
        return x + y
    
    @functools.lru_cache(maxsize=128)
    def cached_method(self, n):
        if n <= 1:
            return n
        return self.cached_method(n-1) + self.cached_method(n-2)

# Decorator class
class CountCalls:
    def __init__(self, func):
        self.func = func
        self.count = 0
    
    def __call__(self, *args, **kwargs):
        self.count += 1
        print(f"Call {self.count} to {self.func.__name__}")
        return self.func(*args, **kwargs)

@CountCalls
def counted_function():
    return "result"

# Parameterized decorator class
class MaxCalls:
    def __init__(self, max_calls=3):
        self.max_calls = max_calls
    
    def __call__(self, func):
        self.calls = 0
        
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            if self.calls >= self.max_calls:
                raise Exception(f"Max calls ({self.max_calls}) exceeded")
            self.calls += 1
            return func(*args, **kwargs)
        
        return wrapper

@MaxCalls(5)
def limited_function():
    return "called"

# Context manager decorator
from contextlib import contextmanager

@contextmanager
def my_context():
    print("Entering context")
    try:
        yield "context value"
    finally:
        print("Exiting context")

# Async decorator
def async_decorator(func):
    async def wrapper(*args, **kwargs):
        print(f"Before async {func.__name__}")
        result = await func(*args, **kwargs)
        print(f"After async {func.__name__}")
        return result
    return wrapper

@async_decorator
async def async_function():
    return "async result"

# Decorator factory
def validate_type(expected_type):
    def decorator(func):
        @functools.wraps(func)
        def wrapper(value):
            if not isinstance(value, expected_type):
                raise TypeError(f"Expected {expected_type}, got {type(value)}")
            return func(value)
        return wrapper
    return decorator

@validate_type(int)
def process_integer(value):
    return value * 2

@validate_type(str)
def process_string(value):
    return value.upper()