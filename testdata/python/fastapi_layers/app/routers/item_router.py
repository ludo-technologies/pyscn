"""Item router - handles HTTP endpoints for item operations."""

from app.domain import item_model
from app.repositories import item_repo


class ItemRouter:
    """Routes item-related HTTP requests."""

    def __init__(self):
        self.repo = item_repo.ItemRepository()

    def get_item(self, item_id: int):
        return self.repo.find_by_id(item_id)

    def create_item(self, name: str):
        item = item_model.Item(name)
        return self.repo.save(item)
