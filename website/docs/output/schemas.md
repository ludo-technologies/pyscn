# Output Schemas

This specification defines the exact shape of JSON, YAML, and CSV output produced by pyscn. All field names, types, and semantics documented here are stable across patch releases within the same major version.

## Stability contract

| Guarantee          | Scope                                                                             |
| ------------------ | --------------------------------------------------------------------------------- |
| Stable             | field names, field types, field semantics, enum values                            |
| May change         | field ordering within an object, ordering of array elements, inclusion of new fields |
| Breaking           | removal or rename of fields, change of field type, removal of enum values         |

Breaking changes are restricted to major version bumps. Consumers MUST ignore unknown fields.

<!-- Field naming note: in `pyscn analyze` JSON/YAML, nested analyzer objects (`complexity`, `cbo`, `lcom`, `system`) use Go-style PascalCase field names because their response structs do not carry JSON tags. Top-level keys, `dead_code`, `clone`, `suggestions`, and `summary` use snake_case. -->

## Top-level structure (`pyscn analyze`)

JSON and YAML outputs serialize the `AnalyzeResponse` Go struct defined in `domain/analyze.go`. The top-level keys are:

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

| Field         | Type              | Description                                            | Stability |
| ------------- | ----------------- | ------------------------------------------------------ | --------- |
| `complexity`  | object \| absent  | Present when complexity analysis ran.                  | stable    |
| `dead_code`   | object \| absent  | Present when dead code analysis ran.                   | stable    |
| `clone`       | object \| absent  | Present when clone detection ran.                      | stable    |
| `cbo`         | object \| absent  | Present when CBO analysis ran.                         | stable    |
| `lcom`        | object \| absent  | Present when LCOM analysis ran.                        | stable    |
| `system`      | object \| absent  | Present when dependency or architecture analysis ran.  | stable    |
| `mock_data`   | object \| absent  | Present when mock data detection ran.                  | stable    |
| `suggestions` | array \| absent   | Derived suggestions. Omitted when empty.               | stable    |
| `summary`     | object            | Always present. See [`summary`](#summary-object).      | stable    |
| `generated_at`| string (RFC 3339) | Analysis completion time.                              | stable    |
| `duration_ms` | integer           | Total analysis duration in milliseconds.               | stable    |
| `version`     | string            | pyscn semantic version.                                | stable    |

## `summary` object

Mirrors `domain.AnalyzeSummary`. All numeric counters default to `0` when the corresponding analyzer is disabled. All fields are always present.

### File statistics

| Field            | Type    | Description                                      |
| ---------------- | ------- | ------------------------------------------------ |
| `total_files`    | integer | Number of Python files discovered.               |
| `analyzed_files` | integer | Number of files successfully analyzed.           |
| `skipped_files`  | integer | Files skipped due to parse errors or filters.    |

### Analyzer status flags

| Field                | Type    | Description                                                |
| -------------------- | ------- | ---------------------------------------------------------- |
| `complexity_enabled` | boolean | `true` if complexity analysis produced results.            |
| `dead_code_enabled`  | boolean | `true` if dead code analysis produced results.             |
| `clone_enabled`      | boolean | `true` if clone detection produced results.                |
| `cbo_enabled`        | boolean | `true` if CBO analysis produced results.                   |
| `lcom_enabled`       | boolean | `true` if LCOM analysis produced results.                  |
| `deps_enabled`       | boolean | `true` if dependency analysis produced results.            |
| `arch_enabled`       | boolean | `true` if architecture validation produced results.        |
| `mock_data_enabled`  | boolean | `true` if mock data detection produced results.            |

### Complexity metrics

| Field                   | Type    | Description                                      |
| ----------------------- | ------- | ------------------------------------------------ |
| `total_functions`       | integer | Total functions analyzed.                        |
| `average_complexity`    | number  | Mean cyclomatic complexity. `0` when no functions. |
| `high_complexity_count` | integer | Functions with complexity > 10 (medium threshold). |

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
| `deps_max_depth`               | integer | Longest dependency chain length.                               |
| `deps_main_sequence_deviation` | number  | Average distance from Martin's main sequence, `0`–`1`.         |

### Architecture metrics

| Field             | Type   | Description                                                       |
| ----------------- | ------ | ----------------------------------------------------------------- |
| `arch_compliance` | number | Architecture compliance ratio, `0`–`1`. `1.0` = fully compliant.  |

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

## `complexity` object

Mirrors `domain.ComplexityResponse`. Nested field names are Go PascalCase.

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

### `Functions[]` element (`FunctionComplexity`)

| Field         | Type    | Description                                                  |
| ------------- | ------- | ------------------------------------------------------------ |
| `Name`        | string  | Function name. `__main__` for module-level code.             |
| `FilePath`    | string  | Path to source file.                                         |
| `StartLine`   | integer | 1-based start line.                                          |
| `StartColumn` | integer | 0-based start column.                                        |
| `EndLine`     | integer | 1-based end line.                                            |
| `Metrics`     | object  | See [`ComplexityMetrics`](#complexitymetrics-object).        |
| `RiskLevel`   | string  | One of: `low`, `medium`, `high`.                             |

### `ComplexityMetrics` object

| Field                 | Type    | Description                                        |
| --------------------- | ------- | -------------------------------------------------- |
| `Complexity`          | integer | McCabe cyclomatic complexity.                      |
| `CognitiveComplexity` | integer | Cognitive complexity (SonarQube-style).            |
| `Nodes`               | integer | CFG node count.                                    |
| `Edges`               | integer | CFG edge count.                                    |
| `NestingDepth`        | integer | Maximum nesting depth.                             |
| `IfStatements`        | integer | Count of `if` statements.                          |
| `LoopStatements`      | integer | Count of `for`/`while` loops.                      |
| `ExceptionHandlers`   | integer | Count of `except` clauses.                         |
| `SwitchCases`         | integer | Count of `match` cases (Python 3.10+).             |

### `Summary` object (`ComplexitySummary`)

| Field                    | Type    | Description                                                            |
| ------------------------ | ------- | ---------------------------------------------------------------------- |
| `TotalFunctions`         | integer | Total functions analyzed.                                              |
| `AverageComplexity`      | number  | Arithmetic mean of `Complexity` across all functions.                  |
| `MaxComplexity`          | integer | Highest observed complexity.                                           |
| `MinComplexity`          | integer | Lowest observed complexity.                                            |
| `FilesAnalyzed`          | integer | Files contributing at least one function.                              |
| `LowRiskFunctions`       | integer | Functions with `RiskLevel = low`.                                      |
| `MediumRiskFunctions`    | integer | Functions with `RiskLevel = medium`.                                   |
| `HighRiskFunctions`      | integer | Functions with `RiskLevel = high`.                                     |
| `ComplexityDistribution` | object  | Histogram keyed by complexity bucket (string) to count (integer), or `null`. |

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
  "generated_at": "",
  "version": "",
  "config": null
}
```

### `files[]` element (`FileDeadCode`)

| Field               | Type    | Description                                    |
| ------------------- | ------- | ---------------------------------------------- |
| `file_path`         | string  | Path to source file.                           |
| `functions`         | array   | Per-function results (see below).              |
| `total_findings`    | integer | Sum of findings across functions in this file. |
| `total_functions`   | integer | Functions analyzed in this file.               |
| `affected_functions`| integer | Functions with at least one finding.           |
| `dead_code_ratio`   | number  | Dead blocks / total blocks, `0`–`1`.           |

### `files[].functions[]` element (`FunctionDeadCode`)

| Field             | Type    | Description                                  |
| ----------------- | ------- | -------------------------------------------- |
| `name`            | string  | Function name.                               |
| `file_path`       | string  | Path to source file.                         |
| `findings`        | array   | Findings in this function (see below).       |
| `total_blocks`    | integer | Total CFG blocks in the function.            |
| `dead_blocks`     | integer | Unreachable CFG blocks.                      |
| `reachable_ratio` | number  | `(total_blocks - dead_blocks) / total_blocks`, `0`–`1`. |
| `critical_count`  | integer | Findings of severity `critical`.             |
| `warning_count`   | integer | Findings of severity `warning`.              |
| `info_count`      | integer | Findings of severity `info`.                 |

### `files[].functions[].findings[]` element (`DeadCodeFinding`)

| Field           | Type    | Description                                                   |
| --------------- | ------- | ------------------------------------------------------------- |
| `location`      | object  | See [`DeadCodeLocation`](#deadcodelocation-object).           |
| `function_name` | string  | Enclosing function name.                                      |
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

### `DeadCodeLocation` object

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
| `total_findings`           | integer | Total findings across all files.                 |
| `files_with_dead_code`     | integer | Files with at least one finding.                 |
| `functions_with_dead_code` | integer | Functions with at least one finding.             |
| `critical_findings`        | integer | Findings with severity `critical`.               |
| `warning_findings`         | integer | Findings with severity `warning`.                |
| `info_findings`            | integer | Findings with severity `info`.                   |
| `findings_by_reason`       | object \| null | Histogram keyed by `reason` value.         |
| `total_blocks`             | integer | CFG blocks across all functions.                 |
| `dead_blocks`              | integer | Unreachable CFG blocks across all functions.     |
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
  "error": ""
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

### `CloneLocation` object

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

## `cbo` object

Mirrors `domain.CBOResponse`. Nested field names are Go PascalCase.

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

### `Classes[]` element (`ClassCoupling`)

| Field         | Type    | Description                                 |
| ------------- | ------- | ------------------------------------------- |
| `Name`        | string  | Class name.                                 |
| `FilePath`    | string  | Path to source file.                        |
| `StartLine`   | integer | 1-based start line.                         |
| `EndLine`     | integer | 1-based end line.                           |
| `Metrics`     | object  | See [`CBOMetrics`](#cbometrics-object).     |
| `RiskLevel`   | string  | One of: `low`, `medium`, `high`.            |
| `IsAbstract`  | boolean | `true` if the class is abstract.            |
| `BaseClasses` | array of string \| null | Direct base classes.        |

### `CBOMetrics` object

| Field                         | Type    | Description                                               |
| ----------------------------- | ------- | --------------------------------------------------------- |
| `CouplingCount`               | integer | CBO value: distinct classes this class depends on.        |
| `InheritanceDependencies`     | integer | Dependencies from base classes.                           |
| `TypeHintDependencies`        | integer | Dependencies from type annotations.                       |
| `InstantiationDependencies`   | integer | Dependencies from object instantiation.                   |
| `AttributeAccessDependencies` | integer | Dependencies from method calls and attribute access.      |
| `ImportDependencies`          | integer | Dependencies from explicit imports.                       |
| `DependentClasses`            | array of string \| null | Names of coupled classes.                 |

### `Summary` object (`CBOSummary`)

| Field                      | Type    | Description                                       |
| -------------------------- | ------- | ------------------------------------------------- |
| `TotalClasses`             | integer | Total classes analyzed.                           |
| `AverageCBO`               | number  | Mean CBO.                                         |
| `MaxCBO`                   | integer | Maximum observed CBO.                             |
| `MinCBO`                   | integer | Minimum observed CBO.                             |
| `ClassesAnalyzed`          | integer | Classes with valid metrics.                       |
| `FilesAnalyzed`            | integer | Files contributing at least one class.            |
| `LowRiskClasses`           | integer | Classes with CBO ≤ low threshold (default `3`).   |
| `MediumRiskClasses`        | integer | Classes with low < CBO ≤ medium threshold.        |
| `HighRiskClasses`          | integer | Classes with CBO > medium threshold (default `7`).|
| `CBODistribution`          | object \| null | Histogram keyed by bucket label to count.  |
| `MostCoupledClasses`       | array \| null | Top 10 classes by CBO (`ClassCoupling`).    |
| `MostDependedUponClasses`  | array of string \| null | Classes with highest in-degree.   |

## `lcom` object

Mirrors `domain.LCOMResponse`. Nested field names are Go PascalCase.

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

### `Classes[]` element (`ClassCohesion`)

| Field       | Type    | Description                                      |
| ----------- | ------- | ------------------------------------------------ |
| `Name`      | string  | Class name.                                      |
| `FilePath`  | string  | Path to source file.                             |
| `StartLine` | integer | 1-based start line.                              |
| `EndLine`   | integer | 1-based end line.                                |
| `Metrics`   | object  | See [`LCOMMetrics`](#lcommetrics-object).        |
| `RiskLevel` | string  | One of: `low`, `medium`, `high`.                 |

### `LCOMMetrics` object

| Field               | Type    | Description                                                 |
| ------------------- | ------- | ----------------------------------------------------------- |
| `LCOM4`             | integer | Connected components in the method-variable graph.          |
| `TotalMethods`      | integer | All methods in the class.                                   |
| `ExcludedMethods`   | integer | Methods excluded from LCOM4 (`@classmethod`, `@staticmethod`). |
| `InstanceVariables` | integer | Distinct `self.x` variables accessed.                       |
| `MethodGroups`      | array of array of string \| null | Method names grouped by connected component. |

### `Summary` object (`LCOMSummary`)

| Field                  | Type    | Description                                     |
| ---------------------- | ------- | ----------------------------------------------- |
| `TotalClasses`         | integer | Classes analyzed.                               |
| `AverageLCOM`          | number  | Mean LCOM4.                                     |
| `MaxLCOM`              | integer | Maximum observed LCOM4.                         |
| `MinLCOM`              | integer | Minimum observed LCOM4.                         |
| `ClassesAnalyzed`      | integer | Classes with valid metrics.                     |
| `FilesAnalyzed`        | integer | Files contributing at least one class.          |
| `LowRiskClasses`       | integer | Classes with LCOM4 ≤ low threshold (default `2`). |
| `MediumRiskClasses`    | integer | Classes with low < LCOM4 ≤ medium threshold.    |
| `HighRiskClasses`      | integer | Classes with LCOM4 > medium threshold (default `5`). |
| `LCOMDistribution`     | object \| null | Histogram keyed by bucket label to count. |
| `LeastCohesiveClasses` | array \| null | Top 10 classes by LCOM4 (`ClassCohesion`). |

## `system` object

Mirrors `domain.SystemAnalysisResponse`. Nested field names are Go PascalCase.

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

### `Summary` object (`SystemAnalysisSummary`)

| Field                      | Type    | Description                                     |
| -------------------------- | ------- | ----------------------------------------------- |
| `TotalModules`             | integer | Total modules analyzed.                         |
| `TotalPackages`            | integer | Total packages.                                 |
| `TotalDependencies`        | integer | Total dependency edges.                         |
| `ProjectRoot`              | string  | Project root directory.                         |
| `OverallQualityScore`      | number  | Composite quality score, `0`–`100`.             |
| `MaintainabilityScore`     | number  | Average maintainability index.                  |
| `ArchitectureScore`        | number  | Architecture compliance score.                  |
| `ModularityScore`          | number  | System modularity score.                        |
| `TechnicalDebtHours`       | number  | Total estimated technical debt in hours.        |
| `AverageCoupling`          | number  | Average module coupling.                        |
| `AverageInstability`       | number  | Average instability (I).                        |
| `CyclicDependencies`       | integer | Modules participating in cycles.                |
| `ArchitectureViolations`   | integer | Count of architecture rule violations.          |
| `HighRiskModules`          | integer | Modules flagged high risk.                      |
| `CriticalIssues`           | integer | Critical issue count.                           |
| `RefactoringCandidates`    | integer | Modules needing refactoring.                    |
| `ArchitectureImprovements` | integer | Suggested architecture improvements.            |

### `DependencyAnalysis` object

| Field                  | Type    | Description                                                          |
| ---------------------- | ------- | -------------------------------------------------------------------- |
| `TotalModules`         | integer | Total modules in the dependency graph.                               |
| `TotalDependencies`    | integer | Total edges.                                                         |
| `RootModules`          | array of string | Modules with no outgoing dependencies.                       |
| `LeafModules`          | array of string | Modules with no incoming dependencies.                       |
| `ModuleMetrics`        | object  | Map from module name to `ModuleDependencyMetrics`.                   |
| `DependencyMatrix`     | object  | Map from module to map of module to boolean.                         |
| `CircularDependencies` | object  | Cycle detection results; contains `Cycles` (array) and `TotalCycles` (integer). |
| `CouplingAnalysis`     | object  | Per-module coupling metrics: `Ca`, `Ce`, `Instability`, `Abstractness`, `Distance`. |
| `LongestChains`        | array   | Array of `DependencyPath` objects.                                   |
| `MaxDepth`             | integer | Maximum dependency depth.                                            |

### `ModuleDependencyMetrics` object

| Field                    | Type    | Description                                              |
| ------------------------ | ------- | -------------------------------------------------------- |
| `ModuleName`             | string  | Fully qualified module name.                             |
| `Package`                | string  | Parent package.                                          |
| `FilePath`               | string  | Path to source file.                                     |
| `IsPackage`              | boolean | `true` if this is a package (has `__init__.py`).         |
| `LinesOfCode`            | integer | Total lines of code.                                     |
| `FunctionCount`          | integer | Number of functions.                                     |
| `ClassCount`             | integer | Number of classes.                                       |
| `PublicInterface`        | array of string | Names in `__all__` or top-level public names.    |
| `AfferentCoupling`       | integer | Ca — modules depending on this one.                      |
| `EfferentCoupling`       | integer | Ce — modules this one depends on.                        |
| `Instability`            | number  | `I = Ce / (Ca + Ce)`, `0`–`1`.                           |
| `Abstractness`           | number  | A — abstract elements / total elements, `0`–`1`.         |
| `Distance`               | number  | `D = |A + I - 1|`, `0`–`1`. Distance from main sequence. |
| `Maintainability`        | number  | Maintainability index, `0`–`100`.                        |
| `TechnicalDebt`          | number  | Estimated technical debt in hours.                       |
| `RiskLevel`              | string  | One of: `low`, `medium`, `high`.                         |
| `DirectDependencies`     | array of string | Direct dependencies.                             |
| `TransitiveDependencies` | array of string | All transitive dependencies.                     |
| `Dependents`             | array of string | Modules depending on this one.                   |

### `CircularDependencyAnalysis` object

| Field                      | Type    | Description                                           |
| -------------------------- | ------- | ----------------------------------------------------- |
| `HasCircularDependencies`  | boolean | `true` if any cycles exist.                           |
| `TotalCycles`              | integer | Number of cycles.                                     |
| `TotalModulesInCycles`     | integer | Modules involved in cycles.                           |
| `CircularDependencies`     | array   | Array of `CircularDependency` objects.                |
| `CycleBreakingSuggestions` | array of string | Suggestions for breaking cycles.              |
| `CoreInfrastructure`       | array of string | Modules appearing in multiple cycles.         |

`CircularDependency.Severity` enumeration: `low`, `medium`, `high`, `critical`.

### `CouplingAnalysis` object

| Field                   | Type    | Description                                        |
| ----------------------- | ------- | -------------------------------------------------- |
| `AverageCoupling`       | number  | Average coupling across modules.                   |
| `CouplingDistribution`  | object  | Map from coupling value (integer key) to count.    |
| `HighlyCoupledModules`  | array of string | Modules with high coupling.                |
| `LooselyCoupledModules` | array of string | Modules with low coupling.                 |
| `AverageInstability`    | number  | Average instability.                               |
| `StableModules`         | array of string | Low-instability modules.                   |
| `InstableModules`       | array of string | High-instability modules.                  |
| `MainSequenceDeviation` | number  | Average distance from main sequence, `0`–`1`.      |
| `ZoneOfPain`            | array of string | Stable + concrete modules.                 |
| `ZoneOfUselessness`     | array of string | Unstable + abstract modules.               |
| `MainSequence`          | array of string | Well-positioned modules.                   |

### `ArchitectureAnalysis` object

| Field                 | Type    | Description                                                 |
| --------------------- | ------- | ----------------------------------------------------------- |
| `ComplianceScore`     | number  | Compliance score, `0`–`1`. `1.0` = fully compliant.         |
| `TotalViolations`     | integer | Total violations found.                                     |
| `TotalRules`          | integer | Total rules evaluated.                                      |
| `LayerAnalysis`       | object \| null | Layer analysis results.                              |
| `CohesionAnalysis`    | object \| null | Package cohesion analysis.                           |
| `ResponsibilityAnalysis` | object \| null | SRP violation analysis.                           |
| `Violations`          | array   | Array of `ArchitectureViolation` objects.                   |
| `SeverityBreakdown`   | object  | Map from severity to count.                                 |
| `Recommendations`     | array   | Array of `ArchitectureRecommendation` objects.              |
| `RefactoringTargets`  | array of string | Modules needing refactoring.                        |

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

Summary only. Two columns. Literal UTF-8 strings, no type annotations.

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

pyscn does not currently expose per-analyzer CSV schemas through the CLI — `--csv` produces only the summary above. For per-finding detail, use `--json` or `--yaml`.

## Timestamps and versioning

| Field          | Format                    | Notes                                                   |
| -------------- | ------------------------- | ------------------------------------------------------- |
| `generated_at` | RFC 3339 (ISO 8601)       | `time.Time` serialization; may include sub-second precision and timezone offset. |
| `duration_ms`  | integer (milliseconds)    | Wall-clock analysis time.                               |
| `version`      | string (semantic version) | pyscn release version, e.g. `"0.14.0"`.                 |

## Invoking each format

`pyscn analyze` takes one of `--json`, `--yaml`, `--csv`, `--html` (default). There is no `--format` flag, and there are no standalone `complexity` / `deadcode` / `clone` / `deps` subcommands. Run a single analyzer via `--select`.

```bash
pyscn analyze --json src/
pyscn analyze --yaml src/
pyscn analyze --csv  src/
pyscn analyze --html src/    # default
pyscn analyze --json --select complexity src/
pyscn analyze --csv  --select deadcode   src/
pyscn analyze --yaml --select clones     src/
```

Output files land in `.pyscn/reports/`; see [Output Formats](index.md) for path and filename details.

## Related

- [HTML Report](html-report.md) — HTML output specification.
- [Health Score](health-score.md) — derivation of `summary.health_score` and per-category scores.
