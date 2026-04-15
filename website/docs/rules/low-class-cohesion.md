# low-class-cohesion

**Category**: Class Design  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags classes whose methods don't share instance state (the LCOM4 metric — Lack of Cohesion of Methods, version 4). pyscn builds a graph where two methods are connected if they touch a common `self.` attribute, then counts the connected components. `LCOM4 = 1` means every method is related to every other; `LCOM4 = N` means the class is really `N` unrelated sub-classes glued together.

Methods decorated with `@staticmethod` or `@classmethod` do not reference `self` and are excluded from the graph.

In plain terms: *this class is doing unrelated jobs — split it, or make it a module of functions.*

## Why is this a problem?

A class is meant to bundle state with the operations that act on it. When the methods don't touch the same state:

- **The class name lies** — it claims to be one thing but behaves like two or three.
- **Changes scatter** — a bug in one responsibility can only be found by reading code that has nothing to do with it.
- **Reuse is blocked** — you cannot lift out the part you need without dragging the rest along.
- **It often predates a real abstraction** — the "Utilities" or "Manager" class is a classic symptom.

## Example

```python
class UserUtility:
    def __init__(self, db, smtp, clock):
        self.db = db
        self.smtp = smtp
        self.clock = clock
        self.cache = {}

    # --- persistence ---
    def load(self, user_id):
        if user_id in self.cache:
            return self.cache[user_id]
        row = self.db.fetch("users", user_id)
        self.cache[user_id] = row
        return row

    def save(self, user):
        self.db.upsert("users", user)
        self.cache[user.id] = user

    # --- email ---
    def send_welcome(self, address):
        self.smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self.smtp.send(address, f"Reset: {token}")

    # --- formatting ---
    def format_joined_at(self, user):
        return self.clock.format(user.joined_at)
```

`LCOM4 = 3`: `{load, save}` share `db` and `cache`, `{send_welcome, send_reset}` share `smtp`, `{format_joined_at}` stands alone. Three components, one class.

## Use instead

Split into cohesive classes, and move the stateless part out to free functions.

```python
class UserRepository:
    def __init__(self, db):
        self._db = db
        self._cache = {}

    def load(self, user_id):
        if user_id in self._cache:
            return self._cache[user_id]
        row = self._db.fetch("users", user_id)
        self._cache[user_id] = row
        return row

    def save(self, user):
        self._db.upsert("users", user)
        self._cache[user.id] = user


class UserMailer:
    def __init__(self, smtp):
        self._smtp = smtp

    def send_welcome(self, address):
        self._smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self._smtp.send(address, f"Reset: {token}")


# user_formatting.py — no class, no state
def format_joined_at(user, clock):
    return clock.format(user.joined_at)
```

Each class now has `LCOM4 = 1`, and the formatter is a one-line function where it belongs.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`lcom.low_threshold`](../configuration/reference.md#lcom) | `2` | At or below this, the class is reported as low risk. |
| [`lcom.medium_threshold`](../configuration/reference.md#lcom) | `5` | Above this, the class is high risk. |

## References

- Hitz, M. & Montazeri, B. *Chidamber and Kemerer's Metrics Suite: A Measurement Theory Perspective.* IEEE TSE, 1996 (LCOM4 definition).
- Implementation: `internal/analyzer/lcom.go`.
- [Rule catalog](index.md) · [high-class-coupling](high-class-coupling.md)
