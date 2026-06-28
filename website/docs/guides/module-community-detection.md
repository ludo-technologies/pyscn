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

### Package mismatch

Community detection groups modules by import topology. Package mismatch metrics compare that partition to declared package boundaries (from module paths and `ModuleNode.Package` metadata).

| Field | What it suggests |
| --- | --- |
| `package_alignment_score` (0–1) | How well communities respect package boundaries. **1.0** means every package's modules live in exactly one community; **0.0** means every package is split. |
| `split_packages` | Packages whose modules appear in two or more communities — refactor candidates when you expect package = subsystem. |
| `mixed_communities` | Communities that contain modules from two or more packages — cross-cutting clusters that span declared boundaries. |
| `dominant_package` / `package_count` / `package_alignment` | Per-community composition: which package dominates, how many packages are represented, and how cohesive internal edges are within that community. |

**Example:** A bridge fixture where `mod.a`/`mod.b` cluster separately from `mod.c`/`mod.d` yields `package_alignment_score: 0` and `split_packages: ["mod"]` even though each community is internally cohesive. A billing/inventory fixture with no cross-package imports yields `package_alignment_score: 1`.

This is distinct from `SystemMetrics.ModularityIndex` in dependency analysis (an intra-package edge ratio), which measures a different cohesion signal.

### Layer mismatch

When `[architecture]` layers are configured (explicit `style`, `[[architecture.layers]]`, or `[[architecture.rules]]`), community detection also compares inferred clusters to configured layer boundaries. Layer mismatch is omitted when no architecture rules are configured — community analysis still succeeds.

Enable both dependency graph construction and architecture config, e.g. `pyscn analyze --select deps,communities .` or `[communities] enabled = true` with `[architecture]` configured.

| Field | What it suggests |
| --- | --- |
| `layer_alignment_score` (0–1) | How well communities respect configured layers. **1.0** means every layer's modules live in exactly one community; **0.0** means every layer is split. Distinct from architecture `compliance_score` / violation counts. |
| `cross_layer_communities` | Communities containing modules from two or more configured layers — clusters that span layer boundaries. |
| `layer_bridge_modules` | Bridge modules whose home community layer differs from a target community's dominant layer. |
| `dominant_layer` / `layer_count` / `layers[]` / `layer_alignment` | Per-community layer composition and internal edge cohesion within the community. |

**Example:** A bridge fixture with `api` and `infra` layers where `api.a`/`api.b` cluster separately from `infra.c`/`infra.d` yields `layer_alignment_score: 1` and `layer_bridge_modules: ["bridge", "infra.c"]` because the bridge couples api and infra communities. A fixture where both layers appear inside each community yields `layer_alignment_score: 0` and populates `cross_layer_communities`.

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

## Macro-architecture visualization

When communities are enabled, the unified HTML report's **Communities** tab includes an interactive macro-architecture graph above the existing metric cards and tables (the tables stay for accessibility and CI assertions).

- **Nodes** are modules, colored by community.
- **Bridge modules** are outlined in red and filled with a light tint so coupling points stand out.
- **Edges** that cross community boundaries are drawn in red; intra-community edges are gray.
- **Hover** a node for its name, community id, bridge flag, cross-edge count, target communities, and the dominant package/layer (when package/layer mismatch data is available).
- **Click** a node to highlight it and its direct neighbors; click again to clear.
- **Drag** a node to reposition it.

Filters above the canvas let you focus on one community, show only bridge modules, or hide isolated nodes.

The graph is rendered client-side from a compact JSON blob embedded in the report (`<script id="community-graph-data">`), so reports stay self-contained and work offline. No external graph viewer is required, and `--no-open` behavior is unchanged.

**Performance guard.** To keep the layout responsive, the graph collapses to one node per community (sized by module count) when a project exceeds the display threshold of `100` modules. Below that, every module is rendered individually. This visualization complements—and does not replace—the [DOT export](#output-formats), which remains available for large graphs and external tooling.

## Example interpretation

Given a bridge fixture where `bridge.py` imports one cluster and is imported by another:

```json
{
  "total_communities": 2,
  "modularity": 0.2188,
  "package_alignment_score": 0,
  "split_packages": ["mod"],
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

Read this as: two natural clusters exist, and `bridge` is the coupling point between them. The `mod` package is split across both communities (`package_alignment_score: 0`), so import topology disagrees with the declared package boundary. Extracting shared logic from `bridge` or regrouping `mod` modules may improve alignment.

## Follow-up work (Phase 2 and 3)

Module-level community detection is Phase 1 of [GitHub issue #564](https://github.com/ludo-technologies/pyscn/issues/564). Phase 2 mismatch scoring is available via package fields (`package_alignment_score`, `split_packages`, `mixed_communities`) and layer fields (`layer_alignment_score`, `cross_layer_communities`, `layer_bridge_modules`) when architecture layers are configured. DOT export with community-colored subgraphs and the [HTML macro-architecture visualization](#macro-architecture-visualization) are also available. Remaining planned extensions include:
- AI-agent context maps
- Richer edge weights from call/type/reference data
- Function- and class-level clustering

Track progress on issue #564.

## See also

- [Output Schemas](../output/schemas.md#community-analysis-object) — JSON field reference
- [`pyscn analyze`](../cli/analyze.md) — CLI flags including `--select communities`
- [Configuration Reference](../configuration/reference.md#communities) — `[communities]` keys
- [Circular import rule](../rules/circular-import.md) — related dependency-structure analysis