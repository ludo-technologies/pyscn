<div align="center">

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [Français](README.fr.md)

<br>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="assets/logo-light.svg">
  <img alt="pyscn" src="assets/logo-light.svg" width="320">
</picture>

**A code quality analyzer for Python vibe coders.**

Building with Cursor, Claude, or ChatGPT? pyscn performs structural analysis to keep your codebase maintainable.

[![Article](https://img.shields.io/badge/dev.to-Article-0A0A0A?style=flat-square&logo=dev.to)](https://dev.to/daisukeyoda/pyscn-the-code-quality-analyzer-for-vibe-coders-18hk)
[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Downloads](https://img.shields.io/pypi/dm/pyscn?style=flat-square&logo=pypi&label=downloads)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/github/go-mod/go-version/ludo-technologies/pyscn?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/ludo-technologies/pyscn?style=flat-square)](LICENSE)

*Working with JavaScript/TypeScript? Check out [jscan](https://github.com/ludo-technologies/jscan)*

</div>

## Quick Start

```bash
# Run analysis without installation
uvx pyscn@latest analyze .
# or
pipx run pyscn analyze .
```

## Demo

https://github.com/user-attachments/assets/71d7a126-9c5e-4254-99f4-f2cdedd526ad

## Features

One command scores your whole codebase (0-100 with an A-F grade) and generates an HTML report that shows what to fix first.

pyscn looks at your code from five angles:

- 🧹 **Dead code** - unreachable code you can safely delete
- 📋 **Duplicate code** - copy-pasted and structurally similar code worth merging (Type 1-4 clone detection)
- 🌀 **Complexity** - functions that are hard to read and test (cyclomatic and cognitive complexity)
- 🏗️ **Architecture** - circular imports, layer rule violations (clean / layered / hexagonal / MVC presets), and auto-detected module communities that reveal how your code is actually structured
- 🧩 **Class design** - classes that do too much or depend on too much (CBO coupling, LCOM4 cohesion, DI anti-patterns)

**100,000+ lines/sec** • Built with Go + tree-sitter

## AI Agent Integration

pyscn ships Agent Skills that teach AI coding agents when and how to run each analysis: health checks, refactoring, architecture review, and CI-friendly reports.

### Agent Skills (Recommended)

```bash
uvx add-skills ludo-technologies/pyscn
```

This installs the Skills into your project. They work with Claude Code, Cursor, Codex, Gemini CLI, and [many other agents](https://github.com/ludo-technologies/add-skills) (add `--agent cursor` etc. to target one, `--global` for all projects).

Then just ask your agent:

1. "Analyze the code quality of the app/ directory"

2. "Find duplicate code and help me refactor it"

3. "Show me complex code and help me simplify it"

### MCP Server (Optional)

For tighter integration, the bundled `pyscn-mcp` server exposes the same analyses as MCP tools to Claude Code, Cursor, ChatGPT, and other MCP clients.

**Claude Code plugin (sets up the MCP server and the Skills together):**

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

**Manual setup for Claude Code:**

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

**Cursor / Claude Desktop:** add to your MCP settings (`~/.config/claude-desktop/config.json` or Cursor settings):

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
pyscn analyze --skip-communities .           # Skip module community detection
```

### `pyscn check`
Fast CI-friendly quality gate
```bash
pyscn check .                         # Quick pass/fail check
pyscn check --max-complexity 15 .     # Custom thresholds
pyscn check --max-cycles 0 .          # Only allow 0 cycle dependency
pyscn check --select deps .           # Check only for circular dependencies
pyscn check --allow-circular-deps .   # Allow circular dependencies (warning only)
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

## Pyscn Bot (GitHub App)

[Pyscn Bot](https://github.com/marketplace/pyscn-bot) monitors your Python code quality automatically.

### Features

- **PR Code Review** - Automatic code review on every pull request
- **Weekly Code Audit** - Scans your entire repository and creates issues for architectural problems

---

## Documentation

📖 **[pyscn documentation site](https://ludo-technologies.github.io/pyscn/)** — installation, rule catalog, CLI reference, configuration, output specification

For contributors: **[Development Guide](docs/DEVELOPMENT.md)** • **[Architecture](docs/ARCHITECTURE.md)** • **[Testing](docs/TESTING.md)**

## Enterprise Support

For commercial support, custom integrations, or consulting services, contact us at contact@ludo-tech.org

## License

MIT License — see [LICENSE](LICENSE)

---

*Built with ❤️ using Go and tree-sitter*
