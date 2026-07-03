---
name: cli-analysis
description: Run the pyscn command-line tool for Python code quality analysis - CI/CD quality gates, HTML/JSON/CSV reports, and full analysis runs. Use when the pyscn MCP tools are unavailable, or when user wants a CI check, a shareable report file, or to configure pyscn for a project.
---

# Python Code Quality Analysis with the pyscn CLI

Use the `pyscn` command-line tool when MCP tools are not connected or when the task needs report files, CI integration, or project configuration.

Install if missing: `pip install pyscn` (or `uvx pyscn <command>` without installing).

## Commands

| Command | Purpose |
|---------|---------|
| `pyscn analyze <path>` | Comprehensive analysis: complexity, dead code, clones, coupling (CBO), cohesion (LCOM), dependencies, architecture |
| `pyscn check <path>` | Fast pass/fail quality gate for CI with predefined thresholds |
| `pyscn init` | Generate a `.pyscn.toml` config file |

## analyze — full analysis and reports

```bash
pyscn analyze .                          # human-readable summary + health score
pyscn analyze --json src/                # also write a JSON report file
pyscn analyze --html src/                # interactive HTML report
pyscn analyze --select complexity,deadcode src/   # only specific analyses
pyscn analyze --skip-clones src/         # clones are the slowest analysis
pyscn analyze --min-complexity 10 --min-severity critical src/
```

Key flags:
- `--json` / `--html` / `--csv` / `--yaml`: generate a report file
- `--select`: `complexity,deadcode,clones,cbo,lcom,deps,communities`
- `--skip-clones`, `--skip-cbo`, `--skip-deps`, ...: skip individual analyses
- `--clone-threshold` (default 0.65), `--min-complexity` (default 5), `--min-cbo`
- `--no-open`: don't auto-open HTML reports in a browser (use in scripts)

**Report files are NOT written to stdout.** They go to `.pyscn/reports/` under the current working directory with timestamped names like `analyze_20260702_221139.json`, and the path is printed on completion. To consume results programmatically: run with `--json`, capture the printed path (or take the newest file in `.pyscn/reports/`), then read that file.

## check — CI quality gate

```bash
pyscn check .                                        # complexity + deadcode + clones
pyscn check --select complexity --max-complexity 10 src/
pyscn check --select deps src/                       # fail on circular dependencies
pyscn check --allow-dead-code --skip-clones src/
pyscn check -q .                                     # quiet: output only on issues
```

Exit codes: `0` no issues, `1` quality issues found, `2` analysis failed (invalid input, missing files). Default gates: complexity > 10 fails, critical dead code fails, clones warn only, dependency cycles fail.

`--select` accepts: `complexity`, `deadcode`, `clones`, `deps`, `mockdata`, `di` (dependency-injection anti-patterns).

## Configuration

Settings load in priority order: command-line flags > `.pyscn.toml` > `[tool.pyscn]` in `pyproject.toml` > defaults. Run `pyscn init` to scaffold `.pyscn.toml` when a project wants persistent thresholds.

## Reporting Results

Summarize the health score and grade, list the specific functions/files behind each failing category, and suggest fixes. For CI setup, recommend `pyscn check` in the pipeline and `pyscn analyze --html` for periodic deep reviews.
