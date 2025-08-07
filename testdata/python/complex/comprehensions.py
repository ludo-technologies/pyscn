# Comprehension patterns

# List comprehensions
simple_list = [x for x in range(10)]
squared_list = [x**2 for x in range(10)]
filtered_list = [x for x in range(10) if x % 2 == 0]
conditional_list = [x if x % 2 == 0 else -x for x in range(10)]

# Nested list comprehensions
matrix = [[i*j for j in range(3)] for i in range(3)]
flattened = [item for row in matrix for item in row]

# Multiple conditions
complex_filter = [x for x in range(100) if x % 2 == 0 if x % 3 == 0]
and_condition = [x for x in range(100) if x % 2 == 0 and x % 3 == 0]

# Multiple iterables
pairs = [(x, y) for x in range(3) for y in range(3)]
combined = [x + y for x in [1, 2, 3] for y in [10, 20, 30]]

# Dictionary comprehensions
simple_dict = {x: x**2 for x in range(5)}
filtered_dict = {k: v for k, v in {'a': 1, 'b': 2, 'c': 3}.items() if v > 1}
swapped_dict = {v: k for k, v in {'a': 1, 'b': 2}.items()}

# Nested dictionary comprehension
nested_dict = {
    outer: {inner: outer * inner for inner in range(3)}
    for outer in range(3)
}

# Set comprehensions
simple_set = {x for x in range(10)}
unique_chars = {char for word in ['hello', 'world'] for char in word}
filtered_set = {x for x in range(100) if x % 7 == 0}

# Generator expressions
simple_gen = (x for x in range(10))
squared_gen = (x**2 for x in range(10))
filtered_gen = (x for x in range(100) if x % 2 == 0)

# Generator with multiple conditions
complex_gen = (
    x * y
    for x in range(10)
    if x > 5
    for y in range(10)
    if y < 5
)

# Using comprehensions with functions
def is_prime(n):
    if n < 2:
        return False
    for i in range(2, int(n**0.5) + 1):
        if n % i == 0:
            return False
    return True

primes = [x for x in range(100) if is_prime(x)]
prime_dict = {x: is_prime(x) for x in range(20)}

# Comprehensions with enumerate
indexed_list = [(i, val) for i, val in enumerate(['a', 'b', 'c'])]
indexed_dict = {i: val for i, val in enumerate(['x', 'y', 'z'])}

# Comprehensions with zip
zipped_list = [x + y for x, y in zip([1, 2, 3], [10, 20, 30])]
zipped_dict = {k: v for k, v in zip(['a', 'b', 'c'], [1, 2, 3])}

# Complex nested comprehensions
nested_complex = [
    [
        [k for k in range(i*j)]
        for j in range(1, 4)
    ]
    for i in range(1, 4)
]

# Comprehension with exception handling
safe_conversion = [
    int(x) if x.isdigit() else 0
    for x in ['1', '2', 'a', '4', 'b']
]

# Walrus operator in comprehensions (Python 3.8+)
filtered_computed = [
    y for x in range(10)
    if (y := x**2) > 25
]

# Comprehension with slicing
sliced_comp = [arr[i:i+3] for i in range(0, 10, 3) if (arr := list(range(10)))]

# Async comprehensions
async def async_range(n):
    for i in range(n):
        yield i

async def async_comprehensions():
    # Async list comprehension
    async_list = [x async for x in async_range(10)]
    
    # Async dict comprehension
    async_dict = {x: x**2 async for x in async_range(5)}
    
    # Async set comprehension
    async_set = {x async for x in async_range(10) if x % 2 == 0}
    
    # Async generator expression
    async_gen = (x**2 async for x in async_range(10))
    
    return async_list, async_dict, async_set, async_gen