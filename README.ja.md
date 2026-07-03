<div align="center">

[English](README.md) | [日本語](README.ja.md) | [简体中文](README.zh-CN.md) | [Français](README.fr.md)

<br>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="assets/logo.svg">
  <source media="(prefers-color-scheme: light)" srcset="assets/logo-light.svg">
  <img alt="pyscn" src="assets/logo-light.svg" width="320">
</picture>

**Python のバイブコーダー向けコード品質アナライザー。**

Cursor、Claude、ChatGPT で開発していますか？pyscn は構造解析により、コードベースの保守性を保ちます。

[![Article](https://img.shields.io/badge/dev.to-Article-0A0A0A?style=flat-square&logo=dev.to)](https://dev.to/daisukeyoda/pyscn-the-code-quality-analyzer-for-vibe-coders-18hk)
[![PyPI](https://img.shields.io/pypi/v/pyscn?style=flat-square&logo=pypi)](https://pypi.org/project/pyscn/)
[![Downloads](https://img.shields.io/pypi/dm/pyscn?style=flat-square&logo=pypi&label=downloads)](https://pypi.org/project/pyscn/)
[![Go](https://img.shields.io/github/go-mod/go-version/ludo-technologies/pyscn?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/ludo-technologies/pyscn?style=flat-square)](LICENSE)

*JavaScript/TypeScript を扱っていますか？[jscan](https://github.com/ludo-technologies/jscan) をご覧ください*

</div>

## クイックスタート

```bash
# インストール不要で解析を実行
uvx pyscn@latest analyze .
# または
pipx run pyscn analyze .
```

## デモ

https://github.com/user-attachments/assets/71d7a126-9c5e-4254-99f4-f2cdedd526ad

## 機能

- 🔍 **CFG ベースのデッドコード検出** – 網羅的な if-elif-else チェーンの後に残された到達不能コードを検出
- 📋 **マルチアルゴリズムのクローン検出（Type 1-4）** – LSH による高速化でリファクタリング候補を特定
- 🔗 **結合度メトリクス（CBO）** – アーキテクチャ品質とモジュール依存関係を追跡
- 📊 **サイクロマティック複雑度解析** – 分割が必要な関数を発見

**100,000 行/秒以上** • Go + tree-sitter で構築

## MCP 連携

Model Context Protocol（MCP）を通じて、AI コーディングアシスタントから直接 pyscn の解析を実行できます。同梱の `pyscn-mcp` サーバーは、CLI と同じツールを Claude Code、Cursor、ChatGPT などの MCP クライアントに公開します。

### MCP の活用例

AI コーディングツールから pyscn を操作できます：

1. 「app/ ディレクトリのコード品質を解析して」

2. 「重複コードを見つけてリファクタリングを手伝って」

3. 「複雑なコードを見せて、シンプルにするのを手伝って」

### Claude Code のセットアップ

**オプション 1: プラグインマーケットプレイスからインストール（推奨）**

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

プラグインは MCP サーバーに加えて、各分析の使いどころを Claude に教える Agent Skills（ヘルスチェック、リファクタリング、アーキテクチャレビュー、CI・レポート向け CLI 利用）もセットアップします。

**オプション 2: 手動 MCP セットアップ**

```bash
claude mcp add pyscn-mcp uvx -- pyscn-mcp
```

### Cursor / Claude Desktop のセットアップ

MCP 設定（`~/.config/claude-desktop/config.json` または Cursor の設定）に追加：

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

「コード品質を解析して」といった指示で MCP 経由の pyscn が起動します。

セットアップ手順の詳細は `mcp/README.md`、アーキテクチャの詳細は `docs/MCP_INTEGRATION.md` をご覧ください。

## インストール

```bash
# pipx でインストール（推奨）
pipx install pyscn

# または uv でインストール
uv tool install pyscn
```

<details>
<summary>その他のインストール方法</summary>

### ソースからビルド
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

## よく使うコマンド

### `pyscn analyze`
HTML レポート付きの包括的な解析を実行
```bash
pyscn analyze .                              # 全解析 + HTML レポート
pyscn analyze --json .                       # JSON レポートを生成
pyscn analyze --select complexity .          # 複雑度解析のみ
pyscn analyze --select deps .                # 依存関係解析のみ
pyscn analyze --select complexity,deps,deadcode . # 複数の解析
```

### `pyscn check`
CI 向けの高速品質ゲート
```bash
pyscn check .                         # クイックな合否チェック
pyscn check --max-complexity 15 .     # カスタム閾値
pyscn check --max-cycles 0 .          # 循環依存 0 のみ許可
pyscn check --select deps .           # 循環依存のみチェック
pyscn check --allow-circular-deps .   # 循環依存を許可（警告のみ）
```

### `pyscn init`
設定ファイルを作成
```bash
pyscn init                         # .pyscn.toml を生成
```

> 💡 すべてのオプションは `pyscn --help` または `pyscn <command> --help` で確認できます

## 設定

`.pyscn.toml` を作成するか、`pyproject.toml` に `[tool.pyscn]` を追加します：

```toml
# .pyscn.toml
[complexity]
max_complexity = 15

[dead_code]
min_severity = "warning"

[output]
directory = "reports"
```

> ⚙️ `pyscn init` を実行すると、利用可能なすべてのオプションを含む完全な設定ファイルが生成されます

## Pyscn Bot（GitHub App）

[Pyscn Bot](https://github.com/marketplace/pyscn-bot) は Python コード品質を自動的に監視します。

### 機能

- **PR コードレビュー** - すべてのプルリクエストで自動コードレビュー
- **週次コード監査** - リポジトリ全体をスキャンし、アーキテクチャ上の問題を Issue として作成

---

## ドキュメント

📖 **[pyscn ドキュメントサイト](https://ludo-technologies.github.io/pyscn/ja/)** — インストール、ルールカタログ、CLI リファレンス、設定、出力仕様

コントリビューター向け: **[開発ガイド](docs/DEVELOPMENT.md)** • **[アーキテクチャ](docs/ARCHITECTURE.md)** • **[テスト](docs/TESTING.md)**

## エンタープライズサポート

商用サポート、カスタム連携、コンサルティングについては contact@ludo-tech.org までお問い合わせください

## ライセンス

MIT License — 詳細は [LICENSE](LICENSE) をご覧ください

---

*Go と tree-sitter で ❤️ を込めて構築*
