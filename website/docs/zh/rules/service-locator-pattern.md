# service-locator-pattern

**Category**: 依赖注入  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select di`

## 检测内容

标记接收定位器、注册表或容器并在方法调用时从中拉取命名依赖的类，例如 `self.locator.get("payment_service")`。

## 为什么这是一个问题

服务定位器用一个清晰的依赖换来了许多隐藏的依赖。构造函数签名暗示类只需要一个对象，但真正的契约是"这个类在运行时碰巧查找的任何字符串"。阅读者必须遍历每个方法才能发现 `OrderService` 实际上需要支付网关、通知器和时钟。

测试必须提供一个知道类可能请求的每个键的假定位器，而缺少的键通常在方法深处以 `AttributeError` 或 `KeyError` 的形式出现，而不是以清晰的构造失败呈现。重命名一个服务需要找到每个字符串查找；静态分析和 IDE 重构工具也帮不上忙。

直接通过 `__init__` 传递实际服务可以为类提供一个可检查的契约，并消除基于字符串的间接引用。

## 示例

```python
class OrderService:
    def __init__(self, locator):
        self.locator = locator

    def place(self, order):
        self.locator.get("order_repo").save(order)
        self.locator.get("payment_service").charge(order)
        self.locator.get("notifier").send(order.user, "placed")
```

## 修正示例

直接接收每个服务，使依赖可见且可进行类型检查。

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

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | 必须为 `true` 才能在 `analyze` 中运行 DI 规则。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | 提高到 `"error"` 可抑制此规则。 |

## 参考

- 服务定位器检测（`internal/analyzer/service_locator_detector.go`）。
- [规则目录](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
