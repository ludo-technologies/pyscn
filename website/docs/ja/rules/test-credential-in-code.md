# test-credential-in-code

**カテゴリ**: モックデータ  
**重大度**: Warning  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

明らかにプレースホルダーの認証情報に見える文字列リテラルを検出します: `password123`、`secret123`、`testpassword`、`token0`、`api_key_test`、および認証情報風の単語に単純な接尾辞を付けた類似パターン。

## なぜ問題なのか

これらは本物のシークレットではありません。それが問題なのです。クライアントのセットアップ中、最初のテストの作成中、または必須フィールドの入力中に打ち込まれた値です。チェックインされると、2つの失敗モードがあります:

- **恥ずかしさ / 正確性**: リテラルがデフォルトとして使用されてユーザーに出荷され、「デフォルト管理者パスワード」が本当に `password123` になります。
- **ローテーションの見落とし**: リリース前にプレースホルダーを本物のシークレットに置き換えるはずだったのに、置き換えられなかったことに誰も気づきません。

pyscn はセキュリティスキャナーではなく、高エントロピーのシークレットの検出は試みません。このルールは逆のケースを捕捉します: そもそもソースに含まれるべきではなかった、低エントロピーの明らかに偽の認証情報です。

## 例

```python
DEFAULT_PASSWORD = "password123"
```

## 修正例

認証情報は環境変数またはシークレットマネージャーから読み取ります。ローカル開発用にデフォルト値が本当に必要な場合は、本番では読み込まれない別の設定ファイルに保管します。

```python
import os

DEFAULT_PASSWORD = os.environ["APP_DEFAULT_PASSWORD"]
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` に上げると、このレベルの検出結果のみを保持します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストファイルをスキップします。テスト用認証情報はテストに属します。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。 |

## 参照

- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [repetitive-string-literal](repetitive-string-literal.md) · [mock-keyword-in-code](mock-keyword-in-code.md)
