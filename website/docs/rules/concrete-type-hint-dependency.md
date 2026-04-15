# concrete-type-hint-dependency

**Category**: Dependency Injection  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## What it does

Flags an `__init__` parameter whose type hint is a concrete class rather than a `Protocol`, abstract base class, or interface.

## Why is this a problem?

A concrete type hint tells the reader — and the type checker — that this class will only accept one specific implementation. Even if the runtime happily accepts duck-typed substitutes, the declared contract says otherwise, and tools that honour the hint (mypy, IDE autocompletion, test mocks built from the type) will refuse alternatives.

In practice this makes tests harder to write: a test that wants to substitute an in-memory fake has to either inherit from the concrete class (picking up all its behaviour) or suppress the type error. It also ties the consumer to whatever imports the concrete class needs, so a small utility ends up pulling in the full database stack.

Depending on a `Protocol` or abstract interface documents what the class actually uses — a method or two — and leaves room for fakes, adapters, and future implementations.

## Example

```python
class SqlUserRepository:
    def find(self, user_id): ...

class UserService:
    def __init__(self, repo: SqlUserRepository):
        self.repo = repo
```

## Use instead

Declare a `Protocol` describing the methods you rely on and depend on that.

```python
from typing import Protocol

class UserRepository(Protocol):
    def find(self, user_id: str) -> User: ...

class UserService:
    def __init__(self, repo: UserRepository):
        self.repo = repo
```

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | Must be `true` for `analyze` to run DI rules. |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | This rule reports at `info`; lower `min_severity` to `"info"` to see it. |

## References

- Concrete dependency detection (`internal/analyzer/concrete_dependency_detector.go`).
- [Rule catalog](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
