# module-variable-dependency

**カテゴリ**: 依存性注入  
**重大度**: Warning  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

`global` 文を使わずにモジュールレベルのミュータブル変数を直接読み書きするクラスを検出します。これはモジュール状態への暗黙的な結合です。

## なぜ問題なのか

`global` による代入とは異なり、モジュールレベルの名前の読み取りはサイレントです。クラスは自己完結しているように見えますが、実際の動作は呼び出し時にそのモジュール変数に格納されている値に依存します。クラスを単独でインスタンス化するテストが、無関係なインポートによって成功したり失敗したりする可能性があります。

また、代替可能性も損なわれます。モジュールをモンキーパッチしない限り、クラスに別の協調オブジェクトを与えることはできず、これは脆弱で順序に依存します。クラスの2つのインスタンスは、望むと望まざるとにかかわらず、同じバッキングオブジェクトを共有することを強いられます。

協調オブジェクトをコンストラクタパラメータにすることで、依存が文書化され、インスタンスごとの設定が可能になり、テストではモジュールグローバルに触れることなくフェイクを注入できます。

## 例

```python
config = load_config()

class UserRepository:
    def find(self, user_id):
        conn = connect(config.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## 修正例

協調オブジェクトをコンストラクタパラメータとして受け取ります。

```python
class UserRepository:
    def __init__(self, database_url):
        self.database_url = database_url

    def find(self, user_id):
        conn = connect(self.database_url)
        return conn.query("SELECT * FROM users WHERE id = ?", user_id)
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | `"error"` に上げるとこのルールは非表示になります。 |

## 参照

- 隠れた依存の検出 (`internal/analyzer/hidden_dependency_detector.go`)。
- [ルールカタログ](index.md) · [Global state dependency](global-state-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
