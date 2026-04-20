# ヘルススコア

## 概要

ヘルススコアは、有効なすべての分析（複雑度、デッドコード、重複、結合度、凝集度、依存関係、アーキテクチャ）を単一の値にまとめた 0〜100 の整数です。計算は純粋で決定論的です。同じ `AnalyzeSummary` 入力が与えられれば、`CalculateHealthScore()` は常に同じスコアを生成します。CI でコミットごとにスコアを保存し、経時的な変化を追跡するために設計されています。

## グレード

スコアは厳密な `≥` 閾値を使用してレターグレードにマッピングされます:

| グレード | スコア範囲 | 閾値定数 |
| ----- | ----------- | ------------------ |
| A     | 90〜100      | `GradeAThreshold = 90` |
| B     | 75〜89       | `GradeBThreshold = 75` |
| C     | 60〜74       | `GradeCThreshold = 60` |
| D     | 45〜59       | `GradeDThreshold = 45` |
| F     | 0〜44        | （D 未満）              |

`domain/analyze.go:90-93` で定義されています。`HealthScore ≥ 70`（`HealthyThreshold`, `domain/analyze.go:103`）のプロジェクトが「健全」とみなされます。

## 計算式

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

各ペナルティには個別の上限があります:

| カテゴリ     | 最大ペナルティ | 定数                          |
| ------------ | ----------- | --------------------------------- |
| 複雑度   | 20          | 計算式内のリテラル `20.0`         |
| デッドコード    | 20          | `MaxDeadCodePenalty = 20`         |
| 重複  | 20          | 計算式内のリテラル `20.0`         |
| 結合度     | 20          | 計算式内のリテラル `20.0`         |
| 凝集度     | 20          | 計算式内のリテラル `20.0`         |
| 依存関係 | 16          | `MaxDependencyPenalty = 10+3+3`   |
| アーキテクチャ | 12          | `MaxArchitecturePenalty = 12`     |

スコアの下限は `MinimumScore = 0`（`domain/analyze.go:102`）で、ペナルティ合算後に適用されます。

## ペナルティの仕様

### 複雑度

**入力.** `AverageComplexity` (float64)。

**計算式.**

```
if AverageComplexity <= 2.0:
    penalty = 0
else:
    penalty = min(20, round((AverageComplexity - 2.0) / 13.0 * 20.0))
```

**定数.**

| 名前      | 値 | 意味                                    |
| --------- | ----- | ------------------------------------------ |
| baseline  | 2.0   | この値以下の平均複雑度ではペナルティ = 0 |
| range     | 13.0  | 分母 — 平均 = 15 で最大ペナルティに到達 |
| max       | 20.0  | ペナルティ上限                                |

**飽和.** `AverageComplexity >= 15.0` で 20 に到達します。

**エッジケース.** `AverageComplexity = 0` または `≤ 2.0` のすべての値でペナルティは 0 です。`Validate()` は負の値を拒否します。

ソース: `domain/analyze.go:266-279`。

### デッドコード

**入力.** `CriticalDeadCode`, `WarningDeadCode`, `InfoDeadCode` (int)。`TotalFiles` (int) は `CalculateHealthScore()` 内で正規化係数の導出に使用されます。

**計算式.**

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

**定数.**

| 名前                      | 値 | 意味                          |
| ------------------------- | ----- | -------------------------------- |
| critical の重み           | 1.0   | Critical 検出結果のフルウェイト |
| warning の重み            | 0.5   | Warning 検出結果のハーフウェイト  |
| info の重み               | 0.2   | Info 検出結果の最小ウェイト  |
| 正規化閾値   | 10    | この値未満のファイル数では正規化係数 = 1 |
| `MaxDeadCodePenalty`      | 20    | ペナルティ上限                       |

**飽和.** `weighted / normalization >= 20.0` のとき 20 に到達します。

**エッジケース.** 検出結果がゼロの場合、ペナルティは 0 です。切り捨ては `int()`（ゼロ方向）であり、`math.Round` ではありません。計算値 `1.99` は `1` になります。

ソース: `domain/analyze.go:283-296`。正規化係数は `CalculateHealthScore()` の `domain/analyze.go:474-477` で導出されます。

