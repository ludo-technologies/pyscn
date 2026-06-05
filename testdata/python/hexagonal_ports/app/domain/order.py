"""Intentional DIP violation: a domain module importing an adapter.

The hexagonal preset forbids domain -> adapters, so this must be flagged.
"""

from app.adapters.user_repo import UserRepository


class Order:
    def __init__(self, order_id: int):
        self.order_id = order_id
        self._repo = UserRepository()
