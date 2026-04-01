"""User service - domain layer module that violates architecture rules.

This module intentionally imports from the presentation layer (pages),
which is forbidden by the architecture rules (domain can only depend on domain).
"""

from app.pages import user_page
from app.domain import user_model


class UserService:
    """Domain service with an architecture violation."""

    def get_user_view(self, user_id: int):
        user = user_model.User(user_id, "test")
        return user_page.render_user_profile(user)
