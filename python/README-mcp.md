# pyscn-mcp

MCP (Model Context Protocol) server for pyscn Python code analyzer.

<!-- The MCP Registry reads this marker off the published PyPI description to
     verify that we own the pyscn-mcp package. It must stay identical to the
     "name" field in server.json, or publishing to the registry fails. -->

mcp-name: io.github.ludo-technologies/pyscn

## Installation

```bash
# Install via pipx (recommended)
pipx install pyscn-mcp

# Or with uv
uvx pyscn-mcp

# Or with pip
pip install pyscn-mcp
```

## Usage

### As MCP Server

Configure in your MCP client (e.g., Claude Desktop):

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

### Standalone

```bash
pyscn-mcp
```

## Documentation

For full documentation, visit the [pyscn repository](https://github.com/ludo-technologies/pyscn).

## License

MIT License - see the [LICENSE](../../LICENSE) file for details.
