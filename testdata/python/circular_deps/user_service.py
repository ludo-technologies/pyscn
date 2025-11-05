"""User service module - Part of a 3-module circular dependency."""

import auth_service
import notification_service


class UserService:
    """Manages user operations."""

    def __init__(self):
        self.auth = auth_service.AuthService()
        self.notifier = notification_service.NotificationService()

    def create_user(self, username, email):
        """Create a new user."""
        if self.auth.validate_credentials(username, email):
            user = {"username": username, "email": email}
            self.notifier.send_welcome_email(user)
            return user
        return None

    def get_user_profile(self, user_id):
        """Get user profile data."""
        return {"id": user_id, "name": "User"}


class UserRepository:
    """Repository for user data access."""

    def find_by_id(self, user_id):
        """Find user by ID."""
        return {"id": user_id}
