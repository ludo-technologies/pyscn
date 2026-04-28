# `pyscn check`

CI/CD パイプライン向けの品質ゲートです。リンター形式の検出結果を **stderr** に出力し、しきい値を超える問題がある場合はゼロ以外の終了コードで終了します。

```text
pyscn check [flags] [paths...]
```

パスを省略した場合はカレントディレクトリが対象になります。

## 動作内容

`check` は [`analyze`](analyze.md) の CI 向けコンパニオンです:

- **検出結果は stderr** にリンター形式（`file:line:col: message`）で出力されます。
- パスすれば **exit 0**、問題が見つかるか実行エラーが発生すれば **exit 1**。
- **厳格なデフォルト** — 複雑度 10 を超える関数があれば失敗、循環依存があれば失敗（`--select deps` 指定時）。
- **高速** — 選択した解析のみ実行し、レポート生成をスキップします。

## フラグ

### 解析の選択

| フラグ | 説明 |
| --- | --- |
| `-s, --select <list>` | 指定した解析のみ実行します。値: `complexity`、`deadcode`、`clones`、`deps`（エイリアス `circular`）、`mockdata`、`di`。 |
| `--skip-clones`       | クローン検出を実行しません。 |

デフォルト（`--select` なし）: `complexity`、`deadcode`、**および `clones`** を実行します。`deps`、`mockdata`、`di` は `--select` で明示的に指定する必要があります。`--select` に切り替えずにクローン検出をスキップするには `--skip-clones` を使ってください。

### しきい値のオーバーライド

| フラグ | デフォルト | 説明 |
| --- | --- | --- |
| `--max-complexity <N>`   | `10` | いずれかの関数がこのサイクロマティック複雑度を超えた場合に失敗します。 |
| `--max-cycles <N>`       | `0`  | 失敗するまでに許容する循環依存サイクルの最大数。 |
| `--allow-dead-code`      | off  | デッドコードを警告のみとし、チェックを失敗させません。 |
| `--allow-circular-deps`  | off  | 循環依存を警告のみとし、チェックを失敗させません。 |

### 出力

| フラグ | 説明 |
| --- | --- |
| `-q, --quiet`          | 問題が見つかった場合のみ出力します。 |
| `-c, --config <path>`  | 指定したファイルから設定を読み込みます。 |
| `-v, --verbose`        | 詳細な進捗を表示します。 |

## 終了コード

| コード | 意味 |
| --- | --- |
| `0` | すべてのチェックに合格しました。 |
| `1` | 1 つ以上のチェックが失敗したか、実行エラーが発生しました。 |

`check` コマンドは「問題が見つかった」と「ツールの実行エラー」を異なる終了コードで区別しません。CI では stderr の出力と pyscn のゼロ以外の終了コードを合否判定にのみ利用してください。

## 使用例

```bash
# Standard CI gate (runs complexity, deadcode, clones)
pyscn check .

# Faster gate: skip clone detection
pyscn check --skip-clones .

# Complexity only, with a higher threshold for legacy code
pyscn check --select complexity --max-complexity 15 src/

# Check for circular imports
pyscn check --select deps src/

# Allow existing dead code while you clean it up
pyscn check --allow-dead-code src/

# Detect DI anti-patterns (opt-in)
pyscn check --select di src/

# Quiet mode — ideal for CI logs
pyscn check --quiet .
```

## `analyze` との関係

`check` は `analyze` と同じ解析エンジンおよび設定ファイルを使用します。違いは以下のとおりです:

| 観点 | `analyze` | `check` |
| --- | --- | --- |
| 出力 | レポートファイル（HTML/JSON/YAML/CSV） | リンター形式の stderr |
| 問題検出時の終了コード | 常に `0`（エラー時を除く） | しきい値を超える問題があれば exit `1` |
| クローン検出 | デフォルトで有効 | デフォルトで有効（`--skip-clones` でスキップ） |
| 依存関係解析 | デフォルトで有効 | デフォルトで無効（`--select deps` でオプトイン） |
| 速度 | 遅い（全解析 + レポート生成） | 高速（選択した解析のみ、レポートなし） |
| ユースケース | 対話的なレビュー | CI 品質ゲート |

両方を使い分けましょう: `analyze` で問題を理解し、`check` でリグレッションを防ぎます。

## 関連項目

- [CI/CD Integration](../integrations/ci-cd.md) — GitHub Actions / pre-commit / GitLab の設定例。
- [`pyscn analyze`](analyze.md) — レポート付きのフル解析。
