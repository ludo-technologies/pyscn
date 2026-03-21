"""User domain model."""


class User:
    """Core user entity."""

    def __init__(self, user_id: int, name: str):
        self.user_id = user_id
        self.name = name

    def is_active(self):
        return self.name != ""
