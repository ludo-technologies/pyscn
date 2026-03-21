"""User repository - data access for users."""

from app.domain import user_model


class UserRepository:
    """Handles user persistence."""

    def __init__(self):
        self._store = {}

    def find_by_id(self, user_id: int):
        return self._store.get(user_id)

    def save(self, user):
        self._store[user.user_id] = user
        return user
