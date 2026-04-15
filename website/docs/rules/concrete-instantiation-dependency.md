# concrete-instantiation-dependency

**Category**: Dependency Injection  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags a class that constructs a concrete collaborator inside `__init__` (e.g. `self.repo = SqlUserRepository()`) rather than receiving it as a parameter.

## Why is this a problem?

When a class builds its own collaborators, it also owns their configuration, their lifetime, and their transitive dependencies. `OrderService()` suddenly requires a database connection string, an HTTP client, and credentials — all reached for implicitly — because `SqlUserRepository()` and `StripeGateway()` do.

Tests feel this first. There is no seam to swap in a fake: every test has to either spin up the real collaborator or monkey-patch the class being constructed. Integration tests and unit tests blur together, and the test suite slows down.

Passing the collaborator in makes the dependency explicit, keeps the class focused on its own logic, and lets different call sites wire different implementations — a real repository in production, an in-memory one in tests.

## Example

```python
class OrderService:
    def __init__(self):
        self.repo = SqlOrderRepository(DATABASE_URL)
        self.payments = StripeGateway(api_key=STRIPE_KEY)

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## Use instead

Receive the collaborators via `__init__` and construct them once at the composition root.

```python
class OrderService:
    def __init__(self, repo, payments):
        self.repo = repo
        self.payments = payments

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | Raise to `"error"` to suppress this rule. |

## References

- Concrete dependency detection (`internal/analyzer/concrete_dependency_detector.go`).
- [Rule catalog](index.md) · [Concrete type hint dependency](concrete-type-hint-dependency.md) · [Service locator pattern](service-locator-pattern.md)
