# 健康评分

## 概述

健康评分是一个 0–100 的整数，汇总了所有已启用分析的结果（复杂度、死代码、代码重复、耦合度、内聚度、依赖、架构）为一个单一值。计算是纯粹且确定性的 — 给定相同的 `AnalyzeSummary` 输入，`CalculateHealthScore()` 始终产生相同的评分。它被设计为可在 CI 中按提交存储，以便追踪变化趋势。

## 等级

评分使用严格的 `≥` 阈值映射到字母等级：

| 等级 | 分数范围 | 阈值常量 |
| ----- | ----------- | ------------------ |
| A     | 90–100      | `GradeAThreshold = 90` |
| B     | 75–89       | `GradeBThreshold = 75` |
| C     | 60–74       | `GradeCThreshold = 60` |
| D     | 45–59       | `GradeDThreshold = 45` |
| F     | 0–44        | （低于 D）              |

定义在 `domain/analyze.go:90-93`。项目在 `HealthScore ≥ 70`（`HealthyThreshold`，`domain/analyze.go:103`）时被视为"健康"。

## 公式

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

每个惩罚值有各自的上限：

| 分类         | 最大惩罚 | 常量                              |
| ------------ | -------- | --------------------------------- |
| 复杂度       | 20       | 公式中的字面量 `20.0`             |
| 死代码       | 20       | `MaxDeadCodePenalty = 20`         |
| 代码重复     | 20       | 公式中的字面量 `20.0`             |
| 耦合度       | 20       | 公式中的字面量 `20.0`             |
| 内聚度       | 20       | 公式中的字面量 `20.0`             |
| 依赖         | 16       | `MaxDependencyPenalty = 10+3+3`   |
| 架构         | 12       | `MaxArchitecturePenalty = 12`     |

评分下限为 `MinimumScore = 0`（`domain/analyze.go:102`），在惩罚求和后应用。

## 惩罚规则详述

### 复杂度

**输入。** `AverageComplexity`（float64）。

**公式。**

```
if AverageComplexity <= 2.0:
    penalty = 0
else:
    penalty = min(20, round((AverageComplexity - 2.0) / 13.0 * 20.0))
```

**常量。**

| 名称      | 值    | 含义                                       |
| --------- | ----- | ------------------------------------------ |
| baseline  | 2.0   | 平均复杂度低于此值时惩罚为 0               |
| range     | 13.0  | 分母 — 在平均值 = 15 时达到最大惩罚        |
| max       | 20.0  | 惩罚上限                                   |

**饱和。** 在 `AverageComplexity >= 15.0` 时达到 20。

**边界情况。** `AverageComplexity = 0` 或任何 `≤ 2.0` 的值产生惩罚 0。`Validate()` 拒绝负值。

来源：`domain/analyze.go:266-279`。

### 死代码

**输入。** `CriticalDeadCode`、`WarningDeadCode`、`InfoDeadCode`（int）。`TotalFiles`（int）用于在 `CalculateHealthScore()` 中计算归一化因子。

**公式。**

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

**常量。**

| 名称                      | 值    | 含义                             |
| ------------------------- | ----- | -------------------------------- |
| critical 权重             | 1.0   | Critical 发现的完整权重          |
| warning 权重              | 0.5   | Warning 发现的半权重             |
| info 权重                 | 0.2   | Info 发现的最小权重              |
| 归一化阈值                | 10    | 文件数低于此值时归一化因子 = 1   |
| `MaxDeadCodePenalty`      | 20    | 惩罚上限                         |

**饱和。** 当 `weighted / normalization >= 20.0` 时达到 20。

**边界情况。** 零发现时惩罚为 0。截断使用 `int()`（向零截断），而非 `math.Round` — 因此计算值 `1.99` 变为 `1`。

来源：`domain/analyze.go:283-296`。归一化因子在 `CalculateHealthScore()` 的 `domain/analyze.go:474-477` 中计算。

### 代码重复

**输入。** `CodeDuplication`（float64）。此字段是预先计算的重复百分比，上游计算限制为 10。

**公式。**

```
if CodeDuplication <= 0:
    penalty = 0
else:
    penalty = min(20, round(CodeDuplication / 10.0 * 20.0))
```

**常量。**

