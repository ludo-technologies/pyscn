package service

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// ComplexityServiceImpl implements the ComplexityService interface
type ComplexityServiceImpl struct {
	parser *parser.Parser
}

// NewComplexityService creates a new complexity service implementation
func NewComplexityService() *ComplexityServiceImpl {
	return &ComplexityServiceImpl{
		parser: parser.New(),
	}
}

// Analyze performs complexity analysis on multiple files
func (s *ComplexityServiceImpl) Analyze(ctx context.Context, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	var allScopes []domain.FunctionComplexity
	var allRawMetrics []domain.RawMetrics
	var rawMetricResults []*analyzer.RawMetricsResult
	var warnings []string
	var issues []analysisIssue
	filesProcessed := 0
	filesSkipped := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("complexity analysis cancelled: %w", ctx.Err())
		default:
		}

		// Progress reporting removed - file parsing is fast

		// Analyze single file
		scopes, rawMetrics, fileWarnings, fileIssues := s.analyzeFile(ctx, filePath, req)

		if rawMetrics != nil {
			allRawMetrics = append(allRawMetrics, *s.convertRawMetrics(rawMetrics))
			rawMetricResults = append(rawMetricResults, rawMetrics)
		}

		if len(fileIssues) > 0 {
			issues = append(issues, fileIssues...)
			filesSkipped++
			continue // Skip this file but continue with others
		}

		allScopes = append(allScopes, scopes...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	if len(allScopes) == 0 && len(allRawMetrics) == 0 {
		return nil, domain.NewAnalysisError("no execution scopes found to analyze", nil)
	}

	// Filter and sort results
	allFunctions, allClassScopes := partitionComplexityScopes(allScopes)
	moduleRollups := domain.AggregateComplexityByModule(allFunctions)
	filteredFunctions, _ := s.filterScopes(allFunctions, req)
	filteredClassScopes, _ := s.filterScopes(allClassScopes, req)
	sortedFunctions, err := domain.SortComplexityScopesBy(filteredFunctions, req.SortBy)
	if err != nil {
		return nil, domain.NewAnalysisError("invalid complexity sort", err)
	}
	sortedClassScopes, err := domain.SortComplexityScopesBy(filteredClassScopes, req.SortBy)
	if err != nil {
		return nil, domain.NewAnalysisError("invalid complexity sort", err)
	}

	// Generate summary over the complete population so min_complexity only
	// affects which functions are displayed, not the aggregate metrics.
	summary := s.generateSummary(allFunctions, allClassScopes, complexitySummaryCounts{
		filesAnalyzed: filesProcessed,
		filesSkipped:  filesSkipped,
	})
	rawMetricsSummary := s.convertAggregateRawMetrics(analyzer.CalculateAggregateRawMetrics(rawMetricResults))

	response := &domain.ComplexityResponse{
		Functions:           sortedFunctions,
		ClassScopes:         sortedClassScopes,
		AnalyzedFunctions:   allFunctions,
		AnalyzedClassScopes: allClassScopes,
		Summary:             summary,
		ModuleRollups:       moduleRollups,
		RawMetrics:          allRawMetrics,
		RawMetricsSummary:   rawMetricsSummary,
		Warnings:            warnings,
		Errors:              analysisIssueMessages(issues),
		Failures:            analyzerFailures(domain.AnalysisKindComplexity, issues),
		GeneratedAt:         time.Now().Format(time.RFC3339),
		Version:             version.Version, // Get version from version package
		Config:              s.buildConfigForResponse(req),
		Request:             &req,
	}
	if err := response.ValidateAnalyzedScopes(); err != nil {
		return nil, domain.NewAnalysisError("invalid complexity analysis result", err)
	}
	return response, nil
}

