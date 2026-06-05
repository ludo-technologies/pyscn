"""Use case depending on an entity (allowed).

It also imports an interface adapter, which is an outward dependency the
clean preset forbids (use_cases -> interface_adapters).
"""

from app.entities.user import User
from app.interface_adapters.user_gateway import UserGateway


class CreateUser:
    def __init__(self):
        self._gateway = UserGateway()

    def execute(self, user_id: int, name: str) -> User:
        user = User(user_id, name)
        self._gateway.persist(user)
        return user
