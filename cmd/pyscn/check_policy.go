package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"

	"github.com/spf13/cobra"
)

func (c *CheckCommand) runCoreAnalysis(
	cmd *cobra.Command,
	paths []string,
	skipComplexity bool,
	skipDeadCode bool,
	skipClones bool,
	skipDependencies bool,
) (*domain.AnalyzeResponse, *service.ProjectSnapshot, error) {
	useCase, err := buildAnalyzeUseCase(cmd, false)
	if err != nil {
		return nil, nil, fmt.Errorf("build analyzer: %w", err)
	}

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	result, err := useCase.ExecuteProjectWithOverrides(ctx, app.AnalyzeUseCaseConfig{
		ConfigFile:              c.configFile,
		SkipComplexity:          skipComplexity,
		SkipDeadCode:            skipDeadCode,
		SkipClones:              skipClones,
		SkipCBO:                 true,
		SkipLCOM:                true,
		SkipSystem:              skipDependencies,
		SkipCommunities:         true,
		SkipCommunitiesExplicit: true,
		SelectAnalysesUsed:      true,
		MinSeverity:             domain.DeadCodeSeverityCritical,
	}, paths, app.AnalyzeRequestOverrides{
		ComplexityEnabled:         domain.BoolPtr(!skipComplexity),
		DeadCodeEnabled:           domain.BoolPtr(!skipDeadCode),
		SystemEnabled:             domain.BoolPtr(!skipDependencies),
		SystemAnalyzeDependencies: domain.BoolPtr(!skipDependencies),
		SystemAnalyzeArchitecture: domain.BoolPtr(false),
	})
	if result == nil {
		return nil, nil, err
	}
	return result.Response, result.Snapshot, err
}

func (c *CheckCommand) reportProjectDiagnostics(writer io.Writer, diagnostics []domain.AnalysisDiagnostic) error {
	if len(diagnostics) == 0 {
		return nil
	}

	if !c.quiet || !c.allowParseErrors {
		for _, diagnostic := range diagnostics {
			fmt.Fprintf(writer, "%s: %s: %s\n", diagnostic.FilePath, diagnostic.Code, diagnostic.Message)
		}
	}
	if c.allowParseErrors {
		return nil
	}
	return newAnalysisError(fmt.Errorf("%d file(s) could not be analyzed (use --allow-parse-errors to ignore)", len(diagnostics)))
}

func (c *CheckCommand) countComplexityIssues(cmd *cobra.Command, response *domain.ComplexityResponse) int {
	if response == nil {
		return 0
	}

	maxComplexity := c.maxComplexity
	if !cmd.Flags().Changed("max-complexity") && response.Request != nil && response.Request.MaxComplexity > 0 {
		maxComplexity = response.Request.MaxComplexity
	}
	slocThreshold := domain.DefaultFunctionSLOCCriticalThreshold
	if response.Request != nil && response.Request.FunctionSLOCCriticalThreshold > 0 {
		slocThreshold = response.Request.FunctionSLOCCriticalThreshold
	}

	issueCount := 0
	for _, function := range response.Functions {
		if function.Metrics.Complexity > maxComplexity {
			issueCount++
			if !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: %s is too complex (%d > %d)\n",
					function.FilePath, function.StartLine, function.StartColumn+1, function.Name, function.Metrics.Complexity, maxComplexity)
			}
		}
		if function.ExceedsSLOC(slocThreshold) {
			issueCount++
			if !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: %s is too long (%d SLOC > %d)\n",
					function.FilePath, function.StartLine, function.StartColumn+1, function.Name, function.Metrics.SLOC, slocThreshold)
			}
		}
	}
	return issueCount
}

func (c *CheckCommand) countDeadCodeIssues(cmd *cobra.Command, response *domain.DeadCodeResponse) int {
	if response == nil {
		return 0
	}

	minSeverity := domain.DeadCodeSeverityCritical
	if response.Request != nil && response.Request.MinSeverity != "" {
		minSeverity = response.Request.MinSeverity
	}

	issueCount := 0
	for _, file := range response.Files {
		for _, function := range file.Functions {
			for _, finding := range function.Findings {
				if !finding.Severity.IsAtLeast(minSeverity) {
					continue
				}
				issueCount++
				if !c.quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: %s (%s)\n",
						finding.Location.FilePath,
						finding.Location.StartLine,
						finding.Location.StartColumn+1,
						finding.Reason,
						finding.Severity)
				}
			}
		}
	}
	return issueCount
}

func (c *CheckCommand) countCloneIssues(cmd *cobra.Command, response *domain.CloneResponse) int {
	if response == nil {
		return 0
	}
	for _, pair := range response.ClonePairs {
		if !c.quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: clone of %s:%d:%d (similarity: %.1f%%)\n",
				pair.Clone1.Location.FilePath,
				pair.Clone1.Location.StartLine,
				pair.Clone1.Location.StartCol+1,
				pair.Clone2.Location.FilePath,
				pair.Clone2.Location.StartLine,
				pair.Clone2.Location.StartCol+1,
				pair.Similarity*100)
		}
	}
	return len(response.ClonePairs)
}

func (c *CheckCommand) countCircularDependencyIssues(cmd *cobra.Command, response *domain.SystemAnalysisResponse) int {
	if response == nil || response.DependencyAnalysis == nil || response.DependencyAnalysis.CircularDependencies == nil {
		return 0
	}

	result := response.DependencyAnalysis
	cycles := result.CircularDependencies
	if !cycles.HasCircularDependencies {
		return 0
	}
	for _, cycle := range cycles.CircularDependencies {
		if len(cycle.Modules) == 0 {
			continue
		}
		firstModule := cycle.Modules[0]
		filePath := firstModule
		if metric, ok := result.ModuleMetrics[firstModule]; ok && metric != nil && metric.FilePath != "" {
			filePath = metric.FilePath
		}
		if !c.quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s:1:1: circular dependency detected: %s\n",
				filePath, strings.Join(cycle.Modules, " -> "))
		}
	}
	return cycles.TotalCycles
}