| 名称                         | 值    | 含义                        |
| ---------------------------- | ----- | --------------------------- |
| `DuplicationThresholdLow`    | 0.0   | 0% 重复 = 0 惩罚           |
| `DuplicationThresholdHigh`   | 10.0  | 10% 重复 = 最大惩罚        |
| max                          | 20.0  | 惩罚上限                    |

**上游计算。** `CodeDuplication` 本身在 `app/analyze_usecase.go:570-583` 中计算：

```
lines_in_thousands = max(GroupDensityMinLines, total_lines / GroupDensityLinesUnit)
group_density      = clone_groups / lines_in_thousands
CodeDuplication    = min(DuplicationThresholdHigh, group_density * GroupDensityCoefficient)
```

其中 `GroupDensityLinesUnit = 1000.0`、`GroupDensityMinLines = 1.0`、`GroupDensityCoefficient = 20.0`、`DuplicationThresholdHigh = 10.0`（`domain/analyze.go:52, 62-64`）。

**饱和。** 当 `CodeDuplication >= 10.0` 时达到 20。由于上游限制为 10，饱和恰好在每 1000 行分析代码有 0.5 个克隆组时达到。

**边界情况。** 无克隆组或无分析行 → `CodeDuplication = 0` → 惩罚 0。`Validate()` 拒绝 `[0, 100]` 范围外的值。

来源：`domain/analyze.go:300-314`。

### 耦合度（CBO）

**输入。** `CBOClasses`、`HighCouplingClasses`（CBO > 7）、`MediumCouplingClasses`（3 < CBO ≤ 7）。

**公式。**

```
if CBOClasses == 0:
    penalty = 0
else:
    weighted = HighCouplingClasses + 0.5 * MediumCouplingClasses
    ratio    = weighted / CBOClasses
    penalty  = min(20, round(ratio / 0.25 * 20.0))
```

**常量。**

| 名称            | 值    | 含义                                         |
| --------------- | ----- | -------------------------------------------- |
| high 权重       | 1.0   | 高耦合类的完整权重                           |
| medium 权重     | 0.5   | 中等耦合类的半权重                           |
| 饱和比率        | 0.25  | 25% 加权问题类 → 最大惩罚                    |
| max             | 20.0  | 惩罚上限                                     |

**饱和。** 当加权比率 `≥ 0.25` 时达到 20。

**边界情况。** `CBOClasses = 0` → 惩罚 0。`Validate()` 确保 high + medium ≤ total。

来源：`domain/analyze.go:318-336`。

### 内聚度（LCOM）

**输入。** `LCOMClasses`、`HighLCOMClasses`（LCOM4 > 5）、`MediumLCOMClasses`（2 < LCOM4 ≤ 5）。

**公式。**

```
if LCOMClasses == 0:
    penalty = 0
else:
    weighted = HighLCOMClasses + 0.5 * MediumLCOMClasses
    ratio    = weighted / LCOMClasses
    penalty  = min(20, round(ratio / 0.30 * 20.0))
```

**常量。**

| 名称             | 值    | 含义                                        |
| ---------------- | ----- | ------------------------------------------- |
| high 权重        | 1.0   | 高 LCOM 类的完整权重                        |
| medium 权重      | 0.5   | 中等 LCOM 类的半权重                        |
| 饱和比率         | 0.30  | 30% 加权问题类 → 最大惩罚                   |
| max              | 20.0  | 惩罚上限                                    |

**饱和。** 当加权比率 `≥ 0.30` 时达到 20。

**边界情况。** `LCOMClasses = 0` → 惩罚 0。`Validate()` 确保 high + medium ≤ total。

来源：`domain/analyze.go:340-358`。

### 依赖

**输入。** `DepsEnabled`（bool）、`DepsTotalModules`、`DepsModulesInCycles`、`DepsMaxDepth`（int）、`DepsMainSequenceDeviation`（float64，0–1）。

**公式。** 三个子惩罚求和：

```
if not DepsEnabled:
    return 0

# 循环子惩罚 (max 10)
if DepsTotalModules > 0:
    ratio  = clamp(0, 1, DepsModulesInCycles / DepsTotalModules)
    cycles = round(10 * ratio)
else:
    cycles = 0

# 深度子惩罚 (max 3)
if DepsTotalModules > 0:
    expected = max(3, ceil(log2(DepsTotalModules + 1)) + 1)
    depth    = clamp(0, 3, DepsMaxDepth - expected)
else:
    depth = 0

# 主序列偏差子惩罚 (max 3)
if DepsMainSequenceDeviation > 0:
    msd = round(3 * clamp(0, 1, DepsMainSequenceDeviation))
else:
    msd = 0

penalty = cycles + depth + msd     # max 16
```

