# 配置示例

适用于常见场景的可复制粘贴的起始配置。

## 最小覆盖

仅设置几个严格的阈值；其余保持默认值。

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "critical"
```

## 严格 CI 门禁

在任何质量退化时使构建失败。与 `pyscn check` 配合使用。

```toml
[complexity]
max_complexity = 10

[dead_code]
min_severity = "warning"
detect_after_return = true
detect_after_raise = true
detect_unreachable_branches = true

[clones]
# 仅标记几乎完全相同的代码
similarity_threshold = 0.90
min_lines = 15

[cbo]
medium_threshold = 7

[dependencies]
enabled = true
detect_cycles = true
```

运行：

```bash
pyscn check --select complexity,deadcode,deps --max-cycles 0 src/
```

## 遗留代码库过渡期

你正在旧项目上引入 pyscn — 你希望获得分析信号而不会被大量失败所淹没。

```toml
[complexity]
max_complexity = 25    # 允许现有的复杂度

[dead_code]
min_severity = "critical"   # 仅报告最严重的问题

[clones]
min_lines = 20              # 仅报告较长的重复代码
similarity_threshold = 0.90

[analysis]
exclude_patterns = [
  "legacy/**",     # 隔离旧代码
  "**/_archive/*",
  "generated/**",
]
```

随着时间推移逐步收紧阈值。

## 大型代码库（10k+ 文件）

优化吞吐量。LSH 会自动启用，但需要提高并行度。

```toml
[clones]
lsh_enabled = true
max_goroutines = 16
max_memory_mb = 2048
batch_size = 500
timeout_seconds = 600
min_lines = 15           # 更少但更有意义的候选片段

[analysis]
exclude_patterns = [
  "**/test_*.py", "**/*_test.py",
  "**/migrations/**",
  "**/__generated__/**",
  "**/node_modules/**",
  ".venv/**", "venv/**",
]
```

## 整洁架构验证

强制分层架构：表现层 → 应用层 → 领域层，基础设施层在边缘。

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

## 数据密集型 ML / 研究代码库

在 notebook 转化而来的模块中，高复杂度是正常的。重点关注代码重复和死代码。

```toml
[complexity]
max_complexity = 30    # 数据管道天然具有较多分支

[dead_code]
min_severity = "critical"

[clones]
# 研究代码常常有几乎相同的实验变体；
# 提高阈值以避免被大量结果淹没
min_lines = 20
similarity_threshold = 0.85

[analysis]
exclude_patterns = [
  "notebooks/**",
  "experiments/**/*.ipynb",
]
```

## 与 `pyproject.toml` 共存

如果你已经有 `pyproject.toml`，可以将 pyscn 配置放在其中，而无需创建新文件：

```toml
# pyproject.toml
[project]
name = "my-package"
# ... 其他项目元数据

[tool.pyscn.complexity]
max_complexity = 15

[tool.pyscn.dead_code]
min_severity = "critical"

[tool.pyscn.clones]
similarity_threshold = 0.85
```

!!! note
    如果两者都存在，`.pyscn.toml` 优先于 `pyproject.toml`。建议只使用其中一个以避免混淆。

## 另请参阅

- [配置文件格式](format.md)
- [配置参考手册](reference.md)
