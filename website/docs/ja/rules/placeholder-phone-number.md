# placeholder-phone-number

**カテゴリ**: モックデータ  
**重大度**: Warning  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

明らかに偽のパターンに従う文字列リテラル内の電話番号を検出します: すべてゼロ（`000-0000-0000`）、連番（`123-456-7890`、`012-345-6789`）、または同じ数字の長い繰り返し。

## なぜ問題なのか

プレースホルダーの電話番号は、フォームの最初のドラフトから残り、二度と見直されない類の値です。バリデーションを通過し、フォーマットされ、データベースを往復します。そのため、実際のユーザーが確認画面でそれを見るか、サポート担当者がその番号に電話をかけようとするまで、何も壊れません。

## 例

```python
default_phone = "000-0000-0000"
```

## 修正例

フィールドを空のままにするか、呼び出し元に要求するか、設定から取得します。不明な電話番号は、偽の値ではなく不在であるべきです。

```python
default_phone: str | None = None
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` に上げると、このレベルの検出結果のみを保持します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストファイルをスキップします。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。 |

## 参照

- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [placeholder-uuid](placeholder-uuid.md) · [repetitive-string-literal](repetitive-string-literal.md)
