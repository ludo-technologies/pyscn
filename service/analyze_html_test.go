package service

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestWriteAnalyzeHTML_EmptyResponseRendersOverviewOnly(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(&domain.AnalyzeResponse{}, &buf))

	html := buf.String()
	assert.Contains(t, html, `data-tab="overview"`)
	assert.NotContains(t, html, `data-tab="functions"`)
	assert.NotContains(t, html, `data-tab="duplication"`)
	assert.NotContains(t, html, `data-tab="classes"`)
	assert.NotContains(t, html, `data-tab="architecture"`)
	assert.Contains(t, html, "No analyses were enabled for this run.")
}

func TestWriteAnalyzeHTML_ComplexityOnly(t *testing.T) {
	response := &domain.AnalyzeResponse{
		Summary: domain.AnalyzeSummary{
			ComplexityEnabled:   true,
			ComplexityScore:     100,
			HealthScore:         100,
			Grade:               "A",
			TotalFiles:          1,
			TotalFunctions:      2,
			HighComplexityCount: 1,
		},
		Complexity: &domain.ComplexityResponse{
			Functions: []domain.FunctionComplexity{
				{Name: "simple", FilePath: "pkg/a.py", Metrics: domain.ComplexityMetrics{Complexity: 1, SLOC: 3}, RiskLevel: domain.RiskLevelLow},
				{Name: "gnarly", FilePath: "pkg/a.py", Metrics: domain.ComplexityMetrics{Complexity: 25, SLOC: 80, NestingDepth: 6}, RiskLevel: domain.RiskLevelHigh},
				{Name: domain.ModuleFunctionName, FilePath: "pkg/a.py", Metrics: domain.ComplexityMetrics{Complexity: 1, SLOC: 900}},
			},
			Summary: domain.ComplexitySummary{TotalFunctions: 3, FunctionsParsed: 3, MaxComplexity: 25},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(response, &buf))
	html := buf.String()

	assert.Contains(t, html, `data-tab="functions"`)
	assert.NotContains(t, html, `data-tab="duplication"`)
	assert.Contains(t, html, "Complexity distribution")
	assert.Contains(t, html, "risk from CC 10")
	// The module pseudo-function is never the "longest function" fact.
	assert.Contains(t, html, "80 SLOC (gnarly)")
	assert.NotContains(t, html, "900 SLOC")
	assert.Contains(t, html, "Complexity scores 100/100 across 1 file.")
}

func TestWriteAnalyzeHTML_ClassExecutionScopesRemainAdditive(t *testing.T) {
	response := &domain.AnalyzeResponse{
		Summary: domain.AnalyzeSummary{
			ComplexityEnabled:             true,
			ComplexityScore:               100,
			HealthScore:                   100,
			Grade:                         "A",
			TotalClassScopes:              1,
			HighComplexityClassScopeCount: 1,
		},
		Complexity: &domain.ComplexityResponse{
			ClassScopes: []domain.FunctionComplexity{{
				Name:      "Config",
				ScopeKind: domain.AnalysisScopeClass,
				FilePath:  "/repo/app/config.py",
				StartLine: 2,
				RiskLevel: domain.RiskLevelHigh,
				Metrics:   domain.ComplexityMetrics{Complexity: 12, CognitiveComplexity: 9, NestingDepth: 2},
			}},
			Summary: domain.ComplexitySummary{
				TotalClassScopes:            1,
				MaxClassComplexity:          12,
				MaxClassCognitiveComplexity: 9,
				MaxClassNestingDepth:        2,
				HighRiskClassScopes:         1,
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(response, &buf))
	html := buf.String()
	assert.Contains(t, html, `data-tab="functions"`)
	assert.Contains(t, html, "Class execution scopes")
	assert.Contains(t, html, "Config")
	assert.Contains(t, html, "/repo/app/config.py:2")
	assert.Contains(t, html, "/repo/app")
}

func TestBuildReportVerdict_NamesWeakDimensions(t *testing.T) {
	response := &domain.AnalyzeResponse{Summary: domain.AnalyzeSummary{
		ComplexityEnabled: true, ComplexityScore: 100,
		DeadCodeEnabled: true, DeadCodeScore: 95,
		CloneEnabled: true, DuplicationScore: 40, CodeDuplication: 22.3, CloneGroups: 27,
		CBOEnabled: true, CouplingScore: 70, AverageCoupling: 2.7,
		Grade: "C",
	}}
	dims := buildReportDimensions(response)
	verdict := buildReportVerdict(response, dims)

	assert.Equal(t, "Fair, with clear debt to pay down", verdict.Headline)
	var text strings.Builder
	var strong []string
	for _, seg := range verdict.Body {
		text.WriteString(seg.Text)
		if seg.Strong {
			strong = append(strong, seg.Text)
		}
	}
	assert.Equal(t, "Complexity and dead code are clean. Most of the remaining debt is in duplication (22.3% of fragments, 27 groups) and coupling (avg CBO 2.7, 0 of 0 high).", text.String())
	assert.Equal(t, []string{"duplication", "coupling"}, strong)
}

func TestJoinNames(t *testing.T) {
	assert.Equal(t, "", joinNames(nil))
	assert.Equal(t, "Complexity", joinNames([]string{"complexity"}))
	assert.Equal(t, "Complexity and cohesion", joinNames([]string{"complexity", "cohesion"}))
	assert.Equal(t, "Complexity, dead code, and cohesion", joinNames([]string{"complexity", "dead code", "cohesion"}))
}

func TestBuildReportHistogram_BinsFollowRiskThresholds(t *testing.T) {
	complexity := &domain.ComplexityResponse{
		Request: &domain.ComplexityRequest{LowThreshold: 9, MediumThreshold: 19},
	}
	for _, cc := range []int{1, 1, 3, 9, 10, 19, 20, 50} {
		complexity.Functions = append(complexity.Functions, domain.FunctionComplexity{Name: "f", Metrics: domain.ComplexityMetrics{Complexity: cc}})
	}

	hist := buildReportHistogram(complexity)
	require.NotNil(t, hist)
	require.Len(t, hist.Bins, 5)

	labels := make([]string, 0, len(hist.Bins))
	counts := make([]int, 0, len(hist.Bins))
	for _, bin := range hist.Bins {
		labels = append(labels, bin.Label)
		counts = append(counts, bin.Count)
	}
	assert.Equal(t, []string{"1", "2–5", "6–9", "10–19", "20+"}, labels)
	assert.Equal(t, []int{2, 1, 1, 2, 2}, counts)
	assert.Equal(t, "warn", hist.Bins[3].Band)
	assert.Equal(t, "bad", hist.Bins[4].Band)
	assert.Equal(t, "risk from CC 10", hist.Threshold)
	assert.Equal(t, "8", hist.Total)
	assert.Nil(t, buildReportHistogram(nil))
	assert.Nil(t, buildReportHistogram(&domain.ComplexityResponse{}))
}

func TestCountClonesByFile(t *testing.T) {
	loc := func(file string) *domain.CloneLocation { return &domain.CloneLocation{FilePath: file} }

	withGroups := &domain.CloneResponse{
		CloneGroups: []*domain.CloneGroup{{Clones: []*domain.Clone{{Location: loc("a.py")}, {Location: loc("a.py")}, {Location: loc("b.py")}}}},
		ClonePairs:  []*domain.ClonePair{{Clone1: &domain.Clone{Location: loc("zzz.py")}, Clone2: &domain.Clone{Location: loc("zzz.py")}}},
	}
	assert.Equal(t, map[string]int{"a.py": 2, "b.py": 1}, countClonesByFile(withGroups))

	pairsOnly := &domain.CloneResponse{
		ClonePairs: []*domain.ClonePair{{Clone1: &domain.Clone{Location: loc("a.py")}, Clone2: &domain.Clone{Location: loc("b.py")}}},
	}
	assert.Equal(t, map[string]int{"a.py": 1, "b.py": 1}, countClonesByFile(pairsOnly))
	assert.Empty(t, countClonesByFile(nil))
}

func TestBuildReportHotspots_RanksAndFormats(t *testing.T) {
	response := &domain.AnalyzeResponse{
		Summary: domain.AnalyzeSummary{CloneEnabled: true},
		ModuleQuality: []domain.ModuleQualityMetrics{
			{FilePath: "pkg/quiet.py", LinesOfCode: 5000, ModuleComplexityMetrics: domain.ModuleComplexityMetrics{MaxComplexity: 3}},
			{FilePath: "pkg/sub/hot.py", LinesOfCode: 1234, ModuleComplexityMetrics: domain.ModuleComplexityMetrics{MaxComplexity: 30, HighRiskFunctionCount: 2}},
		},
		Clone: &domain.CloneResponse{CloneGroups: []*domain.CloneGroup{{Clones: []*domain.Clone{{Location: &domain.CloneLocation{FilePath: "pkg/sub/hot.py"}}}}}},
	}

	rows := buildReportHotspots(response)
	require.Len(t, rows, 2)
	assert.Equal(t, "pkg/sub/", rows[0].Dir)
	assert.Equal(t, "hot.py", rows[0].File)
	assert.Equal(t, "1,234", rows[0].Lines)
	assert.Equal(t, 100, rows[0].MaxCCPct)
	assert.Equal(t, "bad", rows[0].MaxCCBand)
	assert.Equal(t, 1, rows[0].Clones)
	assert.Equal(t, "quiet.py", rows[1].File)
	assert.Equal(t, 10, rows[1].MaxCCPct)
	assert.Equal(t, "", rows[1].MaxCCBand)
}

func TestFormatThousands(t *testing.T) {
	assert.Equal(t, "0", formatThousands(0))
	assert.Equal(t, "999", formatThousands(999))
	assert.Equal(t, "1,000", formatThousands(1000))
	assert.Equal(t, "1,234,567", formatThousands(1234567))
}

func TestReportProject_FallsBackToCommonDirectory(t *testing.T) {
	response := &domain.AnalyzeResponse{ModuleQuality: []domain.ModuleQualityMetrics{
		{FilePath: "/repo/app/a.py"},
		{FilePath: "/repo/app/sub/b.py"},
	}}
	name, root := reportProject(response)
	assert.Equal(t, "app", name)
	assert.Equal(t, "/repo/app", root)

	name, root = reportProject(&domain.AnalyzeResponse{ModuleQuality: []domain.ModuleQualityMetrics{{FilePath: "a.py"}}})
	assert.Equal(t, "", name)
	assert.Equal(t, "", root)
}

func TestBuildReportFixes_SplitsTopAndRemainderWithSteps(t *testing.T) {
	response := &domain.AnalyzeResponse{}
	for i := 0; i < reportSuggestionLimit+5; i++ {
		response.Suggestions = append(response.Suggestions, domain.Suggestion{
			Category: domain.SuggestionCategoryDeadCode,
			Severity: domain.SuggestionSeverityWarning,
			Effort:   domain.SuggestionEffortEasy,
			Title:    fmt.Sprintf("s%d", i),
			Steps:    []string{"first", "second"},
		})
	}

	top, more, total := buildReportFixes(response)
	require.Len(t, top, reportFixLimit)
	assert.Len(t, more, reportSuggestionLimit-reportFixLimit)
	assert.Equal(t, reportSuggestionLimit+5, total)
	assert.Equal(t, "dead code", top[0].Category)
	assert.Equal(t, []string{"first", "second"}, top[0].Steps)
	assert.Equal(t, fmt.Sprintf("s%d", reportFixLimit), more[0].Title)

	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(response, &buf))
	html := buf.String()
	assert.Contains(t, html, fmt.Sprintf("Show %d more suggestions", reportSuggestionLimit-reportFixLimit))
	assert.Contains(t, html, fmt.Sprintf("Showing %d of %d suggestions", reportSuggestionLimit, reportSuggestionLimit+5))
	assert.Contains(t, html, `<details class="steps"><summary>2 steps</summary><ol><li>first</li><li>second</li></ol></details>`)
}

func TestWriteAnalyzeHTML_SurfacesSkippedFiles(t *testing.T) {
	response := &domain.AnalyzeResponse{Summary: domain.AnalyzeSummary{
		TotalFiles: 100, AnalyzedFiles: 60, SkippedFiles: 40,
		ComplexityEnabled: true, ComplexityScore: 100, Grade: "C", HealthScore: 70,
	}}
	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(response, &buf))
	html := buf.String()
	assert.Contains(t, html, "60 of 100 files analyzed, 40 skipped")
	assert.Contains(t, html, "<strong>40 files of 100 could not be analyzed</strong>")
}

func TestBuildReportDimensions_UnlinkedWhenTabMissing(t *testing.T) {
	// Dependency scoring ran but the system payload is absent, so there is no
	// Architecture tab for the card to point at.
	response := &domain.AnalyzeResponse{Summary: domain.AnalyzeSummary{DepsEnabled: true, DependencyScore: 80, ComplexityEnabled: true}}
	dims := buildReportDimensions(response)
	require.Len(t, dims, 2)
	assert.Equal(t, "functions", dims[0].Tab)
	assert.Equal(t, "Dependencies", dims[1].Name)
	assert.Equal(t, "", dims[1].Tab)

	var buf bytes.Buffer
	require.NoError(t, writeAnalyzeHTML(response, &buf))
	assert.NotContains(t, buf.String(), `data-goto="architecture"`)
	assert.Contains(t, buf.String(), `class="dim static ok"`)
}

func TestCountClonesByFile_PairsDedupeFragments(t *testing.T) {
	frag := func(file string, start int) *domain.Clone {
		return &domain.Clone{Location: &domain.CloneLocation{FilePath: file, StartLine: start, EndLine: start + 5}}
	}
	// a.py:1 pairs with both b.py:1 and c.py:1 but is one fragment.
	pairs := &domain.CloneResponse{ClonePairs: []*domain.ClonePair{
		{Clone1: frag("a.py", 1), Clone2: frag("b.py", 1)},
		{Clone1: frag("a.py", 1), Clone2: frag("c.py", 1)},
	}}
	assert.Equal(t, map[string]int{"a.py": 1, "b.py": 1, "c.py": 1}, countClonesByFile(pairs))
}

func TestHistogramBins_CollapseForLowThresholds(t *testing.T) {
	labels := func(bins []histogramBin) []string {
		out := make([]string, 0, len(bins))
		for _, bin := range bins {
			out = append(out, bin.label)
		}
		return out
	}
	assert.Equal(t, []string{"1", "2–5", "6–9", "10–19", "20+"}, labels(histogramBins(9, 19)))
	assert.Equal(t, []string{"1", "2–3", "4+"}, labels(histogramBins(1, 3)))
	assert.Equal(t, []string{"1", "2–5", "6–20", "21+"}, labels(histogramBins(5, 20)))
	// A degenerate medium <= low still yields increasing bins.
	assert.Equal(t, []string{"1", "2–5", "6–9", "10", "11+"}, labels(histogramBins(9, 9)))
	assert.Equal(t, "warn", histogramBins(1, 3)[1].band)
	assert.Equal(t, "bad", histogramBins(1, 3)[2].band)
}

func TestMedian(t *testing.T) {
	assert.Equal(t, 0.0, median(nil))
	assert.Equal(t, 5.0, median([]int{9, 1}))
	assert.Equal(t, 2.0, median([]int{3, 1, 2}))
	assert.Equal(t, "5", formatMedian(5))
	assert.Equal(t, "2.5", formatMedian(2.5))
}
