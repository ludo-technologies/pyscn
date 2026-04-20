# 出力スキーマ

この仕様は、pyscn が生成する JSON、YAML、CSV 出力の正確な構造を定義します。ここに記載されたすべてのフィールド名、型、およびセマンティクスは、同一メジャーバージョン内のパッチリリースにおいて安定しています。

## 安定性契約

| 保証               | 範囲                                                                              |
| ------------------ | --------------------------------------------------------------------------------- |
| 安定               | フィールド名、フィールド型、フィールドセマンティクス、列挙値                      |
| 変更の可能性あり   | オブジェクト内のフィールド順序、配列要素の順序、新規フィールドの追加              |
| 破壊的変更         | フィールドの削除またはリネーム、フィールド型の変更、列挙値の削除                  |

破壊的変更はメジャーバージョンアップ時のみ行われます。利用者は未知のフィールドを無視しなければなりません（MUST）。

<!-- Field naming note: in `pyscn analyze` JSON/YAML, nested analyzer objects (`complexity`, `cbo`, `lcom`, `system`) use Go-style PascalCase field names because their response structs do not carry JSON tags. Top-level keys, `dead_code`, `clone`, `suggestions`, and `summary` use snake_case. -->

## トップレベル構造 (`pyscn analyze`)

JSON および YAML 出力は、`domain/analyze.go` で定義された `AnalyzeResponse` Go 構造体をシリアライズしたものです。トップレベルのキーは以下の通りです：

```json
{
  "complexity":    { /* ComplexityResponse, present when enabled */ },
  "dead_code":     { /* DeadCodeResponse, present when enabled */ },
  "clone":         { /* CloneResponse, present when enabled */ },
  "cbo":           { /* CBOResponse, present when enabled */ },
  "lcom":          { /* LCOMResponse, present when enabled */ },
  "system":        { /* SystemAnalysisResponse, present when deps/arch enabled */ },
  "mock_data":     { /* MockDataResponse, present when enabled */ },
  "suggestions":   [ /* Suggestion array, omitted when empty */ ],
  "summary":       { /* AnalyzeSummary, always present */ },
  "generated_at":  "2026-04-14T10:18:23Z",
  "duration_ms":   2347,
  "version":       "0.14.0"
}
```

