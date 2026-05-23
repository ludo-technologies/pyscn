# FAQ

## Généralités

### En quoi pyscn diffère-t-il de ruff, pylint ou mypy ?

Ruff et pylint sont des linters ; mypy est un vérificateur de types. pyscn est un analyseur structurel : il construit des graphes de flux de contrôle, des représentations arborescentes et des graphes de dépendances pour mesurer la complexité, l'accessibilité, la duplication et le couplage. Ils sont complémentaires.

### Que ne détecte pas pyscn ?

- Les bugs à l'exécution
- Les vulnérabilités de sécurité
- Les problèmes de performance
- Les violations de style (utilisez ruff)
- Les erreurs de typage (utilisez mypy / pyright)

### pyscn a-t-il besoin d'un accès réseau ?

Non. pyscn s'exécute entièrement en local. Aucune télémétrie, aucun appel distant.

### Fonctionne-t-il avec du code asynchrone ?

Oui. `async def` et `await` sont analysés de la même manière que le code synchrone.

### Puis-je analyser des notebooks Jupyter ?

Non. Convertissez d'abord avec `jupyter nbconvert --to script`.

## Configuration

### Où placer mon fichier de configuration ?

Placez `.pyscn.toml` à la racine du dépôt. pyscn le découvre en remontant depuis les fichiers analysés.

### J'ai déjà `pyproject.toml`. Devrais-je utiliser `[tool.pyscn]` à la place ?

Les deux fonctionnent. `.pyscn.toml` l'emporte si les deux existent.

### Comment exclure le code généré / les migrations / les dépendances vendues ?

```toml
[analysis]
exclude_patterns = [
  "**/migrations/**",
  "**/__generated__/**",
  "vendor/**",
]
```

### Puis-je avoir des seuils différents selon les parties du projet ?

Pas dans un seul fichier de configuration. Utilisez des configurations par répertoire :

```bash
pyscn check --config backend/.pyscn.toml backend/
pyscn check --config scripts/.pyscn.toml scripts/
```

## Exécution de pyscn

### Le rapport HTML n'a pas ouvert mon navigateur.

L'ouverture automatique est supprimée lorsque stdin n'est pas un TTY, en SSH, ou lorsque `CI` est défini. Le chemin du rapport est imprimé sur stderr. Forcez avec `--no-open`.

### `pyscn analyze` est lent sur mon dépôt.

- `--skip-clones` (les clones sont l'analyseur le plus lent)
- Restreignez le périmètre : `pyscn analyze src/`
- Augmentez `min_lines` / `min_nodes` sous `[clones]`
- Augmentez `max_goroutines` sous `[clones]`

### J'obtiens des erreurs d'analyse syntaxique sur du code Python valide.

pyscn utilise tree-sitter, qui prend en charge Python jusqu'à la 3.13. Ouvrez une issue avec une reproduction minimale.

### pyscn peut-il corriger automatiquement les problèmes ?

Non. pyscn rapporte ; il ne modifie jamais le code source.

## Scores et seuils

### Mon score de santé a chuté du jour au lendemain.

Vérifiez les scores par catégorie. Causes fréquentes :

- Une nouvelle fonction volumineuse a augmenté la complexité moyenne.
- Un refactoring a laissé du code mort.
- Un copier-coller a créé des clones.
- Un nouvel import a introduit un cycle.

Comparez la sortie JSON entre deux exécutions pour identifier le changement.

### Pourquoi mon score de couplage est-il si bas sur une petite base de code ?

La pénalité est calculée en pourcentage : un ratio problématique de 3/10 produit la même pénalité que 300/1000. Pour les projets de moins de 20 classes, examinez les valeurs CBO brutes plutôt que le score de la catégorie.

### Qu'est-ce qu'un « bon » score de santé ?

| Plage | Signification |
| --- | --- |
| 90+ | Excellent |
| 70–90 | Normal pour une base de code saine |
| 50–70 | Travail réel nécessaire ; récupérable |
| < 50 | Refactoring ciblé requis |

La tendance compte plus que la valeur absolue.

## MCP

### L'assistant voit les outils pyscn, mais les appels échouent.

Vérifiez que le binaire est dans le PATH, ou utilisez `uvx pyscn-mcp`. Testez directement :

```bash
uvx pyscn-mcp
```

Voir le [guide MCP](integrations/mcp.md).

### Le serveur MCP peut-il refactorer mon code ?

Non. Le MCP de pyscn est en lecture seule.

## Dépannage

### `pyscn: command not found` après installation via pip.

L'emplacement d'installation n'est pas dans le PATH. Inspectez avec :

```bash
python -m pip show -f pyscn | grep bin/pyscn
```

Sous Linux/macOS, ajoutez `~/.local/bin` au PATH, ou installez avec `uvx pyscn@latest <command>` (sans étape d'installation), `uv tool install pyscn`, ou `pipx install pyscn`.

### Le rapport affiche 0 fichier analysé.

Les motifs include/exclude ont tout exclu. Valeurs par défaut : `include_patterns = ["**/*.py"]` ; les exclusions comprennent `test_*.py` et `*_test.py`. Surchargez pour analyser les tests :

```toml
[analysis]
exclude_patterns = [
  "**/__pycache__/*",
  "**/*.pyc",
  ".venv/**",
]
```

### `Warning: parse error in <file>`.

Le fichier contient une erreur de syntaxe dont tree-sitter n'a pas pu récupérer. Le fichier est ignoré ; les autres fichiers sont analysés normalement.

## Obtenir de l'aide

- [GitHub Issues](https://github.com/ludo-technologies/pyscn/issues) — bugs et demandes de fonctionnalités.
- [GitHub Discussions](https://github.com/ludo-technologies/pyscn/discussions) — questions et idées.
- [Code source](https://github.com/ludo-technologies/pyscn).
