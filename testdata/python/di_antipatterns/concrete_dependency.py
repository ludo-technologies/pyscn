"""Test file for concrete dependency anti-patterns."""

from abc import ABC, abstractmethod
from typing import Protocol


# Abstract types (good)
class UserRepositoryInterface(ABC):
    @abstractmethod
    def get_user(self, user_id: int):
        pass


class LoggerProtocol(Protocol):
    def log(self, message: str) -> None:
        ...


# Concrete types (potentially problematic when injected)
class MySQLUserRepository:
    """Concrete implementation of user repository."""

    def get_user(self, user_id: int):
        return {"id": user_id}


class FileLogger:
    """Concrete logger implementation."""

    def log(self, message: str):
        print(message)


class ConcreteTypeHintService:
    """Service with concrete type hints - DI anti-pattern (info severity)."""

    def __init__(self, repo: MySQLUserRepository, logger: FileLogger):
        self.repo = repo
        self.logger = logger


class AbstractTypeHintService:
    """Service with abstract type hints - good practice."""

    def __init__(self, repo: UserRepositoryInterface, logger: LoggerProtocol):
        self.repo = repo
        self.logger = logger


class DirectInstantiationService:
    """Service that instantiates dependencies in constructor - DI anti-pattern."""

    def __init__(self):
        self.repo = MySQLUserRepository()
        self.logger = FileLogger()


class PartialDirectInstantiation:
    """Service with some injected and some instantiated dependencies."""

    def __init__(self, cache):
        self.cache = cache
        self.logger = FileLogger()  # Anti-pattern


class GoodService:
    """Service with all dependencies properly injected."""

    def __init__(self, repo: UserRepositoryInterface, logger: LoggerProtocol, cache):
        self.repo = repo
        self.logger = logger
        self.cache = cache
