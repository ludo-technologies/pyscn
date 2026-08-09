package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// DIAntipatternServiceImpl implements the DIAntipatternService interface
type DIAntipatternServiceImpl struct {
	parser *parser.Parser
}

// NewDIAntipatternService creates a new DI anti-pattern service
func NewDIAntipatternService() *DIAntipatternServiceImpl {
	return &DIAntipatternServiceImpl{
		parser: parser.New(),
	}
}

// Analyze performs DI anti-pattern analysis on multiple files
func (s *DIAntipatternServiceImpl) Analyze(ctx context.Context, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	var allFindings []domain.DIAntipatternFinding
	var warnings []string
	var errors []string
	var diagnostics []domain.AnalysisDiagnostic
	var failures []domain.AnalysisFailure
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("DI anti-pattern analysis cancelled: %w", ctx.Err())
		default:
		}

		// Analyze single file
		fileFindings, fileWarnings, fileDiagnostics, fileFailures := s.analyzeFile(ctx, filePath, req)

		if len(fileDiagnostics) > 0 || len(fileFailures) > 0 {
			diagnostics = append(diagnostics, fileDiagnostics...)
			failures = append(failures, fileFailures...)
			errors = append(errors, diagnosticMessages(fileDiagnostics)...)
			errors = append(errors, failureMessages(fileFailures)...)
			continue
		}

		allFindings = append(allFindings, fileFindings...)
		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	// Sort findings
	sortedFindings := analyzer.SortFindings(allFindings, req.SortBy)

	// Generate summary
	summary := analyzer.GenerateSummary(sortedFindings, filesProcessed)

	return &domain.DIAntipatternResponse{
		Findings:    sortedFindings,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      errors,
		Diagnostics: diagnostics,
		Failures:    failures,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeSnapshot performs DI analysis from the canonical parsed project.
func (s *DIAntipatternServiceImpl) AnalyzeSnapshot(ctx context.Context, snapshot *ProjectSnapshot, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("project snapshot is required")
	}
	var findings []domain.DIAntipatternFinding
	var failures []domain.AnalysisFailure
	filesProcessed := 0
	for _, file := range snapshot.files {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("DI anti-pattern analysis cancelled: %w", err)
		}
		if !file.Parsed() || s.isTestFile(file.Path) {
			continue
		}
		fileFindings, err := s.calculateFindings(file.AST, file.Path, req)
		if err != nil {
			failures = append(failures, domain.AnalysisFailure{Analysis: domain.AnalysisKindDI, Code: domain.AnalysisFailureCodeExecution, FilePath: file.Path, Message: err.Error()})
			continue
		}
		findings = append(findings, fileFindings...)
		filesProcessed++
	}
	findings = analyzer.SortFindings(findings, req.SortBy)
	return &domain.DIAntipatternResponse{
		Findings:    findings,
		Summary:     analyzer.GenerateSummary(findings, filesProcessed),
		Errors:      failureMessages(failures),
		Failures:    failures,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file
func (s *DIAntipatternServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	singleFileReq := req
	singleFileReq.Paths = []string{filePath}
	return s.Analyze(ctx, singleFileReq)
}

// analyzeFile performs DI anti-pattern analysis on a single file
func (s *DIAntipatternServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.DIAntipatternRequest) ([]domain.DIAntipatternFinding, []string, []domain.AnalysisDiagnostic, []domain.AnalysisFailure) {
	var findings []domain.DIAntipatternFinding
	var warnings []string
	var diagnostics []domain.AnalysisDiagnostic
	var failures []domain.AnalysisFailure

	// Skip test files by default
	if s.isTestFile(filePath) {
		return findings, warnings, diagnostics, failures
	}

	// Read the file
	content, err := s.readFile(filePath)
	if err != nil {
		diagnostics = append(diagnostics, domain.AnalysisDiagnostic{FilePath: filePath, Code: domain.DiagnosticCodeRead, Message: err.Error()})
		return findings, warnings, diagnostics, failures
	}

	// Parse the file
	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		diagnostics = append(diagnostics, domain.AnalysisDiagnostic{FilePath: filePath, Code: domain.DiagnosticCodeParse, Message: err.Error()})
		return findings, warnings, diagnostics, failures
	}

	fileFindings, err := s.calculateFindings(result.AST, filePath, req)
	if err != nil {
		failures = append(failures, domain.AnalysisFailure{
			Analysis: domain.AnalysisKindDI,
			Code:     domain.AnalysisFailureCodeExecution,
			Message:  err.Error(),
			FilePath: filePath,
		})
		return findings, warnings, diagnostics, failures
	}

	findings = append(findings, fileFindings...)

	return findings, warnings, diagnostics, failures
}

func (s *DIAntipatternServiceImpl) calculateFindings(ast *parser.Node, filePath string, req domain.DIAntipatternRequest) ([]domain.DIAntipatternFinding, error) {
	return analyzer.CalculateDIAntipatternsWithConfig(ast, filePath, s.buildOptions(req))
}

// buildOptions converts domain request to analyzer options
func (s *DIAntipatternServiceImpl) buildOptions(req domain.DIAntipatternRequest) *analyzer.DIAntipatternOptions {
	threshold := req.ConstructorParamThreshold
	if threshold <= 0 {
		threshold = domain.DefaultDIConstructorParamThreshold
	}

	minSeverity := req.MinSeverity
	if minSeverity == "" {
		minSeverity = domain.DIAntipatternSeverityWarning
	}

	return &analyzer.DIAntipatternOptions{
		ConstructorParamThreshold: threshold,
		MinSeverity:               minSeverity,
	}
}

// buildConfigForResponse creates config info for response
func (s *DIAntipatternServiceImpl) buildConfigForResponse(req domain.DIAntipatternRequest) interface{} {
	return map[string]interface{}{
		"constructor_param_threshold": req.ConstructorParamThreshold,
		"min_severity":                req.MinSeverity,
		"sort_by":                     req.SortBy,
	}
}

// readFile reads file content
func (s *DIAntipatternServiceImpl) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// isTestFile checks if the file is a test file that should be skipped
func (s *DIAntipatternServiceImpl) isTestFile(filePath string) bool {
	baseName := filepath.Base(filePath)
	dir := filepath.Dir(filePath)

	// Check file name patterns: test_*.py, *_test.py
	if strings.HasPrefix(baseName, "test_") && strings.HasSuffix(baseName, ".py") {
		return true
	}
	if strings.HasSuffix(baseName, "_test.py") {
		return true
	}

	// Check special files: conftest.py
	if baseName == "conftest.py" {
		return true
	}

	// Check directory patterns: tests/, test/, testing/, __tests__/
	testDirs := []string{"tests", "test", "testing", "__tests__"}
	dirParts := strings.Split(dir, string(filepath.Separator))
	for _, part := range dirParts {
		for _, testDir := range testDirs {
			if part == testDir {
				return true
			}
		}
	}

	return false
}
