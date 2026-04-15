# service-locator-pattern

**Category**: Dependency Injection  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags a class that receives a locator, registry, or container and pulls named dependencies from it at method-call time, e.g. `self.locator.get("payment_service")`.

## Why is this a problem?

A service locator trades one clear dependency for many hidden ones. The constructor signature suggests the class needs a single object, but the real contract is "whatever strings this class happens to look up at runtime." A reader has to grep through every method to discover that `OrderService` actually needs a payment gateway, a notifier, and a clock.

Tests have to supply a fake locator that knows every key the class might ask for, and a missing key usually surfaces as an `AttributeError` or `KeyError` deep in a method rather than as a clear construction failure. Renaming a service requires finding every string lookup; static analysis and IDE refactoring tools can't help.

Passing the actual services through `__init__` gives the class a checkable contract and eliminates the string-based indirection.

## Example

```python
class OrderService:
    def __init__(self, locator):
        self.locator = locator

    def place(self, order):
        self.locator.get("order_repo").save(order)
        self.locator.get("payment_service").charge(order)
        self.locator.get("notifier").send(order.user, "placed")
```

## Use instead

Receive each service directly, so the dependencies are visible and type-checkable.

```python
class OrderService:
    def __init__(self, repo, payments, notifier):
        self.repo = repo
        self.payments = payments
        self.notifier = notifier

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
        self.notifier.send(order.user, "placed")
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Raise to `"error"` to suppress this rule. |

## References

- Service locator detection (`internal/analyzer/service_locator_detector.go`).
- [Rule catalog](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
