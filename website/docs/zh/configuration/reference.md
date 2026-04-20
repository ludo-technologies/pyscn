# 配置参考手册

`.pyscn.toml`（或 `pyproject.toml` 中的 `[tool.pyscn.*]`）中所有可配置的键。运行 `pyscn init` 可生成带注释的初始配置文件。

---

## `[output]`

控制结果的输出方式。

| 键               | 类型    | 默认值        | 说明 |
| ---------------- | ------- | ------------- | --- |
| `format`         | string  | `"text"`      | `text`、`json`、`yaml`、`csv` 或 `html`。`--json` 等 CLI 标志会覆盖此设置。 |
| `directory`      | string  | `""`          | 输出目录。为空时默认为当前工作目录下的 `.pyscn/reports/`。 |
| `show_details`   | bool    | `false`       | 在摘要中包含每个发现的详细信息。 |
| `sort_by`        | string  | `"complexity"`| `name`、`complexity` 或 `risk`。 |
| `min_complexity` | int     | `1`           | 过滤掉低于此复杂度的函数。设置后会覆盖 `[complexity].min_complexity`。 |

---

## `[complexity]`

圈复杂度分析。

| 键                 | 类型 | 默认值 | 说明 |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | 运行分析器。 |
| `low_threshold`    | int  | `9`     | "低风险"的上限（含）。 |
| `medium_threshold` | int  | `19`    | "中等风险"的上限。 |
| `max_complexity`   | int  | `0`     | CI 失败阈值。`0` = 无限制。 |
| `min_complexity`   | int  | `1`     | 不报告低于此值的函数。 |
| `report_unchanged` | bool | `true`  | 包含复杂度为 1 的函数。 |

详见 [high-cyclomatic-complexity](../rules/high-cyclomatic-complexity.md) 中的阈值指南。

---

## `[dead_code]`

死代码检测。

| 键                               | 类型     | 默认值       | 说明 |
| -------------------------------- | -------- | ------------ | --- |
| `enabled`                        | bool     | `true`       | 运行分析器。 |
| `min_severity`                   | string   | `"warning"`  | `info`、`warning` 或 `critical`。 |
| `show_context`                   | bool     | `false`      | 包含周围的源代码行。 |
| `context_lines`                  | int      | `3`          | 上下文行数（0–20）。 |
| `sort_by`                        | string   | `"severity"` | `severity`、`line`、`file` 或 `function`。 |
| `detect_after_return`            | bool     | `true`       | 标记 `return` 之后的语句。 |
| `detect_after_break`             | bool     | `true`       | 标记 `break` 之后的语句。 |
| `detect_after_continue`          | bool     | `true`       | 标记 `continue` 之后的语句。 |
| `detect_after_raise`             | bool     | `true`       | 标记 `raise` 之后的语句。 |
| `detect_unreachable_branches`    | bool     | `true`       | 标记永远不会执行的分支。 |
| `ignore_patterns`                | string[] | `[]`         | 要忽略的行的正则表达式模式。 |

---

## `[clones]`

克隆检测（配置项最多的分析器）。

### 片段选择

| 键               | 类型 | 默认值 | 说明 |
| ---------------- | ---- | ------- | --- |
| `min_lines`      | int  | `10`    | 被视为片段的最小行数。 |
| `min_nodes`      | int  | `20`    | 最小 AST 节点数。 |
| `skip_docstrings`| bool | `true`  | 哈希时跳过文档字符串。 |

### 类型阈值 (0.0–1.0)

| 键                     | 默认值 | 克隆类型 |
| ---------------------- | ------- | --- |
| `type1_threshold`      | `0.85`  | 完全相同（仅空白/注释不同）。 |
| `type2_threshold`      | `0.75`  | 重命名了标识符/字面量。 |
| `type3_threshold`      | `0.70`  | 结构相似但有修改。 |
| `type4_threshold`      | `0.65`  | 语义等价。 |
| `similarity_threshold` | `0.65`  | 任何克隆的全局最小相似度。 |

### 算法

| 键                  | 类型   | 默认值     | 说明 |
| ------------------- | ------ | ---------- | --- |
| `cost_model_type`   | string | `"python"` | `default`、`python` 或 `weighted`。 |
| `ignore_literals`   | bool   | `false`    | 将不同的字面量视为等价。 |
| `ignore_identifiers`| bool   | `false`    | 将不同的变量名视为等价。 |
| `max_edit_distance` | float  | `50.0`     | 树编辑距离上限。 |
| `enable_dfa`        | bool   | `true`     | 为 Type-4 启用数据流分析。 |
| `enabled_clone_types` | string[] | all     | `type1`、`type2`、`type3`、`type4` 的子集。 |

