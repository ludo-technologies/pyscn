---
name: pyscn-mcp
description: Analyze Python code quality using MCP tools - complexity, clones, dead code, coupling. Use when user asks about code quality, refactoring, maintainability, duplicates, or technical debt.
---

# Python Code Quality Analysis with pyscn MCP

Use the pyscn MCP tools for Python code quality analysis.

## Available Tools

| Tool | Purpose |
|------|---------|
| `get_health_score` | Overall code health score (0-100) with grade |
| `analyze_code` | Comprehensive analysis (complexity, dead code, clones, coupling, deps) |
| `check_complexity` | Cyclomatic complexity of functions |
| `detect_clones` | Duplicate code detection |
| `find_dead_code` | Unreachable code detection |
| `check_coupling` | Class coupling (CBO) metrics |

## Tool Selection Guide

| User Request | Tool |
|-------------|------|
| "How healthy is this code?" | `get_health_score` |
| "Analyze code quality" | `analyze_code` |
| "Find complex functions" | `check_complexity` |
| "Find duplicate code" | `detect_clones` |
| "Find dead code" | `find_dead_code` |
| "Check class coupling" | `check_coupling` |

## Common Parameters

- `path` (required): Path to Python file or directory
- `recursive` (analyze_code): Recursively analyze directories (default: true)
- `analyses` (analyze_code): Array of analyses to run - `complexity`, `dead_code`, `clone`, `cbo`, `deps`

## Examples

### Quick Health Check
Use `get_health_score` for a quick overview:
- Returns score 0-100 with letter grade (A-F)
- Category breakdowns for maintainability, reliability, etc.

### Detailed Analysis
Use `analyze_code` with specific analyses:
- `analyses: ["complexity"]` - Only complexity
- `analyses: ["clone"]` - Only duplicates
- `analyses: ["dead_code"]` - Only dead code
- `analyses: ["complexity", "dead_code"]` - Multiple

### Complexity Thresholds
Use `check_complexity` with:
- `min_complexity`: Minimum to report (default: 1)
- `max_complexity`: Maximum allowed (default: 0 = no limit)

### Clone Detection
Use `detect_clones` with:
- `similarity_threshold`: 0.0-1.0 (default: 0.8)
- `min_lines`: Minimum lines to consider (default: 5)

Always explain results and suggest improvements based on findings.
