# pyscn - Python Code Quality Analyzer

[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://img.shields.io/badge/CI-Passing-brightgreen.svg)](https://github.com/ludo-technologies/pyscn/actions)

**pyscn** is an intelligent Python code quality analyzer that performs deep structural analysis to help you write cleaner, more maintainable code.

## What Can It Do For You?

🔍 **Find Complex Code** - Identify functions that are hard to test and maintain
🧹 **Remove Dead Code** - Detect unreachable code that clutters your codebase
📋 **Detect Duplicates** - Find copy-pasted code for refactoring opportunities
🔗 **Analyze Dependencies** - Understand module relationships and coupling
📊 **Validate Architecture** - Ensure your code follows architectural patterns
📈 **Generate Reports** - Export findings in HTML, JSON, YAML, or CSV

## Quick Start

```bash
# 1. Install
pip install pyscn

# 2. Run analysis
pyscn analyze .
```

## Common Commands

### `pyscn analyze`
Run comprehensive analysis with HTML report
```bash
pyscn analyze .                              # All analyses with HTML report
pyscn analyze --json .                       # Generate JSON report
pyscn analyze --select complexity .          # Only complexity analysis
pyscn analyze --select deps .                # Only dependency analysis
pyscn analyze --select complexity,deps,deadcode . # Multiple analyses
```

### `pyscn check`
Fast CI-friendly quality gate
```bash
pyscn check .                      # Quick pass/fail check
pyscn check --max-complexity 15 .  # Custom thresholds
```

### `pyscn init`
Create configuration file
```bash
pyscn init                         # Generate .pyscn.toml
```

> 💡 Run `pyscn --help` or `pyscn <command> --help` for complete options

## Configuration

Create a `.pyscn.toml` file or add `[tool.pyscn]` to your `pyproject.toml`:

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "warning"

[output]
directory = "reports"
```

> ⚙️ Run `pyscn init` to generate a full configuration file with all available options

## Installation

### Recommended: pip or uv

```bash
pip install pyscn
# or
uv add pyscn
```

<details>
<summary>Alternative installation methods</summary>

### Build from source
```bash
git clone https://github.com/ludo-technologies/pyscn.git
cd pyscn
make build
```

### Go install
```bash
go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest
```

</details>

## CI/CD Integration

```yaml
# .github/workflows/code-quality.yml
name: Code Quality
on: [push, pull_request]

jobs:
  quality-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: pip install pyscn
      - name: Quick quality check
        run: pyscn check .
      - name: Generate detailed report
        run: pyscn analyze --json --select complexity,deadcode,deps src/
      - name: Upload report
        uses: actions/upload-artifact@v4
        with:
          name: code-quality-report
          path: .pyscn/reports/
```

---

## Documentation

📚 **[Development Guide](docs/DEVELOPMENT.md)** • **[Architecture](docs/ARCHITECTURE.md)** • **[Testing](docs/TESTING.md)**

## License

MIT License — see [LICENSE](LICENSE)

---

*Built with ❤️ using Go and tree-sitter*
