package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

// AnalyzeFormatter handles formatting of unified analysis reports
type AnalyzeFormatter struct {
	complexityFormatter *OutputFormatterImpl
	deadCodeFormatter   *DeadCodeFormatterImpl
	cloneFormatter      *CloneOutputFormatter
}

// NewAnalyzeFormatter creates a new analyze formatter
func NewAnalyzeFormatter() *AnalyzeFormatter {
	return &AnalyzeFormatter{
		complexityFormatter: NewOutputFormatter(),
		deadCodeFormatter:   NewDeadCodeFormatter(),
		cloneFormatter:      NewCloneOutputFormatter(),
	}
}

// Write formats and writes the unified analysis response
func (f *AnalyzeFormatter) Write(response *domain.AnalyzeResponse, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		return f.writeText(response, writer)
	case domain.OutputFormatJSON:
		return WriteJSON(writer, normalizeAnalyzeResponseForJSON(response))
	case domain.OutputFormatYAML:
		return WriteYAML(writer, response)
	case domain.OutputFormatCSV:
		return f.writeCSV(response, writer)
	case domain.OutputFormatHTML:
		return f.writeHTML(response, writer)
	default:
		return domain.NewUnsupportedFormatError(string(format))
	}
}

// writeText formats the response as plain text
func (f *AnalyzeFormatter) writeText(response *domain.AnalyzeResponse, writer io.Writer) error {
	utils := NewFormatUtils()
	target := writer
	var report strings.Builder
	writer = &report

	// Header
	fmt.Fprint(writer, utils.FormatMainHeader("Comprehensive Analysis Report"))

	// Overall health and duration
	healthStats := map[string]interface{}{
		"Health Score":      fmt.Sprintf("%d/100 (%s)", response.Summary.HealthScore, response.Summary.Grade),
		"Analysis Duration": fmt.Sprintf("%.2fs", float64(response.Duration)/1000.0),
		"Generated":         response.GeneratedAt.Format(time.RFC3339),
	}
	fmt.Fprint(writer, utils.FormatSummaryStats(healthStats))

	// File statistics
	fmt.Fprint(writer, utils.FormatFileStats(
		response.Summary.AnalyzedFiles,
		response.Summary.TotalFiles,
		response.Summary.TotalFiles-response.Summary.AnalyzedFiles))

	// Analysis modules results
	if response.Summary.ComplexityEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("COMPLEXITY ANALYSIS"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Total Functions", formatFunctionCoverage(response.Summary.TotalFunctions, response.Summary.FunctionsParsed)))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Class Scopes", response.Summary.TotalClassScopes))
		if response.Summary.TotalClassScopes > 0 {
			fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Max Class Complexity", response.Summary.MaxClassComplexity))
		}
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Average Complexity", fmt.Sprintf("%.1f", response.Summary.AverageComplexity)))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "High Complexity Count", response.Summary.HighComplexityCount))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.DeadCodeEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("DEAD CODE DETECTION"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Total Issues", response.Summary.DeadCodeCount))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Critical Issues", response.Summary.CriticalDeadCode))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.CloneEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("CLONE DETECTION"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Unique Fragments", response.Summary.TotalClones))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Clone Groups", response.Summary.CloneGroups))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Fragments Cloned", utils.FormatPercentage(response.Summary.CodeDuplication)))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.CBOEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("DEPENDENCY ANALYSIS"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Classes Analyzed", response.Summary.CBOClasses))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "High Coupling Classes", response.Summary.HighCouplingClasses))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Average Coupling", fmt.Sprintf("%.1f", response.Summary.AverageCoupling)))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if len(response.ModuleQuality) > 0 {
		fmt.Fprint(writer, utils.FormatSectionHeader("MODULE QUALITY HOTSPOTS"))
		for index, module := range response.ModuleQuality {
			if index >= 10 {
				break
			}

			label := module.FilePath
			if module.ModuleName != "" {
				label = fmt.Sprintf("%s (%s)", module.ModuleName, module.FilePath)
			}
			fmt.Fprintf(writer, "  %s\n", label)
			if module.FunctionCount > 0 {
				fmt.Fprintf(writer, "    Definitions: %d functions\n", module.FunctionCount)
			}
			fmt.Fprintf(writer, "    Function complexity: %d analyzed, avg %.2f, max %d, high-risk %d, handlers %d\n",
				module.AnalyzedFunctionCount,
				module.AverageComplexity, module.MaxComplexity, module.HighRiskFunctionCount, module.ExceptionHandlerCount)
			fmt.Fprintf(writer, "    Cognitive: avg %.2f\n", module.AverageCognitiveComplexity)
			fmt.Fprintf(writer, "    Dead code: %d findings, %d blocks\n",
				module.DeadCodeFindingCount, module.DeadCodeBlockCount)
		}
		if len(response.ModuleQuality) > 10 {
			fmt.Fprintf(writer, "  Showing top 10 of %d modules\n", len(response.ModuleQuality))
		}
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Complexity != nil && len(response.Complexity.ByDirectory) > 0 {
		fmt.Fprint(writer, formatDirectoryComplexityText(response.Complexity.ByDirectory, 10))
	}

	if response.Summary.CommunitiesEnabled && response.Communities != nil {
		WriteCommunityTextSummary(writer, response.Communities)
	}

	if _, err := io.WriteString(target, report.String()); err != nil {
		return domain.NewOutputError("failed to write analysis text", err)
	}
	return nil
}

