# Score de santé

## Vue d'ensemble

Le score de santé est un entier sur 0–100 qui synthétise en une seule valeur toutes les analyses activées (complexité, code mort, duplication, couplage, cohésion, dépendances, architecture). Le calcul est pur et déterministe — avec les mêmes entrées `AnalyzeSummary`, `CalculateHealthScore()` produit toujours le même score. Il est conçu pour être stocké par commit dans la CI afin de suivre les évolutions au fil du temps.

## Note

Le score est converti en une note littérale à l'aide de seuils stricts `≥` :

| Note | Plage de score | Constante de seuil   |
| ---- | -------------- | -------------------- |
| A    | 90–100         | `GradeAThreshold = 90` |
| B    | 75–89          | `GradeBThreshold = 75` |
| C    | 60–74          | `GradeCThreshold = 60` |
| D    | 45–59          | `GradeDThreshold = 45` |
| F    | 0–44           | (sous D)               |

Définies dans `domain/analyze.go:90-93`. Un projet est considéré « en bonne santé » à `HealthScore ≥ 70` (`HealthyThreshold`, `domain/analyze.go:103`).

## Formule

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

Chaque pénalité est plafonnée à son propre maximum :

| Catégorie    | Pénalité max | Constante                       |
| ------------ | ------------ | ------------------------------- |
| Complexité   | 20           | littéral `20.0` dans la formule |
| Code mort    | 20           | `MaxDeadCodePenalty = 20`       |
| Duplication  | 20           | littéral `20.0` dans la formule |
| Couplage     | 20           | littéral `20.0` dans la formule |
| Cohésion     | 20           | littéral `20.0` dans la formule |
| Dépendances  | 16           | `MaxDependencyPenalty = 10+3+3` |
| Architecture | 12           | `MaxArchitecturePenalty = 12`   |

Le plancher du score est `MinimumScore = 0` (`domain/analyze.go:102`), appliqué après la sommation des pénalités.

## Spécifications des pénalités

### Complexité

**Entrées.** `AverageComplexity` (float64).

**Formule.**

```
if AverageComplexity <= 2.0:
    penalty = 0
else:
    penalty = min(20, round((AverageComplexity - 2.0) / 13.0 * 20.0))
```

**Constantes.**

| Nom       | Valeur | Signification                                   |
| --------- | ------ | ----------------------------------------------- |
| baseline  | 2.0    | Complexité moyenne en dessous de laquelle pénalité = 0 |
| range     | 13.0   | Dénominateur — pénalité max atteinte à moyenne = 15 |
| max       | 20.0   | Plafond de pénalité                             |

**Saturation.** Atteint 20 à `AverageComplexity >= 15.0`.

**Cas limites.** `AverageComplexity = 0` ou toute valeur `≤ 2.0` donne une pénalité de 0. `Validate()` rejette les valeurs négatives.

Source : `domain/analyze.go:266-279`.

### Code mort

**Entrées.** `CriticalDeadCode`, `WarningDeadCode`, `InfoDeadCode` (int). `TotalFiles` (int) sert à dériver le facteur de normalisation dans `CalculateHealthScore()`.

**Formule.**

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

**Constantes.**

| Nom                       | Valeur | Signification                          |
| ------------------------- | ------ | -------------------------------------- |
| poids critical            | 1.0    | Poids complet pour les constats Critical |
| poids warning             | 0.5    | Demi-poids pour les constats Warning   |
| poids info                | 0.2    | Poids minimal pour les constats Info   |
| seuil de normalisation    | 10     | Fichiers en dessous desquels le facteur de normalisation = 1 |
| `MaxDeadCodePenalty`      | 20     | Plafond de pénalité                    |

**Saturation.** Atteint 20 lorsque `weighted / normalization >= 20.0`.

**Cas limites.** Zéro constat donne une pénalité de 0. La troncature utilise `int()` (vers zéro), pas `math.Round` — ainsi une valeur calculée de `1.99` devient `1`.

Source : `domain/analyze.go:283-296`. Facteur de normalisation dérivé dans `CalculateHealthScore()` à `domain/analyze.go:474-477`.

### Duplication

**Entrées.** `CodeDuplication` (float64). Ce champ est le ratio de fragments : le pourcentage de tous les fragments de code extraits qui participent à une paire ou un groupe de clones, plafonné à 30 par le calcul en amont.

