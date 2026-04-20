# よくある質問

## 全般

### pyscn は ruff、pylint、mypy とどう違いますか?

ruff と pylint はリンター、mypy は型チェッカーです。pyscn は構造的解析ツールで、制御フローグラフ・ツリー表現・依存グラフを構築し、複雑度・到達可能性・コード重複・結合度を計測します。これらのツールとは相互補完の関係にあります。

### pyscn が検出しないものは何ですか?

- ランタイムバグ
- セキュリティ脆弱性
- パフォーマンス問題
- スタイル違反（ruff を使ってください）
- 型エラー（mypy / pyright を使ってください）

### pyscn はネットワークアクセスが必要ですか?

いいえ。pyscn は完全にローカルで動作します。テレメトリやリモート通信は一切ありません。

### 非同期コードに対応していますか?

はい。`async def` と `await` は同期コードと同じように解析されます。

### Jupyter notebook を解析できますか?

できません。先に `jupyter nbconvert --to script` で変換してください。

## 設定

### 設定ファイルはどこに置くべきですか?

`.pyscn.toml` をリポジトリのルートに置いてください。pyscn は解析対象のファイルから親ディレクトリをたどって設定ファイルを探します。

### `pyproject.toml` があります。`[tool.pyscn]` を使うべきですか?

どちらでも構いません。両方が存在する場合は `.pyscn.toml` が優先されます。

### 生成コード / マイグレーション / ベンダー依存を除外するには?

```toml
[analysis]
exclude_patterns = [
  "**/migrations/**",
  "**/__generated__/**",
  "vendor/**",
]
```

### プロジェクトの部分ごとに異なるしきい値を設定できますか?

1 つの設定ファイルではできません。ディレクトリごとに設定ファイルを分けてください:

```bash
pyscn check --config backend/.pyscn.toml backend/
pyscn check --config scripts/.pyscn.toml scripts/
```

## pyscn の実行

### HTML レポートがブラウザで開かれません。

標準入力が TTY でない場合、SSH 接続時、または環境変数 `CI` が設定されている場合は自動オープンが無効になります。レポートのパスは stderr に出力されます。`--no-open` で明示的に制御できます。

### `pyscn analyze` がリポジトリで遅いです。

- `--skip-clones`（クローン検出は最も時間がかかる解析です）
- 対象を絞る: `pyscn analyze src/`
- `[clones]` の `min_lines` / `min_nodes` を引き上げる
- `[clones]` の `max_goroutines` を増やす

### 正しい Python コードなのにパースエラーが出ます。

pyscn は tree-sitter を使用しており、Python 3.13 までサポートしています。最小の再現コードを添えて Issue を報告してください。

### pyscn は問題を自動修正できますか?

できません。pyscn はレポートのみを行い、ソースコードを変更することはありません。

## スコアとしきい値

### ヘルススコアが突然下がりました。

カテゴリスコアを確認してください。よくある原因:

- 大きな関数が追加され、平均複雑度が上がった。
- リファクタリングでデッドコードが残った。
- コピー&ペーストでクローンが発生した。
- 新しいインポートが循環依存を生んだ。

実行間の JSON 出力を差分比較すると、変化の原因を特定できます。

### 小規模なコードベースで結合度スコアが低いのはなぜですか?

ペナルティはパーセンテージベースのため、3/10 の問題比率は 300/1000 と同じペナルティになります。クラスが 20 未満のプロジェクトでは、カテゴリスコアよりも CBO の生値を確認してください。

### 「良い」ヘルススコアとは?

| 範囲 | 意味 |
| --- | --- |
| 90 以上 | 優秀 |
| 70〜90 | 健全なコードベースの一般的な範囲 |
| 50〜70 | 改善が必要だが対処可能 |
| 50 未満 | 集中的なリファクタリングが必要 |

絶対値よりもトレンド（推移）が重要です。

## MCP

### アシスタントが pyscn のツールを認識しているのに呼び出しが失敗します。

バイナリが PATH 上にあることを確認するか、`uvx pyscn-mcp` を使ってください。直接テストするには:

```bash
uvx pyscn-mcp
```

詳しくは [MCP guide](integrations/mcp.md) をご覧ください。

### MCP サーバーでコードのリファクタリングはできますか?

できません。pyscn の MCP は読み取り専用です。

## トラブルシューティング

### pip でインストールしたのに `pyscn: command not found` と出ます。

インストール先が PATH に含まれていません。以下で確認してください:

```bash
python -m pip show -f pyscn | grep bin/pyscn
```

Linux/macOS では `~/.local/bin` を PATH に追加するか、`uvx pyscn@latest <command>`（インストール不要）、`uv tool install pyscn`、`pipx install pyscn` のいずれかを使用してください。

### レポートに「解析ファイル数 0」と表示されます。

include/exclude パターンですべてのファイルが除外されています。デフォルトは `include_patterns = ["**/*.py"]` で、除外パターンには `test_*.py` や `*_test.py` が含まれます。テストも解析するには以下のように設定を上書きしてください:

```toml
[analysis]
exclude_patterns = [
  "**/__pycache__/*",
  "**/*.pyc",
  ".venv/**",
]
```

### `Warning: parse error in <file>` と表示されます。

tree-sitter が回復できない構文エラーがファイルにあります。そのファイルはスキップされますが、他のファイルは正常に解析されます。

## ヘルプ

- [GitHub Issues](https://github.com/ludo-technologies/pyscn/issues) — バグ報告と機能リクエスト。
- [GitHub Discussions](https://github.com/ludo-technologies/pyscn/discussions) — 質問とアイデア。
- [ソースコード](https://github.com/ludo-technologies/pyscn)。
