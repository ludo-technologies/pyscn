# pyscn MCP Server

Model Context Protocol (MCP) integration for pyscn - Python Code Quality Analyzer.

## Overview

pyscn MCP Server exposes pyscn's code analysis capabilities as MCP tools, allowing AI assistants like Claude, Cursor, and ChatGPT to directly analyze Python code quality.

## Features

### Available Tools

| Tool | Description | Key Metrics |
|------|-------------|-------------|
| `analyze_code` | Comprehensive code analysis | All metrics combined |
| `check_complexity` | Cyclomatic complexity analysis | McCabe complexity, nesting depth |
| `detect_clones` | Code clone detection | APTED + LSH, Type 1-4 clones |
| `check_coupling` | Class coupling analysis | CBO (Coupling Between Objects) |
| `find_dead_code` | Dead code detection | CFG-based unreachable code |
| `get_health_score` | Overall code health | Score (0-100), Grade (A-F) |

## Quick Start

### 1. Build the MCP Server

#### Option A: Build Go-based MCP Server

```bash
# Build the MCP server binary
make build-mcp

# Or install globally
make install-mcp
```

This creates the `pyscn-mcp` binary.

#### Option B: Use FastMCP Python Server (Recommended)

```bash
# Install pyscn with MCP support
pip install "pyscn[mcp]"
# or
pip install pyscn fastmcp

# Start MCP server
pyscn --start-mcp
```

This uses the FastMCP-based Python server, which is easier to install and maintain.

### 2. Configure Your AI Assistant

#### Cherry Studio

Add to Cherry Studio MCP settings:

```json
{
  "mcpServers": {
    "pyscn": {
      "isActive": true,
      "disabled": false,
      "command": "pyscn",
      "args": ["--start-mcp"],
      "transportType": "stdio",
      "timeout": 120,
      "name": "pyscn"
    }
  }
}
```

#### Claude Desktop

Edit configuration file:
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "pyscn": {
      "command": "pyscn",
      "args": ["--start-mcp"]
    }
  }
}
```

#### Cursor

Add to Cursor settings (Settings → Features → Model Context Protocol):

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "name": "pyscn",
        "transport": {
          "type": "stdio",
          "command": "pyscn",
          "args": ["--start-mcp"]
        }
      }
    ]
  }
}
```

**Note**: Replace `"pyscn"` with `"pyscn-mcp"` if using the Go-based server.

See [../MCP_FASTMCP_GUIDE.md](../MCP_FASTMCP_GUIDE.md) for FastMCP-specific configuration.

### 3. Restart Your AI Assistant

Restart Claude Desktop or Cursor to load the MCP server.

### 4. Test It Out

Try asking your AI assistant:

- "Analyze the code quality of /path/to/my/project"
- "Check the complexity of functions in main.py"
- "Find duplicate code in my project"
- "What's the health score of my codebase?"

## Tool Usage

### analyze_code

**Description**: Comprehensive Python code quality analysis

**Parameters**:
- `path` (required): Path to Python code (file or directory)
- `analyses` (optional): Array of analyses to run
  - Options: `["complexity", "dead_code", "clone", "cbo", "deps"]`
  - Default: All analyses
- `recursive` (optional): Recursively analyze directories (default: `true`)

**Example**:
```
Analyze the code at /home/user/project with all metrics
```

**Output**: JSON with comprehensive analysis results
```json
{
  "complexity": { ... },
  "dead_code": { ... },
  "clone": { ... },
  "cbo": { ... },
  "summary": {
    "health_score": 85,
    "grade": "A",
    "total_files": 42,
    "average_complexity": 5.2,
    "dead_code_count": 2,
    "clone_pairs": 5
  }
}
```

### check_complexity

**Description**: Analyze cyclomatic complexity of Python functions

**Parameters**:
- `path` (required): Path to Python code
- `min_complexity` (optional): Minimum complexity to report (default: `1`)
- `max_complexity` (optional): Maximum allowed complexity, 0 = no limit (default: `0`)
- `show_details` (optional): Include detailed metrics (default: `true`)

**Example**:
```
Check complexity of functions with complexity > 10 in src/
```

**Output**: Complexity analysis with risk levels
```json
{
  "functions": [
    {
      "name": "complex_function",
      "file_path": "src/main.py",
      "start_line": 42,
      "metrics": {
        "complexity": 15,
        "nesting_depth": 4
      },
      "risk_level": "high"
    }
  ],
  "summary": {
    "average_complexity": 8.5,
    "high_risk_functions": 3
  }
}
```

