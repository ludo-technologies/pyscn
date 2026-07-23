# Analyze Scoring Reference

This document explains how the current `pyscn analyze` command derives the health score and the category scores that appear in CLI and HTML outputs. The implementation lives primarily in `domain/analyze.go` with orchestration in `app/analyze_usecase.go`.

## Calculation Flow

1. Each analyzer populates an `AnalyzeResponse`. The `AnalyzeUseCase` composes the project summary (`AnalyzeSummary`) with aggregate metrics (function counts, average complexity, clone duplication, dependency stats, etc.).
2. `AnalyzeSummary.CalculateHealthScore()` validates the inputs, computes penalties per category, converts those penalties to scores on a 0–100 scale, and subtracts the penalties from an overall score that starts at 100.
3. If validation fails, the CLI logs a warning, applies a lightweight fallback scorer, and still surfaces the grade.

All scores are bounded to 0–100. The overall health score can reach 0 for projects with severe quality issues.

## Category Penalties and Scores

Penalties are additive. Each category subtracts up to the maximum listed points from the base score (100). The same penalty value is then converted to a category score via `100 - (penalty / maxPenalty * 100)`.

| Category            | Metric(s)                                                                                                                                                 | Penalty Formula                                                                                                        | Max Penalty |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|-------------|
| Complexity          | Average cyclomatic complexity, cognitive complexity, and nesting depth across functions                                                                  | Max of continuous linear penalties:<br/>McCabe starts at avg=2, reaches max at avg=15<br/>Cognitive starts at avg=15, reaches max at avg=25<br/>Nesting starts at avg=3, reaches max at avg=7 | 20          |
| Dead Code           | Weighted count of dead code issues (Critical=1.0, Warning=0.5, Info=0.2), normalised by logarithm of total files                                         | `min(20, weightedDeadCode / normalizationFactor)`<br/>Normalization: `log10(max(1, totalFiles/10))`                    | 20          |
| Duplication         | Percentage of duplicated code across clone groups                                                                                                        | Continuous linear: `min(20, max(0, (duplication - 1) / 7 * 20))`<br/>Starts at 1%, reaches max at 8%                | 20          |
| Coupling (CBO)      | Weighted ratio of high-risk (`CBO > 7`) and medium-risk (`3 < CBO ≤ 7`) classes using weight 1.0 and 0.5 respectively, divided by total measured classes | Continuous linear: `min(20, ratio / 0.12 * 20)`<br/>Starts at 0%, reaches max at 12%                                  | 20          |
| Dependencies        | Module dependency graph: proportion of modules in cycles, dependency depth above `log₂(N)+1`, Main Sequence Deviation                                    | Cycles: up to 10 pts (`max` of proportional and `log₂(modulesInCycles + 1)` floor)<br/>Depth: up to 3 pts (excess over expected)<br/>MSD: up to 3 pts (proportional) | 16          |
| Architecture        | Architecture rules compliance ratio (0–1)                                                                                                                | `round((1 - compliance) * 12)`                                                                                         | 12          |
| Communities         | Module community structure: low modularity Q, cross-community edge ratio, bridge-module count, and (when available) package/layer alignment              | `round(riskRatio * 10)` where `riskRatio` is the weighted risk blend below                                            | 10          |

When a category is disabled (e.g., `--skip-clones`), its penalty is zero and the prior score (100) carries forward so the missing analysis does not hurt the overall grade.

## Community Risk Score

Community detection runs in default analyze unless disabled. The community penalty only applies when communities ran **and** at least two communities were detected; otherwise the category scores 100 with a zero penalty, so disabling communities or analyzing trivial graphs never changes existing grades (backward compatible).

The system-level **community risk score** (`community_risk_score`, 0–100, higher = worse) is a weighted blend of normalised risk factors. The category quality score is its inverse: `CommunityScore = 100 - community_risk_score`. The health-score penalty is `round(riskRatio * 10)`.

Each factor is normalised to `0..1` (1 = worst). The blend is a weighted average that is renormalised over whichever factors are available, so optional factors only count when their metadata exists:

| Factor                      | Weight | Formula                                                                 | Availability                       |
|-----------------------------|--------|------------------------------------------------------------------------|------------------------------------|
| Low modularity Q            | 0.40   | `clamp01((0.30 - Q) / 0.30)` — risk rises as Q falls below 0.30        | Always (when ≥ 2 communities)      |
| Cross-community edge ratio  | 0.30   | `clamp01(crossRatio / 0.50)`, `crossRatio = crossEdges / (internal + cross)` | When the graph has edges     |
| Bridge-module count         | 0.30   | `clamp01(bridgeModules / communityCount)` — saturates at ~1 per community | When communities exist          |
| Low package alignment       | 0.25   | `clamp01(1 - package_alignment_score)`                                 | When package metadata is present   |
| Low layer alignment         | 0.25   | `clamp01(1 - layer_alignment_score)`                                   | When architecture layers configured |

The cross-community edge ratio also captures the aggregate `external_dependency_ratio` at the system level, so that signal is not double-counted.

### Per-community `risk_level`

Each community is also classified `low` / `medium` / `high` from a local risk ratio blending its `external_dependency_ratio` (weight 0.5) with `1 - package_alignment` (0.25, when available) and `1 - layer_alignment` (0.25, when available):

- `high`: ratio ≥ 0.60
- `medium`: 0.30 ≤ ratio < 0.60
- `low`: ratio < 0.30

## Overall Health Score and Grade

`HealthScore = max(0, 100 - Σ penalties)`

The minimum score is 0, allowing truly low scores for severely problematic code. Grades use stricter thresholds that mirror the score quality thresholds the CLI uses for emoji indicators:

- A: ≥90 (Excellent ✅)
- B: ≥75 (Good 👍)
- C: ≥60 (Fair ⚠️)
- D: ≥45
- F: <45 (Poor ❌)

The CLI treats a project as "healthy" when `HealthScore ≥ 70`.

## Presentation Details

- The CLI summary shows the overall score, letter grade, and per-category scores with emojis (`✅` ≥90, `👍` ≥75, `⚠️` ≥60, `❌` otherwise).
- HTML and JSON outputs expose the same scores and include additional per-category context (e.g., high-risk counts).
- When dependency or architecture analyses are disabled, their sections are omitted from the detailed summary, but the rest of the scoring remains unchanged.

## Fallback Behaviour

If the validator detects inconsistent summary metrics (negative averages, duplication >100%, etc.), the application:

1. Logs a warning about the failure to calculate the health score.
2. Uses `CalculateFallbackScore()`, which applies simple penalties:
   - −10 for average complexity above 10,
   - −5 if any dead code exists,
   - −5 if any high-complexity functions exist.
3. Enforces the same minimum score (0) and derives the grade from the fallback score.

This ensures the CLI still produces a meaningful result even when upstream metrics are incomplete or malformed.

## Key Changes (Issue #212)

The scoring system has been made stricter to better reflect code quality issues:

1. **Continuous Penalties**: Complexity, duplication, and coupling now use continuous linear functions instead of step functions, providing more granular scoring.
2. **Weighted Dead Code**: Warning and Info severity dead code issues now contribute to the score (Critical=1.0, Warning=0.5, Info=0.2).
3. **Stricter Grades**: Grade thresholds increased (A: 85→90, B: 70→75, C: 55→60, D: 40→45) to better distinguish code quality levels.
4. **Increased System Penalties**: Dependencies (12→16) and Architecture (8→12) penalties increased to better reflect structural issues.
5. **No Minimum Floor**: Minimum score reduced from 10 to 0 to allow truly low scores for severely problematic code.
