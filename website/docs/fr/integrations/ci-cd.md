# Intégration CI/CD

`pyscn check` renvoie un code de sortie non nul en cas de problème et produit une sortie de style linter. Voir la [référence `check`](../cli/check.md).

```bash
pyscn check .                        # exit 0 = pass, 1 = fail (issues or error)
pyscn check --max-complexity 15 .
pyscn check --select complexity,deadcode,deps .
```

Les constats sont écrits sur **stderr** au format linter. Capturez-les avec `2>` dans le shell ; la plupart des systèmes CI consignent stderr par défaut.

## GitHub Actions

Minimal (recommandé — uvx, sans étape d'installation) :

```yaml
# .github/workflows/quality.yml
name: Quality
on: [pull_request, push]

jobs:
  pyscn:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
      - run: uvx pyscn@latest check .
```

Avec le rapport complet en tant qu'artefact :

```yaml
jobs:
  pyscn:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3

      - name: Quality gate
        run: uvx pyscn@latest check --max-complexity 15 src/

      - name: Full report
        if: always()
        run: uvx pyscn@latest analyze --no-open --html src/

      - uses: actions/upload-artifact@v4
        if: always()
        with:
          name: pyscn-report
          path: .pyscn/reports/*.html
```

N'exécuter que sur les modifications Python :

```yaml
on:
  pull_request:
    paths:
      - '**/*.py'
      - '.pyscn.toml'
      - 'pyproject.toml'
```

Avec `pip` plutôt qu'uvx :

```yaml
      - uses: actions/setup-python@v5
        with:
          python-version: '3.12'
      - run: pip install pyscn
      - run: pyscn check .
```

## pre-commit

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: pyscn
        name: pyscn check
        entry: pyscn check
        language: python
        additional_dependencies: [pyscn]
        pass_filenames: false
        files: '\.py$'
```

Limiter aux fichiers indexés :

```yaml
      - id: pyscn
        name: pyscn check (staged)
        entry: bash -c 'pyscn check --quiet "$@"' --
        language: python
        additional_dependencies: [pyscn]
        files: '\.py$'
```

## GitLab CI

```yaml
# .gitlab-ci.yml
stages: [quality]

pyscn:
  stage: quality
  image: python:3.12-slim
  script:
    - pip install pyscn
    - pyscn check src/
  artifacts:
    when: always
    paths:
      - .pyscn/reports/
    expire_in: 1 week
  rules:
    - changes:
        - '**/*.py'
        - '.pyscn.toml'
```

## CircleCI

```yaml
version: 2.1
jobs:
  pyscn:
    docker:
      - image: cimg/python:3.12
    steps:
      - checkout
      - run: pip install pyscn
      - run: pyscn check .
      - store_artifacts:
          path: .pyscn/reports
          destination: pyscn-reports
```

## Bitbucket Pipelines

```yaml
# bitbucket-pipelines.yml
image: python:3.12
pipelines:
  default:
    - step:
        name: Code quality
        script:
          - pip install pyscn
          - pyscn check .
        artifacts:
          - .pyscn/reports/**
```

## Codes de sortie

| Code | Signification                        | Action |
| ---- | ------------------------------------ | ------ |
| `0`  | Aucun problème                       | Réussi |
| `1`  | Problèmes trouvés ou erreur d'exécution | Échec |

`check` renvoie le code `1` à la fois pour « problèmes ayant dépassé les seuils » et « l'analyse n'a pas pu se terminer » — les deux cas ne sont pas distinguables par le code de sortie. Inspectez stderr pour les différencier.

## Stratégies

Projet neuf :

```bash
pyscn check --max-complexity 10 --max-cycles 0 .
```

Adoption sur du code existant : commencez de manière permissive et resserrez à chaque sprint :

```bash
pyscn check --max-complexity 25 .
```

Monorepo avec des standards mixtes :

```bash
pyscn check --config packages/backend/.pyscn.toml packages/backend
pyscn check --config packages/tooling/.pyscn.toml packages/tooling
```

## Commentaire de PR à partir de JSON

`pyscn analyze --json` écrit dans un fichier horodaté sous `.pyscn/reports/`, pas sur stdout. Récupérez le fichier généré :

```bash
pyscn analyze --json --no-open .
report=$(ls -t .pyscn/reports/analyze_*.json | head -1)
jq -r '
  "## pyscn report\n" +
  "- **Health Score:** " + (.summary.health_score | tostring) + " / 100 (" + .summary.grade + ")\n" +
  "- Complexity: " + (.summary.complexity_score | tostring) + "\n" +
  "- Dead code: " + (.summary.dead_code_score | tostring) + "\n"
' "$report" > comment.md

gh pr comment $PR_NUMBER --body-file comment.md
```

## Voir aussi

- [`pyscn check`](../cli/check.md)
- [Exemples de configuration — Garde-fou strict pour la CI](../configuration/examples.md#strict-ci-gate)
- [Score de santé](../output/health-score.md)
