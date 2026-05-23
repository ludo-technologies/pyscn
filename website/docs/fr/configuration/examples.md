# Exemples de configuration

Points de départ à copier-coller pour les scénarios courants.

## Surcharge minimale

Juste quelques seuils stricts ; tout le reste reste par défaut.

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "critical"
```

## Garde-fou CI strict { #strict-ci-gate }

Faites échouer le build en cas de régression de qualité. À associer à `pyscn check`.

```toml
[complexity]
max_complexity = 10

[dead_code]
min_severity = "warning"
detect_after_return = true
detect_after_raise = true
detect_unreachable_branches = true

[clones]
# Ne signaler que du code quasi-identique
similarity_threshold = 0.90
min_lines = 15

[cbo]
medium_threshold = 7

[dependencies]
enabled = true
detect_cycles = true
```

Exécution :

```bash
pyscn check --select complexity,deadcode,deps --max-cycles 0 src/
```

## Période de grâce pour un codebase hérité

Vous adoptez pyscn sur un projet plus ancien — vous voulez du signal sans un déluge immédiat d'échecs.

```toml
[complexity]
max_complexity = 25    # autorise la complexité existante

[dead_code]
min_severity = "critical"   # uniquement le pire

[clones]
min_lines = 20              # uniquement les duplications longues
similarity_threshold = 0.90

[analysis]
exclude_patterns = [
  "legacy/**",     # mettre en quarantaine l'ancien code
  "**/_archive/*",
  "generated/**",
]
```

Resserrez les seuils progressivement au fil du temps.

## Gros codebase (10 000+ fichiers)

Optimisez pour le débit. LSH s'active automatiquement, mais poussez le parallélisme.

```toml
[clones]
lsh_enabled = true
max_goroutines = 16
max_memory_mb = 2048
batch_size = 500
timeout_seconds = 600
min_lines = 15           # candidats moins nombreux et plus significatifs

[analysis]
exclude_patterns = [
  "**/test_*.py", "**/*_test.py",
  "**/migrations/**",
  "**/__generated__/**",
  "**/node_modules/**",
  ".venv/**", "venv/**",
]
```

## Validation d'architecture en couches

Imposez une architecture en couches : presentation → application → domain, infrastructure à la périphérie.

```toml
[architecture]
enabled = true
strict_mode = true
fail_on_violations = true

[[architecture.layers]]
name = "presentation"
packages = ["api", "routers", "handlers", "views"]

[[architecture.layers]]
name = "application"
packages = ["services", "usecases"]

[[architecture.layers]]
name = "domain"
packages = ["models", "entities", "core"]

[[architecture.layers]]
name = "infrastructure"
packages = ["repositories", "db", "adapters", "clients"]

[[architecture.rules]]
from = "presentation"
allow = ["application", "domain"]
deny = ["infrastructure"]

[[architecture.rules]]
from = "application"
allow = ["domain", "infrastructure"]
deny = ["presentation"]

[[architecture.rules]]
from = "domain"
deny = ["presentation", "application", "infrastructure"]
```

## Codebase ML / recherche orienté données

Une forte complexité est attendue dans les modules issus de notebooks. Concentrez-vous sur la duplication et le code mort.

```toml
[complexity]
max_complexity = 30    # les pipelines de données sont naturellement très branchés

[dead_code]
min_severity = "critical"

[clones]
# Le code de recherche comporte souvent des variantes d'expérience quasi-identiques ;
# relevez les seuils pour ne pas être submergé
min_lines = 20
similarity_threshold = 0.85

[analysis]
exclude_patterns = [
  "notebooks/**",
  "experiments/**/*.ipynb",
]
```

## Cohabitation avec `pyproject.toml`

Si vous avez déjà un `pyproject.toml`, vous pouvez y placer la configuration pyscn au lieu de créer un nouveau fichier :

```toml
# pyproject.toml
[project]
name = "my-package"
# ... autres métadonnées du projet

[tool.pyscn.complexity]
max_complexity = 15

[tool.pyscn.dead_code]
min_severity = "critical"

[tool.pyscn.clones]
similarity_threshold = 0.85
```

!!! note
    `.pyscn.toml` est prioritaire sur `pyproject.toml` si les deux existent. Choisissez-en un pour éviter la confusion.

## Voir aussi

- [Format du fichier de configuration](format.md)
- [Référence de configuration](reference.md)
