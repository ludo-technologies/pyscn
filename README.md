# pyscn - An Intelligent Python Code Quality Analyzer

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://img.shields.io/badge/CI-Passing-brightgreen.svg)](https://github.com/ludo-technologies/pyscn/actions)

An intelligent Python code quality analyzer that performs deep, structural static analysis to help you write cleaner, more maintainable code.

pyscn complements traditional linters with analyses based on control-flow graphs and tree edit distance:

- Cyclomatic complexity: precise CFG-based metrics with risk thresholds and sorting
- Dead code detection: unreachable code after return/break/continue/raise and unreachable branches
- Code clone detection: APTED-based structural similarity (Type 1–4) with grouping and thresholds
- Coupling Between Objects (CBO): class dependency metrics and risk levels

All analyses are available as dedicated commands and via a unified analyze command that can generate HTML/JSON/YAML/CSV reports.

## Quick Start

```bash
# Install via pip or uv
pip install pyscn
# or: uv add pyscn

# Fast quality check (CI-friendly)
pyscn check .

# Comprehensive analysis with unified report
pyscn analyze .
```

## Features

- Unified analyze workflow: run complexity, dead code, clone, and CBO together with a single command
- Multiple formats: text, JSON, YAML, CSV, and HTML (auto-open supported)
- Config discovery: Ruff-style hierarchical search and project-scoped defaults
- Timestamped reports: output files named with timestamps; defaults to `.pyscn/reports/` under the current working directory unless overridden by configuration
- Clean Architecture: maintainable Go code with domain/use‑case separation and comprehensive tests

## Performance

- Analyzes thousands of lines per second with parallel processing

## Commands Overview

Run `pyscn --help` or `pyscn <command> --help` for all options. Root `--verbose` is supported.

### analyze
Run all major analyses concurrently and generate a unified report.

Default behavior (no format flags):
- Generates a unified HTML report under `.pyscn/reports/`
- Prints a concise analysis summary to stderr
- Auto-opens the HTML only in interactive (non‑CI) sessions; use `--no-open` to suppress

```bash
# Unified HTML report (explicit)
pyscn analyze --html src/

# Default (no flags) also generates HTML under `.pyscn/reports/`
pyscn analyze src/

# JSON/YAML/CSV are also supported (auto-generates timestamped files)
pyscn analyze --json src/  # Creates: .pyscn/reports/analyze_YYYYMMDD_HHMMSS.json (unless [output].directory overrides)

# Skip specific analyses or tune thresholds
pyscn analyze --skip-clones --min-complexity 10 --min-severity critical --min-cbo 5 src/

# Run only specific analyses (preferred over individual commands)
pyscn analyze --select complexity,deadcode src/  # Only run complexity and dead code detection
pyscn analyze --select cbo src/                  # Only run CBO analysis
```

The unified report summarizes files analyzed, average complexity, high-complexity count, dead code findings, clone statistics (including duplication percentage), and CBO metrics, plus a health score.


### check
Fast CI‑friendly gate with sensible defaults.

```bash
pyscn check .
pyscn check --max-complexity 15 --skip-clones src/
pyscn check --allow-dead-code src/
```

Exit codes: 0 (ok), 1 (quality issues), 2 (analysis error). Prints concise findings suitable for CI logs.

### init
Generate a starter `.pyscn.toml` with comprehensive, documented options.

```bash
pyscn init            # create .pyscn.toml in current directory
pyscn init --config myconfig.toml
pyscn init --force    # overwrite if exists
```

## Configuration

pyscn supports hierarchical configuration discovery:

1. `pyproject.toml` with `[tool.pyscn]` section (recommended)
2. `.pyscn.toml` (dedicated config file)
3. Target directory upward to filesystem root (nearest file wins)

Supported formats: TOML only (unified configuration strategy like ruff).

Example:

```toml
# .pyscn.toml or pyproject.toml with [tool.pyscn] section
[output]
directory = "build/reports"    # Override default 'reports/' if desired
sort_by = "name"               # name | complexity | risk  
min_complexity = 1

[complexity]
enabled = true
low_threshold = 9
medium_threshold = 19
max_complexity = 0             # 0 = no limit
report_unchanged = true

[dead_code]
enabled = true
min_severity = "warning"       # critical | warning | info
show_context = false
context_lines = 3
sort_by = "severity"           # severity | line | file | function
detect_after_return = true
detect_after_break = true
detect_after_continue = true
detect_after_raise = true
detect_unreachable_branches = true

[clones]                       # unified clone settings (see CLI for full options)
min_lines = 5
min_nodes = 5
similarity_threshold = 0.8

[cbo]                          # CBO settings (maps to cbo command flags)
enabled = true
low_threshold = 5
medium_threshold = 10
include_builtins = false
```

Notes:

- Default include patterns: `*.py` (`.pyi` stub files supported but not included by default); default exclude: `test_*.py`, `*_test.py`
- **Output formats**:
  - `text` (default): Prints to stdout (for individual commands)
  - `analyze` default (no flags): Generates HTML + prints stderr summary (hybrid); auto-open only in interactive (non‑CI) sessions; use `--no-open` to suppress
  - `json`, `yaml`, `csv`, `html`: Auto-generate timestamped files when selected
  - Default location: `.pyscn/reports/` under the current working directory (created if missing); override with `[output].directory`
  - All commands generate timestamped output files for structured formats

## Installation

### Install via pip or uv (recommended for Python users)

```bash
# Using pip
pip install pyscn

# Using uv (faster, modern Python package manager)
uv add pyscn
```

> Note: pyscn is available on PyPI and GitHub Releases.

If you prefer to build wheels locally (e.g., for development), see the Python section below.

### Build from source (Go)

```bash
git clone https://github.com/ludo-technologies/pyscn.git
cd pyscn
make build     # or: go build -o pyscn ./cmd/pyscn

# Install globally
go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest
```

### Python wheel (optional)

This repo includes a Python wrapper that bundles platform binaries for `pyscn`. If you’re packaging or testing wheels locally:

```bash
# Build wheel for current platform
make python-wheel
pip install dist/*.whl

# Run
pyscn version
```

For all-platform wheels and cross-compiling, see `python/scripts/build_all_wheels.sh`.

## CI/CD Example

```yaml
# .github/workflows/code-quality.yml
name: Code Quality
on: [push, pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
      - run: go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest
      - name: Unified analysis (JSON)
        run: pyscn analyze --json src/ > analyze.json
      - name: Enforce thresholds
        run: pyscn check .
```

## Development

- Go 1.24+, Make, Git
- Commands: `make build`, `make test`, `make coverage`, `make bench`, `make clean`, `make dev`
- See docs for details: [DEVELOPMENT.md](docs/DEVELOPMENT.md), [TESTING.md](docs/TESTING.md), [ARCHITECTURE.md](docs/ARCHITECTURE.md)

## License

MIT License — see [LICENSE](LICENSE).

## Version

Run `pyscn version` for build/version details (commit, date, platform). The repository may be in active development; prefer the binary’s version output over hardcoded README text.

## Acknowledgments

- Tree-sitter for robust parsing
- Go community for tooling and libraries
- Research and open-source work on static analysis and tree edit distance

— Built with ❤️ by the pyscn team
