package service

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

//go:embed templates/analyze/report.html templates/analyze/report.css templates/analyze/report.js
var analyzeTemplateFS embed.FS

var (
	analyzeReportCSS = template.CSS(mustReadTemplateAsset("templates/analyze/report.css"))
	analyzeReportJS  = template.JS(mustReadTemplateAsset("templates/analyze/report.js"))

	analyzeReportTemplate = template.Must(
		template.New("report.html").
			Funcs(analyzeTemplateFuncs()).
			ParseFS(analyzeTemplateFS, "templates/analyze/report.html"),
	)
)

func mustReadTemplateAsset(name string) string {
	data, err := analyzeTemplateFS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("embedded template asset %s: %v", name, err))
	}
	return string(data)
}

func analyzeTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"join":   strings.Join,
		"add":    func(a, b int) int { return a + b },
		"sub":    func(a, b int) int { return a - b },
		"mul100": func(v float64) float64 { return v * 100.0 },
		"addf":   func(a, b float64) float64 { return a + b },
		"divf":   func(a, b float64) float64 { return a / b },
		"int":    func(t domain.CloneType) int { return int(t) },
		"previewContent": func(content string) string {
			const maxLines = 8
			lines := strings.Split(content, "\n")
			if len(lines) <= maxLines {
				return content
			}
			return strings.Join(lines[:maxLines], "\n") + "\n..."
		},
		"scoreBand":     scoreBand,
		"longFunctions": collectLongFunctions,
		"communitySummaryHTML": func(result *domain.CommunityAnalysisResult) template.HTML {
			if result == nil {
				return ""
			}
			var builder strings.Builder
			WriteCommunityHTMLSummary(&builder, result)
			return template.HTML(builder.String())
		},
	}
}

