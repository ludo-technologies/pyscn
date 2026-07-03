---
name: health-check
description: Get an overall Python code quality health score using pyscn. Use when user asks how healthy or good the code is, wants a quality overview, a grade, a summary of technical debt, or a before/after quality comparison.
---

# Python Code Health Check with pyscn

Run the pyscn CLI to give a quick, quantified picture of code quality. No install needed:

```bash
uvx pyscn@latest analyze <path>
```

The summary output includes a health score (0-100), a letter grade (A-F), and per-category metrics.

## Commands

| User Request | Command |
|-------------|---------|
| "How healthy is this code?" | `uvx pyscn@latest analyze <path>` |
| "Give me a quality overview" | Same command; walk through the category breakdown |
| "Did my refactoring improve quality?" | Run before and after, compare scores |

For machine-readable detail add `--json`. **Report files are NOT written to stdout.** They go to `.pyscn/reports/` under the current working directory and the path is printed on completion; read that file.

## Interpreting Results

- Score 0-100 with letter grade; category breakdowns cover complexity, dead code, duplication, coupling (CBO), cohesion (LCOM), dependencies, and architecture.
- Lead with the grade and the weakest categories, then name the top offenders (files/functions) driving them.
- For deeper follow-up, hand off to the focused skills: refactoring targets → `refactoring`, module structure → `architecture-review`.

Always explain the score in plain terms and suggest the highest-impact next step.

## MCP Alternative

If the `pyscn-mcp` MCP server is connected, you can use its tools instead of the CLI: `get_health_score` (score + grade) or `analyze_code` (`output_mode: "summary"` gives the health score plus high-level metrics per category). Results are returned inline with no report files.
