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

// HandleAnalyzeCode handles the analyze_code tool
func HandleAnalyzeCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Parse recursive flag
	recursive := true
	if r, ok := args["recursive"].(bool); ok {
		recursive = r
	}

	// Build use case with dependencies
	fileReader := service.NewFileReader()

	// Create config for analyze use case
	config := app.AnalyzeUseCaseConfig{
		SkipComplexity: !contains(analyses, "complexity") && len(analyses) > 0,
		SkipDeadCode:   !contains(analyses, "dead_code") && len(analyses) > 0,
		SkipClones:     !contains(analyses, "clone") && len(analyses) > 0,
		SkipCBO:        !contains(analyses, "cbo") && len(analyses) > 0,
		SkipSystem:     !contains(analyses, "deps") && len(analyses) > 0,
		MinSeverity:    domain.DeadCodeSeverityWarning,
	}

	// Build analyze use case using builder pattern
	analyzeUC, err := buildAnalyzeUseCase(fileReader)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create analyzer: %v", err)), nil
	}

	// Collect files
	paths := []string{path}
	if !recursive {
		// Just use the provided path
		paths = []string{path}
	}

	// Execute analysis
	result, err := analyzeUC.Execute(ctx, config, paths)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("analysis failed: %v", err)), nil
	}

	// Convert result to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleCheckComplexity handles the check_complexity tool
func HandleCheckComplexity(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	minComplexity := 1
	if mc, ok := args["min_complexity"].(float64); ok {
		minComplexity = int(mc)
	}

	maxComplexity := 0
	if mc, ok := args["max_complexity"].(float64); ok {
		maxComplexity = int(mc)
	}

	showDetails := true
	if sd, ok := args["show_details"].(bool); ok {
		showDetails = sd
	}

	// Create complexity request
	req := domain.ComplexityRequest{
		Paths:           []string{path},
		MinComplexity:   minComplexity,
		MaxComplexity:   maxComplexity,
		ShowDetails:     showDetails,
		Recursive:       true,
		OutputFormat:    domain.OutputFormatJSON,
		OutputWriter:    io.Discard,
		LowThreshold:    9,
		MediumThreshold: 19,
		SortBy:          domain.SortByComplexity,
	}

	// Build use case with all required dependencies
	complexityService := service.NewComplexityService()
	fileReader := service.NewFileReader()
	formatter := service.NewOutputFormatter()

	useCase := app.NewComplexityUseCase(
		complexityService,
		fileReader,
		formatter,
		nil, // config loader
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("complexity analysis failed: %v", err)), nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleDetectClones handles the detect_clones tool
func HandleDetectClones(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	similarityThreshold := 0.8
	if st, ok := args["similarity_threshold"].(float64); ok {
		similarityThreshold = st
	}

	minLines := 5
	if ml, ok := args["min_lines"].(float64); ok {
		minLines = int(ml)
	}

	groupClones := true
	if gc, ok := args["group_clones"].(bool); ok {
		groupClones = gc
	}

	// Create clone request
	req := domain.DefaultCloneRequest()
	req.Paths = []string{path}
	req.SimilarityThreshold = similarityThreshold
	req.MinLines = minLines
	req.GroupClones = groupClones
	req.Recursive = true
	req.OutputFormat = domain.OutputFormatJSON
	req.OutputWriter = io.Discard

	// Build use case with all required dependencies
	cloneService := service.NewCloneService()
	fileReader := service.NewFileReader()
	formatter := service.NewCloneOutputFormatter()

	useCase := app.NewCloneUseCase(
		cloneService,
		fileReader,
		formatter,
		nil, // config loader
	)

	// Execute analysis
	result, err := useCase.ExecuteAndReturn(ctx, *req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("clone detection failed: %v", err)), nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleCheckCoupling handles the check_coupling tool
func HandleCheckCoupling(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	// Create CBO request
	req := domain.CBORequest{
		Paths:           []string{path},
		Recursive:       true,
		OutputFormat:    domain.OutputFormatJSON,
		OutputWriter:    io.Discard,
		LowThreshold:    5,
		MediumThreshold: 10,
		SortBy:          domain.SortByCoupling,
	}

	// Build use case with all required dependencies
	cboService := service.NewCBOService()
	fileReader := service.NewFileReader()
	formatter := service.NewCBOFormatter()

	useCase := app.NewCBOUseCase(
		cboService,
		fileReader,
		formatter,
		nil, // config loader
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("coupling analysis failed: %v", err)), nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleFindDeadCode handles the find_dead_code tool
func HandleFindDeadCode(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	minSeverity := domain.DeadCodeSeverityWarning
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
	}

	// Build use case with all required dependencies
	deadCodeService := service.NewDeadCodeService()
	fileReader := service.NewFileReader()
	formatter := service.NewDeadCodeFormatter()

	useCase := app.NewDeadCodeUseCase(
		deadCodeService,
		fileReader,
		formatter,
		nil, // config loader
	)

	// Execute analysis
	result, err := useCase.AnalyzeAndReturn(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("dead code analysis failed: %v", err)), nil
	}

	// Convert to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// HandleGetHealthScore handles the get_health_score tool
func HandleGetHealthScore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	fileReader := service.NewFileReader()

	analyzeUC, err := buildAnalyzeUseCase(fileReader)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create analyzer: %v", err)), nil
	}

	config := app.AnalyzeUseCaseConfig{
		SkipComplexity: false,
		SkipDeadCode:   false,
		SkipClones:     false,
		SkipCBO:        false,
		MinSeverity:    domain.DeadCodeSeverityWarning,
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

func buildAnalyzeUseCase(fileReader domain.FileReader) (*app.AnalyzeUseCase, error) {
	// Build complexity use case
	complexityService := service.NewComplexityService()
	complexityFormatter := service.NewOutputFormatter()
	complexityUC := app.NewComplexityUseCase(complexityService, fileReader, complexityFormatter, nil)

	// Build dead code use case
	deadCodeService := service.NewDeadCodeService()
	deadCodeFormatter := service.NewDeadCodeFormatter()
	deadCodeUC := app.NewDeadCodeUseCase(deadCodeService, fileReader, deadCodeFormatter, nil)

	// Build clone use case
	cloneService := service.NewCloneService()
	cloneFormatter := service.NewCloneOutputFormatter()
	cloneUC := app.NewCloneUseCase(cloneService, fileReader, cloneFormatter, nil)

	// Build CBO use case
	cboService := service.NewCBOService()
	cboFormatter := service.NewCBOFormatter()
	cboUC := app.NewCBOUseCase(cboService, fileReader, cboFormatter, nil)

	// Build system analysis use case
	systemService := service.NewSystemAnalysisService()
	systemFormatter := service.NewSystemAnalysisFormatter()
	systemUC := app.NewSystemAnalysisUseCase(systemService, fileReader, systemFormatter, nil)

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
