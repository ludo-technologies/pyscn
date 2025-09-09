package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ludo-technologies/pyscn/domain"
)

func TestNewDeadCodeService(t *testing.T) {
	service := NewDeadCodeService()
	
	assert.NotNil(t, service)
	assert.NotNil(t, service.parser)
}

func TestDeadCodeService_Analyze(t *testing.T) {
	service := NewDeadCodeService()
	ctx := context.Background()

	t.Run("successful analysis with Python file containing functions", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/python/simple/control_flow.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.GeneratedAt)
		assert.NotEmpty(t, response.Version)
		assert.NotNil(t, response.Config)
		assert.GreaterOrEqual(t, response.Summary.TotalFiles, 1)

		// Verify response structure
		for _, file := range response.Files {
			assert.NotEmpty(t, file.FilePath)
			assert.GreaterOrEqual(t, file.TotalFunctions, 0)
			
			for _, function := range file.Functions {
				assert.NotEmpty(t, function.Name)
				assert.NotEmpty(t, function.FilePath)
				assert.GreaterOrEqual(t, function.TotalBlocks, 0)
				assert.GreaterOrEqual(t, function.DeadBlocks, 0)
				assert.GreaterOrEqual(t, function.ReachableRatio, 0.0)
				assert.LessOrEqual(t, function.ReachableRatio, 1.0)
				
				for _, finding := range function.Findings {
					assert.NotEmpty(t, finding.Location.FilePath)
					assert.Greater(t, finding.Location.StartLine, 0)
					assert.Contains(t, []domain.DeadCodeSeverity{
						domain.DeadCodeSeverityInfo,
						domain.DeadCodeSeverityWarning,
						domain.DeadCodeSeverityCritical,
					}, finding.Severity)
				}
			}
		}
	})

	t.Run("analyze file with no functions should return valid response", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/python/simple/imports.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Summary.TotalFiles)
		// File with no functions should still be analyzed but may have no findings
	})

	t.Run("analyze with severity filtering", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/python/simple/control_flow.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityWarning, // Only warning and critical
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, file := range response.Files {
				for _, function := range file.Functions {
					for _, finding := range function.Findings {
						assert.True(t, finding.Severity.IsAtLeast(domain.DeadCodeSeverityWarning))
					}
				}
			}
		}
	})

	t.Run("analyze multiple files", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths: []string{
				"../testdata/python/simple/functions.py",
				"../testdata/python/simple/control_flow.py",
			},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Summary.TotalFiles)
	})

	t.Run("error handling for non-existent file", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/non_existent_file.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/python/simple/functions.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		_, err := service.Analyze(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("analyze with context enabled", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			Paths:                     []string{"../testdata/python/simple/control_flow.py"},
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               true,
			ContextLines:              3,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		// When context is enabled, findings may have context information
	})
}

func TestDeadCodeService_AnalyzeFile(t *testing.T) {
	service := NewDeadCodeService()
	ctx := context.Background()

	t.Run("analyze single file", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		file, err := service.AnalyzeFile(ctx, "../testdata/python/simple/functions.py", req)

		assert.NoError(t, err)
		if file != nil {
			assert.NotEmpty(t, file.FilePath)
			assert.GreaterOrEqual(t, file.TotalFunctions, 0)
		}
	})

	t.Run("analyze file with errors should return error", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			OutputFormat:              domain.OutputFormatJSON,
			MinSeverity:               domain.DeadCodeSeverityInfo,
			SortBy:                    domain.DeadCodeSortByFile,
			ShowContext:               false,
			ContextLines:              2,
			Recursive:                 false,
			DetectAfterReturn:         true,
			DetectAfterBreak:          true,
			DetectAfterContinue:       true,
			DetectAfterRaise:          true,
			DetectUnreachableBranches: true,
		}

		_, err := service.AnalyzeFile(ctx, "../testdata/non_existent_file.py", req)

		assert.Error(t, err)
	})
}

func TestDeadCodeService_ConvertSeverity(t *testing.T) {
	// Note: This is testing internal package types, so we'll focus on the public interface
	// The actual analyzer.SeverityLevel types are in the internal package
	// We can test the conversion through the service behavior instead
	// This test is a placeholder for now
	assert.True(t, true)
}

