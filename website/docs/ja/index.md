---
hide:
  - navigation
  - toc
---

<div class="pyscn-hero" markdown="1">

<div class="pyscn-hero__copy" markdown="1">

<p class="pyscn-hero__eyebrow">Python 向けの構造的静的解析</p>

# pyscn

<p class="pyscn-hero__lede">pyscn はコンパイラのように Python を読みます — 制御フローグラフ、構文木、インポートグラフ。行単位のリンターでは見つからないもの、<code>return</code> の後に取り残されたデッドコード、別名で複製されたロジック、静かに循環するモジュール依存を検出します。</p>

```bash
uvx pyscn@latest analyze .
```

[はじめる :material-arrow-right:](getting-started/quick-start.md){ .md-button .md-button--primary } [GitHub で見る :fontawesome-brands-github:](https://github.com/ludo-technologies/pyscn){ .md-button }

<p class="pyscn-hero__meta">Go 製バイナリ · Python ランタイム依存なし · 100,000+ 行/秒 · 33 のルール</p>

</div>

--8<-- "includes/cfg-diagram.html"

</div>

## 検出内容

<div class="grid cards" markdown>

-   :material-source-branch:{ .lg .middle } __到達不能コード__

    ---

    CFG ベースの到達可能性解析で、`return` / `raise` / `break` / `continue` の後や常に真になる分岐の先に残されたデッドコードを検出します。

-   :material-content-duplicate:{ .lg .middle } __重複コード__

    ---

    APTED の木編集距離と LSH の組み合わせで、完全一致・名前変更・構造変更・意味的類似の4種類のクローンを検出します。

-   :material-gauge:{ .lg .middle } __複雑度__

    ---

    関数ごとのサイクロマティック複雑度を計測します。しきい値はプロジェクトごとに調整できます。

-   :material-shape-outline:{ .lg .middle } __クラス設計__

    ---

    CBO 結合度と LCOM4 凝集度のメトリクスで、責務を持ちすぎる、あるいは持たなすぎるクラスを可視化します。

-   :material-sync:{ .lg .middle } __循環インポート__

    ---

    Tarjan の SCC アルゴリズムで、実行時に `ImportError` になる前に循環依存を発見します。

-   :material-sitemap:{ .lg .middle } __モジュール構造__

    ---

    インポートグラフに対する Leiden クラスタリングで、本来ひとまとまりであるべきモジュールとそうでないモジュールを明らかにします。

</div>

## インストール

=== "uvx（推奨）"

    ```bash
    uvx pyscn@latest analyze .
    ```

    インストールせずに最新版を実行します。

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

詳しくは [Installation](getting-started/installation.md) をご覧ください。

## クイックスタート

```bash
pyscn analyze .                                  # full analysis, HTML report
pyscn check --select complexity,deadcode src/    # CI gate
pyscn init                                       # generate .pyscn.toml
```

詳しくは [Quick Start](getting-started/quick-start.md) と [ルールカタログ](rules/index.md) をご覧ください。

## AI エージェント連携

```bash
uvx add-skills ludo-technologies/pyscn
```

Claude Code, Cursor, Codex, Gemini CLI、その他のコーディングエージェントに、各分析の使いどころを教える Agent Skills をインストールします。詳しくは [Agent Skills](integrations/skills.md) をご覧ください。構造化されたツール呼び出しが必要な場合は [MCP サーバー](integrations/mcp.md) を使用してください。
