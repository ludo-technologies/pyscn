"""Interface adapter. May depend on use_cases and entities."""

from app.entities.user import User


class UserGateway:
    def __init__(self):
        self._store = {}

    def persist(self, user: User) -> None:
        self._store[user.user_id] = user