| フィールド     | 型                | 説明                                                   | 安定性    |
| ------------- | ----------------- | ------------------------------------------------------ | --------- |
| `complexity`  | object \| absent  | 複雑度分析が実行された場合に存在します。               | stable    |
| `dead_code`   | object \| absent  | デッドコード分析が実行された場合に存在します。         | stable    |
| `clone`       | object \| absent  | クローン検出が実行された場合に存在します。             | stable    |
| `cbo`         | object \| absent  | CBO 分析が実行された場合に存在します。                 | stable    |
| `lcom`        | object \| absent  | LCOM 分析が実行された場合に存在します。                | stable    |
| `system`      | object \| absent  | 依存関係またはアーキテクチャ分析が実行された場合に存在します。 | stable    |
| `mock_data`   | object \| absent  | モックデータ検出が実行された場合に存在します。         | stable    |
| `suggestions` | array \| absent   | 導出された提案。空の場合は省略されます。               | stable    |
| `summary`     | object            | 常に存在します。[`summary`](#summary-オブジェクト) を参照してください。 | stable    |
| `generated_at`| string (RFC 3339) | 分析完了時刻。                                         | stable    |
| `duration_ms` | integer           | 分析の合計所要時間（ミリ秒）。                         | stable    |
| `version`     | string            | pyscn のセマンティックバージョン。                     | stable    |

## `summary` オブジェクト

`domain.AnalyzeSummary` に対応します。対応するアナライザが無効の場合、すべての数値カウンタはデフォルト値 `0` になります。すべてのフィールドは常に存在します。

### ファイル統計

| フィールド        | 型      | 説明                                             |
| ---------------- | ------- | ------------------------------------------------ |
| `total_files`    | integer | 検出された Python ファイルの数。                 |
| `analyzed_files` | integer | 正常に分析されたファイルの数。                   |
| `skipped_files`  | integer | パースエラーまたはフィルタによりスキップされたファイル。 |

### アナライザステータスフラグ

| フィールド            | 型      | 説明                                                       |
| -------------------- | ------- | ---------------------------------------------------------- |
| `complexity_enabled` | boolean | 複雑度分析が結果を生成した場合に `true`。                  |
| `dead_code_enabled`  | boolean | デッドコード分析が結果を生成した場合に `true`。            |
| `clone_enabled`      | boolean | クローン検出が結果を生成した場合に `true`。                |
| `cbo_enabled`        | boolean | CBO 分析が結果を生成した場合に `true`。                    |
| `lcom_enabled`       | boolean | LCOM 分析が結果を生成した場合に `true`。                   |
| `deps_enabled`       | boolean | 依存関係分析が結果を生成した場合に `true`。                |
| `arch_enabled`       | boolean | アーキテクチャ検証が結果を生成した場合に `true`。          |
| `mock_data_enabled`  | boolean | モックデータ検出が結果を生成した場合に `true`。            |

### 複雑度メトリクス

| フィールド               | 型      | 説明                                             |
| ----------------------- | ------- | ------------------------------------------------ |
| `total_functions`       | integer | 分析された関数の合計数。                         |
| `average_complexity`    | number  | サイクロマティック複雑度の平均値。関数がない場合は `0`。 |
| `high_complexity_count` | integer | 複雑度が 10 を超える関数の数（中閾値）。         |

### デッドコードメトリクス

| フィールド            | 型      | 説明                                         |
| -------------------- | ------- | -------------------------------------------- |
| `dead_code_count`    | integer | 検出結果の合計数。                           |
| `critical_dead_code` | integer | 重大度 `critical` の検出結果。               |
| `warning_dead_code`  | integer | 重大度 `warning` の検出結果。                |
| `info_dead_code`     | integer | 重大度 `info` の検出結果。                   |

### クローンメトリクス

| フィールド                     | 型      | 説明                                                      |
| ----------------------------- | ------- | --------------------------------------------------------- |
| `total_clones`                | integer | クローンとして識別された個別のコード断片。                |
| `clone_pairs`                 | integer | クローンペアの数。                                        |
| `clone_groups`                | integer | クローングループの数。                                    |
| `code_duplication_percentage` | number  | 推定重複率、`0`–`100`。                                   |

### CBO メトリクス

| フィールド                 | 型      | 説明                                                      |
| ------------------------- | ------- | --------------------------------------------------------- |
| `cbo_classes`             | integer | 分析されたクラスの合計数。                                |
| `high_coupling_classes`   | integer | CBO > 7 のクラス。                                        |
| `medium_coupling_classes` | integer | 3 < CBO ≤ 7 のクラス。                                    |
| `average_coupling`        | number  | CBO の平均値。                                            |

### LCOM メトリクス

| フィールド             | 型      | 説明                                         |
| --------------------- | ------- | -------------------------------------------- |
| `lcom_classes`        | integer | 分析されたクラスの合計数。                   |
| `high_lcom_classes`   | integer | LCOM4 > 5 のクラス。                         |
| `medium_lcom_classes` | integer | 2 < LCOM4 ≤ 5 のクラス。                     |
| `average_lcom`        | number  | LCOM4 の平均値。                             |

### 依存関係メトリクス

| フィールド                      | 型      | 説明                                                           |
| ------------------------------ | ------- | -------------------------------------------------------------- |
| `deps_total_modules`           | integer | 分析されたモジュールの合計数。                                 |
| `deps_modules_in_cycles`       | integer | 少なくとも1つの循環依存に関与しているモジュール。              |
| `deps_max_depth`               | integer | 最長の依存チェーンの長さ。                                     |
| `deps_main_sequence_deviation` | number  | Martin のメインシーケンスからの平均距離、`0`–`1`。             |

### アーキテクチャメトリクス

| フィールド         | 型     | 説明                                                              |
| ----------------- | ------ | ----------------------------------------------------------------- |
| `arch_compliance` | number | アーキテクチャ準拠率、`0`–`1`。`1.0` = 完全準拠。                 |

### モックデータメトリクス

| フィールド               | 型      | 説明                                                |
| ----------------------- | ------- | --------------------------------------------------- |
| `mock_data_count`       | integer | モックデータ検出結果の合計数。                      |
| `mock_data_error_count` | integer | 重大度 error の検出結果。                           |
| `mock_data_warning_count` | integer | 重大度 warning の検出結果。                       |
| `mock_data_info_count`  | integer | 重大度 info の検出結果。                            |

### ヘルススコアリング

| フィールド            | 型      | 説明                                                               |
| -------------------- | ------- | ------------------------------------------------------------------ |
| `health_score`       | integer | 総合スコア、`0`–`100`。[ヘルススコア](health-score.md) を参照。    |
| `grade`              | string  | レターグレード。`A`、`B`、`C`、`D`、`F`、`N/A` のいずれか。        |
| `complexity_score`   | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `dead_code_score`    | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `duplication_score`  | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `coupling_score`     | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `cohesion_score`     | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `dependency_score`   | integer | カテゴリ別スコア、`0`–`100`。                                      |
| `architecture_score` | integer | カテゴリ別スコア、`0`–`100`。                                      |

## `complexity` オブジェクト

`domain.ComplexityResponse` に対応します。ネストされたフィールド名は Go の PascalCase です。

```json
{
  "Functions": [ /* FunctionComplexity array */ ],
  "Summary": { /* ComplexitySummary */ },
  "raw_metrics": [ /* RawMetrics array, present when computed */ ],
  "raw_metrics_summary": { /* RawMetricsSummary, present when computed */ },
  "Warnings": [ "..." ],
  "Errors": [ "..." ],
  "GeneratedAt": "2026-04-14T10:18:23Z",
  "Version": "0.14.0",
  "Config": null
}
```

### `Functions[]` 要素 (`FunctionComplexity`)

| フィールド     | 型      | 説明                                                         |
| ------------- | ------- | ------------------------------------------------------------ |
| `Name`        | string  | 関数名。モジュールレベルのコードの場合は `__main__`。        |
| `FilePath`    | string  | ソースファイルのパス。                                       |
| `StartLine`   | integer | 1始まりの開始行。                                            |
| `StartColumn` | integer | 0始まりの開始列。                                            |
| `EndLine`     | integer | 1始まりの終了行。                                            |
| `Metrics`     | object  | [`ComplexityMetrics`](#complexitymetrics-オブジェクト) を参照。 |
| `RiskLevel`   | string  | `low`、`medium`、`high` のいずれか。                         |

### `ComplexityMetrics` オブジェクト

| フィールド             | 型      | 説明                                               |
| --------------------- | ------- | -------------------------------------------------- |
| `Complexity`          | integer | McCabe サイクロマティック複雑度。                  |
| `CognitiveComplexity` | integer | 認知的複雑度（SonarQube スタイル）。               |
| `Nodes`               | integer | CFG ノード数。                                     |
| `Edges`               | integer | CFG エッジ数。                                     |
| `NestingDepth`        | integer | 最大ネスト深度。                                   |
| `IfStatements`        | integer | `if` 文の数。                                      |
| `LoopStatements`      | integer | `for`/`while` ループの数。                         |
| `ExceptionHandlers`   | integer | `except` 節の数。                                  |
| `SwitchCases`         | integer | `match` ケースの数（Python 3.10+）。               |

### `Summary` オブジェクト (`ComplexitySummary`)

| フィールド                | 型      | 説明                                                                   |
| ------------------------ | ------- | ---------------------------------------------------------------------- |
| `TotalFunctions`         | integer | 分析された関数の合計数。                                               |
| `AverageComplexity`      | number  | すべての関数の `Complexity` の算術平均。                                |
| `MaxComplexity`          | integer | 観測された最大複雑度。                                                 |
| `MinComplexity`          | integer | 観測された最小複雑度。                                                 |
| `FilesAnalyzed`          | integer | 少なくとも1つの関数を含むファイル。                                    |
| `LowRiskFunctions`       | integer | `RiskLevel = low` の関数。                                             |
| `MediumRiskFunctions`    | integer | `RiskLevel = medium` の関数。                                          |
| `HighRiskFunctions`      | integer | `RiskLevel = high` の関数。                                            |
| `ComplexityDistribution` | object  | 複雑度バケット（string）からカウント（integer）へのヒストグラム、または `null`。 |

### `raw_metrics[]` 要素 (`RawMetrics`)

| フィールド         | 型      | 説明                                                |
| ----------------- | ------- | --------------------------------------------------- |
| `file_path`       | string  | ソースファイルのパス。                              |
| `sloc`            | integer | ソースコード行数（空行・コメント行を除く）。        |
| `lloc`            | integer | 論理コード行数。                                    |
| `comment_lines`   | integer | コメントを含む行数。                                |
| `docstring_lines` | integer | docstring 内の行数。                                |
| `blank_lines`     | integer | 空白行または空行。                                  |
| `total_lines`     | integer | 物理行の合計数。                                    |
| `comment_ratio`   | number  | `(comment_lines + docstring_lines) / total_lines`、`0`–`1`。 |


## `dead_code` オブジェクト

`domain.DeadCodeResponse` に対応します。全体を通して snake_case のフィールド名を使用します。

```json
{
  "files": [ /* FileDeadCode array */ ],
  "summary": { /* DeadCodeSummary */ },
  "warnings": null,
  "errors": null,
  "generated_at": "",
  "version": "",
  "config": null
}
```

### `files[]` 要素 (`FileDeadCode`)

| フィールド           | 型      | 説明                                           |
| ------------------- | ------- | ---------------------------------------------- |
| `file_path`         | string  | ソースファイルのパス。                         |
| `functions`         | array   | 関数ごとの結果（下記参照）。                   |
| `total_findings`    | integer | このファイル内の関数全体の検出結果の合計。     |
| `total_functions`   | integer | このファイルで分析された関数の数。             |
| `affected_functions`| integer | 少なくとも1つの検出結果がある関数。            |
| `dead_code_ratio`   | number  | デッドブロック / 全ブロック、`0`–`1`。         |

### `files[].functions[]` 要素 (`FunctionDeadCode`)

| フィールド         | 型      | 説明                                         |
| ----------------- | ------- | -------------------------------------------- |
| `name`            | string  | 関数名。                                     |
| `file_path`       | string  | ソースファイルのパス。                       |
| `findings`        | array   | この関数内の検出結果（下記参照）。           |
| `total_blocks`    | integer | 関数内の CFG ブロックの合計数。              |
| `dead_blocks`     | integer | 到達不能な CFG ブロック。                    |
| `reachable_ratio` | number  | `(total_blocks - dead_blocks) / total_blocks`、`0`–`1`。 |
| `critical_count`  | integer | 重大度 `critical` の検出結果。               |
| `warning_count`   | integer | 重大度 `warning` の検出結果。                |
| `info_count`      | integer | 重大度 `info` の検出結果。                   |

### `files[].functions[].findings[]` 要素 (`DeadCodeFinding`)

| フィールド       | 型      | 説明                                                          |
| --------------- | ------- | ------------------------------------------------------------- |
| `location`      | object  | [`DeadCodeLocation`](#deadcodelocation-オブジェクト) を参照。 |
| `function_name` | string  | 包含する関数名。                                              |
| `code`          | string  | デッドコードのソースコード断片。                              |
| `reason`        | string  | 分類 — 下記の列挙を参照。                                    |
| `severity`      | string  | `critical`、`warning`、`info` のいずれか。                    |
| `description`   | string  | 人間が読める説明。                                            |
| `context`       | array of string \| absent | 周囲のソース行。`--show-context` 指定時に存在。 |
| `block_id`      | string \| absent | CFG ブロック識別子。                                  |

`reason` 列挙:

| 値                    | 意味                                         |
| --------------------- | -------------------------------------------- |
| `after_return`        | `return` 文の後のコード。                    |
| `after_break`         | `break` 文の後のコード。                     |
| `after_continue`      | `continue` 文の後のコード。                  |
| `after_raise`         | `raise` 文の後のコード。                     |
| `unreachable_branch`  | 到達されない条件分岐。                       |

### `DeadCodeLocation` オブジェクト

| フィールド      | 型      | 説明                       |
| -------------- | ------- | -------------------------- |
| `file_path`    | string  | ソースファイルのパス。     |
| `start_line`   | integer | 1始まりの開始行。          |
| `end_line`     | integer | 1始まりの終了行。          |
| `start_column` | integer | 0始まりの開始列。          |
| `end_column`   | integer | 0始まりの終了列。          |

### `summary` オブジェクト (`DeadCodeSummary`)

| フィールド                  | 型      | 説明                                             |
| -------------------------- | ------- | ------------------------------------------------ |
| `total_files`              | integer | 分析されたファイル数。                           |
| `total_functions`          | integer | 分析された関数数。                               |
| `total_findings`           | integer | 全ファイルの検出結果の合計。                     |
| `files_with_dead_code`     | integer | 少なくとも1つの検出結果があるファイル。           |
| `functions_with_dead_code` | integer | 少なくとも1つの検出結果がある関数。               |
| `critical_findings`        | integer | 重大度 `critical` の検出結果。                   |
| `warning_findings`         | integer | 重大度 `warning` の検出結果。                    |
| `info_findings`            | integer | 重大度 `info` の検出結果。                       |
| `findings_by_reason`       | object \| null | `reason` 値をキーとしたヒストグラム。      |
| `total_blocks`             | integer | 全関数の CFG ブロック数。                        |
| `dead_blocks`              | integer | 全関数の到達不能な CFG ブロック数。              |
| `overall_dead_ratio`       | number  | `dead_blocks / total_blocks`、`0`–`1`。          |

## `clone` オブジェクト

`domain.CloneResponse` に対応します。全体を通して snake_case のフィールド名を使用します。

```json
{
  "clones": [ /* Clone array, or null */ ],
  "clone_pairs": [ /* ClonePair array, or null */ ],
  "clone_groups": [ /* CloneGroup array, or null */ ],
  "statistics": { /* CloneStatistics */ },
  "duration_ms": 123,
  "success": true,
  "error": ""
}
```

### `clones[]` 要素 (`Clone`)

| フィールド    | 型      | 説明                                                         |
| ------------ | ------- | ------------------------------------------------------------ |
| `id`         | integer | レスポンス内で一意のクローン識別子。                         |
| `type`       | integer | 整数値のクローンタイプ: `1`、`2`、`3`、または `4`。          |
| `location`   | object  | [`CloneLocation`](#clonelocation-オブジェクト) を参照。      |
| `content`    | string  | 生のソーステキスト。`--show-content` 設定時のみ存在。        |
| `hash`       | string  | フィンガープリントハッシュ（アルゴリズムはクローンタイプに依存）。 |
| `size`       | integer | AST ノード数。                                               |
| `line_count` | integer | 断片の行数。                                                 |
| `complexity` | integer | 断片のサイクロマティック複雑度。                             |

`type` 列挙（整数値）:

| 値    | 意味                                                                 |
| ----- | -------------------------------------------------------------------- |
| `1`   | Type-1: 空白/コメントを除いて同一。                                  |
| `2`   | Type-2: 構文的に同一、識別子/リテラルが異なる。                      |
| `3`   | Type-3: 変更を伴う構造的類似。                                       |
| `4`   | Type-4: 意味的に等価、構文的に異なる。                               |

### `CloneLocation` オブジェクト

| フィールド    | 型      | 説明                     |
| ------------ | ------- | ------------------------ |
| `file_path`  | string  | ソースファイルのパス。   |
| `start_line` | integer | 1始まりの開始行。        |
| `end_line`   | integer | 1始まりの終了行。        |
| `start_col`  | integer | 0始まりの開始列。        |
| `end_col`    | integer | 0始まりの終了列。        |

### `clone_pairs[]` 要素 (`ClonePair`)

| フィールド    | 型      | 説明                                                   |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | ペア識別子。                                           |
| `clone1`     | object  | 1つ目のクローン（`Clone` オブジェクト）。              |
| `clone2`     | object  | 2つ目のクローン（`Clone` オブジェクト）。              |
| `similarity` | number  | 類似度スコア、`0`–`1`。                                |
| `distance`   | number  | ツリー編集距離（Type-3）、それ以外は `0`。             |
| `type`       | integer | クローンタイプ（`clones[].type` と同じ列挙）。         |
| `confidence` | number  | 検出器の信頼度、`0`–`1`。                              |

### `clone_groups[]` 要素 (`CloneGroup`)

| フィールド    | 型      | 説明                                                   |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | グループ識別子。                                       |
| `clones`     | array   | メンバーの `Clone` オブジェクト。                      |
| `type`       | integer | 支配的なクローンタイプ。                               |
| `similarity` | number  | 代表的な類似度、`0`–`1`。                              |
| `size`       | integer | メンバー数（`len(clones)`）。                          |

### `statistics` オブジェクト (`CloneStatistics`)

| フィールド            | 型      | 説明                                                     |
| -------------------- | ------- | -------------------------------------------------------- |
| `total_fragments`    | integer | 抽出されたすべての断片（関数、クラスなど）。             |
| `total_clones`       | integer | クローンとして分類された断片。                           |
| `total_clone_pairs`  | integer | 検出されたペアの数。                                     |
| `total_clone_groups` | integer | グループの数。                                           |
| `clones_by_type`     | object \| null | タイプラベル（`Type-1`…`Type-4`）からカウントへのマップ。 |
| `average_similarity` | number  | ペア全体の平均類似度、`0`–`1`。                          |
| `lines_analyzed`     | integer | 考慮されたソース行の合計数。                             |
| `nodes_analyzed`     | integer | 考慮された AST ノードの合計数。                          |
| `files_analyzed`     | integer | 断片を提供した個別ファイル数。                           |

その他の `CloneResponse` フィールド:

| フィールド     | 型      | 説明                                               |
| ------------- | ------- | -------------------------------------------------- |
| `duration_ms` | integer | クローン検出の所要時間（ミリ秒）。                 |
| `success`     | boolean | 正常完了時に `true`。                              |
| `error`       | string \| absent | `success=false` の場合のエラーメッセージ。 |

## `cbo` オブジェクト

`domain.CBOResponse` に対応します。ネストされたフィールド名は Go の PascalCase です。

```json
{
  "Classes": [ /* ClassCoupling array */ ],
  "Summary": { /* CBOSummary */ },
  "Warnings": null,
  "Errors": null,
  "GeneratedAt": "",
  "Version": "",
  "Config": null
}
```

### `Classes[]` 要素 (`ClassCoupling`)

| フィールド     | 型      | 説明                                        |
| ------------- | ------- | ------------------------------------------- |
| `Name`        | string  | クラス名。                                  |
| `FilePath`    | string  | ソースファイルのパス。                      |
| `StartLine`   | integer | 1始まりの開始行。                           |
| `EndLine`     | integer | 1始まりの終了行。                           |
| `Metrics`     | object  | [`CBOMetrics`](#cbometrics-オブジェクト) を参照。 |
| `RiskLevel`   | string  | `low`、`medium`、`high` のいずれか。        |
| `IsAbstract`  | boolean | クラスが抽象クラスの場合に `true`。         |
| `BaseClasses` | array of string \| null | 直接の基底クラス。          |

### `CBOMetrics` オブジェクト

| フィールド                     | 型      | 説明                                                      |
| ----------------------------- | ------- | --------------------------------------------------------- |
| `CouplingCount`               | integer | CBO 値: このクラスが依存する個別クラスの数。              |
| `InheritanceDependencies`     | integer | 基底クラスからの依存。                                    |
| `TypeHintDependencies`        | integer | 型アノテーションからの依存。                              |
| `InstantiationDependencies`   | integer | オブジェクト生成からの依存。                              |
| `AttributeAccessDependencies` | integer | メソッド呼び出しおよび属性アクセスからの依存。            |
| `ImportDependencies`          | integer | 明示的インポートからの依存。                              |
| `DependentClasses`            | array of string \| null | 結合しているクラスの名前。            |

### `Summary` オブジェクト (`CBOSummary`)

| フィールド                  | 型      | 説明                                              |
| -------------------------- | ------- | ------------------------------------------------- |
| `TotalClasses`             | integer | 分析されたクラスの合計数。                        |
| `AverageCBO`               | number  | CBO の平均値。                                    |
| `MaxCBO`                   | integer | 観測された最大 CBO。                              |
| `MinCBO`                   | integer | 観測された最小 CBO。                              |
| `ClassesAnalyzed`          | integer | 有効なメトリクスを持つクラス。                    |
| `FilesAnalyzed`            | integer | 少なくとも1つのクラスを含むファイル。              |
| `LowRiskClasses`           | integer | CBO ≤ 低閾値（デフォルト `3`）のクラス。          |
| `MediumRiskClasses`        | integer | 低 < CBO ≤ 中閾値のクラス。                       |
| `HighRiskClasses`          | integer | CBO > 中閾値（デフォルト `7`）のクラス。          |
| `CBODistribution`          | object \| null | バケットラベルからカウントへのヒストグラム。 |
| `MostCoupledClasses`       | array \| null | CBO 上位10クラス（`ClassCoupling`）。       |
| `MostDependedUponClasses`  | array of string \| null | 被依存度が最も高いクラス。    |

## `lcom` オブジェクト

`domain.LCOMResponse` に対応します。ネストされたフィールド名は Go の PascalCase です。

```json
{
  "Classes": [ /* ClassCohesion array */ ],
  "Summary": { /* LCOMSummary */ },
  "Warnings": null,
  "Errors": null,
  "GeneratedAt": "",
  "Version": "",
  "Config": null
}
```

### `Classes[]` 要素 (`ClassCohesion`)

| フィールド   | 型      | 説明                                             |
| ----------- | ------- | ------------------------------------------------ |
| `Name`      | string  | クラス名。                                       |
| `FilePath`  | string  | ソースファイルのパス。                           |
| `StartLine` | integer | 1始まりの開始行。                                |
| `EndLine`   | integer | 1始まりの終了行。                                |
| `Metrics`   | object  | [`LCOMMetrics`](#lcommetrics-オブジェクト) を参照。 |
| `RiskLevel` | string  | `low`、`medium`、`high` のいずれか。             |

### `LCOMMetrics` オブジェクト

| フィールド           | 型      | 説明                                                        |
| ------------------- | ------- | ----------------------------------------------------------- |
| `LCOM4`             | integer | メソッド-変数グラフの連結成分数。                           |
| `TotalMethods`      | integer | クラス内の全メソッド数。                                    |
| `ExcludedMethods`   | integer | LCOM4 から除外されたメソッド（`@classmethod`、`@staticmethod`）。 |
| `InstanceVariables` | integer | アクセスされた個別の `self.x` 変数の数。                    |
| `MethodGroups`      | array of array of string \| null | 連結成分ごとにグループ化されたメソッド名。 |

### `Summary` オブジェクト (`LCOMSummary`)

| フィールド              | 型      | 説明                                            |
| ---------------------- | ------- | ----------------------------------------------- |
| `TotalClasses`         | integer | 分析されたクラス数。                            |
| `AverageLCOM`          | number  | LCOM4 の平均値。                                |
| `MaxLCOM`              | integer | 観測された最大 LCOM4。                          |
| `MinLCOM`              | integer | 観測された最小 LCOM4。                          |
| `ClassesAnalyzed`      | integer | 有効なメトリクスを持つクラス。                  |
| `FilesAnalyzed`        | integer | 少なくとも1つのクラスを含むファイル。            |
| `LowRiskClasses`       | integer | LCOM4 ≤ 低閾値（デフォルト `2`）のクラス。      |
| `MediumRiskClasses`    | integer | 低 < LCOM4 ≤ 中閾値のクラス。                   |
| `HighRiskClasses`      | integer | LCOM4 > 中閾値（デフォルト `5`）のクラス。      |
| `LCOMDistribution`     | object \| null | バケットラベルからカウントへのヒストグラム。 |
| `LeastCohesiveClasses` | array \| null | LCOM4 上位10クラス（`ClassCohesion`）。    |

## `system` オブジェクト

`domain.SystemAnalysisResponse` に対応します。ネストされたフィールド名は Go の PascalCase です。

```json
{
  "DependencyAnalysis":   { /* DependencyAnalysisResult, or null */ },
  "ArchitectureAnalysis": { /* ArchitectureAnalysisResult, or null */ },
  "Summary":              { /* SystemAnalysisSummary */ },
  "Issues":               [ /* SystemIssue array */ ],
  "Recommendations":      [ /* SystemRecommendation array */ ],
  "Warnings":             [ ],
  "Errors":               [ ],
  "GeneratedAt":          "0001-01-01T00:00:00Z",
  "Duration":             0,
  "Version":              "",
  "Config":               null
}
```

### `Summary` オブジェクト (`SystemAnalysisSummary`)

| フィールド                  | 型      | 説明                                            |
| -------------------------- | ------- | ----------------------------------------------- |
| `TotalModules`             | integer | 分析されたモジュールの合計数。                  |
| `TotalPackages`            | integer | パッケージの合計数。                            |
| `TotalDependencies`        | integer | 依存関係エッジの合計数。                        |
| `ProjectRoot`              | string  | プロジェクトのルートディレクトリ。              |
| `OverallQualityScore`      | number  | 総合品質スコア、`0`–`100`。                     |
| `MaintainabilityScore`     | number  | 保守性インデックスの平均値。                    |
| `ArchitectureScore`        | number  | アーキテクチャ準拠スコア。                      |
| `ModularityScore`          | number  | システムのモジュール性スコア。                  |
| `TechnicalDebtHours`       | number  | 推定技術的負債の合計（時間単位）。              |
| `AverageCoupling`          | number  | モジュール結合度の平均値。                      |
| `AverageInstability`       | number  | 不安定性 (I) の平均値。                         |
| `CyclicDependencies`       | integer | 循環依存に関与するモジュール。                  |
| `ArchitectureViolations`   | integer | アーキテクチャルール違反の数。                  |
| `HighRiskModules`          | integer | 高リスクとフラグされたモジュール。              |
| `CriticalIssues`           | integer | 重大な問題の数。                                |
| `RefactoringCandidates`    | integer | リファクタリングが必要なモジュール。            |
| `ArchitectureImprovements` | integer | 提案されたアーキテクチャ改善。                  |

### `DependencyAnalysis` オブジェクト

| フィールド              | 型      | 説明                                                                 |
| ---------------------- | ------- | -------------------------------------------------------------------- |
| `TotalModules`         | integer | 依存関係グラフ内のモジュールの合計数。                               |
| `TotalDependencies`    | integer | エッジの合計数。                                                     |
| `RootModules`          | array of string | 外向き依存のないモジュール。                                 |
| `LeafModules`          | array of string | 内向き依存のないモジュール。                                 |
| `ModuleMetrics`        | object  | モジュール名から `ModuleDependencyMetrics` へのマップ。              |
| `DependencyMatrix`     | object  | モジュールからモジュールから boolean へのマップ。                    |
| `CircularDependencies` | object  | 循環検出結果。`Cycles`（array）と `TotalCycles`（integer）を含む。   |
| `CouplingAnalysis`     | object  | モジュールごとの結合度メトリクス: `Ca`、`Ce`、`Instability`、`Abstractness`、`Distance`。 |
| `LongestChains`        | array   | `DependencyPath` オブジェクトの配列。                                |
| `MaxDepth`             | integer | 最大依存深度。                                                       |

### `ModuleDependencyMetrics` オブジェクト

| フィールド                | 型      | 説明                                                     |
| ------------------------ | ------- | -------------------------------------------------------- |
| `ModuleName`             | string  | 完全修飾モジュール名。                                   |
| `Package`                | string  | 親パッケージ。                                           |
| `FilePath`               | string  | ソースファイルのパス。                                   |
| `IsPackage`              | boolean | パッケージ（`__init__.py` あり）の場合に `true`。        |
| `LinesOfCode`            | integer | コードの合計行数。                                       |
| `FunctionCount`          | integer | 関数の数。                                               |
| `ClassCount`             | integer | クラスの数。                                             |
| `PublicInterface`        | array of string | `__all__` またはトップレベルのパブリック名。      |
| `AfferentCoupling`       | integer | Ca — このモジュールに依存するモジュール。                |
| `EfferentCoupling`       | integer | Ce — このモジュールが依存するモジュール。                |
| `Instability`            | number  | `I = Ce / (Ca + Ce)`、`0`–`1`。                          |
| `Abstractness`           | number  | A — 抽象要素 / 全要素、`0`–`1`。                         |
| `Distance`               | number  | `D = |A + I - 1|`、`0`–`1`。メインシーケンスからの距離。 |
| `Maintainability`        | number  | 保守性インデックス、`0`–`100`。                          |
| `TechnicalDebt`          | number  | 推定技術的負債（時間単位）。                             |
| `RiskLevel`              | string  | `low`、`medium`、`high` のいずれか。                     |
| `DirectDependencies`     | array of string | 直接依存。                                       |
| `TransitiveDependencies` | array of string | すべての推移的依存。                             |
| `Dependents`             | array of string | このモジュールに依存するモジュール。             |

### `CircularDependencyAnalysis` オブジェクト

| フィールド                  | 型      | 説明                                                  |
| -------------------------- | ------- | ----------------------------------------------------- |
| `HasCircularDependencies`  | boolean | 循環が存在する場合に `true`。                         |
| `TotalCycles`              | integer | 循環の数。                                            |
| `TotalModulesInCycles`     | integer | 循環に関与するモジュール。                            |
| `CircularDependencies`     | array   | `CircularDependency` オブジェクトの配列。             |
| `CycleBreakingSuggestions` | array of string | 循環を解消するための提案。                    |
| `CoreInfrastructure`       | array of string | 複数の循環に出現するモジュール。              |

`CircularDependency.Severity` 列挙: `low`、`medium`、`high`、`critical`。

### `CouplingAnalysis` オブジェクト

| フィールド               | 型      | 説明                                               |
| ----------------------- | ------- | -------------------------------------------------- |
| `AverageCoupling`       | number  | モジュール全体の平均結合度。                       |
| `CouplingDistribution`  | object  | 結合度値（integer キー）からカウントへのマップ。   |
| `HighlyCoupledModules`  | array of string | 高結合度のモジュール。                     |
| `LooselyCoupledModules` | array of string | 低結合度のモジュール。                     |
| `AverageInstability`    | number  | 平均不安定性。                                     |
| `StableModules`         | array of string | 低不安定性のモジュール。                   |
| `InstableModules`       | array of string | 高不安定性のモジュール。                   |
| `MainSequenceDeviation` | number  | メインシーケンスからの平均距離、`0`–`1`。          |
| `ZoneOfPain`            | array of string | 安定かつ具象のモジュール。                 |
| `ZoneOfUselessness`     | array of string | 不安定かつ抽象のモジュール。               |
| `MainSequence`          | array of string | 適切に配置されたモジュール。               |

### `ArchitectureAnalysis` オブジェクト

| フィールド                 | 型      | 説明                                                        |
| ------------------------- | ------- | ----------------------------------------------------------- |
| `ComplianceScore`         | number  | 準拠スコア、`0`–`1`。`1.0` = 完全準拠。                     |
| `TotalViolations`         | integer | 検出された違反の合計数。                                    |
| `TotalRules`              | integer | 評価されたルールの合計数。                                  |
| `LayerAnalysis`           | object \| null | レイヤー分析結果。                                   |
| `CohesionAnalysis`        | object \| null | パッケージ凝集度分析。                               |
| `ResponsibilityAnalysis`  | object \| null | SRP 違反分析。                                       |
| `Violations`              | array   | `ArchitectureViolation` オブジェクトの配列。                |
| `SeverityBreakdown`       | object  | 重大度からカウントへのマップ。                              |
| `Recommendations`         | array   | `ArchitectureRecommendation` オブジェクトの配列。           |
| `RefactoringTargets`      | array of string | リファクタリングが必要なモジュール。                |

`ArchitectureViolation.Type` 列挙: `layer`、`cycle`、`coupling`、`responsibility`、`cohesion`。

`ArchitectureViolation.Severity` 列挙: `info`、`warning`、`error`、`critical`。

## `suggestions` 配列

`Suggestion` オブジェクトの配列です。snake_case のフィールド名を使用します。

| フィールド     | 型      | 必須     | 説明                                              |
| ------------- | ------- | -------- | ------------------------------------------------- |
| `category`    | string  | yes      | 下記の列挙を参照。                                |
| `severity`    | string  | yes      | `critical`、`warning`、`info` のいずれか。        |
| `effort`      | string  | yes      | `easy`、`moderate`、`hard` のいずれか。            |
| `title`       | string  | yes      | 短い人間が読めるタイトル。                        |
| `description` | string  | yes      | 完全な説明。                                      |
| `steps`       | array of string | no | 実行可能な手順。空の場合は省略。             |
| `file_path`   | string  | no       | ソースファイル参照。                              |
| `function`    | string  | no       | 関数名参照。                                      |
| `class_name`  | string  | no       | クラス名参照。                                    |
| `start_line`  | integer | no       | 1始まりの行参照。`0` の場合は省略。               |
| `metric_value`| string  | no       | 観測されたメトリクス値（文字列）。                |
| `threshold`   | string  | no       | 閾値（文字列）。                                  |

`category` 列挙: `complexity`、`dead_code`、`clone`、`coupling`、`cohesion`、`dependency`、`architecture`。

提案は優先度（重大度 x 労力）でソートされます。正確なソート関数については `domain/suggestion.go` を参照してください。

## CSV スキーマ

CSV 出力は Go の `encoding/csv` パッケージを使用し、RFC 4180 に準拠したクォートで書き込まれます。

### `pyscn analyze --csv`

サマリーのみです。2列構成。リテラル UTF-8 文字列、型アノテーションはありません。

| 列       | 型     | 説明                     |
| -------- | ------ | ------------------------ |
| `Metric` | string | メトリクス名。           |
| `Value`  | string | メトリクス値（文字列）。 |

行（この固定順序で出力）:

```csv
Metric,Value
Health Score,<integer>
Grade,<A|B|C|D|F|N/A>
Total Files,<integer>
Analyzed Files,<integer>
Average Complexity,<float with 2 decimals>
High Complexity Count,<integer>
Dead Code Count,<integer>
Critical Dead Code,<integer>
Unique Fragments,<integer>
Clone Groups,<integer>
Code Duplication,<float with 2 decimals>
Total Classes Analyzed,<integer>
High Coupling (CBO) Classes,<integer>
Average CBO,<float with 2 decimals>
```

pyscn は現在、CLI を通じてアナライザごとの CSV スキーマを公開していません。`--csv` は上記のサマリーのみを生成します。検出結果の詳細については、`--json` または `--yaml` を使用してください。

## タイムスタンプとバージョニング

| フィールド      | フォーマット               | 備考                                                                    |
| -------------- | ------------------------- | ----------------------------------------------------------------------- |
| `generated_at` | RFC 3339 (ISO 8601)       | `time.Time` のシリアライズ。サブ秒精度やタイムゾーンオフセットを含む場合があります。 |
| `duration_ms`  | integer（ミリ秒）          | 実時間による分析所要時間。                                              |
| `version`      | string（セマンティックバージョン） | pyscn のリリースバージョン、例: `"0.14.0"`。                      |

## 各フォーマットの呼び出し方

`pyscn analyze` は `--json`、`--yaml`、`--csv`、`--html`（デフォルト）のいずれかを受け付けます。`--format` フラグはなく、個別の `complexity` / `deadcode` / `clone` / `deps` サブコマンドもありません。単一のアナライザを実行するには `--select` を使用してください。

```bash
pyscn analyze --json src/
pyscn analyze --yaml src/
pyscn analyze --csv  src/
pyscn analyze --html src/    # default
pyscn analyze --json --select complexity src/
pyscn analyze --csv  --select deadcode   src/
pyscn analyze --yaml --select clones     src/
```

出力ファイルは `.pyscn/reports/` に生成されます。パスとファイル名の詳細については [出力フォーマット](index.md) を参照してください。

## 関連ドキュメント

- [HTML レポート](html-report.md) — HTML 出力の仕様。
- [ヘルススコア](health-score.md) — `summary.health_score` およびカテゴリ別スコアの導出方法。
