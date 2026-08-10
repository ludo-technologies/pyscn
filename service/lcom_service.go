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

// LCOMServiceImpl implements the LCOMService interface
type LCOMServiceImpl struct {
	parser *parser.Parser
}

// NewLCOMService creates a new LCOM service implementation
func NewLCOMService() *LCOMServiceImpl {
	return &LCOMServiceImpl{
		parser: parser.New(),
	}
}

// Analyze performs LCOM analysis on multiple files
func (s *LCOMServiceImpl) Analyze(ctx context.Context, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	var allClasses []domain.ClassCohesion
	var warnings []string
	var issues []analysisIssue
	filesProcessed := 0

	for _, filePath := range req.Paths {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("LCOM analysis cancelled: %w", ctx.Err())
		default:
		}

		classes, fileWarnings, fileIssues := s.analyzeFile(ctx, filePath, req)

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
		return &domain.LCOMResponse{
			Classes:     []domain.ClassCohesion{},
			Summary:     s.generateSummary([]domain.ClassCohesion{}, filesProcessed, req),
			Warnings:    warnings,
			Errors:      analysisIssueMessages(issues),
			Failures:    analyzerFailures(domain.AnalysisKindLCOM, issues),
			GeneratedAt: time.Now().Format(time.RFC3339),
			Version:     version.Version,
			Config:      s.buildConfigForResponse(req),
		}, nil
	}

	filteredClasses := s.filterClasses(allClasses, req)
	sortedClasses := s.sortClasses(filteredClasses, req.SortBy)
	summary := s.generateSummary(sortedClasses, filesProcessed, req)

	return &domain.LCOMResponse{
		Classes:     sortedClasses,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      analysisIssueMessages(issues),
		Failures:    analyzerFailures(domain.AnalysisKindLCOM, issues),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeSnapshot performs LCOM analysis using already parsed project files.
func (s *LCOMServiceImpl) AnalyzeSnapshot(ctx context.Context, snapshot *ProjectSnapshot, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("project snapshot cannot be nil")
	}

	var allClasses []domain.ClassCohesion
	var warnings []string
	var issues []analysisIssue
	filesProcessed := 0

	for _, file := range snapshot.analysisProjectFiles() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("LCOM analysis cancelled: %w", ctx.Err())
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
		return &domain.LCOMResponse{
			Classes:     []domain.ClassCohesion{},
			Summary:     s.generateSummary([]domain.ClassCohesion{}, filesProcessed, req),
			Warnings:    warnings,
			Errors:      analysisIssueMessages(issues),
			Failures:    analyzerFailures(domain.AnalysisKindLCOM, issues),
			GeneratedAt: time.Now().Format(time.RFC3339),
			Version:     version.Version,
			Config:      s.buildConfigForResponse(req),
		}, nil
	}

	filteredClasses := s.filterClasses(allClasses, req)
	sortedClasses := s.sortClasses(filteredClasses, req.SortBy)
	summary := s.generateSummary(sortedClasses, filesProcessed, req)

	return &domain.LCOMResponse{
		Classes:     sortedClasses,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      analysisIssueMessages(issues),
		Failures:    analyzerFailures(domain.AnalysisKindLCOM, issues),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file
func (s *LCOMServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	singleFileReq := req
	singleFileReq.Paths = []string{filePath}
	return s.Analyze(ctx, singleFileReq)
}

// analyzeFile performs LCOM analysis on a single file
func (s *LCOMServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.LCOMRequest) ([]domain.ClassCohesion, []string, []analysisIssue) {
	var classes []domain.ClassCohesion
	var warnings []string
	var issues []analysisIssue

	content, err := os.ReadFile(filePath)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Failed to read file: %v", err)})
		return classes, warnings, issues
	}

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("Parse error: %v", err)})
		return classes, warnings, issues
	}

	options := s.buildLCOMOptions(req)
	lcomResults, err := analyzer.CalculateLCOMWithConfig(result.AST, filePath, options)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: filePath, message: fmt.Sprintf("LCOM analysis failed: %v", err)})
		return classes, warnings, issues
	}

	if len(lcomResults) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No classes found in file", filePath))
		return classes, warnings, issues
	}

	classes = s.convertLCOMResults(lcomResults)
	return classes, warnings, issues
}

func (s *LCOMServiceImpl) analyzeProjectFile(file *ProjectFile, req domain.LCOMRequest) ([]domain.ClassCohesion, []string, []analysisIssue) {
	var classes []domain.ClassCohesion
	var warnings []string
	var issues []analysisIssue

	if file == nil {
		issues = append(issues, analysisIssue{filePath: "unknown", message: "Invalid project file"})
		return classes, warnings, issues
	}
	if file.ReadErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Failed to read file: %v", file.ReadErr)})
		return classes, warnings, issues
	}
	if file.ParseErr != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("Parse error: %v", file.ParseErr)})
		return classes, warnings, issues
	}

	options := s.buildLCOMOptions(req)
	lcomResults, err := analyzer.CalculateLCOMWithConfig(file.AST, file.Path, options)
	if err != nil {
		issues = append(issues, analysisIssue{filePath: file.Path, message: fmt.Sprintf("LCOM analysis failed: %v", err)})
		return classes, warnings, issues
	}

	if len(lcomResults) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No classes found in file", file.Path))
		return classes, warnings, issues
	}

	classes = s.convertLCOMResults(lcomResults)
	return classes, warnings, issues
}