### LSH 加速

| 键                         | 类型           | 默认值   | 说明 |
| -------------------------- | -------------- | -------- | --- |
| `lsh_enabled`              | `true\|false\|"auto"` | `"auto"` | 启用 LSH（`auto` = 根据片段数量决定）。 |
| `lsh_auto_threshold`       | int            | `500`    | 自动启用的片段数量阈值。 |
| `lsh_similarity_threshold` | float          | `0.50`   | LSH 候选预过滤阈值。 |
| `lsh_bands`                | int            | `32`     | LSH 分段数。 |
| `lsh_rows`                 | int            | `4`      | 每段的行数。 |
| `lsh_hashes`               | int            | `128`    | 哈希函数数量。 |

### 分组

| 键                   | 类型   | 默认值        | 说明 |
| -------------------- | ------ | ------------- | --- |
| `grouping_mode`      | string | `"connected"` | `connected`、`star`、`complete_linkage`、`k_core`。 |
| `grouping_threshold` | float  | `0.65`        | 分组的最小相似度。 |
| `k_core_k`           | int    | `2`           | `k_core` 模式的 k 参数。 |

### 性能

| 键                | 类型 | 默认值 | 说明 |
| ----------------- | ---- | ------- | --- |
| `max_memory_mb`   | int  | `100`   | 内存上限（MB）。`0` = 无限制。 |
| `batch_size`      | int  | `100`   | 每批处理的文件数。 |
| `enable_batching` | bool | `true`  | 启用批处理。 |
| `max_goroutines`  | int  | `4`     | 并发工作协程数。 |
| `timeout_seconds` | int  | `300`   | 单次分析超时时间。 |

### 输出过滤

| 键              | 类型  | 默认值          | 说明 |
| --------------- | ----- | --------------- | --- |
| `min_similarity`| float | `0.0`           | 过滤掉低于此值的克隆对。 |
| `max_similarity`| float | `1.0`           | 过滤掉高于此值的克隆对。 |
| `max_results`   | int   | `10000`         | 报告的最大克隆对数。`0` = 无限制。 |
| `show_details`  | bool  | `false`         | 详细输出。 |
| `show_content`  | bool  | `false`         | 在报告中包含源代码。 |
| `sort_by`       | string| `"similarity"`  | `similarity`、`size`、`location`、`type`。 |
| `group_clones`  | bool  | `true`          | 对相关克隆进行分组。 |

---

## `[cbo]`

对象间耦合度（类耦合）。

| 键                 | 类型 | 默认值 | 说明 |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | 运行分析器。 |
| `low_threshold`    | int  | `3`     | "低风险"的上限。 |
| `medium_threshold` | int  | `7`     | "中等风险"的上限。 |
| `min_cbo`          | int  | `0`     | 过滤掉低于此 CBO 值的类。 |
| `max_cbo`          | int  | `0`     | 过滤掉高于此值的类。`0` = 无限制。 |
| `show_zeros`       | bool | `false` | 包含 CBO 为 0 的类。 |
| `include_builtins` | bool | `false` | 将 `list`/`dict`/`str` 计为依赖项。 |
| `include_imports`  | bool | `true`  | 将导入的模块引用计为依赖项。 |

---

## `[lcom]`

方法缺乏凝聚度（LCOM4）。

| 键                 | 类型 | 默认值 | 说明 |
| ------------------ | ---- | ------- | --- |
| `low_threshold`    | int  | `2`     | "低风险"的上限（良好的内聚性）。 |
| `medium_threshold` | int  | `5`     | "中等风险"的上限。 |

---

## `[analysis]`

文件发现规则。

| 键                 | 类型     | 默认值        | 说明 |
| ------------------ | -------- | ------------- | --- |
| `recursive`        | bool     | `true`        | 递归进入子目录。 |
| `follow_symlinks`  | bool     | `false`       | 跟随符号链接。 |
| `include_patterns` | string[] | `["**/*.py"]` | 包含的 glob 模式。 |
| `exclude_patterns` | string[] | 见下文        | 排除的 glob 模式。 |