### 重複

**入力.** `CodeDuplication` (float64)。このフィールドは事前計算された重複率で、上流の計算で 10 に上限が設定されています。

**計算式.**

```
if CodeDuplication <= 0:
    penalty = 0
else:
    penalty = min(20, round(CodeDuplication / 10.0 * 20.0))
```

**定数.**

| 名前                         | 値 | 意味                     |
| ---------------------------- | ----- | --------------------------- |
| `DuplicationThresholdLow`    | 0.0   | 重複 0% = ペナルティ 0  |
| `DuplicationThresholdHigh`   | 10.0  | 重複 10% = 最大ペナルティ |
| max                          | 20.0  | ペナルティ上限                 |

**上流の計算.** `CodeDuplication` 自体は `app/analyze_usecase.go:570-583` で計算されます:

```
lines_in_thousands = max(GroupDensityMinLines, total_lines / GroupDensityLinesUnit)
group_density      = clone_groups / lines_in_thousands
CodeDuplication    = min(DuplicationThresholdHigh, group_density * GroupDensityCoefficient)
```

ここで `GroupDensityLinesUnit = 1000.0`, `GroupDensityMinLines = 1.0`, `GroupDensityCoefficient = 20.0`, `DuplicationThresholdHigh = 10.0`（`domain/analyze.go:52, 62-64`）です。

**飽和.** `CodeDuplication >= 10.0` で 20 に到達します。上流が 10 で上限を設定するため、分析対象 1000 行あたり 0.5 クローングループで飽和に達します。

**エッジケース.** クローングループなしまたは分析対象行なしの場合 → `CodeDuplication = 0` → ペナルティ 0。`Validate()` は `[0, 100]` の範囲外の値を拒否します。

ソース: `domain/analyze.go:300-314`。

### 結合度（CBO）

**入力.** `CBOClasses`, `HighCouplingClasses`（CBO > 7）, `MediumCouplingClasses`（3 < CBO ≤ 7）。

**計算式.**

```
if CBOClasses == 0:
    penalty = 0
else:
    weighted = HighCouplingClasses + 0.5 * MediumCouplingClasses
    ratio    = weighted / CBOClasses
    penalty  = min(20, round(ratio / 0.25 * 20.0))
```

**定数.**

| 名前            | 値 | 意味                                          |
| --------------- | ----- | ------------------------------------------------ |
| high の重み     | 1.0   | 高結合度クラスのフルウェイト            |
| medium の重み   | 0.5   | 中結合度クラスのハーフウェイト          |
| 飽和比率 | 0.25 | 重み付き問題クラスが 25% で最大ペナルティ   |
| max             | 20.0  | ペナルティ上限                                      |

**飽和.** 重み付き比率が `≥ 0.25` のとき 20 に到達します。

**エッジケース.** `CBOClasses = 0` → ペナルティ 0。`Validate()` は high + medium ≤ total を保証します。

ソース: `domain/analyze.go:318-336`。

### 凝集度（LCOM）

**入力.** `LCOMClasses`, `HighLCOMClasses`（LCOM4 > 5）, `MediumLCOMClasses`（2 < LCOM4 ≤ 5）。

**計算式.**

```
if LCOMClasses == 0:
    penalty = 0
else:
    weighted = HighLCOMClasses + 0.5 * MediumLCOMClasses
    ratio    = weighted / LCOMClasses
    penalty  = min(20, round(ratio / 0.30 * 20.0))
```

**定数.**

| 名前             | 値 | 意味                                         |
| ---------------- | ----- | ----------------------------------------------- |
| high の重み      | 1.0   | 高 LCOM クラスのフルウェイト               |
| medium の重み    | 0.5   | 中 LCOM クラスのハーフウェイト             |
| 飽和比率 | 0.30  | 重み付き問題クラスが 30% で最大ペナルティ  |
| max              | 20.0  | ペナルティ上限                                     |

**飽和.** 重み付き比率が `≥ 0.30` のとき 20 に到達します。

**エッジケース.** `LCOMClasses = 0` → ペナルティ 0。`Validate()` は high + medium ≤ total を保証します。

ソース: `domain/analyze.go:340-358`。

