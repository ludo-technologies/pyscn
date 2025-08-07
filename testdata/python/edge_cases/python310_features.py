# Python 3.10+ features for testing modern syntax

# Match statements (Pattern Matching)
def process_command(command):
    match command:
        case "quit" | "exit":
            return "Goodbye"
        case "help":
            return "Available commands: help, quit, exit"
        case ["move", direction]:
            return f"Moving {direction}"
        case ["attack", target, weapon]:
            return f"Attacking {target} with {weapon}"
        case {"action": "jump", "height": h}:
            return f"Jumping {h} meters"
        case {"x": x, "y": y, **rest}:
            return f"Position: ({x}, {y}), extras: {rest}"
        case _:
            return "Unknown command"

# Match with guards
def categorize_number(num):
    match num:
        case 0:
            return "zero"
        case n if n > 0:
            return "positive"
        case n if n < 0:
            return "negative"

# Match with as patterns
def process_data(data):
    match data:
        case [x, y] as pair:
            return f"Pair: {pair}"
        case [x, y, *rest] as list_data:
            return f"List with {len(list_data)} items"
        case str() as text:
            return f"String: {text}"
        case int() | float() as number:
            return f"Number: {number}"

# Match with class patterns
class Point:
    def __init__(self, x, y):
        self.x = x
        self.y = y

def process_point(point):
    match point:
        case Point(x=0, y=0):
            return "Origin"
        case Point(x=0, y=y):
            return f"On Y-axis at {y}"
        case Point(x=x, y=0):
            return f"On X-axis at {x}"
        case Point(x=x, y=y) if x == y:
            return f"On diagonal at {x}"
        case Point(x=x, y=y):
            return f"Point at ({x}, {y})"

# Union types (Type Hints)
from typing import Union

def process_value(value: int | str | None) -> str | None:
    if value is None:
        return None
    return str(value)

# Type alias with union
JsonValue = None | bool | int | float | str | list["JsonValue"] | dict[str, "JsonValue"]

# Parenthesized context managers
with (
    open("file1.txt") as f1,
    open("file2.txt") as f2,
    open("file3.txt") as f3
):
    content = f1.read() + f2.read() + f3.read()

# Better error messages (syntax improvement)
def complex_function(
    param1: int,
    param2: str,
    /,  # Positional-only parameters
    param3: float = 0.0,
    *,  # Keyword-only parameters
    param4: bool = False,
    param5: list[int] | None = None,
) -> tuple[int, str]:
    return (param1, param2)

# Structural pattern matching with sequences
def analyze_sequence(seq):
    match seq:
        case []:
            return "Empty"
        case [x]:
            return f"Single: {x}"
        case [x, y]:
            return f"Pair: {x}, {y}"
        case [x, *middle, y]:
            return f"First: {x}, Last: {y}, Middle: {middle}"

# Match with nested patterns
def process_nested(data):
    match data:
        case {"user": {"name": name, "age": age}} if age >= 18:
            return f"Adult user: {name}"
        case {"user": {"name": name, "age": age}}:
            return f"Minor user: {name}"
        case {"error": {"code": code, "message": msg}}:
            return f"Error {code}: {msg}"

# Match with or patterns
def classify_value(value):
    match value:
        case 0 | 1 | 2:
            return "Small"
        case int() if 3 <= value <= 10:
            return "Medium"
        case int() if value > 10:
            return "Large"
        case str() | bytes():
            return "Text"
        case list() | tuple() | set():
            return "Collection"
        case _:
            return "Other"

# Walrus operator in match
def find_pattern(text):
    import re
    match text:
        case s if (m := re.match(r"(\d+)", s)):
            return f"Number: {m.group(1)}"
        case s if (m := re.match(r"([a-z]+)", s, re.I)):
            return f"Word: {m.group(1)}"
        case _:
            return "No pattern"

# Complex match with multiple features
def advanced_match(obj):
    match obj:
        case [Point(x=0, y=y1), Point(x=0, y=y2)]:
            return f"Vertical line from {y1} to {y2}"
        case [Point(x=x1, y=0), Point(x=x2, y=0)]:
            return f"Horizontal line from {x1} to {x2}"
        case [Point(x=x1, y=y1), Point(x=x2, y=y2)] if x1 == x2:
            return f"Vertical line at x={x1}"
        case [Point(x=x1, y=y1), Point(x=x2, y=y2)] if y1 == y2:
            return f"Horizontal line at y={y1}"
        case [Point() as p1, Point() as p2, *rest]:
            return f"Polyline with {len(rest) + 2} points"

# Type annotations with new syntax
from typing import TypeAlias, TypeVar, Generic

T = TypeVar("T")
NumberList: TypeAlias = list[int | float]

class Container(Generic[T]):
    def __init__(self, value: T) -> None:
        self.value = value
    
    def get(self) -> T:
        return self.value

# String formatting improvements
name = "Python"
version = 3.10
formatted = f"{name} {version:.1f}"
debug_formatted = f"{name=} {version=}"