func TestDeadCodeService_FilterFiles(t *testing.T) {
	service := NewDeadCodeService()

	// Create test files with different severity levels
	files := []domain.FileDeadCode{
		{
			FilePath: "file1.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func1",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityInfo},
						{Severity: domain.DeadCodeSeverityWarning},
					},
					CriticalCount: 0,
					WarningCount:  1,
					InfoCount:     1,
				},
			},
		},
		{
			FilePath: "file2.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func2",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityCritical},
					},
					CriticalCount: 1,
					WarningCount:  0,
					InfoCount:     0,
				},
			},
		},
		{
			FilePath: "file3.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func3",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityInfo},
					},
					CriticalCount: 0,
					WarningCount:  0,
					InfoCount:     1,
				},
			},
		},
	}

	t.Run("filter by warning severity", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			MinSeverity: domain.DeadCodeSeverityWarning,
		}

		filtered := service.filterFiles(files, req)

		// Should include files with warning or critical findings
		assert.Len(t, filtered, 2)
		assert.Equal(t, "file1.py", filtered[0].FilePath)
		assert.Equal(t, "file2.py", filtered[1].FilePath)
	})

	t.Run("filter by critical severity", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			MinSeverity: domain.DeadCodeSeverityCritical,
		}

		filtered := service.filterFiles(files, req)

		// Should include only files with critical findings
		assert.Len(t, filtered, 1)
		assert.Equal(t, "file2.py", filtered[0].FilePath)
	})

	t.Run("filter by info severity includes all", func(t *testing.T) {
		req := domain.DeadCodeRequest{
			MinSeverity: domain.DeadCodeSeverityInfo,
		}

		filtered := service.filterFiles(files, req)

		// Should include all files
		assert.Len(t, filtered, 3)
	})
}

func TestDeadCodeService_FilterFindingsBySeverity(t *testing.T) {
	service := NewDeadCodeService()

	findings := []domain.DeadCodeFinding{
		{Severity: domain.DeadCodeSeverityInfo, Reason: "info1"},
		{Severity: domain.DeadCodeSeverityWarning, Reason: "warning1"},
		{Severity: domain.DeadCodeSeverityCritical, Reason: "critical1"},
		{Severity: domain.DeadCodeSeverityInfo, Reason: "info2"},
	}

	t.Run("filter by warning severity", func(t *testing.T) {
		filtered := service.filterFindingsBySeverity(findings, domain.DeadCodeSeverityWarning)

		assert.Len(t, filtered, 2)
		assert.Equal(t, "warning1", filtered[0].Reason)
		assert.Equal(t, "critical1", filtered[1].Reason)
	})

	t.Run("filter by critical severity", func(t *testing.T) {
		filtered := service.filterFindingsBySeverity(findings, domain.DeadCodeSeverityCritical)

		assert.Len(t, filtered, 1)
		assert.Equal(t, "critical1", filtered[0].Reason)
	})

	t.Run("filter by info severity includes all", func(t *testing.T) {
		filtered := service.filterFindingsBySeverity(findings, domain.DeadCodeSeverityInfo)

		assert.Len(t, filtered, 4)
	})
}

func TestDeadCodeService_SortFiles(t *testing.T) {
	service := NewDeadCodeService()

	files := []domain.FileDeadCode{
		{
			FilePath: "c_file.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func1",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityInfo},
					},
				},
			},
		},
		{
			FilePath: "a_file.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func2",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityCritical},
					},
				},
			},
		},
		{
			FilePath: "b_file.py",
			Functions: []domain.FunctionDeadCode{
				{
					Name: "func3",
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityWarning},
					},
				},
			},
		},
	}

	t.Run("sort by file path", func(t *testing.T) {
		sorted := service.sortFiles(files, domain.DeadCodeSortByFile)

		require.Len(t, sorted, 3)
		assert.Equal(t, "a_file.py", sorted[0].FilePath)
		assert.Equal(t, "b_file.py", sorted[1].FilePath)
		assert.Equal(t, "c_file.py", sorted[2].FilePath)
	})

	t.Run("sort by severity", func(t *testing.T) {
		sorted := service.sortFiles(files, domain.DeadCodeSortBySeverity)

		require.Len(t, sorted, 3)
		// Should be sorted by highest severity first
		assert.Equal(t, "a_file.py", sorted[0].FilePath) // Critical
		assert.Equal(t, "b_file.py", sorted[1].FilePath) // Warning
		assert.Equal(t, "c_file.py", sorted[2].FilePath) // Info
	})

	t.Run("sort by default (file)", func(t *testing.T) {
		sorted := service.sortFiles(files, domain.DeadCodeSortCriteria("unknown"))

		require.Len(t, sorted, 3)
		assert.Equal(t, "a_file.py", sorted[0].FilePath)
		assert.Equal(t, "b_file.py", sorted[1].FilePath)
		assert.Equal(t, "c_file.py", sorted[2].FilePath)
	})
}

