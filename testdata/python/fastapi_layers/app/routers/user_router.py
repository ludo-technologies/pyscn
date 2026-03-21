"""User router - handles HTTP endpoints for user operations."""

from app.domain import user_model
from app.pages import user_page


class UserRouter:
    """Routes user-related HTTP requests."""

    def get_user(self, user_id: int):
        user = user_model.User(user_id, "test")
        return user_page.render_user_profile(user)

    def list_users(self):
        return {"users": []}
