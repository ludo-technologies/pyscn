"""Notification service - Part of a 3-module circular dependency."""

import user_service


class NotificationService:
    """Handles sending notifications to users."""

    def __init__(self):
        self.user_repo = None

    def send_welcome_email(self, user):
        """Send welcome email to new user."""
        email_content = f"Welcome {user['username']}!"
        self._send_email(user['email'], email_content)

    def send_login_notification(self, username):
        """Send notification when user logs in."""
        if self.user_repo is None:
            self.user_repo = user_service.UserRepository()
        user = self.user_repo.find_by_id(username)
        notification = f"Login detected for {user}"
        self._send_notification(notification)

    def _send_email(self, to_address, content):
        """Internal method to send email."""
        pass

    def _send_notification(self, message):
        """Internal method to send notification."""
        pass


class EmailTemplate:
    """Email template management."""

    def get_template(self, template_name):
        """Get email template."""
        return f"Template: {template_name}"
