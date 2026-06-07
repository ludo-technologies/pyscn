# Référence de configuration

Chaque clé configurable dans `.pyscn.toml` (ou `[tool.pyscn.*]` dans `pyproject.toml`). Exécutez `pyscn init` pour générer un fichier de démarrage commenté.

---

## `[output]`

Contrôle la façon dont les résultats sont rapportés.

| Clé              | Type    | Défaut       | Description |
| ---------------- | ------- | ------------- | --- |
| `format`         | string  | `"text"`      | `text`, `json`, `yaml`, `csv` ou `html`. Les options en ligne de commande comme `--json` la surchargent. |
| `directory`      | string  | `""`          | Répertoire de sortie. Vide = `.pyscn/reports/` sous le CWD. |
| `show_details`   | bool    | `false`       | Inclut le détail par constatation dans le résumé. |
| `sort_by`        | string  | `"complexity"`| `name`, `complexity` ou `risk`. |
| `min_complexity` | int     | `1`           | Filtre les fonctions en dessous de cette complexité. Surcharge `[complexity].min_complexity` lorsqu'elle est définie. |

---

## `[complexity]`

Analyse de complexité cyclomatique.

| Clé                | Type | Défaut | Description |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | Exécute l'analyseur. |
| `low_threshold`    | int  | `9`     | Borne supérieure du « risque faible » (incluse). |
| `medium_threshold` | int  | `19`    | Borne supérieure du « risque moyen ». |
| `max_complexity`   | int  | `0`     | Seuil d'échec en CI. `0` = aucune limite. |
| `min_complexity`   | int  | `1`     | N'affiche pas les fonctions en dessous. |
| `report_unchanged` | bool | `true`  | Inclut les fonctions de complexité = 1. |

Voir [high-cyclomatic-complexity](../rules/high-cyclomatic-complexity.md) pour des recommandations de seuils.

---

## `[dead_code]`

Détection de code mort.

| Clé                              | Type   | Défaut      | Description |
| -------------------------------- | ------ | ------------ | --- |
| `enabled`                        | bool   | `true`       | Exécute l'analyseur. |
| `min_severity`                   | string | `"warning"`  | `info`, `warning` ou `critical`. |
| `show_context`                   | bool   | `false`      | Inclut les lignes source environnantes. |
| `context_lines`                  | int    | `3`          | Lignes de contexte (0–20). |
| `sort_by`                        | string | `"severity"` | `severity`, `line`, `file` ou `function`. |
| `detect_after_return`            | bool   | `true`       | Signale les instructions après `return`. |
| `detect_after_break`             | bool   | `true`       | Signale les instructions après `break`. |
| `detect_after_continue`          | bool   | `true`       | Signale les instructions après `continue`. |
| `detect_after_raise`             | bool   | `true`       | Signale les instructions après `raise`. |
| `detect_unreachable_branches`    | bool   | `true`       | Signale les branches qui ne peuvent jamais être empruntées. |
| `ignore_patterns`                | string[] | `[]`       | Motifs regex pour les lignes à ignorer. |

---

## `[clones]`

