# global-state-dependency

**Category**: Dependency Injection  
**Severity**: Error  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags a class method that uses a `global` statement to read or mutate module-level state.

## Why is this a problem?

A `global` inside a method ties the class to a specific module variable that isn't visible from the class's interface. Nothing in `OrderService(...)` tells a reader that constructing it is not enough — some module-level value must also be primed, or the method will behave unexpectedly.

Tests suffer the most. To exercise a method that touches global state, each test has to reach into the module, save the old value, install a new one, and restore it on teardown — and any test that forgets to clean up leaks state into the next one. Running tests in parallel becomes unsafe.

The dependency is real; it's just hidden. Making it an explicit constructor parameter removes the surprise.

## Example

```python
_current_user = None

class AuditLog:
    def record(self, action):
        global _current_user
        entry = {"user": _current_user, "action": action}
        db.insert("audit", entry)
```

## Use instead

Pass the value in through `__init__` so the dependency is visible and swappable.

```python
class AuditLog:
    def __init__(self, user):
        self.user = user

    def record(self, action):
        db.insert("audit", {"user": self.user, "action": action})
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | This rule reports at `error`; it is surfaced unless `min_severity` is raised above `error`. |

## References

- Hidden dependency detection (`internal/analyzer/hidden_dependency_detector.go`).
- [Rule catalog](index.md) · [Module variable dependency](module-variable-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
