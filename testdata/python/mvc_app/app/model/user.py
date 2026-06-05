"""Model layer: data only, no dependency on view or controller."""


class User:
    def __init__(self, user_id: int, name: str):
        self.user_id = user_id
        self.name = name
