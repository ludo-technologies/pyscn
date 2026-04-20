# `pyscn analyze`

对 Python 文件运行所有可用分析并生成报告。

```text
pyscn analyze [flags] <paths...>
```

`<paths...>` 是一个或多个文件或目录。目录会根据配置中的 `include_patterns` 和 `exclude_patterns` 递归遍历。

## 功能说明

默认情况下，`analyze` 会并发运行所有已启用的分析器：

- 圈复杂度
- 死代码检测
- 克隆检测（Type 1-4）
- 类耦合度（CBO）
- 类内聚度（LCOM4）
- 模块依赖分析
- 架构层验证

分析结果汇总到一份报告中，包含[健康评分](../output/health-score.md)。

## 选项

### 输出格式

每次调用只能设置其中一个。如果均未设置，默认生成 HTML。

| 选项        | 说明 |
| ----------- | --- |
| `--html`    | 生成 HTML 报告（默认）。 |
| `--json`    | 生成 JSON 报告。 |
| `--yaml`    | 生成 YAML 报告。 |
| `--csv`     | 生成 CSV 摘要（仅包含度量数据，不含逐项详情）。 |
| `--no-open` | 不在浏览器中打开 HTML 报告。 |

输出文件默认保存在 `.pyscn/reports/` 目录下，文件名格式为 `analyze_YYYYMMDD_HHMMSS.{ext}`。可通过 `[output] directory = "..."` 配置输出目录。

### 分析选择

| 选项 | 说明 |
| --- | --- |
| `--select <list>` | 仅运行指定的分析。逗号分隔：`complexity,deadcode,clones,cbo,lcom,deps`。 |
| `--skip-complexity` | 跳过复杂度分析。 |
| `--skip-deadcode`   | 跳过死代码检测。 |
| `--skip-clones`     | 跳过克隆检测（最慢的分析）。 |
| `--skip-cbo`        | 跳过类耦合度分析。 |
| `--skip-lcom`       | 跳过类内聚度分析。 |
| `--skip-deps`       | 跳过模块依赖分析。 |

`--select` 和 `--skip-*` 可以组合使用；先应用选择，再应用跳过。

### 快速阈值覆盖

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| `--min-complexity <N>`    | `5`        | 仅报告复杂度 >= N 的函数。 |
| `--min-severity <level>`  | `warning`  | 死代码最低严重级别：`info`、`warning`、`critical`。 |
| `--clone-threshold <F>`   | `0.65`     | 克隆检测的最低相似度（0.0-1.0）。 |
| `--min-cbo <N>`           | `0`        | 仅报告 CBO >= N 的类。 |

### 配置

| 选项 | 说明 |
| --- | --- |
| `-c, --config <path>` | 从指定文件加载配置，而非自动发现 `.pyscn.toml` / `pyproject.toml`。 |
| `-v, --verbose`        | 输出详细进度和逐文件日志。 |

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 分析完成。报告中发现的问题不影响退出码。 |
| `1` | 分析失败 -- 无效参数、无法读取的文件、解析错误。 |

`analyze` 不会因为发现问题而使进程失败；如需通过/失败语义，请使用 [`pyscn check`](check.md)。

## 示例

```bash
# 对当前目录进行完整分析，生成 HTML 报告
pyscn analyze .

# 在流水线中生成 JSON
pyscn analyze --json src/

# 跳过最慢的分析器
pyscn analyze --skip-clones src/

# 仅分析复杂度和死代码
pyscn analyze --select complexity,deadcode src/

# 更严格的阈值
pyscn analyze --min-complexity 10 --min-severity critical src/

# 使用指定的配置文件
pyscn analyze --config ./configs/strict.toml src/

# 不打开浏览器（适用于沙箱或容器环境）
pyscn analyze --no-open .
```

## `analyze` 与 `check` 的使用场景

| 使用场景 | 命令 |
| --- | --- |
| 本地开发，查看全面分析结果 | `pyscn analyze` |
| CI 中的通过/失败质量门禁 | [`pyscn check`](check.md) |
| 机器可读输出，用于自定义工具 | `pyscn analyze --json` |

## 另请参阅

- [配置参考](../configuration/reference.md) -- 所有可配置项。
- [健康评分](../output/health-score.md) -- 0-100 分的计算方式。
- [输出模式](../output/schemas.md) -- JSON / YAML / CSV 字段定义。
