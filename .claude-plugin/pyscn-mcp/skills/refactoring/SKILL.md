---
name: refactoring
description: Find refactoring targets in Python code using pyscn - duplicate code (clones), overly complex functions, and dead code. Use when user asks about refactoring, code duplication, complexity hotspots, unreachable code, or cleaning up a codebase.
---

# Python Refactoring Analysis with pyscn

Run the pyscn CLI to locate concrete refactoring targets. No install needed: `uvx pyscn@latest <command>`.

## Commands

| User Request | Command |
|-------------|---------|
| "Find complex functions" | `uvx pyscn@latest analyze --select complexity <path>` |
| "Find duplicate code" | `uvx pyscn@latest analyze --select clones <path>` |
| "Find dead code" | `uvx pyscn@latest analyze --select deadcode <path>` |
| "What should I refactor first?" | `uvx pyscn@latest analyze --select complexity,deadcode,clones <path>` |

## Key Flags

- `--min-complexity <n>`: minimum complexity to report (default: 5). Risk levels: Low (<10), Medium (10-20), High (>20)
- `--clone-threshold <0.0-1.0>`: minimum similarity for clone detection (default: 0.65)
- `--min-severity <critical|warning|info>`: dead code severity floor (default: warning). Critical means code after return/break/continue/raise that can never execute.
- `--json`: write a detailed report with line-level findings. **Report files are NOT written to stdout**; they go to `.pyscn/reports/` and the path is printed on completion.

## Prioritizing Findings

1. Critical dead code: safe deletions, do these first.
2. High-complexity functions (>20): extract functions, flatten conditionals.
3. Clone groups spanning multiple files: extract shared helpers; clones within one file are usually quicker wins.

When suggesting a refactor, cite the specific function names, files, and line ranges from the results, and re-run the same command afterward to confirm the improvement.

## MCP Alternative

If the `pyscn-mcp` MCP server is connected, you can use its tools instead of the CLI: `check_complexity` (`min_complexity`), `detect_clones` (`similarity_threshold`, `min_lines`), `find_dead_code`, or `analyze_code` with `analyses: ["complexity", "clone", "dead_code"]` to run several at once. Results are returned inline with no report files.
