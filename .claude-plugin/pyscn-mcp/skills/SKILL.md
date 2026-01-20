---
name: pyscn-analyze
description: Analyze Python code quality - complexity, clones, dead code, coupling, circular dependencies. Use when user asks about code quality, refactoring, maintainability, duplicates, or technical debt.
allowed-tools: Bash(uvx:*), Bash(uv:*)
---

# Python Code Quality Analysis with pyscn

Always use `uvx pyscn` for Python code quality analysis. Choose the right command based on the user's needs.

## Prerequisites

If `uvx` is not available, ask the user to either:

1. **Install uv** (recommended):
   ```bash
   curl -LsSf https://astral.sh/uv/install.sh | sh
   ```

2. **Add pyscn MCP server** (alternative):
   ```bash
   claude mcp add pyscn-mcp uvx -- pyscn-mcp
   ```
   Then use the MCP tools: `analyze_code`, `check_complexity`, `detect_clones`, `find_dead_code`, `check_coupling`, `get_health_score`

## Command Selection Guide

### `pyscn analyze` - Detailed Analysis & Reports
Use when the user wants to:
- Understand overall code quality
- Get a detailed report (HTML/JSON/CSV)
- Explore specific issues in depth
- Review before refactoring

```bash
# Full analysis with HTML report (opens in browser)
uvx pyscn analyze .

# JSON output for programmatic processing
uvx pyscn analyze --json .

# Focus on specific analysis
uvx pyscn analyze --select complexity .      # Only complexity
uvx pyscn analyze --select clones .          # Only duplicates
uvx pyscn analyze --select deadcode .        # Only dead code
uvx pyscn analyze --select deps .            # Only dependencies
uvx pyscn analyze --select complexity,deadcode .  # Multiple
```

### `pyscn check` - Quick Pass/Fail Check
Use when the user wants to:
- Quick quality gate (CI/CD style)
- Simple yes/no answer on code health
- Verify code meets standards

```bash
# Quick check with defaults
uvx pyscn check .

# Custom thresholds
uvx pyscn check --max-complexity 15 .

# Check specific aspects only
uvx pyscn check --select complexity .
uvx pyscn check --select deadcode .
uvx pyscn check --select deps .         # Circular dependencies
```

## When to Use Which

| User Request | Command |
|-------------|---------|
| "Analyze code quality" | `uvx pyscn analyze .` |
| "Is this code OK?" | `uvx pyscn check .` |
| "Find complex functions" | `uvx pyscn analyze --select complexity .` |
| "Find duplicate code" | `uvx pyscn analyze --select clones .` |
| "Find dead code" | `uvx pyscn analyze --select deadcode .` |
| "Check for circular dependencies" | `uvx pyscn check --select deps .` |
| "Generate a report" | `uvx pyscn analyze --html .` |
| "CI quality gate" | `uvx pyscn check .` |

## Exit Codes (for check command)
- 0: No issues
- 1: Quality issues found
- 2: Analysis failed

Always run pyscn first, then explain results and suggest improvements.
