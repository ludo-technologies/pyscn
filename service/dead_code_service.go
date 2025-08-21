package service

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/analyzer"
	"github.com/pyqol/pyqol/internal/parser"
	"github.com/pyqol/pyqol/internal/version"
)

// DeadCodeServiceImpl implements the DeadCodeService interface
type DeadCodeServiceImpl struct {
	parser   *parser.Parser
	progress domain.ProgressReporter
}

// NewDeadCodeService creates a new dead code service implementation
func NewDeadCodeService(progress domain.ProgressReporter) *DeadCodeServiceImpl {
	return &DeadCodeServiceImpl{
		parser:   parser.New(),
		progress: progress,
	}
}

// Analyze performs dead code analysis on multiple files
func (s *DeadCodeServiceImpl) Analyze(ctx context.Context, req domain.DeadCodeRequest) (*domain.DeadCodeResponse, error) {
	var allFiles []domain.FileDeadCode
	var warnings []string
	var errors []string
	filesProcessed := 0

	for i, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Update progress
		if s.progress != nil {
			s.progress.UpdateProgress(filePath, i, len(req.Paths))
		}

		// Analyze single file
		fileResult, fileWarnings, fileErrors := s.analyzeFile(ctx, filePath, req)

		if len(fileErrors) > 0 {
			errors = append(errors, fileErrors...)
			continue // Skip this file but continue with others
		}

		// Only include files that have dead code findings
		if fileResult != nil && (len(fileResult.Functions) > 0 || fileResult.TotalFindings > 0) {
			allFiles = append(allFiles, *fileResult)
		}

		warnings = append(warnings, fileWarnings...)
		filesProcessed++
	}

	// Filter and sort results
	filteredFiles := s.filterFiles(allFiles, req)
	sortedFiles := s.sortFiles(filteredFiles, req.SortBy)

	// Generate summary
	summary := s.generateSummary(sortedFiles, filesProcessed, req)

	return &domain.DeadCodeResponse{
		Files:       sortedFiles,
		Summary:     summary,
		Warnings:    warnings,
		Errors:      errors,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     version.Version,
		Config:      s.buildConfigForResponse(req),
	}, nil
}

// AnalyzeFile analyzes a single Python file for dead code
func (s *DeadCodeServiceImpl) AnalyzeFile(ctx context.Context, filePath string, req domain.DeadCodeRequest) (*domain.FileDeadCode, error) {
	fileResult, _, fileErrors := s.analyzeFile(ctx, filePath, req)

	if len(fileErrors) > 0 {
		return nil, domain.NewAnalysisError(fmt.Sprintf("failed to analyze file %s", filePath), fmt.Errorf("%v", fileErrors))
	}

	return fileResult, nil
}

// AnalyzeFunction analyzes a single function for dead code
func (s *DeadCodeServiceImpl) AnalyzeFunction(ctx context.Context, functionCFG interface{}, req domain.DeadCodeRequest) (*domain.FunctionDeadCode, error) {
	cfg, ok := functionCFG.(*analyzer.CFG)
	if !ok {
		return nil, domain.NewInvalidInputError("invalid CFG type", nil)
	}

	result := analyzer.DetectInFunction(cfg)
	if result == nil {
		return nil, domain.NewAnalysisError("failed to analyze function", nil)
	}

	funcResult := s.convertToFunctionDeadCode(result, req)
	return &funcResult, nil
}

