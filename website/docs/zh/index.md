---
hide:
  - navigation
  - toc
---

<div class="pyscn-hero" markdown="1">

<div class="pyscn-hero__copy" markdown="1">

<p class="pyscn-hero__eyebrow">面向 Python 的结构化静态分析</p>

# pyscn

<p class="pyscn-hero__lede">pyscn 像编译器一样阅读 Python 代码 —— 控制流图、语法树、导入图。这正是它能发现逐行 lint 工具无法发现的问题的原因：<code>return</code> 之后残留的死代码、以不同名字重复的逻辑，以及悄悄形成循环的模块依赖。</p>

```bash
uvx pyscn@latest analyze .
```

[开始使用 :material-arrow-right:](getting-started/quick-start.md){ .md-button .md-button--primary } [在 GitHub 上查看 :fontawesome-brands-github:](https://github.com/ludo-technologies/pyscn){ .md-button }

<p class="pyscn-hero__meta">Go 编写 · 无 Python 运行时依赖 · 100,000+ 行/秒 · 33 条规则</p>

</div>

--8<-- "includes/cfg-diagram.html"

</div>

## 检测内容

<div class="grid cards" markdown>

-   :material-source-branch:{ .lg .middle } __不可达代码__

    ---

    基于 CFG 的可达性分析，可发现 `return` / `raise` / `break` / `continue` 之后的死代码，以及恒真分支之后的不可达代码。

-   :material-content-duplicate:{ .lg .middle } __重复代码__

    ---

    APTED 树编辑距离结合 LSH，支持四种克隆类型：完全相同、重命名、修改、语义相似。

-   :material-gauge:{ .lg .middle } __复杂度__

    ---

    计算每个函数的圈复杂度，阈值可按项目自行调整。

-   :material-shape-outline:{ .lg .middle } __类设计__

    ---

    CBO 耦合度和 LCOM4 内聚度度量，揭示职责过多或过少的类。

-   :material-sync:{ .lg .middle } __循环导入__

    ---

    基于 Tarjan 强连通分量算法，在运行时抛出 `ImportError` 之前发现循环依赖。

-   :material-sitemap:{ .lg .middle } __模块结构__

    ---

    对导入图进行 Leiden 聚类，揭示哪些模块真正应该在一起，哪些不该。

</div>

## 安装

=== "uvx（推荐）"

    ```bash
    uvx pyscn@latest analyze .
    ```

    无需安装即可运行最新版本。

=== "uv"

    ```bash
    uv tool install pyscn
    ```

=== "pipx"

    ```bash
    pipx install pyscn
    ```

=== "pip"

    ```bash
    pip install pyscn
    ```

详见[安装指南](getting-started/installation.md)。

## 快速开始

```bash
pyscn analyze .                                  # 完整分析，生成 HTML 报告
pyscn check --select complexity,deadcode src/    # CI 质量门禁
pyscn init                                       # 生成 .pyscn.toml
```

详见[快速开始](getting-started/quick-start.md)和[规则目录](rules/index.md)。

## AI 代理集成

```bash
uvx add-skills ludo-technologies/pyscn
```

安装 Agent Skills，教 Claude Code、Cursor、Codex、Gemini CLI 及其他编码代理何时以及如何运行每种分析。详见 [Agent Skills](integrations/skills.md)；如需结构化的工具调用，可改用 [MCP 服务器](integrations/mcp.md)。
