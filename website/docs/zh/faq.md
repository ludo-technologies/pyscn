# 常见问题

## 基本概念

### pyscn 与 ruff、pylint 或 mypy 有什么区别？

Ruff 和 pylint 是代码检查工具（linter）；mypy 是类型检查器。pyscn 是结构化分析工具：它构建控制流图、语法树表示和依赖关系图，用于衡量复杂度、可达性、重复度和耦合度。它们之间是互补关系。

### pyscn 不能检测什么？

- 运行时错误
- 安全漏洞
- 性能问题
- 代码风格问题（请使用 ruff）
- 类型错误（请使用 mypy / pyright）

### pyscn 需要网络访问吗？

不需要。pyscn 完全在本地运行，无遥测数据上报，无远程调用。

### 支持异步代码吗？

支持。`async def` 和 `await` 的分析方式与同步代码相同。

### 能分析 Jupyter notebook 吗？

不能。请先使用 `jupyter nbconvert --to script` 转换为 Python 脚本。

## 配置

### 配置文件应该放在哪里？

将 `.pyscn.toml` 放在仓库根目录。pyscn 会从被分析文件所在位置向上查找配置文件。

### 我已经有 `pyproject.toml` 了，应该用 `[tool.pyscn]` 吗？

两种方式都可以。如果两者同时存在，`.pyscn.toml` 优先。

### 如何排除生成的代码 / 数据库迁移 / 第三方依赖？

```toml
[analysis]
exclude_patterns = [
  "**/migrations/**",
  "**/__generated__/**",
  "vendor/**",
]
```

### 能为项目的不同部分设置不同的阈值吗？

单个配置文件不支持。可以使用目录级配置：

```bash
pyscn check --config backend/.pyscn.toml backend/
pyscn check --config scripts/.pyscn.toml scripts/
```

## 运行 pyscn

### HTML 报告没有自动打开浏览器。

当 stdin 不是 TTY、通过 SSH 连接或设置了 `CI` 环境变量时，自动打开会被禁用。报告路径会输出到 stderr。使用 `--no-open` 可显式禁用自动打开。

### `pyscn analyze` 在我的仓库上运行很慢。

- 使用 `--skip-clones`（克隆检测是最慢的分析器）
- 缩小分析范围：`pyscn analyze src/`
- 在 `[clones]` 下增大 `min_lines` / `min_nodes`
- 在 `[clones]` 下增大 `max_goroutines`

### 我在合法的 Python 代码上遇到了解析错误。

pyscn 使用 tree-sitter，支持 Python 3.13 及以下版本。请提交包含最小复现代码的 issue。

### pyscn 能自动修复问题吗？

不能。pyscn 只做报告，不会修改源代码。

## 评分与阈值

### 我的健康评分突然下降了。

请检查各类别的分数。常见原因：

- 新增了一个大函数，拉高了平均复杂度。
- 重构后遗留了死代码。
- 复制粘贴产生了克隆代码。
- 新增的导入引入了循环依赖。

对比两次运行的 JSON 输出可以精确定位变化原因。

### 为什么小代码库的耦合评分这么低？

惩罚是基于百分比的，因此 3/10 的问题比例与 300/1000 的问题比例产生相同的惩罚。对于类少于 20 个的项目，建议直接查看原始 CBO 值而非类别评分。

### 多少分算"好"的健康评分？

| 范围 | 含义 |
| --- | --- |
| 90+ | 优秀 |
| 70-90 | 健康代码库的正常水平 |
| 50-70 | 需要改进，但可以挽救 |
| < 50 | 需要集中重构 |

趋势比绝对数值更重要。

## MCP

### AI 助手能看到 pyscn 工具，但调用失败。

确认二进制文件在 PATH 中，或使用 `uvx pyscn-mcp`。直接测试：

```bash
uvx pyscn-mcp
```

详见 [MCP 指南](integrations/mcp.md)。

### MCP 服务器能重构我的代码吗？

不能。pyscn 的 MCP 是只读的。

## 故障排除

### 使用 pip 安装后出现 `pyscn: command not found`。

安装路径不在 PATH 中。使用以下命令检查：

```bash
python -m pip show -f pyscn | grep bin/pyscn
```

在 Linux/macOS 上，将 `~/.local/bin` 添加到 PATH，或使用 `uvx pyscn@latest <command>`（无需安装）、`uv tool install pyscn` 或 `pipx install pyscn`。

### 报告显示分析了 0 个文件。

包含/排除模式过滤掉了所有文件。默认值：`include_patterns = ["**/*.py"]`；排除模式包括 `test_*.py` 和 `*_test.py`。如需分析测试文件，请覆盖排除模式：

```toml
[analysis]
exclude_patterns = [
  "**/__pycache__/*",
  "**/*.pyc",
  ".venv/**",
]
```

### `Warning: parse error in <file>`。

该文件存在 tree-sitter 无法恢复的语法错误。该文件会被跳过，其他文件正常分析。

## 获取帮助

- [GitHub Issues](https://github.com/ludo-technologies/pyscn/issues) -- 报告 bug 和功能请求。
- [GitHub Discussions](https://github.com/ludo-technologies/pyscn/discussions) -- 提问和讨论。
- [源代码](https://github.com/ludo-technologies/pyscn)。
