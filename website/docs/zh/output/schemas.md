# 输出模式

本规范定义了 pyscn 生成的 JSON、YAML 和 CSV 输出的确切结构。此处记录的所有字段名称、类型和语义在同一主版本内的补丁版本之间保持稳定。

## 稳定性约定

| 保证               | 范围                                                                              |
| ------------------ | --------------------------------------------------------------------------------- |
| 稳定               | 字段名称、字段类型、字段语义、枚举值                                              |
| 可能变更           | 对象内的字段顺序、数组元素的顺序、新字段的添加                                    |
| 破坏性变更         | 字段的删除或重命名、字段类型的更改、枚举值的删除                                  |

破坏性变更仅限于主版本升级。使用方必须忽略未知字段。

<!-- Field naming note: in `pyscn analyze` JSON/YAML, nested analyzer objects (`complexity`, `cbo`, `lcom`, `system`) use Go-style PascalCase field names because their response structs do not carry JSON tags. Top-level keys, `dead_code`, `clone`, `suggestions`, and `summary` use snake_case. -->

## 顶层结构（`pyscn analyze`）

JSON 和 YAML 输出序列化 `domain/analyze.go` 中定义的 `AnalyzeResponse` Go 结构体。顶层键为：

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

| 字段          | 类型              | 说明                                                   | 稳定性    |
| ------------- | ----------------- | ------------------------------------------------------ | --------- |
| `complexity`  | object \| absent  | 复杂度分析运行时存在。                                 | 稳定      |
| `dead_code`   | object \| absent  | 死代码分析运行时存在。                                 | 稳定      |
| `clone`       | object \| absent  | 克隆检测运行时存在。                                   | 稳定      |
| `cbo`         | object \| absent  | CBO 分析运行时存在。                                   | 稳定      |
| `lcom`        | object \| absent  | LCOM 分析运行时存在。                                  | 稳定      |
| `system`      | object \| absent  | 依赖或架构分析运行时存在。                             | 稳定      |
| `mock_data`   | object \| absent  | 模拟数据检测运行时存在。                               | 稳定      |
| `suggestions` | array \| absent   | 衍生建议。为空时省略。                                 | 稳定      |
| `summary`     | object            | 始终存在。见 [`summary`](#summary-object)。            | 稳定      |
| `generated_at`| string (RFC 3339) | 分析完成时间。                                         | 稳定      |
| `duration_ms` | integer           | 总分析耗时（毫秒）。                                   | 稳定      |
| `version`     | string            | pyscn 语义版本号。                                     | 稳定      |

## `summary` 对象

对应 `domain.AnalyzeSummary`。当对应分析器未启用时，所有数值计数器默认为 `0`。所有字段始终存在。

### 文件统计

| 字段             | 类型    | 说明                                     |
| ---------------- | ------- | ---------------------------------------- |
| `total_files`    | integer | 发现的 Python 文件数。                   |
| `analyzed_files` | integer | 成功分析的文件数。                       |
| `skipped_files`  | integer | 因解析错误或过滤器跳过的文件数。         |

### 分析器状态标志

| 字段                 | 类型    | 说明                                                   |
| -------------------- | ------- | ------------------------------------------------------ |
| `complexity_enabled` | boolean | 复杂度分析产生结果时为 `true`。                        |
| `dead_code_enabled`  | boolean | 死代码分析产生结果时为 `true`。                        |
| `clone_enabled`      | boolean | 克隆检测产生结果时为 `true`。                          |
| `cbo_enabled`        | boolean | CBO 分析产生结果时为 `true`。                          |
| `lcom_enabled`       | boolean | LCOM 分析产生结果时为 `true`。                         |
| `deps_enabled`       | boolean | 依赖分析产生结果时为 `true`。                          |
| `arch_enabled`       | boolean | 架构验证产生结果时为 `true`。                          |
| `mock_data_enabled`  | boolean | 模拟数据检测产生结果时为 `true`。                      |

### 复杂度指标

| 字段                    | 类型    | 说明                                     |
| ----------------------- | ------- | ---------------------------------------- |
| `total_functions`       | integer | 分析的函数总数。                         |
| `average_complexity`    | number  | 平均圈复杂度。无函数时为 `0`。           |
| `high_complexity_count` | integer | 复杂度 > 10（中等阈值）的函数数。       |

### 死代码指标

| 字段                 | 类型    | 说明                                 |
| -------------------- | ------- | ------------------------------------ |
| `dead_code_count`    | integer | 发现总数。                           |
| `critical_dead_code` | integer | 严重性为 `critical` 的发现数。       |
| `warning_dead_code`  | integer | 严重性为 `warning` 的发现数。        |
| `info_dead_code`     | integer | 严重性为 `info` 的发现数。           |

### 克隆指标

| 字段                          | 类型    | 说明                                              |
| ----------------------------- | ------- | ------------------------------------------------- |
| `total_clones`                | integer | 被识别为克隆的独立代码片段数。                    |
| `clone_pairs`                 | integer | 克隆对的数量。                                    |
| `clone_groups`                | integer | 克隆组的数量。                                    |
| `code_duplication_percentage` | number  | 估计的代码重复率，`0`–`100`。                     |

### CBO 指标

| 字段                      | 类型    | 说明                                              |
| ------------------------- | ------- | ------------------------------------------------- |
| `cbo_classes`             | integer | 分析的类总数。                                    |
| `high_coupling_classes`   | integer | CBO > 7 的类数。                                  |
| `medium_coupling_classes` | integer | 3 < CBO ≤ 7 的类数。                              |
| `average_coupling`        | number  | 平均 CBO 值。                                     |

### LCOM 指标

| 字段                  | 类型    | 说明                                 |
| --------------------- | ------- | ------------------------------------ |
| `lcom_classes`        | integer | 分析的类总数。                       |
| `high_lcom_classes`   | integer | LCOM4 > 5 的类数。                   |
| `medium_lcom_classes` | integer | 2 < LCOM4 ≤ 5 的类数。               |
| `average_lcom`        | number  | 平均 LCOM4 值。                      |

### 依赖指标

| 字段                           | 类型    | 说明                                                   |
| ------------------------------ | ------- | ------------------------------------------------------ |
| `deps_total_modules`           | integer | 分析的模块总数。                                       |
| `deps_modules_in_cycles`       | integer | 参与至少一个循环依赖的模块数。                         |
| `deps_max_depth`               | integer | 最长依赖链长度。                                       |
| `deps_main_sequence_deviation` | number  | 与 Martin 主序列的平均偏差，`0`–`1`。                  |

### 架构指标

| 字段              | 类型   | 说明                                                      |
| ----------------- | ------ | --------------------------------------------------------- |
| `arch_compliance` | number | 架构合规率，`0`–`1`。`1.0` = 完全合规。                   |

### 模拟数据指标

| 字段                    | 类型    | 说明                                        |
| ----------------------- | ------- | ------------------------------------------- |
| `mock_data_count`       | integer | 模拟数据发现总数。                          |
| `mock_data_error_count` | integer | error 严重性的发现数。                      |
| `mock_data_warning_count` | integer | warning 严重性的发现数。                  |
| `mock_data_info_count`  | integer | info 严重性的发现数。                       |

### 健康评分

| 字段                 | 类型    | 说明                                                       |
| -------------------- | ------- | ---------------------------------------------------------- |
| `health_score`       | integer | 综合评分，`0`–`100`。见[健康评分](health-score.md)。       |
| `grade`              | string  | 等级字母。取值：`A`、`B`、`C`、`D`、`F`、`N/A`。           |
| `complexity_score`   | integer | 分类评分，`0`–`100`。                                      |
| `dead_code_score`    | integer | 分类评分，`0`–`100`。                                      |
| `duplication_score`  | integer | 分类评分，`0`–`100`。                                      |
| `coupling_score`     | integer | 分类评分，`0`–`100`。                                      |
| `cohesion_score`     | integer | 分类评分，`0`–`100`。                                      |
| `dependency_score`   | integer | 分类评分，`0`–`100`。                                      |
| `architecture_score` | integer | 分类评分，`0`–`100`。                                      |

## `complexity` 对象

对应 `domain.ComplexityResponse`。嵌套字段名为 Go PascalCase。

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

### `Functions[]` 元素（`FunctionComplexity`）

| 字段          | 类型    | 说明                                                 |
| ------------- | ------- | ---------------------------------------------------- |
| `Name`        | string  | 函数名。模块级代码为 `__main__`。                    |
| `FilePath`    | string  | 源文件路径。                                         |
| `StartLine`   | integer | 从 1 开始的起始行。                                  |
| `StartColumn` | integer | 从 0 开始的起始列。                                  |
| `EndLine`     | integer | 从 1 开始的结束行。                                  |
| `Metrics`     | object  | 见 [`ComplexityMetrics`](#complexitymetrics-object)。|
| `RiskLevel`   | string  | 取值：`low`、`medium`、`high`。                      |

### `ComplexityMetrics` 对象

| 字段                  | 类型    | 说明                                       |
| --------------------- | ------- | ------------------------------------------ |
| `Complexity`          | integer | McCabe 圈复杂度。                          |
| `CognitiveComplexity` | integer | 认知复杂度（SonarQube 风格）。             |
| `Nodes`               | integer | CFG 节点数。                               |
| `Edges`               | integer | CFG 边数。                                 |
| `NestingDepth`        | integer | 最大嵌套深度。                             |
| `IfStatements`        | integer | `if` 语句数。                              |
| `LoopStatements`      | integer | `for`/`while` 循环数。                     |
| `ExceptionHandlers`   | integer | `except` 子句数。                          |
| `SwitchCases`         | integer | `match` 分支数（Python 3.10+）。           |

### `Summary` 对象（`ComplexitySummary`）

| 字段                     | 类型    | 说明                                                           |
| ------------------------ | ------- | -------------------------------------------------------------- |
| `TotalFunctions`         | integer | 分析的函数总数。                                               |
| `AverageComplexity`      | number  | 所有函数 `Complexity` 的算术平均值。                           |
| `MaxComplexity`          | integer | 观测到的最高复杂度。                                           |
| `MinComplexity`          | integer | 观测到的最低复杂度。                                           |
| `FilesAnalyzed`          | integer | 贡献了至少一个函数的文件数。                                   |
| `LowRiskFunctions`       | integer | `RiskLevel = low` 的函数数。                                   |
| `MediumRiskFunctions`    | integer | `RiskLevel = medium` 的函数数。                                |
| `HighRiskFunctions`      | integer | `RiskLevel = high` 的函数数。                                  |
| `ComplexityDistribution` | object  | 按复杂度区间（字符串）到数量（整数）的直方图，或 `null`。      |

### `raw_metrics[]` 元素（`RawMetrics`）

| 字段              | 类型    | 说明                                        |
| ----------------- | ------- | ------------------------------------------- |
| `file_path`       | string  | 源文件路径。                                |
| `sloc`            | integer | 源代码行数（非空行、非注释行）。            |
| `lloc`            | integer | 逻辑代码行数。                              |
| `comment_lines`   | integer | 包含注释的行数。                            |
| `docstring_lines` | integer | 文档字符串内的行数。                        |
| `blank_lines`     | integer | 空行或仅含空白的行数。                      |
| `total_lines`     | integer | 物理总行数。                                |
| `comment_ratio`   | number  | `(comment_lines + docstring_lines) / total_lines`，`0`–`1`。 |


## `dead_code` 对象

对应 `domain.DeadCodeResponse`。全部使用 snake_case 字段名。

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

### `files[]` 元素（`FileDeadCode`）

| 字段                | 类型    | 说明                                   |
| ------------------- | ------- | -------------------------------------- |
| `file_path`         | string  | 源文件路径。                           |
| `functions`         | array   | 每个函数的结果（见下文）。             |
| `total_findings`    | integer | 此文件中所有函数的发现总数。           |
| `total_functions`   | integer | 此文件中分析的函数数。                 |
| `affected_functions`| integer | 至少有一个发现的函数数。               |
| `dead_code_ratio`   | number  | 死代码块 / 总块数，`0`–`1`。           |

### `files[].functions[]` 元素（`FunctionDeadCode`）

| 字段              | 类型    | 说明                                 |
| ----------------- | ------- | ------------------------------------ |
| `name`            | string  | 函数名。                             |
| `file_path`       | string  | 源文件路径。                         |
| `findings`        | array   | 此函数中的发现（见下文）。           |
| `total_blocks`    | integer | 函数中的 CFG 块总数。                |
| `dead_blocks`     | integer | 不可达的 CFG 块数。                  |
| `reachable_ratio` | number  | `(total_blocks - dead_blocks) / total_blocks`，`0`–`1`。 |
| `critical_count`  | integer | 严重性为 `critical` 的发现数。       |
| `warning_count`   | integer | 严重性为 `warning` 的发现数。        |
| `info_count`      | integer | 严重性为 `info` 的发现数。           |

### `files[].functions[].findings[]` 元素（`DeadCodeFinding`）

| 字段            | 类型    | 说明                                                  |
| --------------- | ------- | ----------------------------------------------------- |
| `location`      | object  | 见 [`DeadCodeLocation`](#deadcodelocation-object)。   |
| `function_name` | string  | 所在函数名。                                          |
| `code`          | string  | 死代码片段。                                          |
| `reason`        | string  | 分类 — 见下方枚举。                                   |
| `severity`      | string  | 取值：`critical`、`warning`、`info`。                 |
| `description`   | string  | 人类可读的描述。                                      |
| `context`       | array of string \| absent | 周围的源代码行。使用 `--show-context` 时存在。 |
| `block_id`      | string \| absent | CFG 块标识符。                                 |

`reason` 枚举值：

| 值                    | 含义                                         |
| --------------------- | -------------------------------------------- |
| `after_return`        | `return` 语句之后的代码。                    |
| `after_break`         | `break` 语句之后的代码。                     |
| `after_continue`      | `continue` 语句之后的代码。                  |
| `after_raise`         | `raise` 语句之后的代码。                     |
| `unreachable_branch`  | 永远不会执行的条件分支。                     |

### `DeadCodeLocation` 对象

| 字段           | 类型    | 说明               |
| -------------- | ------- | ------------------ |
| `file_path`    | string  | 源文件路径。       |
| `start_line`   | integer | 从 1 开始的起始行。|
| `end_line`     | integer | 从 1 开始的结束行。|
| `start_column` | integer | 从 0 开始的起始列。|
| `end_column`   | integer | 从 0 开始的结束列。|

### `summary` 对象（`DeadCodeSummary`）

| 字段                       | 类型    | 说明                                     |
| -------------------------- | ------- | ---------------------------------------- |
| `total_files`              | integer | 分析的文件数。                           |
| `total_functions`          | integer | 分析的函数数。                           |
| `total_findings`           | integer | 所有文件的发现总数。                     |
| `files_with_dead_code`     | integer | 至少有一个发现的文件数。                 |
| `functions_with_dead_code` | integer | 至少有一个发现的函数数。                 |
| `critical_findings`        | integer | 严重性为 `critical` 的发现数。           |
| `warning_findings`         | integer | 严重性为 `warning` 的发现数。            |
| `info_findings`            | integer | 严重性为 `info` 的发现数。               |
| `findings_by_reason`       | object \| null | 按 `reason` 值分组的直方图。       |
| `total_blocks`             | integer | 所有函数的 CFG 块总数。                  |
| `dead_blocks`              | integer | 所有函数中不可达的 CFG 块数。            |
| `overall_dead_ratio`       | number  | `dead_blocks / total_blocks`，`0`–`1`。  |

## `clone` 对象

对应 `domain.CloneResponse`。全部使用 snake_case 字段名。

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

### `clones[]` 元素（`Clone`）

| 字段         | 类型    | 说明                                                 |
| ------------ | ------- | ---------------------------------------------------- |
| `id`         | integer | 克隆标识符，在响应内唯一。                           |
| `type`       | integer | 克隆类型（整数）：`1`、`2`、`3` 或 `4`。            |
| `location`   | object  | 见 [`CloneLocation`](#clonelocation-object)。        |
| `content`    | string  | 原始源代码文本。仅在设置 `--show-content` 时存在。   |
| `hash`       | string  | 指纹哈希（算法取决于克隆类型）。                     |
| `size`       | integer | AST 节点数。                                         |
| `line_count` | integer | 片段的行数。                                         |
| `complexity` | integer | 片段的圈复杂度。                                     |

`type` 枚举值（整数）：

| 值    | 含义                                                         |
| ----- | ------------------------------------------------------------ |
| `1`   | Type-1：除空白/注释外完全相同。                              |
| `2`   | Type-2：语法相同，但标识符/字面量不同。                      |
| `3`   | Type-3：结构相似但有修改。                                   |
| `4`   | Type-4：语义等价，语法不同。                                 |

### `CloneLocation` 对象

| 字段         | 类型    | 说明               |
| ------------ | ------- | ------------------ |
| `file_path`  | string  | 源文件路径。       |
| `start_line` | integer | 从 1 开始的起始行。|
| `end_line`   | integer | 从 1 开始的结束行。|
| `start_col`  | integer | 从 0 开始的起始列。|
| `end_col`    | integer | 从 0 开始的结束列。|

### `clone_pairs[]` 元素（`ClonePair`）

| 字段         | 类型    | 说明                                           |
| ------------ | ------- | ---------------------------------------------- |
| `id`         | integer | 克隆对标识符。                                 |
| `clone1`     | object  | 第一个克隆（`Clone` 对象）。                   |
| `clone2`     | object  | 第二个克隆（`Clone` 对象）。                   |
| `similarity` | number  | 相似度评分，`0`–`1`。                          |
| `distance`   | number  | 树编辑距离（Type-3）或 `0`。                   |
| `type`       | integer | 克隆类型（与 `clones[].type` 相同的枚举）。    |
| `confidence` | number  | 检测器置信度，`0`–`1`。                        |

### `clone_groups[]` 元素（`CloneGroup`）

| 字段         | 类型    | 说明                                           |
| ------------ | ------- | ---------------------------------------------- |
| `id`         | integer | 组标识符。                                     |
| `clones`     | array   | 成员 `Clone` 对象。                            |
| `type`       | integer | 主要克隆类型。                                 |
| `similarity` | number  | 代表性相似度，`0`–`1`。                        |
| `size`       | integer | 成员数量（`len(clones)`）。                    |

### `statistics` 对象（`CloneStatistics`）

| 字段                 | 类型    | 说明                                             |
| -------------------- | ------- | ------------------------------------------------ |
| `total_fragments`    | integer | 提取的所有片段（函数、类等）。                   |
| `total_clones`       | integer | 被分类为克隆的片段数。                           |
| `total_clone_pairs`  | integer | 检测到的克隆对数。                               |
| `total_clone_groups` | integer | 组的数量。                                       |
| `clones_by_type`     | object \| null | 按类型标签（`Type-1`…`Type-4`）到数量的映射。 |
| `average_similarity` | number  | 所有克隆对的平均相似度，`0`–`1`。                |
| `lines_analyzed`     | integer | 被考虑的源代码总行数。                           |
| `nodes_analyzed`     | integer | 被考虑的 AST 节点总数。                          |
| `files_analyzed`     | integer | 贡献了片段的不同文件数。                         |

`CloneResponse` 的其他字段：

| 字段          | 类型    | 说明                                       |
| ------------- | ------- | ------------------------------------------ |
| `duration_ms` | integer | 克隆检测耗时（毫秒）。                    |
| `success`     | boolean | 正常完成时为 `true`。                      |
| `error`       | string \| absent | `success=false` 时的错误信息。    |

## `cbo` 对象

对应 `domain.CBOResponse`。嵌套字段名为 Go PascalCase。

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

### `Classes[]` 元素（`ClassCoupling`）

| 字段          | 类型    | 说明                                |
| ------------- | ------- | ----------------------------------- |
| `Name`        | string  | 类名。                              |
| `FilePath`    | string  | 源文件路径。                        |
| `StartLine`   | integer | 从 1 开始的起始行。                 |
| `EndLine`     | integer | 从 1 开始的结束行。                 |
| `Metrics`     | object  | 见 [`CBOMetrics`](#cbometrics-object)。 |
| `RiskLevel`   | string  | 取值：`low`、`medium`、`high`。     |
| `IsAbstract`  | boolean | 抽象类时为 `true`。                 |
| `BaseClasses` | array of string \| null | 直接基类。          |

### `CBOMetrics` 对象

| 字段                          | 类型    | 说明                                              |
| ----------------------------- | ------- | ------------------------------------------------- |
| `CouplingCount`               | integer | CBO 值：此类依赖的不同类数。                      |
| `InheritanceDependencies`     | integer | 来自基类的依赖数。                                |
| `TypeHintDependencies`        | integer | 来自类型注解的依赖数。                            |
| `InstantiationDependencies`   | integer | 来自对象实例化的依赖数。                          |
| `AttributeAccessDependencies` | integer | 来自方法调用和属性访问的依赖数。                  |
| `ImportDependencies`          | integer | 来自显式导入的依赖数。                            |
| `DependentClasses`            | array of string \| null | 耦合类的名称。            |

### `Summary` 对象（`CBOSummary`）

| 字段                       | 类型    | 说明                                      |
| -------------------------- | ------- | ----------------------------------------- |
| `TotalClasses`             | integer | 分析的类总数。                            |
| `AverageCBO`               | number  | 平均 CBO。                                |
| `MaxCBO`                   | integer | 观测到的最大 CBO。                        |
| `MinCBO`                   | integer | 观测到的最小 CBO。                        |
| `ClassesAnalyzed`          | integer | 具有有效指标的类数。                      |
| `FilesAnalyzed`            | integer | 贡献了至少一个类的文件数。                |
| `LowRiskClasses`           | integer | CBO ≤ 低阈值（默认 `3`）的类数。          |
| `MediumRiskClasses`        | integer | 低 < CBO ≤ 中等阈值的类数。               |
| `HighRiskClasses`          | integer | CBO > 中等阈值（默认 `7`）的类数。        |
| `CBODistribution`          | object \| null | 按区间标签到数量的直方图。         |
| `MostCoupledClasses`       | array \| null | 按 CBO 排名的前 10 个类（`ClassCoupling`）。 |
| `MostDependedUponClasses`  | array of string \| null | 入度最高的类。        |

## `lcom` 对象

对应 `domain.LCOMResponse`。嵌套字段名为 Go PascalCase。

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

### `Classes[]` 元素（`ClassCohesion`）

| 字段        | 类型    | 说明                                     |
| ----------- | ------- | ---------------------------------------- |
| `Name`      | string  | 类名。                                   |
| `FilePath`  | string  | 源文件路径。                             |
| `StartLine` | integer | 从 1 开始的起始行。                      |
| `EndLine`   | integer | 从 1 开始的结束行。                      |
| `Metrics`   | object  | 见 [`LCOMMetrics`](#lcommetrics-object)。|
| `RiskLevel` | string  | 取值：`low`、`medium`、`high`。          |

### `LCOMMetrics` 对象

| 字段                | 类型    | 说明                                                |
| ------------------- | ------- | --------------------------------------------------- |
| `LCOM4`             | integer | 方法-变量图中的连通分量数。                         |
| `TotalMethods`      | integer | 类中的所有方法数。                                  |
| `ExcludedMethods`   | integer | 从 LCOM4 中排除的方法（`@classmethod`、`@staticmethod`）。 |
| `InstanceVariables` | integer | 访问的不同 `self.x` 变量数。                        |
| `MethodGroups`      | array of array of string \| null | 按连通分量分组的方法名。 |

### `Summary` 对象（`LCOMSummary`）

| 字段                   | 类型    | 说明                                    |
| ---------------------- | ------- | --------------------------------------- |
| `TotalClasses`         | integer | 分析的类数。                            |
| `AverageLCOM`          | number  | 平均 LCOM4。                            |
| `MaxLCOM`              | integer | 观测到的最大 LCOM4。                    |
| `MinLCOM`              | integer | 观测到的最小 LCOM4。                    |
| `ClassesAnalyzed`      | integer | 具有有效指标的类数。                    |
| `FilesAnalyzed`        | integer | 贡献了至少一个类的文件数。              |
| `LowRiskClasses`       | integer | LCOM4 ≤ 低阈值（默认 `2`）的类数。      |
| `MediumRiskClasses`    | integer | 低 < LCOM4 ≤ 中等阈值的类数。           |
| `HighRiskClasses`      | integer | LCOM4 > 中等阈值（默认 `5`）的类数。    |
| `LCOMDistribution`     | object \| null | 按区间标签到数量的直方图。       |
| `LeastCohesiveClasses` | array \| null | 按 LCOM4 排名的前 10 个类（`ClassCohesion`）。 |

## `system` 对象

对应 `domain.SystemAnalysisResponse`。嵌套字段名为 Go PascalCase。

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

### `Summary` 对象（`SystemAnalysisSummary`）

| 字段                       | 类型    | 说明                                    |
| -------------------------- | ------- | --------------------------------------- |
| `TotalModules`             | integer | 分析的模块总数。                        |
| `TotalPackages`            | integer | 包总数。                                |
| `TotalDependencies`        | integer | 依赖边总数。                            |
| `ProjectRoot`              | string  | 项目根目录。                            |
| `OverallQualityScore`      | number  | 综合质量评分，`0`–`100`。               |
| `MaintainabilityScore`     | number  | 平均可维护性指数。                      |
| `ArchitectureScore`        | number  | 架构合规评分。                          |
| `ModularityScore`          | number  | 系统模块化评分。                        |
| `TechnicalDebtHours`       | number  | 估计的技术债务总时长（小时）。          |
| `AverageCoupling`          | number  | 平均模块耦合度。                        |
| `AverageInstability`       | number  | 平均不稳定性（I）。                     |
| `CyclicDependencies`       | integer | 参与循环的模块数。                      |
| `ArchitectureViolations`   | integer | 架构规则违规数。                        |
| `HighRiskModules`          | integer | 标记为高风险的模块数。                  |
| `CriticalIssues`           | integer | 严重问题数。                            |
| `RefactoringCandidates`    | integer | 需要重构的模块数。                      |
| `ArchitectureImprovements` | integer | 建议的架构改进数。                      |

### `DependencyAnalysis` 对象

| 字段                   | 类型    | 说明                                                         |
| ---------------------- | ------- | ------------------------------------------------------------ |
| `TotalModules`         | integer | 依赖图中的模块总数。                                         |
| `TotalDependencies`    | integer | 边总数。                                                     |
| `RootModules`          | array of string | 没有出向依赖的模块。                                 |
| `LeafModules`          | array of string | 没有入向依赖的模块。                                 |
| `ModuleMetrics`        | object  | 模块名到 `ModuleDependencyMetrics` 的映射。                  |
| `DependencyMatrix`     | object  | 模块到模块到布尔值的映射。                                   |
| `CircularDependencies` | object  | 循环检测结果；包含 `Cycles`（数组）和 `TotalCycles`（整数）。|
| `CouplingAnalysis`     | object  | 每模块耦合指标：`Ca`、`Ce`、`Instability`、`Abstractness`、`Distance`。 |
| `LongestChains`        | array   | `DependencyPath` 对象数组。                                  |
| `MaxDepth`             | integer | 最大依赖深度。                                               |

### `ModuleDependencyMetrics` 对象

| 字段                     | 类型    | 说明                                             |
| ------------------------ | ------- | ------------------------------------------------ |
| `ModuleName`             | string  | 完全限定模块名。                                 |
| `Package`                | string  | 父包。                                           |
| `FilePath`               | string  | 源文件路径。                                     |
| `IsPackage`              | boolean | 如果是包（有 `__init__.py`）则为 `true`。        |
| `LinesOfCode`            | integer | 代码总行数。                                     |
| `FunctionCount`          | integer | 函数数量。                                       |
| `ClassCount`             | integer | 类数量。                                         |
| `PublicInterface`        | array of string | `__all__` 中或顶级公共名称。             |
| `AfferentCoupling`       | integer | Ca — 依赖此模块的模块数。                        |
| `EfferentCoupling`       | integer | Ce — 此模块依赖的模块数。                        |
| `Instability`            | number  | `I = Ce / (Ca + Ce)`，`0`–`1`。                  |
| `Abstractness`           | number  | A — 抽象元素 / 总元素，`0`–`1`。                 |
| `Distance`               | number  | `D = |A + I - 1|`，`0`–`1`。与主序列的距离。     |
| `Maintainability`        | number  | 可维护性指数，`0`–`100`。                        |
| `TechnicalDebt`          | number  | 估计的技术债务（小时）。                         |
| `RiskLevel`              | string  | 取值：`low`、`medium`、`high`。                  |
| `DirectDependencies`     | array of string | 直接依赖。                               |
| `TransitiveDependencies` | array of string | 所有传递依赖。                           |
| `Dependents`             | array of string | 依赖此模块的模块。                       |

### `CircularDependencyAnalysis` 对象

| 字段                       | 类型    | 说明                                          |
| -------------------------- | ------- | --------------------------------------------- |
| `HasCircularDependencies`  | boolean | 存在循环时为 `true`。                         |
| `TotalCycles`              | integer | 循环数量。                                    |
| `TotalModulesInCycles`     | integer | 涉及循环的模块数。                            |
| `CircularDependencies`     | array   | `CircularDependency` 对象数组。               |
| `CycleBreakingSuggestions` | array of string | 打破循环的建议。                      |
| `CoreInfrastructure`       | array of string | 出现在多个循环中的模块。              |

`CircularDependency.Severity` 枚举值：`low`、`medium`、`high`、`critical`。

### `CouplingAnalysis` 对象

| 字段                    | 类型    | 说明                                       |
| ----------------------- | ------- | ------------------------------------------ |
| `AverageCoupling`       | number  | 跨模块的平均耦合度。                       |
| `CouplingDistribution`  | object  | 耦合值（整数键）到数量的映射。             |
| `HighlyCoupledModules`  | array of string | 高耦合模块。                       |
| `LooselyCoupledModules` | array of string | 低耦合模块。                       |
| `AverageInstability`    | number  | 平均不稳定性。                             |
| `StableModules`         | array of string | 低不稳定性模块。                   |
| `InstableModules`       | array of string | 高不稳定性模块。                   |
| `MainSequenceDeviation` | number  | 与主序列的平均距离，`0`–`1`。              |
| `ZoneOfPain`            | array of string | 稳定 + 具体的模块。               |
| `ZoneOfUselessness`     | array of string | 不稳定 + 抽象的模块。             |
| `MainSequence`          | array of string | 位置良好的模块。                   |

### `ArchitectureAnalysis` 对象

| 字段                  | 类型    | 说明                                                |
| --------------------- | ------- | --------------------------------------------------- |
| `ComplianceScore`     | number  | 合规评分，`0`–`1`。`1.0` = 完全合规。               |
| `TotalViolations`     | integer | 发现的违规总数。                                    |
| `TotalRules`          | integer | 评估的规则总数。                                    |
| `LayerAnalysis`       | object \| null | 分层分析结果。                               |
| `CohesionAnalysis`    | object \| null | 包内聚性分析。                               |
| `ResponsibilityAnalysis` | object \| null | 单一职责原则违规分析。                    |
| `Violations`          | array   | `ArchitectureViolation` 对象数组。                  |
| `SeverityBreakdown`   | object  | 严重性到数量的映射。                                |
| `Recommendations`     | array   | `ArchitectureRecommendation` 对象数组。             |
| `RefactoringTargets`  | array of string | 需要重构的模块。                            |

`ArchitectureViolation.Type` 枚举值：`layer`、`cycle`、`coupling`、`responsibility`、`cohesion`。

`ArchitectureViolation.Severity` 枚举值：`info`、`warning`、`error`、`critical`。

## `suggestions` 数组

`Suggestion` 对象数组。使用 snake_case 字段名。

| 字段          | 类型    | 必需 | 说明                                      |
| ------------- | ------- | ---- | ----------------------------------------- |
| `category`    | string  | 是   | 见下方枚举。                              |
| `severity`    | string  | 是   | 取值：`critical`、`warning`、`info`。     |
| `effort`      | string  | 是   | 取值：`easy`、`moderate`、`hard`。        |
| `title`       | string  | 是   | 简短的人类可读标题。                      |
| `description` | string  | 是   | 完整描述。                                |
| `steps`       | array of string | 否 | 可操作的步骤。为空时省略。           |
| `file_path`   | string  | 否   | 源文件引用。                              |
| `function`    | string  | 否   | 函数名引用。                              |
| `class_name`  | string  | 否   | 类名引用。                                |
| `start_line`  | integer | 否   | 从 1 开始的行引用。为 `0` 时省略。        |
| `metric_value`| string  | 否   | 观测到的指标值（字符串形式）。            |
| `threshold`   | string  | 否   | 阈值（字符串形式）。                      |

`category` 枚举值：`complexity`、`dead_code`、`clone`、`coupling`、`cohesion`、`dependency`、`architecture`。

建议按优先级排序（严重性 x 工作量）。详见 `domain/suggestion.go` 中的确切排序函数。

## CSV 模式

CSV 输出通过 Go `encoding/csv` 包使用 RFC 4180 引用格式。

### `pyscn analyze --csv`

仅输出摘要。两列。纯 UTF-8 字符串，无类型注解。

| 列       | 类型   | 说明             |
| -------- | ------ | ---------------- |
| `Metric` | string | 指标名称。       |
| `Value`  | string | 指标值（字符串）。|

行（固定顺序）：

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

pyscn 目前不通过 CLI 公开每个分析器的 CSV 模式 — `--csv` 仅生成上述摘要。如需每个发现的详细信息，请使用 `--json` 或 `--yaml`。

## 时间戳和版本信息

| 字段           | 格式                    | 说明                                                    |
| -------------- | ----------------------- | ------------------------------------------------------- |
| `generated_at` | RFC 3339 (ISO 8601)     | `time.Time` 序列化；可能包含亚秒精度和时区偏移。       |
| `duration_ms`  | integer（毫秒）         | 实际分析耗时。                                          |
| `version`      | string（语义版本号）    | pyscn 发布版本，例如 `"0.14.0"`。                       |

## 调用各格式

`pyscn analyze` 接受 `--json`、`--yaml`、`--csv`、`--html`（默认）之一。没有 `--format` 标志，也没有独立的 `complexity` / `deadcode` / `clone` / `deps` 子命令。通过 `--select` 运行单个分析器。

```bash
pyscn analyze --json src/
pyscn analyze --yaml src/
pyscn analyze --csv  src/
pyscn analyze --html src/    # default
pyscn analyze --json --select complexity src/
pyscn analyze --csv  --select deadcode   src/
pyscn analyze --yaml --select clones     src/
```

输出文件保存在 `.pyscn/reports/` 中；路径和文件名详情见[输出格式](index.md)。

## 相关文档

- [HTML 报告](html-report.md) — HTML 输出规范。
- [健康评分](health-score.md) — `summary.health_score` 和分类评分的推导。
