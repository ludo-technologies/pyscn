"""User page - renders user-related views."""

from app.domain import user_model


def render_user_profile(user):
    """Render user profile page."""
    return {"template": "user_profile.html", "context": {"user": user}}


def render_user_list(users: list):
    """Render user list page."""
    return {"template": "user_list.html", "context": {"users": users}}
