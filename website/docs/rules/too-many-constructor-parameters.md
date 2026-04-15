# too-many-constructor-parameters

**Category**: Dependency Injection  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags an `__init__` method whose parameter count (excluding `self`) exceeds `di.constructor_param_threshold` (default `5`).

## Why is this a problem?

A long constructor signature is a symptom of a class that has taken on too many responsibilities. Each parameter is a collaborator the class knows about, and every call site has to supply all of them — including tests, which end up building elaborate setup fixtures just to construct the object.

Over time, adding one more dependency feels cheap, so the list grows. Readers lose track of what the class actually needs versus what is incidental, and reordering or defaulting parameters becomes risky.

When you see more than five dependencies, the class is usually doing two or three jobs that could be separated, or several of the parameters belong together as a single object.

## Example

```python
class OrderService:
    def __init__(
        self,
        user_repo,
        order_repo,
        payment_gateway,
        inventory,
        notifier,
        audit_log,
        clock,
    ):
        self.user_repo = user_repo
        self.order_repo = order_repo
        ...
```

## Use instead

Group related collaborators into a parameter object, or split the class along its responsibilities.

```python
@dataclass
class OrderDependencies:
    users: UserRepository
    orders: OrderRepository
    payments: PaymentGateway
    inventory: Inventory
    notifier: Notifier

class OrderService:
    def __init__(self, deps: OrderDependencies, clock: Clock):
        self.deps = deps
        self.clock = clock
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. `check --select di` enables it implicitly. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Raise to `"error"` to hide this rule, or lower to `"info"` to surface more. |
| [`di.constructor_param_threshold`](../configuration/reference.md#di) | `5` | Parameter count above which `__init__` is flagged. |

## References

- Constructor over-injection detection (`internal/analyzer/constructor_analyzer.go`).
- [Rule catalog](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Service locator pattern](service-locator-pattern.md)
