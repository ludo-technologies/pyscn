# MCP 集成

`pyscn-mcp` 是一个 Model Context Protocol 服务器，将 pyscn 的分析器作为工具暴露给 MCP 客户端（Claude Code、Cursor、ChatGPT 桌面版等）。

## 工具

| 工具 | 等效 CLI |
| --- | --- |
| `analyze_code` | `pyscn analyze` |
| `check_complexity` | 复杂度分析器 |
| `detect_clones` | 克隆检测器 |
| `check_coupling` | CBO 分析器 |
| `find_dead_code` | 死代码分析器 |
| `get_health_score` | 健康评分摘要 |

所有工具接受路径参数和可选的阈值覆盖。结果为结构化 JSON。

## 安装

| 方式 | 命令 |
| --- | --- |
| uvx（按需运行） | — |
| uv tool | `uv tool install pyscn` |
| pipx | `pipx install pyscn` |
| pip | `pip install pyscn` |

## 客户端配置

### Claude Code / Claude Desktop

编辑 `~/Library/Application Support/Claude/claude_desktop_config.json`（macOS）或 `%APPDATA%/Claude/claude_desktop_config.json`（Windows）：

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

重启应用。

### Cursor

Settings → Features → Model Context Protocol → Add server：

```json
{
  "pyscn": {
    "command": "uvx",
    "args": ["pyscn-mcp"]
  }
}
```

### 固定版本

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

### 自定义配置文件

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

未设置 `PYSCN_CONFIG` 时，服务器会从分析路径向上查找配置。

### 已安装的二进制文件

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

## 示例提示词

> 对这个项目运行 pyscn，告诉我应该优先修复什么。

## 测试

```bash
uvx pyscn-mcp
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | uvx pyscn-mcp
npx @modelcontextprotocol/inspector uvx pyscn-mcp
```

## 安全模型

- 只读：静态分析，不执行代码。
- 路径经过目录遍历验证。
- 每次调用有超时和内存限制。

服务器可读取调用进程有权读取的任何文件。

## 限制

- 无增量模式；每次调用都从头开始分析。
- 在 10k 文件的仓库上 `detect_clones` 可能需要 30 秒以上。
- 无写入工具；重构使用助手自身的文件编辑工具。

## 另请参阅

- [CLI 参考](../cli/index.md)
- [配置](../configuration/index.md)
