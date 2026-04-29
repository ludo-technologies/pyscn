# 設定リファレンス

`.pyscn.toml`（または `pyproject.toml` の `[tool.pyscn.*]`）で設定可能なすべてのキーです。`pyscn init` を実行するとコメント付きのスターターファイルが生成されます。

---

## `[output]`

結果の出力方法を制御します。

| キー              | 型    | デフォルト       | 説明 |
| ---------------- | ------- | ------------- | --- |
| `format`         | string  | `"text"`      | `text`, `json`, `yaml`, `csv`, `html` のいずれか。`--json` などの CLI フラグで上書きされます。 |
| `directory`      | string  | `""`          | 出力ディレクトリ。空の場合は CWD 配下の `.pyscn/reports/`。 |
| `show_details`   | bool    | `false`       | サマリーに検出結果ごとの詳細を含めます。 |
| `sort_by`        | string  | `"complexity"`| `name`, `complexity`, `risk` のいずれか。 |
| `min_complexity` | int     | `1`           | この複雑度未満の関数を除外します。設定時に `[complexity].min_complexity` を上書きします。 |

---

## `[complexity]`

循環的複雑度の分析です。

| キー                | 型 | デフォルト | 説明 |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | 分析器を実行します。 |
| `low_threshold`    | int  | `9`     | 「低リスク」の上限（この値を含む）。 |
| `medium_threshold` | int  | `19`    | 「中リスク」の上限。 |
| `max_complexity`   | int  | `0`     | CI 失敗の閾値。`0` = 制限なし。 |
| `min_complexity`   | int  | `1`     | この値未満の関数をレポートしません。 |
| `report_unchanged` | bool | `true`  | 複雑度 = 1 の関数を含めます。 |

閾値のガイダンスは [high-cyclomatic-complexity](../rules/high-cyclomatic-complexity.md) を参照してください。

---

## `[dead_code]`

デッドコード検出です。

| キー                              | 型   | デフォルト      | 説明 |
| -------------------------------- | ------ | ------------ | --- |
| `enabled`                        | bool   | `true`       | 分析器を実行します。 |
| `min_severity`                   | string | `"warning"`  | `info`, `warning`, `critical` のいずれか。 |
| `show_context`                   | bool   | `false`      | 周囲のソース行を含めます。 |
| `context_lines`                  | int    | `3`          | コンテキスト行数（0〜20）。 |
| `sort_by`                        | string | `"severity"` | `severity`, `line`, `file`, `function` のいずれか。 |
| `detect_after_return`            | bool   | `true`       | `return` の後の文を検出します。 |
| `detect_after_break`             | bool   | `true`       | `break` の後の文を検出します。 |
| `detect_after_continue`          | bool   | `true`       | `continue` の後の文を検出します。 |
| `detect_after_raise`             | bool   | `true`       | `raise` の後の文を検出します。 |
| `detect_unreachable_branches`    | bool   | `true`       | 到達不能なブランチを検出します。 |
| `ignore_patterns`                | string[] | `[]`       | 無視する行の正規表現パターン。 |

---

## `[clones]`

クローン検出（最も設定項目が多い分析器）です。

### フラグメント選択

| キー              | 型 | デフォルト | 説明 |
| ---------------- | ---- | ------- | --- |
| `min_lines`      | int  | `10`    | フラグメントとみなす最小行数。 |
| `min_nodes`      | int  | `20`    | 最小 AST ノード数。 |
| `skip_docstrings`| bool | `true`  | ハッシュ時にドキュメント文字列をスキップします。 |

### タイプ閾値（0.0〜1.0）

| キー                    | デフォルト | クローンタイプ |
| ---------------------- | ------- | --- |
| `type1_threshold`      | `0.85`  | 同一（空白/コメントのみ異なる）。 |
| `type2_threshold`      | `0.75`  | 識別子/リテラルの名前変更。 |
| `type3_threshold`      | `0.70`  | 変更を伴う構造的類似。 |
| `type4_threshold`      | `0.65`  | 意味的等価。 |
| `similarity_threshold` | `0.65`  | すべてのクローンに対するグローバル最小値。 |

### アルゴリズム

| キー                 | 型   | デフォルト    | 説明 |
| ------------------- | ------ | ---------- | --- |
| `cost_model_type`   | string | `"python"` | `default`, `python`, `weighted` のいずれか。 |
| `ignore_literals`   | bool   | `false`    | 異なるリテラルを等価として扱います。 |
| `ignore_identifiers`| bool   | `false`    | 異なる変数名を等価として扱います。 |
| `max_edit_distance` | float  | `50.0`     | 木編集距離の上限。 |
| `enable_dfa`        | bool   | `true`     | Type-4 のデータフロー解析。 |
| `enabled_clone_types` | string[] | all     | `type1`, `type2`, `type3`, `type4` のサブセット。 |

