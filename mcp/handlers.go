package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/mark3labs/mcp-go/mcp"
)

// HandlerSet exposes MCP tool handlers with shared dependencies.
type HandlerSet struct {
	deps *Dependencies
}

// NewHandlerSet constructs a handler set.
func NewHandlerSet(deps *Dependencies) *HandlerSet {
	if deps == nil {
		deps = NewDependencies(nil, "")
	}
	return &HandlerSet{deps: deps}
}

// HandleAnalyzeCode handles the analyze_code tool
func (h *HandlerSet) HandleAnalyzeCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse arguments with type assertion
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	// Validate path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	// Parse analyses array
	analyses := []string{}
	if rawAnalyses, ok := args["analyses"].([]interface{}); ok {
		for _, a := range rawAnalyses {
			if str, ok := a.(string); ok {
				analyses = append(analyses, str)
			}
		}
	}

	// Create config for analyze use case
	config := app.AnalyzeUseCaseConfig{
		SkipComplexity:  !contains(analyses, "complexity") && len(analyses) > 0,
		SkipDeadCode:    !contains(analyses, "dead_code") && len(analyses) > 0,
		SkipClones:      !contains(analyses, "clone") && len(analyses) > 0,
		SkipCBO:         !contains(analyses, "cbo") && len(analyses) > 0,
		SkipSystem:      !contains(analyses, "deps") && len(analyses) > 0,
		MinComplexity:   1,
		MinSeverity:     domain.DeadCodeSeverityWarning,
		CloneSimilarity: 0.8,
		ConfigFile:      h.deps.ConfigPath(),
	}
	if cfg := h.deps.Config(); cfg != nil {
		if cfg.Output.MinComplexity > 0 {
			config.MinComplexity = cfg.Output.MinComplexity
		}
		switch cfg.DeadCode.MinSeverity {
		case "info":
			config.MinSeverity = domain.DeadCodeSeverityInfo
		case "critical", "error":
			config.MinSeverity = domain.DeadCodeSeverityCritical
		default:
			config.MinSeverity = domain.DeadCodeSeverityWarning
		}
		if cfg.Clones != nil && cfg.Clones.Thresholds.SimilarityThreshold > 0 {
			config.CloneSimilarity = cfg.Clones.Thresholds.SimilarityThreshold
		}
	}

	// Build analyze use case using builder pattern
	analyzeUC, err := h.deps.BuildAnalyzeUseCase()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create analyzer: %v", err)), nil
	}

	// Collect files
	paths := []string{path}

	// Execute analysis
	result, err := analyzeUC.Execute(ctx, config, paths)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	// Parse output_mode parameter (default: "summary")
	outputMode := "summary"
	if om, ok := args["output_mode"].(string); ok {
		outputMode = om
	}

	// Format output based on mode
	var responseData interface{}
	switch outputMode {
	case "full":
		responseData = result
	default: // "summary" - return health score and high-level metrics
		responseData = map[string]interface{}{
			"health_score": result.Summary.HealthScore,
			"grade":        result.Summary.Grade,
			"is_healthy":   result.Summary.IsHealthy(),
			"summary": map[string]interface{}{
				"total_files":           result.Summary.TotalFiles,
				"total_functions":       result.Summary.TotalFunctions,
				"complexity_score":      result.Summary.ComplexityScore,
				"dead_code_score":       result.Summary.DeadCodeScore,
				"duplication_score":     result.Summary.DuplicationScore,
				"coupling_score":        result.Summary.CouplingScore,
				"dependency_score":      result.Summary.DependencyScore,
				"high_complexity_count": result.Summary.HighComplexityCount,
				"dead_code_count":       result.Summary.DeadCodeCount,
				"clone_pairs":           result.Summary.ClonePairs,
				"high_coupling_classes": result.Summary.HighCouplingClasses,
			},
		}
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleCheckComplexity handles the check_complexity tool
func (h *HandlerSet) HandleCheckComplexity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	// Parse optional parameters
	cfg := h.deps.Config()

	minComplexity := 1
	if cfg != nil && cfg.Output.MinComplexity > 0 {
		minComplexity = cfg.Output.MinComplexity
	}
	if mc, ok := args["min_complexity"].(float64); ok {
		minComplexity = int(mc)
	}

	maxComplexity := 0
	if cfg != nil && cfg.Complexity.MaxComplexity > 0 {
		maxComplexity = cfg.Complexity.MaxComplexity
	}
	if mc, ok := args["max_complexity"].(float64); ok {
		maxComplexity = int(mc)
	}

	showDetails := true
	if cfg != nil {
		showDetails = cfg.Output.ShowDetails
	}
	if sd, ok := args["show_details"].(bool); ok {
		showDetails = sd
	}

	lowThreshold := 9
	mediumThreshold := 19
	if cfg != nil {
		if cfg.Complexity.LowThreshold > 0 {
			lowThreshold = cfg.Complexity.LowThreshold
		}
		if cfg.Complexity.MediumThreshold > 0 {
			mediumThreshold = cfg.Complexity.MediumThreshold
		}
	}

	sortBy := domain.SortByComplexity
	if cfg != nil {
		switch cfg.Output.SortBy {
		case "name":
			sortBy = domain.SortByName
		case "risk":
			sortBy = domain.SortByRisk
		}
	}

	includePatterns := []string{}
	excludePatterns := []string{}
	if cfg != nil {
		if len(cfg.Analysis.IncludePatterns) > 0 {
			includePatterns = cfg.Analysis.IncludePatterns
		}
		if len(cfg.Analysis.ExcludePatterns) > 0 {
			excludePatterns = cfg.Analysis.ExcludePatterns
		}
	}

	// Create complexity request
	req := domain.ComplexityRequest{
		Paths:           []string{path},
		MinComplexity:   minComplexity,
		MaxComplexity:   maxComplexity,
		ShowDetails:     showDetails,
		Recursive:       cfg == nil || cfg.Analysis.Recursive,
		OutputFormat:    domain.OutputFormatJSON,
		OutputWriter:    io.Discard,
		LowThreshold:    lowThreshold,
		MediumThreshold: mediumThreshold,
		SortBy:          sortBy,
		IncludePatterns: includePatterns,
		ExcludePatterns: excludePatterns,
		ConfigPath:      h.deps.ConfigPath(),
	}

	// Build use case with all required dependencies
	complexityService := service.NewComplexityService()
	fileReader := service.NewFileReader()
	formatter := service.NewOutputFormatter()
	configLoader := service.NewConfigurationLoader()

	useCase := app.NewComplexityUseCase(
		complexityService,
		fileReader,
		formatter,
		configLoader,
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("complexity analysis failed: %v", err)), nil
	}

	// Parse output_mode parameter (default: "summary")
	outputMode := "summary"
	if om, ok := args["output_mode"].(string); ok {
		outputMode = om
	}

	// Parse max_results parameter (default: unlimited for summary mode)
	maxResults := 0
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Format output based on mode
	var responseData interface{}
	switch outputMode {
	case "full":
		responseData = result
	case "detailed":
		responseData = formatComplexityDetailed(result, maxComplexity, maxResults)
	default: // "summary"
		responseData = formatComplexitySummary(result, maxComplexity, maxResults)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleDetectClones handles the detect_clones tool
func (h *HandlerSet) HandleDetectClones(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	// Load defaults from configuration
	cfg := h.deps.Config()
	req := domain.DefaultCloneRequest()
	if cfg != nil && cfg.Clones != nil {
		req.SimilarityThreshold = cfg.Clones.Thresholds.SimilarityThreshold
		req.MinLines = cfg.Clones.Analysis.MinLines
		req.MinNodes = cfg.Clones.Analysis.MinNodes
		req.GroupClones = cfg.Clones.Output.GroupClones
		req.Recursive = cfg.Clones.Input.Recursive
		if len(cfg.Clones.Input.IncludePatterns) > 0 {
			req.IncludePatterns = cfg.Clones.Input.IncludePatterns
		}
		if len(cfg.Clones.Input.ExcludePatterns) > 0 {
			req.ExcludePatterns = cfg.Clones.Input.ExcludePatterns
		}
	} else if cfg != nil {
		req.Recursive = cfg.Analysis.Recursive
		if len(cfg.Analysis.IncludePatterns) > 0 {
			req.IncludePatterns = cfg.Analysis.IncludePatterns
		}
		if len(cfg.Analysis.ExcludePatterns) > 0 {
			req.ExcludePatterns = cfg.Analysis.ExcludePatterns
		}
	}

	// Parse optional parameters
	similarityThreshold := req.SimilarityThreshold
	if st, ok := args["similarity_threshold"].(float64); ok {
		similarityThreshold = st
	}

	minLines := req.MinLines
	if ml, ok := args["min_lines"].(float64); ok {
		minLines = int(ml)
	}

	groupClones := req.GroupClones
	if gc, ok := args["group_clones"].(bool); ok {
		groupClones = gc
	}

	req.Paths = []string{path}
	req.SimilarityThreshold = similarityThreshold
	req.MinLines = minLines
	req.GroupClones = groupClones
	// Preserve MinNodes from defaults/config
	req.OutputFormat = domain.OutputFormatJSON
	req.OutputWriter = io.Discard
	req.ConfigPath = h.deps.ConfigPath()

	// Build use case with all required dependencies
	cloneService := service.NewCloneService()
	fileReader := service.NewFileReader()
	formatter := service.NewCloneOutputFormatter()
	configLoader := service.NewCloneConfigurationLoader()

	useCase := app.NewCloneUseCase(
		cloneService,
		fileReader,
		formatter,
		configLoader,
	)

	// Execute analysis
	result, err := useCase.ExecuteAndReturn(ctx, *req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("clone detection failed: %v", err)), nil
	}

	// Parse output_mode parameter (default: "summary")
	outputMode := "summary"
	if om, ok := args["output_mode"].(string); ok {
		outputMode = om
	}

	// Parse max_results parameter
	maxResults := 0
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Format output based on mode
	var responseData interface{}
	switch outputMode {
	case "full":
		responseData = result
	case "detailed":
		responseData = formatClonesDetailed(result, maxResults)
	default: // "summary"
		responseData = formatClonesSummary(result, maxResults)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleCheckCoupling handles the check_coupling tool
func (h *HandlerSet) HandleCheckCoupling(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	cfg := h.deps.Config()
	req := domain.DefaultCBORequest()
	req.Paths = []string{path}
	req.OutputFormat = domain.OutputFormatJSON
	req.OutputWriter = io.Discard
	req.ConfigPath = h.deps.ConfigPath()
	req.LowThreshold = 5
	req.MediumThreshold = 10
	req.SortBy = domain.SortByCoupling

	if cfg != nil {
		req.Recursive = domain.BoolPtr(cfg.Analysis.Recursive)
		if len(cfg.Analysis.IncludePatterns) > 0 {
			req.IncludePatterns = cfg.Analysis.IncludePatterns
		}
		if len(cfg.Analysis.ExcludePatterns) > 0 {
			req.ExcludePatterns = cfg.Analysis.ExcludePatterns
		}
	}

	// Build use case with all required dependencies
	cboService := service.NewCBOService()
	fileReader := service.NewFileReader()
	formatter := service.NewCBOFormatter()

	useCase := app.NewCBOUseCase(
		cboService,
		fileReader,
		formatter,
		nil, // CBO config loader is optional
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, *req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("coupling analysis failed: %v", err)), nil
	}

	// Parse output_mode parameter (default: "summary")
	outputMode := "summary"
	if om, ok := args["output_mode"].(string); ok {
		outputMode = om
	}

	// Parse max_results parameter
	maxResults := 0
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Parse min_cbo parameter (default threshold for filtering)
	minCBO := 10
	if mc, ok := args["min_cbo"].(float64); ok {
		minCBO = int(mc)
	}

	// Format output based on mode
	var responseData interface{}
	switch outputMode {
	case "full":
		responseData = result
	case "detailed":
		responseData = formatCouplingDetailed(result, minCBO, maxResults)
	default: // "summary"
		responseData = formatCouplingSummary(result, minCBO, maxResults)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleFindDeadCode handles the find_dead_code tool
func (h *HandlerSet) HandleFindDeadCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	// Parse min_severity
	cfg := h.deps.Config()
	minSeverity := domain.DeadCodeSeverityWarning
	if cfg != nil {
		switch cfg.DeadCode.MinSeverity {
		case "info":
			minSeverity = domain.DeadCodeSeverityInfo
		case "critical", "error":
			minSeverity = domain.DeadCodeSeverityCritical
		}
	}
	if ms, ok := args["min_severity"].(string); ok {
		switch ms {
		case "info":
			minSeverity = domain.DeadCodeSeverityInfo
		case "warning":
			minSeverity = domain.DeadCodeSeverityWarning
		case "critical", "error":
			minSeverity = domain.DeadCodeSeverityCritical
		}
	}

	// Create dead code request
	req := domain.DeadCodeRequest{
		Paths:        []string{path},
		MinSeverity:  minSeverity,
		Recursive:    true,
		OutputFormat: domain.OutputFormatJSON,
		OutputWriter: io.Discard,
		SortBy:       domain.DeadCodeSortBySeverity,
		ConfigPath:   h.deps.ConfigPath(),
	}
	if cfg != nil {
		req.Recursive = cfg.Analysis.Recursive
		if len(cfg.Analysis.IncludePatterns) > 0 {
			req.IncludePatterns = cfg.Analysis.IncludePatterns
		}
		if len(cfg.Analysis.ExcludePatterns) > 0 {
			req.ExcludePatterns = cfg.Analysis.ExcludePatterns
		}
	}

	// Build use case with all required dependencies
	deadCodeService := service.NewDeadCodeService()
	fileReader := service.NewFileReader()
	formatter := service.NewDeadCodeFormatter()
	configLoader := service.NewDeadCodeConfigurationLoader()

	useCase := app.NewDeadCodeUseCase(
		deadCodeService,
		fileReader,
		formatter,
		configLoader,
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("dead code analysis failed: %v", err)), nil
	}

	// Parse output_mode parameter (default: "summary")
	outputMode := "summary"
	if om, ok := args["output_mode"].(string); ok {
		outputMode = om
	}

	// Parse max_results parameter
	maxResults := 0
	if mr, ok := args["max_results"].(float64); ok {
		maxResults = int(mr)
	}

	// Format output based on mode
	var responseData interface{}
	switch outputMode {
	case "full":
		responseData = result
	case "detailed":
		responseData = formatDeadCodeDetailed(result, maxResults)
	default: // "summary"
		responseData = formatDeadCodeSummary(result, maxResults)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleGetHealthScore handles the get_health_score tool
func (h *HandlerSet) HandleGetHealthScore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("invalid arguments format"), nil
	}

	path, ok := args["path"].(string)
	if !ok {
		return mcp.NewToolResultError("path parameter is required and must be a string"), nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return mcp.NewToolResultError(fmt.Sprintf("path does not exist: %s", path)), nil
	}

	// Run comprehensive analysis to get health score
	analyzeUC, err := h.deps.BuildAnalyzeUseCase()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create analyzer: %v", err)), nil
	}

	config := app.AnalyzeUseCaseConfig{
		SkipComplexity:  false,
		SkipDeadCode:    false,
		SkipClones:      false,
		SkipCBO:         false,
		MinSeverity:     domain.DeadCodeSeverityWarning,
		MinComplexity:   1,
		CloneSimilarity: 0.8,
		ConfigFile:      h.deps.ConfigPath(),
	}
	if cfg := h.deps.Config(); cfg != nil {
		if cfg.Output.MinComplexity > 0 {
			config.MinComplexity = cfg.Output.MinComplexity
		}
		switch cfg.DeadCode.MinSeverity {
		case "info":
			config.MinSeverity = domain.DeadCodeSeverityInfo
		case "critical", "error":
			config.MinSeverity = domain.DeadCodeSeverityCritical
		default:
			config.MinSeverity = domain.DeadCodeSeverityWarning
		}
		if cfg.Clones != nil && cfg.Clones.Thresholds.SimilarityThreshold > 0 {
			config.CloneSimilarity = cfg.Clones.Thresholds.SimilarityThreshold
		}
	}

	result, err := analyzeUC.Execute(ctx, config, []string{path})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	// Extract health score summary
	healthScoreResult := map[string]interface{}{
		"health_score": result.Summary.HealthScore,
		"grade":        result.Summary.Grade,
		"is_healthy":   result.Summary.IsHealthy(),
		"category_scores": map[string]int{
			"complexity_score":   result.Summary.ComplexityScore,
			"dead_code_score":    result.Summary.DeadCodeScore,
			"duplication_score":  result.Summary.DuplicationScore,
			"coupling_score":     result.Summary.CouplingScore,
			"dependency_score":   result.Summary.DependencyScore,
			"architecture_score": result.Summary.ArchitectureScore,
		},
		"summary": map[string]interface{}{
			"total_files":           result.Summary.TotalFiles,
			"average_complexity":    result.Summary.AverageComplexity,
			"high_complexity_count": result.Summary.HighComplexityCount,
			"dead_code_count":       result.Summary.DeadCodeCount,
			"clone_pairs":           result.Summary.ClonePairs,
			"high_coupling_classes": result.Summary.HighCouplingClasses,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(healthScoreResult)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// formatComplexitySummary formats complexity results in compact summary mode
func formatComplexitySummary(result *domain.ComplexityResponse, threshold int, maxResults int) map[string]interface{} {
	issues := []string{}
	totalIssues := 0

	// Default threshold to 10 if not specified
	if threshold == 0 {
		threshold = 10
	}

	// Filter functions that exceed threshold
	for _, fn := range result.Functions {
		if fn.Metrics.Complexity > threshold {
			totalIssues++

			// Only add to issues array if within max_results limit
			if maxResults == 0 || len(issues) < maxResults {
				// Format: "file:line:col: function is too complex (X > threshold)"
				issue := fmt.Sprintf("%s:%d:%d: %s is too complex (%d > %d)",
					fn.FilePath, fn.StartLine, fn.StartColumn+1, fn.Name,
					fn.Metrics.Complexity, threshold)
				issues = append(issues, issue)
			}
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_issues":       totalIssues,
			"total_functions":    result.Summary.TotalFunctions,
			"max_complexity":     result.Summary.MaxComplexity,
			"average_complexity": result.Summary.AverageComplexity,
			"threshold":          threshold,
		},
	}
}

// formatComplexityDetailed formats complexity results with structured details
func formatComplexityDetailed(result *domain.ComplexityResponse, threshold int, maxResults int) map[string]interface{} {
	type Issue struct {
		File       string `json:"file"`
		Line       int    `json:"line"`
		Column     int    `json:"column"`
		Function   string `json:"function"`
		Complexity int    `json:"complexity"`
		Threshold  int    `json:"threshold"`
		Message    string `json:"message"`
	}

	issues := []Issue{}
	totalIssues := 0

	if threshold == 0 {
		threshold = 10
	}

	for _, fn := range result.Functions {
		if fn.Metrics.Complexity > threshold {
			totalIssues++

			// Only add to issues array if within max_results limit
			if maxResults == 0 || len(issues) < maxResults {
				issue := Issue{
					File:       fn.FilePath,
					Line:       fn.StartLine,
					Column:     fn.StartColumn + 1,
					Function:   fn.Name,
					Complexity: fn.Metrics.Complexity,
					Threshold:  threshold,
					Message:    fmt.Sprintf("is too complex (%d > %d)", fn.Metrics.Complexity, threshold),
				}
				issues = append(issues, issue)
			}
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_issues":       totalIssues,
			"total_functions":    result.Summary.TotalFunctions,
			"max_complexity":     result.Summary.MaxComplexity,
			"average_complexity": result.Summary.AverageComplexity,
			"threshold":          threshold,
		},
	}
}

// formatDeadCodeSummary formats dead code results in compact summary mode
func formatDeadCodeSummary(result *domain.DeadCodeResponse, maxResults int) map[string]interface{} {
	issues := []string{}
	totalIssues := 0
	criticalCount := 0
	warningCount := 0
	infoCount := 0

	for _, file := range result.Files {
		for _, function := range file.Functions {
			for _, finding := range function.Findings {
				totalIssues++

				switch finding.Severity {
				case domain.DeadCodeSeverityCritical:
					criticalCount++
				case domain.DeadCodeSeverityWarning:
					warningCount++
				case domain.DeadCodeSeverityInfo:
					infoCount++
				}

				if maxResults == 0 || len(issues) < maxResults {
					// Format: "file:line:col: message (severity)"
					issue := fmt.Sprintf("%s:%d:%d: %s (%s)",
						finding.Location.FilePath,
						finding.Location.StartLine,
						finding.Location.StartColumn+1,
						finding.Reason,
						finding.Severity)
					issues = append(issues, issue)
				}
			}
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_issues":    totalIssues,
			"critical_issues": criticalCount,
			"warning_issues":  warningCount,
			"info_issues":     infoCount,
			"files_analyzed":  len(result.Files),
		},
	}
}

// formatDeadCodeDetailed formats dead code results with structured details
func formatDeadCodeDetailed(result *domain.DeadCodeResponse, maxResults int) map[string]interface{} {
	type Issue struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Column   int    `json:"column"`
		Function string `json:"function"`
		Reason   string `json:"reason"`
		Severity string `json:"severity"`
	}

	issues := []Issue{}
	totalIssues := 0
	criticalCount := 0
	warningCount := 0
	infoCount := 0

	for _, file := range result.Files {
		for _, function := range file.Functions {
			for _, finding := range function.Findings {
				totalIssues++

				switch finding.Severity {
				case domain.DeadCodeSeverityCritical:
					criticalCount++
				case domain.DeadCodeSeverityWarning:
					warningCount++
				case domain.DeadCodeSeverityInfo:
					infoCount++
				}

				if maxResults == 0 || len(issues) < maxResults {
					issue := Issue{
						File:     finding.Location.FilePath,
						Line:     finding.Location.StartLine,
						Column:   finding.Location.StartColumn + 1,
						Function: function.Name,
						Reason:   finding.Reason,
						Severity: string(finding.Severity),
					}
					issues = append(issues, issue)
				}
			}
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_issues":    totalIssues,
			"critical_issues": criticalCount,
			"warning_issues":  warningCount,
			"info_issues":     infoCount,
			"files_analyzed":  len(result.Files),
		},
	}
}

// formatClonesSummary formats clone detection results in compact summary mode
func formatClonesSummary(result *domain.CloneResponse, maxResults int) map[string]interface{} {
	issues := []string{}
	filesWithClones := make(map[string]bool)

	for _, pair := range result.ClonePairs {
		filesWithClones[pair.Clone1.Location.FilePath] = true
		filesWithClones[pair.Clone2.Location.FilePath] = true

		if maxResults == 0 || len(issues) < maxResults {
			// Format: "file1:line:col: clone of file2:line:col (similarity%)"
			issue := fmt.Sprintf("%s:%d:%d: clone of %s:%d:%d (%.1f%%)",
				pair.Clone1.Location.FilePath,
				pair.Clone1.Location.StartLine,
				pair.Clone1.Location.StartCol+1,
				pair.Clone2.Location.FilePath,
				pair.Clone2.Location.StartLine,
				pair.Clone2.Location.StartCol+1,
				pair.Similarity*100)
			issues = append(issues, issue)
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_clone_pairs":  len(result.ClonePairs),
			"total_clone_groups": len(result.CloneGroups),
			"files_with_clones":  len(filesWithClones),
		},
	}
}

// formatClonesDetailed formats clone detection results with structured details
func formatClonesDetailed(result *domain.CloneResponse, maxResults int) map[string]interface{} {
	type Issue struct {
		File1      string  `json:"file1"`
		Line1      int     `json:"line1"`
		Col1       int     `json:"col1"`
		File2      string  `json:"file2"`
		Line2      int     `json:"line2"`
		Col2       int     `json:"col2"`
		Similarity float64 `json:"similarity"`
		Lines      int     `json:"lines"`
	}

	issues := []Issue{}
	filesWithClones := make(map[string]bool)

	for _, pair := range result.ClonePairs {
		filesWithClones[pair.Clone1.Location.FilePath] = true
		filesWithClones[pair.Clone2.Location.FilePath] = true

		if maxResults == 0 || len(issues) < maxResults {
			issue := Issue{
				File1:      pair.Clone1.Location.FilePath,
				Line1:      pair.Clone1.Location.StartLine,
				Col1:       pair.Clone1.Location.StartCol + 1,
				File2:      pair.Clone2.Location.FilePath,
				Line2:      pair.Clone2.Location.StartLine,
				Col2:       pair.Clone2.Location.StartCol + 1,
				Similarity: pair.Similarity,
				Lines:      pair.Clone1.Location.EndLine - pair.Clone1.Location.StartLine + 1,
			}
			issues = append(issues, issue)
		}
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"total_clone_pairs":  len(result.ClonePairs),
			"total_clone_groups": len(result.CloneGroups),
			"files_with_clones":  len(filesWithClones),
		},
	}
}

// formatCouplingSummary formats coupling results in compact summary mode
func formatCouplingSummary(result *domain.CBOResponse, threshold int, maxResults int) map[string]interface{} {
	issues := []string{}
	highCouplingCount := 0
	totalCBO := 0
	maxCBO := 0

	for _, class := range result.Classes {
		cbo := class.Metrics.CouplingCount
		if cbo > maxCBO {
			maxCBO = cbo
		}
		totalCBO += cbo

		if cbo > threshold {
			highCouplingCount++

			// Only add to issues array if within max_results limit
			if maxResults == 0 || len(issues) < maxResults {
				// Format: "file:line: ClassName has high coupling (CBO=X)"
				issue := fmt.Sprintf("%s:%d: %s has high coupling (CBO=%d)",
					class.FilePath, class.StartLine, class.Name, cbo)
				issues = append(issues, issue)
			}
		}
	}

	avgCBO := 0.0
	if len(result.Classes) > 0 {
		avgCBO = float64(totalCBO) / float64(len(result.Classes))
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"high_coupling_classes": highCouplingCount,
			"total_classes":         len(result.Classes),
			"max_cbo":               maxCBO,
			"average_cbo":           avgCBO,
			"threshold":             threshold,
		},
	}
}