**Formule.**

```
if CodeDuplication <= 0:
    penalty = 0
else:
    penalty = min(20, round(CodeDuplication / 30.0 * 20.0))
```

**Constantes.**

| Nom                          | Valeur | Signification                       |
| ---------------------------- | ------ | ----------------------------------- |
| `DuplicationThresholdLow`    | 0.0    | 0 % de fragments clonés = pénalité 0 |
| `DuplicationThresholdHigh`   | 30.0   | 30 % de fragments clonés = pénalité max |
| max                          | 20.0   | Plafond de pénalité                 |

**Calcul en amont.** `CodeDuplication` est calculé dans `app/analyze_usecase.go` :

```
CodeDuplication = min(DuplicationThresholdHigh, TotalClones / TotalFragments * 100)
```

Où `TotalClones` est le nombre de fragments uniques participant à au moins une paire ou un groupe de clones, et `TotalFragments` est le nombre total de fragments de code extraits (`domain/clone.go:128-131`).

**Saturation.** Atteint 20 lorsque `CodeDuplication >= 30.0`.

**Cas limites.** Aucun fragment cloné ou aucun fragment analysé → `CodeDuplication = 0` → pénalité 0. `Validate()` rejette les valeurs hors de `[0, 100]`.

Source : `domain/analyze.go:336-350`.

### Couplage (CBO)

**Entrées.** `CBOClasses`, `HighCouplingClasses` (CBO > 7), `MediumCouplingClasses` (3 < CBO ≤ 7).

**Formule.**

```
if CBOClasses == 0:
    penalty = 0
else:
    weighted = HighCouplingClasses + 0.5 * MediumCouplingClasses
    ratio    = weighted / CBOClasses
    penalty  = min(20, round(ratio / 0.25 * 20.0))
```

**Constantes.**

| Nom              | Valeur | Signification                                       |
| ---------------- | ------ | --------------------------------------------------- |
| poids haut       | 1.0    | Poids complet pour les classes à haut couplage      |
| poids moyen      | 0.5    | Demi-poids pour les classes à couplage moyen        |
| ratio saturation | 0.25   | 25 % de classes problématiques pondérées → pénalité max |
| max              | 20.0   | Plafond de pénalité                                 |

**Saturation.** Atteint 20 lorsque le ratio pondéré est `≥ 0.25`.

**Cas limites.** `CBOClasses = 0` → pénalité 0. `Validate()` garantit que haut + moyen ≤ total.

Source : `domain/analyze.go:318-336`.

### Cohésion (LCOM)

**Entrées.** `LCOMClasses`, `HighLCOMClasses` (LCOM4 > 5), `MediumLCOMClasses` (2 < LCOM4 ≤ 5).

**Formule.**

```
if LCOMClasses == 0:
    penalty = 0
else:
    weighted = HighLCOMClasses + 0.5 * MediumLCOMClasses
    ratio    = weighted / LCOMClasses
    penalty  = min(20, round(ratio / 0.30 * 20.0))
```

**Constantes.**

| Nom              | Valeur | Signification                                      |
| ---------------- | ------ | -------------------------------------------------- |
| poids haut       | 1.0    | Poids complet pour les classes à LCOM élevé        |
| poids moyen      | 0.5    | Demi-poids pour les classes à LCOM moyen           |
| ratio saturation | 0.30   | 30 % de classes problématiques pondérées → pénalité max |
| max              | 20.0   | Plafond de pénalité                                |

**Saturation.** Atteint 20 lorsque le ratio pondéré est `≥ 0.30`.

**Cas limites.** `LCOMClasses = 0` → pénalité 0. `Validate()` garantit que haut + moyen ≤ total.

Source : `domain/analyze.go:340-358`.

### Dépendances

**Entrées.** `DepsEnabled` (bool), `DepsTotalModules`, `DepsModulesInCycles`, `DepsMaxDepth` (int), `DepsMainSequenceDeviation` (float64, 0–1).

**Formule.** Somme de trois sous-pénalités :

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

**Constantes.**

| Nom                    | Valeur | Signification                       |
| ---------------------- | ------ | ----------------------------------- |
| `MaxCyclesPenalty`     | 10     | Plafond pour la sous-pénalité cycles |
| `MaxDepthPenalty`      | 3      | Plafond pour la sous-pénalité profondeur |
| `MaxMSDPenalty`        | 3      | Plafond pour la sous-pénalité MSD   |
| `MaxDependencyPenalty` | 16     | Somme des trois plafonds            |

