"""Test fixtures for Pydantic model patterns that should NOT be detected as clones.

These Pydantic models have similar structure (all inherit from BaseModel, all have
Field() declarations) but represent completely different domain concepts. Clone
detection should NOT report these as clones.
"""

from pydantic import BaseModel, Field, validator
from typing import Optional, List
from datetime import datetime
from enum import Enum


class PaymentStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    FAILED = "failed"
    REFUNDED = "refunded"


class CustomerAddress(BaseModel):
    """Customer shipping/billing address."""
    street_line1: str = Field(..., min_length=1, max_length=200)
    street_line2: Optional[str] = Field(None, max_length=200)
    city: str = Field(..., min_length=1, max_length=100)
    state: str = Field(..., min_length=2, max_length=50)
    postal_code: str = Field(..., pattern=r"^\d{5}(-\d{4})?$")
    country: str = Field(default="US", min_length=2, max_length=2)

    @validator("country")
    def validate_country(cls, v):
        """Validate country code is uppercase."""
        return v.upper()

    def format_address(self) -> str:
        """Format address as a multi-line string."""
        lines = [self.street_line1]
        if self.street_line2:
            lines.append(self.street_line2)
        lines.append(f"{self.city}, {self.state} {self.postal_code}")
        lines.append(self.country)
        return "\n".join(lines)


class PaymentTransaction(BaseModel):
    """Payment transaction record."""
    transaction_id: str = Field(..., min_length=10, max_length=50)
    amount_cents: int = Field(..., gt=0)
    currency: str = Field(default="USD", min_length=3, max_length=3)
    status: PaymentStatus = Field(default=PaymentStatus.PENDING)
    processor_reference: Optional[str] = Field(None, max_length=100)
    created_at: datetime = Field(default_factory=datetime.utcnow)

    @validator("currency")
    def validate_currency(cls, v):
        """Validate currency code is uppercase."""
        return v.upper()

    def get_amount_dollars(self) -> float:
        """Convert amount to dollars."""
        return self.amount_cents / 100.0

    def mark_completed(self, reference: str) -> None:
        """Mark transaction as completed."""
        self.status = PaymentStatus.COMPLETED
        self.processor_reference = reference


class ProductReview(BaseModel):
    """Customer product review."""
    review_id: str = Field(..., min_length=8)
    product_id: str = Field(..., min_length=8)
    customer_id: str = Field(..., min_length=8)
    rating: int = Field(..., ge=1, le=5)
    title: str = Field(..., min_length=1, max_length=100)
    content: str = Field(..., min_length=10, max_length=5000)
    verified_purchase: bool = Field(default=False)
    helpful_votes: int = Field(default=0, ge=0)
    created_at: datetime = Field(default_factory=datetime.utcnow)

    @validator("content")
    def strip_content(cls, v):
        """Strip whitespace from content."""
        return v.strip()

    def is_positive(self) -> bool:
        """Check if this is a positive review."""
        return self.rating >= 4


class EmailNotification(BaseModel):
    """Email notification configuration."""
    notification_id: str = Field(..., min_length=8)
    recipient_email: str = Field(..., pattern=r"^[\w\.-]+@[\w\.-]+\.\w+$")
    subject: str = Field(..., min_length=1, max_length=200)
    body_html: str = Field(..., min_length=1)
    body_text: Optional[str] = Field(None)
    priority: int = Field(default=3, ge=1, le=5)
    retry_count: int = Field(default=0, ge=0, le=5)
    scheduled_at: Optional[datetime] = Field(None)
    sent_at: Optional[datetime] = Field(None)

    @validator("subject")
    def strip_subject(cls, v):
        """Strip whitespace from subject."""
        return v.strip()

    def is_sent(self) -> bool:
        """Check if notification was sent."""
        return self.sent_at is not None


class ApiRateLimit(BaseModel):
    """API rate limiting configuration."""
    client_id: str = Field(..., min_length=8)
    endpoint_pattern: str = Field(..., min_length=1, max_length=200)
    requests_per_minute: int = Field(default=60, ge=1, le=10000)
    requests_per_hour: int = Field(default=1000, ge=1, le=100000)
    burst_size: int = Field(default=10, ge=1, le=100)
    enabled: bool = Field(default=True)
    last_reset: datetime = Field(default_factory=datetime.utcnow)

    def should_throttle(self, current_count: int) -> bool:
        """Check if request should be throttled."""
        return self.enabled and current_count >= self.requests_per_minute