// AnalyzeSnapshot performs complexity analysis using already parsed project files.
func (s *ComplexityServiceImpl) AnalyzeSnapshot(ctx context.Context, snapshot *ProjectSnapshot, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("project snapshot cannot be nil")
	}

	var allScopes []domain.FunctionComplexity
	var allRawMetrics []domain.RawMetrics
	var rawMetricResults []*analyzer.RawMetricsResult
	var warnings []string
	var issues []analysisIssue
	filesProcessed := 0
	filesSkipped := 0
	parsedFiles := 0

	for _, file := range snapshot.analysisProjectFiles() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("complexity analysis cancelled: %w", ctx.Err())
		default:
		}
		if !file.Parsed() {
			continue
		}
		parsedFiles++

		scopes, rawMetrics, fileWarnings, fileIssues := s.analyzeProjectFile(file, req)

		if rawMetrics != nil {
			allRawMetrics = append(allRawMetrics, *s.convertRawMetrics(rawMetrics))
			rawMetricResults = append(rawMetricResults, rawMetrics)
		}

		if len(fileIssues) > 0 {
			issues = append(issues, fileIssues...)
			filesSkipped++
			continue
		}

		allScopes = append(allScopes, scopes...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}
	if parsedFiles > 0 && len(allScopes) == 0 && len(allRawMetrics) == 0 {
		return nil, domain.NewAnalysisError("no execution scopes found to analyze", nil)
	}

	allFunctions, allClassScopes := partitionComplexityScopes(allScopes)
	moduleRollups := domain.AggregateComplexityByModule(allFunctions)
	filteredFunctions, _ := s.filterScopes(allFunctions, req)
	filteredClassScopes, _ := s.filterScopes(allClassScopes, req)
	sortedFunctions, err := domain.SortComplexityScopesBy(filteredFunctions, req.SortBy)
	if err != nil {
		return nil, domain.NewAnalysisError("invalid complexity sort", err)
	}
	sortedClassScopes, err := domain.SortComplexityScopesBy(filteredClassScopes, req.SortBy)
	if err != nil {
		return nil, domain.NewAnalysisError("invalid complexity sort", err)
	}
	summary := s.generateSummary(allFunctions, allClassScopes, complexitySummaryCounts{
		filesAnalyzed: filesProcessed,
		filesSkipped:  filesSkipped,
	})
	rawMetricsSummary := s.convertAggregateRawMetrics(analyzer.CalculateAggregateRawMetrics(rawMetricResults))

	response := &domain.ComplexityResponse{
		Functions:           sortedFunctions,
		ClassScopes:         sortedClassScopes,
		AnalyzedFunctions:   allFunctions,
		AnalyzedClassScopes: allClassScopes,
		Summary:             summary,
		ModuleRollups:       moduleRollups,
		RawMetrics:          allRawMetrics,
		RawMetricsSummary:   rawMetricsSummary,
		Warnings:            warnings,
		Errors:              analysisIssueMessages(issues),
		Failures:            analyzerFailures(domain.AnalysisKindComplexity, issues),
		GeneratedAt:         time.Now().Format(time.RFC3339),
		Version:             version.Version,
		Config:              s.buildConfigForResponse(req),
		Request:             &req,
	}
	if err := response.ValidateAnalyzedScopes(); err != nil {
		return nil, domain.NewAnalysisError("invalid complexity analysis result", err)
	}
	return response, nil
}

// AnalyzeFile analyzes a single Python file
func (s *ComplexityServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	// Update the request to analyze only this file
	singleFileReq := req
	singleFileReq.Paths = []string{filePath}

	return s.Analyze(ctx, singleFileReq)
}

