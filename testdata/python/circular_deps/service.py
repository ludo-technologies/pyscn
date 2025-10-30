"""Service module - Part of larger dependency chain."""

import repository
import cache


class BusinessService:
    """Business logic service."""

    def __init__(self):
        self.repo = repository.DataRepository()
        self.cache = cache.CacheService()

    def process(self, data):
        """Process business logic."""
        cached = self.cache.get(f"process_{data}")
        if cached:
            return cached

        result = self.repo.fetch_data(data)
        processed = self._transform(result)
        self.cache.set(f"process_{data}", processed)
        return processed

    def _transform(self, data):
        """Transform data."""
        return f"transformed_{data}"


class ValidationService:
    """Data validation service."""

    def validate(self, data):
        """Validate input data."""
        return len(str(data)) > 0