默认 `exclude_patterns`：

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

分层验证。所有键均为可选 — 如果未定义层，架构分析将以宽松模式运行。

| 键                         | 类型  | 默认值 | 说明 |
| -------------------------- | ----- | ------- | --- |
| `enabled`                  | bool  | `true`  | 运行分层验证。 |
| `validate_layers`          | bool  | `true`  | 检查层间规则。 |
| `validate_cohesion`        | bool  | `true`  | 检查层内聚性。 |
| `validate_responsibility`  | bool  | `true`  | 检查每层的职责上限。 |
| `strict_mode`              | bool  | `true`  | 严格验证。 |
| `fail_on_violations`       | bool  | `false` | 违规时以非零退出码退出。 |
| `min_cohesion`             | float | `0.5`   | 最小层内聚度。 |
| `max_coupling`             | int   | `10`    | 最大层间耦合度。 |
| `max_responsibilities`     | int   | `3`     | 每层最大职责数。 |

### 层定义

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

### 层规则

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

模块依赖分析。对于 `pyscn check` 为**可选启用**；`pyscn analyze` 始终运行（除非跳过）。

| 键                   | 类型   | 默认值  | 说明 |
| -------------------- | ------ | ------- | --- |
| `enabled`            | bool   | `false` | 运行分析器（analyze 命令无论此设置都会运行）。 |
| `include_stdlib`     | bool   | `false` | 包含标准库导入。 |
| `include_third_party`| bool   | `true`  | 包含第三方库导入。 |
| `follow_relative`    | bool   | `true`  | 跟随相对导入。 |
| `detect_cycles`      | bool   | `true`  | 查找循环导入。 |
| `calculate_metrics`  | bool   | `true`  | 计算 Ca/Ce/I/A/D 指标。 |
| `find_long_chains`   | bool   | `true`  | 报告最长依赖链。 |
| `cycle_reporting`    | string | `"summary"` | `all`、`critical`、`summary`。 |
| `max_cycles_to_show` | int    | `10`    | 报告的最大循环数。 |
| `sort_by`            | string | `"name"` | `name`、`coupling`、`instability`、`distance`、`risk`。 |
| `show_matrix`        | bool   | `false` | 包含依赖矩阵。 |
| `generate_dot_graph` | bool   | `false` | 输出 Graphviz DOT 格式。 |

---

## `[mock_data]`

模拟/占位数据检测。**可选启用**。

| 键               | 类型     | 默认值      | 说明 |
| ---------------- | -------- | ----------- | --- |
| `enabled`        | bool     | `false`     | 运行分析器。 |
| `min_severity`   | string   | `"warning"` | `info`、`warning`、`error`。 |
| `ignore_tests`   | bool     | `true`      | 跳过测试文件。 |
| `keywords`       | string[] | 内置        | 被标记为模拟数据指标的关键词。 |
| `domains`        | string[] | 内置        | 被标记的域名（`example.com`、`test.com` 等）。 |
| `ignore_patterns`| string[] | `[]`        | 要跳过的文件/正则表达式模式。 |

---

## `[di]`

依赖注入反模式检测。**可选启用**。

| 键                             | 类型   | 默认值      | 说明 |
| ------------------------------ | ------ | ----------- | --- |
| `enabled`                      | bool   | `false`     | 运行分析器。 |
| `min_severity`                 | string | `"warning"` | `info`、`warning`、`error`。 |
| `constructor_param_threshold`  | int    | `5`         | 标记参数超过此数量的 `__init__`。 |

---

## CLI 标志 → 配置键映射

不直接映射到配置键的标志（`--select`、`--skip-*`、`--no-open`）会在已加载配置的基础上生效。

| CLI 标志                | 配置键                            |
| ----------------------- | --------------------------------- |
| `--config <path>`       | — （覆盖自动发现）               |
| `--json/--yaml/--csv/--html` | `[output] format`            |
| `--min-complexity`      | `[complexity] min_complexity`     |
| `--max-complexity`      | `[complexity] max_complexity`     |
| `--min-severity`        | `[dead_code] min_severity`        |
| `--clone-threshold`     | `[clones] similarity_threshold`   |
| `--min-cbo`             | `[cbo] min_cbo`                   |
| `--max-cycles`          | — （仅 check 命令）              |

## 另请参阅

- [配置文件格式](format.md) — 发现与优先级。
- [示例](examples.md) — 即用型配置。
