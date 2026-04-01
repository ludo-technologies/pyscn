"""Item domain model."""


class Item:
    """Core item entity."""

    def __init__(self, name: str):
        self.name = name

    def validate(self):
        return len(self.name) > 0
