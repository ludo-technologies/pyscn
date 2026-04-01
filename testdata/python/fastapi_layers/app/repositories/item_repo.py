"""Item repository - data access for items."""

from app.domain import item_model


class ItemRepository:
    """Handles item persistence."""

    def __init__(self):
        self._store = {}

    def find_by_id(self, item_id: int):
        return self._store.get(item_id)

    def save(self, item):
        self._store[id(item)] = item
        return item
