# `pyscn init`

生成 `.pyscn.toml` 配置文件，所有选项均带有行内文档说明。

```text
pyscn init [flags]
```

## 功能说明

生成一个带注释的 TOML 文件，包含最常调整的配置节：

- `[output]`、`[complexity]`、`[dead_code]`、`[clones]`、`[cbo]`、`[analysis]`、`[architecture]`（附带 `[[architecture.layers]]` 和 `[[architecture.rules]]` 示例）
- 填入默认值
- 注释说明每个配置项

生成的文件**不包含**所有可配置节。LCOM4 内聚度（`[lcom]`）、模块依赖分析（`[dependencies]`）、mock 数据检测（`[mock_data]`）和 DI 反模式（`[di]`）的选项是有效的，但需要手动添加。详见[配置参考](../configuration/reference.md)。

文件创建后，该项目（及其所有子目录）中后续的每次 `pyscn analyze` / `pyscn check` 运行都会自动加载它。

## 选项

| 选项 | 默认值 | 说明 |
| --- | --- | --- |
| `-c, --config <path>` | `.pyscn.toml` | 输出文件路径。 |
| `-f, --force`         | 关闭          | 覆盖已有文件。 |

## 退出码

| 退出码 | 含义 |
| --- | --- |
| `0` | 文件写入成功。 |
| `1` | 文件已存在（使用 `--force` 覆盖）或写入失败。 |

## 示例

```bash
# 在当前目录创建 .pyscn.toml
pyscn init

# 使用自定义文件名
pyscn init --config tools/pyscn.toml

# 覆盖已有配置
pyscn init --force
```

## 建议优先调整的设置

运行 `init` 后，大多数项目最终会调整的设置：

| 设置 | 典型调整 |
| --- | --- |
| `[complexity].max_complexity` | 根据 CI 的严格程度设为 `10`、`15` 或 `20`。 |
| `[dead_code].min_severity`     | 如果警告太多，提高到 `"critical"`。 |
| `[clones].similarity_threshold`| 降低到 `0.80` 以发现更多克隆，提高到 `0.90` 以减少噪音。 |
| `[analysis].exclude_patterns`  | 添加生成代码路径、数据库迁移等。 |

详见完整的[配置参考](../configuration/reference.md)。

## 另请参阅

- [配置参考](../configuration/reference.md) -- 所有选项说明。
- [配置示例](../configuration/examples.md) -- 严格 CI、大型代码库、最小化覆盖等场景。
