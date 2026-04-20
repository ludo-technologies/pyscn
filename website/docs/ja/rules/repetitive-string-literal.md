# repetitive-string-literal

**カテゴリ**: モックデータ  
**重大度**: Info  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

長さ4から20の、非常に反復的な文字パターンを持つ文字列リテラルを検出します: `aaaa`、`1111`、`xxxxxxxx`、および同様の単一文字または2文字の繰り返し。

## なぜ問題なのか

`"aaaaaaaaaaaaaaaa"` のような文字列が実際の値であることはほぼありません。これは、開発者が他の何かを配線している間にバリデーターを通過させるために入力した形のものです。本番に残ると、データに見え、長さチェックを通過するが意味を持たない API キー、ハッシュ入力、またはトークンになります。

このルールは長さ制限付き（4--20文字）で、パディング定数や反復文字を正当に必要とするテストベクターなどの意図的な埋め込みを検出しないようにしています。

## 例

```python
api_key = "aaaaaaaaaaaaaaaa"
```

## 修正例

シークレットやトークンは設定またはシークレットストアから読み取ります。プレースホルダー値をソースに埋め込まないでください。

```python
import os

api_key = os.environ["SERVICE_API_KEY"]
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
- [ルールカタログ](index.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-uuid](placeholder-uuid.md)
