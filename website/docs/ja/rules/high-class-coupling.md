# high-class-coupling

**カテゴリ**: クラス設計  
**重大度**: しきい値により設定可能  
**検出コマンド**: `pyscn analyze`, `pyscn check`

## 検出内容

他のクラスに過度に依存しているクラスを検出します（Chidamber & Kemerer による CBO: Coupling Between Objects メトリクス）。pyscn は、継承、型ヒント、直接的なインスタンス化、インポートされたモジュールの属性アクセス、およびインポートを通じてクラスが参照する、異なるクラスの数をカウントします。

簡単に言えば: *このクラスが動作するために、あまりにも多くのものが揃っていなければなりません。*

## なぜ問題なのか

結合度の高いクラスは扱いにくいものです:

- **テストが困難** — ユニットテストでインスタンスを構築すると、多数の協調オブジェクトが連鎖的に必要となり、テストが統合テストになるか、大量のモックに依存することになります。
- **変更が困難** — いずれかの依存先のシグネチャ変更が、このクラスに波及します。
- **再利用が困難** — 近隣の全体を一緒に持ち出さなければ、別のプロジェクトに移すことができません。
- **抽象化の欠如の兆候** — このクラスはおそらく、より小さなインターフェースの背後にあるべきものを直接扱っています。

## 例

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

`OrderService` は8つの具象ベンダークラスに結合しています。Stripe を Adyen に切り替えたり、ライブ Postgres なしでテストを実行するには、`OrderService` を編集する必要があります。

## 修正例

小さなプロトコルに依存し、`__init__` を通じて協調オブジェクトを注入します。サービスはもはやどのベンダーが相手側にいるかを知りません。

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

1つのクラスがそれでも多くの協調オブジェクトを正当に必要とする場合は、責務ごとに分割し（例: `Checkout`、`Fulfillment`、`Receipt`）、オーケストレーターがそれらを呼び出すようにします。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`cbo.low_threshold`](../configuration/reference.md#cbo) | `3` | この値以下のクラスは低リスクとして報告されます。 |
| [`cbo.medium_threshold`](../configuration/reference.md#cbo) | `7` | この値を超えると高リスクとなります。 |
| [`cbo.min_cbo`](../configuration/reference.md#cbo) | `0` | 結合度がこの値未満のクラスはレポートから除外されます。 |
| [`cbo.include_builtins`](../configuration/reference.md#cbo) | `false` | 組み込み型（`list`、`dict`、`Exception` など）を依存としてカウントします。 |
| [`cbo.include_imports`](../configuration/reference.md#cbo) | `true` | `import` 文を通じてのみ到達するクラスをカウントします。 |

## 参照

- Chidamber, S. R. & Kemerer, C. F. *A Metrics Suite for Object Oriented Design.* IEEE TSE, 1994.
- 実装: `internal/analyzer/cbo.go`。
- [ルールカタログ](index.md) · [low-class-cohesion](low-class-cohesion.md)
