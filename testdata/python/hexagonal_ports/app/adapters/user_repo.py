"""Adapter implementing the user port. May depend on ports and domain."""

from app.domain.user import User
from app.ports.user_port import UserPort


class UserRepository(UserPort):
    def __init__(self):
        self._store = {}

    def save(self, user: User) -> User:
        self._store[user.user_id] = user
        return user
