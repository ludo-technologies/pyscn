# mock-email-address

**カテゴリ**: モックデータ  
**重大度**: Warning  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

ドメイン部分がテスト用に予約されたメールアドレスを検出します: `test@example.com`、`admin@test.com`、`foo@localhost`、および同様のもの。

## なぜ問題なのか

アプリケーションコード内のテストドメインのメールアドレスは、ほぼ確実にフィクスチャ、チュートリアル、または「後で記入する」スタブの残りです。不正なアドレスとは異なり、バリデーションとシリアライゼーションを正常に通過するため、データベースの行、通知キュー、「from」ヘッダーに静かに紛れ込みます。通常、パスワードリセットリンクが届かない理由を尋ねるサポートチケットで発見されます。

このルールは [mock-domain-in-string](mock-domain-in-string.md) を補完しますが、ドメインリストを絞り込みマッチを正確にするため、メール形式を特別に扱います。

## 例

```python
admin_email = "admin@example.com"
```

## 修正例

アドレスを設定から読み取るか、パラメータとして受け取ります。

```python
admin_email = settings.admin_email
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。 |
| [`mock_data.domains`](../configuration/reference.md#mock_data) | *（RFC 2606 リスト）* | プレースホルダーとみなすドメイン。`mock-domain-in-string` と共有されます。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` に上げると、このレベルの検出結果のみを保持します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストファイルをスキップします。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。 |

## 参照

- RFC 2606: *Reserved Top Level DNS Names.*
- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md)