**常量。**

| 名称                 | 值    | 含义                            |
| -------------------- | ----- | ------------------------------- |
| `MaxCyclesPenalty`   | 10    | 循环子惩罚上限                  |
| `MaxDepthPenalty`    | 3     | 深度子惩罚上限                  |
| `MaxMSDPenalty`      | 3     | 主序列偏差子惩罚上限            |
| `MaxDependencyPenalty` | 16  | 三个上限之和                    |

**饱和。** 当每个模块都处于循环中、深度超出预期 ≥ 3、且 MSD ≥ 1 时达到 16。

**边界情况。** `DepsEnabled = false` → 惩罚 0。`DepsTotalModules = 0` → 循环和深度贡献 0；仅 MSD 可能生效。`Validate()` 强制 `MSD ∈ [0, 1]` 且 `DepsModulesInCycles ≤ DepsTotalModules`。

来源：`domain/analyze.go:361-406`。

### 架构

**输入。** `ArchEnabled`（bool）、`ArchCompliance`（float64，0–1）。

**公式。**

```
if not ArchEnabled:
    penalty = 0
else:
    penalty = round(12 * (1 - clamp(0, 1, ArchCompliance)))
```

**常量。**

| 名称                     | 值    | 含义           |
| ------------------------ | ----- | -------------- |
| `MaxArchPenalty`         | 12    | 惩罚上限       |
| `MaxArchitecturePenalty` | 12    | 上限的别名     |

**饱和。** 在 `ArchCompliance = 0.0` 时达到 12。

**边界情况。** `ArchEnabled = false` → 惩罚 0。`Validate()` 在启用时强制 `ArchCompliance ∈ [0, 1]`。

来源：`domain/analyze.go:409-422`。

## 分类评分

每个分类在报告中展示一个 0–100 的评分。转换方式取决于该分类的惩罚尺度。

**大多数分类（复杂度、死代码、代码重复、耦合度、内聚度）。** 它们的惩罚上限均为 20（`MaxScoreBase = 20`）。评分通过 `penaltyToScore(penalty, 20)` 计算：

```
score = 100 - round(penalty * 100 / 20) = 100 - penalty * 5
```

即每单位惩罚扣 5 分。惩罚为 0 时评分为 100；惩罚为 20 时评分为 0。

**依赖。** 惩罚上限为 16，因此评分需先通过 `normalizeToScoreBase` 归一化到 20 分制：

```
normalized = round(dependencyPenalty / 16 * 20)
DependencyScore = 100 - round(normalized * 100 / 20) = 100 - normalized * 5
```

**架构。** 特殊情况 — 评分直接取自合规度：

```
ArchitectureScore = round(ArchCompliance * 100)
```

因此 `ArchCompliance = 0.98` 产生 `ArchitectureScore = 98`，无论计入总体评分的惩罚如何舍入。

来源：`domain/analyze.go:426-453`、`domain/analyze.go:483-510`。

## 计算示例

输入：

| 字段                          | 值     |
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

**复杂度惩罚。** `(8.0 − 2.0) / 13.0 × 20.0 = 9.2308 → round → 9`。

**死代码惩罚。** `weighted = 2×1.0 + 1×0.5 + 0×0.2 = 2.5`。`normalization = 1.0 + log10(50/10) = 1.0 + log10(5) ≈ 1.6990`。`2.5 / 1.6990 ≈ 1.4715`。`int(min(20.0, 1.4715)) = 1`。

**代码重复惩罚。** `7.5 / 10.0 × 20.0 = 15.0 → round → 15`。

**耦合度惩罚。** `weighted = 3 + 0.5×2 = 4`。`ratio = 4/20 = 0.20`。`0.20 / 0.25 × 20.0 = 16.0 → round → 16`。

**内聚度惩罚。** `weighted = 1 + 0.5×3 = 2.5`。`ratio = 2.5/20 = 0.125`。`0.125 / 0.30 × 20.0 ≈ 8.333 → round → 8`。

