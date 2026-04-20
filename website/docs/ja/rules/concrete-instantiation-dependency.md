# concrete-instantiation-dependency

**カテゴリ**: 依存性注入  
**重大度**: Warning  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

パラメータとして受け取るのではなく、`__init__` 内部で具象の協調オブジェクトを構築するクラスを検出します（例: `self.repo = SqlUserRepository()`）。

## なぜ問題なのか

クラスが自身の協調オブジェクトを構築すると、その設定、ライフタイム、および推移的な依存も所有することになります。`OrderService()` は、`SqlUserRepository()` や `StripeGateway()` が暗黙的に取得するデータベース接続文字列、HTTP クライアント、認証情報を突然必要とします。

テストが最初にこの影響を受けます。フェイクに差し替えるための接合点がなく、すべてのテストが本物の協調オブジェクトを起動するか、構築されるクラスをモンキーパッチする必要があります。統合テストとユニットテストの境界が曖昧になり、テストスイートが遅くなります。

協調オブジェクトを渡すようにすることで、依存が明示的になり、クラスは自身のロジックに集中でき、異なる呼び出し元が異なる実装を接続できます — 本番では実際のリポジトリ、テストではインメモリのもの。

## 例

```python
class OrderService:
    def __init__(self):
        self.repo = SqlOrderRepository(DATABASE_URL)
        self.payments = StripeGateway(api_key=STRIPE_KEY)

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## 修正例

`__init__` で協調オブジェクトを受け取り、コンポジションルートで一度だけ構築します。

```python
class OrderService:
    def __init__(self, repo, payments):
        self.repo = repo
        self.payments = payments

    def place(self, order):
        self.repo.save(order)
        self.payments.charge(order)
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | `"error"` に上げるとこのルールは非表示になります。 |

## 参照

- 具象依存の検出 (`internal/analyzer/concrete_dependency_detector.go`)。
- [ルールカタログ](index.md) · [Concrete type hint dependency](concrete-type-hint-dependency.md) · [Service locator pattern](service-locator-pattern.md)