Détection de clones (l'analyseur le plus configurable).

### Sélection des fragments

| Clé              | Type | Défaut | Description |
| ---------------- | ---- | ------- | --- |
| `min_lines`      | int  | `10`    | Nombre minimal de lignes pour considérer un fragment. |
| `min_nodes`      | int  | `20`    | Nombre minimal de nœuds AST. |
| `skip_docstrings`| bool | `true`  | Ignore les docstrings lors du hachage. |

### Seuils par type (0.0–1.0)

| Clé                    | Défaut | Type de clone |
| ---------------------- | ------- | --- |
| `type1_threshold`      | `0.85`  | Identique (espaces/commentaires uniquement). |
| `type2_threshold`      | `0.75`  | Identifiants/littéraux renommés. |
| `type3_threshold`      | `0.70`  | Structurellement similaire avec modifications. |
| `type4_threshold`      | `0.65`  | Équivalence sémantique. |
| `similarity_threshold` | `0.65`  | Minimum global pour tout clone. |

### Algorithme

| Clé                 | Type   | Défaut    | Description |
| ------------------- | ------ | ---------- | --- |
| `cost_model_type`   | string | `"python"` | `default`, `python` ou `weighted`. |
| `ignore_literals`   | bool   | `false`    | Traite les littéraux différents comme équivalents. |
| `ignore_identifiers`| bool   | `false`    | Traite les noms de variables différents comme équivalents. |
| `max_edit_distance` | float  | `50.0`     | Plafond sur la distance d'édition d'arbre. |
| `enable_dfa`        | bool   | `true`     | Analyse de flot de données pour Type-4. |
| `enabled_clone_types` | string[] | tous     | Sous-ensemble de `type1`, `type2`, `type3`, `type4`. |

### Accélération LSH

| Clé                        | Type           | Défaut  | Description |
| -------------------------- | -------------- | -------- | --- |
| `lsh_enabled`              | `true\|false\|"auto"` | `"auto"` | Active LSH (`auto` = en fonction du nombre de fragments ou du nombre estimé de paires). |
| `lsh_auto_threshold`       | int            | `500`    | Seuil de fragments pour l'activation automatique ; auto s'active aussi au-delà de 10 000 paires estimées. |
| `lsh_similarity_threshold` | float          | `0.50`   | Pré-filtre des candidats LSH. |
| `lsh_bands`                | int            | `32`     | Bandes LSH. |
| `lsh_rows`                 | int            | `4`      | Lignes par bande. |
| `lsh_hashes`               | int            | `128`    | Nombre de fonctions de hachage. |

### Regroupement

| Clé                  | Type   | Défaut       | Description |
| -------------------- | ------ | ------------- | --- |
| `grouping_mode`      | string | `"connected"` | `connected`, `star`, `complete_linkage`, `k_core`. |
| `grouping_threshold` | float  | `0.65`        | Similarité minimale pour le regroupement. |
| `k_core_k`           | int    | `2`           | Paramètre k pour le mode `k_core`. |

### Performance

| Clé               | Type | Défaut | Description |
| ----------------- | ---- | ------- | --- |
| `max_memory_mb`   | int  | `100`   | Plafond mémoire (Mo). `0` = pas de limite. |
| `batch_size`      | int  | `100`   | Fichiers par lot. |
| `enable_batching` | bool | `true`  | Traitement par lots. |
| `max_goroutines`  | int  | `4`     | Workers concurrents. |
| `timeout_seconds` | int  | `300`   | Délai par analyse. |

### Filtrage de la sortie

| Clé             | Type  | Défaut          | Description |
| --------------- | ----- | --------------- | --- |
| `min_similarity`| float | `0.0`           | Filtre les paires en dessous. |
| `max_similarity`| float | `1.0`           | Filtre les paires au-dessus. |
| `max_results`   | int   | `10000`         | Nombre maximal de paires à rapporter. `0` = pas de limite. |
| `show_details`  | bool  | `false`         | Sortie verbeuse. |
| `show_content`  | bool  | `false`         | Inclut le code source dans le rapport. |
| `sort_by`       | string| `"similarity"`  | `similarity`, `size`, `location`, `type`. |
| `group_clones`  | bool  | `true`          | Regroupe les clones associés. |

---

## `[cbo]`

Coupling Between Objects (couplage entre classes).

| Clé                | Type | Défaut | Description |
| ------------------ | ---- | ------- | --- |
| `enabled`          | bool | `true`  | Exécute l'analyseur. |
| `low_threshold`    | int  | `3`     | Borne supérieure du « risque faible ». |
| `medium_threshold` | int  | `7`     | Borne supérieure du « risque moyen ». |
| `min_cbo`          | int  | `0`     | Filtre les classes dont le CBO est en dessous. |
| `max_cbo`          | int  | `0`     | Filtre les classes au-dessus. `0` = pas de limite. |
| `show_zeros`       | bool | `false` | Inclut les classes avec CBO = 0. |
| `include_builtins` | bool | `false` | Compte `list`/`dict`/`str` comme des dépendances. |
| `include_imports`  | bool | `true`  | Compte les références aux modules importés. |

---

## `[lcom]`

Lack of Cohesion of Methods (LCOM4).

| Clé                | Type | Défaut | Description |
| ------------------ | ---- | ------- | --- |
| `low_threshold`    | int  | `2`     | Borne supérieure du « risque faible » (bonne cohésion). |
| `medium_threshold` | int  | `5`     | Borne supérieure du « risque moyen ». |

---

## `[analysis]`

Règles de découverte des fichiers.

| Clé                | Type     | Défaut       | Description |
| ------------------ | -------- | ------------- | --- |
| `recursive`        | bool     | `true`        | Descend dans les sous-répertoires. |
| `follow_symlinks`  | bool     | `false`       | Suit les liens symboliques. |
| `include_patterns` | string[] | `["**/*.py"]` | Motifs glob à inclure. |
| `exclude_patterns` | string[] | voir ci-dessous | Motifs glob à exclure. |

`exclude_patterns` par défaut :

```toml
[
  "test_*.py", "*_test.py",
  "**/__pycache__/*", "**/*.pyc",
  "**/.pytest_cache/", ".tox/",
  "venv/", "env/", ".venv/", ".env/",
]
```

---

## `[architecture]`

Validation des couches. Toutes les clés sont facultatives — si vous ne définissez pas de couches, l'analyse architecturale s'exécute en mode permissif.

| Clé                        | Type  | Défaut | Description |
| -------------------------- | ----- | ------- | --- |
| `enabled`                  | bool  | `true`  | Exécute la validation des couches. |
| `style`                    | string | `""`   | Applique un préréglage intégré de couches + règles : `layered`, `hexagonal`, `clean`, `mvc`. Les `layers`/`rules` explicites ci-dessous remplacent le préréglage. Sans aucun `style`, `layers` ni `rules`, pyscn détecte automatiquement l'architecture au lieu d'appliquer un préréglage. |
| `validate_layers`          | bool  | `true`  | Vérifie les règles inter-couches. |
| `validate_cohesion`        | bool  | `true`  | Vérifie la cohésion des paquets. |
| `validate_responsibility`  | bool  | `true`  | Vérifie le nombre de responsabilités des modules. |
| `strict_mode`              | bool  | `true`  | Validation stricte. |
| `fail_on_violations`       | bool  | `false` | Code de sortie non nul en cas de violation. |
| `min_cohesion`             | float | `0.5`   | Cohésion minimale des paquets. |
| `max_coupling`             | int   | `10`    | Couplage inter-couches maximal. |
| `max_responsibilities`     | int   | `3`     | Préoccupations maximales par module. |
| `neutral_prefixes`         | string[] | `[]` | Segments de module de premier niveau à retirer avant la correspondance des paquets de couche. Utile lorsque chaque module commence par le même préfixe projet (ex. `app`, `src`). |

### Préréglages de style

Définissez `style` pour appliquer un ensemble prêt à l'emploi de définitions de couches et de règles plutôt que de les écrire à la main. Le préréglage est le point de départ ; tout `[[architecture.layers]]` ou `[[architecture.rules]]` que vous ajoutez le remplace (les règles utilisateur l'emportent par couche `from`). `layered` reproduit les couches/règles de la configuration par défaut générée ; sans aucun `style`, `layers` ni `rules`, pyscn détecte automatiquement l'architecture au lieu d'appliquer `layered`.

