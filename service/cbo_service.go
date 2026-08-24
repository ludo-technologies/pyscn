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
	var issues []analysisIssue
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
		classes, fileWarnings, fileIssues := s.analyzeFile(ctx, filePath, req)

		if len(fileIssues) > 0 {
			issues = append(issues, fileIssues...)
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
			Errors:      analysisIssueMessages(issues),
			Failures:    analyzerFailures(domain.AnalysisKindCBO, issues),
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
		Errors:      analysisIssueMessages(issues),
		Failures:    analyzerFailures(domain.AnalysisKindCBO, issues),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeSnapshot performs CBO analysis using already parsed project files.
func (s *CBOServiceImpl) AnalyzeSnapshot(ctx context.Context, snapshot *ProjectSnapshot, req domain.CBORequest) (*domain.CBOResponse, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("project snapshot cannot be nil")
	}

	var allClasses []domain.ClassCoupling
	var warnings []string
	var issues []analysisIssue
	filesProcessed := 0

	for _, file := range snapshot.analysisProjectFiles() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("CBO analysis cancelled: %w", ctx.Err())
		default:
		}
		if !file.Parsed() {
			continue
		}

		classes, fileWarnings, fileIssues := s.analyzeProjectFile(file, req)

		if len(fileIssues) > 0 {
			issues = append(issues, fileIssues...)
			continue
		}

		allClasses = append(allClasses, classes...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	if len(allClasses) == 0 {
		warnings = append(warnings, "No classes found to analyze")
		return &domain.CBOResponse{
			Classes:     []domain.ClassCoupling{},
			Summary:     s.generateSummary([]domain.ClassCoupling{}, filesProcessed, req),
			Warnings:    warnings,
			Errors:      analysisIssueMessages(issues),
			Failures:    analyzerFailures(domain.AnalysisKindCBO, issues),
			GeneratedAt: time.Now().Format(time.RFC3339),
			Version:     version.Version,
			Config:      s.buildConfigForResponse(req),
		}, nil
	}

	filteredClasses := s.filterClasses(allClasses, req)
	sortedClasses := s.sortClasses(filteredClasses, req.SortBy)
	summary := s.generateSummary(sortedClasses, filesProcessed, req)

	return &domain.CBOResponse{
		Classes:     sortedClasses,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      analysisIssueMessages(issues),
		Failures:    analyzerFailures(domain.AnalysisKindCBO, issues),
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
func (s *CBOServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.CBORequest) ([]domain.ClassCoupling, []string, []analysisIssue) {
	var classes []domain.ClassCoupling
	var warnings []string
	var issues []analysisIssue

	// Parse the file
	content, err := s.readFile(filePath)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Failed to read file: %v", err), cause: err})
		return classes, warnings, issues
	}

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Parse error: %v", err), cause: err})
		return classes, warnings, issues
	}

	// Configure CBO analysis options
	options := s.buildCBOOptions(req)

	// Perform CBO analysis
	cboResults, err := analyzer.CalculateCBOWithConfig(result.AST, filePath, options)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("CBO analysis failed: %v", err), cause: err})
		return classes, warnings, issues
	}

	if len(cboResults) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No classes found in file", filePath))
		return classes, warnings, issues
	}

	classes = s.convertCBOResults(cboResults)
	return classes, warnings, issues
}

func (s *CBOServiceImpl) analyzeProjectFile(file *ProjectFile, req domain.CBORequest) ([]domain.ClassCoupling, []string, []analysisIssue) {
	var classes []domain.ClassCoupling
	var warnings []string
	var issues []analysisIssue

	if file == nil {
		issues = append(issues, analysisIssue{filePath: "unknown", message: "Invalid project file"})
		return classes, warnings, issues
	}
	if file.ReadErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Failed to read file: %v", file.ReadErr), cause: file.ReadErr})
		return classes, warnings, issues
	}
	if file.ParseErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Parse error: %v", file.ParseErr), cause: file.ParseErr})
		return classes, warnings, issues
	}

	options := s.buildCBOOptions(req)
	cboResults, err := analyzer.CalculateCBOWithConfig(file.AST, file.Path, options)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("CBO analysis failed: %v", err), cause: err})
		return classes, warnings, issues
	}

	if len(cboResults) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No classes found in file", file.Path))
		return classes, warnings, issues
	}

	classes = s.convertCBOResults(cboResults)
	return classes, warnings, issues
}

