"""Authentication service - Part of a 3-module circular dependency."""

import notification_service
import user_service


class AuthService:
    """Handles authentication and authorization."""

    def __init__(self):
        self.notifier = None
        self.user_service = None

    def validate_credentials(self, username, password):
        """Validate user credentials."""
        return len(username) > 0 and len(password) > 0

    def login(self, username, password):
        """Authenticate user and send notification."""
        if self.validate_credentials(username, password):
            if self.notifier is None:
                self.notifier = notification_service.NotificationService()
            self.notifier.send_login_notification(username)
            return True
        return False

    def get_user_permissions(self, user_id):
        """Get user permissions by querying user service."""
        if self.user_service is None:
            self.user_service = user_service.UserService()
        user = self.user_service.get_user_profile(user_id)
        return ["read", "write"] if user else []


class TokenManager:
    """Manages authentication tokens."""

    def generate_token(self, user_id):
        """Generate authentication token."""
        return f"token_{user_id}"
