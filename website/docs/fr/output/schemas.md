# Schémas de sortie

Cette spécification définit la forme exacte des sorties JSON, YAML et CSV produites par pyscn. Tous les noms de champs, types et sémantiques documentés ici sont stables d'une version corrective à l'autre au sein d'une même version majeure.

## Contrat de stabilité

| Garantie          | Portée                                                                            |
| ----------------- | --------------------------------------------------------------------------------- |
| Stable            | noms de champs, types de champs, sémantique des champs, valeurs d'énumération     |
| Peut changer      | ordre des champs au sein d'un objet, ordre des éléments d'un tableau, ajout de nouveaux champs |
| Incompatible      | suppression ou renommage de champs, modification du type d'un champ, suppression de valeurs d'énumération |

Les changements incompatibles sont limités aux changements de version majeure. Les consommateurs DOIVENT ignorer les champs inconnus.

<!-- Field naming note: every object key in `pyscn analyze` JSON/YAML is snake_case. Releases up to 1.29.1 emitted Go-style PascalCase inside `complexity`, `cbo`, `lcom`, and `system`, and lowerCamelCase inside the `config` objects of `cbo`, `lcom`, and `community_analysis`; both were renamed to snake_case. -->

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
| `skipped_files`  | integer | Fichiers écartés faute de pouvoir être lus ou analysés. Une valeur non nulle signifie que les scores ci-dessous couvrent moins que `total_files`. |

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

Reflet de `domain.ComplexityResponse`.

```json
{
  "functions": [ /* FunctionComplexity array */ ],
  "summary": { /* ComplexitySummary */ },
  "raw_metrics": [ /* RawMetrics array, present when computed */ ],
  "raw_metrics_summary": { /* RawMetricsSummary, present when computed */ },
  "warnings": [ "..." ],
  "errors": [ "..." ],
  "generated_at": "2026-04-14T10:18:23Z",
  "version": "0.14.0",
  "config": null
}
```

### Élément de `functions[]` (`FunctionComplexity`)

