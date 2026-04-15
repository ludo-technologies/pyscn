# Health Score

## Overview

The health score is a 0–100 integer summarizing every enabled analysis (complexity, dead code, duplication, coupling, cohesion, dependencies, architecture) into a single value. The calculation is pure and deterministic — given the same `AnalyzeSummary` inputs, `CalculateHealthScore()` always produces the same score. It is designed to be stored per commit in CI so changes are tracked over time.

## Grade

The score maps to a letter grade using strict `≥` thresholds:

| Grade | Score range | Threshold constant |
| ----- | ----------- | ------------------ |
| A     | 90–100      | `GradeAThreshold = 90` |
| B     | 75–89       | `GradeBThreshold = 75` |
| C     | 60–74       | `GradeCThreshold = 60` |
| D     | 45–59       | `GradeDThreshold = 45` |
| F     | 0–44        | (below D)              |

Defined in `domain/analyze.go:90-93`. A project is considered "healthy" at `HealthScore ≥ 70` (`HealthyThreshold`, `domain/analyze.go:103`).

## Formula

```
score = 100
      - complexityPenalty       (0–20)
      - deadCodePenalty         (0–20)
      - duplicationPenalty      (0–20)
      - couplingPenalty         (0–20)
      - cohesionPenalty         (0–20)
      - dependencyPenalty       (0–16)
      - architecturePenalty     (0–12)

HealthScore = max(0, score)
```

Each penalty is capped at its individual maximum:

| Category     | Max penalty | Constant                          |
| ------------ | ----------- | --------------------------------- |
| Complexity   | 20          | literal `20.0` in formula         |
| Dead Code    | 20          | `MaxDeadCodePenalty = 20`         |
| Duplication  | 20          | literal `20.0` in formula         |
| Coupling     | 20          | literal `20.0` in formula         |
| Cohesion     | 20          | literal `20.0` in formula         |
| Dependencies | 16          | `MaxDependencyPenalty = 10+3+3`   |
| Architecture | 12          | `MaxArchitecturePenalty = 12`     |

The score floor is `MinimumScore = 0` (`domain/analyze.go:102`), applied after penalty summation.

## Penalty specifications

### Complexity

**Inputs.** `AverageComplexity` (float64).

**Formula.**

```
if AverageComplexity <= 2.0:
    penalty = 0
else:
    penalty = min(20, round((AverageComplexity - 2.0) / 13.0 * 20.0))
```

**Constants.**

| Name      | Value | Meaning                                    |
| --------- | ----- | ------------------------------------------ |
| baseline  | 2.0   | Average complexity below which penalty = 0 |
| range     | 13.0  | Denominator — max penalty reached at avg = 15 |
| max       | 20.0  | Penalty cap                                |

**Saturation.** Reaches 20 at `AverageComplexity >= 15.0`.

**Edge cases.** `AverageComplexity = 0` or any value `≤ 2.0` yields penalty 0. `Validate()` rejects negative values.

Source: `domain/analyze.go:266-279`.

### Dead Code

**Inputs.** `CriticalDeadCode`, `WarningDeadCode`, `InfoDeadCode` (int). `TotalFiles` (int) is used to derive the normalization factor in `CalculateHealthScore()`.

**Formula.**

```
weighted = CriticalDeadCode * 1.0
         + WarningDeadCode  * 0.5
         + InfoDeadCode     * 0.2

if TotalFiles <= 10:
    normalization = 1.0
else:
    normalization = 1.0 + log10(TotalFiles / 10.0)

if weighted <= 0:
    penalty = 0
else:
    penalty = int(min(20.0, weighted / normalization))
```

**Constants.**

| Name                      | Value | Meaning                          |
| ------------------------- | ----- | -------------------------------- |
| critical weight           | 1.0   | Full weight for Critical findings |
| warning weight            | 0.5   | Half weight for Warning findings  |
| info weight               | 0.2   | Minimal weight for Info findings  |
| normalization threshold   | 10    | Files below which norm factor = 1 |
| `MaxDeadCodePenalty`      | 20    | Penalty cap                       |

**Saturation.** Reaches 20 when `weighted / normalization >= 20.0`.

**Edge cases.** Zero findings yields penalty 0. The truncation is `int()` (toward zero), not `math.Round` — so a computed value of `1.99` becomes `1`.

Source: `domain/analyze.go:283-296`. Normalization factor derived in `CalculateHealthScore()` at `domain/analyze.go:474-477`.

### Duplication

**Inputs.** `CodeDuplication` (float64). This field is a pre-computed duplication percentage, capped at 10 by the upstream calculation.

**Formula.**

