---
hide:
  - navigation
  - toc
---

# pyscn

Python 结构化静态分析工具。通过控制流和语法树分析检测死代码、重复代码、复杂度和耦合问题。

```bash
uvx pyscn@latest analyze .
```

## 功能

- **33 条规则**，涵盖不可达代码、重复代码、复杂度、类设计、依赖注入、模块结构和 mock 数据。
- **基于 CFG 的可达性分析**，可发现 `return` / `raise` / `break` / `continue` 之后的死代码和不可达分支。
- **APTED + LSH 克隆检测**，支持四种克隆类型（完全相同、重命名、修改、语义）。
- **CBO / LCOM4** 类耦合度和内聚度度量。
- **循环导入检测**，基于 Tarjan 强连通分量算法。
- **健康评分**（0-100），按类别分项评估。
- **CI 就绪**，提供 `pyscn check` 命令、linter 风格输出和确定性退出码。
- **MCP 服务器**（`pyscn-mcp`），支持 Claude Code、Cursor 及其他 MCP 客户端。

使用 Go 编写。在常见硬件上分析速度超过 100,000 行/秒。无 Python 运行时依赖。

## 安装

```bash
uvx pyscn@latest <command>   # 免安装直接运行（推荐）
uv tool install pyscn        # 使用 uv 安装
pipx install pyscn           # 使用 pipx 安装
pip install pyscn            # 使用 pip 安装
```

详见[安装指南](getting-started/installation.md)。

## 快速开始

```bash
pyscn analyze .                         # 完整分析，生成 HTML 报告
pyscn check --select complexity,deadcode src/   # CI 质量门禁
pyscn init                              # 生成 .pyscn.toml
```

详见[快速开始](getting-started/quick-start.md)和[规则目录](rules/index.md)。