// formatCouplingDetailed formats coupling results with structured details
func formatCouplingDetailed(result *domain.CBOResponse, threshold int, maxResults int) map[string]interface{} {
	type Issue struct {
		File      string `json:"file"`
		Line      int    `json:"line"`
		ClassName string `json:"class_name"`
		CBO       int    `json:"cbo"`
		Message   string `json:"message"`
	}

	issues := []Issue{}
	highCouplingCount := 0
	totalCBO := 0
	maxCBO := 0

	for _, class := range result.Classes {
		cbo := class.Metrics.CouplingCount
		if cbo > maxCBO {
			maxCBO = cbo
		}
		totalCBO += cbo

		if cbo > threshold {
			highCouplingCount++

			// Only add to issues array if within max_results limit
			if maxResults == 0 || len(issues) < maxResults {
				issue := Issue{
					File:      class.FilePath,
					Line:      class.StartLine,
					ClassName: class.Name,
					CBO:       cbo,
					Message:   fmt.Sprintf("has high coupling (CBO=%d)", cbo),
				}
				issues = append(issues, issue)
			}
		}
	}

	avgCBO := 0.0
	if len(result.Classes) > 0 {
		avgCBO = float64(totalCBO) / float64(len(result.Classes))
	}

	return map[string]interface{}{
		"issues": issues,
		"summary": map[string]interface{}{
			"high_coupling_classes": highCouplingCount,
			"total_classes":         len(result.Classes),
			"max_cbo":               maxCBO,
			"average_cbo":           avgCBO,
			"threshold":             threshold,
		},
	}
}

