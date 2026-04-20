# concrete-type-hint-dependency

**カテゴリ**: 依存性注入  
**重大度**: Info  
**検出コマンド**: `pyscn analyze`, `pyscn check --select di`

## 検出内容

`__init__` パラメータの型ヒントが `Protocol`、抽象基底クラス、またはインターフェースではなく具象クラスである場合を検出します。

## なぜ問題なのか

具象の型ヒントは、読み手と型チェッカーに対して、このクラスが特定の1つの実装のみを受け入れることを伝えます。ランタイムではダックタイピングによる代替を問題なく受け入れますが、宣言された契約はそうではなく、ヒントに従うツール（mypy、IDE の自動補完、型から構築されるテストモックなど）は代替を拒否します。

実際にはテストが書きにくくなります。インメモリのフェイクを代替として使いたいテストは、具象クラスを継承するか（その全ての動作を引き継ぐことになる）、型エラーを抑制する必要があります。また、具象クラスが必要とするインポートにコンシューマーが結び付けられるため、小さなユーティリティがデータベーススタック全体を引き込むことになりかねません。

`Protocol` や抽象インターフェースに依存することで、クラスが実際に使用するもの — 1つや2つのメソッド — が文書化され、フェイク、アダプタ、将来の実装の余地が残ります。

## 例

```python
class SqlUserRepository:
    def find(self, user_id): ...

class UserService:
    def __init__(self, repo: SqlUserRepository):
        self.repo = repo
```

## 修正例

依存するメソッドを記述する `Protocol` を宣言し、それに依存します。

```python
from typing import Protocol

class UserRepository(Protocol):
    def find(self, user_id: str) -> User: ...

class UserService:
    def __init__(self, repo: UserRepository):
        self.repo = repo
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`di.enabled`](../configuration/reference.md#di) | `false` | `analyze` で DI ルールを実行するには `true` にする必要があります。 |
| [`di.min_severity`](../configuration/reference.md#di) | `"warning"` | このルールは `info` で報告されます。表示するには `min_severity` を `"info"` に下げてください。 |

## 参照

- 具象依存の検出 (`internal/analyzer/concrete_dependency_detector.go`)。
- [ルールカタログ](index.md) · [Concrete instantiation dependency](concrete-instantiation-dependency.md) · [Too many constructor parameters](too-many-constructor-parameters.md)
