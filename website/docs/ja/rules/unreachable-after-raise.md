# unreachable-after-raise

**Category**: 到達不能コード  
**Severity**: Critical  
**Triggered by**: `pyscn analyze`, `pyscn check`

## 検出内容

同じコードブロック内で `raise` 文の後に記述された文を検出します。

## なぜ問題なのか

`raise` は無条件にスタックを巻き戻します。同じブロック内でその後に続く文は決して実行されません。

これは通常、以下を示します：

- **古いクリーンアップ** -- 例外が送出される前に実行されるべきだったコード。
- **リファクタリングの痕跡** -- `raise` が以前の分岐に置き換わり、周囲の行が残された。
- **ロジックのバグ** -- 作者は `raise` の後も実行が続くと想定していた。

`raise` の後のデッドコードはテストで実行されることがなく、本番環境でも表面化しないため、隠れたバグは無言のままです。

## 例

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
        account.balance -= amount   # ← 実行されない
    account.balance -= amount
```

## 修正例

文を `raise` の前に移動するか、削除してください。

```python
def withdraw(account, amount):
    if amount > account.balance:
        raise InsufficientFunds(account.id)
    account.balance -= amount
```

## オプション

| Option | Default | Description |
| --- | --- | --- |
| [`dead_code.detect_after_raise`](../configuration/reference.md#dead_code) | `true` | `false` に設定するとこのルールを無効にします。 |
| [`dead_code.min_severity`](../configuration/reference.md#dead_code) | `"warning"` | `"critical"` にするとこの種類の検出結果のみ保持します。`"info"` にするとより多く表示します。 |
| [`dead_code.ignore_patterns`](../configuration/reference.md#dead_code) | `[]` | ソース行に対してマッチする正規表現パターン。マッチした場合は抑制されます。 |

## 参照

- 制御フローグラフの到達可能性分析 (`internal/analyzer/dead_code.go`)。
- [ルールカタログ](index.md) · [Unreachable after return](unreachable-after-return.md) · [Unreachable branch](unreachable-branch.md)
