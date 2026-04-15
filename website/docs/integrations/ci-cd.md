# CI/CD Integration

`pyscn check` exits non-zero on issues and produces linter-style output. See the [`check` reference](../cli/check.md).

```bash
pyscn check .                        # exit 0 = pass, 1 = fail (issues or error)
pyscn check --max-complexity 15 .
pyscn check --select complexity,deadcode,deps .
```

Findings are written to **stderr** in linter format. Capture them with `2>` in shell; most CI systems log stderr by default.

## GitHub Actions

Minimal (recommended — uvx, no install step):

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

With full report as an artifact:

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

Run only on Python changes:

```yaml
on:
  pull_request:
    paths:
      - '**/*.py'
      - '.pyscn.toml'
      - 'pyproject.toml'
```

With `pip` instead of uvx:

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

Scope to staged files:

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

## Exit codes

| Code | Meaning | Action |
| --- | --- | --- |
| `0` | No issues | Pass |
| `1` | Issues found or execution error | Fail |

`check` returns exit `1` for both "issues exceeded thresholds" and "analysis could not complete" — the two cases are not distinguishable by exit code. Inspect stderr to tell them apart.

## Strategies

Green-field:

```bash
pyscn check --max-complexity 10 --max-cycles 0 .
```

Legacy adoption: start permissive and tighten each sprint:

```bash
pyscn check --max-complexity 25 .
```

Monorepo with mixed standards:

```bash
pyscn check --config packages/backend/.pyscn.toml packages/backend
pyscn check --config packages/tooling/.pyscn.toml packages/tooling
```

## PR comment from JSON

`pyscn analyze --json` writes to a timestamped file in `.pyscn/reports/`, not stdout. Pick up the generated file:

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

## See also

- [`pyscn check`](../cli/check.md)
- [Configuration Examples — Strict CI gate](../configuration/examples.md#strict-ci-gate)
- [Health Score](../output/health-score.md)
