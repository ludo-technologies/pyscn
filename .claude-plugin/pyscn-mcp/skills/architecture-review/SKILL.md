---
name: architecture-review
description: Analyze Python module architecture using pyscn - class coupling (CBO), cohesion (LCOM), dependency cycles, and module communities. Use when user asks about architecture, module structure, coupling, circular dependencies, or which files to review or change together.
---

# Python Architecture Review with pyscn

Run the pyscn CLI to understand module structure and coupling. No install needed: `uvx pyscn@latest <command>`.

## Commands

| User Request | Command |
|-------------|---------|
| "Check class coupling" | `uvx pyscn@latest analyze --select cbo <path>` |
| "Find circular dependencies" | `uvx pyscn@latest analyze --select deps <path>` |
| "Check class cohesion" | `uvx pyscn@latest analyze --select lcom <path>` |
| "Which files should I review together?" | `uvx pyscn@latest analyze --select communities --json <path>` |
| "Map the module architecture" | `uvx pyscn@latest analyze --select communities --json <path>` |

Useful flags: `--min-cbo <n>` (minimum CBO to report), `--json` (detailed report file). **Report files are NOT written to stdout**; they go to `.pyscn/reports/` and the path is printed on completion.

## Module Communities (context map for agents)

Module community detection is **opt-in**: run with `--select communities --json` and read the report file.

The JSON report includes `community_analysis.community_context_map`, a compact, deterministic map telling you which modules to inspect together:
- `bundles[]`: clusters of modules that change together. Each has `modules`, `packages`, `risk_level`, the cluster's `bridge_modules`, an optional `suggested_review_scope` (path prefix to focus a review on), and a one-line `summary`. Large bundles cap the module list with a `... +N more` marker (`module_count` always holds the true total).
- `bridge_modules[]`: modules that couple two or more communities (`module`, the communities it `connects`, and a `reason`). Widen the review scope to include these when touching a bundle they connect.

Use the map to scope reviews and refactors: load all modules in a bundle together, and pull in connected bridge modules before changing cluster boundaries.

## Interpreting Coupling Results

- High CBO classes depend on many others; changes ripple widely. Suggest interface extraction or dependency inversion.
- High LCOM4 classes bundle unrelated responsibilities; suggest splitting along the disconnected method groups.
- Dependency cycles are the highest-priority architectural issue; name the modules in each cycle and the weakest edge to break.

Always tie findings back to concrete modules and suggest a specific structural change.

## MCP Alternative

If the `pyscn-mcp` MCP server is connected, you can use its tools instead of the CLI: `check_coupling` for CBO metrics, or `analyze_code` with `analyses: ["deps"]`, `["lcom"]`, or `["communities"]` (communities require `output_mode: "full"`; the full response includes the same `community_context_map`). Results are returned inline with no report files.