### detect_clones

**Description**: Detect code clones using APTED tree edit distance and LSH

**Parameters**:
- `path` (required): Path to Python code
- `similarity_threshold` (optional): Minimum similarity 0.0-1.0 (default: `0.8`)
- `min_lines` (optional): Minimum lines to consider (default: `5`)
- `group_clones` (optional): Group related clones (default: `true`)

**Example**:
```
Find duplicate code with similarity > 0.85 in my project
```

**Output**: Clone pairs and groups
```json
{
  "clone_pairs": [
    {
      "clone1": {
        "file_path": "src/a.py",
        "start_line": 10,
        "end_line": 25
      },
      "clone2": {
        "file_path": "src/b.py",
        "start_line": 42,
        "end_line": 57
      },
      "similarity": 0.92,
      "type": "Type-2"
    }
  ],
  "statistics": {
    "total_clone_pairs": 5,
    "average_similarity": 0.87
  }
}
```

### check_coupling

**Description**: Analyze class coupling (CBO - Coupling Between Objects)

**Parameters**:
- `path` (required): Path to Python code

**Example**:
```
Check the coupling of classes in src/
```

**Output**: CBO metrics per class
```json
{
  "classes": [
    {
      "name": "MyClass",
      "file_path": "src/service.py",
      "cbo": 8,
      "risk_level": "high",
      "coupled_classes": ["ClassA", "ClassB", ...]
    }
  ],
  "summary": {
    "average_coupling": 4.2,
    "high_coupling_classes": 3
  }
}
```

### find_dead_code

**Description**: Find unreachable code using CFG analysis

**Parameters**:
- `path` (required): Path to Python code
- `min_severity` (optional): Minimum severity: `info`, `warning`, `error` (default: `warning`)

**Example**:
```
Find dead code with severity >= warning in my project
```

**Output**: List of dead code locations
```json
{
  "dead_code": [
    {
      "file_path": "src/util.py",
      "function": "process_data",
      "line": 42,
      "severity": "warning",
      "reason": "Unreachable after exhaustive if-elif-else"
    }
  ],
  "summary": {
    "dead_code_count": 5,
    "critical_dead_code": 2
  }
}
```

### get_health_score

**Description**: Get overall code health score (0-100) with grade

**Parameters**:
- `path` (required): Path to Python code

**Example**:
```
What's the health score of my codebase?
```

**Output**: Health score and breakdown
```json
{
  "health_score": 85,
  "grade": "A",
  "is_healthy": true,
  "category_scores": {
    "complexity_score": 90,
    "dead_code_score": 95,
    "duplication_score": 80,
    "coupling_score": 85,
    "dependency_score": 88,
    "architecture_score": 82
  },
  "summary": {
    "total_files": 42,
    "average_complexity": 5.2,
    "dead_code_count": 2,
    "clone_pairs": 5,
    "high_coupling_classes": 3
  }
}
```

## Use Cases

### 1. AI Code Review

**Scenario**: Get AI-powered code review with actual metrics

**Example Prompts**:
- "Review this function for complexity and suggest improvements"
- "What are the quality issues in this module?"
- "Is this code maintainable?"

**Benefits**:
- Objective metrics instead of subjective opinions
- Specific refactoring suggestions
- Quantifiable improvement targets

### 2. Refactoring Assistant

**Scenario**: Find refactoring opportunities

**Example Prompts**:
- "Find duplicate code that can be refactored"
- "Which functions are too complex?"
- "What classes have high coupling?"

**Benefits**:
- Data-driven refactoring decisions
- Prioritized by impact
- Clear before/after metrics

### 3. Quality Gate

**Scenario**: Check if code is ready for review/deployment

**Example Prompts**:
- "Is this code ready for production?"
- "Check if the code meets our quality standards"
- "What's the overall quality of this PR?"

**Benefits**:
- Automated quality checks
- Consistent standards
- Fast feedback loop

### 4. Code Understanding

**Scenario**: Understand unfamiliar code

**Example Prompts**:
- "Explain this complex function"
- "What are the dependencies in this module?"
- "Which parts of this code are tightly coupled?"

**Benefits**:
- Context-aware explanations
- Visual dependency maps
- Complexity hotspots

