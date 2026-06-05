"""View accessing the model directly.

The mvc preset discourages view -> model; this should produce a warning.
"""

from app.model.user import User


def render(user: User) -> str:
    return f"<h1>{user.name}</h1>"