// analyzeFile performs dead code analysis on a single file
func (s *DeadCodeServiceImpl) analyzeFile(ctx context.Context, filePath string, req domain.DeadCodeRequest) (*domain.FileDeadCode, []string, []string) {
	var warnings []string
	var errors []string

	// Parse the file
	content, err := s.readFile(filePath)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Failed to read file: %v", filePath, err))
		return nil, warnings, errors
	}

	result, err := s.parser.Parse(ctx, content)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] Parse error: %v", filePath, err))
		return nil, warnings, errors
	}

	// Build CFGs for all functions
	builder := analyzer.NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	if err != nil {
		errors = append(errors, fmt.Sprintf("[%s] CFG construction failed: %v", filePath, err))
		return nil, warnings, errors
	}

	if len(cfgs) == 0 {
		warnings = append(warnings, fmt.Sprintf("[%s] No functions found in file", filePath))
		return &domain.FileDeadCode{
			FilePath:          filePath,
			Functions:         []domain.FunctionDeadCode{},
			TotalFindings:     0,
			TotalFunctions:    0,
			AffectedFunctions: 0,
			DeadCodeRatio:     0.0,
		}, warnings, errors
	}

	// Analyze dead code for each function
	var functions []domain.FunctionDeadCode
	totalFindings := 0
	affectedFunctions := 0

	for functionName, cfg := range cfgs {
		// Skip the main module CFG for now, focus on functions
		if functionName == "__main__" {
			continue
		}

		deadCodeResults := analyzer.DetectInFunction(cfg)
		if deadCodeResults == nil {
			warnings = append(warnings, fmt.Sprintf("[%s:%s] Failed to analyze dead code for function", filePath, functionName))
			continue
		}

		functionResult := s.convertToFunctionDeadCode(deadCodeResults, req)
		functionResult.Name = functionName
		functionResult.FilePath = filePath

		// Apply severity filtering
		filteredFindings := s.filterFindingsBySeverity(functionResult.Findings, req.MinSeverity)
		functionResult.Findings = filteredFindings

		// Only include functions that have findings after filtering
		if len(functionResult.Findings) > 0 {
			functions = append(functions, functionResult)
			totalFindings += len(functionResult.Findings)
			affectedFunctions++
		}
	}

	// Calculate file-level metrics
	deadCodeRatio := 0.0
	if len(cfgs) > 0 {
		totalDeadBlocks := 0
		totalBlocks := 0
		for _, function := range functions {
			totalDeadBlocks += function.DeadBlocks
			totalBlocks += function.TotalBlocks
		}
		if totalBlocks > 0 {
			deadCodeRatio = float64(totalDeadBlocks) / float64(totalBlocks)
		}
	}

	fileResult := &domain.FileDeadCode{
		FilePath:          filePath,
		Functions:         functions,
		TotalFindings:     totalFindings,
		TotalFunctions:    len(cfgs) - 1, // Exclude __main__
		AffectedFunctions: affectedFunctions,
		DeadCodeRatio:     deadCodeRatio,
	}

	return fileResult, warnings, errors
}

// convertToFunctionDeadCode converts analyzer results to domain model
func (s *DeadCodeServiceImpl) convertToFunctionDeadCode(result *analyzer.DeadCodeResult, req domain.DeadCodeRequest) domain.FunctionDeadCode {
	var findings []domain.DeadCodeFinding

	for _, analyzerFinding := range result.Findings {
		finding := domain.DeadCodeFinding{
			Location: domain.DeadCodeLocation{
				FilePath:  analyzerFinding.FilePath,
				StartLine: analyzerFinding.StartLine,
				EndLine:   analyzerFinding.EndLine,
			},
			FunctionName: analyzerFinding.FunctionName,
			Code:         analyzerFinding.Code,
			Reason:       string(analyzerFinding.Reason),
			Severity:     s.convertSeverity(analyzerFinding.Severity),
			Description:  analyzerFinding.Description,
			Context:      analyzerFinding.Context,
			BlockID:      analyzerFinding.BlockID,
		}
		findings = append(findings, finding)
	}

	functionResult := domain.FunctionDeadCode{
		Name:           result.FunctionName,
		FilePath:       result.FilePath,
		Findings:       findings,
		TotalBlocks:    result.TotalBlocks,
		DeadBlocks:     result.DeadBlocks,
		ReachableRatio: result.ReachableRatio,
	}

	// Calculate severity counts
	functionResult.CalculateSeverityCounts()

	return functionResult
}

// convertSeverity converts analyzer severity to domain severity
func (s *DeadCodeServiceImpl) convertSeverity(analyzerSeverity analyzer.SeverityLevel) domain.DeadCodeSeverity {
	switch analyzerSeverity {
	case analyzer.SeverityLevelCritical:
		return domain.DeadCodeSeverityCritical
	case analyzer.SeverityLevelWarning:
		return domain.DeadCodeSeverityWarning
	case analyzer.SeverityLevelInfo:
		return domain.DeadCodeSeverityInfo
	default:
		return domain.DeadCodeSeverityWarning
	}
}