// writeAnalyzeHTML renders the unified HTML report.
func writeAnalyzeHTML(response *domain.AnalyzeResponse, writer io.Writer) error {
	view := buildAnalyzeReportView(response)
	if err := analyzeReportTemplate.Execute(writer, view); err != nil {
		return fmt.Errorf("failed to render HTML report: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// View model
// ---------------------------------------------------------------------------

const (
	reportFixLimit         = 5
	reportSuggestionLimit  = 30
	reportHotspotLimit     = 8
	reportScoreRingCircumf = 351.86 // 2 * pi * r for r=56
)

// analyzeReportView is the data handed to the report template. It embeds the
// response so detail tabs can reach the raw results, and adds the pre-computed
// overview blocks so the template stays free of arithmetic.
type analyzeReportView struct {
	*domain.AnalyzeResponse

	CSS template.CSS
	JS  template.JS

	ProjectName string
	ProjectPath string
	GradeClass  string
	ScoreBand   string
	RingOffset  float64

	Tabs        []reportTab
	Verdict     reportVerdict
	Facts       []reportFact
	Dimensions  []reportDimension
	Fixes       []reportFix // shown inline on the overview
	MoreFixes   []reportFix // remaining suggestions, collapsed
	FixTotal    int
	Hotspots    []reportHotspot
	Histogram   *reportHistogram
	Duplication *reportDuplication
	Classes     *reportClasses
	Structure   *reportStructure

	SkippedFiles        int
	ShowExecutionScopes bool
	ShowDeadColumn      bool
	ShowCloneColumn     bool
}

type reportTab struct {
	ID        string
	Label     string
	Count     int
	CountBand string // "", "warn", "bad"
}

type reportVerdict struct {
	Headline string
	Body     []reportSegment
}

// reportSegment is a run of verdict text; Strong runs are emphasized.
type reportSegment struct {
	Text   string
	Strong bool
}

type reportFact struct {
	Value string
	Label string
}

type reportDimension struct {
	Name  string
	Score int
	Band  string
	Left  string
	Right string
	Tab   string
}

type reportFix struct {
	Severity string
	Effort   string
	Category string
	Title    string
	Location string
	Why      string
	Steps    []string
}

type reportHotspot struct {
	Dir       string
	File      string
	Lines     string
	Functions int
	MaxCC     int
	MaxCCPct  int
	MaxCCBand string
	HighRisk  int
	DeadCode  int
	Clones    int
}

type reportHistogram struct {
	Total      string
	Bins       []reportHistogramBin
	Ticks      []reportHistogramTick
	ThresholdX float64
	Threshold  string
	Facts      []reportKV
}

type reportHistogramBin struct {
	Label  string
	Count  int
	X      float64
	Y      float64
	Width  float64
	Height float64
	Band   string // "", "warn", "bad"
}

type reportHistogramTick struct {
	Label string
	Y     float64
}

type reportKV struct {
	Key   string
	Value string
	Band  string
	Mono  bool
}

type reportDuplication struct {
	Percent   float64
	Fragments int
	Types     []reportShare
	Facts     []reportKV
}

type reportShare struct {
	Label   string
	Percent float64
	Class   string
}

type reportClasses struct {
	Total int
	Facts []reportKV
}

type reportStructure struct {
	Cycles int
	Facts  []reportKV
}

// scoreBand maps a 0-100 score onto the report's three semantic colors.
func scoreBand(score int) string {
	switch {
	case score >= domain.ScoreThresholdGood:
		return "ok"
	case score >= domain.ScoreThresholdFair:
		return "watch"
	default:
		return "poor"
	}
}

func buildAnalyzeReportView(response *domain.AnalyzeResponse) *analyzeReportView {
	view := &analyzeReportView{
		AnalyzeResponse:     response,
		CSS:                 analyzeReportCSS,
		JS:                  analyzeReportJS,
		GradeClass:          "grade-" + SafeHTMLID(strings.ToLower(response.Summary.Grade)),
		ScoreBand:           scoreBand(response.Summary.HealthScore),
		RingOffset:          reportScoreRingCircumf * (1 - float64(clampScore(response.Summary.HealthScore))/100),
		SkippedFiles:        response.Summary.SkippedFiles,
		ShowExecutionScopes: reportHasExecutionScopes(response),
		ShowDeadColumn:      response.Summary.DeadCodeEnabled,
		ShowCloneColumn:     response.Summary.CloneEnabled,
	}
	view.ProjectName, view.ProjectPath = reportProject(response)
	view.Tabs = buildReportTabs(response)
	view.Dimensions = buildReportDimensions(response)
	view.Verdict = buildReportVerdict(response, view.Dimensions)
	view.Facts = buildReportFacts(response)
	view.Fixes, view.MoreFixes, view.FixTotal = buildReportFixes(response)
	view.Hotspots = buildReportHotspots(response)
	view.Histogram = buildReportHistogram(response.Complexity)
	view.Duplication = buildReportDuplication(response)
	view.Classes = buildReportClasses(response)
	view.Structure = buildReportStructure(response)
	return view
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// reportProject derives a project label. The unified response carries no
// explicit root, so the system summary root is preferred and the common
// directory of the analyzed files is the fallback.
func reportProject(response *domain.AnalyzeResponse) (name, root string) {
	if response.System != nil && response.System.Summary.ProjectRoot != "" {
		root = response.System.Summary.ProjectRoot
		return path.Base(root), abbreviateHome(root)
	}
	var files []string
	for _, module := range response.ModuleQuality {
		files = append(files, module.FilePath)
	}
	if len(files) == 0 && response.Complexity != nil {
		for _, function := range response.Complexity.Functions {
			files = append(files, function.FilePath)
		}
		for _, classScope := range response.Complexity.ClassScopes {
			files = append(files, classScope.FilePath)
		}
	}
	root = commonDirectory(files)
	if root == "" || root == "." || root == "/" {
		return "", ""
	}
	return path.Base(root), root
}

// abbreviateHome swaps the user's home directory for "~" so the report header
// stays short and does not leak the account name when shared.
func abbreviateHome(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || !strings.HasPrefix(p, home) {
		return p
	}
	return "~" + strings.TrimPrefix(p, home)
}

func commonDirectory(files []string) string {
	if len(files) == 0 {
		return ""
	}
	prefix := strings.Split(path.Dir(files[0]), "/")
	for _, file := range files[1:] {
		parts := strings.Split(path.Dir(file), "/")
		n := 0
		for n < len(prefix) && n < len(parts) && prefix[n] == parts[n] {
			n++
		}
		prefix = prefix[:n]
		if len(prefix) == 0 {
			return ""
		}
	}
	return strings.Join(prefix, "/")
}

func buildReportTabs(response *domain.AnalyzeResponse) []reportTab {
	s := response.Summary
	tabs := []reportTab{{ID: "overview", Label: "Overview"}}

	if reportHasExecutionScopes(response) {
		tab := reportTab{ID: "functions", Label: "Functions", Count: s.HighComplexityCount + s.HighComplexityClassScopeCount + s.DeadCodeCount}
		switch {
		case s.HighComplexityCount > 0 || s.HighComplexityClassScopeCount > 0 || s.CriticalDeadCode > 0:
			tab.CountBand = "bad"
		case tab.Count > 0:
			tab.CountBand = "warn"
		}
		tabs = append(tabs, tab)
	}
	if s.CloneEnabled {
		tab := reportTab{ID: "duplication", Label: "Duplication", Count: s.CloneGroups}
		if tab.Count == 0 {
			tab.Count = s.ClonePairs
		}
		if tab.Count > 0 {
			tab.CountBand = "warn"
		}
		tabs = append(tabs, tab)
	}
	if s.CBOEnabled || s.LCOMEnabled {
		tab := reportTab{ID: "classes", Label: "Classes", Count: s.HighCouplingClasses + s.HighLCOMClasses}
		if tab.Count > 0 {
			tab.CountBand = "bad"
		}
		tabs = append(tabs, tab)
	}
	if reportHasArchitecture(response) {
		tab := reportTab{ID: "architecture", Label: "Architecture"}
		cycles, violations := 0, 0
		if response.System != nil {
			if deps := response.System.DependencyAnalysis; deps != nil && deps.CircularDependencies != nil {
				cycles = deps.CircularDependencies.TotalCycles
			}
			if arch := response.System.ArchitectureAnalysis; arch != nil {
				violations = arch.TotalViolations
			}
		}
		tab.Count = cycles + violations
		switch {
		case cycles > 0:
			tab.CountBand = "bad"
		case violations > 0:
			tab.CountBand = "warn"
		}
		tabs = append(tabs, tab)
	}
	return tabs
}

// reportHasExecutionScopes reports whether the execution-scope tab has
// function or class-suite analysis, or function-derived rollups to show.
func reportHasExecutionScopes(response *domain.AnalyzeResponse) bool {
	if response.Summary.ComplexityEnabled || response.Summary.DeadCodeEnabled || len(response.ModuleQuality) > 0 {
		return true
	}
	return response.Complexity != nil && (len(response.Complexity.ClassScopes) > 0 || len(response.Complexity.ByDirectory) > 0)
}

func reportHasArchitecture(response *domain.AnalyzeResponse) bool {
	if response.System != nil && (response.System.DependencyAnalysis != nil || response.System.ArchitectureAnalysis != nil) {
		return true
	}
	return response.Summary.CommunitiesEnabled && response.Communities != nil
}

func buildReportDimensions(response *domain.AnalyzeResponse) []reportDimension {
	s := response.Summary
	var dims []reportDimension
	tabs := make(map[string]bool)
	for _, tab := range buildReportTabs(response) {
		tabs[tab.ID] = true
	}
	// A dimension can be enabled while its detail tab is absent (for example
	// dependency scoring ran but the system analysis payload is missing), so
	// only link cards whose target tab is actually rendered.
	add := func(name string, score int, left, right, tab string) {
		if !tabs[tab] {
			tab = ""
		}
		dims = append(dims, reportDimension{Name: name, Score: score, Band: scoreBand(score), Left: left, Right: right, Tab: tab})
	}
	if s.ComplexityEnabled {
		add("Complexity", s.ComplexityScore,
			fmt.Sprintf("avg CC %.2f", s.AverageComplexity),
			fmt.Sprintf("%d high-risk", s.HighComplexityCount), "functions")
	}
	if s.DeadCodeEnabled {
		add("Dead code", s.DeadCodeScore,
			pluralize(s.DeadCodeCount, "finding", "findings"),
			fmt.Sprintf("%d critical", s.CriticalDeadCode), "functions")
	}
	if s.CloneEnabled {
		add("Duplication", s.DuplicationScore,
			fmt.Sprintf("%.1f%% of fragments", s.CodeDuplication),
			pluralize(s.CloneGroups, "group", "groups"), "duplication")
	}
	if s.CBOEnabled {
		add("Coupling", s.CouplingScore,
			fmt.Sprintf("avg CBO %.1f", s.AverageCoupling),
			fmt.Sprintf("%d of %d high", s.HighCouplingClasses, s.CBOClasses), "classes")
	}
	if s.LCOMEnabled {
		add("Cohesion", s.CohesionScore,
			fmt.Sprintf("avg LCOM4 %.2f", s.AverageLCOM),
			fmt.Sprintf("%d of %d low", s.HighLCOMClasses, s.LCOMClasses), "classes")
	}
	if s.DepsEnabled {
		cycles := "no cycles"
		if s.DepsModulesInCycles > 0 {
			cycles = fmt.Sprintf("%d modules in cycles", s.DepsModulesInCycles)
		}
		add("Dependencies", s.DependencyScore, cycles, fmt.Sprintf("depth %d", s.DepsMaxDepth), "architecture")
	}
	if s.ArchEnabled {
		violations := 0
		if response.System != nil && response.System.ArchitectureAnalysis != nil {
			violations = response.System.ArchitectureAnalysis.TotalViolations
		}
		add("Architecture", s.ArchitectureScore,
			fmt.Sprintf("%.0f%% compliant", s.ArchCompliance*100),
			pluralize(violations, "violation", "violations"), "architecture")
	}
	if s.CommunitiesEnabled && response.Communities != nil {
		add("Communities", s.CommunityScore,
			fmt.Sprintf("%d communities, Q %.2f", response.Communities.TotalCommunities, response.Communities.Modularity),
			pluralize(len(response.Communities.BridgeModules), "bridge", "bridges"), "architecture")
	}
	return dims
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

func buildReportVerdict(response *domain.AnalyzeResponse, dims []reportDimension) reportVerdict {
	verdict := reportVerdict{Headline: gradeHeadline(response.Summary.Grade)}

	var clean, weak []reportDimension
	for _, dim := range dims {
		switch {
		case dim.Score >= domain.ScoreThresholdExcellent:
			clean = append(clean, dim)
		case dim.Score < domain.ScoreThresholdGood:
			weak = append(weak, dim)
		}
	}
	sort.SliceStable(weak, func(i, j int) bool { return weak[i].Score < weak[j].Score })

	text := func(s string) { verdict.Body = append(verdict.Body, reportSegment{Text: s}) }
	strong := func(s string) { verdict.Body = append(verdict.Body, reportSegment{Text: s, Strong: true}) }

	if skipped := response.Summary.SkippedFiles; skipped > 0 {
		strong(fmt.Sprintf("%s of %d could not be analyzed", pluralize(skipped, "file", "files"), response.Summary.TotalFiles))
		text(" and were skipped; the health score is penalized for them. ")
	}

	switch {
	case len(dims) == 0:
		text("No analyses were enabled for this run.")
		return verdict
	case len(clean) == len(dims):
		files := pluralize(response.Summary.TotalFiles, "file", "files")
		if len(dims) == 1 {
			text(fmt.Sprintf("%s scores %d/100 across %s.", joinNames(lowerNames(dims)), dims[0].Score, files))
		} else {
			text(fmt.Sprintf("All %d dimensions score %d or above across %s.", len(dims), domain.ScoreThresholdExcellent, files))
		}
		return verdict
	case len(clean) > 0:
		text(joinNames(lowerNames(clean)) + " " + isAre(len(clean)) + " clean. ")
	}

	if len(weak) == 0 {
		text(fmt.Sprintf("No dimension scores below %d.", domain.ScoreThresholdGood))
		return verdict
	}
	if len(weak) > 3 {
		weak = weak[:3]
	}
	if len(clean) > 0 {
		text("Most of the remaining debt is in ")
	} else {
		text("Most of the debt is in ")
	}
	for i, dim := range weak {
		if i > 0 {
			if i == len(weak)-1 {
				text(" and ")
			} else {
				text(", ")
			}
		}
		strong(strings.ToLower(dim.Name))
		text(fmt.Sprintf(" (%s, %s)", dim.Left, dim.Right))
	}
	text(".")
	return verdict
}

func gradeHeadline(grade string) string {
	switch strings.ToUpper(grade) {
	case "A":
		return "Healthy codebase"
	case "B":
		return "Good shape overall"
	case "C":
		return "Fair, with clear debt to pay down"
	case "D":
		return "Quality needs attention"
	default:
		return "Serious quality problems"
	}
}

func lowerNames(dims []reportDimension) []string {
	names := make([]string, 0, len(dims))
	for _, dim := range dims {
		names = append(names, strings.ToLower(dim.Name))
	}
	return names
}

// joinNames renders "A", "A and B", or "A, B, and C" with the first name
// capitalized for sentence position.
func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	first := strings.ToUpper(names[0][:1]) + names[0][1:]
	switch len(names) {
	case 1:
		return first
	case 2:
		return first + " and " + names[1]
	default:
		return first + ", " + strings.Join(names[1:len(names)-1], ", ") + ", and " + names[len(names)-1]
	}
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

func buildReportFacts(response *domain.AnalyzeResponse) []reportFact {
	s := response.Summary
	var facts []reportFact
	lines := 0
	for _, module := range response.ModuleQuality {
		lines += module.LinesOfCode
	}
	if lines == 0 && response.Clone != nil && response.Clone.Statistics != nil {
		lines = response.Clone.Statistics.LinesAnalyzed
	}
	if lines > 0 {
		facts = append(facts, reportFact{Value: formatThousands(lines), Label: "lines"})
	}
	facts = append(facts, reportFact{Value: formatThousands(s.TotalFiles), Label: "files"})
	if s.ComplexityEnabled {
		facts = append(facts, reportFact{Value: formatThousands(s.TotalFunctions), Label: "functions"})
	}
	if classes := max(s.CBOClasses, s.LCOMClasses); s.CBOEnabled || s.LCOMEnabled {
		facts = append(facts, reportFact{Value: formatThousands(classes), Label: "classes"})
	}
	return facts
}

func formatThousands(n int) string {
	digits := fmt.Sprintf("%d", n)
	if len(digits) <= 3 {
		return digits
	}
	var builder strings.Builder
	head := len(digits) % 3
	if head > 0 {
		builder.WriteString(digits[:head])
	}
	for i := head; i < len(digits); i += 3 {
		if builder.Len() > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString(digits[i : i+3])
	}
	return builder.String()
}

// buildReportFixes splits suggestions into the inline top list and the
// collapsed remainder. The remainder is capped so huge runs do not bloat the
// report; the cap matches the previous report's suggestion table.
func buildReportFixes(response *domain.AnalyzeResponse) (top, more []reportFix, total int) {
	for i, suggestion := range response.Suggestions {
		if i == reportSuggestionLimit {
			break
		}
		fix := reportFix{
			Severity: string(suggestion.Severity),
			Effort:   string(suggestion.Effort),
			Category: strings.ReplaceAll(string(suggestion.Category), "_", " "),
			Title:    suggestion.Title,
			Why:      suggestion.Description,
			Steps:    suggestion.Steps,
		}
		if suggestion.FilePath != "" {
			fix.Location = suggestion.FilePath
			if suggestion.StartLine > 0 {
				fix.Location = fmt.Sprintf("%s:%d", suggestion.FilePath, suggestion.StartLine)
			}
		}
		if i < reportFixLimit {
			top = append(top, fix)
		} else {
			more = append(more, fix)
		}
	}
	return top, more, len(response.Suggestions)
}

func buildReportHotspots(response *domain.AnalyzeResponse) []reportHotspot {
	if len(response.ModuleQuality) == 0 {
		return nil
	}
	clonesByFile := countClonesByFile(response.Clone)

	modules := make([]domain.ModuleQualityMetrics, len(response.ModuleQuality))
	copy(modules, response.ModuleQuality)
	sort.SliceStable(modules, func(i, j int) bool {
		a, b := modules[i], modules[j]
		if a.HighRiskFunctionCount != b.HighRiskFunctionCount {
			return a.HighRiskFunctionCount > b.HighRiskFunctionCount
		}
		if a.MaxComplexity != b.MaxComplexity {
			return a.MaxComplexity > b.MaxComplexity
		}
		if a.DeadCodeFindingCount != b.DeadCodeFindingCount {
			return a.DeadCodeFindingCount > b.DeadCodeFindingCount
		}
		if clonesByFile[a.FilePath] != clonesByFile[b.FilePath] {
			return clonesByFile[a.FilePath] > clonesByFile[b.FilePath]
		}
		return a.LinesOfCode > b.LinesOfCode
	})
	if len(modules) > reportHotspotLimit {
		modules = modules[:reportHotspotLimit]
	}

	maxCC := 1
	for _, module := range modules {
		maxCC = max(maxCC, module.MaxComplexity)
	}
	low, medium := complexityThresholds(response.Complexity)

	rows := make([]reportHotspot, 0, len(modules))
	for _, module := range modules {
		dir, file := path.Split(module.FilePath)
		rows = append(rows, reportHotspot{
			Dir:       dir,
			File:      file,
			Lines:     formatThousands(module.LinesOfCode),
			Functions: module.FunctionCount,
			MaxCC:     module.MaxComplexity,
			MaxCCPct:  module.MaxComplexity * 100 / maxCC,
			MaxCCBand: complexityBand(module.MaxComplexity, low, medium),
			HighRisk:  module.HighRiskFunctionCount,
			DeadCode:  module.DeadCodeFindingCount,
			Clones:    clonesByFile[module.FilePath],
		})
	}
	return rows
}

// countClonesByFile counts clone fragments per file. Groups are authoritative
// when present; otherwise pairs are counted (each side once).
func countClonesByFile(clone *domain.CloneResponse) map[string]int {
	counts := make(map[string]int)
	if clone == nil {
		return counts
	}
	if len(clone.CloneGroups) > 0 {
		for _, group := range clone.CloneGroups {
			for _, fragment := range group.Clones {
				if fragment != nil && fragment.Location != nil {
					counts[fragment.Location.FilePath]++
				}
			}
		}
		return counts
	}
	// A fragment can sit in several pairs; count it once, keyed by its span,
	// to match how CloneStatistics.TotalClones deduplicates.
	seen := make(map[string]struct{})
	for _, pair := range clone.ClonePairs {
		for _, fragment := range []*domain.Clone{pair.Clone1, pair.Clone2} {
			if fragment == nil || fragment.Location == nil {
				continue
			}
			key := fmt.Sprintf("%s:%d-%d", fragment.Location.FilePath, fragment.Location.StartLine, fragment.Location.EndLine)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			counts[fragment.Location.FilePath]++
		}
	}
	return counts
}

func complexityThresholds(complexity *domain.ComplexityResponse) (low, medium int) {
	low, medium = domain.DefaultComplexityLowThreshold, domain.DefaultComplexityMediumThreshold
	if complexity != nil && complexity.Request != nil {
		if complexity.Request.LowThreshold > 0 {
			low = complexity.Request.LowThreshold
		}
		if complexity.Request.MediumThreshold > 0 {
			medium = complexity.Request.MediumThreshold
		}
	}
	return low, medium
}

func complexityBand(cc, low, medium int) string {
	switch {
	case cc > medium:
		return "bad"
	case cc > low:
		return "warn"
	default:
		return ""
	}
}

// Histogram geometry (SVG user units).
const (
	histLeft       = 44.0
	histRight      = 412.0
	histTop        = 30.0
	histBaseline   = 150.0
	histBarFill    = 0.72
	histTickCount  = 3
	histLabelSpace = 6.0
)

func buildReportHistogram(complexity *domain.ComplexityResponse) *reportHistogram {
	if complexity == nil || len(complexity.Functions) == 0 {
		return nil
	}
	low, medium := complexityThresholds(complexity)

	defs := histogramBins(low, medium)

	counts := make([]int, len(defs))
	ccs := make([]int, 0, len(complexity.Functions))
	slocs := make([]int, 0, len(complexity.Functions))
	var deepest, longest *domain.FunctionComplexity
	for i := range complexity.Functions {
		function := &complexity.Functions[i]
		cc := function.Metrics.Complexity
		ccs = append(ccs, cc)
		slocs = append(slocs, function.Metrics.SLOC)
		for b, def := range defs {
			if def.upper == 0 || cc <= def.upper {
				counts[b]++
				break
			}
		}
		if function.Name == domain.ModuleFunctionName {
			continue
		}
		if deepest == nil || function.Metrics.NestingDepth > deepest.Metrics.NestingDepth {
			deepest = function
		}
		if longest == nil || function.Metrics.SLOC > longest.Metrics.SLOC {
			longest = function
		}
	}

	maxCount := 1
	for _, count := range counts {
		maxCount = max(maxCount, count)
	}
	niceMax := niceCeiling(maxCount)
	scale := (histBaseline - histTop) / float64(niceMax)

	slot := (histRight - histLeft) / float64(len(defs))
	barWidth := slot * histBarFill
	hist := &reportHistogram{Total: formatThousands(len(complexity.Functions))}
	for i, def := range defs {
		height := float64(counts[i]) * scale
		if counts[i] > 0 && height < 1 {
			height = 1
		}
		x := histLeft + slot*float64(i) + (slot-barWidth)/2
		hist.Bins = append(hist.Bins, reportHistogramBin{
			Label:  def.label,
			Count:  counts[i],
			X:      round1(x),
			Y:      round1(histBaseline - height),
			Width:  round1(barWidth),
			Height: round1(height),
			Band:   def.band,
		})
		if def.band == "warn" {
			hist.ThresholdX = round1(histLeft + slot*float64(i) - histLabelSpace/2)
			hist.Threshold = fmt.Sprintf("risk from CC %d", low+1)
		}
	}
	for i := 0; i <= histTickCount; i++ {
		value := niceMax * i / histTickCount
		hist.Ticks = append(hist.Ticks, reportHistogramTick{
			Label: formatThousands(value),
			Y:     round1(histBaseline - float64(value)*scale),
		})
	}

	hist.Facts = append(hist.Facts, reportKV{Key: "Median function", Value: fmt.Sprintf("CC %s, %s SLOC", formatMedian(median(ccs)), formatMedian(median(slocs)))})
	if deepest != nil {
		hist.Facts = append(hist.Facts, reportKV{Key: "Deepest nesting", Value: fmt.Sprintf("%d levels (%s)", deepest.Metrics.NestingDepth, deepest.Name)})
	}
	if longest != nil {
		hist.Facts = append(hist.Facts, reportKV{Key: "Longest function", Value: fmt.Sprintf("%d SLOC (%s)", longest.Metrics.SLOC, longest.Name)})
	}
	return hist
}

type histogramBin struct {
	label string
	upper int    // inclusive; 0 means open-ended
	band  string // "", "warn", "bad"
}

// histogramBins builds bins that follow the configured risk thresholds:
// 1 | 2–5 | 6–low | low+1–medium | medium+1+, collapsing bins that the
// thresholds make empty (for example low_threshold = 1 removes 2–5 and 6–low).
func histogramBins(low, medium int) []histogramBin {
	if medium <= low {
		medium = low + 1
	}
	candidates := []int{1, low, medium}
	if low > 5 {
		candidates = []int{1, 5, low, medium}
	}
	uppers := []int{}
	for _, upper := range candidates {
		if len(uppers) == 0 || upper > uppers[len(uppers)-1] {
			uppers = append(uppers, upper)
		}
	}
	bins := make([]histogramBin, 0, len(uppers)+1)
	prev := 0
	for _, upper := range uppers {
		bin := histogramBin{upper: upper, band: complexityBand(upper, low, medium)}
		if upper == prev+1 {
			bin.label = fmt.Sprintf("%d", upper)
		} else {
			bin.label = fmt.Sprintf("%d–%d", prev+1, upper)
		}
		bins = append(bins, bin)
		prev = upper
	}
	return append(bins, histogramBin{label: fmt.Sprintf("%d+", prev+1), band: "bad"})
}

// niceCeiling rounds n up to a tidy axis maximum (1, 2, 5 × 10^k) so ticks
// land on round numbers.
func niceCeiling(n int) int {
	if n <= 3 {
		return 3
	}
	magnitude := 1
	for n/magnitude >= 10 {
		magnitude *= 10
	}
	for _, step := range []int{1, 2, 3, 5, 6, 10} {
		if candidate := step * magnitude; candidate >= n {
			return candidate
		}
	}
	return magnitude * 10
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}

func median(values []int) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]int, len(values))
	copy(sorted, values)
	sort.Ints(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return float64(sorted[mid])
	}
	return float64(sorted[mid-1]+sorted[mid]) / 2
}

// formatMedian prints whole medians without a decimal and half values with one.
func formatMedian(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func buildReportDuplication(response *domain.AnalyzeResponse) *reportDuplication {
	if !response.Summary.CloneEnabled || response.Clone == nil || response.Clone.Statistics == nil {
		return nil
	}
	stats := response.Clone.Statistics
	dup := &reportDuplication{Percent: response.Summary.CodeDuplication, Fragments: stats.TotalFragments}

	byType := make(map[string]int)
	files := make(map[string]struct{})
	unit := "fragments"
	if len(response.Clone.CloneGroups) > 0 {
		for _, group := range response.Clone.CloneGroups {
			for _, fragment := range group.Clones {
				if fragment == nil {
					continue
				}
				byType[group.Type.String()]++
				if fragment.Location != nil {
					files[fragment.Location.FilePath] = struct{}{}
				}
			}
		}
	} else {
		unit = "pairs"
		for _, pair := range response.Clone.ClonePairs {
			byType[pair.Type.String()]++
			for _, fragment := range []*domain.Clone{pair.Clone1, pair.Clone2} {
				if fragment != nil && fragment.Location != nil {
					files[fragment.Location.FilePath] = struct{}{}
				}
			}
		}
	}
	total := 0
	for _, count := range byType {
		total += count
	}
	for _, cloneType := range []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone, domain.Type4Clone} {
		count := byType[cloneType.String()]
		if count == 0 {
			continue
		}
		dup.Types = append(dup.Types, reportShare{
			Label:   fmt.Sprintf("%s %s · %d %s", cloneType.String(), cloneTypeNoun(cloneType), count, unit),
			Percent: float64(count) * 100 / float64(total),
			Class:   fmt.Sprintf("t%d", int(cloneType)),
		})
	}

	dup.Facts = []reportKV{
		{Key: "Clone groups", Value: formatThousands(stats.TotalCloneGroups)},
		{Key: "Clone pairs", Value: formatThousands(stats.TotalClonePairs)},
	}
	if stats.TotalClonePairs > 0 || stats.TotalCloneGroups > 0 {
		dup.Facts = append(dup.Facts, reportKV{Key: "Avg similarity", Value: fmt.Sprintf("%.2f", stats.AverageSimilarity)})
	}
	if stats.FilesAnalyzed > 0 {
		dup.Facts = append(dup.Facts, reportKV{Key: "Files with clones", Value: fmt.Sprintf("%d of %d", len(files), stats.FilesAnalyzed)})
	}
	return dup
}

func cloneTypeNoun(cloneType domain.CloneType) string {
	switch cloneType {
	case domain.Type1Clone:
		return "identical"
	case domain.Type2Clone:
		return "renamed"
	case domain.Type3Clone:
		return "modified"
	case domain.Type4Clone:
		return "semantic"
	default:
		return ""
	}
}

func buildReportClasses(response *domain.AnalyzeResponse) *reportClasses {
	s := response.Summary
	if !s.CBOEnabled && !s.LCOMEnabled {
		return nil
	}
	classes := &reportClasses{Total: max(s.CBOClasses, s.LCOMClasses)}
	if s.CBOEnabled && response.CBO != nil {
		classes.Facts = append(classes.Facts,
			reportKV{Key: "High coupling", Value: formatThousands(response.CBO.Summary.HighRiskClasses), Band: warnIfPositive(response.CBO.Summary.HighRiskClasses)},
			reportKV{Key: "Medium coupling", Value: formatThousands(response.CBO.Summary.MediumRiskClasses)},
		)
	}
	if s.LCOMEnabled && response.LCOM != nil {
		classes.Facts = append(classes.Facts,
			reportKV{Key: "Low cohesion", Value: formatThousands(response.LCOM.Summary.HighRiskClasses), Band: warnIfPositive(response.LCOM.Summary.HighRiskClasses)},
		)
	}
	if s.CBOEnabled && response.CBO != nil {
		if top := mostCoupledClass(response.CBO.Classes); top != nil {
			classes.Facts = append(classes.Facts, reportKV{Key: "Most coupled", Value: fmt.Sprintf("%s (%d)", top.Name, top.Metrics.CouplingCount), Mono: true})
		}
	}
	if s.LCOMEnabled && response.LCOM != nil {
		if top := leastCohesiveClass(response.LCOM.Classes); top != nil {
			classes.Facts = append(classes.Facts, reportKV{Key: "Least cohesive", Value: fmt.Sprintf("%s (%d)", top.Name, top.Metrics.LCOM4), Mono: true})
		}
	}
	return classes
}

func warnIfPositive(n int) string {
	if n > 0 {
		return "warn"
	}
	return "good"
}

func mostCoupledClass(classes []domain.ClassCoupling) *domain.ClassCoupling {
	var top *domain.ClassCoupling
	for i := range classes {
		if top == nil || classes[i].Metrics.CouplingCount > top.Metrics.CouplingCount {
			top = &classes[i]
		}
	}
	return top
}

func leastCohesiveClass(classes []domain.ClassCohesion) *domain.ClassCohesion {
	var top *domain.ClassCohesion
	for i := range classes {
		if top == nil || classes[i].Metrics.LCOM4 > top.Metrics.LCOM4 {
			top = &classes[i]
		}
	}
	return top
}

func buildReportStructure(response *domain.AnalyzeResponse) *reportStructure {
	if !reportHasArchitecture(response) {
		return nil
	}
	structure := &reportStructure{}
	if response.System != nil {
		if deps := response.System.DependencyAnalysis; deps != nil {
			if deps.CircularDependencies != nil {
				structure.Cycles = deps.CircularDependencies.TotalCycles
			}
			structure.Facts = append(structure.Facts,
				reportKV{Key: "Modules / edges", Value: fmt.Sprintf("%d / %d", deps.TotalModules, deps.TotalDependencies)},
				reportKV{Key: "Max dependency depth", Value: formatThousands(deps.MaxDepth)},
			)
		}
		if arch := response.System.ArchitectureAnalysis; arch != nil {
			band := "good"
			if arch.TotalViolations > 0 {
				band = "warn"
			}
			structure.Facts = append(structure.Facts,
				reportKV{Key: "Layer compliance", Value: fmt.Sprintf("%.0f%%", arch.ComplianceScore*100), Band: band},
			)
		}
	}
	if response.Summary.CommunitiesEnabled && response.Communities != nil {
		communities := response.Communities
		structure.Facts = append(structure.Facts,
			reportKV{Key: fmt.Sprintf("Communities (%s)", communities.Algorithm), Value: fmt.Sprintf("%d, Q %.2f", communities.TotalCommunities, communities.Modularity)},
			reportKV{Key: "Bridge modules", Value: formatThousands(len(communities.BridgeModules))},
		)
	}
	return structure
}