```
if CodeDuplication <= 0:
    penalty = 0
else:
    penalty = min(20, round(CodeDuplication / 10.0 * 20.0))
```

**Constants.**

| Name                         | Value | Meaning                     |
| ---------------------------- | ----- | --------------------------- |
| `DuplicationThresholdLow`    | 0.0   | 0% duplication = 0 penalty  |
| `DuplicationThresholdHigh`   | 10.0  | 10% duplication = max penalty |
| max                          | 20.0  | Penalty cap                 |

**Upstream computation.** `CodeDuplication` itself is computed in `app/analyze_usecase.go:570-583`:

```
lines_in_thousands = max(GroupDensityMinLines, total_lines / GroupDensityLinesUnit)
group_density      = clone_groups / lines_in_thousands
CodeDuplication    = min(DuplicationThresholdHigh, group_density * GroupDensityCoefficient)
```

Where `GroupDensityLinesUnit = 1000.0`, `GroupDensityMinLines = 1.0`, `GroupDensityCoefficient = 20.0`, `DuplicationThresholdHigh = 10.0` (`domain/analyze.go:52, 62-64`).

**Saturation.** Reaches 20 when `CodeDuplication >= 10.0`. Since the upstream caps at 10, saturation is reached at exactly 0.5 clone groups per 1000 analyzed lines.

**Edge cases.** No clone groups or no analyzed lines → `CodeDuplication = 0` → penalty 0. `Validate()` rejects values outside `[0, 100]`.

Source: `domain/analyze.go:300-314`.

### Coupling (CBO)

**Inputs.** `CBOClasses`, `HighCouplingClasses` (CBO > 7), `MediumCouplingClasses` (3 < CBO ≤ 7).

**Formula.**

```
if CBOClasses == 0:
    penalty = 0
else:
    weighted = HighCouplingClasses + 0.5 * MediumCouplingClasses
    ratio    = weighted / CBOClasses
    penalty  = min(20, round(ratio / 0.25 * 20.0))
```

**Constants.**

| Name            | Value | Meaning                                          |
| --------------- | ----- | ------------------------------------------------ |
| high weight     | 1.0   | Full weight for high-coupling classes            |
| medium weight   | 0.5   | Half weight for medium-coupling classes          |
| saturation ratio | 0.25 | 25% weighted-problematic classes → max penalty   |
| max             | 20.0  | Penalty cap                                      |

**Saturation.** Reaches 20 when the weighted ratio is `≥ 0.25`.

**Edge cases.** `CBOClasses = 0` → penalty 0. `Validate()` ensures high + medium ≤ total.

Source: `domain/analyze.go:318-336`.

### Cohesion (LCOM)

**Inputs.** `LCOMClasses`, `HighLCOMClasses` (LCOM4 > 5), `MediumLCOMClasses` (2 < LCOM4 ≤ 5).

**Formula.**

```
if LCOMClasses == 0:
    penalty = 0
else:
    weighted = HighLCOMClasses + 0.5 * MediumLCOMClasses
    ratio    = weighted / LCOMClasses
    penalty  = min(20, round(ratio / 0.30 * 20.0))
```

**Constants.**

| Name             | Value | Meaning                                         |
| ---------------- | ----- | ----------------------------------------------- |
| high weight      | 1.0   | Full weight for high-LCOM classes               |
| medium weight    | 0.5   | Half weight for medium-LCOM classes             |
| saturation ratio | 0.30  | 30% weighted-problematic classes → max penalty  |
| max              | 20.0  | Penalty cap                                     |

**Saturation.** Reaches 20 when the weighted ratio is `≥ 0.30`.

**Edge cases.** `LCOMClasses = 0` → penalty 0. `Validate()` ensures high + medium ≤ total.

Source: `domain/analyze.go:340-358`.

### Dependencies

**Inputs.** `DepsEnabled` (bool), `DepsTotalModules`, `DepsModulesInCycles`, `DepsMaxDepth` (int), `DepsMainSequenceDeviation` (float64, 0–1).

**Formula.** Three sub-penalties summed:

```
if not DepsEnabled:
    return 0

# Cycles sub-penalty (max 10)
if DepsTotalModules > 0:
    ratio  = clamp(0, 1, DepsModulesInCycles / DepsTotalModules)
    cycles = round(10 * ratio)
else:
    cycles = 0

# Depth sub-penalty (max 3)
if DepsTotalModules > 0:
    expected = max(3, ceil(log2(DepsTotalModules + 1)) + 1)
    depth    = clamp(0, 3, DepsMaxDepth - expected)
else:
    depth = 0

# Main Sequence Deviation sub-penalty (max 3)
if DepsMainSequenceDeviation > 0:
    msd = round(3 * clamp(0, 1, DepsMainSequenceDeviation))
else:
    msd = 0

penalty = cycles + depth + msd     # max 16
```

