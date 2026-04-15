# module-variable-dependency

**Category**: Dependency Injection  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags a class that directly reads or writes a module-level mutable variable without a `global` statement — an implicit coupling to module state.

## Why is this a problem?

Unlike a `global` assignment, reading a module-level name is silent: the class looks self-contained, but its behaviour actually depends on whatever lives in that module variable at call time. A test that instantiates the class in isolation may still pass or fail based on an unrelated import.

This also breaks substitutability. You can't give the class a different collaborator without monkey-patching the module, which is fragile and order-dependent. Two instances of the class are forced to share the same backing object whether you want that or not.

Making the collaborator a constructor parameter documents the dependency, allows per-instance configuration, and lets tests inject a fake without touching module globals.

## Example

```python
config = load_config()

class UserRepository:
    def find(self, user_id):
        conn = connect(config.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## Use instead

Accept the collaborator as a constructor parameter.

```python
class UserRepository:
    def __init__(self, database_url):
        self.database_url = database_url

    def find(self, user_id):
        conn = connect(self.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Raise to `"error"` to suppress this rule. |

## References

- Hidden dependency detection (`internal/analyzer/hidden_dependency_detector.go`).
- [Rule catalog](index.md) · [Global state dependency](global-state-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
