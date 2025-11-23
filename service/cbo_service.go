package service

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// CBOServiceImpl implements the CBOService interface
type CBOServiceImpl struct {
	parser *parser.Parser
}

// NewCBOService creates a new CBO service implementation
func NewCBOService() *CBOServiceImpl {
	return &CBOServiceImpl{
		parser: parser.New(),
	}
}

// Analyze performs CBO analysis on multiple files
func (s *CBOServiceImpl) Analyze(ctx context.Context, req domain.CBORequest) (*domain.CBOResponse, error) {
	var allClasses []domain.ClassCoupling
	var warnings []string
	var errors []string
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("CBO analysis cancelled: %w", ctx.Err())
		default:
		}

		// Progress reporting removed - file parsing is fast

		// Analyze single file
		classes, fileWarnings, fileErrors := s.analyzeFile(ctx, filePath, req)

		if len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
			continue // Skip this file but continue with others
		}

		allClasses = append(allClasses, classes...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	if len(allClasses) == 0 {
		warnings = append(warnings, "No classes found to analyze")
		// Return empty but valid response instead of error
		return &domain.CBOResponse{
			Classes:     []domain.ClassCoupling{},
			Summary:     s.generateSummary([]domain.ClassCoupling{}, filesProcessed, req),
			Warnings:    warnings,
			Errors:      errors,
			GeneratedAt: time.Now().Format(time.RFC3339),
			Version:     version.Version,
			Config:      s.buildConfigForResponse(req),
		}, nil
	}

	// Filter and sort results
	filteredClasses := s.filterClasses(allClasses, req)
	sortedClasses := s.sortClasses(filteredClasses, req.SortBy)

	// Generate summary
	summary := s.generateSummary(sortedClasses, filesProcessed, req)

	return &domain.CBOResponse{
		Classes:     sortedClasses,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      errors,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file
func (s *CBOServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.CBORequest) (*domain.CBOResponse, error) {
	// Update the request to analyze only this file
	singleFileReq := req
	singleFileReq.Paths = []string{filePath}

	return s.Analyze(ctx, singleFileReq)
}

// analyzeFile performs CBO analysis on a single file
func (s *CBOServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.CBORequest) ([]domain.ClassCoupling, []string, []string) {
	var classes []domain.ClassCoupling
	var warnings []string
	var errors []string

	// Parse the file
	content, err := s.readFile(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Failed to read file: %v", filePath, err))
		return classes, warnings, errors
	}

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Parse error: %v", filePath, err))
		return classes, warnings, errors
	}

	// Configure CBO analysis options
	options := s.buildCBOOptions(req)

	// Perform CBO analysis
	cboResults, err := analyzer.CalculateCBOWithConfig(result.AST, filePath, options)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] CBO analysis failed: %v", filePath, err))
		return classes, warnings, errors
	}

	if len(cboResults) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No classes found in file", filePath))
		return classes, warnings, errors
	}

	// Convert analyzer results to domain objects
	for _, cboResult := range cboResults {
		class := domain.ClassCoupling{
			Name:      cboResult.ClassName,
			FilePath:  cboResult.FilePath,
			StartLine: cboResult.StartLine,
			EndLine:   cboResult.EndLine,
			Metrics: domain.CBOMetrics{
				CouplingCount:               cboResult.CouplingCount,
				InheritanceDependencies:     cboResult.InheritanceDependencies,
				TypeHintDependencies:        cboResult.TypeHintDependencies,
				InstantiationDependencies:   cboResult.InstantiationDependencies,
				AttributeAccessDependencies: cboResult.AttributeAccessDependencies,
				ImportDependencies:          cboResult.ImportDependencies,
				DependentClasses:            cboResult.DependentClasses,
			},
			RiskLevel:   domain.RiskLevel(cboResult.RiskLevel),
			IsAbstract:  cboResult.IsAbstract,
			BaseClasses: cboResult.BaseClasses,
		}

		classes = append(classes, class)
	}

	return classes, warnings, errors
}

// filterClasses filters classes based on request criteria
func (s *CBOServiceImpl) filterClasses(classes []domain.ClassCoupling, req domain.CBORequest) []domain.ClassCoupling {
	var filtered []domain.ClassCoupling

	for _, class := range classes {
		// Filter by minimum CBO
		if class.Metrics.CouplingCount < req.MinCBO {
			continue
		}

		// Filter by maximum CBO (0 means no limit)
		if req.MaxCBO > 0 && class.Metrics.CouplingCount > req.MaxCBO {
			continue
		}

		// Filter out zero CBO classes if not requested
		if !domain.BoolValue(req.ShowZeros, false) && class.Metrics.CouplingCount == 0 {
			continue
		}

		filtered = append(filtered, class)
	}

	return filtered
}

