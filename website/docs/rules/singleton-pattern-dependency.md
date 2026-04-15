# singleton-pattern-dependency

**Category**: Dependency Injection  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags a class that implements the singleton pattern by caching itself on a class-level `_instance` attribute.

## Why is this a problem?

A singleton is global state wearing a class-shaped costume. Every caller that writes `PaymentGateway.instance()` depends on the one object the class chooses to return, and that object survives across tests unless each test remembers to reset it. One forgetful test and the next one inherits stale state.

Because the singleton decides its own lifetime, callers can't give it different collaborators in different contexts — a second configuration, a fake for tests, a per-tenant instance. Substitution requires reaching into the class and resetting `_instance`, which is exactly the coupling the singleton was supposed to hide.

The pattern also obscures real dependencies: reading the code of a method that calls `X.instance()` tells you nothing about what `X` needs or where it was configured.

## Example

```python
class PaymentGateway:
    _instance = None

    @classmethod
    def instance(cls):
        if cls._instance is None:
            cls._instance = cls()
        return cls._instance

    def charge(self, order):
        ...
```

## Use instead

Construct the object once at the edge of the application and pass it to whatever needs it.

```python
class PaymentGateway:
    def charge(self, order):
        ...

# wiring, done once at startup
gateway = PaymentGateway()
order_service = OrderService(gateway)
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Raise to `"error"` to suppress this rule. |

## References

- Hidden dependency detection (`internal/analyzer/hidden_dependency_detector.go`).
- [Rule catalog](index.md) · [Global state dependency](global-state-dependency.md) · [Module variable dependency](module-variable-dependency.md)
