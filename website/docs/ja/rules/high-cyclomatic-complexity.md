# high-cyclomatic-complexity

**カテゴリ**: 複雑度  
**重大度**: しきい値により設定可能  
**検出コマンド**: `pyscn analyze`, `pyscn check`

## 検出内容

McCabe サイクロマティック複雑度が設定されたしきい値を超える関数を検出します。`if`、`elif`、`for`、`while`、`except`、`match case`、および内包表記内のブール節がそれぞれカウントに1を加算します。直線的な関数の初期値は1です。

pyscn は `and` / `or` 短絡演算子を個別の分岐としてカウントしません。

## なぜ問題なのか

分岐数が多いということは、以下を意味します:

- **読むべきパスが増える** — レビュアーは関数の動作を理解するために、すべての分岐を頭の中でシミュレーションしなければなりません。
- **テストすべきパスが増える** — 完全な分岐カバレッジにはパスごとに1つのテストが必要であり、実際には分岐の多い関数のほとんどがテスト不足です。
- **欠陥密度が高くなる** — McCabe (1976) 以降の実証研究で、複雑度とバグ率の相関が示されています。
- **安全な変更が困難になる** — ある分岐での小さな修正が、別の分岐を気づかないうちに壊す可能性があります。

複雑度が約10を超える関数は、通常、名前を付けて分離できる複数の役割を担っています。

## 例

```python
def price_for(user, cart, coupon, region):
    total = 0
    for item in cart:
        if item.category == "book":
            if region == "EU":
                total += item.price * 0.95
            elif region == "US":
                total += item.price
            else:
                total += item.price * 1.10
        elif item.category == "food":
            if user.is_student:
                total += item.price * 0.90
            else:
                total += item.price
        else:
            total += item.price
    if coupon:
        if coupon.kind == "percent":
            total *= 1 - coupon.value
        elif coupon.kind == "fixed":
            total -= coupon.value
    if total < 0:
        total = 0
    return total
```

サイクロマティック複雑度: 13。

## 修正例

商品ごとの価格計算とクーポン処理を抽出し、ネストした条件分岐をディスパッチテーブルに置き換えます。

```python
REGION_BOOK_MULTIPLIER = {"EU": 0.95, "US": 1.00}

def _book_price(item, region):
    return item.price * REGION_BOOK_MULTIPLIER.get(region, 1.10)

def _food_price(item, user):
    return item.price * (0.90 if user.is_student else 1.00)

PRICERS = {"book": _book_price, "food": _food_price}

def _item_price(item, user, region):
    pricer = PRICERS.get(item.category)
    return pricer(item, region) if item.category == "book" else \
           pricer(item, user)   if item.category == "food" else \
           item.price

def _apply_coupon(total, coupon):
    if coupon is None:
        return total
    if coupon.kind == "percent":
        return total * (1 - coupon.value)
    return total - coupon.value

def price_for(user, cart, coupon, region):
    subtotal = sum(_item_price(i, user, region) for i in cart)
    return max(0, _apply_coupon(subtotal, coupon))
```

各ヘルパー関数の複雑度は1〜3で、単一の責任を持ちます。ガード節（`if coupon is None: return`）により、残りの分岐がフラットになります。

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`complexity.max_complexity`](../configuration/reference.md#complexity) | `0` | `pyscn check` で適用されるハードリミット。`0` は `analyze` での制限なしを意味します。`pyscn check --max-complexity` は未設定の場合 `10` を使用します。 |
| [`complexity.low_threshold`](../configuration/reference.md#complexity) | `9` | この値以下の関数は低リスクとして報告されます。 |
| [`complexity.medium_threshold`](../configuration/reference.md#complexity) | `19` | この値を超えると高リスクとなります。 |
| [`complexity.min_complexity`](../configuration/reference.md#complexity) | `1` | この値未満の関数はレポートから除外されます。 |

## 参照

- McCabe, T. J. *A Complexity Measure.* IEEE Transactions on Software Engineering, 1976.
- 制御フローグラフの構築とサイクロマティック複雑度の計算: `internal/analyzer/complexity.go`, `internal/analyzer/complexity_analyzer.go`, `internal/analyzer/cfg_builder.go`。
- [ルールカタログ](index.md) · [too-many-constructor-parameters](too-many-constructor-parameters.md)
