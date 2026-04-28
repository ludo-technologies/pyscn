# low-class-cohesion

**カテゴリ**: クラス設計  
**重大度**: しきい値により設定可能  
**検出コマンド**: `pyscn analyze`, `pyscn check`

## 検出内容

メソッド間でインスタンス状態を共有していないクラスを検出します（LCOM4 メトリクス — Lack of Cohesion of Methods, バージョン4）。pyscn は、共通の `self.` 属性にアクセスする2つのメソッドを接続するグラフを構築し、連結成分の数をカウントします。`LCOM4 = 1` はすべてのメソッドが互いに関連していることを意味し、`LCOM4 = N` はそのクラスが実質的に `N` 個の無関係なサブクラスを1つにまとめたものであることを意味します。

`@staticmethod` または `@classmethod` で装飾されたメソッドは `self` を参照しないため、グラフから除外されます。

簡単に言えば: *このクラスは無関係な仕事をしています — 分割するか、関数のモジュールにしましょう。*

## なぜ問題なのか

クラスは状態とそれに対する操作をまとめるためのものです。メソッドが同じ状態にアクセスしない場合:

- **クラス名が嘘をつく** — 1つのものであると主張しながら、2つや3つのように振る舞います。
- **変更が分散する** — ある責務のバグを見つけるために、それとは無関係なコードを読む必要が生じます。
- **再利用が阻害される** — 必要な部分だけを取り出すことができず、残りも引きずることになります。
- **真の抽象化の前身であることが多い** — 「Utilities」や「Manager」クラスは典型的な症状です。

## 例

```python
class UserUtility:
    def __init__(self, db, smtp, clock):
        self.db = db
        self.smtp = smtp
        self.clock = clock
        self.cache = {}

    # --- 永続化 ---
    def load(self, user_id):
        if user_id in self.cache:
            return self.cache[user_id]
        row = self.db.fetch("users", user_id)
        self.cache[user_id] = row
        return row

    def save(self, user):
        self.db.upsert("users", user)
        self.cache[user.id] = user

    # --- メール ---
    def send_welcome(self, address):
        self.smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self.smtp.send(address, f"Reset: {token}")

    # --- フォーマット ---
    def format_joined_at(self, user):
        return self.clock.format(user.joined_at)
```

`LCOM4 = 3`: `{load, save}` は `db` と `cache` を共有し、`{send_welcome, send_reset}` は `smtp` を共有し、`{format_joined_at}` は単独です。3つの成分、1つのクラス。

## 修正例

凝集度の高いクラスに分割し、ステートレスな部分はフリー関数に移動します。

```python
class UserRepository:
    def __init__(self, db):
        self._db = db
        self._cache = {}

    def load(self, user_id):
        if user_id in self._cache:
            return self._cache[user_id]
        row = self._db.fetch("users", user_id)
        self._cache[user_id] = row
        return row

    def save(self, user):
        self._db.upsert("users", user)
        self._cache[user.id] = user


class UserMailer:
    def __init__(self, smtp):
        self._smtp = smtp

    def send_welcome(self, address):
        self._smtp.send(address, "Welcome")

    def send_reset(self, address, token):
        self._smtp.send(address, f"Reset: {token}")


# user_formatting.py — クラスなし、状態なし
def format_joined_at(user, clock):
    return clock.format(user.joined_at)
```

各クラスの `LCOM4 = 1` となり、フォーマッターは本来あるべき場所に1行の関数として配置されます。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`lcom.low_threshold`](../configuration/reference.md#lcom) | `2` | この値以下のクラスは低リスクとして報告されます。 |
| [`lcom.medium_threshold`](../configuration/reference.md#lcom) | `5` | この値を超えると高リスクとなります。 |

## 参照

- Hitz, M. & Montazeri, B. *Chidamber and Kemerer's Metrics Suite: A Measurement Theory Perspective.* IEEE TSE, 1996 (LCOM4 の定義).
- 実装: `internal/analyzer/lcom.go`。
- [ルールカタログ](index.md) · [high-class-coupling](high-class-coupling.md)
