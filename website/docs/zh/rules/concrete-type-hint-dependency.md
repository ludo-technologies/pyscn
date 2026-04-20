# concrete-type-hint-dependency

**Category**: 依赖注入  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记 `__init__` 参数的类型提示为具体类而非 `Protocol`、抽象基类或接口的情况。

## 为什么这是一个问题

具体类型提示告诉阅读者和类型检查器，这个类只接受一种特定的实现。即使运行时可以愉快地接受鸭子类型的替代品，声明的契约说的并非如此，遵循类型提示的工具（mypy、IDE 自动补全、基于类型构建的测试 mock）也会拒绝替代品。

在实践中，这使测试更难编写：想要替换一个内存假对象的测试必须要么继承具体类（继承其所有行为），要么抑制类型错误。这也将消费者绑定到具体类所需的任何导入上，导致一个小工具最终拉入了整个数据库技术栈。

依赖 `Protocol` 或抽象接口可以记录类实际使用了什么 — 一两个方法 — 并为假对象、适配器和未来的实现留有余地。

## 示例

```python
class SqlUserRepository:
    def find(self, user_id): ...

class UserService:
    def __init__(self, repo: SqlUserRepository):
        self.repo = repo
```

## 修正示例

声明一个 `Protocol` 来描述你依赖的方法，然后依赖它。

```python
from typing import Protocol

class UserRepository(Protocol):
    def find(self, user_id: str) -> User: ...

class UserService:
    def __init__(self, repo: UserRepository):
        self.repo = repo
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 此规则以 `info` 级别报告；将 `min_severity` 降低到 `"info"` 才能看到。 |

## 参考

- 具体依赖检测（`internal/analyzer/concrete_dependency_detector.go`）。
- [规则目录](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
