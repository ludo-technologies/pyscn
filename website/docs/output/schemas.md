# Output Schemas

This specification defines the exact shape of JSON, YAML, and CSV output produced by pyscn. All field names, types, and semantics documented here are stable across patch releases within the same major version.

## Stability contract

| Guarantee          | Scope                                                                             |
| ------------------ | --------------------------------------------------------------------------------- |
| Stable             | field names, field types, field semantics, enum values                            |
| May change         | field ordering within an object, ordering of array elements, inclusion of new fields |
| Breaking           | removal or rename of fields, change of field type, removal of enum values         |

Breaking changes are restricted to major version bumps. Consumers MUST ignore unknown fields.

<!-- Field naming note: every object key in `pyscn analyze` JSON/YAML is snake_case. Releases up to 1.29.1 emitted Go-style PascalCase inside `complexity`, `cbo`, `lcom`, and `system`, and lowerCamelCase inside the `config` objects of `cbo`, `lcom`, and `community_analysis`; both were renamed to snake_case. -->

## Top-level structure (`pyscn analyze`)

JSON and YAML outputs serialize the `AnalyzeResponse` Go struct defined in `domain/analyze.go`. The top-level keys are:

```json
{
  "complexity":    { /* ComplexityResponse, present when enabled */ },
  "dead_code":     { /* DeadCodeResponse, present when enabled */ },
  "clone":         { /* CloneResponse, present when enabled */ },
  "cbo":           { /* CBOResponse, present when enabled */ },
  "lcom":          { /* LCOMResponse, present when enabled */ },
  "system":             { /* SystemAnalysisResponse, present when deps/arch enabled */ },
  "community_analysis": { /* CommunityAnalysisResult, present when communities enabled */ },
  "mock_data":          { /* MockDataResponse, present when enabled */ },
  "module_quality":     [ /* ModuleQualityMetrics array, omitted when empty */ ],
  "suggestions":   [ /* Suggestion array, omitted when empty */ ],
  "diagnostics":   [ /* AnalysisDiagnostic array, omitted when empty */ ],
  "failures":      [ /* AnalysisFailure array, omitted when empty */ ],
  "summary":       { /* AnalyzeSummary, always present */ },
  "generated_at":  "2026-04-14T10:18:23Z",
  "duration_ms":   2347,
  "version":       "0.14.0"
}
```

