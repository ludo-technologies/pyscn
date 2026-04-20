# `pyscn init`

すべてのオプションがインラインでドキュメント化された `.pyscn.toml` 設定ファイルを生成します。

```text
pyscn init [flags]
```

## 動作内容

よく調整されるセクションをコメント付きの TOML ファイルとして出力します:

- `[output]`、`[complexity]`、`[dead_code]`、`[clones]`、`[cbo]`、`[analysis]`、`[architecture]`（`[[architecture.layers]]` と `[[architecture.rules]]` の例を含む）
- デフォルト値が設定済み
- 各キーの説明コメント付き

生成されるファイルにはすべての設定セクションが含まれるわけではありません。LCOM4 凝集度（`[lcom]`）、モジュール依存関係解析（`[dependencies]`）、モックデータ検出（`[mock_data]`）、DI アンチパターン（`[di]`）のオプションは有効ですが、手動で追加する必要があります。すべてのキーについては [Configuration Reference](../configuration/reference.md) を参照してください。

ファイルが作成されると、以降のこのプロジェクト（またはサブディレクトリ）での `pyscn analyze` / `pyscn check` の実行時に自動的に読み込まれます。

## フラグ

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `-c, --config <path>` | `.pyscn.toml` | 出力ファイルのパス。 |
| `-f, --force`         | off          | 既存のファイルを上書きします。 |

## 終了コード

| コード | 意味 |
| --- | --- |
| `0` | ファイルが正常に書き込まれました。 |
| `1` | ファイルが既に存在します（`--force` で上書き可能）、または書き込みに失敗しました。 |

## 使用例

```bash
# Create .pyscn.toml in the current directory
pyscn init

# Use a custom filename
pyscn init --config tools/pyscn.toml

# Overwrite an existing config
pyscn init --force
```

## まず調整すべき設定

`init` を実行した後、多くのプロジェクトで調整される設定は以下のとおりです:

| 設定 | よくある調整 |
| --- | --- |
| `[complexity].max_complexity` | CI の厳しさに応じて `10`、`15`、`20` のいずれかに設定します。 |
| `[dead_code].min_severity`     | 警告が多すぎる場合は `"critical"` に引き上げます。 |
| `[clones].similarity_threshold`| クローンをより多く検出するには `0.80` に下げ、ノイズを減らすには `0.90` に上げます。 |
| `[analysis].exclude_patterns`  | 生成コードのパス、マイグレーションなどを追加します。 |

詳しくは [Configuration Reference](../configuration/reference.md) をご覧ください。

## 関連項目

- [Configuration Reference](../configuration/reference.md) — すべてのオプションの説明。
- [Configuration Examples](../configuration/examples.md) — 厳格な CI、大規模コードベース、最小限のオーバーライドの例。
