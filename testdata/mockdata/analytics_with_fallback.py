"""
Analytics module with problematic fallback data.
This file demonstrates mock data that should NOT be in production.
"""

import requests


def fetch_user_metrics():
    """Fetch user metrics from API with fallback."""
    try:
        response = requests.get("https://api.company.com/metrics")
        return response.json()
    except Exception:
        # BAD: Fallback to mock data - will silently use fake data if API fails
        return {
            "total_users": 1000,
            "active_users": 500,
            "email": "test@example.com",
            "phone": "000-0000-0000",
        }


def get_config():
    """Get configuration with hardcoded test values."""
    return {
        "api_endpoint": "http://localhost:8080/api",
        "debug_mode": True,
        "admin_email": "admin@test.org",
    }


# Hardcoded credentials - security risk
API_KEY = "password123"
DB_PASSWORD = "secret123"
AUTH_TOKEN = "testtoken"


class UserService:
    """Service with mock data embedded."""

    def __init__(self):
        # Placeholder UUID - will cause issues in production
        self.session_id = "00000000-0000-0000-0000-000000000000"
        self.mock_user = {"name": "Test User", "id": 12345}

    def get_default_user(self):
        """Return default user for testing."""
        return {
            "id": "11111111-1111-1111-1111-111111111111",
            "name": "John Doe",
            "email": "user@example.net",
        }


# Fake data for development
fake_response = {
    "status": "ok",
    "data": "lorem ipsum dolor sit amet",
}

dummy_config = {
    "foo": "bar",
    "baz": "qux",
}
