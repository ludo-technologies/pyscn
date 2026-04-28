# duplicate-code-modified

**Category**: 重複コード  
**Severity**: Info  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## 検出内容

構造の大部分を共有しながら、文の追加、削除、または変更があるコードブロックを検出します（Type-3 クローン、類似度 >= 0.70）。デフォルトでは無効です。`clones.enabled_clone_types` に `"type3"` を追加して有効にしてください。

## なぜ問題なのか

ニアデュプリケートは実際のコードベースで最も一般的な種類のクローンであり、クリーンアップが最も困難なことが多いです。2つの関数がほぼ同じことをしていますが、一方には追加のバリデーションステップがあったり、エラーパスがわずかに異なっていたりします。共有部分は1箇所に存在すべきであり、異なる部分だけが呼び出し元ごとに変化すべきです。

放置すると、変更されたクローンは乖離していきます。バグが一方のコピーで修正されてもう一方では修正されない、新機能が一方に追加されてもう一方では忘れられる、といったことが起こります。レビュアーは、似たように見えるコードが同じ振る舞いをすると信頼できなくなります。

このルールは Info です。これは、正しいリファクタリングが Type-1 や Type-2 クローンほど機械的ではないためです。違いこそが重要なポイントである場合もあり、マージすると明確さを損なうことがあります。検出結果は確認する価値のある候補として扱い、自動的な欠陥とは見なさないでください。

## 例

```python
def export_users_csv(users, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "name", "email"])
        for u in users:
            writer.writerow([u.id, u.name, u.email])

def export_orders_csv(orders, path):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(["id", "total", "status"])
        for o in orders:
            if o.status != "draft":
                writer.writerow([o.id, o.total, o.status])
```

## 修正例

共通のスキャフォールディングを抽出し、ヘッダー、行の形状、フィルターをパラメータ化してください。

```python
def export_csv(rows, path, header, to_row, keep=lambda _: True):
    with open(path, "w") as f:
        writer = csv.writer(f)
        writer.writerow(header)
        for row in rows:
            if keep(row):
                writer.writerow(to_row(row))

def export_users_csv(users, path):
    export_csv(users, path, ["id", "name", "email"],
               lambda u: [u.id, u.name, u.email])

def export_orders_csv(orders, path):
    export_csv(orders, path, ["id", "total", "status"],
               lambda o: [o.id, o.total, o.status],
               keep=lambda o: o.status != "draft")
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type3_threshold`](../configuration/reference.md#clones) | `0.70` | 変更ありとして報告されるための最小類似度。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | `"type3"` を追加するとこのルールが有効になります。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | タイプ別の閾値の前に適用されるグローバルな下限値。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | フラグメントの最小行数。 |
| [`clones.min_nodes`](../configuration/reference.md#clones) | `10` | フラグメントの最小ASTノード数。 |

## 参照

- クローン検出の実装 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [ルールカタログ](index.md) · [同一クローン](duplicate-code-identical.md) · [リネームされたクローン](duplicate-code-renamed.md) · [意味的クローン](duplicate-code-semantic.md)
