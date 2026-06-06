"""Controller mediating between model and view (allowed)."""

from app.model.user import User
from app.view import profile


def show(user_id: int, name: str) -> str:
    user = User(user_id, name)
    return profile.render(user)
