# MCP 連携

`pyscn-mcp` は、pyscn の分析機能を MCP クライアント（Claude Code, Cursor, ChatGPT desktop など）向けのツールとして公開する Model Context Protocol サーバーです。

## ツール

| ツール | 対応する CLI |
| --- | --- |
| `analyze_code` | `pyscn analyze` |
| `check_complexity` | 複雑度分析器 |
| `detect_clones` | クローン検出器 |
| `check_coupling` | CBO 分析器 |
| `find_dead_code` | デッドコード分析器 |
| `get_health_score` | サマリースコア |

すべてのツールはパス引数とオプションの閾値オーバーライドを受け付けます。結果は構造化された JSON です。

## インストール

| 方法 | コマンド |
| --- | --- |
| uvx（オンデマンド） | — |
| uv tool | `uv tool install pyscn` |
| pipx | `pipx install pyscn` |
| pip | `pip install pyscn` |

## クライアント設定

### Claude Code / Claude Desktop

`~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）または `%APPDATA%/Claude/claude_desktop_config.json`（Windows）を編集します:

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp"]
    }
  }
}
```

アプリを再起動してください。

### Cursor

Settings → Features → Model Context Protocol → Add server:

```json
{
  "pyscn": {
    "command": "uvx",
    "args": ["pyscn-mcp"]
  }
}
```

### バージョン固定

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp==0.2.0"]
    }
  }
}
```

### カスタム設定ファイル

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "uvx",
      "args": ["pyscn-mcp"],
      "env": {
        "PYSCN_CONFIG": "/abs/path/to/.pyscn.toml"
      }
    }
  }
}
```

`PYSCN_CONFIG` を指定しない場合、サーバーは分析対象パスから上位に向かって設定を探索します。

### インストール済みバイナリ

=== "macOS"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "/Users/you/Library/Application Support/uv/tools/pyscn/bin/pyscn-mcp"
        }
      }
    }
    ```

=== "Linux"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "/home/you/.local/share/uv/tools/pyscn/bin/pyscn-mcp"
        }
      }
    }
    ```

=== "Windows"

    ```json
    {
      "mcpServers": {
        "pyscn": {
          "command": "C:/Users/you/AppData/Local/uv/tools/pyscn/bin/pyscn-mcp.exe"
        }
      }
    }
    ```

## プロンプト例

> このプロジェクトに pyscn を実行して、最初に何を修正すべきか教えてください。

## テスト

```bash
uvx pyscn-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | uvx pyscn-mcp
npx @modelcontextprotocol/inspector uvx pyscn-mcp
```

## セキュリティモデル

- 読み取り専用: 静的解析のみ、コード実行なし。
- ディレクトリトラバーサルに対するパスバリデーション。
- 呼び出しごとのタイムアウトとメモリ制限。

サーバーは呼び出しプロセスが読み取り可能なすべてのファイルを読み取ります。

## 制限事項

- インクリメンタルモードなし。各呼び出しでゼロから再分析します。
- 10k ファイル規模のリポジトリでは `detect_clones` に 30 秒以上かかる場合があります。
- 書き込みツールなし。リファクタリングはアシスタント自身のファイル編集ツールを使用します。

## 関連項目

- [CLI リファレンス](../cli/index.md)
- [設定](../configuration/index.md)
