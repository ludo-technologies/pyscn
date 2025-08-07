# Control flow statements

# If statements
x = 10

if x > 0:
    print("positive")

if x > 0:
    print("positive")
else:
    print("non-positive")

if x > 10:
    print("greater than 10")
elif x == 10:
    print("equal to 10")
else:
    print("less than 10")

if x > 5:
    if x < 15:
        print("between 5 and 15")
    else:
        print("15 or greater")
else:
    print("5 or less")

# Conditional expressions
result = "positive" if x > 0 else "non-positive"
nested_conditional = "big" if x > 100 else "medium" if x > 10 else "small"

# For loops
for i in range(10):
    print(i)

for i in range(0, 10, 2):
    print(i)

for item in [1, 2, 3, 4, 5]:
    if item == 3:
        continue
    if item == 5:
        break
    print(item)

for i, value in enumerate(['a', 'b', 'c']):
    print(i, value)

for key, value in {'a': 1, 'b': 2}.items():
    print(key, value)

# For with else
for i in range(5):
    if i == 10:
        break
else:
    print("completed normally")

# Nested loops
for i in range(3):
    for j in range(3):
        print(i, j)

# While loops
count = 0
while count < 5:
    print(count)
    count += 1

while True:
    if count > 10:
        break
    count += 1

# While with else
x = 0
while x < 5:
    x += 1
else:
    print("while completed")

# Try/except statements
try:
    result = 10 / 2
except ZeroDivisionError:
    print("division by zero")

try:
    value = int("abc")
except ValueError as e:
    print(f"error: {e}")
except Exception:
    print("unexpected error")

try:
    result = 10 / 1
except:
    print("error")
else:
    print("success")
finally:
    print("cleanup")

# Nested try
try:
    try:
        raise ValueError("inner")
    except ValueError:
        raise
except:
    print("caught in outer")

# With statements
with open("file.txt") as f:
    content = f.read()

with open("file1.txt") as f1, open("file2.txt") as f2:
    data1 = f1.read()
    data2 = f2.read()

# Assert statements
assert x > 0
assert x == 10, "x should be 10"

# Pass statement
if x > 100:
    pass
else:
    print("not greater than 100")

# Match statement (Python 3.10+)
def process_value(value):
    match value:
        case 0:
            return "zero"
        case 1 | 2 | 3:
            return "small"
        case int(x) if x > 10:
            return "large"
        case [x, y]:
            return f"pair: {x}, {y}"
        case {"key": value}:
            return f"dict with key: {value}"
        case _:
            return "other"