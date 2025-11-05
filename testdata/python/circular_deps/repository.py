"""Repository module - Part of larger dependency chain."""

import database
import controller


class DataRepository:
    """Data access layer."""

    def __init__(self):
        self.db = database.DatabaseConnection()
        self.controller = None

    def fetch_data(self, key):
        """Fetch data from database."""
        result = self.db.query(f"SELECT * FROM data WHERE key='{key}'")
        return result

    def save_data(self, key, value):
        """Save data to database."""
        self.db.query(f"INSERT INTO data VALUES ('{key}', '{value}')")

    def validate_access(self, user):
        """Validate user access by checking with controller."""
        if self.controller is None:
            # This creates a cycle: repository -> controller -> service -> repository
            self.controller = controller.UserController()
        return True


class QueryBuilder:
    """SQL query builder."""

    def build_query(self, table, conditions):
        """Build SQL query."""
        return f"SELECT * FROM {table} WHERE {conditions}"
