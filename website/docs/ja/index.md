---
hide:
  - navigation
  - toc
---

# pyscn

Python 向けの構造的静的解析ツールです。制御フローグラフとツリー解析を用いて、デッドコード・コード重複・複雑度・結合度の問題を検出します。

```bash
uvx pyscn@latest analyze .
```

## 機能

- **33 のルール** — 到達不能コード、重複コード、複雑度、クラス設計、依存性注入、モジュール構造、モックデータにまたがるルールセット。
- **CFG ベースの到達可能性解析** — `return` / `raise` / `break` / `continue` の後に残されたデッドコードや到達不能な分岐を検出します。
- **APTED + LSH クローン検出** — 4 種類のクローン（完全一致、名前変更、構造変更、意味的類似）に対応します。
- **CBO / LCOM4** — クラスの結合度と凝集度のメトリクスを計算します。
- **循環インポート検出** — Tarjan の SCC アルゴリズムで循環依存を発見します。
- **ヘルススコア**（0〜100）— カテゴリごとの内訳付き。
- **CI 対応** — `pyscn check` によるリンター形式の出力と確定的な終了コード。
- **MCP サーバー**（`pyscn-mcp`）— Claude Code、Cursor、その他の MCP クライアントから利用できます。

Go で実装されています。一般的なハードウェアで 100,000 行/秒以上の解析速度です。Python ランタイムへの依存はありません。

## インストール

```bash
uvx pyscn@latest <command>   # run without installing (recommended)
uv tool install pyscn        # install with uv
pipx install pyscn           # install with pipx
pip install pyscn            # install with pip
```

詳しくは [Installation](getting-started/installation.md) をご覧ください。

## クイックスタート

```bash
pyscn analyze .                         # full analysis, HTML report
pyscn check --select complexity,deadcode src/   # CI gate
pyscn init                              # generate .pyscn.toml
```

詳しくは [Quick Start](getting-started/quick-start.md) と [Rule catalog](rules/index.md) をご覧ください。
