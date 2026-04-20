# CI/CD 集成

`pyscn check` 在发现问题时以非零退出码退出，并产生 linter 风格的输出。详见 [`check` 参考](../cli/check.md)。

```bash
pyscn check .                        # exit 0 = 通过, 1 = 失败（问题或错误）
pyscn check --max-complexity 15 .
pyscn check --select complexity,deadcode,deps .
```

发现结果以 linter 格式写入 **stderr**。在 shell 中通过 `2>` 捕获；大多数 CI 系统默认记录 stderr。

## GitHub Actions

最简配置（推荐 — 使用 uvx，无需安装步骤）：

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

生成完整报告并作为构件上传：

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

仅在 Python 文件变更时运行：

```yaml
on:
  pull_request:
    paths:
      - '**/*.py'
      - '.pyscn.toml'
      - 'pyproject.toml'
```

使用 `pip` 代替 uvx：

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

限定为暂存文件：

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

## 退出码

| 代码 | 含义 | 操作 |
| --- | --- | --- |
| `0` | 无问题 | 通过 |
| `1` | 发现问题或执行错误 | 失败 |

`check` 对"问题超出阈值"和"分析无法完成"都返回退出码 `1` — 两种情况无法通过退出码区分。检查 stderr 以区分它们。

## 策略

全新项目：

```bash
pyscn check --max-complexity 10 --max-cycles 0 .
```

遗留项目引入：从宽松开始，每个迭代逐步收紧：

```bash
pyscn check --max-complexity 25 .
```

混合标准的单仓库：

```bash
pyscn check --config packages/backend/.pyscn.toml packages/backend
pyscn check --config packages/tooling/.pyscn.toml packages/tooling
```

## 从 JSON 生成 PR 评论

`pyscn analyze --json` 将结果写入 `.pyscn/reports/` 下的带时间戳文件，而非标准输出。获取生成的文件：

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

## 另请参阅

- [`pyscn check`](../cli/check.md)
- [配置示例 — 严格 CI 门禁](../configuration/examples.md#strict-ci-gate)
- [健康评分](../output/health-score.md)
