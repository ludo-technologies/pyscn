# `pyscn check`

Quality gate for CI/CD pipelines. Writes linter-style findings to **stderr** and exits non-zero if any issue fails a threshold.

```text
pyscn check [flags] [paths...]
```

Paths default to the current directory.

## What it does

`check` is the CI companion to [`analyze`](analyze.md):

- **Findings go to stderr** in linter format (`file:line:col: message`).
- **Exit 0** on pass, **exit 1** on issues found, **exit 2** when the analysis could not complete.
- **Strict defaults** — any function over complexity 10 fails; any circular dependency fails (when `--select deps` is set); any file that cannot be parsed fails.
- **Fast** — only runs the analyses you select; skips report generation.

!!! note "Parse-error coverage"
    Every selected analysis uses the same project discovery ledger. A file that cannot be read or parsed fails the check, including `pyscn check --select deps`, unless `--allow-parse-errors` is set.

## Flags

### Analysis selection

| Flag | Description |
| --- | --- |
| `-s, --select <list>` | Run only the listed analyses. Values: `complexity`, `deadcode`, `clones`, `deps` (alias `circular`), `mockdata`, `di`. |
| `--skip-clones`       | Don't run clone detection. |

Default (no `--select`): runs `complexity`, `deadcode`, **and `clones`**. `deps`, `mockdata`, and `di` are opt-in via `--select`. Pass `--skip-clones` to skip clone detection without switching to `--select`.

### Threshold overrides

| Flag | Default | Description |
| --- | --- | --- |
| `--max-complexity <N>`   | `10` | Fail if any module, executable class suite, or function exceeds this cyclomatic complexity. |
| `--max-cycles <N>`       | `0`  | Maximum number of circular dependency cycles before failing. |
| `--allow-dead-code`      | off  | Treat dead code as warnings only; don't fail the check. |
| `--allow-circular-deps`  | off  | Treat cycles as warnings only; don't fail the check. |
| `--allow-parse-errors`   | off  | Treat unreadable or unparseable files as warnings only; don't fail the check. |

By default, a file that cannot be parsed fails the check. Such a file is excluded from every analysis, so it contributes no findings and would otherwise pass every threshold — a syntax error in your source would be reported as a clean run.

### Output

| Flag | Description |
| --- | --- |
| `-q, --quiet`          | Suppress output unless issues are found. |
| `-c, --config <path>`  | Load configuration from a specific file. |
| `-v, --verbose`        | Print detailed progress. |

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | All checks passed. |
| `1` | One or more quality thresholds were exceeded. |
| `2` | The analysis could not complete over the requested targets — invalid input, missing files, or files that could not be parsed. |

Exit `1` is a verdict about your code; exit `2` means the verdict itself is incomplete and should not be read as a pass.

## Examples

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

## Relationship to `analyze`

`check` uses the same analyzers and the same configuration file as `analyze`. The differences:

| Aspect | `analyze` | `check` |
| --- | --- | --- |
| Output | Report file (HTML/JSON/YAML/CSV) | Linter-style stderr |
| Exit on issues | Always `0` (unless error) | Exit `1` if any issue fails threshold |
| Clone detection | On by default | On by default (skip with `--skip-clones`) |
| Dependency analysis | On by default | Off by default (opt-in via `--select deps`) |
| Speed | Slower (all analyzers, report generation) | Fast (only selected, no report) |
| Use case | Interactive review | CI quality gate |

Use both: `analyze` to understand problems, `check` to prevent regressions.

## See also

- [CI/CD Integration](../integrations/ci-cd.md) — GitHub Actions / pre-commit / GitLab examples.
- [`pyscn analyze`](analyze.md) — Full analysis with reports.
