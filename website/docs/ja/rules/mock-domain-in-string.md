# mock-domain-in-string

**カテゴリ**: モックデータ  
**重大度**: Warning  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

ドキュメントやテスト用に予約されたドメインを含む文字列リテラルを検出します: `example.com`、`example.org`、`example.net`、`test.com`、`localhost`、`invalid`、`foo.com`、`bar.com`、および同様の RFC 2606 / RFC 6761 で定義された名前。

## なぜ問題なのか

これらのドメインは、例やテストが実際のトラフィックと衝突しないように存在しています。ドキュメントを書いている間は便利ですが、リテラルが出荷されると問題になります。本番コード内のハードコードされた `example.com` の URL は、通常以下のいずれかです:

- リリース前に置き換えるべきだったプレースホルダー。
- そもそもハードコードすべきではなかった設定値。

どちらの場合も、失敗モードはサイレントです: リクエストは成功し（ドメインはドキュメントページか何もないページに解決され）、例外は発生せず、誰かがサインアップが届かない理由を尋ねるまでバグに気づきません。

## 例

```python
SIGNUP_URL = "https://example.com/signup"
```

## 修正例

値を設定に移動するか、実際の URL をインラインで指定します。

```python
SIGNUP_URL = settings.signup_url
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。 |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *（RFC 2606 リスト）* | 予約ドメインリストを上書きまたは拡張します。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` に上げると、このレベルの検出結果のみを保持します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストファイルをスキップします。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。 |

## 参照

- RFC 2606: *Reserved Top Level DNS Names.*
- RFC 6761: *Special-Use Domain Names.*
- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [mock-email-address](mock-email-address.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
