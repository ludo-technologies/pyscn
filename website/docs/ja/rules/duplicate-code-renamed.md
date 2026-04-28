# duplicate-code-renamed

**Category**: 重複コード  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check --select clones`

## 検出内容

同じ構造を持ちながら、識別子やリテラルが異なるコードブロックを検出します（Type-2 クローン、類似度 >= 0.75）。

## なぜ問題なのか

リネームされたクローンは、誰かが関数をコピーし、変数名を検索置換して先に進んだ場合に発生します。構造は同一で、名詞だけが変わっています。保守コストはテキストが完全に一致するクローンと同じですが、単語が異なるため目視では重複を見つけにくくなります。

これはまた、元のコードがパラメータ化されるべきだったのにされていないことを示すシグナルでもあります。変化する部分（型、フィールド名、定数）は自然な関数の引数です。

## 例

```python
def total_for_orders(orders):
    total = 0
    for order in orders:
        if order.status == "paid":
            total += order.amount
    return total

def total_for_invoices(invoices):
    total = 0
    for invoice in invoices:
        if invoice.status == "settled":
            total += invoice.amount
    return total
```

## 修正例

変化する述語やフィールドアクセサをパラメータとして受け取る汎用ヘルパーを抽出してください。

```python
def total_where(items, is_active):
    return sum(item.amount for item in items if is_active(item))

def total_for_orders(orders):
    return total_where(orders, lambda o: o.status == "paid")

def total_for_invoices(invoices):
    return total_where(invoices, lambda i: i.status == "settled")
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`clones.type2_threshold`](../configuration/reference.md#clones) | `0.75` | リネームとして報告されるための最小類似度。 |
| [`clones.similarity_threshold`](../configuration/reference.md#clones) | `0.65` | タイプ別の閾値の前に適用されるグローバルな下限値。 |
| [`clones.ignore_identifiers`](../configuration/reference.md#clones) | `true` | 類似度計算時に異なる変数名を同等として扱います。 |
| [`clones.ignore_literals`](../configuration/reference.md#clones) | `true` | 異なる数値リテラルや文字列リテラルを同等として扱います。 |
| [`clones.min_lines`](../configuration/reference.md#clones) | `5` | フラグメントの最小行数。 |
| [`clones.enabled_clone_types`](../configuration/reference.md#clones) | `["type1","type2","type4"]` | `"type2"` を含めることでこのルールを有効にします。 |

## 参照

- クローン検出の実装 (`internal/analyzer/clone_detector.go`, `internal/analyzer/apted.go`)。
- [ルールカタログ](index.md) · [同一クローン](duplicate-code-identical.md) · [変更されたクローン](duplicate-code-modified.md) · [意味的クローン](duplicate-code-semantic.md)
