# pyscn - Python Code Quality Analyzer

[![Article](https://img.shields.io/badge/dev.to-Article-0A0A0A?style=flat-square&logo=dev.to)](https://dev.to/daisukeyoda/pyscn-the-code-quality-analyzer-for-vibe-coders-18hk)
[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![CI](https://img.shields.io/badge/CI-Passing-brightgreen.svg)](https://github.com/ludo-technologies/pyscn/actions)

## pyscn is a code quality analyzer for Python vibe coders.

Building with Cursor, Claude, or ChatGPT? pyscn performs structural analysis to keep your codebase maintainable.

## Quick Start

```bash
# Run analysis without installation
uvx pyscn analyze .
# or
pipx run pyscn analyze .
```

## Demo

https://github.com/user-attachments/assets/07f48070-c0dd-437b-9621-cb3963f863ff

## Features

- 🔍 **CFG-based dead code detection** – Find unreachable code after exhaustive if-elif-else chains
- 📋 **Clone detection with APTED + LSH** – Identify refactoring opportunities with tree edit distance
- 🔗 **Coupling metrics (CBO)** – Track architecture quality and module dependencies
- 📊 **Cyclomatic complexity analysis** – Spot functions that need breaking down

**100,000+ lines/sec** • Built with Go + tree-sitter

## MCP Integration

Run pyscn analyses straight from AI coding assistants via the Model Context Protocol (MCP). The bundled `pyscn-mcp` server exposes the same tools used in the CLI to Claude Code, Cursor, ChatGPT, and other MCP clients.

### Claude Code Setup

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

### Cursor / Claude Desktop Setup

Add to your MCP settings (`~/.config/claude-desktop/config.json` or Cursor settings):

```json
{
  "mcpServers": {
    "pyscn-mcp": {
      "command": "uvx",
      "args": ["pyscn-mcp"],
      "env": {
        "PYSCN_CONFIG": "/path/to/.pyscn.toml"
      }
    }
  }
}
```

The instructions like "Analyze the code quality" trigger pyscn via MCP.

Dive deeper in `mcp/README.md` for setup walkthroughs and `docs/MCP_INTEGRATION.md` for architecture details.

## Installation

```bash
# Install with pipx (recommended)
pipx install pyscn

# Or with uv
uv tool install pyscn
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
pyscn check .                         # Quick pass/fail check
pyscn check --max-complexity 15 .     # Custom thresholds
pyscn check --select deps .           # Check only circular dependencies

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

## CI/CD Integration

```yaml
# GitHub Actions
- uses: actions/checkout@v4
- run: pipx run pyscn check .    # Fail on quality issues

# Pre-commit hook
- repo: local
  hooks:
    - id: pyscn
      name: pyscn check
      entry: pipx run pyscn check .
      language: system
      pass_filenames: false
      types: [python]
```

---

## Documentation

📚 **[Development Guide](docs/DEVELOPMENT.md)** • **[Architecture](docs/ARCHITECTURE.md)** • **[Testing](docs/TESTING.md)**

## Enterprise Support

For commercial support, custom integrations, or consulting services, contact us at contact@ludo-tech.org

## License

MIT License — see [LICENSE](LICENSE)

---

*Built with ❤️ using Go and tree-sitter*
