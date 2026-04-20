# too-many-constructor-parameters

**Category**: 依赖注入  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记 `__init__` 方法的参数数量（不含 `self`）超过 `di.constructor_param_threshold`（默认 `5`）的类。

## 为什么这是一个问题

过长的构造函数签名是类承担了过多职责的症状。每个参数都是类需要了解的一个协作对象，而每个调用点都必须提供所有参数 — 包括测试，测试最终需要构建复杂的前置设置才能创建对象。

随着时间推移，再加一个依赖看起来代价很小，因此列表不断增长。阅读者无法分辨类实际需要什么和哪些是附带的，重新排序或设置默认参数也变得有风险。

当你看到超过五个依赖时，这个类通常在做两三件可以分离的工作，或者其中几个参数应该合并为一个对象。

## 示例

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

## 修正示例

将相关的协作对象分组为参数对象，或按职责拆分类。

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

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。`check --select di` 会隐式启用。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 提高到 `"error"` 可隐藏此规则，降低到 `"info"` 可显示更多。 |
| [`di.constructor_param_threshold`](../configuration/reference.md#di) | `5` | `__init__` 参数数量超过此值时将被标记。 |

## 参考

- 构造函数过度注入检测（`internal/analyzer/constructor_analyzer.go`）。
- [规则目录](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Service locator pattern](service-locator-pattern.md)
