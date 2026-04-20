# concrete-instantiation-dependency

**Category**: 依赖注入  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记在 `__init__` 内部构造具体协作对象的类（例如 `self.repo = SqlUserRepository()`），而不是将其作为参数接收。

## 为什么这是一个问题

当一个类自行构建协作对象时，它也拥有了它们的配置、生命周期和传递依赖。`OrderService()` 突然需要数据库连接字符串、HTTP 客户端和凭证 — 全部是隐式获取的 — 因为 `SqlUserRepository()` 和 `StripeGateway()` 需要它们。

测试首先感受到这一点。没有接缝可以注入假对象：每个测试要么必须启动真实的协作对象，要么对正在构造的类进行猴子补丁。集成测试和单元测试的界限模糊了，测试套件也变慢了。

将协作对象作为参数传入可以使依赖显式化，让类专注于自身的逻辑，并允许不同的调用点装配不同的实现 — 生产环境使用真实的仓储，测试中使用内存仓储。

## 示例

```python
class OrderService:
    def __init__(self):
        self.repo = SqlOrderRepository(DATABASE_URL)
        self.payments = StripeGateway(api_key=STRIPE_KEY)

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## 修正示例

通过 `__init__` 接收协作对象，并在组合根中构造它们一次。

```python
class OrderService:
    def __init__(self, repo, payments):
        self.repo = repo
        self.payments = payments

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 提高到 `"error"` 可抑制此规则。 |

## 参考

- 具体依赖检测（`internal/analyzer/concrete_dependency_detector.go`）。
- [规则目录](index.md) · [Concrete type hint dependency](concrete-type-hint-dependency.md) · [Service locator pattern](service-locator-pattern.md)