### 依存関係

**入力.** `DepsEnabled` (bool), `DepsTotalModules`, `DepsModulesInCycles`, `DepsMaxDepth` (int), `DepsMainSequenceDeviation` (float64, 0〜1)。

**計算式.** 3つのサブペナルティの合計:

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

**定数.**

| 名前                 | 値 | 意味                         |
| -------------------- | ----- | ------------------------------- |
| `MaxCyclesPenalty`   | 10    | 循環サブペナルティの上限      |
| `MaxDepthPenalty`    | 3     | 深度サブペナルティの上限       |
| `MaxMSDPenalty`      | 3     | MSD サブペナルティの上限         |
| `MaxDependencyPenalty` | 16  | 3つの上限の合計           |

**飽和.** すべてのモジュールが循環に含まれ、深度が期待値を 3 以上超過し、MSD ≥ 1 のとき 16 に到達します。

**エッジケース.** `DepsEnabled = false` → ペナルティ 0。`DepsTotalModules = 0` → 循環と深度は 0 に寄与し、MSD のみが影響します。`Validate()` は `MSD ∈ [0, 1]` および `DepsModulesInCycles ≤ DepsTotalModules` を保証します。

ソース: `domain/analyze.go:361-406`。

### アーキテクチャ

**入力.** `ArchEnabled` (bool), `ArchCompliance` (float64, 0〜1)。

**計算式.**

```
if not ArchEnabled:
    penalty = 0
else:
    penalty = round(12 * (1 - clamp(0, 1, ArchCompliance)))
```

**定数.**

| 名前                     | 値 | 意味            |
| ------------------------ | ----- | ------------------ |
| `MaxArchPenalty`         | 12    | ペナルティ上限        |
| `MaxArchitecturePenalty` | 12    | 上限のエイリアス  |

**飽和.** `ArchCompliance = 0.0` で 12 に到達します。

**エッジケース.** `ArchEnabled = false` → ペナルティ 0。`Validate()` は有効時に `ArchCompliance ∈ [0, 1]` を保証します。

ソース: `domain/analyze.go:409-422`。

## カテゴリスコア

各カテゴリはレポートで 0〜100 のスコアを公開します。変換はカテゴリのペナルティスケールに依存します。

**ほとんどのカテゴリ（複雑度、デッドコード、重複、結合度、凝集度）.** これらはすべてペナルティ上限 20（`MaxScoreBase = 20`）を持ちます。スコアは `penaltyToScore(penalty, 20)` で計算されます:

```
score = 100 - round(penalty * 100 / 20) = 100 - penalty * 5
```

ペナルティ 1 単位あたり 5 スコアポイントが差し引かれます。ペナルティ 0 はスコア 100、ペナルティ 20 はスコア 0 になります。

**依存関係.** ペナルティ上限は 16 なので、`normalizeToScoreBase` で 20 ポイントスケールに正規化されます:

```
normalized = round(dependencyPenalty / 16 * 20)
DependencyScore = 100 - round(normalized * 100 / 20) = 100 - normalized * 5
```

**アーキテクチャ.** 特殊ケース — スコアはコンプライアンスから直接取得されます:

```
ArchitectureScore = round(ArchCompliance * 100)
```

`ArchCompliance = 0.98` は `ArchitectureScore = 98` になり、全体スコアに影響するペナルティの丸めとは無関係です。

ソース: `domain/analyze.go:426-453`, `domain/analyze.go:483-510`。

## 計算例

入力:

| フィールド                         | 値  |
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

**複雑度ペナルティ.** `(8.0 − 2.0) / 13.0 × 20.0 = 9.2308 → round → 9`。

**デッドコードペナルティ.** `weighted = 2×1.0 + 1×0.5 + 0×0.2 = 2.5`。`normalization = 1.0 + log10(50/10) = 1.0 + log10(5) ≈ 1.6990`。`2.5 / 1.6990 ≈ 1.4715`。`int(min(20.0, 1.4715)) = 1`。

**重複ペナルティ.** `7.5 / 10.0 × 20.0 = 15.0 → round → 15`。

