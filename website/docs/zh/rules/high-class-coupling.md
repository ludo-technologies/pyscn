# high-class-coupling

**Category**: 类设计  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 检测内容

标记依赖了过多其他类的类（即 Chidamber & Kemerer 提出的对象间耦合度 CBO 指标）。pyscn 通过继承、类型提示、直接实例化、对导入模块的属性访问和导入语句来统计一个类引用的不同类的数量。

简单来说：*这个类要正常工作需要太多前置条件。*

## 为什么这是一个问题

高度耦合的类难以维护：

- **难以测试** — 在单元测试中构造它需要引入大量协作对象，导致测试变成集成测试或依赖大量的 mock。
- **难以修改** — 任何一个依赖项的签名变更都会波及到这个类。
- **难以复用** — 你无法将它移到另一个项目中而不连带迁移它的整个关联体系。
- **缺少抽象的信号** — 这个类很可能在编排一些本应隐藏在更小接口背后的东西。

## 示例

```python
from billing.stripe_gateway import StripeGateway
from billing.paypal_gateway import PayPalGateway
from notifications.sendgrid import SendGridClient
from notifications.twilio import TwilioClient
from storage.s3 import S3Bucket
from storage.postgres import PostgresConnection
from audit.datadog import DatadogLogger
from auth.okta import OktaClient

class OrderService:
    def __init__(self):
        self.stripe = StripeGateway()
        self.paypal = PayPalGateway()
        self.email = SendGridClient()
        self.sms = TwilioClient()
        self.blobs = S3Bucket("orders")
        self.db = PostgresConnection()
        self.audit = DatadogLogger()
        self.auth = OktaClient()

    def place(self, user, cart): ...
```

`OrderService` 耦合了 8 个具体的供应商类。要将 Stripe 换成 Adyen，或者在没有实际 Postgres 的情况下运行测试，都意味着要修改 `OrderService`。

## 修正示例

依赖小型协议，并通过 `__init__` 注入协作对象。服务不再需要知道对面是哪个供应商。

```python
from typing import Protocol

class PaymentGateway(Protocol):
    def charge(self, amount: int, token: str) -> str: ...

class Notifier(Protocol):
    def notify(self, user_id: str, message: str) -> None: ...

class OrderRepository(Protocol):
    def save(self, order) -> None: ...

class OrderService:
    def __init__(
        self,
        payments: PaymentGateway,
        notifier: Notifier,
        repo: OrderRepository,
    ):
        self._payments = payments
        self._notifier = notifier
        self._repo = repo

    def place(self, user, cart): ...
```

如果一个类确实需要很多协作对象，可以按职责拆分（例如 `Checkout`、`Fulfillment`、`Receipt`），然后让一个编排器来调用它们。

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| [`cbo.low_threshold`](../configuration/reference.md#cbo) | `3` | 等于或低于此值，类报告为低风险。 |
| [`cbo.medium_threshold`](../configuration/reference.md#cbo) | `7` | 高于此值，类为高风险。 |
| [`cbo.min_cbo`](../configuration/reference.md#cbo) | `0` | 耦合度低于此值的类将从报告中省略。 |
| [`cbo.include_builtins`](../configuration/reference.md#cbo) | `false` | 是否将内置类型（`list`、`dict`、`Exception` 等）计为依赖。 |
| [`cbo.include_imports`](../configuration/reference.md#cbo) | `true` | 是否将仅通过 `import` 语句引用的类计为依赖。 |

## 参考

- Chidamber, S. R. & Kemerer, C. F. *A Metrics Suite for Object Oriented Design.* IEEE TSE, 1994.
- 实现：`internal/analyzer/cbo.go`。
- [规则目录](index.md) · [low-class-cohesion](low-class-cohesion.md)
