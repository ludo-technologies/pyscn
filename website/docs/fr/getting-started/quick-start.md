# Démarrage rapide

## Lancer une analyse

```bash
uvx pyscn@latest analyze .
```

Si `pyscn` est déjà installé (via `uv tool install pyscn`, `pipx install pyscn`, ou `pip install pyscn`), supprimez le préfixe `uvx pyscn@latest` :

```bash
pyscn analyze .
```

Écrit un rapport HTML dans `.pyscn/reports/analyze_YYYYMMDD_HHMMSS.html` et l'ouvre dans le navigateur par défaut.

## Choisir le format de sortie

```bash
pyscn analyze --json .
pyscn analyze --yaml .
pyscn analyze --csv .
pyscn analyze --no-open .       # supprime l'ouverture du navigateur
```

## Exécuter des analyseurs spécifiques

```bash
pyscn analyze --select complexity .
pyscn analyze --select complexity,deadcode .
pyscn analyze --skip-clones .
```

Voir [`analyze`](../cli/analyze.md) pour toutes les options.

## Garde-fou de qualité en CI

```bash
pyscn check .                              # sortie 0 = succès, 1 = échec
pyscn check --max-complexity 15 src/
pyscn check --select complexity,deadcode,deps src/
```

Voir [`check`](../cli/check.md) et [Intégration CI/CD](../integrations/ci-cd.md).

## Générer un fichier de configuration

```bash
pyscn init
```

Crée `.pyscn.toml` avec toutes les options commentées. Voir la [Référence de configuration](../configuration/reference.md).
