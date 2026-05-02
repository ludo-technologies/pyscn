# Configuration Reference

Every configurable key in `.pyscn.toml` (or `[tool.pyscn.*]` in `pyproject.toml`). Run `pyscn init` to generate a commented starter file.

---

## `[output]`

Controls how results are reported.

| Key              | Type    | Default       | Description |
| ---------------- | ------- | ------------- | --- |
| `format`         | string  | `"text"`      | `text`, `json`, `yaml`, `csv`, or `html`. CLI flags like `--json` override this. |
| `directory`      | string  | `""`          | Output directory. Empty = `.pyscn/reports/` under CWD. |
| `show_details`   | bool    | `false`       | Include per-finding detail in the summary. |
| `sort_by`        | string  | `"complexity"`| `name`, `complexity`, or `risk`. |
| `min_complexity` | int     | `1`           | Filter out functions below this complexity. Overrides `[complexity].min_complexity` when set. |

---

## `[complexity]`

Cyclomatic complexity analysis.

| Key                | Type | Default | Description |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | Run the analyzer. |
| `low_threshold`    | int  | `9`     | Upper bound for "low risk" (inclusive). |
| `medium_threshold` | int  | `19`    | Upper bound for "medium risk". |
| `max_complexity`   | int  | `0`     | CI failure threshold. `0` = no limit. |
| `min_complexity`   | int  | `1`     | Don't report functions below this. |
| `report_unchanged` | bool | `true`  | Include functions with complexity = 1. |

See [high-cyclomatic-complexity](../rules/high-cyclomatic-complexity.md) for thresholds guidance.

---

## `[dead_code]`

Dead code detection.

| Key                              | Type   | Default      | Description |
| -------------------------------- | ------ | ------------ | --- |
| `enabled`                        | bool   | `true`       | Run the analyzer. |
| `min_severity`                   | string | `"warning"`  | `info`, `warning`, or `critical`. |
| `show_context`                   | bool   | `false`      | Include surrounding source lines. |
| `context_lines`                  | int    | `3`          | Lines of context (0–20). |
| `sort_by`                        | string | `"severity"` | `severity`, `line`, `file`, or `function`. |
| `detect_after_return`            | bool   | `true`       | Flag statements after `return`. |
| `detect_after_break`             | bool   | `true`       | Flag statements after `break`. |
| `detect_after_continue`          | bool   | `true`       | Flag statements after `continue`. |
| `detect_after_raise`             | bool   | `true`       | Flag statements after `raise`. |
| `detect_unreachable_branches`    | bool   | `true`       | Flag branches that can never be taken. |
| `ignore_patterns`                | string[] | `[]`       | Regex patterns for lines to ignore. |

---

## `[clones]`

Clone detection (the most configurable analyzer).

### Fragment selection

| Key              | Type | Default | Description |
| ---------------- | ---- | ------- | --- |
| `min_lines`      | int  | `10`    | Minimum lines to consider a fragment. |
| `min_nodes`      | int  | `20`    | Minimum AST nodes. |
| `skip_docstrings`| bool | `true`  | Skip docstrings when hashing. |

### Type thresholds (0.0–1.0)

| Key                    | Default | Clone type |
| ---------------------- | ------- | --- |
| `type1_threshold`      | `0.85`  | Identical (whitespace/comments only). |
| `type2_threshold`      | `0.75`  | Renamed identifiers/literals. |
| `type3_threshold`      | `0.70`  | Structurally similar with modifications. |
| `type4_threshold`      | `0.65`  | Semantic equivalence. |
| `similarity_threshold` | `0.65`  | Global minimum for any clone. |

### Algorithm

| Key                 | Type   | Default    | Description |
| ------------------- | ------ | ---------- | --- |
| `cost_model_type`   | string | `"python"` | `default`, `python`, or `weighted`. |
| `ignore_literals`   | bool   | `false`    | Treat different literals as equivalent. |
| `ignore_identifiers`| bool   | `false`    | Treat different variable names as equivalent. |
| `max_edit_distance` | float  | `50.0`     | Cap on tree edit distance. |
| `enable_dfa`        | bool   | `true`     | Data-flow analysis for Type-4. |
| `enabled_clone_types` | string[] | all     | Subset of `type1`, `type2`, `type3`, `type4`. |

