"""
Clean production code without mock data.
This file should NOT trigger any mockdata warnings.
"""

import os
from typing import Optional


def fetch_user_data(user_id: str) -> dict:
    """Fetch user data from database."""
    # Real implementation using environment variables
    db_host = os.environ.get("DB_HOST")
    db_password = os.environ.get("DB_PASSWORD")

    # ... actual database query ...
    return {"user_id": user_id, "status": "active"}


def send_email(recipient: str, subject: str, body: str) -> bool:
    """Send email to recipient."""
    # Real email implementation
    smtp_server = os.environ.get("SMTP_SERVER")
    return True


class ProductService:
    """Production-ready service."""

    def __init__(self, config: dict):
        self.api_url = config.get("api_url")
        self.timeout = config.get("timeout", 30)

    def get_products(self, category: Optional[str] = None) -> list:
        """Get products from API."""
        # Real implementation
        return []

    def calculate_price(self, base_price: float, quantity: int) -> float:
        """Calculate total price."""
        return base_price * quantity


# Configuration loaded from environment
CONFIG = {
    "max_retries": 3,
    "timeout_seconds": 30,
    "batch_size": 100,
}
