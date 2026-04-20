# CI/CD 連携

`pyscn check` は問題が見つかった場合に非ゼロで終了し、リンター形式の出力を生成します。[`check` リファレンス](../cli/check.md)を参照してください。

```bash
pyscn check .                        # exit 0 = パス, 1 = 失敗（問題またはエラー）
pyscn check --max-complexity 15 .
pyscn check --select complexity,deadcode,deps .
```

検出結果はリンター形式で **stderr** に書き込まれます。シェルで `2>` を使用してキャプチャできます。ほとんどの CI システムはデフォルトで stderr をログに記録します。

## GitHub Actions

最小構成（推奨 — uvx を使用、インストールステップ不要）:

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

アーティファクトとして完全なレポートを含む構成:

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

Python ファイルの変更時のみ実行:

```yaml
on:
  pull_request:
    paths:
      - '**/*.py'
      - '.pyscn.toml'
      - 'pyproject.toml'
```

uvx の代わりに `pip` を使用:

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

ステージされたファイルに限定:

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

## 終了コード

| コード | 意味 | アクション |
| --- | --- | --- |
| `0` | 問題なし | パス |
| `1` | 問題が見つかった、または実行エラー | 失敗 |

`check` は「閾値を超える問題が見つかった」場合と「分析を完了できなかった」場合の両方で終了コード `1` を返します。2つのケースは終了コードでは区別できません。stderr を確認して判別してください。

## 戦略

グリーンフィールド:

```bash
pyscn check --max-complexity 10 --max-cycles 0 .
```

レガシーの導入: 寛容な設定から始めて、スプリントごとに厳しくしていきます:

```bash
pyscn check --max-complexity 25 .
```

異なる基準を持つモノレポ:

```bash
pyscn check --config packages/backend/.pyscn.toml packages/backend
pyscn check --config packages/tooling/.pyscn.toml packages/tooling
```

## JSON からの PR コメント

`pyscn analyze --json` は `.pyscn/reports/` 内のタイムスタンプ付きファイルに書き込みます（stdout ではありません）。生成されたファイルを取得します:

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

## 関連項目

- [`pyscn check`](../cli/check.md)
- [設定例 — 厳格な CI ゲート](../configuration/examples.md#strict-ci-gate)
- [ヘルススコア](../output/health-score.md)