// analyzeFile performs complexity analysis on a single file
func (s *ComplexityServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) ([]domain.FunctionComplexity, *analyzer.RawMetricsResult, []string, []analysisIssue) {
	var scopes []domain.FunctionComplexity
	var warnings []string
	var issues []analysisIssue

	// Parse the file
	content, err := s.readFile(filePath)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Failed to read file: %v", err), cause: err})
		return scopes, nil, warnings, issues
	}

	rawMetrics := analyzer.CalculateRawMetrics(content, filePath)

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		// Enhanced error context with file path
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Parse error: %v", err), cause: err})
		return scopes, rawMetrics, warnings, issues
	}

	analyzer.PopulateLogicalLines(rawMetrics, result.AST)

	// Build a typed CFG for each Python execution scope.
	builder := analyzer.NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	if err != nil {
		// Enhanced error context with file path
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("CFG construction failed: %v", err), cause: err})
		return scopes, rawMetrics, warnings, issues
	}

	// Calculate complexity for each execution scope.
	complexityConfig := s.buildComplexityConfig(req)
	scopes, warnings = s.calculateScopeComplexities(filePath, cfgs, complexityConfig, req, rawMetrics)

	return scopes, rawMetrics, warnings, issues
}

func (s *ComplexityServiceImpl) analyzeProjectFile(file *ProjectFile, req domain.ComplexityRequest) ([]domain.FunctionComplexity, *analyzer.RawMetricsResult, []string, []analysisIssue) {
	var scopes []domain.FunctionComplexity
	var warnings []string
	var issues []analysisIssue

	if file == nil {
		issues = append(issues, analysisIssue{filePath: "unknown", message: "Invalid project file"})
		return scopes, nil, warnings, issues
	}
	if file.ReadErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Failed to read file: %v", file.ReadErr), cause: file.ReadErr})
		return scopes, nil, warnings, issues
	}

	rawMetrics := file.RawMetrics
	if rawMetrics == nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: "Project snapshot is missing raw metrics"})
		return scopes, nil, warnings, issues
	}
	if file.ParseErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Parse error: %v", file.ParseErr), cause: file.ParseErr})
		return scopes, rawMetrics, warnings, issues
	}

	cfgs, err := file.CFGs()
	if err != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("CFG construction failed: %v", err), cause: err})
		return scopes, rawMetrics, warnings, issues
	}

	complexityConfig := s.buildComplexityConfig(req)
	scopes, warnings = s.calculateScopeComplexities(file.Path, cfgs, complexityConfig, req, rawMetrics)
	return scopes, rawMetrics, warnings, issues
}

func (s *ComplexityServiceImpl) calculateScopeComplexities(filePath string, cfgs analyzer.ControlFlowGraphs, complexityConfig *config.ComplexityConfig, req domain.ComplexityRequest, rawMetrics *analyzer.RawMetricsResult) ([]domain.FunctionComplexity, []string) {
	var scopes []domain.FunctionComplexity
	var warnings []string

	for _, scopedCFG := range cfgs {
		functionName := scopedCFG.Scope.Name
		cfg := scopedCFG.Graph
		result := analyzer.CalculateComplexityWithConfig(cfg, complexityConfig)
		if result == nil {
			warnings = append(warnings, fmt.Sprintf("[%s:%s] failed to calculate complexity for scope", filePath, functionName))
			continue
		}

		result.SLOC = rawMetrics.FunctionSLOC(result.StartLine, result.EndLine)

		riskLevel := s.calculateRiskLevel(result.Complexity, result.CognitiveComplexity, result.NestingDepth, req)
		if complexityConfig.ShouldReport(result.Complexity) {
			warnings = append(warnings, s.metricThresholdWarnings(filePath, functionName, result, req)...)
		}

		scope := domain.FunctionComplexity{
			Name:        functionName,
			ScopeKind:   scopedCFG.Scope.Kind,
			FilePath:    filePath,
			StartLine:   result.StartLine,
			StartColumn: result.StartCol,
			EndLine:     result.EndLine,
			Metrics: domain.ComplexityMetrics{
				Complexity:          result.Complexity,
				CognitiveComplexity: result.CognitiveComplexity,
				Nodes:               result.Nodes,
				Edges:               result.Edges,
				NestingDepth:        result.NestingDepth,
				IfStatements:        result.IfStatements,
				LoopStatements:      result.LoopStatements,
				ExceptionHandlers:   result.ExceptionHandlers,
				SwitchCases:         result.SwitchCases,
				SLOC:                result.SLOC,
			},
			RiskLevel: riskLevel,
		}

		scopes = append(scopes, scope)
	}

	return scopes, warnings
}

