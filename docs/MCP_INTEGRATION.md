# MCP Integration Design for pyscn

## Overview

This document outlines the design for integrating Model Context Protocol (MCP) into pyscn, allowing AI assistants (Claude, Cursor, ChatGPT) to directly access pyscn's code analysis capabilities.

## Architecture

```
┌─────────────────────────────────────────────────┐
│         AI Assistant (Claude/Cursor)            │
│              MCP Client                          │
└─────────────────┬───────────────────────────────┘
                  │ JSON-RPC over stdio/SSE
                  │
┌─────────────────▼───────────────────────────────┐
│         pyscn MCP Server                         │
│  ┌──────────────────────────────────────────┐   │
│  │  MCP Tool Handlers                       │   │
│  │  - analyze_code                          │   │
│  │  - check_complexity                      │   │
│  │  - detect_clones                         │   │
│  │  - check_coupling                        │   │
│  │  - find_dead_code                        │   │
│  │  - get_health_score                      │   │
│  └──────────────┬───────────────────────────┘   │
│                 │                                │
│  ┌──────────────▼───────────────────────────┐   │
│  │  Existing pyscn Core                     │   │
│  │  - app/ (Use Cases)                      │   │
│  │  - domain/ (Domain Models)               │   │
│  │  - internal/analyzer/ (Core Analyzers)   │   │
│  └──────────────────────────────────────────┘   │
└─────────────────────────────────────────────────┘
```

## Recommended SDK

**Choice: `mark3labs/mcp-go` or official `modelcontextprotocol/go-sdk`**

Reasons:
- Most popular community implementation
- Good documentation and examples
- Active maintenance
- Supports stdio, SSE, WebSocket transports

## MCP Tool Definitions

### 1. analyze_code

**Description**: Comprehensive code quality analysis (all metrics)

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code (file or directory)"
    },
    "analyses": {
      "type": "array",
      "items": {
        "type": "string",
        "enum": ["complexity", "dead_code", "clone", "cbo", "deps", "arch"]
      },
      "description": "Analyses to run (default: all)",
      "default": ["complexity", "dead_code", "clone", "cbo"]
    },
    "recursive": {
      "type": "boolean",
      "description": "Recursively analyze directories",
      "default": true
    }
  },
  "required": ["path"]
}
```

**Output**: `AnalyzeResponse` JSON
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
    "average_complexity": 5.2
  }
}
```

### 2. check_complexity

**Description**: Analyze cyclomatic complexity

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code"
    },
    "min_complexity": {
      "type": "integer",
      "description": "Minimum complexity to report",
      "default": 1
    },
    "max_complexity": {
      "type": "integer",
      "description": "Maximum allowed complexity (0 = no limit)",
      "default": 0
    },
    "show_details": {
      "type": "boolean",
      "description": "Include detailed metrics",
      "default": true
    }
  },
  "required": ["path"]
}
```

**Output**: `ComplexityResponse` JSON

### 3. detect_clones

**Description**: Detect code clones using APTED + LSH

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code"
    },
    "similarity_threshold": {
      "type": "number",
      "description": "Minimum similarity (0.0-1.0)",
      "default": 0.8,
      "minimum": 0.0,
      "maximum": 1.0
    },
    "min_lines": {
      "type": "integer",
      "description": "Minimum lines to consider",
      "default": 5
    },
    "group_clones": {
      "type": "boolean",
      "description": "Group related clones",
      "default": true
    }
  },
  "required": ["path"]
}
```

**Output**: `CloneResponse` JSON

### 4. check_coupling

**Description**: Analyze class coupling (CBO metrics)

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code"
    }
  },
  "required": ["path"]
}
```

**Output**: `CBOResponse` JSON

### 5. find_dead_code

**Description**: Find unreachable code using CFG analysis

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code"
    },
    "min_severity": {
      "type": "string",
      "enum": ["info", "warning", "error"],
      "description": "Minimum severity to report",
      "default": "warning"
    }
  },
  "required": ["path"]
}
```

**Output**: `DeadCodeResponse` JSON

### 6. get_health_score

**Description**: Get overall code health score

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "path": {
      "type": "string",
      "description": "Path to Python code"
    }
  },
  "required": ["path"]
}
```

**Output**:
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
    "clone_pairs": 5
  }
}
```

## Implementation Steps

### Phase 1: Basic MCP Server Setup

1. **Add MCP SDK dependency**
   ```bash
   go get github.com/mark3labs/mcp-go
   # or
   go get github.com/modelcontextprotocol/go-sdk
   ```

2. **Create MCP server entry point**
   ```
   cmd/pyscn-mcp/
   ├── main.go           # MCP server entry point
   └── handlers.go       # Tool handlers
   ```

