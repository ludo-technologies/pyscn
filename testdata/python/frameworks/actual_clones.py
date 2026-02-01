"""Test fixtures for ACTUAL clones within framework patterns.

These examples contain real code duplication that SHOULD be detected as clones,
even though they use framework decorators. The detection should correctly identify
cloned logic while properly weighting the framework boilerplate.
"""

from dataclasses import dataclass, field
from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime


# Type-1/Type-2 Clone: Nearly identical dataclasses with only name differences
# These SHOULD be detected as clones - they have identical logic

@dataclass
class UserMetricsV1:
    """First version of user metrics - actual clone of UserMetricsV2."""
    user_id: str
    login_count: int = 0
    last_login: Optional[datetime] = None
    session_duration_seconds: int = 0
    page_views: int = 0

    def calculate_engagement_score(self) -> float:
        """Calculate user engagement score."""
        if self.login_count == 0:
            return 0.0
        avg_session = self.session_duration_seconds / self.login_count
        avg_pages = self.page_views / self.login_count
        return (avg_session * 0.3 + avg_pages * 0.7) / 100.0

    def record_session(self, duration: int, pages: int) -> None:
        """Record a new session."""
        self.login_count += 1
        self.last_login = datetime.now()
        self.session_duration_seconds += duration
        self.page_views += pages


@dataclass
class UserMetricsV2:
    """Second version of user metrics - actual clone of UserMetricsV1."""
    account_id: str  # Only name changed
    login_count: int = 0
    last_login: Optional[datetime] = None
    session_duration_seconds: int = 0
    page_views: int = 0

    def calculate_engagement_score(self) -> float:
        """Calculate user engagement score."""
        if self.login_count == 0:
            return 0.0
        avg_session = self.session_duration_seconds / self.login_count
        avg_pages = self.page_views / self.login_count
        return (avg_session * 0.3 + avg_pages * 0.7) / 100.0

    def record_session(self, duration: int, pages: int) -> None:
        """Record a new session."""
        self.login_count += 1
        self.last_login = datetime.now()
        self.session_duration_seconds += duration
        self.page_views += pages


# Type-2/Type-3 Clone: Pydantic models with same validation logic
# These SHOULD be detected as clones - validators are copy-pasted

class InternalUserRequest(BaseModel):
    """Internal user request - clone of ExternalUserRequest."""
    employee_id: str = Field(..., min_length=8)
    full_name: str = Field(..., min_length=1, max_length=100)
    email_address: str = Field(..., pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    department: str = Field(..., min_length=1, max_length=50)
    access_level: int = Field(default=1, ge=1, le=5)

    def validate_and_normalize(self) -> dict:
        """Validate and return normalized data."""
        normalized = {
            "id": self.employee_id.strip().upper(),
            "name": self.full_name.strip().title(),
            "email": self.email_address.strip().lower(),
            "dept": self.department.strip().upper(),
            "level": max(1, min(5, self.access_level)),
        }
        if not normalized["email"].endswith(".com"):
            raise ValueError("Invalid email domain")
        return normalized

    def get_display_name(self) -> str:
        """Get formatted display name."""
        return f"{self.full_name} ({self.department})"


class ExternalUserRequest(BaseModel):
    """External user request - clone of InternalUserRequest."""
    customer_id: str = Field(..., min_length=8)
    full_name: str = Field(..., min_length=1, max_length=100)
    email_address: str = Field(..., pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    company: str = Field(..., min_length=1, max_length=50)
    tier_level: int = Field(default=1, ge=1, le=5)

    def validate_and_normalize(self) -> dict:
        """Validate and return normalized data."""
        normalized = {
            "id": self.customer_id.strip().upper(),
            "name": self.full_name.strip().title(),
            "email": self.email_address.strip().lower(),
            "dept": self.company.strip().upper(),
            "level": max(1, min(5, self.tier_level)),
        }
        if not normalized["email"].endswith(".com"):
            raise ValueError("Invalid email domain")
        return normalized

    def get_display_name(self) -> str:
        """Get formatted display name."""
        return f"{self.full_name} ({self.company})"


# Methods that are cloned across different framework classes
# The method bodies are identical - should be detected

@dataclass
class OrderProcessor:
    """Order processing with cloned validation logic."""
    order_id: str
    items: List[str] = field(default_factory=list)
    total_amount: float = 0.0
    discount_applied: float = 0.0

    def apply_discount(self, discount_percent: float) -> float:
        """Apply discount - CLONED logic in PaymentProcessor."""
        if discount_percent < 0:
            discount_percent = 0
        if discount_percent > 50:
            discount_percent = 50

        discount_amount = self.total_amount * (discount_percent / 100.0)
        self.discount_applied = discount_amount
        final_amount = self.total_amount - discount_amount

        if final_amount < 0:
            final_amount = 0

        return round(final_amount, 2)


@dataclass
class PaymentProcessor:
    """Payment processing with cloned validation logic."""
    payment_id: str
    line_items: List[str] = field(default_factory=list)
    subtotal: float = 0.0
    discount_value: float = 0.0

    def apply_discount(self, discount_percent: float) -> float:
        """Apply discount - CLONED logic from OrderProcessor."""
        if discount_percent < 0:
            discount_percent = 0
        if discount_percent > 50:
            discount_percent = 50

        discount_amount = self.subtotal * (discount_percent / 100.0)
        self.discount_value = discount_amount
        final_amount = self.subtotal - discount_amount

        if final_amount < 0:
            final_amount = 0

        return round(final_amount, 2)
