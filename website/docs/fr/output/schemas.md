# Schémas de sortie

Cette spécification définit la forme exacte des sorties JSON, YAML et CSV produites par pyscn. Tous les noms de champs, types et sémantiques documentés ici sont stables d'une version corrective à l'autre au sein d'une même version majeure.

## Contrat de stabilité

| Garantie          | Portée                                                                            |
| ----------------- | --------------------------------------------------------------------------------- |
| Stable            | noms de champs, types de champs, sémantique des champs, valeurs d'énumération     |
| Peut changer      | ordre des champs au sein d'un objet, ordre des éléments d'un tableau, ajout de nouveaux champs |
| Incompatible      | suppression ou renommage de champs, modification du type d'un champ, suppression de valeurs d'énumération |

Les changements incompatibles sont limités aux changements de version majeure. Les consommateurs DOIVENT ignorer les champs inconnus.

<!-- Field naming note: in `pyscn analyze` JSON/YAML, nested analyzer objects (`complexity`, `cbo`, `lcom`, `system`) use Go-style PascalCase field names because their response structs do not carry JSON tags. Top-level keys, `dead_code`, `clone`, `suggestions`, and `summary` use snake_case. -->

## Structure de premier niveau (`pyscn analyze`)

Les sorties JSON et YAML sérialisent la structure Go `AnalyzeResponse` définie dans `domain/analyze.go`. Les clés de premier niveau sont :

```json
{
  "complexity":    { /* ComplexityResponse, present when enabled */ },
  "dead_code":     { /* DeadCodeResponse, present when enabled */ },
  "clone":         { /* CloneResponse, present when enabled */ },
  "cbo":           { /* CBOResponse, present when enabled */ },
  "lcom":          { /* LCOMResponse, present when enabled */ },
  "system":        { /* SystemAnalysisResponse, present when deps/arch enabled */ },
  "mock_data":     { /* MockDataResponse, present when enabled */ },
  "suggestions":   [ /* Suggestion array, omitted when empty */ ],
  "summary":       { /* AnalyzeSummary, always present */ },
  "generated_at":  "2026-04-14T10:18:23Z",
  "duration_ms":   2347,
  "version":       "0.14.0"
}
```

