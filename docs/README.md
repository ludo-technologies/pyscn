# Developer documentation

This directory contains documentation for pyscn **contributors** and **algorithm
designers**. It is intentionally scoped to internals: architecture decisions,
scoring formulas, algorithm details, coding standards, and contribution
conventions.

**If you're looking for end-user docs** (installation, CLI usage, rule
reference, configuration), see:

- <https://ludo-technologies.github.io/pyscn/>
- Source: [`website/`](../website/) in this repository

## Contents

| File / Directory | Purpose |
| --- | --- |
| [`ARCHITECTURE.md`](ARCHITECTURE.md)           | High-level system architecture and package layout. |
| [`DEVELOPMENT.md`](DEVELOPMENT.md)             | Local setup, build targets, debugging. |
| [`TESTING.md`](TESTING.md)                     | Test strategy, fixtures, integration tests. |
| [`CODING_STANDARDS.md`](CODING_STANDARDS.md)   | Go style conventions for this repo. |
| [`COMMIT_CONVENTION.md`](COMMIT_CONVENTION.md) | Commit message format. |
| [`BRANCHING.md`](BRANCHING.md)                 | Branching and release flow. |
| [`ANALYZE_SCORING.md`](ANALYZE_SCORING.md)     | Health Score design rationale. See also the user-facing spec page. |
| [`MCP_INTEGRATION.md`](MCP_INTEGRATION.md)     | MCP server design. |
| [`algorithms/`](algorithms/)                   | Per-analyzer algorithm descriptions (CFG, APTED, LSH, Main Sequence, etc.). |

For the user-facing **rule catalog** (problem-based rule names and fixes), see
<https://ludo-technologies.github.io/pyscn/rules/>.