**結合度ペナルティ.** `weighted = 3 + 0.5×2 = 4`。`ratio = 4/20 = 0.20`。`0.20 / 0.25 × 20.0 = 16.0 → round → 16`。

**凝集度ペナルティ.** `weighted = 1 + 0.5×3 = 2.5`。`ratio = 2.5/20 = 0.125`。`0.125 / 0.30 × 20.0 ≈ 8.333 → round → 8`。

**依存関係ペナルティ.**
- 循環: `ratio = 1/8 = 0.125`。`round(10 × 0.125) = round(1.25) = 1`。
- 深度: `expected = max(3, ceil(log2(9)) + 1) = max(3, 4 + 1) = 5`。超過 = `5 − 5 = 0`。
- MSD: `round(3 × 0.2) = round(0.6) = 1`。
- 合計: `1 + 0 + 1 = 2`。

**アーキテクチャペナルティ.** `round(12 × (1 − 0.85)) = round(1.8) = 2`。

**合計.** `9 + 1 + 15 + 16 + 8 + 2 + 2 = 53`。

**HealthScore.** `max(0, 100 − 53) = 47`。グレード: `47 ≥ 45` → **D**。

**カテゴリスコア（レポート上の表示）.**

| カテゴリ     | ペナルティ | スコア                                |
| ------------ | ------- | ------------------------------------ |
| 複雑度   | 9       | `100 − 9×5 = 55`                     |
| デッドコード    | 1       | `100 − 1×5 = 95`                     |
| 重複  | 15      | `100 − 15×5 = 25`                    |
| 結合度     | 16      | `100 − 16×5 = 20`                    |
| 凝集度     | 8       | `100 − 8×5 = 60`                     |
| 依存関係 | 2       | `normalized = round(2/16 × 20) = 3`; `100 − 3×5 = 85` |
| アーキテクチャ | —       | `round(0.85 × 100) = 85`             |

## フォールバックスコア

`CalculateHealthScore()` はまず summary に対して `Validate()` を呼び出します。バリデーションが失敗した場合（例: `AverageComplexity < 0`、`CodeDuplication` が `[0, 100]` の範囲外、有効時に `ArchCompliance` が `[0, 1]` の範囲外、有効時に `DepsMainSequenceDeviation` が `[0, 1]` の範囲外、LCOM または CBO で high + medium クラスの合計が total を超過）、summary のスコアはゼロにリセットされ、グレードは `"N/A"` に設定され、エラーが返されます。呼び出し元はその後、劣化パスとして `CalculateFallbackScore()` を呼び出すことができます。100 から開始し、`AverageComplexity > 10` の場合に `FallbackComplexityThreshold = 10` を減算し、`DeadCodeCount > 0`、`HighComplexityCount > 0`、`HighLCOMClasses > 0` それぞれに対して `FallbackPenalty = 5` を減算し、`MinimumScore = 0` で下限を設定します。

ソース: `domain/analyze.go:200-262`（`Validate`）, `domain/analyze.go:456-470`（`CalculateHealthScore` 内のバリデーション分岐）, `domain/analyze.go:538-566`（`CalculateFallbackScore`）。

## 丸め処理

すべての非整数中間値は Go の `math.Round` を使用して整数に変換されます。これは正の値に対して四捨五入（正確な `.5` 値では最近接偶数丸め）を適用します。最終的なヘルススコアは `MinimumScore` により `[0, 100]` にクランプされます。カテゴリスコアは `penaltyToScore`（`domain/analyze.go:441-453`）内で `[0, 100]` にクランプされます。`math.Round` の唯一の例外はデッドコードペナルティで、`math.Min` の後に `int()` による切り捨てを使用します（`domain/analyze.go:294`）。

## 経時的な追跡

計算式は決定論的であり、同一の入力は実行やプラットフォームにかかわらず常に同一のスコアを生成します。ヘルスを経時的に追跡するには、JSON 出力の `summary.health_score` をコミットごとに保存し、CI で比較してください。カテゴリごとのスコア（`complexity_score`, `dead_code_score` など）も同様に安定しており、個別に追跡できます。

## リファレンス

すべての行番号は `domain/analyze.go` を参照しています:

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

上流の `CodeDuplication` 計算: `app/analyze_usecase.go:570-583`。
