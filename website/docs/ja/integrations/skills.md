# Agent Skills 連携

pyscn は 4 つの Agent Skills を同梱しています。コーディングエージェントに「いつ・どの分析を実行すべきか」を教えるもので、MCP サーバーのセットアップは不要です。

## インストール

```bash
uvx add-skills ludo-technologies/pyscn
```

プロジェクトに Skills をインストールします。Claude Code, Cursor, Codex, Gemini CLI、その他[多数のエージェント](https://github.com/ludo-technologies/add-skills)に対応しています（特定のエージェントだけに導入する場合は `--agent cursor` のように指定、全プロジェクトに導入する場合は `--global` を指定）。

## Skills 一覧

| Skill | 使うタイミング |
| --- | --- |
| `health-check` | 「このコードは健全か?」品質の概要確認、リファクタリング前後の比較 |
| `refactoring` | リファクタリング対象の検出 — 重複コード、複雑度のホットスポット、デッドコード |
| `architecture-review` | モジュール構造、結合度、循環依存、一緒にレビューすべきファイルの特定 |
| `cli-analysis` | CI/CD の品質ゲート、共有可能なレポート、プロジェクト設定 |

各 Skill は内部で `uvx pyscn@latest <command>` を実行します。事前のインストールは不要です。

## プロンプト例

> app/ ディレクトリのコード品質を分析して

> 重複コードを見つけてリファクタリングを手伝って

> 複雑なコードを教えて、シンプルにするのを手伝って

> マージする前に循環依存がないか確認して

## Claude Code プラグイン

MCP サーバーと Skills をまとめてセットアップします:

```bash
claude plugin marketplace add ludo-technologies/pyscn
claude plugin install pyscn-mcp@pyscn-marketplace
```

## Skills と MCP の違い

Agent Skills は、エージェントに「いつ pyscn を使うか」「どの CLI コマンドを実行するか」を教えるものです。サーバー不要で、Skill 形式に対応した任意のエージェントで動作します。一方 [MCP サーバー](mcp.md)は同じ分析を構造化されたツール呼び出しとして公開します。シェル出力ではなく型付きの JSON 結果をクライアントに直接組み込みたい場合はこちらを使ってください。両者は補完関係にあり、上記の Claude Code プラグインでまとめて導入できます。

## 関連項目

- [MCP 連携](mcp.md)
- [CLI リファレンス](../cli/index.md)
- [add-skills](https://github.com/ludo-technologies/add-skills)
