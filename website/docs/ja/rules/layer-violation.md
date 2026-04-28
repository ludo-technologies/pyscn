# layer-violation

**カテゴリ**: モジュール構造  
**重大度**: `architecture.rules[].severity` で設定可能  
**トリガー**: `pyscn analyze`, `pyscn check --select deps`

## 検出内容

ソースモジュールのレイヤーがターゲットモジュールのレイヤーに依存することが許可されていない場合に、`import` 文を検出します。許可ルールは設定した `[[architecture.rules]]` に基づきます。レイヤーは `[[architecture.layers]]` で定義されたパッケージ名のフラグメントにマッチすることで、モジュールに割り当てられます。

## なぜ問題なのか

レイヤードアーキテクチャは、レイヤーが維持されている間のみ効果を発揮します。`presentation` から `infrastructure` への単一のショートカットだけで、以下の問題が生じます:

- **テスタビリティの低下。** プレゼンテーション層が実際のデータベースや HTTP クライアントなしでは検証できなくなります。
- **隠れた結合の発生。** インフラストラクチャの実装を差し替えると、その存在を知るべきではなかった UI コードが暗黙的に壊れます。
- **違反の常態化。** 一つのショートカットが存在すると、次のショートカットを正当化しやすくなります。

このルールは、設計ドキュメントに描いたアーキテクチャ図を自動的に強制するものです。

## 例

設定:

```toml
[[architecture.layers]]
name = "presentation"
packages = ["api", "handlers"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db"]

[[architecture.rules]]
from = "presentation"
allow = ["application"]
deny = ["infrastructure"]
```

違反するコード:

```python
# myapp/api/orders.py  (presentation)
from myapp.repositories.orders import OrderRepository   # ← 禁止

def list_orders():
    return OrderRepository().all()
```

`presentation` が `application` を飛び越えて直接 `infrastructure` にアクセスしています。

## 修正例

アプリケーション層を経由して呼び出すようにします:

```python
# myapp/services/orders.py  (application)
from myapp.repositories.orders import OrderRepository

def list_orders():
    return OrderRepository().all()
```

```python
# myapp/api/orders.py  (presentation)
from myapp.services.orders import list_orders

def get():
    return list_orders()
```

`api` は `services` にのみ依存するようになり、プレゼンテーション層に触れることなくインフラストラクチャを差し替えられます。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`[[architecture.layers]]`](../configuration/reference.md#architecture) | -- | レイヤーと各レイヤーに属するパッケージフラグメントを定義します。 |
| [`[[architecture.rules]]`](../configuration/reference.md#architecture) | -- | `from` / `allow` / `deny` / オプションで各ルールの `severity`。 |
| [`architecture.validate_layers`](../configuration/reference.md#architecture) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`architecture.strict_mode`](../configuration/reference.md#architecture) | `true` | strict モードでは、明示的に許可されていないものはすべて拒否されます。 |
| [`architecture.fail_on_violations`](../configuration/reference.md#architecture) | `false` | 違反が検出された場合にゼロ以外の終了コードを返します。 |

レイヤーが設定されていない場合、アナライザーは許容モードで動作し、このルールは検出結果を生成しません。

## 参照

- レイヤー解決とルール評価 (`internal/analyzer/module_analyzer.go`)。
- [ルールカタログ](index.md) · [low-package-cohesion](low-package-cohesion.md)