### LSH 高速化

| キー                        | 型           | デフォルト  | 説明 |
| -------------------------- | -------------- | -------- | --- |
| `lsh_enabled`              | `true\|false\|"auto"` | `"auto"` | LSH を有効にします（`auto` = フラグメント数に基づく）。 |
| `lsh_auto_threshold`       | int            | `500`    | 自動有効化のフラグメント数閾値。 |
| `lsh_similarity_threshold` | float          | `0.50`   | LSH 候補のプレフィルター。 |
| `lsh_bands`                | int            | `32`     | LSH バンド数。 |
| `lsh_rows`                 | int            | `4`      | バンドあたりの行数。 |
| `lsh_hashes`               | int            | `128`    | ハッシュ関数の数。 |

### グルーピング

| キー                  | 型   | デフォルト       | 説明 |
| -------------------- | ------ | ------------- | --- |
| `grouping_mode`      | string | `"connected"` | `connected`, `star`, `complete_linkage`, `k_core` のいずれか。 |
| `grouping_threshold` | float  | `0.65`        | グルーピングの最小類似度。 |
| `k_core_k`           | int    | `2`           | `k_core` モードの k パラメータ。 |

### パフォーマンス

| キー               | 型 | デフォルト | 説明 |
| ----------------- | ---- | ------- | --- |
| `max_memory_mb`   | int  | `100`   | メモリ上限（MB）。`0` = 制限なし。 |
| `batch_size`      | int  | `100`   | バッチあたりのファイル数。 |
| `enable_batching` | bool | `true`  | バッチ処理を行います。 |
| `max_goroutines`  | int  | `4`     | 並行ワーカー数。 |
| `timeout_seconds` | int  | `300`   | 分析ごとのタイムアウト。 |

### 出力フィルタリング

| キー             | 型  | デフォルト         | 説明 |
| --------------- | ----- | --------------- | --- |
| `min_similarity`| float | `0.0`           | この値未満のペアを除外します。 |
| `max_similarity`| float | `1.0`           | この値を超えるペアを除外します。 |
| `max_results`   | int   | `10000`         | レポートする最大ペア数。`0` = 制限なし。 |
| `show_details`  | bool  | `false`         | 詳細出力。 |
| `show_content`  | bool  | `false`         | レポートにソースを含めます。 |
| `sort_by`       | string| `"similarity"`  | `similarity`, `size`, `location`, `type` のいずれか。 |
| `group_clones`  | bool  | `true`          | 関連するクローンをグループ化します。 |

---

## `[cbo]`

Coupling Between Objects（クラス結合度）です。

| キー                | 型 | デフォルト | 説明 |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | 分析器を実行します。 |
| `low_threshold`    | int  | `3`     | 「低リスク」の上限。 |
| `medium_threshold` | int  | `7`     | 「中リスク」の上限。 |
| `min_cbo`          | int  | `0`     | この CBO 未満のクラスを除外します。 |
| `max_cbo`          | int  | `0`     | この CBO を超えるクラスを除外します。`0` = 制限なし。 |
| `show_zeros`       | bool | `false` | CBO = 0 のクラスを含めます。 |
| `include_builtins` | bool | `false` | `list`/`dict`/`str` を依存関係としてカウントします。 |
| `include_imports`  | bool | `true`  | インポートされたモジュール参照をカウントします。 |

---

## `[lcom]`

Lack of Cohesion of Methods（LCOM4）です。

| キー                | 型 | デフォルト | 説明 |
| ------------------ | ---- | ------- | --- |
| `low_threshold`    | int  | `2`     | 「低リスク」の上限（良好な凝集度）。 |
| `medium_threshold` | int  | `5`     | 「中リスク」の上限。 |

---

## `[analysis]`

ファイル探索ルールです。

| キー                | 型     | デフォルト       | 説明 |
| ------------------ | -------- | ------------- | --- |
| `recursive`        | bool     | `true`        | サブディレクトリに再帰的に探索します。 |
| `follow_symlinks`  | bool     | `false`       | シンボリックリンクをたどります。 |
| `include_patterns` | string[] | `["**/*.py"]` | 含める glob パターン。 |
| `exclude_patterns` | string[] | 下記参照     | 除外する glob パターン。 |

デフォルトの `exclude_patterns`:

```toml
[
  "test_*.py", "*_test.py",
  "**/__pycache__/*", "**/*.pyc",
  "**/.pytest_cache/", ".tox/",
  "venv/", "env/", ".venv/", ".env/",
]
```

---

## `[architecture]`

レイヤーバリデーションです。すべてのキーはオプションです。レイヤーを定義しない場合、アーキテクチャ分析は許容モードで実行されます。