| Field         | Type              | Description                                            | Stability |
| ------------- | ----------------- | ------------------------------------------------------ | --------- |
| `complexity`  | object \| absent  | Present when complexity analysis ran.                  | stable    |
| `dead_code`   | object \| absent  | Present when dead code analysis ran.                   | stable    |
| `clone`       | object \| absent  | Present when clone detection ran.                      | stable    |
| `cbo`         | object \| absent  | Present when CBO analysis ran.                         | stable    |
| `lcom`        | object \| absent  | Present when LCOM analysis ran.                        | stable    |
| `system`             | object \| absent | Present when dependency or architecture analysis ran. | stable |
| `community_analysis` | object \| absent | Present when module community detection ran.          | stable |
| `mock_data`          | object \| absent | Present when mock data detection ran.                 | stable |
| `module_quality`     | array \| absent  | Per-module quality rollups. Omitted when no analyzer produced module data. | stable |
| `suggestions` | array \| absent   | Derived suggestions. Omitted when empty.               | stable    |
| `diagnostics` | array \| absent   | Project files that could not be read or parsed. See [`AnalysisDiagnostic`](#analysisdiagnostic-object). | stable |
| `failures`    | array \| absent   | Analyzer execution failures. Partial results may still be present. See [`AnalysisFailure`](#analysisfailure-object). | stable |
| `summary`     | object            | Always present. See [`summary`](#summary-object).      | stable    |
| `generated_at`| string (RFC 3339) | Analysis completion time.                              | stable    |
| `duration_ms` | integer           | Total analysis duration in milliseconds.               | stable    |
| `version`     | string            | pyscn semantic version.                                | stable    |

## `AnalysisDiagnostic` object { #analysisdiagnostic-object }

| Field | Type | Description |
| --- | --- | --- |
| `file_path` | string | Source file that could not be analyzed. |
| `code` | string | One of: `read_error`, `parse_error`. Cancellation uses the phase that observed it: `read_error` before parsing starts and `parse_error` during parsing. |
| `message` | string | Human-readable cause. |

## `AnalysisFailure` object { #analysisfailure-object }

| Field | Type | Description |
| --- | --- | --- |
| `analysis` | string | One of: `complexity`, `deadcode`, `clones`, `cbo`, `lcom`, `system`, `communities`, `mockdata`, `di`. |
| `code` | string | Currently `execution_error`. |
| `message` | string | Human-readable analyzer failure. |
| `file_path` | string \| absent | Source file when the failure belongs to one file. |

## `summary` object { #summary-object }

Mirrors `domain.AnalyzeSummary`. All numeric counters default to `0` when the corresponding analyzer is disabled. All fields are always present.

### File statistics

| Field            | Type    | Description                                      |
| ---------------- | ------- | ------------------------------------------------ |
| `total_files`    | integer | Number of Python files required by the requested analysis set. Dependency and community analysis include matching `.pyi` modules. |
| `analyzed_files` | integer | Number of files successfully analyzed.           |
| `skipped_files`  | integer | Files dropped because they could not be read or parsed. A non-zero value means every score below covers less than `total_files`. |

### Analyzer status flags

| Field                | Type    | Description                                                |
| -------------------- | ------- | ---------------------------------------------------------- |
| `complexity_enabled` | boolean | `true` if complexity analysis produced results.            |
| `dead_code_enabled`  | boolean | `true` if dead code analysis produced results.             |
| `clone_enabled`      | boolean | `true` if clone detection produced results.                |
| `cbo_enabled`        | boolean | `true` if CBO analysis produced results.                   |
| `lcom_enabled`       | boolean | `true` if LCOM analysis produced results.                  |
| `deps_enabled`       | boolean | `true` if dependency analysis produced results.            |
| `arch_enabled`        | boolean | `true` if architecture validation produced results.    |
| `communities_enabled` | boolean | `true` if module community detection produced results. |
| `mock_data_enabled`   | boolean | `true` if mock data detection produced results.      |

### Complexity metrics

| Field                   | Type    | Description                                      |
| ----------------------- | ------- | ------------------------------------------------ |
| `total_functions`       | integer | Total functions analyzed.                        |
| `total_class_scopes`    | integer | Executable class suites analyzed.                |
| `average_complexity`    | number  | Mean cyclomatic complexity. `0` when no functions. |
| `high_complexity_count` | integer | Functions with complexity > 10 (medium threshold). |
| `max_class_complexity` | integer | Highest class-suite cyclomatic complexity. |
| `max_class_cognitive_complexity` | integer | Highest class-suite cognitive complexity. |
| `max_class_nesting_depth` | integer | Highest class-suite nesting depth. |
| `high_complexity_class_scope_count` | integer | Class scopes classified as high risk. |

### Dead code metrics

| Field                | Type    | Description                                  |
| -------------------- | ------- | -------------------------------------------- |
| `dead_code_count`    | integer | Total findings.                              |
| `critical_dead_code` | integer | Findings with severity `critical`.           |
| `warning_dead_code`  | integer | Findings with severity `warning`.            |
| `info_dead_code`     | integer | Findings with severity `info`.               |

### Clone metrics

| Field                         | Type    | Description                                               |
| ----------------------------- | ------- | --------------------------------------------------------- |
| `total_clones`                | integer | Distinct code fragments identified as clones.             |
| `clone_pairs`                 | integer | Number of clone pairs.                                    |
| `clone_groups`                | integer | Number of clone groups.                                   |
| `code_duplication_percentage` | number  | Estimated duplication ratio, `0`–`100`.                   |

### CBO metrics

| Field                     | Type    | Description                                               |
| ------------------------- | ------- | --------------------------------------------------------- |
| `cbo_classes`             | integer | Total classes analyzed.                                   |
| `high_coupling_classes`   | integer | Classes with CBO > 7.                                     |
| `medium_coupling_classes` | integer | Classes with 3 < CBO ≤ 7.                                 |
| `average_coupling`        | number  | Mean CBO value.                                           |

### LCOM metrics

| Field                 | Type    | Description                                  |
| --------------------- | ------- | -------------------------------------------- |
| `lcom_classes`        | integer | Total classes analyzed.                      |
| `high_lcom_classes`   | integer | Classes with LCOM4 > 5.                      |
| `medium_lcom_classes` | integer | Classes with 2 < LCOM4 ≤ 5.                  |
| `average_lcom`        | number  | Mean LCOM4 value.                            |

### Dependency metrics

| Field                          | Type    | Description                                                    |
| ------------------------------ | ------- | -------------------------------------------------------------- |
| `deps_total_modules`           | integer | Total modules analyzed.                                        |
| `deps_modules_in_cycles`       | integer | Modules participating in at least one circular dependency.     |
| `deps_max_depth`               | integer | Edges along the longest load-time dependency chain, the first entry of `LongestChains`. |
| `deps_main_sequence_deviation` | number  | Average distance from Martin's main sequence, `0`–`1`.         |

### Architecture metrics

| Field             | Type   | Description                                                       |
| ----------------- | ------ | ----------------------------------------------------------------- |
| `arch_compliance` | number | Architecture compliance ratio, `0`–`1`. Severity-weighted (`error × 5 + warning × 1`); see `system.ArchitectureAnalysis.WeightedViolations` for the numerator. |

### Mock data metrics

| Field                   | Type    | Description                                         |
| ----------------------- | ------- | --------------------------------------------------- |
| `mock_data_count`       | integer | Total mock data findings.                           |
| `mock_data_error_count` | integer | Findings at error severity.                         |
| `mock_data_warning_count` | integer | Findings at warning severity.                     |
| `mock_data_info_count`  | integer | Findings at info severity.                          |

### Health scoring

| Field                | Type    | Description                                                        |
| -------------------- | ------- | ------------------------------------------------------------------ |
| `health_score`       | integer | Composite score, `0`–`100`. See [Health Score](health-score.md).   |
| `grade`              | string  | Letter grade. One of: `A`, `B`, `C`, `D`, `F`, `N/A`.              |
| `complexity_score`   | integer | Per-category score, `0`–`100`.                                     |
| `dead_code_score`    | integer | Per-category score, `0`–`100`.                                     |
| `duplication_score`  | integer | Per-category score, `0`–`100`.                                     |
| `coupling_score`     | integer | Per-category score, `0`–`100`.                                     |
| `cohesion_score`     | integer | Per-category score, `0`–`100`.                                     |
| `dependency_score`   | integer | Per-category score, `0`–`100`.                                     |
| `architecture_score` | integer | Per-category score, `0`–`100`.                                     |

## `module_quality` array

Each entry joins module identity and size from dependency analysis with analyzer-owned complexity and dead-code rollups. The rollups are calculated before presentation filters, do not run another source-analysis pass, and do not contribute a new health-score category.

| Field | Type | Description |
| --- | --- | --- |
| `module_name` | string \| absent | Importable Python module name when dependency analysis resolved it. |
| `file_path` | string | Source path as reported by the analyzer. Relative input paths remain relative. |
| `lines_of_code` | integer | Physical module line count from dependency analysis, or `0` when unavailable. |
| `function_count` | integer | Total module function count from dependency analysis, or `0` when unavailable. |
| `analyzed_function_count` | integer | Function complexity records before `min_complexity` and `report_unchanged` filters. The `<module>` pseudo-record is excluded. |
| `average_complexity` | number | Mean cyclomatic complexity across all function records. |
| `average_cognitive_complexity` | number | Mean cognitive complexity across all function records. |
| `max_complexity` | integer | Maximum cyclomatic complexity before presentation filters. |
| `high_risk_function_count` | integer | Complexity records classified as high risk before presentation filters. |
| `exception_handler_count` | integer | Sum of exception handlers across all function records. |
| `dead_code_finding_count` | integer | Detector findings before `min_severity` filtering. |
| `dead_code_block_count` | integer | Distinct unreachable CFG blocks represented by detector findings before `min_severity` filtering. |

Entries are ordered by high-risk function count, maximum complexity, average complexity, dead-code findings, then file path. This places the most actionable modules first while retaining deterministic ties.

## `complexity.by_directory` array

Each entry groups the reported `complexity.Functions` population by its direct directory relative to the analysis root. Counts and averages therefore reconcile with the function list after `min_complexity` and `report_unchanged` filters. The aggregation performs no source reads or analyzer passes.

| Field | Type | Description |
| --- | --- | --- |
| `directory_path` | string | Directory relative to the analyzed root. The root entry is `.`. |
| `function_count` | integer | Reported `complexity.Functions` entries whose files are directly in this directory, including a `<module>` pseudo-entry when it survives presentation filters. |
| `average_complexity` | number | Mean cyclomatic complexity of those functions. |
| `max_complexity` | integer | Maximum cyclomatic complexity of those functions. |
| `high_risk_function_count` | integer | Reported functions classified as high risk. |
| `average_nesting_depth` | number | Mean maximum nesting depth of the reported functions. |
| `max_nesting_depth` | integer | Maximum nesting depth in the directory. |

When multiple input paths are supplied, their common directory is the analyzed root. File inputs participate through their parent directories. The root entry is `.`. Entries are ordered by high-risk count, maximum complexity, average complexity, then directory path.

When complexity analysis completes with no reported functions, `by_directory` is present as an empty array in JSON and YAML.

## `complexity` object

Mirrors `domain.ComplexityResponse`.

```json
{
  "functions": [ /* FunctionComplexity array */ ],
  "class_scopes": [ /* FunctionComplexity array; omitted when empty */ ],
  "by_directory": [ /* DirectoryComplexityMetrics array; empty when no functions are reported */ ],
  "summary": { /* ComplexitySummary */ },
  "raw_metrics": [ /* RawMetrics array, present when computed */ ],
  "raw_metrics_summary": { /* RawMetricsSummary, present when computed */ },
   "warnings": [ "..." ],
   "errors": [ "..." ],
   "failures": [ /* AnalysisFailure array, absent when empty */ ],
   "generated_at": "2026-04-14T10:18:23Z",
  "version": "0.14.0",
  "config": null
}
```

The standalone complexity formatter uses `by_directory` at the report root beside `results`, optional `class_scopes`, `summary`, and `metadata`. `results` retains the established module/function collection; `class_scopes` contains executable class suites. Directory entries remain function-only and their semantics are identical to unified output.

### `functions[]` and `class_scopes[]` element (`FunctionComplexity`)

| Field          | Type    | Description                                           |
| -------------- | ------- | ----------------------------------------------------- |
| `name`         | string  | Qualified scope name. `<module>` for module-level code. |
| `scope_kind`   | string  | Required execution owner: `module`, `function`, or `class`. |
| `file_path`    | string  | Path to source file.                                  |
| `start_line`   | integer | 1-based start line.                                   |
| `start_column` | integer | 0-based start column.                                 |
| `end_line`     | integer | 1-based end line.                                     |
| `metrics`      | object  | See [`ComplexityMetrics`](#complexitymetrics-object). |
| `risk_level`   | string  | One of: `low`, `medium`, `high`.                      |

### `ComplexityMetrics` object { #complexitymetrics-object }

| Field                  | Type    | Description                             |
| ---------------------- | ------- | --------------------------------------- |
| `complexity`           | integer | McCabe cyclomatic complexity.           |
| `cognitive_complexity` | integer | Cognitive complexity (SonarQube-style). |
| `nodes`                | integer | CFG node count.                         |
| `edges`                | integer | CFG edge count.                         |
| `nesting_depth`        | integer | Maximum nesting depth.                  |
| `if_statements`        | integer | Count of `if` statements.               |
| `loop_statements`      | integer | Count of `for`/`while` loops.           |
| `exception_handlers`   | integer | Count of `except` clauses.              |
| `switch_cases`         | integer | Count of `match` cases (Python 3.10+).  |

### `summary` object (`ComplexitySummary`)

| Field                              | Type    | Description                                                                                                |
| ---------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------- |
| `total_functions`                  | integer | Total module and function scopes analyzed.                                                                 |
| `total_class_scopes`               | integer | Total executable class suites analyzed.                                                                    |
| `functions_parsed`                 | integer | Compatibility count matching the complete `total_functions` population.                                   |
| `average_complexity`               | number  | Arithmetic mean of `complexity` across module and function scopes.                                         |
| `average_cognitive_complexity`     | number  | Arithmetic mean of `cognitive_complexity` across module and function scopes.                               |
| `average_nesting_depth`            | number  | Arithmetic mean of `nesting_depth` across module and function scopes.                                      |
| `max_complexity`                   | integer | Highest complexity among module and function scopes.                                                       |
| `min_complexity`                   | integer | Lowest complexity among module and function scopes.                                                        |
| `max_class_complexity`             | integer | Highest class-suite cyclomatic complexity.                                                                 |
| `max_class_cognitive_complexity`   | integer | Highest class-suite cognitive complexity.                                                                  |
| `max_class_nesting_depth`          | integer | Highest class-suite nesting depth.                                                                          |
| `high_risk_class_scopes`           | integer | Class suites classified as high risk.                                                                       |
| `files_analyzed`                   | integer | Files that were parsed and contributed to the metrics above.                                               |
| `total_files`                      | integer | Files the request covered, parsed or not.                                                                  |
| `skipped_files`                    | integer | Files dropped because they could not be read or parsed. Their contents are absent from every metric above. |
| `low_risk_functions`               | integer | Module and function scopes with `risk_level = low`.                                                        |
| `medium_risk_functions`            | integer | Module and function scopes with `risk_level = medium`.                                                     |
| `high_risk_functions`              | integer | Module and function scopes with `risk_level = high`.                                                       |
| `complexity_distribution`          | object  | Function-only histogram keyed by complexity bucket (string) to count (integer), or `null`.                 |

Class-scope counts and maxima are additive hotspot metrics. Adding a class scope does not change the legacy function collections, counts, averages, extrema, risk distribution, complexity distribution, module or directory rollups, or health score.

### `raw_metrics[]` element (`RawMetrics`)

| Field             | Type    | Description                                         |
| ----------------- | ------- | --------------------------------------------------- |
| `file_path`       | string  | Path to source file.                                |
| `sloc`            | integer | Source lines of code (non-blank, non-comment).      |
| `lloc`            | integer | Logical lines of code.                              |
| `comment_lines`   | integer | Lines containing comments.                          |
| `docstring_lines` | integer | Lines inside docstrings.                            |
| `blank_lines`     | integer | Empty or whitespace-only lines.                     |
| `total_lines`     | integer | Total physical lines.                               |
| `comment_ratio`   | number  | `(comment_lines + docstring_lines) / total_lines`, `0`–`1`. |


## `dead_code` object

Mirrors `domain.DeadCodeResponse`. Uses snake_case field names throughout.

```json
{
  "files": [ /* FileDeadCode array */ ],
  "summary": { /* DeadCodeSummary */ },
   "warnings": null,
   "errors": null,
   "failures": [ /* AnalysisFailure array, absent when empty */ ],
   "generated_at": "",
  "version": "",
  "config": null
}
```

### `files[]` element (`FileDeadCode`)

| Field               | Type    | Description                                    |
| ------------------- | ------- | ---------------------------------------------- |
| `file_path`         | string  | Path to source file.                           |
| `functions`             | array   | Per-function results. Existing field; remains function-only. |
| `class_scopes`          | array \| absent | Executable class-suite results. Uses the same row model as `functions`. |
| `total_findings`        | integer | Sum of findings across functions and class scopes in this file. |
| `total_functions`       | integer | Functions analyzed in this file. Existing field; remains function-only. |
| `affected_functions`    | integer | Functions with at least one finding. Existing field; remains function-only. |
| `total_class_scopes`    | integer | Executable class suites analyzed in this file. |
| `affected_class_scopes` | integer | Executable class suites with at least one finding. |
| `dead_code_ratio`       | number  | Dead blocks / total blocks across both scope collections, `0`–`1`. |

### `files[].functions[]` and `files[].class_scopes[]` element (`FunctionDeadCode`)

| Field             | Type    | Description                                  |
| ----------------- | ------- | -------------------------------------------- |
| `name`            | string  | Function or class name.                      |
| `scope_kind`      | string  | Required execution owner: `function` in `functions`; `class` in `class_scopes`. |
| `file_path`       | string  | Path to source file.                         |
| `findings`        | array   | Findings in this execution scope (see below). |
| `total_blocks`    | integer | Total CFG blocks in the execution scope.     |
| `dead_blocks`     | integer | Unreachable CFG blocks.                      |
| `reachable_ratio` | number  | `(total_blocks - dead_blocks) / total_blocks`, `0`–`1`. |
| `critical_count`  | integer | Findings of severity `critical`.             |
| `warning_count`   | integer | Findings of severity `warning`.              |
| `info_count`      | integer | Findings of severity `info`.                 |

### `findings[]` element (`DeadCodeFinding`)

| Field           | Type    | Description                                                   |
| --------------- | ------- | ------------------------------------------------------------- |
| `location`      | object  | See [`DeadCodeLocation`](#deadcodelocation-object).           |
| `function_name` | string  | Enclosing execution-scope name (historical field name).       |
| `scope_kind`    | string  | Required execution owner: `function` or `class`.              |
| `code`          | string  | The dead source code snippet.                                 |
| `reason`        | string  | Classification — see enumeration below.                       |
| `severity`      | string  | One of: `critical`, `warning`, `info`.                        |
| `description`   | string  | Human-readable description.                                   |
| `context`       | array of string \| absent | Surrounding source lines. Present when `--show-context`. |
| `block_id`      | string \| absent | CFG block identifier.                                  |

`reason` enumeration:

| Value                 | Meaning                                      |
| --------------------- | -------------------------------------------- |
| `after_return`        | Code following a `return` statement.         |
| `after_break`         | Code following a `break` statement.          |
| `after_continue`      | Code following a `continue` statement.       |
| `after_raise`         | Code following a `raise` statement.          |
| `unreachable_branch`  | Conditional branch that is never taken.      |

### `DeadCodeLocation` object { #deadcodelocation-object }

| Field          | Type    | Description                |
| -------------- | ------- | -------------------------- |
| `file_path`    | string  | Path to source file.       |
| `start_line`   | integer | 1-based start line.        |
| `end_line`     | integer | 1-based end line.          |
| `start_column` | integer | 0-based start column.      |
| `end_column`   | integer | 0-based end column.        |

### `summary` object (`DeadCodeSummary`)

| Field                      | Type    | Description                                      |
| -------------------------- | ------- | ------------------------------------------------ |
| `total_files`              | integer | Files analyzed.                                  |
| `total_functions`          | integer | Functions analyzed.                              |
| `total_findings`           | integer | Total findings across functions and class scopes. |
| `files_with_dead_code`     | integer | Files with at least one finding.                 |
| `functions_with_dead_code` | integer | Functions with at least one finding.             |
| `total_class_scopes`       | integer | Executable class suites analyzed.                |
| `class_scopes_with_dead_code` | integer | Executable class suites with at least one finding. |
| `critical_findings`        | integer | Findings with severity `critical`.               |
| `warning_findings`         | integer | Findings with severity `warning`.                |
| `info_findings`            | integer | Findings with severity `info`.                   |
| `findings_by_reason`       | object \| null | Histogram keyed by `reason` value.         |
| `total_blocks`             | integer | CFG blocks across functions and class scopes.    |
| `dead_blocks`              | integer | Unreachable CFG blocks across functions and class scopes. |
| `overall_dead_ratio`       | number  | `dead_blocks / total_blocks`, `0`–`1`.           |

## `clone` object

Mirrors `domain.CloneResponse`. Uses snake_case field names throughout.

```json
{
  "clones": [ /* Clone array, or null */ ],
  "clone_pairs": [ /* ClonePair array, or null */ ],
  "clone_groups": [ /* CloneGroup array, or null */ ],
  "statistics": { /* CloneStatistics */ },
  "duration_ms": 123,
  "success": true,
  "error": "",
  "failures": [ /* AnalysisFailure array, absent when empty */ ]
}
```

### `clones[]` element (`Clone`)

| Field        | Type    | Description                                                  |
| ------------ | ------- | ------------------------------------------------------------ |
| `id`         | integer | Clone identifier, unique within the response.                |
| `type`       | integer | Clone type as integer: `1`, `2`, `3`, or `4`.                |
| `location`   | object  | See [`CloneLocation`](#clonelocation-object).                |
| `content`    | string  | Raw source text. Present only when `--show-content` set.     |
| `hash`       | string  | Fingerprint hash (algorithm depends on clone type).          |
| `size`       | integer | Number of AST nodes.                                         |
| `line_count` | integer | Line count of the fragment.                                  |
| `complexity` | integer | Cyclomatic complexity of the fragment.                       |

`type` enumeration (integer values):

| Value | Meaning                                                              |
| ----- | -------------------------------------------------------------------- |
| `1`   | Type-1: identical except whitespace/comments.                        |
| `2`   | Type-2: syntactically identical, different identifiers/literals.     |
| `3`   | Type-3: structurally similar with modifications.                     |
| `4`   | Type-4: semantically equivalent, syntactically different.            |

### `CloneLocation` object { #clonelocation-object }

| Field        | Type    | Description              |
| ------------ | ------- | ------------------------ |
| `file_path`  | string  | Path to source file.     |
| `start_line` | integer | 1-based start line.      |
| `end_line`   | integer | 1-based end line.        |
| `start_col`  | integer | 0-based start column.    |
| `end_col`    | integer | 0-based end column.      |

### `clone_pairs[]` element (`ClonePair`)

| Field        | Type    | Description                                            |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | Pair identifier.                                       |
| `clone1`     | object  | First clone (`Clone` object).                          |
| `clone2`     | object  | Second clone (`Clone` object).                         |
| `similarity` | number  | Similarity score, `0`–`1`.                             |
| `distance`   | number  | Tree edit distance (Type-3) or `0` otherwise.          |
| `type`       | integer | Clone type (same enumeration as `clones[].type`).      |
| `confidence` | number  | Detector confidence, `0`–`1`.                          |

### `clone_groups[]` element (`CloneGroup`)

| Field        | Type    | Description                                            |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | Group identifier.                                      |
| `clones`     | array   | Member `Clone` objects.                                |
| `type`       | integer | Dominant clone type.                                   |
| `similarity` | number  | Representative similarity, `0`–`1`.                    |
| `size`       | integer | Number of members (`len(clones)`).                     |

### `statistics` object (`CloneStatistics`)

| Field                | Type    | Description                                              |
| -------------------- | ------- | -------------------------------------------------------- |
| `total_fragments`    | integer | All extracted fragments (functions, classes, etc.).      |
| `total_clones`       | integer | Fragments classified as clones.                          |
| `total_clone_pairs`  | integer | Number of pairs detected.                                |
| `total_clone_groups` | integer | Number of groups.                                        |
| `clones_by_type`     | object \| null | Map from type label (`Type-1`…`Type-4`) to count.  |
| `average_similarity` | number  | Mean similarity across pairs, `0`–`1`.                   |
| `lines_analyzed`     | integer | Total source lines considered.                           |
| `nodes_analyzed`     | integer | Total AST nodes considered.                              |
| `files_analyzed`     | integer | Distinct files contributing fragments.                   |

Other `CloneResponse` fields:

| Field         | Type    | Description                                        |
| ------------- | ------- | -------------------------------------------------- |
| `duration_ms` | integer | Clone detection duration in milliseconds.          |
| `success`     | boolean | `true` on normal completion.                       |
| `error`       | string \| absent | Error message if `success=false`.         |
| `errors`      | array \| absent  | Per-file failures. A file listed here was skipped while the run itself succeeded, so its contents are absent from `statistics`. |

## `cbo` object

Mirrors `domain.CBOResponse`.

```json
{
  "classes": [ /* ClassCoupling array */ ],
  "summary": { /* CBOSummary */ },
   "warnings": null,
   "errors": null,
   "failures": [ /* AnalysisFailure array, absent when empty */ ],
   "generated_at": "",
  "version": "",
  "config": null
}
```

### `classes[]` element (`ClassCoupling`)

| Field          | Type                    | Description                             |
| -------------- | ----------------------- | --------------------------------------- |
| `name`         | string                  | Class name.                             |
| `file_path`    | string                  | Path to source file.                    |
| `start_line`   | integer                 | 1-based start line.                     |
| `end_line`     | integer                 | 1-based end line.                       |
| `metrics`      | object                  | See [`CBOMetrics`](#cbometrics-object). |
| `risk_level`   | string                  | One of: `low`, `medium`, `high`.        |
| `is_abstract`  | boolean                 | `true` if the class is abstract.        |
| `base_classes` | array of string \| null | Direct base classes.                    |

### `CBOMetrics` object { #cbometrics-object }

| Field                           | Type                    | Description                                          |
| ------------------------------- | ----------------------- | ---------------------------------------------------- |
| `coupling_count`                | integer                 | CBO value: distinct classes this class depends on.   |
| `inheritance_dependencies`      | integer                 | Dependencies from base classes.                      |
| `type_hint_dependencies`        | integer                 | Dependencies from type annotations.                  |
| `instantiation_dependencies`    | integer                 | Dependencies from object instantiation.              |
| `attribute_access_dependencies` | integer                 | Dependencies from method calls and attribute access. |
| `import_dependencies`           | integer                 | Dependencies from explicit imports.                  |
| `dependent_classes`             | array of string \| null | Names of coupled classes.                            |

### `summary` object (`CBOSummary`)

| Field                        | Type                    | Description                                        |
| ---------------------------- | ----------------------- | -------------------------------------------------- |
| `total_classes`              | integer                 | Total classes analyzed.                            |
| `average_cbo`                | number                  | Mean CBO.                                          |
| `max_cbo`                    | integer                 | Maximum observed CBO.                              |
| `min_cbo`                    | integer                 | Minimum observed CBO.                              |
| `classes_analyzed`           | integer                 | Classes with valid metrics.                        |
| `files_analyzed`             | integer                 | Files contributing at least one class.             |
| `low_risk_classes`           | integer                 | Classes with CBO ≤ low threshold (default `3`).    |
| `medium_risk_classes`        | integer                 | Classes with low < CBO ≤ medium threshold.         |
| `high_risk_classes`          | integer                 | Classes with CBO > medium threshold (default `7`). |
| `cbo_distribution`           | object \| null          | Histogram keyed by bucket label to count.          |
| `most_coupled_classes`       | array \| null           | Top 10 classes by CBO (`ClassCoupling`).           |
| `most_depended_upon_classes` | array of string \| null | Classes with highest in-degree.                    |

## `lcom` object

Mirrors `domain.LCOMResponse`.

```json
{
  "classes": [ /* ClassCohesion array */ ],
  "summary": { /* LCOMSummary */ },
  "warnings": null,
  "errors": null,
  "failures": [ /* AnalysisFailure array, absent when empty */ ],
  "generated_at": "",
  "version": "",
  "config": null
}
```

### `classes[]` element (`ClassCohesion`)

| Field        | Type    | Description                               |
| ------------ | ------- | ----------------------------------------- |
| `name`       | string  | Class name.                               |
| `file_path`  | string  | Path to source file.                      |
| `start_line` | integer | 1-based start line.                       |
| `end_line`   | integer | 1-based end line.                         |
| `metrics`    | object  | See [`LCOMMetrics`](#lcommetrics-object). |
| `risk_level` | string  | One of: `low`, `medium`, `high`.          |

### `LCOMMetrics` object { #lcommetrics-object }

| Field                | Type                             | Description                                                                                                                                           |
| -------------------- | -------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `lcom4`              | integer                          | Connected components in the method-variable graph.                                                                                                    |
| `total_methods`      | integer                          | All methods in the class.                                                                                                                             |
| `excluded_methods`   | integer                          | Methods excluded from the LCOM4 graph (`@classmethod`, `@staticmethod`, `@abstractmethod`, and constructors: `__init__`, `__new__`, `__post_init__`). |
| `instance_variables` | integer                          | Distinct `self.x` variables accessed.                                                                                                                 |
| `method_groups`      | array of array of string \| null | Method names grouped by connected component.                                                                                                          |

### `summary` object (`LCOMSummary`)

| Field                    | Type           | Description                                          |
| ------------------------ | -------------- | ---------------------------------------------------- |
| `total_classes`          | integer        | Classes analyzed.                                    |
| `average_lcom`           | number         | Mean LCOM4.                                          |
| `max_lcom`               | integer        | Maximum observed LCOM4.                              |
| `min_lcom`               | integer        | Minimum observed LCOM4.                              |
| `classes_analyzed`       | integer        | Classes with valid metrics.                          |
| `files_analyzed`         | integer        | Files contributing at least one class.               |
| `low_risk_classes`       | integer        | Classes with LCOM4 ≤ low threshold (default `2`).    |
| `medium_risk_classes`    | integer        | Classes with low < LCOM4 ≤ medium threshold.         |
| `high_risk_classes`      | integer        | Classes with LCOM4 > medium threshold (default `5`). |
| `lcom_distribution`      | object \| null | Histogram keyed by bucket label to count.            |
| `least_cohesive_classes` | array \| null  | Top 10 classes by LCOM4 (`ClassCohesion`).           |

## `system` object

Mirrors `domain.SystemAnalysisResponse`.

```json
{
  "dependency_analysis":   { /* DependencyAnalysisResult, or null */ },
  "architecture_analysis": { /* ArchitectureAnalysisResult, or null */ },
  "summary":              { /* SystemAnalysisSummary */ },
  "issues":               [ /* SystemIssue array */ ],
  "recommendations":      [ /* SystemRecommendation array */ ],
   "warnings":             [ ],
   "errors":               [ ],
   "failures":             [ /* AnalysisFailure array, absent when empty */ ],
   "generated_at":          "0001-01-01T00:00:00Z",
  "duration":             0,
  "version":              "",
  "config":               null
}
```

### `summary` object (`SystemAnalysisSummary`)

| Field                       | Type    | Description                              |
| --------------------------- | ------- | ---------------------------------------- |
| `total_modules`             | integer | Total modules analyzed.                  |
| `total_packages`            | integer | Total packages.                          |
| `total_dependencies`        | integer | Total dependency edges.                  |
| `project_root`              | string  | Project root directory.                  |
| `overall_quality_score`     | number  | Composite quality score, `0`–`100`.      |
| `maintainability_score`     | number  | Average maintainability index.           |
| `architecture_score`        | number  | Architecture compliance score.           |
| `modularity_score`          | number  | System modularity score.                 |
| `technical_debt_hours`      | number  | Total estimated technical debt in hours. |
| `average_coupling`          | number  | Average module coupling.                 |
| `average_instability`       | number  | Average instability (I).                 |
| `cyclic_dependencies`       | integer | Modules participating in cycles.         |
| `architecture_violations`   | integer | Count of architecture rule violations.   |
| `high_risk_modules`         | integer | Modules flagged high risk.               |
| `critical_issues`           | integer | Critical issue count.                    |
| `refactoring_candidates`    | integer | Modules needing refactoring.             |
| `architecture_improvements` | integer | Suggested architecture improvements.     |

### `dependency_analysis` object

| Field                   | Type            | Description                                                                             |
| ----------------------- | --------------- | --------------------------------------------------------------------------------------- |
| `total_modules`         | integer         | Total modules in the dependency graph.                                                  |
| `total_dependencies`    | integer         | Total edges.                                                                            |
| `root_modules`          | array of string | Modules with no outgoing dependencies.                                                  |
| `leaf_modules`          | array of string | Modules with no incoming dependencies.                                                  |
| `module_metrics`        | object          | Map from module name to `ModuleDependencyMetrics`.                                      |
| `dependency_matrix`     | object          | Map from module to map of module to boolean.                                            |
| `circular_dependencies` | object          | Cycle detection results; contains `Cycles` (array) and `TotalCycles` (integer).         |
| `coupling_analysis`     | object          | Per-module coupling metrics: `Ca`, `Ce`, `Instability`, `Abstractness`, `Distance`.     |
| `longest_chains`        | array           | Top load-time SCC-condensed paths, expanded to real module dependency paths.            |
| `max_depth`             | integer         | Edges along the longest load-time dependency chain, the first entry of `LongestChains`. |

### `ModuleDependencyMetrics` object

| Field                     | Type            | Description                                      |           |                                          |
| ------------------------- | --------------- | ------------------------------------------------ | --------- | ---------------------------------------- |
| `module_name`             | string          | Fully qualified module name.                     |           |                                          |
| `package`                 | string          | Parent package.                                  |           |                                          |
| `file_path`               | string          | Path to source file.                             |           |                                          |
| `is_package`              | boolean         | `true` if this is a package (has `__init__.py`). |           |                                          |
| `lines_of_code`           | integer         | Total lines of code.                             |           |                                          |
| `function_count`          | integer         | Number of functions.                             |           |                                          |
| `class_count`             | integer         | Number of classes.                               |           |                                          |
| `abstract_class_count`    | integer         | Number of abstract classes.                      |           |                                          |
| `public_interface`        | array of string | Names in `__all__` or top-level public names.    |           |                                          |
| `afferent_coupling`       | integer         | Ca — modules depending on this one.              |           |                                          |
| `efferent_coupling`       | integer         | Ce — modules this one depends on.                |           |                                          |
| `instability`             | number          | `I = Ce / (Ca + Ce)`, `0`–`1`.                   |           |                                          |
| `abstractness`            | number          | A — abstract classes / total classes, `0`–`1`.   |           |                                          |
| `distance`                | number          | `D =                                             | A + I - 1 | `, `0`–`1`. Distance from main sequence. |
| `maintainability`         | number          | Maintainability index, `0`–`100`.                |           |                                          |
| `technical_debt`          | number          | Estimated technical debt in hours.               |           |                                          |
| `risk_level`              | string          | One of: `low`, `medium`, `high`.                 |           |                                          |
| `direct_dependencies`     | array of string | Direct dependencies.                             |           |                                          |
| `transitive_dependencies` | array of string | All transitive dependencies.                     |           |                                          |
| `dependents`              | array of string | Modules depending on this one.                   |           |                                          |

### `CircularDependencyAnalysis` object

| Field                        | Type            | Description                            |
| ---------------------------- | --------------- | -------------------------------------- |
| `has_circular_dependencies`  | boolean         | `true` if any cycles exist.            |
| `total_cycles`               | integer         | Number of cycles.                      |
| `total_modules_in_cycles`    | integer         | Modules involved in cycles.            |
| `circular_dependencies`      | array           | Array of `CircularDependency` objects. |
| `cycle_breaking_suggestions` | array of string | Suggestions for breaking cycles.       |
| `core_infrastructure`        | array of string | Modules appearing in multiple cycles.  |

`CircularDependency.Severity` enumeration: `low`, `medium`, `high`, `critical`.

### `coupling_analysis` object

| Field                     | Type            | Description                                     |
| ------------------------- | --------------- | ----------------------------------------------- |
| `average_coupling`        | number          | Average coupling across modules.                |
| `coupling_distribution`   | object          | Map from coupling value (integer key) to count. |
| `highly_coupled_modules`  | array of string | Modules with high coupling.                     |
| `loosely_coupled_modules` | array of string | Modules with low coupling.                      |
| `average_instability`     | number          | Average instability.                            |
| `stable_modules`          | array of string | Low-instability modules.                        |
| `instable_modules`        | array of string | High-instability modules.                       |
| `main_sequence_deviation` | number          | Average distance from main sequence, `0`–`1`.   |
| `zone_of_pain`            | array of string | Stable + concrete modules.                      |
| `zone_of_uselessness`     | array of string | Unstable + abstract modules.                    |
| `main_sequence`           | array of string | Well-positioned modules.                        |

### `architecture_analysis` object

| Field                     | Type            | Description                                                                                                               |
| ------------------------- | --------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `compliance_score`        | number          | Compliance score, `0`–`1`. Computed as `max(0, 1 - WeightedViolations / TotalRules)`; `1.0` when no rules were evaluated. |
| `total_violations`        | integer         | Raw count of violations (one per entry in `Violations`).                                                                  |
| `weighted_violations`     | integer         | Severity-weighted violation count used as the `ComplianceScore` numerator: `error × 5 + warning × 1`.                     |
| `total_rules`             | integer         | Total rule invocations evaluated (the `ComplianceScore` denominator).                                                     |
| `layer_analysis`          | object \| null  | Layer analysis results.                                                                                                   |
| `cohesion_analysis`       | object \| null  | Package cohesion analysis.                                                                                                |
| `responsibility_analysis` | object \| null  | SRP violation analysis.                                                                                                   |
| `violations`              | array           | Array of `ArchitectureViolation` objects.                                                                                 |
| `severity_breakdown`      | object          | Map from severity to count.                                                                                               |
| `recommendations`         | array           | Array of `ArchitectureRecommendation` objects.                                                                            |
| `refactoring_targets`     | array of string | Modules needing refactoring.                                                                                              |

`ArchitectureViolation.Type` enumeration: `layer`, `cycle`, `coupling`, `responsibility`, `cohesion`.

`ArchitectureViolation.Severity` enumeration: `info`, `warning`, `error`, `critical`.

## `suggestions` array

Array of `Suggestion` objects. Uses snake_case field names.

| Field         | Type    | Required | Description                                       |
| ------------- | ------- | -------- | ------------------------------------------------- |
| `category`    | string  | yes      | See enumeration below.                            |
| `severity`    | string  | yes      | One of: `critical`, `warning`, `info`.            |
| `effort`      | string  | yes      | One of: `easy`, `moderate`, `hard`.               |
| `title`       | string  | yes      | Short human-readable title.                       |
| `description` | string  | yes      | Full description.                                 |
| `steps`       | array of string | no | Actionable steps. Omitted when empty.        |
| `file_path`   | string  | no       | Source file reference.                            |
| `function`    | string  | no       | Function name reference.                          |
| `class_name`  | string  | no       | Class name reference.                             |
| `start_line`  | integer | no       | 1-based line reference. Omitted when `0`.         |
| `metric_value`| string  | no       | Observed metric value as string.                  |
| `threshold`   | string  | no       | Threshold value as string.                        |

`category` enumeration: `complexity`, `dead_code`, `clone`, `coupling`, `cohesion`, `dependency`, `architecture`.

Suggestions are sorted by priority (severity × effort). See `domain/suggestion.go` for the exact ordering function.

## CSV schemas

CSV outputs are written with RFC 4180 quoting via the Go `encoding/csv` package.

### `pyscn analyze --csv`

Summary, module-quality, optional community, and directory-complexity metrics. Two columns. Literal UTF-8 strings, no type annotations.

| Column   | Type   | Description              |
| -------- | ------ | ------------------------ |
| `Metric` | string | Metric name.             |
| `Value`  | string | Metric value as string.  |

Rows (in this fixed order):

```csv
Metric,Value
Health Score,<integer>
Grade,<A|B|C|D|F|N/A>
Total Files,<integer>
Analyzed Files,<integer>
Skipped Files,<integer>
Total Functions,<integer>
Class Scopes,<integer>
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
Module Quality Count,<integer>
Diagnostic,<file_path> [<code>]: <message>
Analysis Failure,<analysis> <file_path> [<code>]: <message>
Module 1 Name,<string>
Module 1 File Path,<string>
Module 1 Lines of Code,<integer>
Module 1 Function Count,<integer>
Module 1 Analyzed Function Count,<integer>
Module 1 Average Complexity,<float with 2 decimals>
Module 1 Average Cognitive Complexity,<float with 2 decimals>
Module 1 Max Complexity,<integer>
Module 1 High Risk Function Count,<integer>
Module 1 Exception Handler Count,<integer>
Module 1 Dead Code Findings,<integer>
Module 1 Dead Code Blocks,<integer>
Directory Complexity Count,<integer>
Directory 1 Path,<string>
Directory 1 Function Count,<integer>
Directory 1 Average Complexity,<float with 2 decimals>
Directory 1 Max Complexity,<integer>
Directory 1 High Risk Function Count,<integer>
Directory 1 Average Nesting Depth,<float with 2 decimals>
Directory 1 Max Nesting Depth,<integer>
```

The `Diagnostic` and `Analysis Failure` rows repeat once per corresponding entry and are omitted when empty. They appear before the numbered module rows. The numbered module and directory row groups repeat once per corresponding entry in the same order. Directory rows are appended after all summary, diagnostic, failure, module, and optional community rows, and are omitted when complexity analysis is disabled. CSV remains a summary format; use `--json` or `--yaml` for per-scope and per-finding detail.

The standalone complexity formatter emits one row per reported module, function, or class scope. Existing columns remain in place and `Scope Kind` is appended:

```csv
Function,Complexity,Cognitive Complexity,Risk,Nodes,Edges,Nesting Depth,If Statements,Loop Statements,Exception Handlers,SLOC,Scope Kind
```

## `community_analysis` object { #community-analysis-object }

Mirrors `domain.CommunityAnalysisResult`. Emitted as a top-level field in unified `pyscn analyze` JSON/YAML when community detection runs. When `pyscn analyze --json --select communities` is used, the report file contains only this object (standalone JSON).

| Field               | Type    | Description                                                       |
| ------------------- | ------- | ----------------------------------------------------------------- |
| `algorithm`         | string  | Community detection algorithm (currently `leiden`).               |
| `scope`             | string  | Graph scope (currently `module`).                                 |
| `total_communities` | integer | Number of detected communities.                                   |
| `modularity`        | number  | Partition modularity score.                                         |
| `communities`       | array   | Per-community metrics. See [`community`](#community-object).      |
| `bridge_modules`    | array   | Cross-community coupling modules. See [`bridge_module`](#bridge-module-object). |
| `package_alignment_score` | number \| absent | Share of packages whose modules all reside in one community (0–1). Omitted when package metadata is unavailable. |
| `split_packages`    | array \| absent | Packages whose modules appear in two or more communities (sorted). |
| `mixed_communities` | array \| absent | Community ids containing modules from two or more packages (sorted). |
| `layer_alignment_score` | number \| absent | Share of configured layers whose modules all reside in one community (0–1). Omitted when architecture layers are not configured. |
| `cross_layer_communities` | array \| absent | Community ids containing modules from two or more configured layers (sorted). |
| `layer_bridge_modules` | array \| absent | Bridge modules coupling communities mapped to different layers (sorted). |
| `community_risk_score` | integer \| absent | System-level community risk score (0–100, higher = worse). Absent when fewer than two communities were detected. |
| `community_context_map` | object \| absent | Compact, agent-optimized map of which modules to inspect together. See [`community_context_map`](#community-context-map-object). Absent when no communities were detected. |
| `warnings`          | array \| absent | Non-fatal analysis warnings.                              |
| `errors`            | array \| absent | Fatal analysis errors.                                    |
| `failures`          | array \| absent | Typed analyzer execution failures. See [`AnalysisFailure`](#analysisfailure-object). |
| `generated_at`      | string (RFC 3339) | Community analysis completion time.                 |
| `version`           | string  | pyscn semantic version.                                           |
| `config`            | object \| absent | Effective community-detection settings.                    |

### `community` object { #community-object }

| Field                              | Type    | Description                                              |
| ---------------------------------- | ------- | -------------------------------------------------------- |
| `id`                               | string  | Stable community identifier (`community_1`, `community_2`, …). |
| `modules`                          | array   | Module names in this community (sorted for stable diffs). |
| `packages`                         | array   | Package names represented in this community.             |
| `internal_edges`                   | integer | Dependency edges within the community.                   |
| `external_edges`                   | integer | Dependency edges crossing community boundaries.        |
| `external_dependency_ratio`        | number  | `external_edges / (internal_edges + external_edges)`.    |
| `incoming_cross_community_edges`   | integer | Incoming edges from other communities.                   |
| `outgoing_cross_community_edges`   | integer | Outgoing edges to other communities.                   |
| `size`                             | integer | Number of modules in the community.                      |
| `dominant_package`                 | string \| absent | Package with the most modules in this community.   |
| `package_count`                    | integer \| absent | Distinct packages represented in this community.  |
| `package_alignment`                | number \| absent | Cohesion within the community: share of internal edges whose endpoints share a package, or dominant-package module ratio when no qualifying internal edges exist. |
| `dominant_layer`                   | string \| absent | Configured layer with the most modules in this community. |
| `layer_count`                      | integer \| absent | Distinct configured layers represented in this community. |
| `layers`                           | array \| absent | Configured layer names present in this community (sorted). |
| `layer_alignment`                  | number \| absent | Cohesion within the community: share of internal edges whose endpoints share a layer, or dominant-layer module ratio when no qualifying internal edges exist. |
| `risk_level`                       | string \| absent | Per-community risk classification (`low`, `medium`, `high`). |

### `bridge_module` object { #bridge-module-object }

| Field                  | Type    | Description                                           |
| ---------------------- | ------- | ----------------------------------------------------- |
| `module`               | string  | Module name acting as a bridge.                       |
| `community`            | string  | Home community id for the module.                     |
| `cross_community_edges`| integer | Edges that connect to other communities.              |
| `target_communities`   | array   | Destination community ids (sorted for stable diffs).  |

### `community_context_map` object { #community-context-map-object }

A compact, deterministic view of the communities optimized for AI coding/review agents: which modules to inspect together, and which modules bridge otherwise-separate clusters. Derived entirely from the community analysis (no LLM-generated content).

| Field           | Type   | Description                                              |
| --------------- | ------ | -------------------------------------------------------- |
| `version`       | integer | Schema version of the context map (currently `1`).     |
| `bundles`       | array  | Module clusters to review together. See [`context_bundle`](#context-bundle-object). |
| `bridge_modules`| array  | Modules coupling two or more communities. See [`context_bridge_module`](#context-bridge-module-object). |

#### `context_bundle` object { #context-bundle-object }

| Field                    | Type            | Description                                              |
| ------------------------ | --------------- | -------------------------------------------------------- |
| `community_id`           | string          | Community identifier this bundle maps to.               |
| `modules`                | array           | Member module names (sorted). Capped at 10 entries; the remainder is collapsed into a `... +N more` marker. |
| `module_count`           | integer         | True number of modules in the community (before capping). |
| `packages`               | array           | Package names represented in the community (sorted).    |
| `risk_level`             | string          | Per-community risk classification (`low`, `medium`, `high`). |
| `bridge_modules`         | array           | Member modules that bridge to other communities (sorted). |
| `suggested_review_scope` | string \| absent | Filesystem-style path prefix derived from the longest common module prefix (e.g. `app/orders/`). Omitted when members share no common package. |
| `summary`                | string          | One-line, fact-based description of the cluster.        |

#### `context_bridge_module` object { #context-bridge-module-object }

| Field      | Type   | Description                                              |
| ---------- | ------ | -------------------------------------------------------- |
| `module`   | string | Module name acting as a bridge.                         |
| `connects` | array  | Community ids this module couples (sorted, includes its home community). |
| `reason`   | string | Human-readable reason (e.g. `3 cross-community import edges`). |

### Determinism

Community detection is deterministic for a fixed codebase snapshot and configuration: repeated runs yield identical `communities`, `bridge_modules`, and `modularity`. Module and community ordering in JSON is stable (sorted ids and module names). Numeric fields are rounded to four decimal places for diff-friendly output. Results may change across pyscn versions or when `min_community_size`, `resolution`, or `include_lazy_edges` change. See [Module Community Detection](../guides/module-community-detection.md#determinism) for details.

## `mock_data` object

Mirrors `domain.MockDataResponse`.

| Field | Type | Description |
| --- | --- | --- |
| `files` | array | Per-file mock-data findings. |
| `summary` | object | File, finding, severity, and type totals. |
| `warnings` | array | Non-fatal detector warnings. |
| `errors` | array | Fatal detector errors retained for output compatibility. |
| `diagnostics` | array \| absent | Typed file read and parse failures. See [`AnalysisDiagnostic`](#analysisdiagnostic-object). |
| `failures` | array \| absent | Typed detector execution failures. See [`AnalysisFailure`](#analysisfailure-object). |
| `generated_at` | string | Analysis completion time. |
| `version` | string | pyscn semantic version. |
| `config` | object \| null | Effective mock-data settings. |

## Timestamps and versioning

| Field          | Format                    | Notes                                                   |
| -------------- | ------------------------- | ------------------------------------------------------- |
| `generated_at` | RFC 3339 (ISO 8601)       | `time.Time` serialization; may include sub-second precision and timezone offset. |
| `duration_ms`  | integer (milliseconds)    | Wall-clock analysis time.                               |
| `version`      | string (semantic version) | pyscn release version, e.g. `"0.14.0"`.                 |

## Invoking each format

`pyscn analyze` takes one of `--json`, `--yaml`, `--csv`, `--text`, or `--html` (default). There is no `--format` flag, and there are no standalone `complexity` / `deadcode` / `clone` / `deps` subcommands. Run a single analyzer via `--select`.

```bash
pyscn analyze --json src/
pyscn analyze --yaml src/
pyscn analyze --csv  src/
pyscn analyze --text src/
pyscn analyze --html src/    # default
pyscn analyze --json --select complexity src/
pyscn analyze --csv  --select deadcode   src/
pyscn analyze --yaml --select clones        src/
pyscn analyze --json --select communities   src/
```

`--select communities` with `--json` writes standalone community JSON (not the unified `AnalyzeResponse` wrapper). YAML standalone output is not supported yet; use unified `pyscn analyze --yaml` or `--json --select communities`.

Output files land in `.pyscn/reports/`; see [Output Formats](index.md) for path and filename details.

## Related

- [HTML Report](html-report.md) — HTML output specification.
- [Health Score](health-score.md) — derivation of `summary.health_score` and per-category scores.
