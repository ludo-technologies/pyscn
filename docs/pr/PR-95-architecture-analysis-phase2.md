# feat(architecture): layer rules validation (Phase 2)

## Summary
- Implement real architecture analysis based on layer rules.
- Load `[architecture]` config (layers/rules/strict_mode) from `.pyscn.toml` or `pyproject.toml` (`[tool.pyscn.architecture]`).
- Build a dependency graph via `ModuleAnalyzer`, map modules to layers, evaluate allow/deny rules, and compute layer-level metrics.
- If no rules are provided, the architecture section is omitted (no mock/fixed score).

## Rationale
- Remove placeholder (fixed score) behavior and provide actionable, rule-driven architecture validation.
- Keep implementation consistent with the rest of the system by reusing `ModuleAnalyzer`.
- Provide safe defaults: when rules are not defined, analysis is skipped to avoid misleading results.

## Changes
- service/system_analysis_config_loader.go
  - Load `architecture.strict_mode`, `architecture.layers[]` (name, packages, description), `architecture.rules[]` (from, allow[], deny[]).
  - Merge `ArchitectureRules` with CLI request when provided.
- service/system_analysis_service.go
  - AnalyzeArchitecture:
    - Guard: return `nil` when no rules are provided.
    - Build graph via `ModuleAnalyzer` (respects stdlib/third-party/relative, include/exclude patterns).
    - Map modules to layers using package patterns (`*` wildcard supported).
    - Evaluate edges: deny takes precedence; if `allow[]` is set, the target layer must be included. Under `strict_mode`, unknown/no-rule cases raise warnings.
    - Metrics: `LayerCoupling`, `LayerCohesion` (intra/total), `ComplianceScore = 1 - violations/checked_edges`.
    - Populate `ArchitectureAnalysisResult`, `LayerAnalysis`, and `SeverityBreakdown`.
  - Remove legacy ad-hoc layer counting helpers.

## Configuration
- `.pyscn.toml` (TOML)
```toml
[architecture]
strict_mode = true

[[architecture.layers]]
name = "presentation"
packages = ["myapp.web.*", "myapp.ui*"]

[[architecture.layers]]
name = "application"
packages = ["myapp.app*"]

[[architecture.layers]]
name = "domain"
packages = ["myapp.domain*"]

[[architecture.layers]]
name = "infrastructure"
packages = ["myapp.infra*"]

[[architecture.rules]]
from = "presentation"
allow = ["application"]

[[architecture.rules]]
from = "application"
allow = ["domain"]

[[architecture.rules]]
from = "domain"
allow = []

[[architecture.rules]]
from = "infrastructure"
allow = ["domain"]
```

- `pyproject.toml` (TOML)
```toml
[tool.pyscn.architecture]
strict_mode = true

[[tool.pyscn.architecture.layers]]
name = "domain"
packages = ["myapp.domain*"]

[[tool.pyscn.architecture.rules]]
from = "domain"
allow = []
```

Notes:
- Patterns support `*` wildcard; `.` is treated as a literal separator.
- Unknown-layer edges and missing rules produce warnings in `strict_mode`.

## Usage
- Architecture analysis runs when `AnalyzeArchitecture` is enabled (via config) and rules are present.
- Output appears under the Architecture section:
  - Summary: `ComplianceScore`, `TotalViolations`, `TotalRules`.
  - Details: `LayerViolations` (allow/deny violations), `LayerCoupling`, `LayerCohesion`, `ProblematicLayers`.
- `deps` command is unchanged (dependency-only). Architecture analysis is used in system-level analysis paths.

## Tests / QA
- Build: `make build` ✅
- Lint: `make lint` ✅
- Unit tests: `make test` ✅
- Manual checks:
  - With rules → Architecture section rendered with compliance, violations, and coupling/cohesion.
  - Without rules → Architecture section omitted.

## Backward Compatibility
- No breaking CLI changes.
- When rules are not configured, analysis is skipped (previously returned a placeholder score). This is safer and avoids misleading outputs.

## Performance
- Uses `ModuleAnalyzer` for dependency discovery. For very large graphs, evaluation is linear in edge count. Further optimizations can be introduced if needed.

## Known Limitations / Follow-ups
- Tests: add table-driven tests for allow/deny/no-rule/unknown-layer cases.
- Docs: add a dedicated page for architecture configuration and examples.
- Pattern matching is glob-like with `*`; consider richer matching (regex) in future.
- Consider surfacing `fromLayer`/`toLayer` in `LayerViolation` entries for easier scanning.

## Checklist
- [x] Conventional commit title
- [x] Summary and rationale
- [x] Config changes documented
- [x] Tests / Lint / Build passing
- [ ] Additional tests for rule evaluation (follow-up)
- [ ] Docs page for architecture rules (follow-up)

