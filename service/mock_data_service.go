package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/mockdetector"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// MockDataServiceImpl implements the MockDataService interface
type MockDataServiceImpl struct {
	detector *mockdetector.Detector
}

// NewMockDataService creates a new mock data service implementation
func NewMockDataService() *MockDataServiceImpl {
	return &MockDataServiceImpl{
		detector: mockdetector.NewDetector(nil, nil), // Use defaults
	}
}

// NewMockDataServiceWithConfig creates a mock data service with custom configuration
func NewMockDataServiceWithConfig(keywords, domains []string) *MockDataServiceImpl {
	return &MockDataServiceImpl{
		detector: mockdetector.NewDetector(keywords, domains),
	}
}

// Analyze performs mock data analysis on multiple files
func (s *MockDataServiceImpl) Analyze(ctx context.Context, req domain.MockDataRequest) (*domain.MockDataResponse, error) {
	var allFiles []domain.FileMockData
	var warnings []string
	var errors []string
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("mock data analysis cancelled: %w", ctx.Err())
		default:
		}

		// Check if this is a test file and should be ignored
		if domain.BoolValue(req.IgnoreTests, domain.DefaultMockDataIgnoreTests) {
			if s.isTestFile(filePath) {
				continue
			}
		}

		// Analyze single file
		fileResult, fileWarnings, fileErrors := s.analyzeFile(ctx, filePath, req)

		if len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
			continue // Skip this file but continue with others
		}

		// Only include files that have findings
		if fileResult != nil && fileResult.HasFindings() {
			allFiles = append(allFiles, *fileResult)
		}

		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	// Filter and sort results
	filteredFiles := s.filterFiles(allFiles, req)
	sortedFiles := s.sortFiles(filteredFiles, req.SortBy)

	// Generate summary
	summary := s.generateSummary(sortedFiles, filesProcessed)

	return &domain.MockDataResponse{
		Files:       sortedFiles,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      errors,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file for mock data
func (s *MockDataServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.MockDataRequest) (*domain.FileMockData, error) {
	fileResult, _, fileErrors := s.analyzeFile(ctx, filePath, req)

	if len(fileErrors) > 0 {
		return nil, domain.NewAnalysisError(fmt.Sprintf("failed to analyze file %s", filePath), fmt.Errorf("%v", fileErrors))
	}

	return fileResult, nil
}

// analyzeFile performs mock data analysis on a single file
func (s *MockDataServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.MockDataRequest) (*domain.FileMockData, []string, []string) {
	var warnings []string
	var errors []string

	// Read the file
	content, err := s.readFile(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Failed to read file: %v", filePath, err))
		return nil, warnings, errors
	}

	// Detect mock data
	result, err := s.detector.Detect(ctx, content, filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Detection error: %v", filePath, err))
		return nil, warnings, errors
	}

	// Update file path in findings
	for i := range result.Findings {
		result.Findings[i].Location.FilePath = filePath
	}

	// Filter findings by enabled types
	filteredFindings := s.filterByType(result.Findings, req.EnabledTypes)

	// Create file result
	fileResult := &domain.FileMockData{
		FilePath: filePath,
		Findings: filteredFindings,
	}
	fileResult.CalculateSeverityCounts()

	return fileResult, warnings, errors
}

// readFile reads the content of a file
func (s *MockDataServiceImpl) readFile(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return content, nil
}

// isTestFile checks if a file path matches test file patterns
func (s *MockDataServiceImpl) isTestFile(filePath string) bool {
	base := filepath.Base(filePath)
	dir := filepath.Dir(filePath)

	// Check filename patterns
	if strings.HasPrefix(base, "test_") || strings.HasSuffix(base, "_test.py") {
		return true
	}

	// Check if in test directory
	testDirs := []string{"tests", "test", "testing", "__tests__"}
	parts := strings.Split(dir, string(filepath.Separator))
	for _, part := range parts {
		for _, testDir := range testDirs {
			if part == testDir {
				return true
			}
		}
	}

	// Check for conftest.py
	if base == "conftest.py" {
		return true
	}

	return false
}

// filterByType filters findings to only include specified types
func (s *MockDataServiceImpl) filterByType(findings []domain.MockDataFinding, enabledTypes []domain.MockDataType) []domain.MockDataFinding {
	// If no types specified, include all
	if len(enabledTypes) == 0 {
		return findings
	}

	typeSet := make(map[domain.MockDataType]bool)
	for _, t := range enabledTypes {
		typeSet[t] = true
	}

	var filtered []domain.MockDataFinding
	for _, finding := range findings {
		if typeSet[finding.Type] {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// filterFiles filters files based on minimum severity
func (s *MockDataServiceImpl) filterFiles(files []domain.FileMockData, req domain.MockDataRequest) []domain.FileMockData {
	var filtered []domain.FileMockData

	for _, file := range files {
		filteredFindings := file.GetFindingsAtSeverity(req.MinSeverity)
		if len(filteredFindings) > 0 {
			fileCopy := file
			fileCopy.Findings = filteredFindings
			fileCopy.CalculateSeverityCounts()
			filtered = append(filtered, fileCopy)
		}
	}

	return filtered
}

// sortFiles sorts files based on the specified criteria
func (s *MockDataServiceImpl) sortFiles(files []domain.FileMockData, sortBy domain.MockDataSortCriteria) []domain.FileMockData {
	sorted := make([]domain.FileMockData, len(files))
	copy(sorted, files)

	switch sortBy {
	case domain.MockDataSortByFile:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].FilePath < sorted[j].FilePath
		})
	case domain.MockDataSortBySeverity:
		sort.Slice(sorted, func(i, j int) bool {
			// Sort by highest severity finding (descending)
			maxI := s.getMaxSeverity(sorted[i].Findings)
			maxJ := s.getMaxSeverity(sorted[j].Findings)
			if maxI != maxJ {
				return maxI > maxJ
			}
			return sorted[i].TotalFindings > sorted[j].TotalFindings
		})
	case domain.MockDataSortByLine:
		// Sort findings within each file by line
		for i := range sorted {
			sort.Slice(sorted[i].Findings, func(a, b int) bool {
				return sorted[i].Findings[a].Location.StartLine < sorted[i].Findings[b].Location.StartLine
			})
		}
	case domain.MockDataSortByType:
		sort.Slice(sorted, func(i, j int) bool {
			if len(sorted[i].Findings) == 0 || len(sorted[j].Findings) == 0 {
				return len(sorted[i].Findings) > len(sorted[j].Findings)
			}
			return sorted[i].Findings[0].Type < sorted[j].Findings[0].Type
		})
	}

	return sorted
}

// getMaxSeverity returns the maximum severity level in a list of findings
func (s *MockDataServiceImpl) getMaxSeverity(findings []domain.MockDataFinding) int {
	maxLevel := 0
	for _, f := range findings {
		if level := f.Severity.Level(); level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}

// generateSummary generates aggregate statistics
func (s *MockDataServiceImpl) generateSummary(files []domain.FileMockData, filesProcessed int) domain.MockDataSummary {
	summary := domain.MockDataSummary{
		TotalFiles:        filesProcessed,
		FilesWithMockData: len(files),
		FindingsByType:    make(map[domain.MockDataType]int),
	}

	for _, file := range files {
		summary.TotalFindings += file.TotalFindings
		summary.ErrorFindings += file.ErrorCount
		summary.WarningFindings += file.WarningCount
		summary.InfoFindings += file.InfoCount

		for _, finding := range file.Findings {
			summary.FindingsByType[finding.Type]++
		}
	}

	return summary
}

// buildConfigForResponse creates a config representation for the response
func (s *MockDataServiceImpl) buildConfigForResponse(req domain.MockDataRequest) map[string]interface{} {
	return map[string]interface{}{
		"min_severity": string(req.MinSeverity),
		"sort_by":      string(req.SortBy),
		"ignore_tests": domain.BoolValue(req.IgnoreTests, domain.DefaultMockDataIgnoreTests),
		"keywords":     req.Keywords,
		"domains":      req.Domains,
	}
}
