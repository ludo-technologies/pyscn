# unreachable-after-return

**Category**: 到達不能コード  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

同じコードブロック内で `return` 文の後に記述された文を検出します。

## なぜ問題なのか

`return` の後に配置されたコードは決して実行されません。これは通常、以下のいずれかです：

- **リファクタリングの残骸** -- `return` が上に移動され、下のコードが削除し忘れられた。
- **バグ** -- プログラマはコードが実行されると期待していたが、制御フローの変更により到達不能になった。
- **誤配置されたクリーンアップ** -- return する前に実行すべき処理。

いずれの場合でもコードはデッドコードです。読む時間を消費し、テストでカバーされず（カバーできないため）、バグが隠れていても、そのバグがユーザーの動作から報告されることはありません。

## 例

```python
def charge(order):
    if order.total <= 0:
        return None
        log.debug("zero-value charge")   # ← 実行されない
    ...
```

## 修正例

文を `return` の前に移動するか、不要であれば削除してください。

```python
def charge(order):
    if order.total <= 0:
        log.debug("zero-value charge")
        return None
    ...
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_return`](../configuration/reference.md#dead_code) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` にするとこの種類の検出結果のみ保持します。`"info"` にするとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable branch](unreachable-branch.md) · [Unreachable after raise](unreachable-after-raise.md)
