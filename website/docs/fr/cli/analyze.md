# `pyscn analyze`

Exécute toutes les analyses disponibles sur des fichiers Python et produit un rapport.

```text
pyscn analyze [flags] <paths...>
```

`<paths...>` correspond à un ou plusieurs fichiers ou répertoires. Les répertoires sont parcourus récursivement en utilisant les `include_patterns` et `exclude_patterns` de votre configuration.

## Ce qu'elle fait

Par défaut, `analyze` exécute simultanément chaque analyseur activé :

- Complexité cyclomatique
- Détection de code mort
- Détection de clones (Type 1 à 4)
- Couplage entre classes (CBO)
- Cohésion des classes (LCOM4)
- Dépendances entre modules
- Validation des couches d'architecture

Les résultats sont combinés dans un rapport unique avec un [Score de Santé](../output/health-score.md).

## Options

### Format de sortie

Une seule de ces options peut être définie par invocation. Si aucune n'est définie, le HTML est généré.

| Option        | Description |
| ----------- | --- |
| `--html`    | Génère un rapport HTML (par défaut). |
| `--json`    | Génère un rapport JSON. |
| `--yaml`    | Génère un rapport YAML. |
| `--csv`     | Génère un résumé CSV (métriques uniquement, sans détail par constatation). |
| `--no-open` | N'ouvre pas le rapport HTML dans un navigateur. |

Les fichiers de sortie sont placés par défaut dans `.pyscn/reports/`, nommés `analyze_YYYYMMDD_HHMMSS.{ext}`. Configurez le répertoire avec `[output] directory = "..."`.

### Sélection des analyses

| Option | Description |
| --- | --- |
| `--select <list>` | N'exécute que les analyses listées. Séparées par des virgules : `complexity,deadcode,clones,cbo,lcom,deps`. |
| `--skip-complexity` | Ignore l'analyse de complexité. |
| `--skip-deadcode`   | Ignore la détection de code mort. |
| `--skip-clones`     | Ignore la détection de clones (l'analyse la plus lente). |
| `--skip-cbo`        | Ignore l'analyse de couplage entre classes. |
| `--skip-lcom`       | Ignore l'analyse de cohésion des classes. |
| `--skip-deps`       | Ignore l'analyse des dépendances entre modules. |

`--select` et `--skip-*` peuvent être combinées ; la sélection s'applique en premier, puis les exclusions.

### Surcharges rapides de seuils

| Option | Défaut | Description |
| --- | --- | --- |
| `--min-complexity <N>`    | `5`        | N'affiche que les fonctions de complexité ≥ N. |
| `--min-severity <level>`  | `warning`  | Sévérité minimale du code mort : `info`, `warning`, `critical`. |
| `--clone-threshold <F>`   | `0.65`     | Similarité minimale (0.0–1.0) pour la détection de clones. |
| `--min-cbo <N>`           | `0`        | N'affiche que les classes dont le CBO est ≥ N. |

### Configuration

| Option | Description |
| --- | --- |
| `-c, --config <path>` | Charge la configuration depuis un fichier spécifique au lieu de découvrir `.pyscn.toml` / `pyproject.toml`. |
| `-v, --verbose`        | Affiche la progression détaillée et les journaux par fichier. |

## Codes de sortie

| Code | Signification |
| --- | --- |
| `0` | Analyse terminée. Les problèmes signalés dans le rapport n'affectent pas le code de sortie. |
| `1` | Échec de l'analyse — arguments invalides, fichiers illisibles, erreurs d'analyse syntaxique. |

`analyze` n'échoue jamais le processus en fonction des constatations ; utilisez [`pyscn check`](check.md) pour une sémantique succès/échec.

## Exemples

```bash
# Analyse complète du répertoire courant avec rapport HTML
pyscn analyze .

# JSON pour les pipelines
pyscn analyze --json src/

# Ignorer l'analyseur le plus lent
pyscn analyze --skip-clones src/

# Uniquement complexité et code mort
pyscn analyze --select complexity,deadcode src/

# Seuils plus stricts
pyscn analyze --min-complexity 10 --min-severity critical src/

# Utiliser un fichier de configuration spécifique
pyscn analyze --config ./configs/strict.toml src/

# Ne pas ouvrir le navigateur (utile dans des sandbox ou des conteneurs)
pyscn analyze --no-open .
```

## Quand utiliser `analyze` vs `check`

| Cas d'usage | Commande |
| --- | --- |
| Développement local, « donnez-moi la vue d'ensemble » | `pyscn analyze` |
| Garde-fou qualité en CI avec succès/échec | [`pyscn check`](check.md) |
| Sortie lisible par machine pour outillage personnalisé | `pyscn analyze --json` |

## Voir aussi

- [Référence de configuration](../configuration/reference.md) — tous les réglages.
- [Score de Santé](../output/health-score.md) — comment le nombre 0–100 est calculé.
- [Schémas de sortie](../output/schemas.md) — définitions des champs JSON / YAML / CSV.