| Préréglage | Impose | Règle notable |
| --- | --- | --- |
| `layered` | `presentation → application → domain → infrastructure` classique (base de la configuration par défaut générée). | `domain` peut encore atteindre `infrastructure` (DIP souple). |
| `hexagonal` | Ports & Adapters / Onion : `domain` ne dépend de rien vers l'extérieur. | `domain → ports`/`adapters` refusé. |
| `clean` | Clean Architecture : les dépendances ne pointent que vers l'intérieur sur `entities → use_cases → interface_adapters → frameworks`. | Toute dépendance vers l'extérieur refusée. |
| `mvc` | MVC / MVT : `model` / `view` / `controller`. | Une dépendance directe `view → model` est **déconseillée** (avertissement), pas refusée. |

```toml
[architecture]
style = "hexagonal"
```

Les règles utilisent trois listes, évaluées dans l'ordre : `deny` (erreur), `warn` (autorisé mais signalé comme avertissement — utilisé par `mvc` pour `view → model`), puis `allow` (lorsqu'une liste allow non vide est présente, tout ce qui n'est pas listé devient une erreur). Un avertissement réduit moins le score de conformité d'architecture qu'une erreur.

### Définitions de couches

```toml
[[architecture.layers]]
name = "presentation"
packages = ["router", "routers", "handler", "handlers", "controller", "api"]

[[architecture.layers]]
name = "application"
packages = ["service", "services", "usecase", "usecases"]

[[architecture.layers]]
name = "domain"
packages = ["model", "models", "entity", "entities"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repository", "repositories", "db", "database"]
```

