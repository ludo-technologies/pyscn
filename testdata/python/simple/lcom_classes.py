# Test fixtures for LCOM4 analysis

# LCOM4 = 1: All methods share 'self.value' → fully cohesive
class CohesiveClass:
    def __init__(self):
        self.value = 0

    def get_value(self):
        return self.value

    def set_value(self, v):
        self.value = v


# LCOM4 = 2: Two disconnected groups
# Group 1: __init__, get_x, set_x (share self.x)
# Group 2: get_y, set_y (share self.y)
# __init__ also touches self.y, so it connects both → actually LCOM4=1
# Let's make it truly disconnected:
class TwoGroupClass:
    def get_a(self):
        return self.a

    def set_a(self, v):
        self.a = v

    def get_b(self):
        return self.b

    def set_b(self, v):
        self.b = v


# LCOM4 = 3: Three disconnected groups
class ThreeGroupClass:
    def method_x(self):
        return self.x

    def method_y(self):
        return self.y

    def method_z(self):
        return self.z


# LCOM4 = 1: Class with @staticmethod and @classmethod (excluded)
class ClassWithDecorators:
    def __init__(self):
        self.data = []

    def add(self, item):
        self.data.append(item)

    @staticmethod
    def helper(x):
        return x * 2

    @classmethod
    def create(cls):
        return cls()


# LCOM4 = 1: Class with @property (included, accesses self._value)
class ClassWithProperty:
    def __init__(self):
        self._value = 0

    @property
    def value(self):
        return self._value

    def set_value(self, v):
        self._value = v


# LCOM4 = 1: Class with only one method (trivially cohesive)
class SingleMethodClass:
    def do_something(self):
        self.x = 1


# LCOM4 = 1: Empty class (trivially cohesive)
class EmptyClass:
    pass


# LCOM4 = 1: Methods with no self.xxx access form individual components
# But since there's only 0 or 1, it's trivially 1
class NoSelfAccessClass:
    def method_a(self):
        return 42

    def method_b(self):
        return 99


# LCOM4 = 1: All magic methods share self.value
class MagicMethodClass:
    def __init__(self, value):
        self.value = value

    def __str__(self):
        return str(self.value)

    def __repr__(self):
        return "MagicMethodClass(" + str(self.value) + ")"

    def __eq__(self, other):
        return self.value == other.value
