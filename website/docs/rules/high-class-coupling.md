# high-class-coupling

**Category**: Class Design  
**Severity**: Configurable by threshold  
**Triggered by**: `pyscn analyze`, `pyscn check`

## What it does

Flags classes that depend on too many other classes (the Coupling Between Objects, or CBO, metric from Chidamber & Kemerer). pyscn counts the number of distinct other classes a class references through inheritance, type hints, direct instantiation, attribute access on imported modules, and imports.

In plain terms: *too many things have to be in place for this class to work.*

## Why is this a problem?

A highly coupled class is hard to live with:

- **Hard to test** — constructing it in a unit test drags in a web of collaborators, so tests become integration tests or rely on heavy mocking.
- **Hard to change** — a signature change in any one dependency ripples into this class.
- **Hard to reuse** — you cannot lift it into another project without lifting its whole neighborhood.
- **A sign of missing abstraction** — the class is probably orchestrating things that belong behind a smaller interface.

## Example

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

`OrderService` is coupled to 8 concrete vendor classes. Swapping Stripe for Adyen, or running a test without a live Postgres, means editing `OrderService`.

## Use instead

Depend on small protocols, and inject collaborators through `__init__`. The service no longer knows which vendor is on the other end.

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

If one class still legitimately needs many collaborators, split it by responsibility (e.g. `Checkout`, `Fulfillment`, `Receipt`) and let an orchestrator call them.

## Options

| Option | Default | Description |
| --- | --- | --- |
| [`cbo.low_threshold`](../configuration/reference.md#cbo) | `3` | At or below this, the class is reported as low risk. |
| [`cbo.medium_threshold`](../configuration/reference.md#cbo) | `7` | Above this, the class is high risk. |
| [`cbo.min_cbo`](../configuration/reference.md#cbo) | `0` | Classes with coupling below this value are omitted from the report. |
| [`cbo.include_builtins`](../configuration/reference.md#cbo) | `false` | Count built-in types (`list`, `dict`, `Exception`, …) as dependencies. |
| [`cbo.include_imports`](../configuration/reference.md#cbo) | `true` | Count classes reached only through `import` statements. |

## References

- Chidamber, S. R. & Kemerer, C. F. *A Metrics Suite for Object Oriented Design.* IEEE TSE, 1994.
- Implementation: `internal/analyzer/cbo.go`.
- [Rule catalog](index.md) · [low-class-cohesion](low-class-cohesion.md)
