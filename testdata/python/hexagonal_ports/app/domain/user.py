"""Core domain entity. Must not depend on ports or adapters."""


class User:
    def __init__(self, user_id: int, name: str):
        self.user_id = user_id
        self.name = name