**依赖惩罚。**
- 循环：`ratio = 1/8 = 0.125`。`round(10 × 0.125) = round(1.25) = 1`。
- 深度：`expected = max(3, ceil(log2(9)) + 1) = max(3, 4 + 1) = 5`。超出 = `5 − 5 = 0`。
- MSD：`round(3 × 0.2) = round(0.6) = 1`。
- 总计：`1 + 0 + 1 = 2`。

**架构惩罚。** `round(12 × (1 − 0.85)) = round(1.8) = 2`。

**总和。** `9 + 1 + 15 + 16 + 8 + 2 + 2 = 53`。

**HealthScore。** `max(0, 100 − 53) = 47`。等级：`47 ≥ 45` → **D**。

**分类评分（报告中显示）。**

| 分类         | 惩罚 | 评分                                 |
| ------------ | ------- | ------------------------------------ |
| 复杂度       | 9       | `100 − 9×5 = 55`                     |
| 死代码       | 1       | `100 − 1×5 = 95`                     |
| 代码重复     | 15      | `100 − 15×5 = 25`                    |
| 耦合度       | 16      | `100 − 16×5 = 20`                    |
| 内聚度       | 8       | `100 − 8×5 = 60`                     |
| 依赖         | 2       | `normalized = round(2/16 × 20) = 3`; `100 − 3×5 = 85` |
| 架构         | —       | `round(0.85 × 100) = 85`             |

## 降级评分

`CalculateHealthScore()` 首先对摘要调用 `Validate()`。如果验证失败 — 例如 `AverageComplexity < 0`、`CodeDuplication` 超出 `[0, 100]` 范围、启用时 `ArchCompliance` 超出 `[0, 1]` 范围、启用时 `DepsMainSequenceDeviation` 超出 `[0, 1]` 范围、或 LCOM 或 CBO 的 high + medium 类数之和超过总数 — 摘要的评分将被置零，等级设为 `"N/A"`，并返回错误。调用方可以随后调用 `CalculateFallbackScore()` 作为降级路径：从 100 开始，如果 `AverageComplexity > 10` 则减去 `FallbackComplexityThreshold = 10`，对 `DeadCodeCount > 0`、`HighComplexityCount > 0` 和 `HighLCOMClasses > 0` 各减去 `FallbackPenalty = 5`，下限为 `MinimumScore = 0`。

来源：`domain/analyze.go:200-262`（`Validate`）、`domain/analyze.go:456-470`（`CalculateHealthScore` 中的验证分支）、`domain/analyze.go:538-566`（`CalculateFallbackScore`）。

## 舍入

所有非整数中间值使用 Go 的 `math.Round` 归约为整数，对于精确的 `.5` 值使用银行家舍入（Go 实现中正数使用四舍五入）。最终健康评分通过 `MinimumScore` 限制在 `[0, 100]` 范围内。分类评分在 `penaltyToScore` 中限制在 `[0, 100]`（`domain/analyze.go:441-453`）。`math.Round` 的唯一例外是死代码惩罚，它在 `math.Min` 之后使用 `int()` 截断（`domain/analyze.go:294`）。

## 随时间追踪

公式是确定性的 — 相同的输入在不同运行和平台上始终产生相同的评分。要追踪健康状况随时间的变化，请在每次提交时从 JSON 输出中持久化 `summary.health_score` 并在 CI 中进行比较。分类评分（`complexity_score`、`dead_code_score` 等）同样稳定，可以单独追踪。

## 参考

所有行号引用 `domain/analyze.go`：

- `CalculateHealthScore()` — 第 456–534 行
- `calculateComplexityPenalty()` — 第 266–279 行
- `calculateDeadCodePenalty(normalizationFactor)` — 第 283–296 行
- `calculateDuplicationPenalty()` — 第 300–314 行
- `calculateCouplingPenalty()` — 第 318–336 行
- `calculateCohesionPenalty()` — 第 340–358 行
- `calculateDependencyPenalty()` — 第 361–406 行
- `calculateArchitecturePenalty()` — 第 409–422 行
- `CalculateFallbackScore()` — 第 538–566 行
- `Validate()` — 第 200–262 行
- `penaltyToScore()` — 第 441–453 行
- `normalizeToScoreBase()` — 第 426–438 行
- `GetGradeFromScore()` — 第 569–582 行

上游 `CodeDuplication` 计算：`app/analyze_usecase.go:570-583`。
