---
hide:
  - navigation
  - toc
---

<div class="pyscn-hero" markdown="1">

<div class="pyscn-hero__copy" markdown="1">

<p class="pyscn-hero__eyebrow">Structural static analysis for Python</p>

# pyscn

<p class="pyscn-hero__lede">pyscn reads Python the way a compiler does — control-flow graphs, syntax trees, import graphs — to catch what line-by-line linters miss: dead code stranded after a <code>return</code>, logic duplicated under a different name, and modules wired into silent cycles.</p>

```bash
uvx pyscn@latest analyze .
```

[Get started :material-arrow-right:](getting-started/quick-start.md){ .md-button .md-button--primary } [View on GitHub :fontawesome-brands-github:](https://github.com/ludo-technologies/pyscn){ .md-button }

<p class="pyscn-hero__meta">Go binary · zero Python runtime deps · 100,000+ lines/sec · 33 rules</p>

</div>

--8<-- "includes/cfg-diagram.html"

</div>

## What it finds

<div class="grid cards" markdown="1">

-   :material-source-branch:{ .lg .middle } __Unreachable code__

    ---

    CFG-based reachability finds dead code stranded after `return`, `raise`, `break`, `continue`, or an always-true branch.

-   :material-content-duplicate:{ .lg .middle } __Duplicate code__

    ---

    APTED tree-edit distance plus LSH catches four clone types: identical, renamed, modified, and semantic.

-   :material-gauge:{ .lg .middle } __Complexity__

    ---

    Cyclomatic complexity per function, with thresholds you tune per project.

-   :material-shape-outline:{ .lg .middle } __Class design__

    ---

    CBO coupling and LCOM4 cohesion metrics surface classes doing too much, or too little, together.

-   :material-sync:{ .lg .middle } __Circular imports__

    ---

    Tarjan's SCC algorithm finds import cycles before they turn into an `ImportError` at runtime.

-   :material-sitemap:{ .lg .middle } __Module structure__

    ---

    Leiden clustering over the import graph shows which modules actually belong together — and which don't.

</div>

## Install

=== "uvx (recommended)"

    ```bash
    uvx pyscn@latest analyze .
    ```

    Runs the latest release without installing anything.

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

See [Installation](getting-started/installation.md) for all options.

## Quick start

```bash
pyscn analyze .                                  # full analysis, HTML report
pyscn check --select complexity,deadcode src/    # CI gate
pyscn init                                       # generate .pyscn.toml
```

See [Quick Start](getting-started/quick-start.md) and the [rule catalog](rules/index.md).

## AI agent integration

```bash
uvx add-skills ludo-technologies/pyscn
```

Installs Agent Skills that teach Claude Code, Cursor, Codex, Gemini CLI, and other coding agents when and how to run each analysis. See [Agent Skills](integrations/skills.md), or use the [MCP server](integrations/mcp.md) for structured tool calls instead.
