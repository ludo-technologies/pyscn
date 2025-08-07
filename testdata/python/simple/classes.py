# Simple class definitions

class EmptyClass:
    """An empty class."""
    pass

class SimpleClass:
    """A simple class with methods."""
    
    def __init__(self):
        self.value = 0
    
    def get_value(self):
        return self.value
    
    def set_value(self, value):
        self.value = value

class ClassWithClassVar:
    """Class with class variables."""
    class_var = 100
    
    def __init__(self, instance_var):
        self.instance_var = instance_var
    
    @classmethod
    def get_class_var(cls):
        return cls.class_var
    
    @staticmethod
    def static_method(x, y):
        return x + y

class InheritedClass(SimpleClass):
    """Class with inheritance."""
    
    def __init__(self):
        super().__init__()
        self.extra = "extra"
    
    def get_extra(self):
        return self.extra

class MultipleInheritance(SimpleClass, EmptyClass):
    """Class with multiple inheritance."""
    pass

class ClassWithProperties:
    """Class with properties."""
    
    def __init__(self):
        self._value = 0
    
    @property
    def value(self):
        return self._value
    
    @value.setter
    def value(self, val):
        if val >= 0:
            self._value = val
    
    @value.deleter
    def value(self):
        self._value = 0

class ClassWithSlots:
    """Class with __slots__."""
    __slots__ = ['x', 'y']
    
    def __init__(self, x, y):
        self.x = x
        self.y = y

class ClassWithMagicMethods:
    """Class with magic methods."""
    
    def __init__(self, value):
        self.value = value
    
    def __str__(self):
        return f"Value: {self.value}"
    
    def __repr__(self):
        return f"ClassWithMagicMethods({self.value})"
    
    def __eq__(self, other):
        return self.value == other.value
    
    def __add__(self, other):
        return ClassWithMagicMethods(self.value + other.value)
    
    def __len__(self):
        return self.value
    
    def __getitem__(self, key):
        return self.value * key
    
    def __call__(self, x):
        return self.value * x

class AbstractClass:
    """Abstract class example."""
    
    def concrete_method(self):
        return "concrete"
    
    def abstract_method(self):
        raise NotImplementedError

class NestedClass:
    """Class with nested class."""
    
    class Inner:
        def inner_method(self):
            return "inner"
    
    def outer_method(self):
        return self.Inner().inner_method()

# Decorated class
@dataclass
class DataClass:
    name: str
    age: int
    email: str = ""

# Class with type annotations
class TypedClass:
    """Class with type annotations."""
    
    name: str
    age: int
    items: list[str]
    
    def __init__(self, name: str, age: int):
        self.name = name
        self.age = age
        self.items = []
    
    def add_item(self, item: str) -> None:
        self.items.append(item)
    
    def get_items(self) -> list[str]:
        return self.items