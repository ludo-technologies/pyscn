---
name: architecture-review
description: Analyze Python module architecture using pyscn MCP tools - class coupling (CBO), cohesion (LCOM), dependency cycles, and module communities. Use when user asks about architecture, module structure, coupling, circular dependencies, or which files to review or change together.
---

# Python Architecture Review with pyscn MCP

Use the pyscn MCP tools to understand module structure and coupling.

## Tools

| Tool | Purpose |
|------|---------|
| `check_coupling` | Class coupling (CBO) metrics |
| `analyze_code` | Dependencies, cohesion, and module communities via `analyses` |

## Tool Selection

| User Request | Tool |
|-------------|------|
| "Check class coupling" | `check_coupling` |
| "Find circular dependencies" | `analyze_code` (`analyses: ["deps"]`) |
| "Check class cohesion" | `analyze_code` (`analyses: ["lcom"]`) |
| "Which files should I review together?" | `analyze_code` (`analyses: ["communities"]`, `output_mode: "full"`) |
| "Map the module architecture" | `analyze_code` (`analyses: ["communities"]`, `output_mode: "full"`) |

## Parameters

- `path` (required): Path to Python file or directory
- `recursive` (analyze_code): Recursively analyze directories (default: true)
- `analyses` (analyze_code): `cbo`, `lcom`, `deps`, `communities`
- `output_mode` (analyze_code): `summary` (default) or `full` — communities require `full`

## Module Communities (context map for agents)

Module community detection is **opt-in**. Request it explicitly:
- `analyses: ["communities"]`, `output_mode: "full"`

The full response includes `community_analysis.community_context_map`, a compact, deterministic map telling you which modules to inspect together:
- `bundles[]`: clusters of modules that change together. Each has `modules`, `packages`, `risk_level`, the cluster's `bridge_modules`, an optional `suggested_review_scope` (path prefix to focus a review on), and a one-line `summary`. Large bundles cap the module list with a `... +N more` marker (`module_count` always holds the true total).
- `bridge_modules[]`: modules that couple two or more communities (`module`, the communities it `connects`, and a `reason`). Widen the review scope to include these when touching a bundle they connect.

Use the map to scope reviews and refactors: load all modules in a bundle together, and pull in connected bridge modules before changing cluster boundaries.

## Interpreting Coupling Results

- High CBO classes depend on many others; changes ripple widely. Suggest interface extraction or dependency inversion.
- High LCOM4 classes bundle unrelated responsibilities; suggest splitting along the disconnected method groups.
- Dependency cycles are the highest-priority architectural issue; name the modules in each cycle and the weakest edge to break.

Always tie findings back to concrete modules and suggest a specific structural change.
