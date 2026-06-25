---
title: Module Community Detection
description: Discover natural architectural boundaries in a Python codebase from import dependencies. Learn when to run community detection, how to interpret modularity and bridge modules, and how to enable it in CI.
---

# Module Community Detection

Community detection groups Python modules by how they actually import each other, not by folder names or configured architecture layers. pyscn builds a module dependency graph from your code, runs the Leiden clustering algorithm, and reports cohesive communities plus modules that sit on boundaries between them.

Use it when you want answers like:

- What are the natural module clusters in this repository?
- Which modules connect otherwise separate parts of the codebase?
- Do import communities leak dependencies across boundaries?
- Which files should an agent or reviewer inspect together?

Community detection is **opt-in**. It does not run in a default `pyscn analyze` invocation until you enable it through `--select communities` or `[communities] enabled = true`.

## Quick start

```bash
# Standalone community JSON (machine-readable, stable field names)
pyscn analyze --json --select communities src/

# Include communities in a full unified report
pyscn analyze --json --select deps,communities src/
```

Standalone `--json --select communities` writes only the `community_analysis` object to `.pyscn/reports/`. Unified `pyscn analyze --json` embeds the same object under the top-level `community_analysis` key. See [Output Schemas](../output/schemas.md#community-analysis-object).

Enable by default in config:

```toml
[communities]
enabled = true
```

## How it works

1. pyscn parses Python imports and builds a directed module dependency graph (the same foundation used by dependency analysis).
2. The graph is projected to an undirected weighted view for clustering. Opposite import directions between two modules increase edge weight.
3. The native Go Leiden implementation partitions modules into communities and computes modularity.
4. Per-community metrics count internal vs cross-community edges. Modules with edges to multiple communities are surfaced as **bridge modules**.

Cycles, lazy imports, and third-party modules are handled gracefully: cyclic subgraphs cluster together; lazy edges are included by default; stdlib imports are excluded by default.

## Interpreting results

### Communities

Each community is a cluster of modules that import each other more densely than they import outsiders.

| Signal | What it suggests |
| --- | --- |
| High `internal_edges`, low `external_edges` | A cohesive subsystem with few outward dependencies. |
| High `external_dependency_ratio` | The cluster depends heavily on other communities — a coupling hotspot. |
| Large `size` with many `packages` | A cross-cutting area spanning multiple declared packages. |
| Many small single-module communities | A flat or disconnected graph (common in tiny repos or leaf modules). |

Community `id` values (`community_1`, `community_2`, …) are stable for a given graph topology and deterministic tie-breaking order.

### Bridge modules

A bridge module belongs to one community but has import edges into other communities. These are prime review targets when you refactor boundaries or reduce coupling.

`cross_community_edges` counts edges that cross the home community. `target_communities` lists the other community ids reached by those edges.

### Modularity

`modularity` is a quality score for the partition (higher is generally better, range roughly −0.5 to 1.0). Use it to compare runs on the same codebase after refactors, not as an absolute pass/fail threshold across unrelated projects.

- **≈ 0** — weak community structure (e.g. no import edges, or uniformly connected modules).
- **0.3 – 0.7** — typical for repositories with identifiable subsystems.
- **Very high** — strong separation; verify the graph is not trivially disconnected.

## Configuration

All keys live under `[communities]` in `.pyscn.toml` or `[tool.pyscn.communities]` in `pyproject.toml`. Full reference: [Configuration Reference](../configuration/reference.md#communities).

```toml
[communities]
enabled = false                  # opt-in; CLI --select communities also enables per run
algorithm = "leiden"             # currently the only supported value
scope = "module"                 # module-level graph (function/class scope is future work)
min_community_size = 2           # communities smaller than this remain as singleton partitions
include_lazy_edges = true        # count TYPE_CHECKING / lazy imports as edges
report_bridge_modules = true     # populate bridge_modules in output
resolution = 1.0                 # Leiden resolution (higher → more, smaller communities)
```

CLI equivalents:

| Goal | Command |
| --- | --- |
| Run only communities | `pyscn analyze --select communities .` |
| Skip even if config enables | `pyscn analyze --skip-communities .` |
| Standalone JSON artifact | `pyscn analyze --json --select communities .` |

`--select` takes precedence: `[communities] enabled = true` does not run communities unless `communities` appears in `--select` when `--select` is used.

## Determinism

For a fixed repository snapshot and configuration, pyscn guarantees **deterministic** community detection:

- Module names are sorted before graph construction.
- Neighbor lists use stable index ordering.
- Community ids are assigned in sorted module order after partitioning.
- JSON output sorts `communities` by `id`, `modules` alphabetically, and `bridge_modules` by module name.

Repeated runs on the same fixture produce identical `communities`, `bridge_modules`, and `modularity` values. Floating-point fields are rounded for stable JSON diffs (four decimal places).

**Not deterministic across:**

- Different pyscn versions (algorithm or rounding changes).
- Different `min_community_size`, `resolution`, or `include_lazy_edges` settings.
- Repository changes that add/remove modules or imports.

There is no random seed or sampling step in the current Leiden implementation.

## Output formats

| Format | Support |
| --- | --- |
| JSON | Standalone (`--json --select communities`) or unified analyze output |
| YAML | Unified analyze only (`--yaml` with communities enabled) |
| Text / HTML / CSV / DOT | Community summary formatters (see [Output Schemas](../output/schemas.md)) |

Field-level schema documentation: [community_analysis object](../output/schemas.md#community-analysis-object).

## Example interpretation

Given a bridge fixture where `bridge.py` imports one cluster and is imported by another:

```json
{
  "total_communities": 2,
  "modularity": 0.2188,
  "bridge_modules": [
    {
      "module": "bridge",
      "community": "community_1",
      "cross_community_edges": 1,
      "target_communities": ["community_2"]
    }
  ]
}
```

Read this as: two natural clusters exist, and `bridge` is the coupling point between them. Extracting shared logic from `bridge` or inverting dependencies may improve modularity.

## Follow-up work (Phase 2 and 3)

Module-level community detection is Phase 1 of [GitHub issue #564](https://github.com/ludo-technologies/pyscn/issues/564). Planned extensions include:

- Community vs package/layer mismatch scoring
- DOT export with community-colored subgraphs
- HTML macro-architecture visualization
- AI-agent context maps
- Richer edge weights from call/type/reference data
- Function- and class-level clustering

Track progress on issue #564.

## See also

- [Output Schemas](../output/schemas.md#community-analysis-object) — JSON field reference
- [`pyscn analyze`](../cli/analyze.md) — CLI flags including `--select communities`
- [Configuration Reference](../configuration/reference.md#communities) — `[communities]` keys
- [Circular import rule](../rules/circular-import.md) — related dependency-structure analysis