# unreachable-branch

**Category**: 到達不能コード  
**Severity**: Warning  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

先行するすべての分岐が `return`、`raise`、`break`、または `continue` で終了するために、到達できない `if`、`elif`、または `else` 分岐を検出します。

## なぜ問題なのか

先行する各分岐がすでに関数やループを終了している場合、残りの分岐は論理的にデッドコードです。ガード条件は読者にとって意味があるように見えるため、実際の制御フローが隠されます。

これは通常、以下を示します：

- **冗長な条件** -- `if` がかつてフォールスルーしていたために `else` が存在している。
- **微妙なバグ** -- 作者は後の分岐がある条件で実行されると期待していたが、先行する exit により不可能になった。
- **古い防御コード** -- もはや到達できないフォールバック。

テストはその分岐をカバーできず、レビュアーは決して実行されないパスについて推論する時間を浪費します。

## 例

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    elif payment.amount == 0:
        return "empty"
    else:
        return "normal"
    return "unknown"   # ← 到達不能な分岐
```

## 修正例

デッドブランチを削除するか、フォールバックが実際に到達可能になるように先行する分岐を再構成してください。

```python
def classify(payment):
    if payment.amount < 0:
        raise ValueError("negative amount")
    if payment.amount == 0:
        return "empty"
    return "normal"
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_unreachable_branches`](../configuration/reference.md#dead_code) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` に上げるとこれらの検出結果を非表示にします。`"info"` に下げるとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable after raise](unreachable-after-raise.md)
