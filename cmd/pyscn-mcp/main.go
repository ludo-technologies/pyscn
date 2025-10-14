package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ludo-technologies/pyscn/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "pyscn"
	serverVersion = "1.0.0"
)

func main() {
	// Set up logging to stderr (MCP uses stdout for JSON-RPC)
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create MCP server with tool capabilities
	server := mcpserver.NewMCPServer(
		serverName,
		serverVersion,
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithLogging(),
	)

	// Register all pyscn tools
	mcp.RegisterTools(server)

	log.Printf("Starting %s MCP server v%s\n", serverName, serverVersion)
	log.Println("Registered tools:")
	log.Println("  - analyze_code: Comprehensive code analysis")
	log.Println("  - check_complexity: Cyclomatic complexity analysis")
	log.Println("  - detect_clones: Code clone detection")
	log.Println("  - check_coupling: Class coupling analysis")
	log.Println("  - find_dead_code: Dead code detection")
	log.Println("  - get_health_score: Code health score")
	log.Println("")
	log.Println("Server ready - waiting for MCP client connection...")

	// Start server with stdio transport
	// This blocks until the server is terminated
	if err := mcpserver.ServeStdio(server); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
