"""Port (interface) for user persistence. Lives on the domain side."""

from abc import ABC, abstractmethod

from app.domain.user import User


class UserPort(ABC):
    @abstractmethod
    def save(self, user: User) -> User:
        ...
