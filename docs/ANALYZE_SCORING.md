# Analyze Scoring Reference

This document explains how the current `pyscn analyze` command derives the health score and the category scores that appear in CLI and HTML outputs. The implementation lives primarily in `domain/analyze.go` with orchestration in `app/analyze_usecase.go`.

## Calculation Flow

1. Each analyzer populates an `AnalyzeResponse`. The `AnalyzeUseCase` composes the project summary (`AnalyzeSummary`) with aggregate metrics (function counts, average complexity, clone duplication, dependency stats, etc.).
2. `AnalyzeSummary.CalculateHealthScore()` validates the inputs, computes penalties per category, converts those penalties to scores on a 0–100 scale, and subtracts the penalties from an overall score that starts at 100.
3. If validation fails, the CLI logs a warning, applies a lightweight fallback scorer, and still surfaces the grade.

All scores are bounded to 0–100. The overall health score has a floor of 10 to avoid degenerate results for heavily penalised projects.

## Category Penalties and Scores

Penalties are additive. Each category subtracts up to the maximum listed points from the base score (100). The same penalty value is then converted to a category score via `100 - (penalty / maxPenalty * 100)`.

| Category            | Metric(s)                                                                                                                                                 | Thresholds → Penalty                                                                                                  | Max Penalty |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------|-------------|
| Complexity          | Average cyclomatic complexity across functions                                                                                                           | >20 → 20, >10 → 12, >5 → 6                                                                                              | 20          |
| Dead Code           | Count of critical dead code issues, normalised by logarithm of total files (threshold kicks in once more than 10 files are analysed)                      | Up to 20 based on `criticalDeadCode / normalizationFactor`, capped at 20                                               | 20          |
| Duplication         | Percentage of duplicated code across clone groups                                                                                                        | >20% → 20, >10% → 12, >3% → 6                                                                                          | 20          |
| Coupling (CBO)      | Weighted ratio of high-risk (`CBO > 7`) and medium-risk (`3 < CBO ≤ 7`) classes using weight 1.0 and 0.5 respectively, divided by total measured classes | >30% → 20, >15% → 12, >5% → 6                                                                                          | 20          |
| Dependencies        | Module dependency graph: proportion of modules in cycles, dependency depth above `log₂(N)+1`, Main Sequence Deviation                                    | Cycles up to 8 pts + depth up to 2 pts + MSD up to 2 pts (ratio/overflow calculations clamp to [0, max])               | 12          |
| Architecture        | Architecture rules compliance ratio (0–1)                                                                                                                | `round((1 - compliance) * 8)`                                                                                          | 8           |

When a category is disabled (e.g., `--skip-clones`), its penalty is zero and the prior score (100) carries forward so the missing analysis does not hurt the overall grade.

## Overall Health Score and Grade

`HealthScore = max(10, 100 - Σ penalties)`

Grades mirror the score quality thresholds that the CLI uses for emoji indicators:

- A: ≥85
- B: ≥70
- C: ≥55
- D: ≥40
- F: <40

The CLI treats a project as “healthy” when `HealthScore ≥ 70`.

## Presentation Details

- The CLI summary shows the overall score, letter grade, and per-category scores with emojis (`✅` ≥85, `👍` ≥70, `⚠️` ≥55, `❌` otherwise).
- HTML and JSON outputs expose the same scores and include additional per-category context (e.g., high-risk counts).
- When dependency or architecture analyses are disabled, their sections are omitted from the detailed summary, but the rest of the scoring remains unchanged.

## Fallback Behaviour

If the validator detects inconsistent summary metrics (negative averages, duplication >100%, etc.), the application:

1. Logs a warning about the failure to calculate the health score.
2. Uses `CalculateFallbackScore()`, which applies simple penalties:
   - −10 for average complexity above 10,
   - −5 if any dead code exists,
   - −5 if any high-complexity functions exist.
3. Enforces the same minimum score (10) and derives the grade from the fallback score.

This ensures the CLI still produces a meaningful result even when upstream metrics are incomplete or malformed.