// sortClasses sorts classes based on specified criteria
func (s *CBOServiceImpl) sortClasses(classes []domain.ClassCoupling, sortBy domain.SortCriteria) []domain.ClassCoupling {
	sorted := make([]domain.ClassCoupling, len(classes))
	copy(sorted, classes)

	switch sortBy {
	case domain.SortByCoupling:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Metrics.CouplingCount > sorted[j].Metrics.CouplingCount
		})
	case domain.SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
	case domain.SortByRisk:
		sort.Slice(sorted, func(i, j int) bool {
			riskOrder := map[domain.RiskLevel]int{
				domain.RiskLevelHigh:   3,
				domain.RiskLevelMedium: 2,
				domain.RiskLevelLow:    1,
			}
			return riskOrder[sorted[i].RiskLevel] > riskOrder[sorted[j].RiskLevel]
		})
	case domain.SortByLocation:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].FilePath != sorted[j].FilePath {
				return sorted[i].FilePath < sorted[j].FilePath
			}
			return sorted[i].StartLine < sorted[j].StartLine
		})
	default:
		// Default to sorting by coupling count (descending)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Metrics.CouplingCount > sorted[j].Metrics.CouplingCount
		})
	}

	return sorted
}

// generateSummary creates aggregate statistics
func (s *CBOServiceImpl) generateSummary(classes []domain.ClassCoupling, filesAnalyzed int, req domain.CBORequest) domain.CBOSummary {
	if len(classes) == 0 {
		return domain.CBOSummary{
			FilesAnalyzed: filesAnalyzed,
		}
	}

	summary := domain.CBOSummary{
		TotalClasses:       len(classes),
		ClassesAnalyzed:    len(classes),
		FilesAnalyzed:      filesAnalyzed,
		CBODistribution:    make(map[string]int),
		MostCoupledClasses: []domain.ClassCoupling{},
	}

	// Calculate statistics
	totalCBO := 0
	minCBO := classes[0].Metrics.CouplingCount
	maxCBO := classes[0].Metrics.CouplingCount

	for _, class := range classes {
		cbo := class.Metrics.CouplingCount
		totalCBO += cbo

		if cbo < minCBO {
			minCBO = cbo
		}
		if cbo > maxCBO {
			maxCBO = cbo
		}

		// Count by risk level
		switch class.RiskLevel {
		case domain.RiskLevelLow:
			summary.LowRiskClasses++
		case domain.RiskLevelMedium:
			summary.MediumRiskClasses++
		case domain.RiskLevelHigh:
			summary.HighRiskClasses++
		}

		// Build CBO distribution
		cboRange := s.getCBORange(cbo)
		summary.CBODistribution[cboRange]++
	}

	summary.AverageCBO = float64(totalCBO) / float64(len(classes))
	summary.MinCBO = minCBO
	summary.MaxCBO = maxCBO

	// Get top 10 most coupled classes
	sortedByCount := make([]domain.ClassCoupling, len(classes))
	copy(sortedByCount, classes)
	sort.Slice(sortedByCount, func(i, j int) bool {
		return sortedByCount[i].Metrics.CouplingCount > sortedByCount[j].Metrics.CouplingCount
	})

	maxTopClasses := 10
	if len(sortedByCount) < maxTopClasses {
		maxTopClasses = len(sortedByCount)
	}
	summary.MostCoupledClasses = sortedByCount[:maxTopClasses]

	return summary
}

// getCBORange returns a range string for CBO distribution
func (s *CBOServiceImpl) getCBORange(cbo int) string {
	switch {
	case cbo == 0:
		return "0"
	case cbo <= 5:
		return "1-5"
	case cbo <= 10:
		return "6-10"
	case cbo <= 20:
		return "11-20"
	case cbo <= 50:
		return "21-50"
	default:
		return "50+"
	}
}

// buildCBOOptions converts domain request to analyzer options
func (s *CBOServiceImpl) buildCBOOptions(req domain.CBORequest) *analyzer.CBOOptions {
	return &analyzer.CBOOptions{
		IncludeBuiltins:   domain.BoolValue(req.IncludeBuiltins, false),
		IncludeImports:    domain.BoolValue(req.IncludeImports, true),
		PublicClassesOnly: false, // Could add this to domain.CBORequest later
		ExcludePatterns:   req.ExcludePatterns,
		LowThreshold:      req.LowThreshold,
		MediumThreshold:   req.MediumThreshold,
	}
}

// buildConfigForResponse creates config info for response
func (s *CBOServiceImpl) buildConfigForResponse(req domain.CBORequest) interface{} {
	return map[string]interface{}{
		"minCBO":          req.MinCBO,
		"maxCBO":          req.MaxCBO,
		"showZeros":       domain.BoolValue(req.ShowZeros, false),
		"lowThreshold":    req.LowThreshold,
		"mediumThreshold": req.MediumThreshold,
		"includeBuiltins": domain.BoolValue(req.IncludeBuiltins, false),
		"includeImports":  domain.BoolValue(req.IncludeImports, true),
		"outputFormat":    req.OutputFormat,
		"sortBy":          req.SortBy,
	}
}

// readFile reads file content (extracted for testability)
func (s *CBOServiceImpl) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