func (s *LCOMServiceImpl) convertLCOMResults(lcomResults []*analyzer.LCOMResult) []domain.ClassCohesion {
	classes := make([]domain.ClassCohesion, 0, len(lcomResults))

	for _, lcomResult := range lcomResults {
		class := domain.ClassCohesion{
			Name:      lcomResult.ClassName,
			FilePath:  lcomResult.FilePath,
			StartLine: lcomResult.StartLine,
			EndLine:   lcomResult.EndLine,
			Metrics: domain.LCOMMetrics{
				LCOM4:             lcomResult.LCOM4,
				TotalMethods:      lcomResult.TotalMethods,
				ExcludedMethods:   lcomResult.ExcludedMethods,
				InstanceVariables: lcomResult.InstanceVariables,
				MethodGroups:      lcomResult.MethodGroups,
			},
			RiskLevel: domain.RiskLevel(lcomResult.RiskLevel),
		}
		classes = append(classes, class)
	}

	return classes
}

// filterClasses filters classes based on request criteria
func (s *LCOMServiceImpl) filterClasses(classes []domain.ClassCohesion, req domain.LCOMRequest) []domain.ClassCohesion {
	var filtered []domain.ClassCohesion

	for _, class := range classes {
		if class.Metrics.LCOM4 < req.MinLCOM {
			continue
		}
		if req.MaxLCOM > 0 && class.Metrics.LCOM4 > req.MaxLCOM {
			continue
		}
		filtered = append(filtered, class)
	}

	return filtered
}

// sortClasses sorts classes based on specified criteria
func (s *LCOMServiceImpl) sortClasses(classes []domain.ClassCohesion, sortBy domain.SortCriteria) []domain.ClassCohesion {
	sorted := make([]domain.ClassCohesion, len(classes))
	copy(sorted, classes)

	switch sortBy {
	case domain.SortByCohesion:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Metrics.LCOM4 > sorted[j].Metrics.LCOM4
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
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Metrics.LCOM4 > sorted[j].Metrics.LCOM4
		})
	}

	return sorted
}

// generateSummary creates aggregate LCOM statistics
func (s *LCOMServiceImpl) generateSummary(classes []domain.ClassCohesion, filesAnalyzed int, req domain.LCOMRequest) domain.LCOMSummary {
	if len(classes) == 0 {
		return domain.LCOMSummary{
			FilesAnalyzed: filesAnalyzed,
		}
	}

	summary := domain.LCOMSummary{
		TotalClasses:         len(classes),
		ClassesAnalyzed:      len(classes),
		FilesAnalyzed:        filesAnalyzed,
		LCOMDistribution:     make(map[string]int),
		LeastCohesiveClasses: []domain.ClassCohesion{},
	}

	totalLCOM := 0
	minLCOM := classes[0].Metrics.LCOM4
	maxLCOM := classes[0].Metrics.LCOM4

	for _, class := range classes {
		lcom := class.Metrics.LCOM4
		totalLCOM += lcom

		if lcom < minLCOM {
			minLCOM = lcom
		}
		if lcom > maxLCOM {
			maxLCOM = lcom
		}

		switch class.RiskLevel {
		case domain.RiskLevelLow:
			summary.LowRiskClasses++
		case domain.RiskLevelMedium:
			summary.MediumRiskClasses++
		case domain.RiskLevelHigh:
			summary.HighRiskClasses++
		}

		lcomRange := s.getLCOMRange(lcom)
		summary.LCOMDistribution[lcomRange]++
	}

	summary.AverageLCOM = float64(totalLCOM) / float64(len(classes))
	summary.MinLCOM = minLCOM
	summary.MaxLCOM = maxLCOM

	// Top 10 least cohesive classes
	sortedByLCOM := make([]domain.ClassCohesion, len(classes))
	copy(sortedByLCOM, classes)
	sort.Slice(sortedByLCOM, func(i, j int) bool {
		return sortedByLCOM[i].Metrics.LCOM4 > sortedByLCOM[j].Metrics.LCOM4
	})

	maxTopClasses := 10
	if len(sortedByLCOM) < maxTopClasses {
		maxTopClasses = len(sortedByLCOM)
	}
	summary.LeastCohesiveClasses = sortedByLCOM[:maxTopClasses]

	return summary
}

// getLCOMRange returns a range string for LCOM distribution
func (s *LCOMServiceImpl) getLCOMRange(lcom int) string {
	switch {
	case lcom <= 1:
		return "1"
	case lcom == 2:
		return "2"
	case lcom <= 5:
		return "3-5"
	case lcom <= 10:
		return "6-10"
	default:
		return "10+"
	}
}

// buildLCOMOptions converts domain request to analyzer options
func (s *LCOMServiceImpl) buildLCOMOptions(req domain.LCOMRequest) *analyzer.LCOMOptions {
	return &analyzer.LCOMOptions{
		LowThreshold:    req.LowThreshold,
		MediumThreshold: req.MediumThreshold,
	}
}

// buildConfigForResponse creates config info for response
func (s *LCOMServiceImpl) buildConfigForResponse(req domain.LCOMRequest) any {
	return map[string]any{
		"minLCOM":         req.MinLCOM,
		"maxLCOM":         req.MaxLCOM,
		"lowThreshold":    req.LowThreshold,
		"mediumThreshold": req.MediumThreshold,
		"outputFormat":    req.OutputFormat,
		"sortBy":          req.SortBy,
	}
}