**Saturation.** Atteint 16 lorsque chaque module est dans un cycle, que la profondeur dépasse l'attendu de ≥ 3, et que MSD ≥ 1.

**Cas limites.** `DepsEnabled = false` → pénalité 0. `DepsTotalModules = 0` → cycles et profondeur contribuent 0 ; seul MSD peut s'activer. `Validate()` impose `MSD ∈ [0, 1]` et `DepsModulesInCycles ≤ DepsTotalModules`.

Source : `domain/analyze.go:361-406`.

### Architecture

**Entrées.** `ArchEnabled` (bool), `ArchCompliance` (float64, 0–1).

**Formule.**

```
if not ArchEnabled:
    penalty = 0
else:
    penalty = round(12 * (1 - clamp(0, 1, ArchCompliance)))
```

**Constantes.**

| Nom                      | Valeur | Signification          |
| ------------------------ | ------ | ---------------------- |
| `MaxArchPenalty`         | 12     | Plafond de pénalité    |
| `MaxArchitecturePenalty` | 12     | Alias du plafond       |

**Saturation.** Atteint 12 à `ArchCompliance = 0.0`.

**Cas limites.** `ArchEnabled = false` → pénalité 0. `Validate()` impose `ArchCompliance ∈ [0, 1]` lorsqu'activé.

Source : `domain/analyze.go:409-422`.

## Scores par catégorie

Chaque catégorie expose un score sur 0–100 dans le rapport. La conversion dépend de l'échelle de pénalité de la catégorie.

**La plupart des catégories (Complexité, Code mort, Duplication, Couplage, Cohésion).** Toutes ont une pénalité plafonnée à 20 (`MaxScoreBase = 20`). Le score est calculé par `penaltyToScore(penalty, 20)` :

```
score = 100 - round(penalty * 100 / 20) = 100 - penalty * 5
```

Effectivement, chaque unité de pénalité coûte 5 points de score. Une pénalité de 0 donne 100 ; une pénalité de 20 donne 0.

**Dépendances.** Le plafond de pénalité est 16, donc le score est d'abord normalisé sur l'échelle 20 points via `normalizeToScoreBase` :

```
normalized = round(dependencyPenalty / 16 * 20)
DependencyScore = 100 - round(normalized * 100 / 20) = 100 - normalized * 5
```

**Architecture.** Cas particulier — le score est pris directement à partir de la conformité :

```
ArchitectureScore = round(ArchCompliance * 100)
```

Ainsi `ArchCompliance = 0.98` donne `ArchitectureScore = 98`, indépendamment des arrondis de pénalité qui alimentent le score global.

Source : `domain/analyze.go:426-453`, `domain/analyze.go:483-510`.

## Exemple détaillé

Entrées :

| Champ                         | Valeur |
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

**Pénalité de complexité.** `(8.0 − 2.0) / 13.0 × 20.0 = 9.2308 → arrondi → 9`.

**Pénalité de code mort.** `weighted = 2×1.0 + 1×0.5 + 0×0.2 = 2.5`. `normalization = 1.0 + log10(50/10) = 1.0 + log10(5) ≈ 1.6990`. `2.5 / 1.6990 ≈ 1.4715`. `int(min(20.0, 1.4715)) = 1`.

**Pénalité de duplication.** `7.5 / 10.0 × 20.0 = 15.0 → arrondi → 15`.

**Pénalité de couplage.** `weighted = 3 + 0.5×2 = 4`. `ratio = 4/20 = 0.20`. `0.20 / 0.25 × 20.0 = 16.0 → arrondi → 16`.

**Pénalité de cohésion.** `weighted = 1 + 0.5×3 = 2.5`. `ratio = 2.5/20 = 0.125`. `0.125 / 0.30 × 20.0 ≈ 8.333 → arrondi → 8`.

**Pénalité de dépendances.**
- Cycles : `ratio = 1/8 = 0.125`. `round(10 × 0.125) = round(1.25) = 1`.
- Profondeur : `expected = max(3, ceil(log2(9)) + 1) = max(3, 4 + 1) = 5`. Excédent = `5 − 5 = 0`.
- MSD : `round(3 × 0.2) = round(0.6) = 1`.
- Total : `1 + 0 + 1 = 2`.

