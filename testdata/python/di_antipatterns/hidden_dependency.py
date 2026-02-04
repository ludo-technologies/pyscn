"""Test file for hidden dependency anti-patterns."""

# Module-level variable (can be accessed from classes)
_config = {"debug": True}
db_connection = None
settings = {}


class GlobalStatementUser:
    """Class that uses global statement - DI anti-pattern."""

    def __init__(self):
        pass

    def do_something(self):
        global _config
        return _config["debug"]


class ModuleVariableAccessor:
    """Class that directly accesses module-level variables - DI anti-pattern."""

    def __init__(self):
        pass

    def get_connection(self):
        return db_connection

    def read_settings(self):
        return settings


class SingletonExample:
    """Class implementing singleton pattern - DI anti-pattern."""
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance


class AnotherSingleton:
    """Another singleton implementation."""
    __instance = None

    @classmethod
    def get_instance(cls):
        if cls.__instance is None:
            cls.__instance = cls()
        return cls.__instance


class GoodClass:
    """A class with proper dependency injection."""

    def __init__(self, config, connection):
        self._config = config
        self._connection = connection

    def get_debug_mode(self):
        return self._config["debug"]

    def get_connection(self):
        return self._connection
