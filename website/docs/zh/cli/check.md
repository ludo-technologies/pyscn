# `pyscn check`

用于 CI/CD 流水线的质量门禁。以 linter 风格将检查结果输出到 **stderr**，如果任何问题超出阈值则以非零退出码退出。

```text
pyscn check [flags] [paths...]
```

路径默认为当前目录。

## 功能说明

`check` 是 [`analyze`](analyze.md) 的 CI 版本：

- **检查结果输出到 stderr**，使用 linter 格式（`file:line:col: message`）。
- 通过时**退出码 0**，任何失败（发现问题*或*执行错误）时**退出码 1**。
- **严格的默认值** -- 复杂度超过 10 的函数即为失败；任何循环依赖即为失败（当设置了 `--select deps` 时）。
- **速度快** -- 仅运行你选择的分析；跳过报告生成。

## 选项

### 分析选择

| 选项 | 说明 |
| --- | --- |
| `-s, --select <list>` | 仅运行指定的分析。可选值：`complexity`、`deadcode`、`clones`、`deps`（别名 `circular`）、`mockdata`、`di`。 |
| `--skip-clones`       | 不运行克隆检测。 |

默认（不使用 `--select`）：运行 `complexity`、`deadcode` **和 `clones`**。`deps`、`mockdata` 和 `di` 需要通过 `--select` 显式启用。使用 `--skip-clones` 可以在不切换到 `--select` 的情况下跳过克隆检测。

### 阈值覆盖

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| `--max-complexity <N>`   | `10` | 任何函数的圈复杂度超过此值即为失败。 |
| `--max-cycles <N>`       | `0`  | 循环依赖的最大数量，超出则失败。 |
| `--allow-dead-code`      | 关闭  | 将死代码视为警告，不触发检查失败。 |
| `--allow-circular-deps`  | 关闭  | 将循环依赖视为警告，不触发检查失败。 |

### 输出

| 选项 | 说明 |
| --- | --- |
| `-q, --quiet`          | 仅在发现问题时输出信息。 |
| `-c, --config <path>`  | 从指定文件加载配置。 |
| `-v, --verbose`        | 输出详细进度。 |

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 所有检查通过。 |
| `1` | 一项或多项检查失败，或出现执行错误。 |

`check` 不使用不同的退出码区分"发现问题"和"工具故障"。在 CI 中，依据 stderr 输出和 pyscn 的非零退出码来判断通过/失败即可。

## 示例

```bash
# 标准 CI 门禁（运行 complexity、deadcode、clones）
pyscn check .

# 更快的门禁：跳过克隆检测
pyscn check --skip-clones .

# 仅检查复杂度，对遗留代码使用更高的阈值
pyscn check --select complexity --max-complexity 15 src/

# 检查循环导入
pyscn check --select deps src/

# 在清理过程中允许现有死代码
pyscn check --allow-dead-code src/

# 检测 DI 反模式（需显式启用）
pyscn check --select di src/

# 静默模式 -- 适合 CI 日志
pyscn check --quiet .
```

## 与 `analyze` 的关系

`check` 使用与 `analyze` 相同的分析器和配置文件。区别如下：

| 方面 | `analyze` | `check` |
| --- | --- | --- |
| 输出 | 报告文件（HTML/JSON/YAML/CSV） | Linter 风格 stderr 输出 |
| 发现问题时的退出码 | 始终为 `0`（除非出错） | 任何问题超出阈值则退出 `1` |
| 克隆检测 | 默认开启 | 默认开启（用 `--skip-clones` 跳过） |
| 依赖分析 | 默认开启 | 默认关闭（通过 `--select deps` 启用） |
| 速度 | 较慢（所有分析器 + 报告生成） | 快（仅选定项，无报告） |
| 使用场景 | 交互式审查 | CI 质量门禁 |

两者配合使用：`analyze` 用于理解问题，`check` 用于防止回归。

## 另请参阅

- [CI/CD 集成](../integrations/ci-cd.md) -- GitHub Actions / pre-commit / GitLab 示例。
- [`pyscn analyze`](analyze.md) -- 带报告的完整分析。
