# FAQ

## General

### How is pyscn different from ruff, pylint, or mypy?

Ruff and pylint are linters; mypy is a type checker. pyscn is a structural analyzer: it builds control-flow graphs, tree representations, and dependency graphs to measure complexity, reachability, duplication, and coupling. They are complementary.

### What does pyscn not detect?

- Runtime bugs
- Security vulnerabilities
- Performance issues
- Style violations (use ruff)
- Type errors (use mypy / pyright)

### Does pyscn need network access?

No. pyscn runs fully locally. No telemetry, no remote calls.

### Does it work with async code?

Yes. `async def` and `await` are analyzed the same as sync code.

### Can I analyze Jupyter notebooks?

No. Convert with `jupyter nbconvert --to script` first.

## Configuration

### Where should my config file go?

Put `.pyscn.toml` at the repository root. pyscn discovers it by walking up from the analyzed files.

### I have `pyproject.toml`. Should I use `[tool.pyscn]` instead?

Either works. `.pyscn.toml` wins if both exist.

### How do I exclude generated code / migrations / vendored dependencies?

```toml
[analysis]
exclude_patterns = [
  "**/migrations/**",
  "**/__generated__/**",
  "vendor/**",
]
```

### Can I have different thresholds for different parts of my project?

Not in a single config file. Use per-directory configs:

```bash
pyscn check --config backend/.pyscn.toml backend/
pyscn check --config scripts/.pyscn.toml scripts/
```

## Running pyscn

### The HTML report didn't open my browser.

Auto-open is suppressed when stdin is not a TTY, over SSH, or `CI` is set. The report path is printed to stderr. Force with `--no-open`.

### `pyscn analyze` is slow on my repo.

- `--skip-clones` (clones are the slowest analyzer)
- Narrow scope: `pyscn analyze src/`
- Raise `min_lines` / `min_nodes` under `[clones]`
- Increase `max_goroutines` under `[clones]`

### I get parse errors on valid Python code.

pyscn uses tree-sitter, which supports Python through 3.13. File an issue with a minimal reproduction.

### Can pyscn auto-fix issues?

No. pyscn reports; it never modifies source.

## Scores and thresholds

### My Health Score dropped overnight.

Check category scores. Common causes:

- A new large function raised average complexity.
- A refactor left dead code.
- Copy-paste created clones.
- A new import introduced a cycle.

Diff JSON output between runs to pinpoint the change.

### Why is my Coupling score so low on a small codebase?

The penalty is percentage-based, so a 3/10 problematic ratio produces the same penalty as 300/1000. For projects with fewer than 20 classes, inspect raw CBO values rather than the category score.

### What's a "good" Health Score?

| Range | Meaning |
| --- | --- |
| 90+ | Great |
| 70–90 | Normal for healthy codebases |
| 50–70 | Real work needed; salvageable |
| < 50 | Focused refactoring required |

Trend matters more than the absolute number.

## MCP

### The assistant sees pyscn tools but calls fail.

Verify the binary is on PATH, or use `uvx pyscn-mcp`. Test directly:

```bash
uvx pyscn-mcp
```

See the [MCP guide](integrations/mcp.md).

### Can the MCP server refactor my code?

No. pyscn's MCP is read-only.

## Troubleshooting

### `pyscn: command not found` after installing with pip.

The install location isn't on PATH. Inspect with:

```bash
python -m pip show -f pyscn | grep bin/pyscn
```

On Linux/macOS, add `~/.local/bin` to PATH, or install with `uvx pyscn@latest <command>` (no install step), `uv tool install pyscn`, or `pipx install pyscn`.

### The report shows 0 files analyzed.

Include/exclude patterns excluded everything. Defaults: `include_patterns = ["**/*.py"]`; excludes include `test_*.py` and `*_test.py`. Override to analyze tests:

```toml
[analysis]
exclude_patterns = [
  "**/__pycache__/*",
  "**/*.pyc",
  ".venv/**",
]
```

### `Warning: parse error in <file>`.

The file has a syntax error tree-sitter couldn't recover from. The file is skipped; other files analyze normally.

## Getting help

- [GitHub Issues](https://github.com/ludo-technologies/pyscn/issues) — bugs and feature requests.
- [GitHub Discussions](https://github.com/ludo-technologies/pyscn/discussions) — questions and ideas.
- [Source code](https://github.com/ludo-technologies/pyscn).
