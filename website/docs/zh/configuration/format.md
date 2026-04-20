# 配置文件格式

pyscn 从 **TOML** 读取配置。你可以将设置保存在专用的 `.pyscn.toml` 文件中，也可以放在 `pyproject.toml` 的 `[tool.pyscn]` 部分。

## 文件发现

当你运行 `pyscn analyze` 或 `pyscn check` 时，pyscn 会从目标路径向上查找：

1. `.pyscn.toml`（最高优先级）
2. 包含 `[tool.pyscn]` 部分的 `pyproject.toml`

找到的第一个文件将被使用。搜索会沿父目录逐级向上，直到匹配或到达文件系统根目录。如果两个文件都未找到，则使用内置默认值。

你也可以显式指定路径：

```bash
pyscn analyze --config ./configs/strict.toml src/
```

这将跳过自动发现。

## 优先级顺序

当同一设置出现在多个位置时，后者优先：

1. **内置默认值**（最低）
2. **`pyproject.toml` → `[tool.pyscn]`**
3. **`.pyscn.toml`**
4. **CLI 标志**（最高）

CLI 标志仅在**显式设置**时生效 — 未更改的默认值不会覆盖配置文件中的值。

## 两种文件风格

=== ".pyscn.toml"

    ```toml
    [complexity]
    max_complexity = 15

    [dead_code]
    min_severity = "critical"
    ```

=== "pyproject.toml"

    ```toml
    [tool.pyscn.complexity]
    max_complexity = 15

    [tool.pyscn.dead_code]
    min_severity = "critical"
    ```

如果同一目录中两个文件都存在，`.pyscn.toml` 优先。

## 生成初始配置文件

```bash
pyscn init
```

这会生成一个包含完整注释的 `.pyscn.toml`，其中列出每个选项、其默认值和简短说明。编辑你关心的值，其余的可以删除或保留不变。

```bash
pyscn init --force   # 覆盖已有文件
pyscn init --config tools/pyscn.toml   # 自定义路径
```

## 验证

pyscn 在加载时验证配置，如果有误则以退出码 `2` 退出。常见的验证规则：

- 复杂度阈值必须满足 `low ≥ 1` 且 `medium > low`。
- 输出格式必须是 `text`、`json`、`yaml`、`csv`、`html` 之一。
- 死代码严重性必须是 `info`、`warning` 或 `critical`。
- 克隆相似度阈值必须在 `[0.0, 1.0]` 范围内。
- 至少需要指定一个包含模式。

## 环境变量

pyscn **不会**从环境变量读取配置。MCP 服务器是一个例外：`PYSCN_CONFIG` 可以指向配置文件。

## 下一步

- [参考手册](reference.md) — 所有配置键的完整文档。
- [示例](examples.md) — 严格 CI、大型代码库、最小覆盖配置。