3. **Basic server structure**
   ```go
   package main

   import (
       "context"
       "log"
       mcp "github.com/mark3labs/mcp-go"
   )

   func main() {
       server := mcp.NewServer(
           mcp.WithName("pyscn"),
           mcp.WithVersion("1.0.0"),
           mcp.WithDescription("Python Code Quality Analyzer"),
       )

       // Register tools
       server.RegisterTool(analyzeCodeTool())
       server.RegisterTool(checkComplexityTool())
       server.RegisterTool(detectClonesTool())
       server.RegisterTool(checkCouplingTool())
       server.RegisterTool(findDeadCodeTool())
       server.RegisterTool(getHealthScoreTool())

       // Start server with stdio transport
       if err := server.ServeStdio(context.Background()); err != nil {
           log.Fatal(err)
       }
   }
   ```

### Phase 2: Tool Handlers Implementation

4. **Create adapter layer**
   ```
   app/mcp/
   ├── tools.go          # Tool definitions
   ├── handlers.go       # Handler implementations
   └── adapters.go       # Request/response adapters
   ```

5. **Example handler**
   ```go
   func analyzeCodeHandler(ctx context.Context, args map[string]interface{}) (interface{}, error) {
       // Parse arguments
       path := args["path"].(string)
       analyses := args["analyses"].([]string)

       // Create analyze request
       req := &domain.AnalyzeRequest{
           Paths: []string{path},
           SelectAnalyses: analyses,
           OutputFormat: domain.OutputFormatJSON,
       }

       // Call existing use case
       useCase := app.NewAnalyzeUseCase(/* dependencies */)
       result, err := useCase.Execute(ctx, req)
       if err != nil {
           return nil, err
       }

       // Return JSON response
       return result, nil
   }
   ```

### Phase 3: Transport Configuration

6. **Support multiple transports**
   - stdio (for CLI integration)
   - SSE (for web-based clients)
   - WebSocket (for persistent connections)

7. **Configuration file**
   ```toml
   # .pyscn-mcp.toml
   [server]
   name = "pyscn"
   version = "1.0.0"

   [transport]
   type = "stdio"  # or "sse", "websocket"

   [sse]
   host = "localhost"
   port = 8080

   [websocket]
   host = "localhost"
   port = 8081
   ```

### Phase 4: Client Configuration

8. **Claude Desktop configuration**
   ```json
   {
     "mcpServers": {
       "pyscn": {
         "command": "pyscn-mcp",
         "args": [],
         "env": {}
       }
     }
   }
   ```

9. **Cursor configuration**
   ```json
   {
     "experimental": {
       "modelContextProtocolServers": [
         {
           "transport": {
             "type": "stdio",
             "command": "pyscn-mcp",
             "args": []
           }
         }
       ]
     }
   }
   ```

## Benefits of MCP Integration

### For AI Assistants
- **Direct code analysis**: AI can analyze code quality on-demand
- **Context-aware suggestions**: Refactoring suggestions based on actual metrics
- **Automated quality checks**: Run checks during code generation

### For Developers
- **Seamless integration**: Works with existing AI tools
- **Real-time feedback**: Instant quality metrics while coding
- **Workflow automation**: AI can suggest improvements based on analysis

### Example Use Cases

1. **AI Code Review**
   ```
   User: "Review this function for complexity"
   AI: [Uses check_complexity tool]
   AI: "This function has complexity 15, which is HIGH risk.
        Consider breaking it down into smaller functions."
   ```

2. **Refactoring Suggestions**
   ```
   User: "Find duplicate code in my project"
   AI: [Uses detect_clones tool]
   AI: "Found 5 clone pairs. Here are refactoring suggestions..."
   ```

3. **Quality Gates**
   ```
   User: "Is my code ready for review?"
   AI: [Uses get_health_score tool]
   AI: "Health Score: 85 (Grade A). Your code is in excellent shape!"
   ```

## Testing Strategy

### Unit Tests
- Test each tool handler individually
- Mock use case dependencies
- Validate JSON schema compliance

### Integration Tests
- Test MCP server with real client
- Validate tool registration
- Test error handling

### E2E Tests
- Test with actual Python codebases
- Validate all analysis types
- Performance benchmarks

## Performance Considerations

1. **Caching**: Cache analysis results for repeated paths
2. **Streaming**: Stream results for large codebases
3. **Timeout**: Set reasonable timeouts for long analyses
4. **Resource limits**: Memory/CPU limits per request

## Security Considerations

1. **Path validation**: Prevent directory traversal
2. **Resource limits**: Prevent DoS attacks
3. **Sandboxing**: Run analysis in isolated environment
4. **Authentication**: Optional API key for remote access

## Future Enhancements

1. **Resources**: Expose configuration files as MCP resources
2. **Prompts**: Pre-built prompts for common analysis tasks
3. **Sampling**: Support for progressive analysis
4. **Notifications**: Real-time progress updates
5. **Incremental analysis**: Only analyze changed files

## References

- [MCP Official Specification](https://modelcontextprotocol.io)
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go)
- [Official Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Servers Registry](https://github.com/modelcontextprotocol/servers)
