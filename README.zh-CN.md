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

- 🔍 **基于 CFG 的死代码检测** – 发现穷尽 if-elif-else 链之后的不可达代码
- 📋 **多算法克隆检测（Type 1-4）** – 借助 LSH 加速识别重构机会
- 🔗 **耦合度指标（CBO）** – 追踪架构质量和模块依赖
- 📊 **圈复杂度分析** – 发现需要拆分的函数

**100,000+ 行/秒** • 基于 Go + tree-sitter 构建

## MCP 集成

通过 Model Context Protocol（MCP）直接从 AI 编程助手运行 pyscn 分析。内置的 `pyscn-mcp` 服务器将 CLI 使用的相同工具暴露给 Claude Code、Cursor、ChatGPT 及其他 MCP 客户端。

### MCP 使用场景

你可以通过 AI 编程工具与 pyscn 交互：

1. "分析 app/ 目录的代码质量"

2. "找出重复代码并帮我重构"

3. "展示复杂代码并帮我简化"

### Claude Code 配置

**选项 1：通过插件市场安装（推荐）**

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

插件除了配置 MCP 服务器外，还会添加 Agent Skills，教 Claude 何时使用各项分析：健康检查、重构、架构评审，以及面向 CI 和报告的 CLI 用法。

**选项 2：手动 MCP 配置**

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

### Cursor / Claude Desktop 配置

添加到 MCP 设置（`~/.config/claude-desktop/config.json` 或 Cursor 设置）：

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

"分析代码质量"等指令会通过 MCP 触发 pyscn。

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
