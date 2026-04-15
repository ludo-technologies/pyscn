---
hide:
  - navigation
  - toc
---

# pyscn

A structural static analyzer for Python. Detects dead code, duplication, complexity, and coupling issues via control-flow and tree analysis.

```bash
uvx pyscn@latest analyze .
```

## Features

- **32 rules** across unreachable code, duplicate code, complexity, class design, dependency injection, module structure, and mock data.
- **CFG-based reachability** finds dead code past `return` / `raise` / `break` / `continue` and unreachable branches.
- **APTED + LSH clone detection** across four clone types (identical, renamed, modified, semantic).
- **CBO / LCOM4** class coupling and cohesion metrics.
- **Circular import detection** via Tarjan's SCC.
- **Health score** (0–100) with per-category breakdown.
- **CI-ready** with `pyscn check`, linter-style output, and deterministic exit codes.
- **MCP server** (`pyscn-mcp`) for Claude Code, Cursor, and other MCP clients.

Written in Go. 100,000+ lines/sec on typical hardware. No Python runtime dependencies.

## Installation

```bash
uvx pyscn@latest <command>   # run without installing (recommended)
uv tool install pyscn        # install with uv
pipx install pyscn           # install with pipx
pip install pyscn            # install with pip
```

See [Installation](getting-started/installation.md) for all options.

## Quick start

```bash
pyscn analyze .                         # full analysis, HTML report
pyscn check --select complexity,deadcode src/   # CI gate
pyscn init                              # generate .pyscn.toml
```

See [Quick Start](getting-started/quick-start.md) and the [Rule catalog](rules/index.md).
