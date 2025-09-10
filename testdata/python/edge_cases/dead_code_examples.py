"""
Examples of dead code patterns for testing dead code detection.
This file contains various types of unreachable code.
"""


def unreachable_after_return():
    """Function with code after return statement."""
    x = 10
    if x > 5:
        return "greater than 5"
    
    # This code is unreachable - dead code
    print("This will never execute")
    return "fallback"


def unreachable_after_raise():
    """Function with code after raise statement."""
    value = None
    if value is None:
        raise ValueError("Value cannot be None")
    
    # This code is unreachable - dead code
    print("This will never execute")
    return value * 2


def unreachable_conditional_branch():
    """Function with unreachable conditional branches."""
    debug_mode = False
    
    if debug_mode:
        print("Debug mode enabled")
    else:
        print("Normal mode")
        
    # This condition will never be True - dead code
    if debug_mode and not debug_mode:
        print("This is impossible")
        return "impossible"
    
    return "normal"


def unreachable_after_break():
    """Function with code after break statement."""
    items = [1, 2, 3, 4, 5]
    
    for item in items:
        if item == 3:
            break
            # This code is unreachable - dead code
            print(f"Processing {item}")
            item += 1
    
    return "done"


def unreachable_after_continue():
    """Function with code after continue statement."""
    numbers = [1, 2, 3, 4, 5]
    result = []
    
    for num in numbers:
        if num % 2 == 0:
            continue
            # This code is unreachable - dead code
            print(f"Even number: {num}")
            result.append(num)
        result.append(num)
    
    return result


def multiple_returns_with_dead_code():
    """Function with multiple return statements and dead code."""
    status = "active"
    
    if status == "active":
        return "User is active"
    elif status == "inactive":
        return "User is inactive"
    else:
        return "Unknown status"
    
    # This code is unreachable - dead code
    print("This should never execute")
    status = "processed"
    return f"Status: {status}"


def nested_dead_code():
    """Function with nested dead code patterns."""
    data = {"valid": True}
    
    if data.get("valid"):
        if True:  # Always true condition
            return "valid data"
            # Dead code after return in nested block
            print("Never reached in nested block")
        
        # Dead code - the inner if always returns
        print("Never reached in outer block")
    
    # Dead code - the outer if always returns
    return "invalid data"


class DeadCodeClass:
    """Class with dead code examples."""
    
    def __init__(self):
        self.active = True
        return  # Early return in __init__
        # Dead code after return
        self.inactive = False
        self.setup()
    
    def setup(self):
        """This method is never called due to dead code in __init__."""
        print("Setting up...")
        return True
    
    def method_with_dead_code(self):
        """Method with unreachable code."""
        if self.active:
            return "active"
        
        # Dead code - method always returns above
        print("This is dead code")
        self.active = False
        return "inactive"


def infinite_loop_with_break():
    """Function with unreachable code after infinite loop."""
    counter = 0
    
    while True:
        counter += 1
        if counter > 10:
            break
    
    # This code is reachable - NOT dead code
    print(f"Counter reached: {counter}")
    
    while True:
        print("Infinite loop")
        return "exited"
    
    # This code is unreachable - dead code
    print("After infinite loop")
    return "never reached"


def early_exit_pattern():
    """Function with early exit patterns."""
    import sys
    
    condition = False
    if not condition:
        sys.exit(1)
    
    # This code is unreachable if condition is always False - dead code
    print("After sys.exit")
    return "completed"


if __name__ == "__main__":
    # Test the functions
    print(unreachable_after_return())
    print(unreachable_conditional_branch())
    # Note: unreachable_after_raise() would raise an exception