func partitionComplexityScopes(scopes []domain.FunctionComplexity) (functions, classScopes []domain.FunctionComplexity) {
	functions = make([]domain.FunctionComplexity, 0, len(scopes))
	classScopes = make([]domain.FunctionComplexity, 0)
	for _, scope := range scopes {
		if scope.ScopeKind == domain.AnalysisScopeClass {
			classScopes = append(classScopes, scope)
			continue
		}
		functions = append(functions, scope)
	}
	return functions, classScopes
}

// filterScopes returns visible execution scopes and the count before min_complexity.
// report_unchanged remains part of the reporting contract, while module rollups
// consume the complete analyzer population before either presentation filter.
func (s *ComplexityServiceImpl) filterScopes(functions []domain.FunctionComplexity, req domain.ComplexityRequest) ([]domain.FunctionComplexity, int) {
	var filtered []domain.FunctionComplexity
	complexityConfig := s.buildComplexityConfig(req)
	functionsParsed := 0

	for _, function := range functions {
		if !complexityConfig.ShouldReport(function.Metrics.Complexity) {
			continue
		}
		functionsParsed++
		// Apply minimum complexity filter
		if function.Metrics.Complexity < req.MinComplexity {
			continue
		}

		// MaxComplexity is used only as a warning threshold (not for filtering)
		// See check command for threshold violation detection
		filtered = append(filtered, function)
	}

	return filtered, functionsParsed
}

// complexitySummaryCounts carries the labeled counts for generateSummary so
// call sites cannot transpose them.
type complexitySummaryCounts struct {
	filesAnalyzed int
	filesSkipped  int
}

