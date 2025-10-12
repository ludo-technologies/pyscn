"""Cache module - Core infrastructure appearing in multiple cycles."""

import database
import logger


class CacheService:
    """Caching service that depends on database for persistence."""

    def __init__(self):
        self.db = None
        self.logger = logger.Logger()
        self.memory_cache = {}

    def get(self, key):
        """Get value from cache."""
        if key in self.memory_cache:
            return self.memory_cache[key]

        # Fallback to database
        if self.db is None:
            self.db = database.DatabaseConnection()
        result = self.db.query(f"SELECT * FROM cache WHERE key='{key}'")
        return result

    def set(self, key, value):
        """Set value in cache."""
        self.memory_cache[key] = value
        self.logger.log(f"Cache set: {key}")

    def invalidate(self, key):
        """Invalidate cache entry."""
        if key in self.memory_cache:
            del self.memory_cache[key]


class CacheStats:
    """Cache statistics tracker."""

    def get_hit_rate(self):
        """Get cache hit rate."""
        return 0.85
