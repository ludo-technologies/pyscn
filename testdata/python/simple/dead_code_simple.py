"""
Simple examples of dead code for testing.
"""


def simple_dead_code():
    """Simple function with dead code after return."""
    name = "test"
    return f"Hello, {name}"
    
    # Dead code - unreachable
    print("This line will never execute")
    name = "updated"


def conditional_dead_code():
    """Function with dead code in conditional."""
    always_true = True
    
    if always_true:
        return "always returns here"
    
    # Dead code - condition is always true
    print("This is unreachable")
    return "never reached"


def loop_with_dead_code():
    """Function with dead code after loop break."""
    for i in range(5):
        if i == 2:
            break
            # Dead code after break
            print(f"Processing {i}")
        print(f"Number: {i}")
    
    return "loop completed"


def exception_dead_code():
    """Function with dead code after exception."""
    data = None
    
    if data is None:
        raise ValueError("Data is required")
    
    # Dead code - exception always raised
    print("Processing data")
    return len(data)


def nested_return_dead_code():
    """Function with nested returns and dead code."""
    status = "ok"
    
    if status == "ok":
        if True:
            return "success"
            # Dead code in nested block
            print("Never reached nested")
        
        # Dead code in outer block
        print("Never reached outer")
    
    # Dead code at function level
    return "failure"