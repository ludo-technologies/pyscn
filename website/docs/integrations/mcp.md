# MCP Integration

`pyscn-mcp` is a Model Context Protocol server exposing pyscn's analyzers as tools for MCP clients (Claude Code, Cursor, ChatGPT desktop, etc.).

## Tools

| Tool | Equivalent CLI |
| --- | --- |
| `analyze_code` | `pyscn analyze` |
| `check_complexity` | Complexity analyzer |
| `detect_clones` | Clone detector |
| `check_coupling` | CBO analyzer |
| `find_dead_code` | Dead code analyzer |
| `get_health_score` | Summary score |

All tools accept path arguments and optional threshold overrides. Results are structured JSON.

## Installation

| Method | Command |
| --- | --- |
| uvx (on-demand) | — |
| uv tool | `uv tool install pyscn` |
| pipx | `pipx install pyscn` |
| pip | `pip install pyscn` |

## Client configuration

### Claude Code / Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%/Claude/claude_desktop_config.json` (Windows):

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

Restart the app.

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

### Pinned version

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

### Custom config file

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

Without `PYSCN_CONFIG`, the server discovers config by walking up from the analyzed path.

### Installed binary

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

## Example prompt

> Run pyscn on this project and tell me what to fix first.

## Testing

```bash
uvx pyscn-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | uvx pyscn-mcp
npx @modelcontextprotocol/inspector uvx pyscn-mcp
```

## Security model

- Read-only: static analysis, no code execution.
- Paths validated against directory traversal.
- Per-invocation timeout and memory limits.

The server reads any file the invoking process can read.

## Limitations

- No incremental mode; each call re-analyzes from scratch.
- `detect_clones` on 10k-file repos can take 30+ seconds.
- No write tools; refactoring uses the assistant's own file-editing tools.

## See also

- [CLI Reference](../cli/index.md)
- [Configuration](../configuration/index.md)
