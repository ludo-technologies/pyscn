# too-many-constructor-parameters

**カテゴリ**: 依存性注入  
**重大度**: Warning  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

`self` を除くパラメータ数が `di.constructor_param_threshold`（デフォルト `5`）を超える `__init__` メソッドを検出します。

## なぜ問題なのか

長いコンストラクタシグネチャは、クラスが多くの責務を引き受けすぎている症状です。各パラメータはクラスが知っている協調オブジェクトであり、すべての呼び出し元がそれらをすべて提供しなければなりません。テストでは、オブジェクトを構築するためだけに手の込んだセットアップフィクスチャを作成することになります。

時間が経つにつれ、依存を1つ追加するのは安上がりに感じるため、リストは増え続けます。読み手はクラスが実際に何を必要としているのか、何が付随的なものなのかを見失い、パラメータの並べ替えやデフォルト値の設定がリスクを伴うようになります。

5つ以上の依存がある場合、そのクラスは通常、分離可能な2〜3の仕事をしているか、いくつかのパラメータが1つのオブジェクトとしてまとめられるべきものです。

## 例

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

## 修正例

関連する協調オブジェクトをパラメータオブジェクトにまとめるか、責務に沿ってクラスを分割します。

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

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。`check --select di` は暗黙的に有効にします。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | `"error"` に上げるとこのルールは非表示になり、`"info"` に下げるとより多くのルールが表示されます。 |
| [`di.constructor_param_threshold`](../configuration/reference.md#di) | `5` | この数を超える `__init__` のパラメータ数で検出されます。 |

## 参照

- コンストラクタの過剰注入検出 (`internal/analyzer/constructor_analyzer.go`)。
- [ルールカタログ](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Service locator pattern](service-locator-pattern.md)