**Constants.**

| Name                 | Value | Meaning                         |
| -------------------- | ----- | ------------------------------- |
| `MaxCyclesPenalty`   | 10    | Cap for cycles sub-penalty      |
| `MaxDepthPenalty`    | 3     | Cap for depth sub-penalty       |
| `MaxMSDPenalty`      | 3     | Cap for MSD sub-penalty         |
| `MaxDependencyPenalty` | 16  | Sum of the three caps           |

**Saturation.** Reaches 16 when every module is in a cycle, depth exceeds expected by ≥ 3, and MSD ≥ 1.

**Edge cases.** `DepsEnabled = false` → penalty 0. `DepsTotalModules = 0` → cycles and depth contribute 0; only MSD can fire. `Validate()` enforces `MSD ∈ [0, 1]` and `DepsModulesInCycles ≤ DepsTotalModules`.

Source: `domain/analyze.go:361-406`.

### Architecture

**Inputs.** `ArchEnabled` (bool), `ArchCompliance` (float64, 0–1).

**Formula.**

```
if not ArchEnabled:
    penalty = 0
else:
    penalty = round(12 * (1 - clamp(0, 1, ArchCompliance)))
```

**Constants.**

| Name                     | Value | Meaning            |
| ------------------------ | ----- | ------------------ |
| `MaxArchPenalty`         | 12    | Penalty cap        |
| `MaxArchitecturePenalty` | 12    | Alias for the cap  |

**Saturation.** Reaches 12 at `ArchCompliance = 0.0`.

**Edge cases.** `ArchEnabled = false` → penalty 0. `Validate()` enforces `ArchCompliance ∈ [0, 1]` when enabled.

Source: `domain/analyze.go:409-422`.

## Category scores

Each category exposes a 0–100 score in the report. The conversion depends on the category's penalty scale.

**Most categories (Complexity, Dead Code, Duplication, Coupling, Cohesion).** These all have penalty cap 20 (`MaxScoreBase = 20`). The score is computed by `penaltyToScore(penalty, 20)`:

```
score = 100 - round(penalty * 100 / 20) = 100 - penalty * 5
```

Effectively each unit of penalty costs 5 score points. A penalty of 0 yields 100; a penalty of 20 yields 0.

**Dependencies.** Penalty cap is 16, so the score is normalized to the 20-point scale first via `normalizeToScoreBase`:

```
normalized = round(dependencyPenalty / 16 * 20)
DependencyScore = 100 - round(normalized * 100 / 20) = 100 - normalized * 5
```

**Architecture.** Special case — the score is taken directly from compliance:

```
ArchitectureScore = round(ArchCompliance * 100)
```

So `ArchCompliance = 0.98` yields `ArchitectureScore = 98`, regardless of the penalty rounding that feeds the overall score.

Source: `domain/analyze.go:426-453`, `domain/analyze.go:483-510`.

## Worked example

Inputs:

| Field                         | Value  |
| ----------------------------- | ------ |
| `TotalFiles`                  | 50     |
| `AverageComplexity`           | 8.0    |
| `CriticalDeadCode`            | 2      |
| `WarningDeadCode`             | 1      |
| `InfoDeadCode`                | 0      |
| `CodeDuplication`             | 7.5    |
| `CBOClasses`                  | 20     |
| `HighCouplingClasses`         | 3      |
| `MediumCouplingClasses`       | 2      |
| `LCOMClasses`                 | 20     |
| `HighLCOMClasses`             | 1      |
| `MediumLCOMClasses`           | 3      |
| `DepsEnabled`                 | true   |
| `DepsTotalModules`            | 8      |
| `DepsModulesInCycles`         | 1      |
| `DepsMaxDepth`                | 5      |
| `DepsMainSequenceDeviation`   | 0.2    |
| `ArchEnabled`                 | true   |
| `ArchCompliance`              | 0.85   |

**Complexity penalty.** `(8.0 − 2.0) / 13.0 × 20.0 = 9.2308 → round → 9`.

**Dead code penalty.** `weighted = 2×1.0 + 1×0.5 + 0×0.2 = 2.5`. `normalization = 1.0 + log10(50/10) = 1.0 + log10(5) ≈ 1.6990`. `2.5 / 1.6990 ≈ 1.4715`. `int(min(20.0, 1.4715)) = 1`.