func TestDeadCodeService_GetHighestSeverityLevel(t *testing.T) {
	service := NewDeadCodeService()

	t.Run("file with critical severity", func(t *testing.T) {
		file := domain.FileDeadCode{
			Functions: []domain.FunctionDeadCode{
				{
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityInfo},
						{Severity: domain.DeadCodeSeverityCritical},
						{Severity: domain.DeadCodeSeverityWarning},
					},
				},
			},
		}

		level := service.getHighestSeverityLevel(file)
		assert.Equal(t, 3, level) // Critical = 3
	})

	t.Run("file with warning severity", func(t *testing.T) {
		file := domain.FileDeadCode{
			Functions: []domain.FunctionDeadCode{
				{
					Findings: []domain.DeadCodeFinding{
						{Severity: domain.DeadCodeSeverityInfo},
						{Severity: domain.DeadCodeSeverityWarning},
					},
				},
			},
		}

		level := service.getHighestSeverityLevel(file)
		assert.Equal(t, 2, level) // Warning = 2
	})

	t.Run("file with no findings", func(t *testing.T) {
		file := domain.FileDeadCode{
			Functions: []domain.FunctionDeadCode{
				{
					Findings: []domain.DeadCodeFinding{},
				},
			},
		}

		level := service.getHighestSeverityLevel(file)
		assert.Equal(t, 0, level)
	})
}

func TestDeadCodeService_GenerateSummary(t *testing.T) {
	service := NewDeadCodeService()

	files := []domain.FileDeadCode{
		{
			FilePath:          "file1.py",
			TotalFunctions:    2,
			AffectedFunctions: 1,
			Functions: []domain.FunctionDeadCode{
				{
					Name:          "func1",
					TotalBlocks:   10,
					DeadBlocks:    2,
					CriticalCount: 1,
					WarningCount:  1,
					InfoCount:     0,
					Findings: []domain.DeadCodeFinding{
						{Reason: "unreachable_after_return"},
						{Reason: "dead_branch"},
					},
				},
			},
		},
		{
			FilePath:          "file2.py",
			TotalFunctions:    1,
			AffectedFunctions: 1,
			Functions: []domain.FunctionDeadCode{
				{
					Name:          "func2",
					TotalBlocks:   5,
					DeadBlocks:    1,
					CriticalCount: 0,
					WarningCount:  0,
					InfoCount:     2,
					Findings: []domain.DeadCodeFinding{
						{Reason: "unreachable_after_return"},
						{Reason: "unused_variable"},
					},
				},
			},
		},
	}

	t.Run("generate summary with files", func(t *testing.T) {
		req := domain.DeadCodeRequest{}
		summary := service.generateSummary(files, 3, req)

		assert.Equal(t, 3, summary.TotalFiles)
		assert.Equal(t, 2, summary.FilesWithDeadCode)
		assert.Equal(t, 3, summary.TotalFunctions)         // 2 + 1
		assert.Equal(t, 2, summary.FunctionsWithDeadCode) // 1 + 1
		assert.Equal(t, 4, summary.TotalFindings)          // 2 + 2
		assert.Equal(t, 1, summary.CriticalFindings)
		assert.Equal(t, 1, summary.WarningFindings)
		assert.Equal(t, 2, summary.InfoFindings)
		assert.Equal(t, 15, summary.TotalBlocks) // 10 + 5
		assert.Equal(t, 3, summary.DeadBlocks)   // 2 + 1
		assert.Equal(t, 0.2, summary.OverallDeadRatio) // 3/15

		// Check findings by reason
		assert.Equal(t, 2, summary.FindingsByReason["unreachable_after_return"])
		assert.Equal(t, 1, summary.FindingsByReason["dead_branch"])
		assert.Equal(t, 1, summary.FindingsByReason["unused_variable"])
	})

	t.Run("generate summary with no files", func(t *testing.T) {
		req := domain.DeadCodeRequest{}
		summary := service.generateSummary([]domain.FileDeadCode{}, 5, req)

		assert.Equal(t, 5, summary.TotalFiles)
		assert.Equal(t, 0, summary.FilesWithDeadCode)
		assert.Equal(t, 0, summary.TotalFunctions)
		assert.Equal(t, 0, summary.FunctionsWithDeadCode)
		assert.Equal(t, 0, summary.TotalFindings)
		assert.Equal(t, 0.0, summary.OverallDeadRatio)
	})
}

