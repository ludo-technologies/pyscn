package service

import (
	"bytes"
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