**Duplication penalty.** `7.5 / 10.0 × 20.0 = 15.0 → round → 15`.

**Coupling penalty.** `weighted = 3 + 0.5×2 = 4`. `ratio = 4/20 = 0.20`. `0.20 / 0.25 × 20.0 = 16.0 → round → 16`.

**Cohesion penalty.** `weighted = 1 + 0.5×3 = 2.5`. `ratio = 2.5/20 = 0.125`. `0.125 / 0.30 × 20.0 ≈ 8.333 → round → 8`.

**Dependency penalty.**
- Cycles: `ratio = 1/8 = 0.125`. `round(10 × 0.125) = round(1.25) = 1`.
- Depth: `expected = max(3, ceil(log2(9)) + 1) = max(3, 4 + 1) = 5`. Excess = `5 − 5 = 0`.
- MSD: `round(3 × 0.2) = round(0.6) = 1`.
- Total: `1 + 0 + 1 = 2`.

**Architecture penalty.** `round(12 × (1 − 0.85)) = round(1.8) = 2`.

**Sum.** `9 + 1 + 15 + 16 + 8 + 2 + 2 = 53`.

**HealthScore.** `max(0, 100 − 53) = 47`. Grade: `47 ≥ 45` → **D**.

**Category scores (as reported).**

| Category     | Penalty | Score                                |
| ------------ | ------- | ------------------------------------ |
| Complexity   | 9       | `100 − 9×5 = 55`                     |
| Dead Code    | 1       | `100 − 1×5 = 95`                     |
| Duplication  | 15      | `100 − 15×5 = 25`                    |
| Coupling     | 16      | `100 − 16×5 = 20`                    |
| Cohesion     | 8       | `100 − 8×5 = 60`                     |
| Dependencies | 2       | `normalized = round(2/16 × 20) = 3`; `100 − 3×5 = 85` |
| Architecture | —       | `round(0.85 × 100) = 85`             |

## Fallback score

`CalculateHealthScore()` first calls `Validate()` on the summary. If validation fails — e.g. `AverageComplexity < 0`, `CodeDuplication` outside `[0, 100]`, `ArchCompliance` outside `[0, 1]` when enabled, `DepsMainSequenceDeviation` outside `[0, 1]` when enabled, or the sum of high + medium classes exceeding the total for LCOM or CBO — the summary's scores are zeroed, the grade is set to `"N/A"`, and an error is returned. The caller may then invoke `CalculateFallbackScore()` as a degraded path: starting from 100, it subtracts `FallbackComplexityThreshold = 10` if `AverageComplexity > 10`, and `FallbackPenalty = 5` each for `DeadCodeCount > 0`, `HighComplexityCount > 0`, and `HighLCOMClasses > 0`, flooring at `MinimumScore = 0`.

Source: `domain/analyze.go:200-262` (`Validate`), `domain/analyze.go:456-470` (validation branch in `CalculateHealthScore`), `domain/analyze.go:538-566` (`CalculateFallbackScore`).

## Rounding

All non-integer intermediates are reduced to integers using Go's `math.Round`, which applies banker's rounding for exact `.5` values (round-half-away-from-zero for positive values in Go's implementation). The final health score is clamped to `[0, 100]` via `MinimumScore`. Category scores are clamped to `[0, 100]` inside `penaltyToScore` (`domain/analyze.go:441-453`). The one exception to `math.Round` is the dead-code penalty, which uses `int()` truncation after `math.Min` (`domain/analyze.go:294`).

## Tracking over time

The formula is deterministic — identical inputs always produce identical scores across runs and platforms. To track health over time, persist `summary.health_score` from the JSON output per commit and compare in CI. The per-category scores (`complexity_score`, `dead_code_score`, etc.) are likewise stable and can be tracked individually.

## References

All line numbers refer to `domain/analyze.go`:

- `CalculateHealthScore()` — lines 456–534
- `calculateComplexityPenalty()` — lines 266–279
- `calculateDeadCodePenalty(normalizationFactor)` — lines 283–296
- `calculateDuplicationPenalty()` — lines 300–314
- `calculateCouplingPenalty()` — lines 318–336
- `calculateCohesionPenalty()` — lines 340–358
- `calculateDependencyPenalty()` — lines 361–406
- `calculateArchitecturePenalty()` — lines 409–422
- `CalculateFallbackScore()` — lines 538–566
- `Validate()` — lines 200–262
- `penaltyToScore()` — lines 441–453
- `normalizeToScoreBase()` — lines 426–438
- `GetGradeFromScore()` — lines 569–582

Upstream `CodeDuplication` computation: `app/analyze_usecase.go:570-583`.