// generateSummary creates summary statistics.
// functions must be the complete analyzer population: min_complexity and
// report_unchanged are presentation filters, so averages, distribution, and
// all function counts stay stable regardless of what is displayed.
// filesSkipped counts files that produced no metrics because they could not be
// read or parsed, so consumers can see when aggregates cover only part of the
// requested files.
func (s *ComplexityServiceImpl) generateSummary(functions, classScopes []domain.FunctionComplexity, counts complexitySummaryCounts) domain.ComplexitySummary {
	var maxClassComplexity, maxClassCognitiveComplexity, maxClassNestingDepth, highRiskClassScopes int
	for _, classScope := range classScopes {
		maxClassComplexity = max(maxClassComplexity, classScope.Metrics.Complexity)
		maxClassCognitiveComplexity = max(maxClassCognitiveComplexity, classScope.Metrics.CognitiveComplexity)
		maxClassNestingDepth = max(maxClassNestingDepth, classScope.Metrics.NestingDepth)
		if classScope.RiskLevel == domain.RiskLevelHigh {
			highRiskClassScopes++
		}
	}

	if len(functions) == 0 {
		return domain.ComplexitySummary{
			TotalClassScopes:            len(classScopes),
			MaxClassComplexity:          maxClassComplexity,
			MaxClassCognitiveComplexity: maxClassCognitiveComplexity,
			MaxClassNestingDepth:        maxClassNestingDepth,
			HighRiskClassScopes:         highRiskClassScopes,
			FilesAnalyzed:               counts.filesAnalyzed,
			TotalFiles:                  counts.filesAnalyzed + counts.filesSkipped,
			SkippedFiles:                counts.filesSkipped,
		}
	}

	var totalComplexity int
	var totalCognitiveComplexity int
	var totalNestingDepth int
	var maxComplexity int
	minComplexity := functions[0].Metrics.Complexity
	var lowCount, mediumCount, highCount int
	complexityDist := make(map[string]int)

	for _, function := range functions {
		complexity := function.Metrics.Complexity
		totalComplexity += complexity
		totalCognitiveComplexity += function.Metrics.CognitiveComplexity
		totalNestingDepth += function.Metrics.NestingDepth

		if complexity > maxComplexity {
			maxComplexity = complexity
		}
		if complexity < minComplexity {
			minComplexity = complexity
		}

		// Count risk levels
		switch function.RiskLevel {
		case domain.RiskLevelLow:
			lowCount++
		case domain.RiskLevelMedium:
			mediumCount++
		case domain.RiskLevelHigh:
			highCount++
		}

		// Build complexity distribution
		distKey := s.getComplexityDistributionKey(complexity)
		complexityDist[distKey]++
	}
	avgComplexity := float64(totalComplexity) / float64(len(functions))
	avgCognitiveComplexity := float64(totalCognitiveComplexity) / float64(len(functions))
	avgNestingDepth := float64(totalNestingDepth) / float64(len(functions))

	return domain.ComplexitySummary{
		TotalFunctions:              len(functions),
		TotalClassScopes:            len(classScopes),
		MaxClassComplexity:          maxClassComplexity,
		MaxClassCognitiveComplexity: maxClassCognitiveComplexity,
		MaxClassNestingDepth:        maxClassNestingDepth,
		HighRiskClassScopes:         highRiskClassScopes,
		FunctionsParsed:             len(functions),
		AverageComplexity:           avgComplexity,
		AverageCognitiveComplexity:  avgCognitiveComplexity,
		AverageNestingDepth:         avgNestingDepth,
		MaxComplexity:               maxComplexity,
		MinComplexity:               minComplexity,
		FilesAnalyzed:               counts.filesAnalyzed,
		TotalFiles:                  counts.filesAnalyzed + counts.filesSkipped,
		SkippedFiles:                counts.filesSkipped,
		LowRiskFunctions:            lowCount,
		MediumRiskFunctions:         mediumCount,
		HighRiskFunctions:           highCount,
		ComplexityDistribution:      complexityDist,
	}
}

// Helper methods
func (s *ComplexityServiceImpl) calculateRiskLevel(complexity, cognitiveComplexity, nestingDepth int, req domain.ComplexityRequest) domain.RiskLevel {
	cfg := s.buildComplexityConfig(req)
	return domain.RiskLevel(cfg.AssessRiskLevel(complexity, cognitiveComplexity, nestingDepth))
}

func (s *ComplexityServiceImpl) metricThresholdWarnings(filePath string, functionName string, result *analyzer.ComplexityResult, req domain.ComplexityRequest) []string {
	var warnings []string
	complexityConfig := s.buildComplexityConfig(req)

	if result.CognitiveComplexity > complexityConfig.CognitiveComplexityThreshold {
		warnings = append(warnings, fmt.Sprintf("[%s:%d:%d] %s cognitive complexity too high (%d > %d)",
			filePath, result.StartLine, result.StartCol+1, functionName, result.CognitiveComplexity, complexityConfig.CognitiveComplexityThreshold))
	}
	if result.NestingDepth > complexityConfig.NestingDepthThreshold {
		warnings = append(warnings, fmt.Sprintf("[%s:%d:%d] %s nesting depth too high (%d > %d)",
			filePath, result.StartLine, result.StartCol+1, functionName, result.NestingDepth, complexityConfig.NestingDepthThreshold))
	}

	return warnings
}

func (s *ComplexityServiceImpl) getComplexityDistributionKey(complexity int) string {
	if complexity == 1 {
		return "1"
	} else if complexity <= 5 {
		return "2-5"
	} else if complexity <= 10 {
		return "6-10"
	} else if complexity <= 20 {
		return "11-20"
	}
	return "21+"
}

