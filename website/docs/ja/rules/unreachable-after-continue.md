# unreachable-after-continue

**Category**: 到達不能コード  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

ループ内で `continue` 文の後に記述された文を検出します。

## なぜ問題なのか

`continue` は次のイテレーションに直接ジャンプします。同じブロック内でその後に続く文はすべてのイテレーションでスキップされるため、一度も実行されません。

典型的な原因：

- **ロジックの並べ替え** -- ガード条件が `continue` に変換され、後続の処理が残された。
- **誤配置された副作用** -- スキップする前に実行すべきカウンタの更新やログ出力。
- **セマンティクスの誤解** -- 作者は `continue` が `pass` のように動作すると期待していた。

文が到達不能なため、テストでカバーできず、意図された動作が無言のまま実行されません。

## 例

```python
for order in orders:
    if order.status == "cancelled":
        continue
        metrics.record_skip(order.id)   # ← 実行されない
    process(order)
```

## 修正例

`continue` の前に文を実行するか、削除してください。

```python
for order in orders:
    if order.status == "cancelled":
        metrics.record_skip(order.id)
        continue
    process(order)
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_continue`](../configuration/reference.md#dead_code) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` にするとこの種類の検出結果のみ保持します。`"info"` にするとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable after break](unreachable-after-break.md) · [Unreachable after return](unreachable-after-return.md)
