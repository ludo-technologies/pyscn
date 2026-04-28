# mock-keyword-in-code

**カテゴリ**: モックデータ  
**重大度**: Info（文字列内） / Warning（識別子内）  
**トリガー**: `pyscn check --select mockdata`

## 検出内容

一般的なプレースホルダーキーワードを含む識別子および文字列リテラルを検出します: `mock`, `fake`, `dummy`, `test`, `sample`, `example`, `placeholder`, `stub`, `fixture`, `temp`, `foo`, `bar`, `baz`, `lorem`, `ipsum`。

## なぜ問題なのか

これらの単語は、まだ何かを検討中のときに入力するものです。ノートブック、テストファイル、5分間のスパイクでは問題ありませんが、本番コードに残ってしまった場合は、通常スタブが置き換えられなかったことを意味します。チェックインされたモジュール内の `foo` という名前は、作者が出荷しようとしたものではほぼありません。

**識別子**内の一致は、束縛された名前（`foo = get_user()`）が動作を変えるため、警告として扱われます。**文字列リテラル**内の一致は、残された `"fake_user"` が壊れているというよりは見た目の問題であることが多いため、情報レベルですが、リリース前にレビューする価値はあります。

## 例

```python
def create_user():
    name = "fake_user"    # 文字列リテラルが一致
    foo = get_user()      # 識別子 `foo` が一致
    return foo
```

## 修正例

プレースホルダーを削除します。実際のデータを使用するか、設定から読み取るか、スタブをテストフィクスチャに移動して適切な場所に配置します。

```python
def create_user(name: str):
    user = get_user(name)
    return user
```

## オプション

| オプション | デフォルト | 説明 |
| --- | --- | --- |
| [`mock_data.enabled`](../configuration/reference.md#mock_data) | `false` | オプトイン。カテゴリ全体がデフォルトで無効です。 |
| [`mock_data.keywords`](../configuration/reference.md#mock_data) | *（組み込みリスト）* | このルールのキーワードリストを上書きします。 |
| [`mock_data.min_severity`](../configuration/reference.md#mock_data) | `"info"` | `"warning"` にすると識別子の一致のみを保持します。 |
| [`mock_data.ignore_tests`](../configuration/reference.md#mock_data) | `true` | テストと思われるファイルをスキップします。 |
| [`mock_data.ignore_patterns`](../configuration/reference.md#mock_data) | `[]` | ファイルパスに対してマッチする正規表現パターン。一致した場合は抑制されます。 |

## 参照

- 実装: `internal/analyzer/mock_data_detector.go`。
- [ルールカタログ](index.md) · [mock-domain-in-string](mock-domain-in-string.md) · [test-credential-in-code](test-credential-in-code.md) · [placeholder-comment](placeholder-comment.md)
