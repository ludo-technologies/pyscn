package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all pyscn MCP tools with the server
func RegisterTools(s *server.MCPServer) {
	// Tool 1: analyze_code - Comprehensive code analysis
	s.AddTool(mcp.NewTool("analyze_code",
		mcp.WithDescription("Comprehensive Python code quality analysis with complexity, dead code, clone detection, and coupling metrics"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code (file or directory) to analyze")),
		mcp.WithArray("analyses",
			mcp.WithStringEnumItems([]string{"complexity", "dead_code", "clone", "cbo", "deps"}),
			mcp.Description("Array of analyses to run. Options: complexity, dead_code, clone, cbo, deps. Default: all analyses")),
		mcp.WithBoolean("recursive",
			mcp.Description("Recursively analyze directories (default: true)")),
	), HandleAnalyzeCode)

	// Tool 2: check_complexity - Cyclomatic complexity analysis
	s.AddTool(mcp.NewTool("check_complexity",
		mcp.WithDescription("Analyze cyclomatic complexity of Python functions"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code to analyze")),
		mcp.WithNumber("min_complexity",
			mcp.Description("Minimum complexity to report (default: 1)")),
		mcp.WithNumber("max_complexity",
			mcp.Description("Maximum allowed complexity, 0 = no limit (default: 0)")),
		mcp.WithBoolean("show_details",
			mcp.Description("Include detailed metrics (default: true)")),
	), HandleCheckComplexity)

	// Tool 3: detect_clones - Code clone detection
	s.AddTool(mcp.NewTool("detect_clones",
		mcp.WithDescription("Detect code clones using APTED tree edit distance and LSH acceleration"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code to analyze")),
		mcp.WithNumber("similarity_threshold",
			mcp.Description("Minimum similarity threshold 0.0-1.0 (default: 0.8)")),
		mcp.WithNumber("min_lines",
			mcp.Description("Minimum lines to consider as clone (default: 5)")),
		mcp.WithBoolean("group_clones",
			mcp.Description("Group related clones together (default: true)")),
	), HandleDetectClones)

	// Tool 4: check_coupling - Class coupling analysis
	s.AddTool(mcp.NewTool("check_coupling",
		mcp.WithDescription("Analyze class coupling (CBO - Coupling Between Objects) metrics"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code to analyze")),
	), HandleCheckCoupling)

	// Tool 5: find_dead_code - Dead code detection
	s.AddTool(mcp.NewTool("find_dead_code",
		mcp.WithDescription("Find unreachable code using Control Flow Graph (CFG) analysis"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code to analyze")),
		mcp.WithString("min_severity",
			mcp.Description("Minimum severity: info, warning, error (default: warning)")),
	), HandleFindDeadCode)

	// Tool 6: get_health_score - Overall code health score
	s.AddTool(mcp.NewTool("get_health_score",
		mcp.WithDescription("Get overall code health score (0-100) with grade and category scores"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Path to Python code to analyze")),
	), HandleGetHealthScore)
}
