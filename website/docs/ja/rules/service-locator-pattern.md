# service-locator-pattern

**カテゴリ**: 依存性注入  
**重大度**: Warning  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

ロケーター、レジストリ、またはコンテナを受け取り、メソッド呼び出し時に名前で依存を取得するクラスを検出します（例: `self.locator.get("payment_service")`）。

## なぜ問題なのか

サービスロケーターは、1つの明確な依存を多数の隠れた依存に置き換えます。コンストラクタシグネチャはクラスが1つのオブジェクトを必要とすることを示唆しますが、実際の契約は「このクラスが実行時にたまたまルックアップする文字列が何であれ」です。読み手は、`OrderService` が実際にはペイメントゲートウェイ、通知機能、クロックを必要としていることを発見するために、すべてのメソッドを grep で検索しなければなりません。

テストでは、クラスが要求するすべてのキーを知っているフェイクロケーターを提供する必要があり、キーの欠落は通常、明確な構築時の失敗ではなく、メソッドの深い部分での `AttributeError` や `KeyError` として表面化します。サービス名の変更にはすべての文字列ルックアップを見つける必要があり、静的解析やIDEのリファクタリングツールは役に立ちません。

実際のサービスを `__init__` で渡すことで、クラスに検査可能な契約を与え、文字列ベースの間接参照を排除します。

## 例

```python
class OrderService:
    def __init__(self, locator):
        self.locator = locator

    def place(self, order):
        self.locator.get("order_repo").save(order)
        self.locator.get("payment_service").charge(order)
        self.locator.get("notifier").send(order.user, "placed")
```

## 修正例

各サービスを直接受け取り、依存を可視化し型チェック可能にします。

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

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | `"error"` に上げるとこのルールは非表示になります。 |

## 参照

- サービスロケーターの検出 (`internal/analyzer/service_locator_detector.go`)。
- [ルールカタログ](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
