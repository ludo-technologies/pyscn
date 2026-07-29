# Agent Skills 集成

pyscn 附带 4 个 Agent Skills，用于教编码代理"何时以及如何"运行每种分析，无需搭建 MCP 服务器。

## 安装

```bash
uvx add-skills ludo-technologies/pyscn
```

将 Skills 安装到你的项目中。支持 Claude Code、Cursor、Codex、Gemini CLI 及[其他多种代理](https://github.com/ludo-technologies/add-skills)（使用 `--agent cursor` 等指定单个代理，使用 `--global` 应用到所有项目）。

## Skills 列表

| Skill | 使用场景 |
| --- | --- |
| `health-check` | "这段代码健康吗？"、质量概览、重构前后对比 |
| `refactoring` | 查找重构目标 — 重复代码、复杂度热点、死代码 |
| `architecture-review` | 模块结构、耦合度、循环依赖、应一起审查的文件 |
| `cli-analysis` | CI/CD 质量门禁、可分享的报告、项目配置 |

每个 Skill 内部都会运行 `uvx pyscn@latest <command>`，无需提前安装。

## 示例提示词

> 分析 app/ 目录的代码质量

> 找出重复代码并帮我重构

> 找出复杂的代码并帮我简化

> 合并前检查一下有没有循环依赖

## Claude Code 插件

同时安装 MCP 服务器和 Skills：

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

## Skills 与 MCP 的区别

Agent Skills 教代理"何时使用 pyscn"以及"运行哪条 CLI 命令"，无需服务器，适用于任何支持 Skill 格式的代理。而 [MCP 服务器](mcp.md)将相同的分析暴露为结构化的工具调用——如果你希望客户端直接获得类型化的 JSON 结果而不是 shell 输出，请使用它。两者可以互补，并通过上面的 Claude Code 插件一起安装。

## 另请参阅

- [MCP 集成](mcp.md)
- [CLI 参考](../cli/index.md)
- [add-skills](https://github.com/ludo-technologies/add-skills)