func (s *ComplexityServiceImpl) buildComplexityConfig(req domain.ComplexityRequest) *config.ComplexityConfig {
	// Convert domain request to internal complexity config
	// This bridges the domain layer with the internal implementation
	cognitiveThreshold := req.CognitiveComplexityThreshold
	if cognitiveThreshold <= 0 {
		cognitiveThreshold = domain.DefaultCognitiveComplexityThreshold
	}
	nestingThreshold := req.NestingDepthThreshold
	if nestingThreshold <= 0 {
		nestingThreshold = domain.DefaultNestingDepthThreshold
	}
	slocWarnThreshold := req.FunctionSLOCWarnThreshold
	if slocWarnThreshold <= 0 {
		slocWarnThreshold = domain.DefaultFunctionSLOCWarnThreshold
	}
	slocCriticalThreshold := req.FunctionSLOCCriticalThreshold
	if slocCriticalThreshold <= 0 {
		slocCriticalThreshold = domain.DefaultFunctionSLOCCriticalThreshold
	}

	return &config.ComplexityConfig{
		LowThreshold:                  req.LowThreshold,
		MediumThreshold:               req.MediumThreshold,
		CognitiveComplexityThreshold:  cognitiveThreshold,
		NestingDepthThreshold:         nestingThreshold,
		FunctionSLOCWarnThreshold:     slocWarnThreshold,
		FunctionSLOCCriticalThreshold: slocCriticalThreshold,
		Enabled:                       domain.BoolValue(req.Enabled, true),
		ReportUnchanged:               domain.BoolValue(req.ReportUnchanged, true),
		MaxComplexity:                 req.MaxComplexity,
	}
}

func (s *ComplexityServiceImpl) buildConfigForResponse(req domain.ComplexityRequest) interface{} {
	return map[string]interface{}{
		"output_format":                    string(req.OutputFormat),
		"min_complexity":                   req.MinComplexity,
		"max_complexity":                   req.MaxComplexity,
		"low_threshold":                    req.LowThreshold,
		"medium_threshold":                 req.MediumThreshold,
		"cognitive_complexity_threshold":   req.CognitiveComplexityThreshold,
		"nesting_depth_threshold":          req.NestingDepthThreshold,
		"function_sloc_warn_threshold":     req.FunctionSLOCWarnThreshold,
		"function_sloc_critical_threshold": req.FunctionSLOCCriticalThreshold,
		"enabled":                          domain.BoolValue(req.Enabled, true),
		"report_unchanged":                 domain.BoolValue(req.ReportUnchanged, true),
		"sort_by":                          string(req.SortBy),
		"show_details":                     domain.BoolValue(req.ShowDetails, false),
		"recursive":                        domain.BoolValue(req.Recursive, true),
		"include_patterns":                 req.IncludePatterns,
		"exclude_patterns":                 req.ExcludePatterns,
	}
}

func (s *ComplexityServiceImpl) convertRawMetrics(result *analyzer.RawMetricsResult) *domain.RawMetrics {
	if result == nil {
		return nil
	}

	return &domain.RawMetrics{
		FilePath:       result.FilePath,
		SLOC:           result.SLOC,
		LLOC:           result.LLOC,
		CommentLines:   result.CommentLines,
		DocstringLines: result.DocstringLines,
		BlankLines:     result.BlankLines,
		TotalLines:     result.TotalLines,
		CommentRatio:   result.CommentRatio,
	}
}

func (s *ComplexityServiceImpl) convertAggregateRawMetrics(result *analyzer.AggregateRawMetrics) *domain.RawMetricsSummary {
	if result == nil {
		return nil
	}

	return &domain.RawMetricsSummary{
		FilesAnalyzed:  result.FilesAnalyzed,
		SLOC:           result.SLOC,
		LLOC:           result.LLOC,
		CommentLines:   result.CommentLines,
		DocstringLines: result.DocstringLines,
		BlankLines:     result.BlankLines,
		TotalLines:     result.TotalLines,
		CommentRatio:   result.CommentRatio,
	}
}

func (s *ComplexityServiceImpl) readFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}