// writeCSV formats summary and quality metrics as CSV.
func (f *AnalyzeFormatter) writeCSV(response *domain.AnalyzeResponse, writer io.Writer) error {
	directories := []domain.DirectoryComplexityMetrics(nil)
	if response.Complexity != nil {
		directories = response.Complexity.ByDirectory
	}
	rowCapacity := 16 + (12 * len(response.ModuleQuality))
	if response.Complexity != nil {
		rowCapacity += 1 + (7 * len(directories))
	}
	if response.Summary.CommunitiesEnabled && response.Communities != nil {
		rowCapacity += 6
	}
	rows := make([][]string, 0, rowCapacity)
	rows = append(rows,
		[]string{"Metric", "Value"},
		[]string{"Health Score", fmt.Sprint(response.Summary.HealthScore)},
		[]string{"Grade", response.Summary.Grade},
		[]string{"Total Files", fmt.Sprint(response.Summary.TotalFiles)},
		[]string{"Analyzed Files", fmt.Sprint(response.Summary.AnalyzedFiles)},
		[]string{"Total Functions", fmt.Sprint(response.Summary.TotalFunctions)},
		[]string{"Class Scopes", fmt.Sprint(response.Summary.TotalClassScopes)},
		[]string{"Average Complexity", fmt.Sprintf("%.2f", response.Summary.AverageComplexity)},
		[]string{"High Complexity Count", fmt.Sprint(response.Summary.HighComplexityCount)},
		[]string{"Dead Code Count", fmt.Sprint(response.Summary.DeadCodeCount)},
		[]string{"Critical Dead Code", fmt.Sprint(response.Summary.CriticalDeadCode)},
		[]string{"Unique Fragments", fmt.Sprint(response.Summary.TotalClones)},
		[]string{"Clone Groups", fmt.Sprint(response.Summary.CloneGroups)},
		[]string{"Code Duplication", fmt.Sprintf("%.2f", response.Summary.CodeDuplication)},
		[]string{"Total Classes Analyzed", fmt.Sprint(response.Summary.CBOClasses)},
		[]string{"High Coupling (CBO) Classes", fmt.Sprint(response.Summary.HighCouplingClasses)},
		[]string{"Average CBO", fmt.Sprintf("%.2f", response.Summary.AverageCoupling)},
		[]string{"Module Quality Count", fmt.Sprint(len(response.ModuleQuality))},
	)

	for index, module := range response.ModuleQuality {
		prefix := fmt.Sprintf("Module %d ", index+1)
		rows = append(rows,
			[]string{prefix + "Name", module.ModuleName},
			[]string{prefix + "File Path", module.FilePath},
			[]string{prefix + "Lines of Code", fmt.Sprint(module.LinesOfCode)},
			[]string{prefix + "Function Count", fmt.Sprint(module.FunctionCount)},
			[]string{prefix + "Analyzed Function Count", fmt.Sprint(module.AnalyzedFunctionCount)},
			[]string{prefix + "Average Complexity", fmt.Sprintf("%.2f", module.AverageComplexity)},
			[]string{prefix + "Average Cognitive Complexity", fmt.Sprintf("%.2f", module.AverageCognitiveComplexity)},
			[]string{prefix + "Max Complexity", fmt.Sprint(module.MaxComplexity)},
			[]string{prefix + "High Risk Function Count", fmt.Sprint(module.HighRiskFunctionCount)},
			[]string{prefix + "Exception Handler Count", fmt.Sprint(module.ExceptionHandlerCount)},
			[]string{prefix + "Dead Code Findings", fmt.Sprint(module.DeadCodeFindingCount)},
			[]string{prefix + "Dead Code Blocks", fmt.Sprint(module.DeadCodeBlockCount)},
		)
	}

	if response.Summary.CommunitiesEnabled && response.Communities != nil {
		communities := response.Communities
		rows = append(rows,
			[]string{"Communities Enabled", "true"},
			[]string{"Total Communities", fmt.Sprint(communities.TotalCommunities)},
			[]string{"Community Modularity", fmt.Sprintf("%.4f", communities.Modularity)},
			[]string{"Bridge Modules", fmt.Sprint(len(communities.BridgeModules))},
			[]string{"Community Score", fmt.Sprint(response.Summary.CommunityScore)},
			[]string{"Community Risk Score", fmt.Sprint(response.Summary.CommunityRiskScore)},
		)
	}

	if response.Complexity != nil {
		rows = append(rows, []string{"Directory Complexity Count", fmt.Sprint(len(directories))})
		for index, directory := range directories {
			prefix := fmt.Sprintf("Directory %d ", index+1)
			rows = append(rows,
				[]string{prefix + "Path", directory.DirectoryPath},
				[]string{prefix + "Function Count", fmt.Sprint(directory.FunctionCount)},
				[]string{prefix + "Average Complexity", fmt.Sprintf("%.2f", directory.AverageComplexity)},
				[]string{prefix + "Max Complexity", fmt.Sprint(directory.MaxComplexity)},
				[]string{prefix + "High Risk Function Count", fmt.Sprint(directory.HighRiskFunctionCount)},
				[]string{prefix + "Average Nesting Depth", fmt.Sprintf("%.2f", directory.AverageNestingDepth)},
				[]string{prefix + "Max Nesting Depth", fmt.Sprint(directory.MaxNestingDepth)},
			)
		}
	}

	csvWriter := csv.NewWriter(writer)
	for _, row := range rows {
		if err := csvWriter.Write(row); err != nil {
			return domain.NewOutputError("failed to write CSV output", err)
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return domain.NewOutputError("failed to write CSV output", err)
	}

	return nil
}

// longFunctionsView carries the report's long-function section. The threshold
// travels with the rows so the section can name the bar the functions cleared.
type longFunctionsView struct {
	Threshold int
	Total     int
	Functions []domain.FunctionComplexity
}

// longFunctionsDisplayLimit caps the rows rendered in the long-function table,
// matching the top-N convention of the other report tables.
const longFunctionsDisplayLimit = 10

// collectLongFunctions returns the longest functions above the configured warn
// threshold, longest first. The complexity tab's other table is ranked by
// McCabe, where a flat 200-line function never surfaces.
func collectLongFunctions(complexity *domain.ComplexityResponse) longFunctionsView {
	view := longFunctionsView{Threshold: domain.DefaultFunctionSLOCWarnThreshold}
	if complexity == nil {
		return view
	}
	if complexity.Request != nil && complexity.Request.FunctionSLOCWarnThreshold > 0 {
		view.Threshold = complexity.Request.FunctionSLOCWarnThreshold
	}

	for _, function := range complexity.Functions {
		if function.ExceedsSLOC(view.Threshold) {
			view.Functions = append(view.Functions, function)
		}
	}
	view.Total = len(view.Functions)

	sort.SliceStable(view.Functions, func(i, j int) bool {
		return view.Functions[i].Metrics.SLOC > view.Functions[j].Metrics.SLOC
	})
	if len(view.Functions) > longFunctionsDisplayLimit {
		view.Functions = view.Functions[:longFunctionsDisplayLimit]
	}

	return view
}

// writeHTML formats the response as HTML
func (f *AnalyzeFormatter) writeHTML(response *domain.AnalyzeResponse, writer io.Writer) error {
	return writeAnalyzeHTML(response, writer)
}
