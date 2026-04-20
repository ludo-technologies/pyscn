# placeholder-uuid

**カテゴリ**: モックデータ  
**重大度**: Warning  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

エントロピーが非常に低い UUID 形式の文字列リテラルを検出します: nil UUID（`00000000-0000-0000-0000-000000000000`）、全て1、全て `f`、または UUID としてパースできるものの同じ文字の長い繰り返しで構成されているもの。

## なぜ問題なのか

nil UUID はいくつかのコンテキストでは正当な値ですが、ほとんどのアプリケーションコードでは、置き換えられるべきだったスタブ `DEFAULT_USER_ID = "00..."` の残りです。他の UUID と同様にパースおよびシリアライズされるため、外部キーの検索、ログ行、監査証跡のすべてがそれを受け入れ、エラーが発生することなく行が同じ「ユーザー」に集約されていきます。

## 例

```python
DEFAULT_USER_ID = "00000000-0000-0000-0000-000000000000"
```

## 修正例

使用時に実際の UUID を生成するか、呼び出し元に提供を要求します。データに見えるセンチネル値を持ち歩かないでください。

```python
import uuid

def new_user_id() -> uuid.UUID:
    return uuid.uuid4()
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
- [ルールカタログ](index.md) · [placeholder-phone-number](placeholder-phone-number.md) · [repetitive-string-literal](repetitive-string-literal.md)
