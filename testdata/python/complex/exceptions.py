# Exception handling patterns

# Basic exception handling
try:
    result = 10 / 2
except ZeroDivisionError:
    print("Cannot divide by zero")

# Multiple except blocks
try:
    value = int("not a number")
except ValueError:
    print("Invalid value")
except TypeError:
    print("Type error")
except Exception as e:
    print(f"Unexpected error: {e}")

# Except with multiple exceptions
try:
    risky_operation()
except (ValueError, TypeError, AttributeError) as e:
    print(f"Known error: {e}")

# Bare except
try:
    unknown_operation()
except:
    print("Something went wrong")

# Try with else
try:
    safe_value = 10
    result = safe_value * 2
except Exception:
    print("Error occurred")
else:
    print("No error, result is:", result)

# Try with finally
try:
    file = open("data.txt")
    data = file.read()
except FileNotFoundError:
    print("File not found")
finally:
    if 'file' in locals():
        file.close()

# Complete try block
try:
    result = complex_operation()
except ValueError as ve:
    print(f"Value error: {ve}")
except Exception as e:
    print(f"Other error: {e}")
else:
    print("Success:", result)
finally:
    cleanup()

# Nested try blocks
try:
    try:
        inner_operation()
    except ValueError:
        print("Inner value error")
        raise  # Re-raise the exception
except ValueError:
    print("Caught in outer block")

# Custom exceptions
class CustomError(Exception):
    """Custom exception class"""
    pass

class ValidationError(CustomError):
    """Validation error"""
    def __init__(self, field, value):
        self.field = field
        self.value = value
        super().__init__(f"Invalid {field}: {value}")

# Raising exceptions
def validate_age(age):
    if age < 0:
        raise ValueError("Age cannot be negative")
    if age > 150:
        raise ValueError("Age seems unrealistic")
    return age

# Exception chaining
try:
    try:
        dangerous_operation()
    except Exception as e:
        raise RuntimeError("Operation failed") from e
except RuntimeError as re:
    print(f"Runtime error: {re}")
    print(f"Caused by: {re.__cause__}")

# Suppressing exceptions
from contextlib import suppress

with suppress(FileNotFoundError):
    with open("maybe_missing.txt") as f:
        content = f.read()

# Exception in comprehension
safe_numbers = []
for s in ["1", "2", "abc", "4"]:
    try:
        safe_numbers.append(int(s))
    except ValueError:
        safe_numbers.append(0)

# Exception in generator
def safe_generator(items):
    for item in items:
        try:
            yield process_item(item)
        except Exception as e:
            print(f"Error processing {item}: {e}")
            yield None

# Exception in context manager
class ManagedResource:
    def __enter__(self):
        print("Acquiring resource")
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        print("Releasing resource")
        if exc_type is ValueError:
            print("Handling ValueError in context manager")
            return True  # Suppress the exception
        return False

# Using the context manager
with ManagedResource() as resource:
    raise ValueError("Test error")

# Exception in decorator
def exception_handler(default_return=None):
    def decorator(func):
        def wrapper(*args, **kwargs):
            try:
                return func(*args, **kwargs)
            except Exception as e:
                print(f"Exception in {func.__name__}: {e}")
                return default_return
        return wrapper
    return decorator

@exception_handler(default_return=0)
def risky_function(x, y):
    return x / y

# Async exception handling
async def async_exception_handling():
    try:
        result = await async_operation()
    except asyncio.TimeoutError:
        print("Async operation timed out")
    except Exception as e:
        print(f"Async error: {e}")
    finally:
        await cleanup_async()

# Exception groups (Python 3.11+)
def handle_multiple_errors():
    errors = []
    for operation in operations:
        try:
            operation()
        except Exception as e:
            errors.append(e)
    
    if errors:
        raise ExceptionGroup("Multiple errors occurred", errors)

# Assert with custom message
def check_invariant(value):
    assert value > 0, f"Value must be positive, got {value}"
    assert isinstance(value, int), f"Value must be integer, got {type(value)}"

# Warning handling
import warnings

def deprecated_function():
    warnings.warn("This function is deprecated", DeprecationWarning)
    return "old behavior"

# Catching warnings
with warnings.catch_warnings():
    warnings.simplefilter("ignore")
    deprecated_function()