| Champ          | Type    | Description                                                      |
| -------------- | ------- | ---------------------------------------------------------------- |
| `name`         | string  | Nom de la fonction. `__main__` pour le code au niveau du module. |
| `file_path`    | string  | Chemin du fichier source.                                        |
| `start_line`   | integer | Ligne de début, base 1.                                          |
| `start_column` | integer | Colonne de début, base 0.                                        |
| `end_line`     | integer | Ligne de fin, base 1.                                            |
| `metrics`      | object  | Voir [`ComplexityMetrics`](#complexitymetrics-object).           |
| `risk_level`   | string  | L'une de : `low`, `medium`, `high`.                              |

### Objet `ComplexityMetrics` { #complexitymetrics-object }

| Champ                  | Type    | Description                             |
| ---------------------- | ------- | --------------------------------------- |
| `complexity`           | integer | Complexité cyclomatique de McCabe.      |
| `cognitive_complexity` | integer | Complexité cognitive (style SonarQube). |
| `nodes`                | integer | Nombre de nœuds du CFG.                 |
| `edges`                | integer | Nombre d'arêtes du CFG.                 |
| `nesting_depth`        | integer | Profondeur d'imbrication maximale.      |
| `if_statements`        | integer | Nombre d'instructions `if`.             |
| `loop_statements`      | integer | Nombre de boucles `for`/`while`.        |
| `exception_handlers`   | integer | Nombre de clauses `except`.             |
| `switch_cases`         | integer | Nombre de cas `match` (Python 3.10+).   |

### Objet `summary` (`ComplexitySummary`)

| Champ                     | Type    | Description                                                                                                        |
| ------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------ |
| `total_functions`         | integer | Total des fonctions analysées.                                                                                     |
| `average_complexity`      | number  | Moyenne arithmétique de `Complexity` sur toutes les fonctions.                                                     |
| `max_complexity`          | integer | Complexité maximale observée.                                                                                      |
| `min_complexity`          | integer | Complexité minimale observée.                                                                                      |
| `files_analyzed`          | integer | Fichiers analysés avec succès et pris en compte dans les métriques ci-dessus.                                      |
| `total_files`             | integer | Fichiers couverts par la requête, analysables ou non.                                                              |
| `skipped_files`           | integer | Fichiers écartés faute de pouvoir être lus ou analysés. Leur contenu est absent de toutes les métriques ci-dessus. |
| `low_risk_functions`      | integer | Fonctions avec `RiskLevel = low`.                                                                                  |
| `medium_risk_functions`   | integer | Fonctions avec `RiskLevel = medium`.                                                                               |
| `high_risk_functions`     | integer | Fonctions avec `RiskLevel = high`.                                                                                 |
| `complexity_distribution` | object  | Histogramme indexé par tranche de complexité (string) vers compteur (integer), ou `null`.                          |

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
| `errors`      | array \| absent  | Échecs par fichier. Un fichier listé ici a été ignoré alors que l'exécution a réussi : son contenu est absent de `statistics`. |

## Objet `cbo`

Reflet de `domain.CBOResponse`.

```json
{
  "classes": [ /* ClassCoupling array */ ],
  "summary": { /* CBOSummary */ },
  "warnings": null,
  "errors": null,
  "generated_at": "",
  "version": "",
  "config": null
}
```

### Élément de `classes[]` (`ClassCoupling`)

| Champ          | Type                    | Description                              |
| -------------- | ----------------------- | ---------------------------------------- |
| `name`         | string                  | Nom de la classe.                        |
| `file_path`    | string                  | Chemin du fichier source.                |
| `start_line`   | integer                 | Ligne de début, base 1.                  |
| `end_line`     | integer                 | Ligne de fin, base 1.                    |
| `metrics`      | object                  | Voir [`CBOMetrics`](#cbometrics-object). |
| `risk_level`   | string                  | L'une de : `low`, `medium`, `high`.      |
| `is_abstract`  | boolean                 | `true` si la classe est abstraite.       |
| `base_classes` | array of string \| null | Classes de base directes.                |

### Objet `CBOMetrics` { #cbometrics-object }

| Champ                           | Type                    | Description                                                |
| ------------------------------- | ----------------------- | ---------------------------------------------------------- |
| `coupling_count`                | integer                 | Valeur CBO : classes distinctes dont dépend cette classe.  |
| `inheritance_dependencies`      | integer                 | Dépendances par classes de base.                           |
| `type_hint_dependencies`        | integer                 | Dépendances par annotations de type.                       |
| `instantiation_dependencies`    | integer                 | Dépendances par instanciation d'objets.                    |
| `attribute_access_dependencies` | integer                 | Dépendances par appels de méthodes et accès aux attributs. |
| `import_dependencies`           | integer                 | Dépendances par imports explicites.                        |
| `dependent_classes`             | array of string \| null | Noms des classes couplées.                                 |

### Objet `summary` (`CBOSummary`)

| Champ                        | Type                    | Description                                                |
| ---------------------------- | ----------------------- | ---------------------------------------------------------- |
| `total_classes`              | integer                 | Total des classes analysées.                               |
| `average_cbo`                | number                  | CBO moyen.                                                 |
| `max_cbo`                    | integer                 | CBO maximal observé.                                       |
| `min_cbo`                    | integer                 | CBO minimal observé.                                       |
| `classes_analyzed`           | integer                 | Classes avec des métriques valides.                        |
| `files_analyzed`             | integer                 | Fichiers contribuant au moins une classe.                  |
| `low_risk_classes`           | integer                 | Classes avec CBO ≤ seuil bas (par défaut `3`).             |
| `medium_risk_classes`        | integer                 | Classes avec seuil bas < CBO ≤ seuil moyen.                |
| `high_risk_classes`          | integer                 | Classes avec CBO > seuil moyen (par défaut `7`).           |
| `cbo_distribution`           | object \| null          | Histogramme indexé par étiquette de tranche vers compteur. |
| `most_coupled_classes`       | array \| null           | Top 10 des classes par CBO (`ClassCoupling`).              |
| `most_depended_upon_classes` | array of string \| null | Classes avec le plus fort degré entrant.                   |

## Objet `lcom`

Reflet de `domain.LCOMResponse`.

```json
{
  "classes": [ /* ClassCohesion array */ ],
  "summary": { /* LCOMSummary */ },
  "warnings": null,
  "errors": null,
  "generated_at": "",
  "version": "",
  "config": null
}
```

### Élément de `classes[]` (`ClassCohesion`)

| Champ        | Type    | Description                                |
| ------------ | ------- | ------------------------------------------ |
| `name`       | string  | Nom de la classe.                          |
| `file_path`  | string  | Chemin du fichier source.                  |
| `start_line` | integer | Ligne de début, base 1.                    |
| `end_line`   | integer | Ligne de fin, base 1.                      |
| `metrics`    | object  | Voir [`LCOMMetrics`](#lcommetrics-object). |
| `risk_level` | string  | L'une de : `low`, `medium`, `high`.        |

### Objet `LCOMMetrics` { #lcommetrics-object }

| Champ                | Type                             | Description                                                                                                                                        |
| -------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `lcom4`              | integer                          | Composantes connexes dans le graphe méthodes-variables.                                                                                            |
| `total_methods`      | integer                          | Toutes les méthodes de la classe.                                                                                                                  |
| `excluded_methods`   | integer                          | Méthodes exclues du graphe LCOM4 (`@classmethod`, `@staticmethod`, `@abstractmethod` et les constructeurs `__init__`, `__new__`, `__post_init__`). |
| `instance_variables` | integer                          | Variables `self.x` distinctes accédées.                                                                                                            |
| `method_groups`      | array of array of string \| null | Noms de méthodes groupés par composante connexe.                                                                                                   |

### Objet `summary` (`LCOMSummary`)

| Champ                    | Type           | Description                                                |
| ------------------------ | -------------- | ---------------------------------------------------------- |
| `total_classes`          | integer        | Classes analysées.                                         |
| `average_lcom`           | number         | LCOM4 moyen.                                               |
| `max_lcom`               | integer        | LCOM4 maximal observé.                                     |
| `min_lcom`               | integer        | LCOM4 minimal observé.                                     |
| `classes_analyzed`       | integer        | Classes avec des métriques valides.                        |
| `files_analyzed`         | integer        | Fichiers contribuant au moins une classe.                  |
| `low_risk_classes`       | integer        | Classes avec LCOM4 ≤ seuil bas (par défaut `2`).           |
| `medium_risk_classes`    | integer        | Classes avec seuil bas < LCOM4 ≤ seuil moyen.              |
| `high_risk_classes`      | integer        | Classes avec LCOM4 > seuil moyen (par défaut `5`).         |
| `lcom_distribution`      | object \| null | Histogramme indexé par étiquette de tranche vers compteur. |
| `least_cohesive_classes` | array \| null  | Top 10 des classes par LCOM4 (`ClassCohesion`).            |

## Objet `system`

Reflet de `domain.SystemAnalysisResponse`.

```json
{
  "dependency_analysis":   { /* DependencyAnalysisResult, or null */ },
  "architecture_analysis": { /* ArchitectureAnalysisResult, or null */ },
  "summary":              { /* SystemAnalysisSummary */ },
  "issues":               [ /* SystemIssue array */ ],
  "recommendations":      [ /* SystemRecommendation array */ ],
  "warnings":             [ ],
  "errors":               [ ],
  "generated_at":          "0001-01-01T00:00:00Z",
  "duration":             0,
  "version":              "",
  "config":               null
}
```

### Objet `summary` (`SystemAnalysisSummary`)

| Champ                       | Type    | Description                                      |
| --------------------------- | ------- | ------------------------------------------------ |
| `total_modules`             | integer | Total des modules analysés.                      |
| `total_packages`            | integer | Total des paquets.                               |
| `total_dependencies`        | integer | Total des arêtes de dépendances.                 |
| `project_root`              | string  | Répertoire racine du projet.                     |
| `overall_quality_score`     | number  | Score de qualité composite, `0`–`100`.           |
| `maintainability_score`     | number  | Indice de maintenabilité moyen.                  |
| `architecture_score`        | number  | Score de conformité architecturale.              |
| `modularity_score`          | number  | Score de modularité du système.                  |
| `technical_debt_hours`      | number  | Dette technique totale estimée en heures.        |
| `average_coupling`          | number  | Couplage moyen entre modules.                    |
| `average_instability`       | number  | Instabilité moyenne (I).                         |
| `cyclic_dependencies`       | integer | Modules participant à des cycles.                |
| `architecture_violations`   | integer | Nombre de violations des règles architecturales. |
| `high_risk_modules`         | integer | Modules signalés à risque élevé.                 |
| `critical_issues`           | integer | Nombre de problèmes critiques.                   |
| `refactoring_candidates`    | integer | Modules nécessitant un refactoring.              |
| `architecture_improvements` | integer | Améliorations architecturales suggérées.         |

### Objet `dependency_analysis`

| Champ                   | Type            | Description                                                                               |
| ----------------------- | --------------- | ----------------------------------------------------------------------------------------- |
| `total_modules`         | integer         | Total des modules dans le graphe de dépendances.                                          |
| `total_dependencies`    | integer         | Total des arêtes.                                                                         |
| `root_modules`          | array of string | Modules sans dépendance sortante.                                                         |
| `leaf_modules`          | array of string | Modules sans dépendance entrante.                                                         |
| `module_metrics`        | object          | Map du nom de module vers `ModuleDependencyMetrics`.                                      |
| `dependency_matrix`     | object          | Map de module vers map de module vers booléen.                                            |
| `circular_dependencies` | object          | Résultats de détection de cycles ; contient `Cycles` (array) et `TotalCycles` (integer).  |
| `coupling_analysis`     | object          | Métriques de couplage par module : `Ca`, `Ce`, `Instability`, `Abstractness`, `Distance`. |
| `longest_chains`        | array           | Tableau d'objets `DependencyPath`.                                                        |
| `max_depth`             | integer         | Profondeur de dépendance maximale.                                                        |

### Objet `ModuleDependencyMetrics`

| Champ                     | Type            | Description                                             |           |                                                |
| ------------------------- | --------------- | ------------------------------------------------------- | --------- | ---------------------------------------------- |
| `module_name`             | string          | Nom complet du module.                                  |           |                                                |
| `package`                 | string          | Paquet parent.                                          |           |                                                |
| `file_path`               | string          | Chemin du fichier source.                               |           |                                                |
| `is_package`              | boolean         | `true` s'il s'agit d'un paquet (possède `__init__.py`). |           |                                                |
| `lines_of_code`           | integer         | Total des lignes de code.                               |           |                                                |
| `function_count`          | integer         | Nombre de fonctions.                                    |           |                                                |
| `class_count`             | integer         | Nombre de classes.                                      |           |                                                |
| `abstract_class_count`    | integer         | Nombre de classes abstraites.                           |           |                                                |
| `public_interface`        | array of string | Noms dans `__all__` ou noms publics de haut niveau.     |           |                                                |
| `afferent_coupling`       | integer         | Ca — modules dépendant de celui-ci.                     |           |                                                |
| `efferent_coupling`       | integer         | Ce — modules dont celui-ci dépend.                      |           |                                                |
| `instability`             | number          | `I = Ce / (Ca + Ce)`, `0`–`1`.                          |           |                                                |
| `abstractness`            | number          | A — classes abstraites / classes totales, `0`–`1`.      |           |                                                |
| `distance`                | number          | `D =                                                    | A + I - 1 | `, `0`–`1`. Distance à la séquence principale. |
| `maintainability`         | number          | Indice de maintenabilité, `0`–`100`.                    |           |                                                |
| `technical_debt`          | number          | Dette technique estimée en heures.                      |           |                                                |
| `risk_level`              | string          | L'une de : `low`, `medium`, `high`.                     |           |                                                |
| `direct_dependencies`     | array of string | Dépendances directes.                                   |           |                                                |
| `transitive_dependencies` | array of string | Toutes les dépendances transitives.                     |           |                                                |
| `dependents`              | array of string | Modules dépendant de celui-ci.                          |           |                                                |

### Objet `CircularDependencyAnalysis`

| Champ                        | Type            | Description                                 |
| ---------------------------- | --------------- | ------------------------------------------- |
| `has_circular_dependencies`  | boolean         | `true` s'il existe au moins un cycle.       |
| `total_cycles`               | integer         | Nombre de cycles.                           |
| `total_modules_in_cycles`    | integer         | Modules impliqués dans des cycles.          |
| `circular_dependencies`      | array           | Tableau d'objets `CircularDependency`.      |
| `cycle_breaking_suggestions` | array of string | Suggestions pour briser les cycles.         |
| `core_infrastructure`        | array of string | Modules apparaissant dans plusieurs cycles. |

Énumération `CircularDependency.Severity` : `low`, `medium`, `high`, `critical`.

### Objet `coupling_analysis`

| Champ                     | Type            | Description                                               |
| ------------------------- | --------------- | --------------------------------------------------------- |
| `average_coupling`        | number          | Couplage moyen entre les modules.                         |
| `coupling_distribution`   | object          | Map de la valeur de couplage (clé entière) vers compteur. |
| `highly_coupled_modules`  | array of string | Modules à fort couplage.                                  |
| `loosely_coupled_modules` | array of string | Modules à faible couplage.                                |
| `average_instability`     | number          | Instabilité moyenne.                                      |
| `stable_modules`          | array of string | Modules à faible instabilité.                             |
| `instable_modules`        | array of string | Modules à forte instabilité.                              |
| `main_sequence_deviation` | number          | Distance moyenne à la séquence principale, `0`–`1`.       |
| `zone_of_pain`            | array of string | Modules stables + concrets.                               |
| `zone_of_uselessness`     | array of string | Modules instables + abstraits.                            |
| `main_sequence`           | array of string | Modules bien positionnés.                                 |

### Objet `architecture_analysis`

| Champ                     | Type            | Description                                                                                                                |
| ------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `compliance_score`        | number          | Score de conformité, `0`–`1`. Calculé comme `max(0, 1 - WeightedViolations / TotalRules)` ; `1.0` si aucune règle évaluée. |
| `total_violations`        | integer         | Nombre brut de violations (une par entrée de `Violations`).                                                                |
| `weighted_violations`     | integer         | Nombre de violations pondéré par sévérité, utilisé comme numérateur de `ComplianceScore` : `error × 5 + warning × 1`.      |
| `total_rules`             | integer         | Total des invocations de règles évaluées (dénominateur de `ComplianceScore`).                                              |
| `layer_analysis`          | object \| null  | Résultats d'analyse par couches.                                                                                           |
| `cohesion_analysis`       | object \| null  | Analyse de cohésion des paquets.                                                                                           |
| `responsibility_analysis` | object \| null  | Analyse des violations du principe SRP.                                                                                    |
| `violations`              | array           | Tableau d'objets `ArchitectureViolation`.                                                                                  |
| `severity_breakdown`      | object          | Map de la sévérité vers le compteur.                                                                                       |
| `recommendations`         | array           | Tableau d'objets `ArchitectureRecommendation`.                                                                             |
| `refactoring_targets`     | array of string | Modules nécessitant un refactoring.                                                                                        |

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
