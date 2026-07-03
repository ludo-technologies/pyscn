---
name: refactoring
description: Find refactoring targets in Python code using pyscn MCP tools - duplicate code (clones), overly complex functions, and dead code. Use when user asks about refactoring, code duplication, complexity hotspots, unreachable code, or cleaning up a codebase.
---

# Python Refactoring Analysis with pyscn MCP

Use the pyscn MCP tools to locate concrete refactoring targets.

## Tools

| Tool | Purpose |
|------|---------|
| `check_complexity` | Cyclomatic complexity of functions |
| `detect_clones` | Duplicate code detection (Type 1-4 clones) |
| `find_dead_code` | Unreachable code detection |
| `analyze_code` | Run several of the above at once via `analyses` |

## Tool Selection

| User Request | Tool |
|-------------|------|
| "Find complex functions" | `check_complexity` |
| "Find duplicate code" | `detect_clones` |
| "Find dead code" | `find_dead_code` |
| "What should I refactor first?" | `analyze_code` (`analyses: ["complexity", "clone", "dead_code"]`) |

## Parameters

Common:
- `path` (required): Path to Python file or directory

`check_complexity`:
- `min_complexity`: Minimum to report (default: 1)
- `max_complexity`: Maximum allowed (default: 0 = no limit)
- Risk levels: Low (<10), Medium (10-20), High (>20)

`detect_clones`:
- `similarity_threshold`: 0.0-1.0 (default: 0.8)
- `min_lines`: Minimum lines to consider (default: 5)

`find_dead_code`:
- Severity levels: Critical, Warning, Info. Critical means code after return/break/continue/raise that can never execute.

## Prioritizing Findings

1. Critical dead code: safe deletions, do these first.
2. High-complexity functions (>20): extract functions, flatten conditionals.
3. Clone groups spanning multiple files: extract shared helpers; clones within one file are usually quicker wins.

When suggesting a refactor, cite the specific function names, files, and line ranges from the results, and re-run the same tool afterward to confirm the improvement.
