# global-state-dependency

**カテゴリ**: 依存性注入  
**重大度**: Error  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

モジュールレベルの状態を読み取りまたは変更するために `global` 文を使用するクラスメソッドを検出します。

## なぜ問題なのか

メソッド内の `global` は、クラスのインターフェースからは見えない特定のモジュール変数にクラスを結び付けます。`OrderService(...)` のどこにも、インスタンスの構築だけでは不十分であること — 何らかのモジュールレベルの値も事前に設定されている必要があること — を読み手に伝えるものはなく、メソッドが予期しない動作をする可能性があります。

最も影響を受けるのはテストです。グローバル状態にアクセスするメソッドをテストするには、各テストがモジュールに手を伸ばし、古い値を保存し、新しい値を設定し、終了時に復元する必要があります。クリーンアップを忘れたテストは、次のテストに状態をリークさせます。テストの並列実行も安全ではなくなります。

依存は実在しますが、隠されているだけです。明示的なコンストラクタパラメータにすることで、驚きを排除できます。

## 例

```python
_current_user = None

class AuditLog:
    def record(self, action):
        global _current_user
        entry = {"user": _current_user, "action": action}
        db.insert("audit", entry)
```

## 修正例

`__init__` を通じて値を渡すことで、依存を可視化しスワップ可能にします。

```python
class AuditLog:
    def __init__(self, user):
        self.user = user

    def record(self, action):
        db.insert("audit", {"user": self.user, "action": action})
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | このルールは `error` で報告されます。`min_severity` を `error` より上に設定しない限り表示されます。 |

## 参照

- 隠れた依存の検出 (`internal/analyzer/hidden_dependency_detector.go`)。
- [ルールカタログ](index.md) · [Module variable dependency](module-variable-dependency.md) · [Singleton pattern dependency](singleton-pattern-dependency.md)
