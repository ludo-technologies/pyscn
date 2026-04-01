"""Test fixtures for dataclass patterns that should NOT be detected as clones.

These dataclasses have similar structure (all decorated with @dataclass, all have
type-annotated fields) but represent completely different domain concepts with
different semantics. Clone detection should NOT report these as clones.
"""

from dataclasses import dataclass, field
from typing import Optional, List
from datetime import datetime


@dataclass
class UserProfile:
    """Represents a user's profile information."""
    username: str
    email: str
    display_name: Optional[str] = None
    created_at: datetime = field(default_factory=datetime.now)
    is_active: bool = True

    def validate_email(self) -> bool:
        """Validate the email format."""
        return "@" in self.email and "." in self.email.split("@")[1]

    def deactivate(self) -> None:
        """Deactivate the user profile."""
        self.is_active = False


@dataclass
class ProductInventory:
    """Represents product inventory data."""
    product_id: str
    sku: str
    quantity: int = 0
    reorder_threshold: int = 10
    warehouse_location: Optional[str] = None

    def needs_reorder(self) -> bool:
        """Check if product needs to be reordered."""
        return self.quantity <= self.reorder_threshold

    def update_quantity(self, delta: int) -> None:
        """Update the inventory quantity."""
        self.quantity += delta
        if self.quantity < 0:
            self.quantity = 0


@dataclass
class OrderItem:
    """Represents an item in an order."""
    item_id: str
    product_name: str
    unit_price: float
    quantity: int = 1
    discount_percent: float = 0.0

    def calculate_total(self) -> float:
        """Calculate the total price for this item."""
        subtotal = self.unit_price * self.quantity
        discount = subtotal * (self.discount_percent / 100)
        return subtotal - discount

    def apply_discount(self, percent: float) -> None:
        """Apply a discount to this item."""
        self.discount_percent = min(percent, 100.0)


@dataclass
class ServerConfig:
    """Server configuration settings."""
    host: str
    port: int
    debug_mode: bool = False
    max_connections: int = 100
    timeout_seconds: float = 30.0
    ssl_enabled: bool = True

    def get_connection_string(self) -> str:
        """Generate the connection string."""
        protocol = "https" if self.ssl_enabled else "http"
        return f"{protocol}://{self.host}:{self.port}"

    def enable_debug(self) -> None:
        """Enable debug mode with adjusted settings."""
        self.debug_mode = True
        self.timeout_seconds = 60.0


@dataclass
class LogEntry:
    """Represents a log entry."""
    timestamp: datetime
    level: str
    message: str
    source: Optional[str] = None
    metadata: dict = field(default_factory=dict)

    def is_error(self) -> bool:
        """Check if this is an error log."""
        return self.level.upper() in ("ERROR", "CRITICAL", "FATAL")

    def format_entry(self) -> str:
        """Format the log entry as a string."""
        source_str = f"[{self.source}]" if self.source else ""
        return f"{self.timestamp.isoformat()} {self.level} {source_str} {self.message}"
