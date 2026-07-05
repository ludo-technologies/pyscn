<div align="center">

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [Français](README.fr.md)

<br>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="assets/logo-light.svg">
  <img alt="pyscn" src="assets/logo-light.svg" width="320">
</picture>

**面向 Python 氛围编程者的代码质量分析工具。**

使用 Cursor、Claude 或 ChatGPT 开发？pyscn 通过结构化分析帮助保持代码库的可维护性。

[![Article](https://img.shields.io/badge/dev.to-Article-0A0A0A?style=flat-square&logo=dev.to)](https://dev.to/daisukeyoda/pyscn-the-code-quality-analyzer-for-vibe-coders-18hk)
[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Downloads](https://img.shields.io/pypi/dm/pyscn?style=flat-square&logo=pypi&label=downloads)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/github/go-mod/go-version/ludo-technologies/pyscn?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/ludo-technologies/pyscn?style=flat-square)](LICENSE)

*使用 JavaScript/TypeScript？请查看 [jscan](https://github.com/ludo-technologies/jscan)*

</div>

## 快速开始

```bash
# 无需安装即可运行分析
uvx pyscn@latest analyze .
# 或
pipx run pyscn analyze .
```

## 演示

https://github.com/user-attachments/assets/71d7a126-9c5e-4254-99f4-f2cdedd526ad

## 功能

一条命令即可为整个代码库打分（0-100 分 + A-F 等级），并生成 HTML 报告，告诉你应该先修什么。

pyscn 从五个角度分析你的代码：

- 🧹 **死代码** - 找出可以安全删除的不可达代码
- 📋 **重复代码** - 检测复制粘贴和结构相似的代码，提示合并候选（Type 1-4 克隆检测）
- 🌀 **复杂度** - 定位难读、难测试的函数（圈复杂度 + 认知复杂度）
- 🏗️ **架构** - 发现循环导入、分层规则违规（clean / layered / hexagonal / MVC 预设），并自动检测模块社区，展示代码的实际结构
- 🧩 **类设计** - 发现职责过多、依赖过多的类（耦合度 CBO、内聚度 LCOM4、DI 反模式）

**100,000+ 行/秒** • 基于 Go + tree-sitter 构建

## AI 智能体集成

pyscn 内置 Agent Skills，教 AI 编程智能体何时以及如何运行各项分析：健康检查、重构、架构评审，以及面向 CI 的报告。

### Agent Skills（推荐）

```bash
uvx add-skills ludo-technologies/pyscn
```

该命令会将 Skills 安装到你的项目中。支持 Claude Code、Cursor、Codex、Gemini CLI 等[众多智能体](https://github.com/ludo-technologies/add-skills)（用 `--agent cursor` 等指定目标，用 `--global` 安装到所有项目）。

然后直接向你的智能体提问：

1. "分析 app/ 目录的代码质量"

2. "找出重复代码并帮我重构"

3. "展示复杂代码并帮我简化"

### MCP 服务器（可选）

如需更深度的集成，内置的 `pyscn-mcp` 服务器会将相同的分析以 MCP 工具的形式暴露给 Claude Code、Cursor、ChatGPT 及其他 MCP 客户端。

**Claude Code 插件（同时配置 MCP 服务器和 Skills）：**

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

**Claude Code 手动配置：**

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

**Cursor / Claude Desktop：** 添加到 MCP 设置（`~/.config/claude-desktop/config.json` 或 Cursor 设置）：

```json
{
  "mcpServers": {
    "pyscn-mcp": {
      "command": "uvx",
      "args": ["pyscn-mcp"],
      "env": {
        "PYSCN_CONFIG": "/path/to/.pyscn.toml"
      }
    }
  }
}
```

详细配置步骤请参阅 `mcp/README.md`，架构详情请参阅 `docs/MCP_INTEGRATION.md`。

## 安装

```bash
# 使用 pipx 安装（推荐）
pipx install pyscn

# 或使用 uv 安装
uv tool install pyscn
```

<details>
<summary>其他安装方式</summary>

### 从源码构建
```bash
git clone https://github.com/ludo-technologies/pyscn.git
cd pyscn
make build
```

### Go install
```bash
go install github.com/ludo-technologies/pyscn/cmd/pyscn@latest
```

</details>

## 常用命令

### `pyscn analyze`
运行全面分析并生成 HTML 报告
```bash
pyscn analyze .                              # 全部分析 + HTML 报告
pyscn analyze --json .                       # 生成 JSON 报告
pyscn analyze --select complexity .          # 仅复杂度分析
pyscn analyze --select deps .                # 仅依赖分析
pyscn analyze --select complexity,deps,deadcode . # 多项分析
pyscn analyze --skip-communities .           # 跳过模块社区检测
```

### `pyscn check`
适合 CI 的快速质量门禁
```bash
pyscn check .                         # 快速通过/失败检查
pyscn check --max-complexity 15 .     # 自定义阈值
pyscn check --max-cycles 0 .          # 仅允许 0 个循环依赖
pyscn check --select deps .           # 仅检查循环依赖
pyscn check --allow-circular-deps .   # 允许循环依赖（仅警告）
```

### `pyscn init`
创建配置文件
```bash
pyscn init                         # 生成 .pyscn.toml
```

> 💡 运行 `pyscn --help` 或 `pyscn <command> --help` 查看完整选项

## 配置

创建 `.pyscn.toml` 文件，或在 `pyproject.toml` 中添加 `[tool.pyscn]`：

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "warning"

[output]
directory = "reports"
```

> ⚙️ 运行 `pyscn init` 可生成包含所有可用选项的完整配置文件

## Pyscn Bot（GitHub App）

[Pyscn Bot](https://github.com/marketplace/pyscn-bot) 自动监控 Python 代码质量。

### 功能

- **PR 代码审查** - 在每个 Pull Request 上自动进行代码审查
- **每周代码审计** - 扫描整个仓库并为架构问题创建 Issue

---

## 文档

📖 **[pyscn 文档站点](https://ludo-technologies.github.io/pyscn/zh/)** — 安装、规则目录、CLI 参考、配置、输出规范

贡献者请参阅：**[开发指南](docs/DEVELOPMENT.md)** • **[架构](docs/ARCHITECTURE.md)** • **[测试](docs/TESTING.md)**

## 企业支持

如需商业支持、定制集成或咨询服务，请联系 contact@ludo-tech.org

## 许可证

MIT License — 详见 [LICENSE](LICENSE)

---

*使用 Go 和 tree-sitter 用 ❤️ 构建*
