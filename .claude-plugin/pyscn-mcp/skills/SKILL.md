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
| "Which files should I review together?" | `analyze_code` (`analyses: ["communities"]`, `output_mode: "full"`) |
| "Map the module architecture" | `analyze_code` (`analyses: ["communities"]`, `output_mode: "full"`) |

## Common Parameters

- `path` (required): Path to Python file or directory
- `recursive` (analyze_code): Recursively analyze directories (default: true)
- `analyses` (analyze_code): Array of analyses to run - `complexity`, `dead_code`, `clone`, `cbo`, `lcom`, `deps`, `communities`
- `output_mode` (analyze_code): `summary` (default, health score + high-level metrics) or `full` (complete report, including the module `community_context_map`)

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

### Module Communities (context map for agents)
Module community detection is **opt-in**. Request it explicitly and use `output_mode: "full"`:
- `analyses: ["communities"]`, `output_mode: "full"`

The full response includes `community_analysis.community_context_map`, a compact, deterministic map telling you which modules to inspect together:
- `bundles[]`: clusters of modules that change together. Each has `modules`, `packages`, `risk_level`, the cluster's `bridge_modules`, an optional `suggested_review_scope` (path prefix to focus a review on), and a one-line `summary`. Large bundles cap the module list with a `... +N more` marker (`module_count` always holds the true total).
- `bridge_modules[]`: modules that couple two or more communities (`module`, the communities it `connects`, and a `reason`). Widen the review scope to include these when touching a bundle they connect.

Use the map to scope reviews and refactors: load all modules in a bundle together, and pull in connected bridge modules before changing cluster boundaries.

Always explain results and suggest improvements based on findings.