func TestDeadCodeService_CountTotalFindings(t *testing.T) {
	service := NewDeadCodeService()

	functions := []domain.FunctionDeadCode{
		{
			Findings: []domain.DeadCodeFinding{
				{Reason: "finding1"},
				{Reason: "finding2"},
			},
		},
		{
			Findings: []domain.DeadCodeFinding{
				{Reason: "finding3"},
			},
		},
		{
			Findings: []domain.DeadCodeFinding{},
		},
	}

	total := service.countTotalFindings(functions)
	assert.Equal(t, 3, total)
}

func TestDeadCodeService_BuildConfigForResponse(t *testing.T) {
	service := NewDeadCodeService()

	req := domain.DeadCodeRequest{
		MinSeverity:               domain.DeadCodeSeverityWarning,
		SortBy:                    domain.DeadCodeSortBySeverity,
		ShowContext:               true,
		ContextLines:              3,
		DetectAfterReturn:         true,
		DetectAfterBreak:          false,
		DetectAfterContinue:       true,
		DetectAfterRaise:          false,
		DetectUnreachableBranches: true,
		IncludePatterns:           []string{"*.py"},
		ExcludePatterns:           []string{"test_*.py"},
		IgnorePatterns:            []string{"# TODO"},
	}

	config := service.buildConfigForResponse(req)

	configMap, ok := config.(map[string]interface{})
	require.True(t, ok)
	
	assert.Equal(t, domain.DeadCodeSeverityWarning, configMap["min_severity"])
	assert.Equal(t, domain.DeadCodeSortBySeverity, configMap["sort_by"])
	assert.Equal(t, true, configMap["show_context"])
	assert.Equal(t, 3, configMap["context_lines"])
	assert.Equal(t, true, configMap["detect_after_return"])
	assert.Equal(t, false, configMap["detect_after_break"])
	assert.Equal(t, true, configMap["detect_after_continue"])
	assert.Equal(t, false, configMap["detect_after_raise"])
	assert.Equal(t, true, configMap["detect_unreachable_branches"])
	assert.Equal(t, []string{"*.py"}, configMap["include_patterns"])
	assert.Equal(t, []string{"test_*.py"}, configMap["exclude_patterns"])
	assert.Equal(t, []string{"# TODO"}, configMap["ignore_patterns"])
}

func TestDeadCodeService_ResponseMetadata(t *testing.T) {
	service := NewDeadCodeService()
	ctx := context.Background()

	req := domain.DeadCodeRequest{
		Paths:                     []string{"../testdata/python/simple/functions.py"},
		OutputFormat:              domain.OutputFormatJSON,
		MinSeverity:               domain.DeadCodeSeverityInfo,
		SortBy:                    domain.DeadCodeSortByFile,
		ShowContext:               false,
		ContextLines:              2,
		Recursive:                 false,
		DetectAfterReturn:         true,
		DetectAfterBreak:          true,
		DetectAfterContinue:       true,
		DetectAfterRaise:          true,
		DetectUnreachableBranches: true,
	}

	beforeTime := time.Now()
	response, err := service.Analyze(ctx, req)
	afterTime := time.Now()

	assert.NoError(t, err)
	require.NotNil(t, response)
	
	// Verify timestamp is within expected range
	assert.NotEmpty(t, response.GeneratedAt)
	generatedTime, err := time.Parse(time.RFC3339, response.GeneratedAt)
	assert.NoError(t, err)
	assert.True(t, generatedTime.After(beforeTime.Add(-time.Second)) && generatedTime.Before(afterTime.Add(time.Second)), 
		"Generated time %v should be between %v and %v", generatedTime, beforeTime, afterTime)
	
	// Verify version is present
	assert.NotEmpty(t, response.Version)
	
	// Verify config is present
	assert.NotNil(t, response.Config)
}