"""Test file for service locator anti-pattern."""


class Container:
    """Simple DI container for testing."""

    _services = {}

    @classmethod
    def register(cls, name, service):
        cls._services[name] = service

    @classmethod
    def get(cls, name):
        return cls._services.get(name)

    @classmethod
    def resolve(cls, name):
        return cls._services.get(name)


class Locator:
    """Service locator class."""

    @staticmethod
    def get_service(name):
        return None

    @staticmethod
    def locate(name):
        return None


def get_service(name):
    """Global service locator function."""
    return None


def resolve(name):
    """Another service locator function."""
    return None


class ServiceLocatorUser:
    """Class that uses service locator pattern - DI anti-pattern."""

    def __init__(self):
        # Using service locator in constructor
        self.logger = Container.get("logger")
        self.cache = Container.resolve("cache")

    def process(self):
        # Using service locator in methods
        db = Locator.get_service("database")
        return db


class GlobalLocatorUser:
    """Class that uses global locator functions - DI anti-pattern."""

    def __init__(self):
        self.config = get_service("config")

    def execute(self):
        return resolve("executor")


class ProperInjection:
    """Class with proper dependency injection - good practice."""

    def __init__(self, logger, cache, config):
        self.logger = logger
        self.cache = cache
        self.config = config

    def process(self, db):
        return db
