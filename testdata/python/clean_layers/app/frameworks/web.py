"""Outermost layer: frameworks may depend on any inner layer."""

from app.interface_adapters.user_gateway import UserGateway
from app.use_cases.create_user import CreateUser


def handle_request(user_id: int, name: str):
    _ = UserGateway()
    return CreateUser().execute(user_id, name)