// filterFiles filters files based on request criteria
func (s *DeadCodeServiceImpl) filterFiles(files []domain.FileDeadCode, req domain.DeadCodeRequest) []domain.FileDeadCode {
	var filtered []domain.FileDeadCode

	for _, file := range files {
		// Filter functions within each file
		var filteredFunctions []domain.FunctionDeadCode
		for _, function := range file.Functions {
			if function.HasFindingsAtSeverity(req.MinSeverity) {
				filteredFunctions = append(filteredFunctions, function)
			}
		}

		// Only include files that have functions with findings after filtering
		if len(filteredFunctions) > 0 {
			filteredFile := file
			filteredFile.Functions = filteredFunctions
			filteredFile.TotalFindings = s.countTotalFindings(filteredFunctions)
			filteredFile.AffectedFunctions = len(filteredFunctions)
			filtered = append(filtered, filteredFile)
		}
	}

	return filtered
}

// filterFindingsBySeverity filters findings by minimum severity level
func (s *DeadCodeServiceImpl) filterFindingsBySeverity(findings []domain.DeadCodeFinding, minSeverity domain.DeadCodeSeverity) []domain.DeadCodeFinding {
	var filtered []domain.DeadCodeFinding
	for _, finding := range findings {
		if finding.Severity.IsAtLeast(minSeverity) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// sortFiles sorts files based on the specified criteria
func (s *DeadCodeServiceImpl) sortFiles(files []domain.FileDeadCode, sortBy domain.DeadCodeSortCriteria) []domain.FileDeadCode {
	sort.Slice(files, func(i, j int) bool {
		switch sortBy {
		case domain.DeadCodeSortByFile:
			return files[i].FilePath < files[j].FilePath
		case domain.DeadCodeSortBySeverity:
			// Sort by highest severity findings first
			return s.getHighestSeverityLevel(files[i]) > s.getHighestSeverityLevel(files[j])
		default:
			return files[i].FilePath < files[j].FilePath
		}
	})
	return files
}

// getHighestSeverityLevel gets the highest severity level in a file
func (s *DeadCodeServiceImpl) getHighestSeverityLevel(file domain.FileDeadCode) int {
	maxLevel := 0
	for _, function := range file.Functions {
		for _, finding := range function.Findings {
			if level := finding.Severity.Level(); level > maxLevel {
				maxLevel = level
			}
		}
	}
	return maxLevel
}

// generateSummary generates aggregate statistics
func (s *DeadCodeServiceImpl) generateSummary(files []domain.FileDeadCode, filesProcessed int, req domain.DeadCodeRequest) domain.DeadCodeSummary {
	summary := domain.DeadCodeSummary{
		TotalFiles:        filesProcessed,
		FilesWithDeadCode: len(files),
		FindingsByReason:  make(map[string]int),
		TotalBlocks:       0,
		DeadBlocks:        0,
	}

	for _, file := range files {
		summary.TotalFunctions += file.TotalFunctions
		summary.FunctionsWithDeadCode += file.AffectedFunctions

		for _, function := range file.Functions {
			summary.TotalFindings += len(function.Findings)
			summary.CriticalFindings += function.CriticalCount
			summary.WarningFindings += function.WarningCount
			summary.InfoFindings += function.InfoCount
			summary.TotalBlocks += function.TotalBlocks
			summary.DeadBlocks += function.DeadBlocks

			// Count findings by reason
			for _, finding := range function.Findings {
				summary.FindingsByReason[finding.Reason]++
			}
		}
	}

	// Calculate overall dead code ratio
	if summary.TotalBlocks > 0 {
		summary.OverallDeadRatio = float64(summary.DeadBlocks) / float64(summary.TotalBlocks)
	}

	return summary
}

// countTotalFindings counts total findings in functions
func (s *DeadCodeServiceImpl) countTotalFindings(functions []domain.FunctionDeadCode) int {
	total := 0
	for _, function := range functions {
		total += len(function.Findings)
	}
	return total
}

// readFile reads a file and returns its content
func (s *DeadCodeServiceImpl) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// buildConfigForResponse builds configuration for response metadata
func (s *DeadCodeServiceImpl) buildConfigForResponse(req domain.DeadCodeRequest) interface{} {
	return map[string]interface{}{
		"min_severity":                req.MinSeverity,
		"sort_by":                     req.SortBy,
		"show_context":                req.ShowContext,
		"context_lines":               req.ContextLines,
		"detect_after_return":         req.DetectAfterReturn,
		"detect_after_break":          req.DetectAfterBreak,
		"detect_after_continue":       req.DetectAfterContinue,
		"detect_after_raise":          req.DetectAfterRaise,
		"detect_unreachable_branches": req.DetectUnreachableBranches,
		"include_patterns":            req.IncludePatterns,
		"exclude_patterns":            req.ExcludePatterns,
		"ignore_patterns":             req.IgnorePatterns,
	}
}
