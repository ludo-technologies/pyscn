# placeholder-comment

**カテゴリ**: モックデータ  
**重大度**: Info  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

未完了の作業を示すマーカーを含むコメントを検出します: `TODO`、`FIXME`、`XXX`、`HACK`、`BUG`、`NOTE`。

## なぜ問題なのか

ソース内の `# TODO` は、期限もレビュアーもない、作者が将来の自分に対して行った約束です。ほとんどのコードベースでは、解消されるよりも速く蓄積されます。各マーカーは隠れたスコープの一部であり、読者はそれがまだ関連しているか、変更をブロックするか、誰かが実際に追跡しているかを判断しなければなりません。

このルールは、すべてのマーカーがバグであると主張するものではありません。リストを表面化させることで、プロジェクトごとにそれらを解消するか、追跡対象の Issue に変換するか、明示的なポリシーで受け入れるかを判断できるようにします。

## 例

```python
def process_order(order):
    # TODO: handle refunds
    ...
```

## 修正例

作業を実装するか、マーカーを追跡対象の Issue リンクに変換して、意図がクローズ状態を持つ場所に残るようにします。

```python
def process_order(order):
    # Refunds are handled by the billing service: see issue #1423.
    ...
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` に上げると、このルールを除外します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストファイルをスキップします。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。 |

## 参照

- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
