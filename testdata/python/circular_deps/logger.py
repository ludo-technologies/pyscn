"""Logger module - Core infrastructure appearing in multiple cycles."""

import database


class Logger:
    """Logging service that persists logs to database."""

    def __init__(self):
        self.db = None
        self.buffer = []

    def log(self, message):
        """Log a message."""
        self.buffer.append(message)
        if len(self.buffer) >= 10:
            self._flush()

    def _flush(self):
        """Flush logs to database."""
        if self.db is None:
            self.db = database.DatabaseConnection()

        for message in self.buffer:
            # This creates a cycle: logger -> database -> cache -> logger
            self.db.query(f"INSERT INTO logs VALUES ('{message}')")

        self.buffer = []

    def error(self, message):
        """Log an error message."""
        self.log(f"ERROR: {message}")

    def warning(self, message):
        """Log a warning message."""
        self.log(f"WARNING: {message}")


class LogFormatter:
    """Log message formatter."""

    def format(self, message):
        """Format log message."""
        return f"[LOG] {message}"
