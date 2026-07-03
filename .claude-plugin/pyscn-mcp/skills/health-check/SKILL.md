---
name: health-check
description: Get an overall Python code quality health score using pyscn MCP tools. Use when user asks how healthy or good the code is, wants a quality overview, a grade, a summary of technical debt, or a before/after quality comparison.
---

# Python Code Health Check with pyscn MCP

Use the pyscn MCP tools to give a quick, quantified picture of code quality.

## Tools

| Tool | Purpose |
|------|---------|
| `get_health_score` | Overall code health score (0-100) with letter grade (A-F) |
| `analyze_code` | Comprehensive analysis; `output_mode: "summary"` gives health score plus high-level metrics per category |

## Tool Selection

| User Request | Tool |
|-------------|------|
| "How healthy is this code?" | `get_health_score` |
| "Give me a quality overview" | `analyze_code` (defaults) |
| "Did my refactoring improve quality?" | `get_health_score` before and after, compare |

## Parameters

- `path` (required): Path to Python file or directory
- `recursive` (analyze_code): Recursively analyze directories (default: true)
- `output_mode` (analyze_code): `summary` (default) or `full`

## Interpreting Results

- Score 0-100 with letter grade; category breakdowns cover complexity, dead code, duplication, coupling (CBO), cohesion (LCOM), dependencies, and architecture.
- Lead with the grade and the weakest categories, then name the top offenders (files/functions) driving them.
- For deeper follow-up, hand off to the focused skills: refactoring targets → `refactoring`, module structure → `architecture-review`.

Always explain the score in plain terms and suggest the highest-impact next step.