| キー                        | 型  | デフォルト | 説明 |
| -------------------------- | ----- | ------- | --- |
| `enabled`                  | bool  | `true`  | レイヤーバリデーションを実行します。 |
| `validate_layers`          | bool  | `true`  | レイヤー間ルールをチェックします。 |
| `validate_cohesion`        | bool  | `true`  | パッケージ凝集度をチェックします。 |
| `validate_responsibility`  | bool  | `true`  | モジュールごとの責務数をチェックします。 |
| `strict_mode`              | bool  | `true`  | 厳格なバリデーション。 |
| `fail_on_violations`       | bool  | `false` | 違反時に非ゼロ終了します。 |
| `min_cohesion`             | float | `0.5`   | 最小パッケージ凝集度。 |
| `max_coupling`             | int   | `10`    | レイヤー間結合度の最大値。 |
| `max_responsibilities`     | int   | `3`     | モジュールごとの責務の最大数。 |

### レイヤー定義

```toml
[[architecture.layers]]
name = "presentation"
packages = ["router", "routers", "handler", "handlers", "controller", "api"]

[[architecture.layers]]
name = "application"
packages = ["service", "services", "usecase", "usecases"]

[[architecture.layers]]
name = "domain"
packages = ["model", "models", "entity", "entities"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repository", "repositories", "db", "database"]
```

### レイヤールール

```toml
[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain"]
```

---

## `[dependencies]`

モジュール依存関係分析です。`pyscn check` では **オプトイン**。`pyscn analyze` ではスキップしない限り常に実行されます。

| キー                  | 型   | デフォルト | 説明 |
| -------------------- | ------ | ------- | --- |
| `enabled`            | bool   | `false` | 分析器を実行します（analyze は設定に関係なく常に実行します）。 |
| `include_stdlib`     | bool   | `false` | 標準ライブラリのインポートを含めます。 |
| `include_third_party`| bool   | `true`  | サードパーティのインポートを含めます。 |
| `follow_relative`    | bool   | `true`  | 相対インポートをたどります。 |
| `detect_cycles`      | bool   | `true`  | 循環インポートを検出します。 |
| `calculate_metrics`  | bool   | `true`  | Ca/Ce/I/A/D を計算します。 |
| `find_long_chains`   | bool   | `true`  | 最長の依存チェーンをレポートします。 |
| `cycle_reporting`    | string | `"summary"` | `all`, `critical`, `summary` のいずれか。 |
| `max_cycles_to_show` | int    | `10`    | レポートする循環の上限。 |
| `sort_by`            | string | `"name"` | `name`, `coupling`, `instability`, `distance`, `risk` のいずれか。 |
| `show_matrix`        | bool   | `false` | 依存関係マトリクスを含めます。 |
| `generate_dot_graph` | bool   | `false` | Graphviz DOT 出力を生成します。 |

---

## `[mock_data]`

モック/プレースホルダーデータ検出です。**オプトイン**。

| キー              | 型     | デフォルト     | 説明 |
| ---------------- | -------- | ----------- | --- |
| `enabled`        | bool     | `false`     | 分析器を実行します。 |
| `min_severity`   | string   | `"warning"` | `info`, `warning`, `error` のいずれか。 |
| `ignore_tests`   | bool     | `true`      | テストファイルをスキップします。 |
| `keywords`       | string[] | 組み込み    | モック指標として検出されるキーワード。 |
| `domains`        | string[] | 組み込み    | 検出されるドメイン（`example.com`, `test.com` など）。 |
| `ignore_patterns`| string[] | `[]`        | スキップするファイル/正規表現パターン。 |

---

## `[di]`

Dependency Injection のアンチパターン検出です。**オプトイン**。

| キー                            | 型   | デフォルト     | 説明 |
| ------------------------------ | ------ | ----------- | --- |
| `enabled`                      | bool   | `false`     | 分析器を実行します。 |
| `min_severity`                 | string | `"warning"` | `info`, `warning`, `error` のいずれか。 |
| `constructor_param_threshold`  | int    | `5`         | この数を超えるパラメータを持つ `__init__` を検出します。 |

---

## CLI フラグと設定キーの対応表

設定キーに直接マッピングされないフラグ（`--select`, `--skip-*`, `--no-open`）は、読み込まれた設定の上に適用されます。

| CLI フラグ                | 設定キー                        |
| ----------------------- | --------------------------------- |
| `--config <path>`       | —（探索を上書き）           |
| `--json/--yaml/--csv/--html` | `[output] format`            |
| `--min-complexity`      | `[complexity] min_complexity`     |
| `--max-complexity`      | `[complexity] max_complexity`     |
| `--min-severity`        | `[dead_code] min_severity`        |
| `--clone-threshold`     | `[clones] similarity_threshold`   |
| `--min-cbo`             | `[cbo] min_cbo`                   |
| `--max-cycles`          | —（check コマンド専用）            |

## 関連項目

- [設定ファイル形式](format.md) — 探索と優先順位。
- [設定例](examples.md) — すぐに使える設定。
