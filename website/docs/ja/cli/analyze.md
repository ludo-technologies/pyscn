# `pyscn analyze`

Python ファイルに対してすべての解析を実行し、レポートを生成します。

```text
pyscn analyze [flags] <paths...>
```

`<paths...>` には 1 つ以上のファイルまたはディレクトリを指定します。ディレクトリは設定ファイルの `include_patterns` と `exclude_patterns` に従って再帰的に走査されます。

## 動作内容

デフォルトでは `analyze` は有効なすべての解析を並行して実行します:

- サイクロマティック複雑度
- デッドコード検出
- クローン検出（Type 1〜4）
- クラス結合度（CBO）
- クラス凝集度（LCOM4）
- モジュール依存関係
- アーキテクチャレイヤーの検証

結果は [Health Score](../output/health-score.md) を含む 1 つのレポートにまとめられます。

## フラグ

### 出力フォーマット

1 回の実行で設定できるのは 1 つだけです。指定しない場合は HTML が生成されます。

| フラグ | 説明 |
| ----------- | --- |
| `--html`    | HTML レポートを生成します（デフォルト）。 |
| `--json`    | JSON レポートを生成します。 |
| `--yaml`    | YAML レポートを生成します。 |
| `--csv`     | CSV サマリーを生成します（メトリクスのみ、個別の検出結果は含みません）。 |
| `--no-open` | HTML レポートをブラウザで開きません。 |

出力ファイルはデフォルトで `.pyscn/reports/` に `analyze_YYYYMMDD_HHMMSS.{ext}` の名前で保存されます。ディレクトリは `[output] directory = "..."` で変更できます。

### 解析の選択

| フラグ | 説明 |
| --- | --- |
| `--select <list>` | 指定した解析のみを実行します。カンマ区切り: `complexity,deadcode,clones,cbo,lcom,deps`。 |
| `--skip-complexity` | 複雑度解析をスキップします。 |
| `--skip-deadcode`   | デッドコード検出をスキップします。 |
| `--skip-clones`     | クローン検出をスキップします（最も時間がかかる解析です）。 |
| `--skip-cbo`        | クラス結合度解析をスキップします。 |
| `--skip-lcom`       | クラス凝集度解析をスキップします。 |
| `--skip-deps`       | モジュール依存関係解析をスキップします。 |

`--select` と `--skip-*` は組み合わせて使えます。まず選択が適用され、次にスキップが適用されます。

### しきい値のクイックオーバーライド

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `--min-complexity <N>`    | `5`        | 複雑度が N 以上の関数のみレポートします。 |
| `--min-severity <level>`  | `warning`  | デッドコードの最小深刻度: `info`、`warning`、`critical`。 |
| `--clone-threshold <F>`   | `0.65`     | クローン検出の最小類似度（0.0〜1.0）。 |
| `--min-cbo <N>`           | `0`        | CBO が N 以上のクラスのみレポートします。 |

### 設定

| フラグ | 説明 |
| --- | --- |
| `-c, --config <path>` | `.pyscn.toml` / `pyproject.toml` の自動検出の代わりに、指定したファイルから設定を読み込みます。 |
| `-v, --verbose`        | 詳細な進捗とファイルごとのログを表示します。 |

## 終了コード

| コード | 意味 |
| --- | --- |
| `0` | 解析が完了しました。レポート内の検出結果は終了コードに影響しません。 |
| `1` | 解析が失敗しました — 無効な引数、読み取り不可のファイル、パースエラーなど。 |

`analyze` は検出結果に基づいてプロセスを失敗させることはありません。合否判定が必要な場合は [`pyscn check`](check.md) を使用してください。

## 使用例

```bash
# Full analysis of the current directory with HTML report
pyscn analyze .

# JSON for pipelines
pyscn analyze --json src/

# Skip the slowest analyzer
pyscn analyze --skip-clones src/

# Only complexity and dead code
pyscn analyze --select complexity,deadcode src/

# Stricter thresholds
pyscn analyze --min-complexity 10 --min-severity critical src/

# Use a specific config file
pyscn analyze --config ./configs/strict.toml src/

# Don't open the browser (useful in sandboxes or containers)
pyscn analyze --no-open .
```

## `analyze` と `check` の使い分け

| ユースケース | コマンド |
| --- | --- |
| ローカル開発、全体像を把握したいとき | `pyscn analyze` |
| CI での合否判定 | [`pyscn check`](check.md) |
| カスタムツール向けの機械可読出力 | `pyscn analyze --json` |

## 関連項目

- [Configuration Reference](../configuration/reference.md) — すべての設定項目。
- [Health Score](../output/health-score.md) — 0〜100 のスコアの算出方法。
- [Output Schemas](../output/schemas.md) — JSON / YAML / CSV のフィールド定義。
