# circular-import

**カテゴリ**: モジュール構造  
**重大度**: サイクルサイズに応じて設定可能 (Low / Medium / High / Critical)  
**トリガー**: `pyscn analyze`, `pyscn check --select circular`

## 検出内容

インポートサイクルを形成するモジュールグループを検出します。モジュール A が B を（直接的または推移的に）インポートし、B が A をインポートしている状態です。サイクルはモジュール依存グラフに対して Tarjan の強連結成分アルゴリズムを実行することで検出されます。

重大度はサイクルのサイズとメンバーのファンインから割り当てられます:

| サイクルメンバー数 | 重大度 |
| --- | --- |
| 2 | Low |
| 3 -- 5 | Medium |
| 6 -- 9 | High |
| 10以上、またはファンイン > 10 のメンバーを含む場合 | Critical |

## なぜ問題なのか

循環インポートは、2つ以上のモジュールが独立して理解・テスト・リリースできないことを意味します。具体的には:

- **インポート時エラー。** Python は循環インポート中にモジュールを部分的に初期化します。半分しか読み込まれていないモジュールへの属性アクセスは、文の順序に応じて `ImportError` または `AttributeError` を発生させます。
- **密結合。** サイクルのメンバーはファイルに分割された単一の「論理モジュール」を共有しています。一つを変更すると、他のすべてにも変更が波及する傾向があります。
- **リファクタリングの阻害。** サイクルのいずれのメンバーも、他のメンバーに手を加えずに移動・リネーム・削除することができません。
- **サイクルが大きくなるほど深刻化。** 2モジュールのサイクルは煩わしい程度ですが、10モジュールのサイクルはアーキテクチャの破綻です。重大度が段階的に上がるのはそのためです。

## 例

```python
# myapp/orders.py
from myapp.billing import Invoice

class Order:
    def invoice(self) -> Invoice:
        return Invoice(self)
```

```python
# myapp/billing.py
from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"):
        self.order = order
```

`orders` は戻り値の型のために `billing` をインポートし、`billing` はコンストラクタパラメータのために `orders` をインポートしています。どちらのファイルをトップレベルで実行してもサイクルが発生します。

## 修正例

共有する型を第三のモジュールに抽出し、両方がお互いではなくそのモジュールに依存するようにします:

```python
# myapp/domain.py
class Order: ...
class Invoice: ...
```

```python
# myapp/orders.py
from myapp.domain import Order, Invoice
```

```python
# myapp/billing.py
from myapp.domain import Order, Invoice
```

逆方向のインポートが型アノテーションにのみ必要な場合は、`TYPE_CHECKING` でガードして実行時に評価されないようにします:

```python
# myapp/billing.py
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    from myapp.orders import Order

class Invoice:
    def __init__(self, order: "Order"): ...
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`dependencies.detect_cycles`](../configuration/reference.md#dependencies) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dependencies.cycle_reporting`](../configuration/reference.md#dependencies) | `"summary"` | `all`、`critical`、または `summary` -- レポートに表示するサイクル数を制御します。 |
| [`dependencies.max_cycles_to_show`](../configuration/reference.md#dependencies) | `10` | レポートに表示するサイクル数の上限。 |
| `--max-cycles N` (check) | `0` | サイクル数が `N` を超えた場合に `check` コマンドを失敗させます。 |
| `--allow-circular-deps` (check) | off | サイクルを失敗ではなく警告に降格させます。 |

## 参照

- Tarjan SCC の実装 (`internal/analyzer/circular_detector.go`)、モジュールグラフの構築 (`internal/analyzer/module_analyzer.go`)。
- [ルールカタログ](index.md) · [deep-import-chain](deep-import-chain.md)