func buildAnalyzeUseCase(fileReader domain.FileReader) (*app.AnalyzeUseCase, error) {
	// Create config loaders
	complexityConfigLoader := service.NewConfigurationLoader()
	deadCodeConfigLoader := service.NewDeadCodeConfigurationLoader()
	cloneConfigLoader := service.NewCloneConfigurationLoader()
	systemConfigLoader := service.NewSystemAnalysisConfigurationLoader()

	// Build complexity use case
	complexityService := service.NewComplexityService()
	complexityFormatter := service.NewOutputFormatter()
	complexityUC := app.NewComplexityUseCase(complexityService, fileReader, complexityFormatter, complexityConfigLoader)

	// Build dead code use case
	deadCodeService := service.NewDeadCodeService()
	deadCodeFormatter := service.NewDeadCodeFormatter()
	deadCodeUC := app.NewDeadCodeUseCase(deadCodeService, fileReader, deadCodeFormatter, deadCodeConfigLoader)

	// Build clone use case
	cloneService := service.NewCloneService()
	cloneFormatter := service.NewCloneOutputFormatter()
	cloneUC := app.NewCloneUseCase(cloneService, fileReader, cloneFormatter, cloneConfigLoader)

	// Build CBO use case
	cboService := service.NewCBOService()
	cboFormatter := service.NewCBOFormatter()
	cboUC := app.NewCBOUseCase(cboService, fileReader, cboFormatter, nil) // CBO config loader is optional

	// Build system analysis use case
	systemService := service.NewSystemAnalysisService()
	systemFormatter := service.NewSystemAnalysisFormatter()
	systemUC := app.NewSystemAnalysisUseCase(systemService, fileReader, systemFormatter, systemConfigLoader)

	// Build analyze use case
	return app.NewAnalyzeUseCaseBuilder().
		WithComplexityUseCase(complexityUC).
		WithDeadCodeUseCase(deadCodeUC).
		WithCloneUseCase(cloneUC).
		WithCBOUseCase(cboUC).
		WithSystemUseCase(systemUC).
		WithFileReader(fileReader).
		WithProgressManager(service.NewProgressManager()).
		WithParallelExecutor(service.NewParallelExecutor()).
		WithErrorCategorizer(service.NewErrorCategorizer()).
		Build()
}