### Règles de couches

```toml
[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain"]
```

### Préfixes neutres

Si chaque module du projet commence par le même segment racine (`app.`, `src.`, ...), la correspondance des couches peut échouer car le préfixe projet masque le nom de la couche. Listez ces segments sous `neutral_prefixes` et pyscn les retirera avant de résoudre un module vers une couche :

```toml
[architecture]
neutral_prefixes = ["app", "src"]
```

Avec ce réglage, `app.routers.user_router` est mis en correspondance comme `routers.user_router` et résolu vers la couche `presentation`.

---

## `[dependencies]`

Analyse des dépendances entre modules. **Opt-in** pour `pyscn check` ; toujours active pour `pyscn analyze` sauf si ignorée.

| Clé                  | Type   | Défaut | Description |
| -------------------- | ------ | ------- | --- |
| `enabled`            | bool   | `false` | Exécute l'analyseur (analyze l'exécute toujours quoi qu'il en soit). |
| `include_stdlib`     | bool   | `false` | Inclut les imports de la bibliothèque standard. |
| `include_third_party`| bool   | `true`  | Inclut les imports tiers. |
| `follow_relative`    | bool   | `true`  | Suit les imports relatifs. |
| `detect_cycles`      | bool   | `true`  | Détecte les imports circulaires. |
| `calculate_metrics`  | bool   | `true`  | Calcule Ca/Ce/I/A/D. |
| `find_long_chains`   | bool   | `true`  | Signale les plus longues chaînes de dépendance. |
| `cycle_reporting`    | string | `"summary"` | `all`, `critical`, `summary`. |
| `max_cycles_to_show` | int    | `10`    | Plafond sur les cycles rapportés. |
| `sort_by`            | string | `"name"` | `name`, `coupling`, `instability`, `distance`, `risk`. |
| `show_matrix`        | bool   | `false` | Inclut la matrice de dépendances. |
| `generate_dot_graph` | bool   | `false` | Émet une sortie Graphviz DOT. |

---

## `[mock_data]`

Détection de données factices / placeholder. **Opt-in**.

| Clé              | Type     | Défaut      | Description |
| ---------------- | -------- | ----------- | --- |
| `enabled`        | bool     | `false`     | Exécute l'analyseur. |
| `min_severity`   | string   | `"warning"` | `info`, `warning`, `error`. |
| `ignore_tests`   | bool     | `true`      | Ignore les fichiers de test. |
| `keywords`       | string[] | intégré     | Mots signalés comme indicateurs de mock. |
| `domains`        | string[] | intégré     | Domaines signalés (`example.com`, `test.com`, etc.). |
| `ignore_patterns`| string[] | `[]`        | Fichiers/motifs regex à ignorer. |

---

## `[di]`

Détection d'anti-patterns d'injection de dépendances. **Opt-in**.

| Clé                            | Type   | Défaut      | Description |
| ------------------------------ | ------ | ----------- | --- |
| `enabled`                      | bool   | `false`     | Exécute l'analyseur. |
| `min_severity`                 | string | `"warning"` | `info`, `warning`, `error`. |
| `constructor_param_threshold`  | int    | `5`         | Signale les `__init__` avec plus de paramètres. |

---

## Correspondance option CLI → clé de configuration

Les options qui ne se mappent pas directement sur une clé de configuration (`--select`, `--skip-*`, `--no-open`) fonctionnent par-dessus la configuration chargée.

| Option CLI                | Clé de configuration              |
| ----------------------- | --------------------------------- |
| `--config <path>`       | — (contourne la découverte)       |
| `--json/--yaml/--csv/--html` | `[output] format`            |
| `--min-complexity`      | `[complexity] min_complexity`     |
| `--max-complexity`      | `[complexity] max_complexity`     |
| `--min-severity`        | `[dead_code] min_severity`        |
| `--clone-threshold`     | `[clones] similarity_threshold`   |
| `--min-cbo`             | `[cbo] min_cbo`                   |
| `--max-cycles`          | — (commande check uniquement)     |

## Voir aussi

- [Format du fichier de configuration](format.md) — découverte et priorité.
- [Exemples](examples.md) — configurations prêtes à l'emploi.
