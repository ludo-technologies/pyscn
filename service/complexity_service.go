package service

import (
	"context"
	"fmt"
	"os"
	"sort"
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
	var allFunctions []domain.FunctionComplexity
	var warnings []string
	var errors []string
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("complexity analysis cancelled: %w", ctx.Err())
		default:
		}

		// Progress reporting removed - file parsing is fast

		// Analyze single file
		functions, fileWarnings, fileErrors := s.analyzeFile(ctx, filePath, req)

		if len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
			continue // Skip this file but continue with others
		}

		allFunctions = append(allFunctions, functions...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	if len(allFunctions) == 0 {
		return nil, domain.NewAnalysisError("no functions found to analyze", nil)
	}

	// Filter and sort results
	filteredFunctions := s.filterFunctions(allFunctions, req)
	sortedFunctions := s.sortFunctions(filteredFunctions, req.SortBy)

	// Generate summary
	summary := s.generateSummary(sortedFunctions, filesProcessed, req)

	return &domain.ComplexityResponse{
		Functions:   sortedFunctions,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      errors,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version, // Get version from version package
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file
func (s *ComplexityServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	// Update the request to analyze only this file
	singleFileReq := req
	singleFileReq.Paths = []string{filePath}

	return s.Analyze(ctx, singleFileReq)
}

// analyzeFile performs complexity analysis on a single file
func (s *ComplexityServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) ([]domain.FunctionComplexity, []string, []string) {
	var functions []domain.FunctionComplexity
	var warnings []string
	var errors []string

	// Parse the file
	content, err := s.readFile(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Failed to read file: %v", filePath, err))
		return functions, warnings, errors
	}

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		// Enhanced error context with file path
		errors = append(errors, fmt.Sprintf("[%s] Parse error: %v", filePath, err))
		return functions, warnings, errors
	}

	// Build CFGs for all functions
	builder := analyzer.NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	if err != nil {
		// Enhanced error context with file path
		errors = append(errors, fmt.Sprintf("[%s] CFG construction failed: %v", filePath, err))
		return functions, warnings, errors
	}

	if len(cfgs) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No functions found in file", filePath))
		return functions, warnings, errors
	}

	// Calculate complexity for each function
	complexityConfig := s.buildComplexityConfig(req)

	for functionName, cfg := range cfgs {
		result := analyzer.CalculateComplexityWithConfig(cfg, complexityConfig)
		if result == nil {
			warnings = append(warnings, fmt.Sprintf("[%s:%s] Failed to calculate complexity for function", filePath, functionName))
			continue
		}

		riskLevel := s.calculateRiskLevel(result.Complexity, req)

		function := domain.FunctionComplexity{
			Name:     functionName,
			FilePath: filePath,
			Metrics: domain.ComplexityMetrics{
				Complexity:        result.Complexity,
				Nodes:             result.Nodes,
				Edges:             result.Edges,
				IfStatements:      result.IfStatements,
				LoopStatements:    result.LoopStatements,
				ExceptionHandlers: result.ExceptionHandlers,
				SwitchCases:       result.SwitchCases,
			},
			RiskLevel: riskLevel,
		}

		functions = append(functions, function)
	}

	return functions, warnings, errors
}

// filterFunctions filters functions based on complexity thresholds
func (s *ComplexityServiceImpl) filterFunctions(functions []domain.FunctionComplexity, req domain.ComplexityRequest) []domain.FunctionComplexity {
	var filtered []domain.FunctionComplexity

	for _, function := range functions {
		// Apply minimum complexity filter
		if function.Metrics.Complexity < req.MinComplexity {
			continue
		}

		// Apply maximum complexity filter
		if req.MaxComplexity > 0 && function.Metrics.Complexity > req.MaxComplexity {
			continue
		}

		filtered = append(filtered, function)
	}

	return filtered
}

// sortFunctions sorts functions based on the specified criteria
func (s *ComplexityServiceImpl) sortFunctions(functions []domain.FunctionComplexity, sortBy domain.SortCriteria) []domain.FunctionComplexity {
	// Create a copy to avoid modifying the original slice
	sorted := make([]domain.FunctionComplexity, len(functions))
	copy(sorted, functions)

	switch sortBy {
	case domain.SortByComplexity:
		s.sortByComplexity(sorted)
	case domain.SortByName:
		s.sortByName(sorted)
	case domain.SortByRisk:
		s.sortByRisk(sorted)
	}

	return sorted
}