**Pénalité d'architecture.** `round(12 × (1 − 0.85)) = round(1.8) = 2`.

**Somme.** `9 + 1 + 15 + 16 + 8 + 2 + 2 = 53`.

**HealthScore.** `max(0, 100 − 53) = 47`. Note : `47 ≥ 45` → **D**.

**Scores par catégorie (tels que rapportés).**

| Catégorie    | Pénalité | Score                                            |
| ------------ | -------- | ------------------------------------------------ |
| Complexité   | 9        | `100 − 9×5 = 55`                                 |
| Code mort    | 1        | `100 − 1×5 = 95`                                 |
| Duplication  | 15       | `100 − 15×5 = 25`                                |
| Couplage     | 16       | `100 − 16×5 = 20`                                |
| Cohésion     | 8        | `100 − 8×5 = 60`                                 |
| Dépendances  | 2        | `normalized = round(2/16 × 20) = 3` ; `100 − 3×5 = 85` |
| Architecture | —        | `round(0.85 × 100) = 85`                         |

## Score de repli

`CalculateHealthScore()` appelle d'abord `Validate()` sur le résumé. Si la validation échoue — par exemple `AverageComplexity < 0`, `CodeDuplication` hors de `[0, 100]`, `ArchCompliance` hors de `[0, 1]` quand activé, `DepsMainSequenceDeviation` hors de `[0, 1]` quand activé, ou la somme des classes haut + moyen dépassant le total pour LCOM ou CBO — les scores du résumé sont remis à zéro, la note est fixée à `"N/A"`, et une erreur est renvoyée. L'appelant peut alors invoquer `CalculateFallbackScore()` comme chemin dégradé : en partant de 100, il soustrait `FallbackComplexityThreshold = 10` si `AverageComplexity > 10`, et `FallbackPenalty = 5` pour chacun de `DeadCodeCount > 0`, `HighComplexityCount > 0` et `HighLCOMClasses > 0`, avec un plancher à `MinimumScore = 0`.

Source : `domain/analyze.go:200-262` (`Validate`), `domain/analyze.go:456-470` (branche de validation dans `CalculateHealthScore`), `domain/analyze.go:538-566` (`CalculateFallbackScore`).

## Arrondi

Tous les intermédiaires non entiers sont ramenés à des entiers via `math.Round` de Go, qui applique l'arrondi du banquier pour les valeurs exactement à `.5` (round-half-away-from-zero pour les valeurs positives dans l'implémentation Go). Le score de santé final est borné à `[0, 100]` via `MinimumScore`. Les scores par catégorie sont bornés à `[0, 100]` à l'intérieur de `penaltyToScore` (`domain/analyze.go:441-453`). La seule exception à `math.Round` est la pénalité de code mort, qui utilise la troncature `int()` après `math.Min` (`domain/analyze.go:294`).

## Suivi au fil du temps

La formule est déterministe — des entrées identiques produisent toujours des scores identiques entre exécutions et plateformes. Pour suivre la santé au fil du temps, persistez `summary.health_score` issu de la sortie JSON par commit et comparez-le dans la CI. Les scores par catégorie (`complexity_score`, `dead_code_score`, etc.) sont également stables et peuvent être suivis individuellement.

## Références

Tous les numéros de ligne se rapportent à `domain/analyze.go` :

- `CalculateHealthScore()` — lignes 456–534
- `calculateComplexityPenalty()` — lignes 266–279
- `calculateDeadCodePenalty(normalizationFactor)` — lignes 283–296
- `calculateDuplicationPenalty()` — lignes 300–314
- `calculateCouplingPenalty()` — lignes 318–336
- `calculateCohesionPenalty()` — lignes 340–358
- `calculateDependencyPenalty()` — lignes 361–406
- `calculateArchitecturePenalty()` — lignes 409–422
- `CalculateFallbackScore()` — lignes 538–566
- `Validate()` — lignes 200–262
- `penaltyToScore()` — lignes 441–453
- `normalizeToScoreBase()` — lignes 426–438
- `GetGradeFromScore()` — lignes 569–582

Calcul amont de `CodeDuplication` : `app/analyze_usecase.go:570-583`.