## Configuration

### Global Configuration

Create `.pyscn-mcp.toml` in your project root:

```toml
[defaults]
recursive = true
min_complexity = 10
similarity_threshold = 0.85

[performance]
timeout_seconds = 300
```

### Environment Variables

Set in your MCP client configuration:

- `PYSCN_LOG_LEVEL`: Log level (`debug`, `info`, `warn`, `error`)
- `PYSCN_CONFIG`: Path to custom config file

**Example** (Claude Desktop):
```json
{
  "mcpServers": {
    "pyscn": {
      "command": "pyscn-mcp",
      "env": {
        "PYSCN_LOG_LEVEL": "debug",
        "PYSCN_CONFIG": "/path/to/.pyscn-mcp.toml"
      }
    }
  }
}
```

## Development

### Building

```bash
# Build for current platform
make build-mcp

# Build for all platforms
make build-mcp-all

# Install globally
make install-mcp
```

### Testing

#### Manual Testing

```bash
# Run the server directly
./pyscn-mcp
```

Then interact with it via stdin/stdout (JSON-RPC format).

#### MCP Inspector

Use the official MCP Inspector:

```bash
npx @modelcontextprotocol/inspector pyscn-mcp
```

This provides a web UI for:
- Viewing available tools
- Testing tool calls
- Inspecting request/response payloads
- Debugging issues

### Adding New Tools

1. **Define the tool** in `mcp/tools.go`:
```go
s.AddTool(mcp.NewTool("my_new_tool",
    mcp.WithDescription("Tool description"),
    mcp.WithString("path", mcp.Required(), mcp.Description("Path parameter")),
), HandleMyNewTool)
```

2. **Implement the handler** in `mcp/handlers.go`:
```go
func HandleMyNewTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Parse arguments
    // Call existing use cases
    // Return JSON result
}
```

3. **Rebuild**:
```bash
make build-mcp
```

## Troubleshooting

### Server Not Found

**Issue**: `command not found: pyscn-mcp`

**Solutions**:
1. Build the binary: `make build-mcp`
2. Add to PATH or use absolute path in config
3. Verify with: `which pyscn-mcp` (Unix) or `where pyscn-mcp` (Windows)

### Permission Denied

**Issue**: Permission denied when running `pyscn-mcp`

**Solution** (Unix):
```bash
chmod +x pyscn-mcp
```

### Tool Not Appearing

**Issue**: Tool doesn't show up in AI assistant

**Solutions**:
1. Restart the AI assistant completely
2. Check server logs: `./pyscn-mcp 2> mcp.log`
3. Verify config file syntax
4. Test with MCP Inspector

### Analysis Fails

**Issue**: Tool returns error or times out

**Solutions**:
1. Check path exists and is accessible
2. Verify Python code is valid
3. Increase timeout in config
4. Check logs for specific errors

### Slow Performance

**Issue**: Analysis takes too long

**Solutions**:
1. Analyze specific files instead of entire project
2. Disable heavy analyses (clone detection)
3. Increase timeout
4. Use filters (min_complexity, similarity_threshold)

## Architecture

```
AI Assistant (Claude/Cursor)
    ↓ JSON-RPC via stdio
pyscn MCP Server
    ↓ Function calls
Tool Handlers (mcp/handlers.go)
    ↓ Domain requests
Use Cases (app/)
    ↓ Business logic
Analyzers (internal/analyzer/)
    ↓ Tree-sitter parsing
Python Code
```

## Performance

- **Complexity analysis**: ~100,000+ lines/sec
- **Clone detection**: Depends on code size
  - LSH acceleration for large codebases
  - ~1000 fragments/sec without LSH
  - ~10,000+ fragments/sec with LSH
- **CBO analysis**: ~50,000+ lines/sec
- **Dead code**: ~80,000+ lines/sec

## Security

- **Path validation**: Prevents directory traversal
- **Resource limits**: Timeout and memory constraints
- **Sandboxing**: Analysis runs in isolated context
- **No code execution**: Static analysis only

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

## License

MIT License - see [LICENSE](../LICENSE)

## Support

- **Issues**: https://github.com/ludo-technologies/pyscn/issues
- **Discussions**: https://github.com/ludo-technologies/pyscn/discussions
- **Documentation**: https://github.com/ludo-technologies/pyscn/tree/main/docs