// Helper methods for sorting - using efficient Go standard library sorting
func (s *ComplexityServiceImpl) sortByComplexity(functions []domain.FunctionComplexity) {
	// Sort by complexity (descending) - O(n log n) instead of O(n²)
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Metrics.Complexity > functions[j].Metrics.Complexity
	})
}

func (s *ComplexityServiceImpl) sortByName(functions []domain.FunctionComplexity) {
	// Sort by name (ascending) - O(n log n) instead of O(n²)
	sort.Slice(functions, func(i, j int) bool {
		return functions[i].Name < functions[j].Name
	})
}

func (s *ComplexityServiceImpl) sortByRisk(functions []domain.FunctionComplexity) {
	// Sort by risk level (high to low) - O(n log n) instead of O(n²)
	riskOrder := map[domain.RiskLevel]int{
		domain.RiskLevelHigh:   3,
		domain.RiskLevelMedium: 2,
		domain.RiskLevelLow:    1,
	}

	sort.Slice(functions, func(i, j int) bool {
		// Primary sort by risk level (high to low)
		if riskOrder[functions[i].RiskLevel] != riskOrder[functions[j].RiskLevel] {
			return riskOrder[functions[i].RiskLevel] > riskOrder[functions[j].RiskLevel]
		}
		// Secondary sort by complexity within same risk level
		return functions[i].Metrics.Complexity > functions[j].Metrics.Complexity
	})
}

// generateSummary creates summary statistics
func (s *ComplexityServiceImpl) generateSummary(functions []domain.FunctionComplexity, filesAnalyzed int, req domain.ComplexityRequest) domain.ComplexitySummary {
	if len(functions) == 0 {
		return domain.ComplexitySummary{
			FilesAnalyzed: filesAnalyzed,
		}
	}

	var totalComplexity int
	var maxComplexity int
	minComplexity := functions[0].Metrics.Complexity
	var lowCount, mediumCount, highCount int
	complexityDist := make(map[string]int)

	for _, function := range functions {
		complexity := function.Metrics.Complexity
		totalComplexity += complexity

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

	return domain.ComplexitySummary{
		TotalFunctions:         len(functions),
		AverageComplexity:      avgComplexity,
		MaxComplexity:          maxComplexity,
		MinComplexity:          minComplexity,
		FilesAnalyzed:          filesAnalyzed,
		LowRiskFunctions:       lowCount,
		MediumRiskFunctions:    mediumCount,
		HighRiskFunctions:      highCount,
		ComplexityDistribution: complexityDist,
	}
}

// Helper methods
func (s *ComplexityServiceImpl) calculateRiskLevel(complexity int, req domain.ComplexityRequest) domain.RiskLevel {
	if complexity <= req.LowThreshold {
		return domain.RiskLevelLow
	} else if complexity <= req.MediumThreshold {
		return domain.RiskLevelMedium
	}
	return domain.RiskLevelHigh
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
	return &config.ComplexityConfig{
		LowThreshold:    req.LowThreshold,
		MediumThreshold: req.MediumThreshold,
		Enabled:         true,
		ReportUnchanged: true,
		MaxComplexity:   req.MaxComplexity,
	}
}

func (s *ComplexityServiceImpl) buildConfigForResponse(req domain.ComplexityRequest) interface{} {
	return map[string]interface{}{
		"output_format":    string(req.OutputFormat),
		"min_complexity":   req.MinComplexity,
		"max_complexity":   req.MaxComplexity,
		"low_threshold":    req.LowThreshold,
		"medium_threshold": req.MediumThreshold,
		"sort_by":          string(req.SortBy),
		"show_details":     req.ShowDetails,
		"recursive":        req.Recursive,
		"include_patterns": req.IncludePatterns,
		"exclude_patterns": req.ExcludePatterns,
	}
}

func (s *ComplexityServiceImpl) readFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return content, nil
}