### LSH acceleration

| Key                        | Type           | Default  | Description |
| -------------------------- | -------------- | -------- | --- |
| `lsh_enabled`              | `true\|false\|"auto"` | `"auto"` | Enable LSH (`auto` = based on fragment count). |
| `lsh_auto_threshold`       | int            | `500`    | Fragment count threshold for auto-enable. |
| `lsh_similarity_threshold` | float          | `0.50`   | LSH candidate pre-filter. |
| `lsh_bands`                | int            | `32`     | LSH bands. |
| `lsh_rows`                 | int            | `4`      | Rows per band. |
| `lsh_hashes`               | int            | `128`    | Hash function count. |

### Grouping

| Key                  | Type   | Default       | Description |
| -------------------- | ------ | ------------- | --- |
| `grouping_mode`      | string | `"connected"` | `connected`, `star`, `complete_linkage`, `k_core`. |
| `grouping_threshold` | float  | `0.65`        | Minimum similarity for grouping. |
| `k_core_k`           | int    | `2`           | k parameter for `k_core` mode. |

### Performance

| Key               | Type | Default | Description |
| ----------------- | ---- | ------- | --- |
| `max_memory_mb`   | int  | `100`   | Memory cap (MB). `0` = no limit. |
| `batch_size`      | int  | `100`   | Files per batch. |
| `enable_batching` | bool | `true`  | Process in batches. |
| `max_goroutines`  | int  | `4`     | Concurrent workers. |
| `timeout_seconds` | int  | `300`   | Per-analysis timeout. |

### Output filtering

| Key             | Type  | Default         | Description |
| --------------- | ----- | --------------- | --- |
| `min_similarity`| float | `0.0`           | Filter out pairs below this. |
| `max_similarity`| float | `1.0`           | Filter out pairs above this. |
| `max_results`   | int   | `10000`         | Maximum pairs to report. `0` = no limit. |
| `show_details`  | bool  | `false`         | Verbose output. |
| `show_content`  | bool  | `false`         | Include source in the report. |
| `sort_by`       | string| `"similarity"`  | `similarity`, `size`, `location`, `type`. |
| `group_clones`  | bool  | `true`          | Group related clones. |

---

## `[cbo]`

Coupling Between Objects (class coupling).

| Key                | Type | Default | Description |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | Run the analyzer. |
| `low_threshold`    | int  | `3`     | Upper bound for "low risk". |
| `medium_threshold` | int  | `7`     | Upper bound for "medium risk". |
| `min_cbo`          | int  | `0`     | Filter out classes below this CBO. |
| `max_cbo`          | int  | `0`     | Filter out classes above this. `0` = no limit. |
| `show_zeros`       | bool | `false` | Include classes with CBO = 0. |
| `include_builtins` | bool | `false` | Count `list`/`dict`/`str` as dependencies. |
| `include_imports`  | bool | `true`  | Count imported module references. |

---

## `[lcom]`

Lack of Cohesion of Methods (LCOM4).

| Key                | Type | Default | Description |
| ------------------ | ---- | ------- | --- |
| `low_threshold`    | int  | `2`     | Upper bound for "low risk" (good cohesion). |
| `medium_threshold` | int  | `5`     | Upper bound for "medium risk". |

---

## `[analysis]`

File discovery rules.

| Key                | Type     | Default       | Description |
| ------------------ | -------- | ------------- | --- |
| `recursive`        | bool     | `true`        | Descend into subdirectories. |
| `follow_symlinks`  | bool     | `false`       | Follow symbolic links. |
| `include_patterns` | string[] | `["**/*.py"]` | Glob patterns to include. |
| `exclude_patterns` | string[] | see below     | Glob patterns to exclude. |

Default `exclude_patterns`:

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

Layer validation. All keys optional — if you don't define layers, architecture analysis runs in permissive mode.

