# 設定ファイル形式

pyscn は **TOML** 形式で設定を読み込みます。専用の `.pyscn.toml` ファイル、または既存の `pyproject.toml` 内の `[tool.pyscn]` セクションに設定を記述できます。

## ファイル探索

`pyscn analyze` や `pyscn check` を実行すると、pyscn はターゲットパスから上位ディレクトリに向かって以下を検索します:

1. `.pyscn.toml`（最優先）
2. `[tool.pyscn]` セクションを含む `pyproject.toml`

最初に見つかったファイルが使用されます。一致するファイルが見つかるか、ファイルシステムのルートに到達するまで親ディレクトリが検索されます。どちらのファイルも見つからない場合は、組み込みデフォルトが使用されます。

明示的にパスを指定することもできます:

```bash
pyscn analyze --config ./configs/strict.toml src/
```

これにより探索はスキップされます。

## 優先順位

設定が複数の場所に存在する場合、後のものが優先されます:

1. **組み込みデフォルト**（最低）
2. **`pyproject.toml` → `[tool.pyscn]`**
3. **`.pyscn.toml`**
4. **CLI フラグ**（最高）

CLI フラグは **明示的に設定された** 場合のみ考慮されます。変更されていないデフォルト値は設定値を上書きしません。

## 2つのファイル形式

=== ".pyscn.toml"

    ```toml
    [complexity]
    max_complexity = 15

    [dead_code]
    min_severity = "critical"
    ```

=== "pyproject.toml"

    ```toml
    [tool.pyscn.complexity]
    max_complexity = 15

    [tool.pyscn.dead_code]
    min_severity = "critical"
    ```

同じディレクトリに両方のファイルが存在する場合、`.pyscn.toml` が優先されます。

## スターターファイルの生成

```bash
pyscn init
```

すべてのオプション、デフォルト値、簡単な説明を含むコメント付きの `.pyscn.toml` が生成されます。必要な値を編集し、残りは削除するか（そのままにしておいても構いません）。

```bash
pyscn init --force   # 既存ファイルを上書き
pyscn init --config tools/pyscn.toml   # カスタムパス
```

## バリデーション

pyscn は設定の読み込み時にバリデーションを行い、問題がある場合は終了コード `2` で終了します。主なバリデーションルール:

- 複雑度の閾値は `low ≥ 1` かつ `medium > low` を満たす必要があります。
- 出力形式は `text`, `json`, `yaml`, `csv`, `html` のいずれかです。
- デッドコードの重大度は `info`, `warning`, `critical` のいずれかです。
- クローン類似度の閾値は `[0.0, 1.0]` の範囲内である必要があります。
- 少なくとも1つの include パターンが指定されている必要があります。

## 環境変数

pyscn は環境変数から設定を読み込み**ません**。MCP サーバーは例外として、`PYSCN_CONFIG` で設定ファイルを指定できます。

## 次のステップ

- [リファレンス](reference.md) — すべてのキーのドキュメント。
- [設定例](examples.md) — 厳格な CI、大規模コードベース、最小限のオーバーライド。
