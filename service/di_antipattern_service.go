package service

import (
	"context"
	"fmt"
	"os"
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
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("DI anti-pattern analysis cancelled: %w", ctx.Err())
		default:
		}

		// Analyze single file
		fileFindings, fileWarnings, fileErrors := s.analyzeFile(ctx, filePath, req)

		if len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
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
func (s *DIAntipatternServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.DIAntipatternRequest) ([]domain.DIAntipatternFinding, []string, []string) {
	var findings []domain.DIAntipatternFinding
	var warnings []string
	var errors []string

	// Read the file
	content, err := s.readFile(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Failed to read file: %v", filePath, err))
		return findings, warnings, errors
	}

	// Parse the file
	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Parse error: %v", filePath, err))
		return findings, warnings, errors
	}

	// Configure DI anti-pattern detection options
	options := s.buildOptions(req)

	// Perform analysis
	fileFindings, err := analyzer.CalculateDIAntipatternsWithConfig(result.AST, filePath, options)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] DI anti-pattern analysis failed: %v", filePath, err))
		return findings, warnings, errors
	}

	findings = append(findings, fileFindings...)

	return findings, warnings, errors
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
