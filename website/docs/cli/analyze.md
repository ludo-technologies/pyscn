# `pyscn analyze`

Run all available analyses on Python files and produce a report.

```text
pyscn analyze [flags] <paths...>
```

`<paths...>` is one or more files or directories. Directories are traversed recursively using the `include_patterns` and `exclude_patterns` from your config.

## What it does

By default `analyze` runs every enabled analyzer concurrently:

- Cyclomatic complexity
- Dead code detection
- Clone detection (Type 1–4)
- Class coupling (CBO)
- Class cohesion (LCOM4)
- Module dependencies
- Architecture layer validation
- Module community detection

Results are combined into a single report with a [Health Score](../output/health-score.md).

## Flags

### Output format

Only one of these may be set per invocation. If none is set, HTML is generated.

| Flag        | Description |
| ----------- | --- |
| `--html`    | Generate HTML report (default). |
| `--json`    | Generate JSON report. |
| `--yaml`    | Generate YAML report. |
| `--csv`     | Generate CSV summary (metrics only, no per-finding detail). |
| `--no-open` | Do not open the HTML report in a browser. |

Output files land in `.pyscn/reports/` by default, named `analyze_YYYYMMDD_HHMMSS.{ext}`. Configure the directory with `[output] directory = "..."`.

### Analysis selection

| Flag | Description |
| --- | --- |
| `--select <list>` | Only run the listed analyses. Comma-separated: `complexity,deadcode,clones,cbo,lcom,deps,communities`. |
| `--skip-complexity` | Skip complexity analysis. |
| `--skip-deadcode`   | Skip dead code detection. |
| `--skip-clones`     | Skip clone detection (the slowest analysis). |
| `--skip-cbo`        | Skip class coupling analysis. |
| `--skip-lcom`       | Skip class cohesion analysis. |
| `--skip-deps`       | Skip module dependency analysis. |
| `--skip-communities` | Skip module community detection. |

`--select` and `--skip-*` can be combined; selection applies first, then skips.

### Quick threshold overrides

| Flag | Default | Description |
| --- | --- | --- |
| `--min-complexity <N>`    | `5`        | Only report functions with complexity ≥ N. |
| `--min-severity <level>`  | `warning`  | Dead-code minimum severity: `info`, `warning`, `critical`. |
| `--clone-threshold <F>`   | `0.65`     | Minimum similarity (0.0–1.0) for clone detection. |
| `--min-cbo <N>`           | `0`        | Only report classes with CBO ≥ N. |

### Configuration

| Flag | Description |
| --- | --- |
| `-c, --config <path>` | Load configuration from a specific file instead of discovering `.pyscn.toml` / `pyproject.toml`. |
| `-v, --verbose`        | Print detailed progress and per-file logs. |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Analysis completed. Issues found in the report do not affect the exit code. |
| `1` | Analysis failed — invalid arguments, unreadable files, parse errors. |

`analyze` never fails the process based on findings; use [`pyscn check`](check.md) for pass/fail semantics.

## Examples

```bash
# Full analysis of the current directory with HTML report
pyscn analyze .

# JSON for pipelines
pyscn analyze --json src/

# Skip the slowest analyzer
pyscn analyze --skip-clones src/

# Only complexity and dead code
pyscn analyze --select complexity,deadcode src/

# Module community detection (standalone JSON)
pyscn analyze --json --select communities src/

# Stricter thresholds
pyscn analyze --min-complexity 10 --min-severity critical src/

# Use a specific config file
pyscn analyze --config ./configs/strict.toml src/

# Don't open the browser (useful in sandboxes or containers)
pyscn analyze --no-open .
```

## When to use `analyze` vs `check`

| Use case | Command |
| --- | --- |
| Local dev, "give me the full picture" | `pyscn analyze` |
| CI quality gate with pass/fail | [`pyscn check`](check.md) |
| Machine-readable output for custom tooling | `pyscn analyze --json` |

## See also

- [Configuration Reference](../configuration/reference.md) — every knob.
- [Module Community Detection](../guides/module-community-detection.md) — interpret communities, bridge modules, and modularity.
- [Health Score](../output/health-score.md) — how the 0–100 number is computed.
- [Output Schemas](../output/schemas.md) — JSON / YAML / CSV field definitions.