| Key                        | Type  | Default | Description |
| -------------------------- | ----- | ------- | --- |
| `enabled`                  | bool  | `true`  | Run layer validation. |
| `validate_layers`          | bool  | `true`  | Check layer-to-layer rules. |
| `validate_cohesion`        | bool  | `true`  | Check package cohesion. |
| `validate_responsibility`  | bool  | `true`  | Check module responsibility count. |
| `strict_mode`              | bool  | `true`  | Strict validation. |
| `fail_on_violations`       | bool  | `false` | Non-zero exit on violation. |
| `min_cohesion`             | float | `0.5`   | Minimum package cohesion. |
| `max_coupling`             | int   | `10`    | Max inter-layer coupling. |
| `max_responsibilities`     | int   | `3`     | Max concerns per module. |
| `neutral_prefixes`         | string[] | `[]` | Top-level module segments to strip before matching layer packages. Useful when every module starts with the same project prefix (e.g. `app`, `src`). |

### Layer definitions

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

### Layer rules

```toml
[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain"]
```

### Neutral prefixes

If every module in the project starts with the same root segment (`app.`, `src.`, ...), layer matching can fail because the project prefix shadows the layer name. List those segments under `neutral_prefixes` and pyscn will strip them before resolving a module to a layer:

```toml
[architecture]
neutral_prefixes = ["app", "src"]
```

With this set, `app.routers.user_router` is matched as `routers.user_router` and resolves to the `presentation` layer.

---

## `[dependencies]`

Module dependency analysis. **Opt-in** for `pyscn check`; always on for `pyscn analyze` unless skipped.

| Key                  | Type   | Default | Description |
| -------------------- | ------ | ------- | --- |
| `enabled`            | bool   | `false` | Run the analyzer (analyze always runs it regardless). |
| `include_stdlib`     | bool   | `false` | Include standard-library imports. |
| `include_third_party`| bool   | `true`  | Include third-party imports. |
| `follow_relative`    | bool   | `true`  | Follow relative imports. |
| `detect_cycles`      | bool   | `true`  | Find circular imports. |
| `calculate_metrics`  | bool   | `true`  | Compute Ca/Ce/I/A/D. |
| `find_long_chains`   | bool   | `true`  | Report longest dependency chains. |
| `cycle_reporting`    | string | `"summary"` | `all`, `critical`, `summary`. |
| `max_cycles_to_show` | int    | `10`    | Cap on reported cycles. |
| `sort_by`            | string | `"name"` | `name`, `coupling`, `instability`, `distance`, `risk`. |
| `show_matrix`        | bool   | `false` | Include dependency matrix. |
| `generate_dot_graph` | bool   | `false` | Emit Graphviz DOT output. |

---

## `[mock_data]`

Mock/placeholder data detection. **Opt-in**.

| Key              | Type     | Default     | Description |
| ---------------- | -------- | ----------- | --- |
| `enabled`        | bool     | `false`     | Run the analyzer. |
| `min_severity`   | string   | `"warning"` | `info`, `warning`, `error`. |
| `ignore_tests`   | bool     | `true`      | Skip test files. |
| `keywords`       | string[] | built-in    | Words flagged as mock indicators. |
| `domains`        | string[] | built-in    | Domains flagged (`example.com`, `test.com`, etc.). |
| `ignore_patterns`| string[] | `[]`        | Files/regex patterns to skip. |

---

## `[di]`

Dependency Injection anti-pattern detection. **Opt-in**.

| Key                            | Type   | Default     | Description |
| ------------------------------ | ------ | ----------- | --- |
| `enabled`                      | bool   | `false`     | Run the analyzer. |
| `min_severity`                 | string | `"warning"` | `info`, `warning`, `error`. |
| `constructor_param_threshold`  | int    | `5`         | Flag `__init__` with more params. |

---

## CLI flag → config key map

Flags that don't map directly to a config key (`--select`, `--skip-*`, `--no-open`) work on top of whatever config you have loaded.

| CLI flag                | Config key                        |
| ----------------------- | --------------------------------- |
| `--config <path>`       | — (overrides discovery)           |
| `--json/--yaml/--csv/--html` | `[output] format`            |
| `--min-complexity`      | `[complexity] min_complexity`     |
| `--max-complexity`      | `[complexity] max_complexity`     |
| `--min-severity`        | `[dead_code] min_severity`        |
| `--clone-threshold`     | `[clones] similarity_threshold`   |
| `--min-cbo`             | `[cbo] min_cbo`                   |
| `--max-cycles`          | — (check command only)            |

## See also

- [Config File Format](format.md) — discovery and priority.
- [Examples](examples.md) — ready-to-use configurations.