func (s *CBOServiceImpl) convertCBOResults(cboResults []*analyzer.CBOResult) []domain.ClassCoupling {
	classes := make([]domain.ClassCoupling, 0, len(cboResults))

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

	return classes
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
			if sorted[i].Metrics.CouplingCount != sorted[j].Metrics.CouplingCount {
				return sorted[i].Metrics.CouplingCount > sorted[j].Metrics.CouplingCount
			}
			return cboClassLocationLess(sorted[i], sorted[j])
		})
	case domain.SortByName:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Name != sorted[j].Name {
				return sorted[i].Name < sorted[j].Name
			}
			return cboClassLocationLess(sorted[i], sorted[j])
		})
	case domain.SortByRisk:
		sort.Slice(sorted, func(i, j int) bool {
			riskOrder := map[domain.RiskLevel]int{
				domain.RiskLevelHigh:   3,
				domain.RiskLevelMedium: 2,
				domain.RiskLevelLow:    1,
			}
			if riskOrder[sorted[i].RiskLevel] != riskOrder[sorted[j].RiskLevel] {
				return riskOrder[sorted[i].RiskLevel] > riskOrder[sorted[j].RiskLevel]
			}
			return cboClassLocationLess(sorted[i], sorted[j])
		})
	case domain.SortByLocation:
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].FilePath != sorted[j].FilePath {
				return sorted[i].FilePath < sorted[j].FilePath
			}
			if sorted[i].StartLine != sorted[j].StartLine {
				return sorted[i].StartLine < sorted[j].StartLine
			}
			return sorted[i].Name < sorted[j].Name
		})
	default:
		// Default to sorting by coupling count (descending)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Metrics.CouplingCount != sorted[j].Metrics.CouplingCount {
				return sorted[i].Metrics.CouplingCount > sorted[j].Metrics.CouplingCount
			}
			return cboClassLocationLess(sorted[i], sorted[j])
		})
	}

	return sorted
}

func cboClassLocationLess(a, b domain.ClassCoupling) bool {
	if a.FilePath != b.FilePath {
		return a.FilePath < b.FilePath
	}
	if a.StartLine != b.StartLine {
		return a.StartLine < b.StartLine
	}
	return a.Name < b.Name
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
		if sortedByCount[i].Metrics.CouplingCount != sortedByCount[j].Metrics.CouplingCount {
			return sortedByCount[i].Metrics.CouplingCount > sortedByCount[j].Metrics.CouplingCount
		}
		return cboClassLocationLess(sortedByCount[i], sortedByCount[j])
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
		IncludeBuiltins:       domain.BoolValue(req.IncludeBuiltins, false),
		IncludeImports:        domain.BoolValue(req.IncludeImports, true),
		GroupNamespaceImports: domain.BoolValue(req.GroupNamespaceImports, true),
		PublicClassesOnly:     false, // Could add this to domain.CBORequest later
		ExcludePatterns:       req.ExcludePatterns,
		LowThreshold:          req.LowThreshold,
		MediumThreshold:       req.MediumThreshold,
	}
}

// buildConfigForResponse creates config info for response
func (s *CBOServiceImpl) buildConfigForResponse(req domain.CBORequest) interface{} {
	return map[string]interface{}{
		"min_cbo":                 req.MinCBO,
		"max_cbo":                 req.MaxCBO,
		"show_zeros":              domain.BoolValue(req.ShowZeros, false),
		"low_threshold":           req.LowThreshold,
		"medium_threshold":        req.MediumThreshold,
		"include_builtins":        domain.BoolValue(req.IncludeBuiltins, false),
		"include_imports":         domain.BoolValue(req.IncludeImports, true),
		"group_namespace_imports": domain.BoolValue(req.GroupNamespaceImports, true),
		"output_format":           req.OutputFormat,
		"sort_by":                 req.SortBy,
	}
}

// readFile reads file content (extracted for testability)
func (s *CBOServiceImpl) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}