| Champ         | Type              | Description                                                       | Stabilité |
| ------------- | ----------------- | ----------------------------------------------------------------- | --------- |
| `complexity`  | object \| absent  | Présent lorsque l'analyse de complexité a été exécutée.           | stable    |
| `dead_code`   | object \| absent  | Présent lorsque l'analyse de code mort a été exécutée.            | stable    |
| `clone`       | object \| absent  | Présent lorsque la détection de clones a été exécutée.            | stable    |
| `cbo`         | object \| absent  | Présent lorsque l'analyse CBO a été exécutée.                     | stable    |
| `lcom`        | object \| absent  | Présent lorsque l'analyse LCOM a été exécutée.                    | stable    |
| `system`      | object \| absent  | Présent lorsque l'analyse des dépendances ou de l'architecture a été exécutée. | stable |
| `mock_data`   | object \| absent  | Présent lorsque la détection de données fictives a été exécutée.  | stable    |
| `suggestions` | array \| absent   | Suggestions dérivées. Omis lorsque vide.                          | stable    |
| `summary`     | object            | Toujours présent. Voir [`summary`](#summary-object).               | stable    |
| `generated_at`| string (RFC 3339) | Heure de fin d'analyse.                                           | stable    |
| `duration_ms` | integer           | Durée totale d'analyse en millisecondes.                          | stable    |
| `version`     | string            | Version sémantique de pyscn.                                      | stable    |

## Objet `summary` { #summary-object }

Reflet de `domain.AnalyzeSummary`. Tous les compteurs numériques valent `0` par défaut lorsque l'analyseur correspondant est désactivé. Tous les champs sont toujours présents.

### Statistiques de fichiers

| Champ            | Type    | Description                                       |
| ---------------- | ------- | ------------------------------------------------- |
| `total_files`    | integer | Nombre de fichiers Python découverts.             |
| `analyzed_files` | integer | Nombre de fichiers analysés avec succès.          |
| `skipped_files`  | integer | Fichiers ignorés à cause d'erreurs de parsing ou de filtres. |

### Indicateurs d'état des analyseurs

| Champ                | Type    | Description                                                 |
| -------------------- | ------- | ----------------------------------------------------------- |
| `complexity_enabled` | boolean | `true` si l'analyse de complexité a produit des résultats. |
| `dead_code_enabled`  | boolean | `true` si l'analyse de code mort a produit des résultats.  |
| `clone_enabled`      | boolean | `true` si la détection de clones a produit des résultats.  |
| `cbo_enabled`        | boolean | `true` si l'analyse CBO a produit des résultats.           |
| `lcom_enabled`       | boolean | `true` si l'analyse LCOM a produit des résultats.          |
| `deps_enabled`       | boolean | `true` si l'analyse des dépendances a produit des résultats. |
| `arch_enabled`       | boolean | `true` si la validation d'architecture a produit des résultats. |
| `mock_data_enabled`  | boolean | `true` si la détection de données fictives a produit des résultats. |

### Métriques de complexité

| Champ                   | Type    | Description                                       |
| ----------------------- | ------- | ------------------------------------------------- |
| `total_functions`       | integer | Total des fonctions analysées.                    |
| `average_complexity`    | number  | Complexité cyclomatique moyenne. `0` quand il n'y a aucune fonction. |
| `high_complexity_count` | integer | Fonctions avec complexité > 10 (seuil moyen).     |

### Métriques de code mort

| Champ                | Type    | Description                                  |
| -------------------- | ------- | -------------------------------------------- |
| `dead_code_count`    | integer | Total des constats.                          |
| `critical_dead_code` | integer | Constats de sévérité `critical`.             |
| `warning_dead_code`  | integer | Constats de sévérité `warning`.              |
| `info_dead_code`     | integer | Constats de sévérité `info`.                 |

### Métriques de clones

| Champ                         | Type    | Description                                              |
| ----------------------------- | ------- | -------------------------------------------------------- |
| `total_clones`                | integer | Fragments de code distincts identifiés comme clones.     |
| `clone_pairs`                 | integer | Nombre de paires de clones.                              |
| `clone_groups`                | integer | Nombre de groupes de clones.                             |
| `code_duplication_percentage` | number  | Taux de duplication estimé, `0`–`100`.                   |

### Métriques CBO

| Champ                     | Type    | Description                                              |
| ------------------------- | ------- | -------------------------------------------------------- |
| `cbo_classes`             | integer | Total des classes analysées.                             |
| `high_coupling_classes`   | integer | Classes avec CBO > 7.                                    |
| `medium_coupling_classes` | integer | Classes avec 3 < CBO ≤ 7.                                |
| `average_coupling`        | number  | Valeur CBO moyenne.                                      |

### Métriques LCOM

| Champ                 | Type    | Description                                  |
| --------------------- | ------- | -------------------------------------------- |
| `lcom_classes`        | integer | Total des classes analysées.                 |
| `high_lcom_classes`   | integer | Classes avec LCOM4 > 5.                      |
| `medium_lcom_classes` | integer | Classes avec 2 < LCOM4 ≤ 5.                  |
| `average_lcom`        | number  | Valeur LCOM4 moyenne.                        |

### Métriques de dépendances

| Champ                          | Type    | Description                                                    |
| ------------------------------ | ------- | -------------------------------------------------------------- |
| `deps_total_modules`           | integer | Total des modules analysés.                                    |
| `deps_modules_in_cycles`       | integer | Modules participant à au moins une dépendance circulaire.      |
| `deps_max_depth`               | integer | Longueur de la plus longue chaîne de dépendances.              |
| `deps_main_sequence_deviation` | number  | Distance moyenne à la séquence principale de Martin, `0`–`1`.  |

### Métriques d'architecture

| Champ             | Type   | Description                                                            |
| ----------------- | ------ | ---------------------------------------------------------------------- |
| `arch_compliance` | number | Taux de conformité architecturale, `0`–`1`. Pondéré par sévérité (`error × 5 + warning × 1`) ; voir `system.ArchitectureAnalysis.WeightedViolations` pour le numérateur. |

### Métriques de données fictives

| Champ                     | Type    | Description                                         |
| ------------------------- | ------- | --------------------------------------------------- |
| `mock_data_count`         | integer | Total des constats de données fictives.             |
| `mock_data_error_count`   | integer | Constats de sévérité error.                         |
| `mock_data_warning_count` | integer | Constats de sévérité warning.                       |
| `mock_data_info_count`    | integer | Constats de sévérité info.                          |

### Notation de santé

| Champ                | Type    | Description                                                          |
| -------------------- | ------- | -------------------------------------------------------------------- |
| `health_score`       | integer | Score composite, `0`–`100`. Voir [Score de santé](health-score.md).  |
| `grade`              | string  | Note littérale. L'une de : `A`, `B`, `C`, `D`, `F`, `N/A`.           |
| `complexity_score`   | integer | Score par catégorie, `0`–`100`.                                      |
| `dead_code_score`    | integer | Score par catégorie, `0`–`100`.                                      |
| `duplication_score`  | integer | Score par catégorie, `0`–`100`.                                      |
| `coupling_score`     | integer | Score par catégorie, `0`–`100`.                                      |
| `cohesion_score`     | integer | Score par catégorie, `0`–`100`.                                      |
| `dependency_score`   | integer | Score par catégorie, `0`–`100`.                                      |
| `architecture_score` | integer | Score par catégorie, `0`–`100`.                                      |

## Objet `complexity`

Reflet de `domain.ComplexityResponse`. Les noms de champs imbriqués sont en PascalCase Go.

```json
{
  "Functions": [ /* FunctionComplexity array */ ],
  "Summary": { /* ComplexitySummary */ },
  "raw_metrics": [ /* RawMetrics array, present when computed */ ],
  "raw_metrics_summary": { /* RawMetricsSummary, present when computed */ },
  "Warnings": [ "..." ],
  "Errors": [ "..." ],
  "GeneratedAt": "2026-04-14T10:18:23Z",
  "Version": "0.14.0",
  "Config": null
}
```

### Élément de `Functions[]` (`FunctionComplexity`)

| Champ         | Type    | Description                                                  |
| ------------- | ------- | ------------------------------------------------------------ |
| `Name`        | string  | Nom de la fonction. `__main__` pour le code au niveau du module. |
| `FilePath`    | string  | Chemin du fichier source.                                    |
| `StartLine`   | integer | Ligne de début, base 1.                                      |
| `StartColumn` | integer | Colonne de début, base 0.                                    |
| `EndLine`     | integer | Ligne de fin, base 1.                                        |
| `Metrics`     | object  | Voir [`ComplexityMetrics`](#complexitymetrics-object).        |
| `RiskLevel`   | string  | L'une de : `low`, `medium`, `high`.                          |

### Objet `ComplexityMetrics` { #complexitymetrics-object }

| Champ                 | Type    | Description                                        |
| --------------------- | ------- | -------------------------------------------------- |
| `Complexity`          | integer | Complexité cyclomatique de McCabe.                 |
| `CognitiveComplexity` | integer | Complexité cognitive (style SonarQube).            |
| `Nodes`               | integer | Nombre de nœuds du CFG.                            |
| `Edges`               | integer | Nombre d'arêtes du CFG.                            |
| `NestingDepth`        | integer | Profondeur d'imbrication maximale.                 |
| `IfStatements`        | integer | Nombre d'instructions `if`.                        |
| `LoopStatements`      | integer | Nombre de boucles `for`/`while`.                   |
| `ExceptionHandlers`   | integer | Nombre de clauses `except`.                        |
| `SwitchCases`         | integer | Nombre de cas `match` (Python 3.10+).              |

### Objet `Summary` (`ComplexitySummary`)

| Champ                    | Type    | Description                                                                  |
| ------------------------ | ------- | ---------------------------------------------------------------------------- |
| `TotalFunctions`         | integer | Total des fonctions analysées.                                               |
| `AverageComplexity`      | number  | Moyenne arithmétique de `Complexity` sur toutes les fonctions.               |
| `MaxComplexity`          | integer | Complexité maximale observée.                                                |
| `MinComplexity`          | integer | Complexité minimale observée.                                                |
| `FilesAnalyzed`          | integer | Fichiers contribuant au moins une fonction.                                  |
| `LowRiskFunctions`       | integer | Fonctions avec `RiskLevel = low`.                                            |
| `MediumRiskFunctions`    | integer | Fonctions avec `RiskLevel = medium`.                                         |
| `HighRiskFunctions`      | integer | Fonctions avec `RiskLevel = high`.                                           |
| `ComplexityDistribution` | object  | Histogramme indexé par tranche de complexité (string) vers compteur (integer), ou `null`. |

### Élément de `raw_metrics[]` (`RawMetrics`)

| Champ             | Type    | Description                                         |
| ----------------- | ------- | --------------------------------------------------- |
| `file_path`       | string  | Chemin du fichier source.                           |
| `sloc`            | integer | Lignes de code source (non vides, non commentaires). |
| `lloc`            | integer | Lignes de code logiques.                            |
| `comment_lines`   | integer | Lignes contenant des commentaires.                  |
| `docstring_lines` | integer | Lignes à l'intérieur de docstrings.                 |
| `blank_lines`     | integer | Lignes vides ou ne contenant que des espaces.       |
| `total_lines`     | integer | Total des lignes physiques.                         |
| `comment_ratio`   | number  | `(comment_lines + docstring_lines) / total_lines`, `0`–`1`. |


## Objet `dead_code`

Reflet de `domain.DeadCodeResponse`. Utilise des noms de champs en snake_case partout.

```json
{
  "files": [ /* FileDeadCode array */ ],
  "summary": { /* DeadCodeSummary */ },
  "warnings": null,
  "errors": null,
  "generated_at": "",
  "version": "",
  "config": null
}
```

### Élément de `files[]` (`FileDeadCode`)

| Champ               | Type    | Description                                          |
| ------------------- | ------- | ---------------------------------------------------- |
| `file_path`         | string  | Chemin du fichier source.                            |
| `functions`         | array   | Résultats par fonction (voir ci-dessous).            |
| `total_findings`    | integer | Somme des constats sur les fonctions de ce fichier.  |
| `total_functions`   | integer | Fonctions analysées dans ce fichier.                 |
| `affected_functions`| integer | Fonctions avec au moins un constat.                  |
| `dead_code_ratio`   | number  | Blocs morts / blocs totaux, `0`–`1`.                 |

### Élément de `files[].functions[]` (`FunctionDeadCode`)

| Champ             | Type    | Description                                  |
| ----------------- | ------- | -------------------------------------------- |
| `name`            | string  | Nom de la fonction.                          |
| `file_path`       | string  | Chemin du fichier source.                    |
| `findings`        | array   | Constats dans cette fonction (voir ci-dessous). |
| `total_blocks`    | integer | Total des blocs CFG dans la fonction.        |
| `dead_blocks`     | integer | Blocs CFG inatteignables.                    |
| `reachable_ratio` | number  | `(total_blocks - dead_blocks) / total_blocks`, `0`–`1`. |
| `critical_count`  | integer | Constats de sévérité `critical`.             |
| `warning_count`   | integer | Constats de sévérité `warning`.              |
| `info_count`      | integer | Constats de sévérité `info`.                 |

### Élément de `files[].functions[].findings[]` (`DeadCodeFinding`)

| Champ           | Type    | Description                                                   |
| --------------- | ------- | ------------------------------------------------------------- |
| `location`      | object  | Voir [`DeadCodeLocation`](#deadcodelocation-object).           |
| `function_name` | string  | Nom de la fonction englobante.                                |
| `code`          | string  | Extrait du code source mort.                                  |
| `reason`        | string  | Classification — voir l'énumération ci-dessous.               |
| `severity`      | string  | L'une de : `critical`, `warning`, `info`.                     |
| `description`   | string  | Description lisible par un humain.                            |
| `context`       | array of string \| absent | Lignes de code environnantes. Présent avec `--show-context`. |
| `block_id`      | string \| absent | Identifiant du bloc CFG.                               |

Énumération `reason` :

| Valeur                | Signification                                |
| --------------------- | -------------------------------------------- |
| `after_return`        | Code suivant une instruction `return`.       |
| `after_break`         | Code suivant une instruction `break`.        |
| `after_continue`      | Code suivant une instruction `continue`.     |
| `after_raise`         | Code suivant une instruction `raise`.        |
| `unreachable_branch`  | Branche conditionnelle jamais empruntée.     |

### Objet `DeadCodeLocation` { #deadcodelocation-object }

| Champ          | Type    | Description                |
| -------------- | ------- | -------------------------- |
| `file_path`    | string  | Chemin du fichier source.  |
| `start_line`   | integer | Ligne de début, base 1.    |
| `end_line`     | integer | Ligne de fin, base 1.      |
| `start_column` | integer | Colonne de début, base 0.  |
| `end_column`   | integer | Colonne de fin, base 0.    |

### Objet `summary` (`DeadCodeSummary`)

| Champ                      | Type    | Description                                      |
| -------------------------- | ------- | ------------------------------------------------ |
| `total_files`              | integer | Fichiers analysés.                               |
| `total_functions`          | integer | Fonctions analysées.                             |
| `total_findings`           | integer | Total des constats sur tous les fichiers.        |
| `files_with_dead_code`     | integer | Fichiers avec au moins un constat.               |
| `functions_with_dead_code` | integer | Fonctions avec au moins un constat.              |
| `critical_findings`        | integer | Constats de sévérité `critical`.                 |
| `warning_findings`         | integer | Constats de sévérité `warning`.                  |
| `info_findings`            | integer | Constats de sévérité `info`.                     |
| `findings_by_reason`       | object \| null | Histogramme indexé par valeur de `reason`. |
| `total_blocks`             | integer | Blocs CFG sur toutes les fonctions.              |
| `dead_blocks`              | integer | Blocs CFG inatteignables sur toutes les fonctions. |
| `overall_dead_ratio`       | number  | `dead_blocks / total_blocks`, `0`–`1`.           |

## Objet `clone`

Reflet de `domain.CloneResponse`. Utilise des noms de champs en snake_case partout.

```json
{
  "clones": [ /* Clone array, or null */ ],
  "clone_pairs": [ /* ClonePair array, or null */ ],
  "clone_groups": [ /* CloneGroup array, or null */ ],
  "statistics": { /* CloneStatistics */ },
  "duration_ms": 123,
  "success": true,
  "error": ""
}
```

### Élément de `clones[]` (`Clone`)

| Champ        | Type    | Description                                                  |
| ------------ | ------- | ------------------------------------------------------------ |
| `id`         | integer | Identifiant du clone, unique dans la réponse.                |
| `type`       | integer | Type de clone en entier : `1`, `2`, `3` ou `4`.              |
| `location`   | object  | Voir [`CloneLocation`](#clonelocation-object).                |
| `content`    | string  | Texte source brut. Présent uniquement si `--show-content` est défini. |
| `hash`       | string  | Hachage d'empreinte (algorithme dépendant du type de clone). |
| `size`       | integer | Nombre de nœuds AST.                                         |
| `line_count` | integer | Nombre de lignes du fragment.                                |
| `complexity` | integer | Complexité cyclomatique du fragment.                         |

Énumération `type` (valeurs entières) :

| Valeur | Signification                                                        |
| ------ | -------------------------------------------------------------------- |
| `1`    | Type-1 : identiques aux espaces/commentaires près.                   |
| `2`    | Type-2 : syntaxiquement identiques, identifiants/littéraux différents. |
| `3`    | Type-3 : structurellement similaires avec modifications.             |
| `4`    | Type-4 : sémantiquement équivalents, syntaxiquement différents.      |

### Objet `CloneLocation` { #clonelocation-object }

| Champ        | Type    | Description              |
| ------------ | ------- | ------------------------ |
| `file_path`  | string  | Chemin du fichier source. |
| `start_line` | integer | Ligne de début, base 1.   |
| `end_line`   | integer | Ligne de fin, base 1.     |
| `start_col`  | integer | Colonne de début, base 0. |
| `end_col`    | integer | Colonne de fin, base 0.   |

### Élément de `clone_pairs[]` (`ClonePair`)

| Champ        | Type    | Description                                            |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | Identifiant de paire.                                  |
| `clone1`     | object  | Premier clone (objet `Clone`).                         |
| `clone2`     | object  | Second clone (objet `Clone`).                          |
| `similarity` | number  | Score de similarité, `0`–`1`.                          |
| `distance`   | number  | Distance d'édition d'arbre (Type-3) ou `0` sinon.      |
| `type`       | integer | Type de clone (même énumération que `clones[].type`).  |
| `confidence` | number  | Confiance du détecteur, `0`–`1`.                       |

### Élément de `clone_groups[]` (`CloneGroup`)

| Champ        | Type    | Description                                            |
| ------------ | ------- | ------------------------------------------------------ |
| `id`         | integer | Identifiant du groupe.                                 |
| `clones`     | array   | Objets `Clone` membres.                                |
| `type`       | integer | Type de clone dominant.                                |
| `similarity` | number  | Similarité représentative, `0`–`1`.                    |
| `size`       | integer | Nombre de membres (`len(clones)`).                     |

### Objet `statistics` (`CloneStatistics`)

| Champ                | Type    | Description                                                |
| -------------------- | ------- | ---------------------------------------------------------- |
| `total_fragments`    | integer | Tous les fragments extraits (fonctions, classes, etc.).    |
| `total_clones`       | integer | Fragments classés comme clones.                            |
| `total_clone_pairs`  | integer | Nombre de paires détectées.                                |
| `total_clone_groups` | integer | Nombre de groupes.                                         |
| `clones_by_type`     | object \| null | Map de l'étiquette de type (`Type-1`…`Type-4`) vers le compteur. |
| `average_similarity` | number  | Similarité moyenne sur les paires, `0`–`1`.                |
| `lines_analyzed`     | integer | Total des lignes source considérées.                       |
| `nodes_analyzed`     | integer | Total des nœuds AST considérés.                            |
| `files_analyzed`     | integer | Fichiers distincts contribuant à des fragments.            |

Autres champs de `CloneResponse` :

| Champ         | Type    | Description                                        |
| ------------- | ------- | -------------------------------------------------- |
| `duration_ms` | integer | Durée de détection de clones en millisecondes.     |
| `success`     | boolean | `true` en cas d'achèvement normal.                 |
| `error`       | string \| absent | Message d'erreur si `success=false`.        |

## Objet `cbo`

Reflet de `domain.CBOResponse`. Les noms de champs imbriqués sont en PascalCase Go.

```json
{
  "Classes": [ /* ClassCoupling array */ ],
  "Summary": { /* CBOSummary */ },
  "Warnings": null,
  "Errors": null,
  "GeneratedAt": "",
  "Version": "",
  "Config": null
}
```

### Élément de `Classes[]` (`ClassCoupling`)

| Champ         | Type    | Description                                 |
| ------------- | ------- | ------------------------------------------- |
| `Name`        | string  | Nom de la classe.                           |
| `FilePath`    | string  | Chemin du fichier source.                   |
| `StartLine`   | integer | Ligne de début, base 1.                     |
| `EndLine`     | integer | Ligne de fin, base 1.                       |
| `Metrics`     | object  | Voir [`CBOMetrics`](#cbometrics-object).     |
| `RiskLevel`   | string  | L'une de : `low`, `medium`, `high`.         |
| `IsAbstract`  | boolean | `true` si la classe est abstraite.          |
| `BaseClasses` | array of string \| null | Classes de base directes.   |

### Objet `CBOMetrics` { #cbometrics-object }

| Champ                         | Type    | Description                                                |
| ----------------------------- | ------- | ---------------------------------------------------------- |
| `CouplingCount`               | integer | Valeur CBO : classes distinctes dont dépend cette classe.  |
| `InheritanceDependencies`     | integer | Dépendances par classes de base.                           |
| `TypeHintDependencies`        | integer | Dépendances par annotations de type.                       |
| `InstantiationDependencies`   | integer | Dépendances par instanciation d'objets.                    |
| `AttributeAccessDependencies` | integer | Dépendances par appels de méthodes et accès aux attributs. |
| `ImportDependencies`          | integer | Dépendances par imports explicites.                        |
| `DependentClasses`            | array of string \| null | Noms des classes couplées.                 |

### Objet `Summary` (`CBOSummary`)

| Champ                      | Type    | Description                                       |
| -------------------------- | ------- | ------------------------------------------------- |
| `TotalClasses`             | integer | Total des classes analysées.                      |
| `AverageCBO`               | number  | CBO moyen.                                        |
| `MaxCBO`                   | integer | CBO maximal observé.                              |
| `MinCBO`                   | integer | CBO minimal observé.                              |
| `ClassesAnalyzed`          | integer | Classes avec des métriques valides.               |
| `FilesAnalyzed`            | integer | Fichiers contribuant au moins une classe.         |
| `LowRiskClasses`           | integer | Classes avec CBO ≤ seuil bas (par défaut `3`).    |
| `MediumRiskClasses`        | integer | Classes avec seuil bas < CBO ≤ seuil moyen.       |
| `HighRiskClasses`          | integer | Classes avec CBO > seuil moyen (par défaut `7`).  |
| `CBODistribution`          | object \| null | Histogramme indexé par étiquette de tranche vers compteur. |
| `MostCoupledClasses`       | array \| null | Top 10 des classes par CBO (`ClassCoupling`). |
| `MostDependedUponClasses`  | array of string \| null | Classes avec le plus fort degré entrant. |

## Objet `lcom`

Reflet de `domain.LCOMResponse`. Les noms de champs imbriqués sont en PascalCase Go.

```json
{
  "Classes": [ /* ClassCohesion array */ ],
  "Summary": { /* LCOMSummary */ },
  "Warnings": null,
  "Errors": null,
  "GeneratedAt": "",
  "Version": "",
  "Config": null
}
```

### Élément de `Classes[]` (`ClassCohesion`)

| Champ       | Type    | Description                                      |
| ----------- | ------- | ------------------------------------------------ |
| `Name`      | string  | Nom de la classe.                                |
| `FilePath`  | string  | Chemin du fichier source.                        |
| `StartLine` | integer | Ligne de début, base 1.                          |
| `EndLine`   | integer | Ligne de fin, base 1.                            |
| `Metrics`   | object  | Voir [`LCOMMetrics`](#lcommetrics-object).        |
| `RiskLevel` | string  | L'une de : `low`, `medium`, `high`.              |

### Objet `LCOMMetrics` { #lcommetrics-object }

| Champ               | Type    | Description                                                  |
| ------------------- | ------- | ------------------------------------------------------------ |
| `LCOM4`             | integer | Composantes connexes dans le graphe méthodes-variables.      |
| `TotalMethods`      | integer | Toutes les méthodes de la classe.                            |
| `ExcludedMethods`   | integer | Méthodes exclues de LCOM4 (`@classmethod`, `@staticmethod`). |
| `InstanceVariables` | integer | Variables `self.x` distinctes accédées.                      |
| `MethodGroups`      | array of array of string \| null | Noms de méthodes groupés par composante connexe. |

### Objet `Summary` (`LCOMSummary`)

| Champ                  | Type    | Description                                          |
| ---------------------- | ------- | ---------------------------------------------------- |
| `TotalClasses`         | integer | Classes analysées.                                   |
| `AverageLCOM`          | number  | LCOM4 moyen.                                         |
| `MaxLCOM`              | integer | LCOM4 maximal observé.                               |
| `MinLCOM`              | integer | LCOM4 minimal observé.                               |
| `ClassesAnalyzed`      | integer | Classes avec des métriques valides.                  |
| `FilesAnalyzed`        | integer | Fichiers contribuant au moins une classe.            |
| `LowRiskClasses`       | integer | Classes avec LCOM4 ≤ seuil bas (par défaut `2`).     |
| `MediumRiskClasses`    | integer | Classes avec seuil bas < LCOM4 ≤ seuil moyen.        |
| `HighRiskClasses`      | integer | Classes avec LCOM4 > seuil moyen (par défaut `5`).   |
| `LCOMDistribution`     | object \| null | Histogramme indexé par étiquette de tranche vers compteur. |
| `LeastCohesiveClasses` | array \| null | Top 10 des classes par LCOM4 (`ClassCohesion`).|

## Objet `system`

Reflet de `domain.SystemAnalysisResponse`. Les noms de champs imbriqués sont en PascalCase Go.

```json
{
  "DependencyAnalysis":   { /* DependencyAnalysisResult, or null */ },
  "ArchitectureAnalysis": { /* ArchitectureAnalysisResult, or null */ },
  "Summary":              { /* SystemAnalysisSummary */ },
  "Issues":               [ /* SystemIssue array */ ],
  "Recommendations":      [ /* SystemRecommendation array */ ],
  "Warnings":             [ ],
  "Errors":               [ ],
  "GeneratedAt":          "0001-01-01T00:00:00Z",
  "Duration":             0,
  "Version":              "",
  "Config":               null
}
```

### Objet `Summary` (`SystemAnalysisSummary`)

| Champ                      | Type    | Description                                       |
| -------------------------- | ------- | ------------------------------------------------- |
| `TotalModules`             | integer | Total des modules analysés.                       |
| `TotalPackages`            | integer | Total des paquets.                                |
| `TotalDependencies`        | integer | Total des arêtes de dépendances.                  |
| `ProjectRoot`              | string  | Répertoire racine du projet.                      |
| `OverallQualityScore`      | number  | Score de qualité composite, `0`–`100`.            |
| `MaintainabilityScore`     | number  | Indice de maintenabilité moyen.                   |
| `ArchitectureScore`        | number  | Score de conformité architecturale.               |
| `ModularityScore`          | number  | Score de modularité du système.                   |
| `TechnicalDebtHours`       | number  | Dette technique totale estimée en heures.         |
| `AverageCoupling`          | number  | Couplage moyen entre modules.                     |
| `AverageInstability`       | number  | Instabilité moyenne (I).                          |
| `CyclicDependencies`       | integer | Modules participant à des cycles.                 |
| `ArchitectureViolations`   | integer | Nombre de violations des règles architecturales.  |
| `HighRiskModules`          | integer | Modules signalés à risque élevé.                  |
| `CriticalIssues`           | integer | Nombre de problèmes critiques.                    |
| `RefactoringCandidates`    | integer | Modules nécessitant un refactoring.               |
| `ArchitectureImprovements` | integer | Améliorations architecturales suggérées.          |

### Objet `DependencyAnalysis`

| Champ                  | Type    | Description                                                          |
| ---------------------- | ------- | -------------------------------------------------------------------- |
| `TotalModules`         | integer | Total des modules dans le graphe de dépendances.                     |
| `TotalDependencies`    | integer | Total des arêtes.                                                    |
| `RootModules`          | array of string | Modules sans dépendance sortante.                            |
| `LeafModules`          | array of string | Modules sans dépendance entrante.                            |
| `ModuleMetrics`        | object  | Map du nom de module vers `ModuleDependencyMetrics`.                 |
| `DependencyMatrix`     | object  | Map de module vers map de module vers booléen.                       |
| `CircularDependencies` | object  | Résultats de détection de cycles ; contient `Cycles` (array) et `TotalCycles` (integer). |
| `CouplingAnalysis`     | object  | Métriques de couplage par module : `Ca`, `Ce`, `Instability`, `Abstractness`, `Distance`. |
| `LongestChains`        | array   | Tableau d'objets `DependencyPath`.                                   |
| `MaxDepth`             | integer | Profondeur de dépendance maximale.                                   |

### Objet `ModuleDependencyMetrics`

| Champ                    | Type    | Description                                              |
| ------------------------ | ------- | -------------------------------------------------------- |
| `ModuleName`             | string  | Nom complet du module.                                   |
| `Package`                | string  | Paquet parent.                                           |
| `FilePath`               | string  | Chemin du fichier source.                                |
| `IsPackage`              | boolean | `true` s'il s'agit d'un paquet (possède `__init__.py`).  |
| `LinesOfCode`            | integer | Total des lignes de code.                                |
| `FunctionCount`          | integer | Nombre de fonctions.                                     |
| `ClassCount`             | integer | Nombre de classes.                                       |
| `AbstractClassCount`     | integer | Nombre de classes abstraites.                            |
| `PublicInterface`        | array of string | Noms dans `__all__` ou noms publics de haut niveau. |
| `AfferentCoupling`       | integer | Ca — modules dépendant de celui-ci.                      |
| `EfferentCoupling`       | integer | Ce — modules dont celui-ci dépend.                       |
| `Instability`            | number  | `I = Ce / (Ca + Ce)`, `0`–`1`.                           |
| `Abstractness`           | number  | A — classes abstraites / classes totales, `0`–`1`.       |
| `Distance`               | number  | `D = |A + I - 1|`, `0`–`1`. Distance à la séquence principale. |
| `Maintainability`        | number  | Indice de maintenabilité, `0`–`100`.                     |
| `TechnicalDebt`          | number  | Dette technique estimée en heures.                       |
| `RiskLevel`              | string  | L'une de : `low`, `medium`, `high`.                      |
| `DirectDependencies`     | array of string | Dépendances directes.                            |
| `TransitiveDependencies` | array of string | Toutes les dépendances transitives.              |
| `Dependents`             | array of string | Modules dépendant de celui-ci.                   |

### Objet `CircularDependencyAnalysis`

| Champ                      | Type    | Description                                           |
| -------------------------- | ------- | ----------------------------------------------------- |
| `HasCircularDependencies`  | boolean | `true` s'il existe au moins un cycle.                 |
| `TotalCycles`              | integer | Nombre de cycles.                                     |
| `TotalModulesInCycles`     | integer | Modules impliqués dans des cycles.                    |
| `CircularDependencies`     | array   | Tableau d'objets `CircularDependency`.                |
| `CycleBreakingSuggestions` | array of string | Suggestions pour briser les cycles.           |
| `CoreInfrastructure`       | array of string | Modules apparaissant dans plusieurs cycles.   |

Énumération `CircularDependency.Severity` : `low`, `medium`, `high`, `critical`.

### Objet `CouplingAnalysis`

| Champ                   | Type    | Description                                            |
| ----------------------- | ------- | ------------------------------------------------------ |
| `AverageCoupling`       | number  | Couplage moyen entre les modules.                      |
| `CouplingDistribution`  | object  | Map de la valeur de couplage (clé entière) vers compteur. |
| `HighlyCoupledModules`  | array of string | Modules à fort couplage.                       |
| `LooselyCoupledModules` | array of string | Modules à faible couplage.                     |
| `AverageInstability`    | number  | Instabilité moyenne.                                   |
| `StableModules`         | array of string | Modules à faible instabilité.                  |
| `InstableModules`       | array of string | Modules à forte instabilité.                   |
| `MainSequenceDeviation` | number  | Distance moyenne à la séquence principale, `0`–`1`.    |
| `ZoneOfPain`            | array of string | Modules stables + concrets.                    |
| `ZoneOfUselessness`     | array of string | Modules instables + abstraits.                 |
| `MainSequence`          | array of string | Modules bien positionnés.                      |

### Objet `ArchitectureAnalysis`

| Champ                 | Type    | Description                                                 |
| --------------------- | ------- | ----------------------------------------------------------- |
| `ComplianceScore`     | number  | Score de conformité, `0`–`1`. Calculé comme `max(0, 1 - WeightedViolations / TotalRules)` ; `1.0` si aucune règle évaluée. |
| `TotalViolations`     | integer | Nombre brut de violations (une par entrée de `Violations`). |
| `WeightedViolations`  | integer | Nombre de violations pondéré par sévérité, utilisé comme numérateur de `ComplianceScore` : `error × 5 + warning × 1`. |
| `TotalRules`          | integer | Total des invocations de règles évaluées (dénominateur de `ComplianceScore`). |
| `LayerAnalysis`       | object \| null | Résultats d'analyse par couches.                     |
| `CohesionAnalysis`    | object \| null | Analyse de cohésion des paquets.                     |
| `ResponsibilityAnalysis` | object \| null | Analyse des violations du principe SRP.           |
| `Violations`          | array   | Tableau d'objets `ArchitectureViolation`.                   |
| `SeverityBreakdown`   | object  | Map de la sévérité vers le compteur.                        |
| `Recommendations`     | array   | Tableau d'objets `ArchitectureRecommendation`.              |
| `RefactoringTargets`  | array of string | Modules nécessitant un refactoring.                 |

Énumération `ArchitectureViolation.Type` : `layer`, `cycle`, `coupling`, `responsibility`, `cohesion`.

Énumération `ArchitectureViolation.Severity` : `info`, `warning`, `error`, `critical`.

## Tableau `suggestions`

Tableau d'objets `Suggestion`. Utilise des noms de champs en snake_case.

| Champ         | Type    | Requis | Description                                       |
| ------------- | ------- | ------ | ------------------------------------------------- |
| `category`    | string  | oui    | Voir l'énumération ci-dessous.                    |
| `severity`    | string  | oui    | L'une de : `critical`, `warning`, `info`.         |
| `effort`      | string  | oui    | L'une de : `easy`, `moderate`, `hard`.            |
| `title`       | string  | oui    | Titre court lisible par un humain.                |
| `description` | string  | oui    | Description complète.                             |
| `steps`       | array of string | non | Étapes concrètes. Omis lorsque vide.         |
| `file_path`   | string  | non    | Référence vers le fichier source.                 |
| `function`    | string  | non    | Référence vers un nom de fonction.                |
| `class_name`  | string  | non    | Référence vers un nom de classe.                  |
| `start_line`  | integer | non    | Référence de ligne, base 1. Omis lorsque `0`.     |
| `metric_value`| string  | non    | Valeur observée de la métrique sous forme de chaîne. |
| `threshold`   | string  | non    | Valeur du seuil sous forme de chaîne.             |

Énumération `category` : `complexity`, `dead_code`, `clone`, `coupling`, `cohesion`, `dependency`, `architecture`.

Les suggestions sont triées par priorité (sévérité × effort). Voir `domain/suggestion.go` pour la fonction de tri exacte.

## Schémas CSV

Les sorties CSV sont écrites avec les règles de mise en guillemets RFC 4180 via le paquet Go `encoding/csv`.

### `pyscn analyze --csv`

Résumé uniquement. Deux colonnes. Chaînes UTF-8 littérales, sans annotation de type.

| Colonne  | Type   | Description                            |
| -------- | ------ | -------------------------------------- |
| `Metric` | string | Nom de la métrique.                    |
| `Value`  | string | Valeur de la métrique sous forme de chaîne. |

Lignes (dans cet ordre fixe) :

```csv
Metric,Value
Health Score,<integer>
Grade,<A|B|C|D|F|N/A>
Total Files,<integer>
Analyzed Files,<integer>
Average Complexity,<float with 2 decimals>
High Complexity Count,<integer>
Dead Code Count,<integer>
Critical Dead Code,<integer>
Unique Fragments,<integer>
Clone Groups,<integer>
Code Duplication,<float with 2 decimals>
Total Classes Analyzed,<integer>
High Coupling (CBO) Classes,<integer>
Average CBO,<float with 2 decimals>
```

pyscn n'expose actuellement pas de schémas CSV par analyseur via le CLI — `--csv` ne produit que le résumé ci-dessus. Pour les détails par constat, utilisez `--json` ou `--yaml`.

## Horodatages et versionnage

| Champ          | Format                    | Notes                                                   |
| -------------- | ------------------------- | ------------------------------------------------------- |
| `generated_at` | RFC 3339 (ISO 8601)       | Sérialisation `time.Time` ; peut inclure une précision sous-seconde et un décalage de fuseau horaire. |
| `duration_ms`  | integer (millisecondes)   | Durée d'analyse en horloge murale.                      |
| `version`      | string (version sémantique) | Version de pyscn, par exemple `"0.14.0"`.            |

## Invoquer chaque format

`pyscn analyze` accepte l'un de `--json`, `--yaml`, `--csv`, `--html` (défaut). Il n'existe pas de drapeau `--format`, et il n'y a pas de sous-commandes autonomes `complexity` / `deadcode` / `clone` / `deps`. Exécutez un seul analyseur via `--select`.

```bash
pyscn analyze --json src/
pyscn analyze --yaml src/
pyscn analyze --csv  src/
pyscn analyze --html src/    # default
pyscn analyze --json --select complexity src/
pyscn analyze --csv  --select deadcode   src/
pyscn analyze --yaml --select clones     src/
```

Les fichiers de sortie sont placés dans `.pyscn/reports/` ; voir [Formats de sortie](index.md) pour les détails de chemin et de nom de fichier.

## Voir aussi

- [Rapport HTML](html-report.md) — spécification de la sortie HTML.
- [Score de santé](health-score.md) — dérivation de `summary.health_score` et des scores par catégorie.
