package service

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
)

const (
	defaultMinPackageCohesion  = domain.DefaultArchitectureMinCohesion
	defaultMaxResponsibilities = domain.DefaultArchitectureMaxResponsibilities
)

type responsibilityOptions struct {
	minPackageCohesion  float64
	maxResponsibilities int
	cohesionSeverity    domain.ViolationSeverity
	severity            domain.ViolationSeverity
}

func defaultResponsibilityOptions() responsibilityOptions {
	return responsibilityOptions{
		minPackageCohesion:  defaultMinPackageCohesion,
		maxResponsibilities: defaultMaxResponsibilities,
		cohesionSeverity:    domain.ViolationSeverityWarning,
		severity:            domain.ViolationSeverityWarning,
	}
}

func responsibilityOptionsFromRequest(req domain.SystemAnalysisRequest) responsibilityOptions {
	options := defaultResponsibilityOptions()
	if req.MinCohesion > 0 {
		options.minPackageCohesion = req.MinCohesion
	}
	if req.MaxResponsibilities > 0 {
		options.maxResponsibilities = req.MaxResponsibilities
	}
	if req.CohesionViolationSeverity != "" {
		options.cohesionSeverity = req.CohesionViolationSeverity
	}
	if req.ResponsibilityViolationSeverity != "" {
		options.severity = req.ResponsibilityViolationSeverity
	}
	return options
}

func parseViolationSeverity(value string) domain.ViolationSeverity {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(domain.ViolationSeverityInfo):
		return domain.ViolationSeverityInfo
	case string(domain.ViolationSeverityError):
		return domain.ViolationSeverityError
	case string(domain.ViolationSeverityCritical):
		return domain.ViolationSeverityCritical
	default:
		return domain.ViolationSeverityWarning
	}
}

func (s *SystemAnalysisServiceImpl) analyzeResponsibilityForRequest(
	graph *analyzer.DependencyGraph,
	req domain.SystemAnalysisRequest,
) (*domain.ResponsibilityAnalysis, *domain.CohesionAnalysis, []domain.ArchitectureViolation) {
	if !domain.BoolValue(req.ValidateResponsibility, true) && !domain.BoolValue(req.ValidateCohesion, true) {
		return nil, nil, nil
	}

	options := responsibilityOptionsFromRequest(req)
	responsibility, cohesion, responsibilityViolations := s.analyzeResponsibility(graph, options)
	violations := make([]domain.ArchitectureViolation, 0, len(responsibilityViolations)+len(cohesion.LowCohesionPackages))
	if !domain.BoolValue(req.ValidateResponsibility, true) {
		responsibility = nil
	} else {
		violations = append(violations, responsibilityViolations...)
	}
	if !domain.BoolValue(req.ValidateCohesion, true) {
		cohesion = nil
	} else {
		violations = append(violations, cohesionArchitectureViolations(cohesion, options.cohesionSeverity)...)
	}
	return responsibility, cohesion, violations
}

func responsibilitySeverityCounts(violations []domain.ArchitectureViolation) map[domain.ViolationSeverity]int {
	counts := make(map[domain.ViolationSeverity]int)
	for _, violation := range violations {
		counts[violation.Severity]++
	}
	return counts
}

func cohesionArchitectureViolations(cohesion *domain.CohesionAnalysis, severity domain.ViolationSeverity) []domain.ArchitectureViolation {
	if cohesion == nil || len(cohesion.LowCohesionPackages) == 0 {
		return nil
	}
	if severity == "" {
		severity = domain.ViolationSeverityWarning
	}

	violations := make([]domain.ArchitectureViolation, 0, len(cohesion.LowCohesionPackages))
	for _, pkg := range cohesion.LowCohesionPackages {
		score := cohesion.PackageCohesion[pkg]
		suggestion := cohesion.CohesionSuggestions[pkg]
		violations = append(violations, domain.ArchitectureViolation{
			Type:        domain.ViolationTypeCohesion,
			Severity:    severity,
			Module:      pkg,
			Rule:        "package-cohesion",
			Description: fmt.Sprintf("Package '%s' has low cohesion (%.2f)", pkg, score),
			Suggestion:  suggestion,
		})
	}
	return violations
}

func (s *SystemAnalysisServiceImpl) analyzeResponsibility(
	graph *analyzer.DependencyGraph,
	options responsibilityOptions,
) (*domain.ResponsibilityAnalysis, *domain.CohesionAnalysis, []domain.ArchitectureViolation) {
	if options.minPackageCohesion <= 0 {
		options.minPackageCohesion = defaultMinPackageCohesion
	}
	if options.maxResponsibilities <= 0 {
		options.maxResponsibilities = defaultMaxResponsibilities
	}

	fanInLimit, fanOutLimit := responsibilityCouplingLimits(graph)
	moduleResponsibilities := make(map[string][]string, len(graph.Nodes))
	violations := make([]domain.SRPViolation, 0)
	architectureViolations := make([]domain.ArchitectureViolation, 0)
	overloaded := make([]string, 0)

	for _, module := range graph.GetModuleNames() {
		node := graph.Nodes[module]
		responsibilities := inferResponsibilities(module, node)
		moduleResponsibilities[module] = responsibilities

		isHub := node.InDegree >= fanInLimit && node.OutDegree >= fanOutLimit && node.InDegree > 0 && node.OutDegree > 0
		isOverloaded := len(responsibilities) > options.maxResponsibilities
		if !isHub && !isOverloaded {
			continue
		}

		overloaded = append(overloaded, module)
		severity := responsibilitySeverity(options.severity, len(responsibilities), options.maxResponsibilities, isHub)
		suggestion := fmt.Sprintf("Split '%s' by concern so each module owns one stable responsibility boundary", module)
		violations = append(violations, domain.SRPViolation{
			Module:           module,
			Responsibilities: responsibilities,
			Severity:         severity,
			Suggestion:       suggestion,
		})
		architectureViolations = append(architectureViolations, domain.ArchitectureViolation{
			Type:        domain.ViolationTypeResponsibility,
			Severity:    severity,
			Module:      module,
			Rule:        "single-responsibility",
			Description: fmt.Sprintf("Module '%s' mixes %d dependency concerns with fan-in %d and fan-out %d", module, len(responsibilities), node.InDegree, node.OutDegree),
			Suggestion:  suggestion,
		})
	}

	sort.SliceStable(violations, func(i, j int) bool {
		return violations[i].Module < violations[j].Module
	})
	sort.Strings(overloaded)

	return &domain.ResponsibilityAnalysis{
		SRPViolations:          violations,
		ModuleResponsibilities: moduleResponsibilities,
		OverloadedModules:      overloaded,
	}, analyzePackageCohesion(graph, options.minPackageCohesion), architectureViolations
}

func responsibilityCouplingLimits(graph *analyzer.DependencyGraph) (int, int) {
	moduleCount := len(graph.Nodes)
	if moduleCount == 0 {
		return 1, 1
	}

	var totalIn, totalOut float64
	for _, node := range graph.Nodes {
		totalIn += float64(node.InDegree)
		totalOut += float64(node.OutDegree)
	}

	meanIn := totalIn / float64(moduleCount)
	meanOut := totalOut / float64(moduleCount)
	stdIn := couplingStdDev(graph, meanIn, true)
	stdOut := couplingStdDev(graph, meanOut, false)

	return maxSystemAnalysis(2, int(math.Ceil(meanIn+stdIn))), maxSystemAnalysis(2, int(math.Ceil(meanOut+stdOut)))
}

func couplingStdDev(graph *analyzer.DependencyGraph, mean float64, fanIn bool) float64 {
	if len(graph.Nodes) == 0 {
		return 0
	}

	var sum float64
	for _, node := range graph.Nodes {
		value := float64(node.OutDegree)
		if fanIn {
			value = float64(node.InDegree)
		}
		diff := value - mean
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(graph.Nodes)))
}

func inferResponsibilities(module string, node *analyzer.ModuleNode) []string {
	labels := make(map[string]bool)
	for dependency := range node.Dependencies {
		if label := concernLabel(module, dependency); label != "" {
			labels[label] = true
		}
	}
	for dependent := range node.Dependents {
		if label := concernLabel(module, dependent); label != "" {
			labels[label] = true
		}
	}

	result := make([]string, 0, len(labels))
	for label := range labels {
		result = append(result, label)
	}
	sort.Strings(result)
	return result
}

func concernLabel(module, neighbor string) string {
	moduleParts := splitModuleName(module)
	neighborParts := splitModuleName(neighbor)
	limit := minSystemAnalysis(len(moduleParts), len(neighborParts))

	i := 0
	for i < limit && moduleParts[i] == neighborParts[i] {
		i++
	}
	if i < len(neighborParts) {
		return neighborParts[i]
	}
	if len(neighborParts) > 0 {
		return neighborParts[len(neighborParts)-1]
	}
	return ""
}

func splitModuleName(module string) []string {
	parts := strings.Split(module, ".")
	result := parts[:0]
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func responsibilitySeverity(
	defaultSeverity domain.ViolationSeverity,
	responsibilityCount int,
	maxResponsibilities int,
	isHub bool,
) domain.ViolationSeverity {
	if isHub && responsibilityCount > maxResponsibilities+1 {
		return domain.ViolationSeverityError
	}
	if defaultSeverity != "" {
		return defaultSeverity
	}
	return domain.ViolationSeverityWarning
}

func analyzePackageCohesion(graph *analyzer.DependencyGraph, minCohesion float64) *domain.CohesionAnalysis {
	type packageStats struct {
		modules map[string]bool
		intra   int
		inter   int
	}

	stats := make(map[string]*packageStats)
	for module := range graph.Nodes {
		pkg := packageNameForModule(module)
		if pkg == "" {
			continue
		}
		if stats[pkg] == nil {
			stats[pkg] = &packageStats{modules: make(map[string]bool)}
		}
		stats[pkg].modules[module] = true
	}

	for _, edge := range graph.Edges {
		fromPkg := packageNameForModule(edge.From)
		toPkg := packageNameForModule(edge.To)
		if fromPkg == "" || stats[fromPkg] == nil {
			continue
		}
		if fromPkg == toPkg {
			stats[fromPkg].intra++
		} else {
			stats[fromPkg].inter++
		}
	}

	cohesion := make(map[string]float64, len(stats))
	lowCohesion := make([]string, 0)
	suggestions := make(map[string]string)

	for pkg, stat := range stats {
		total := stat.intra + stat.inter
		score := 1.0
		if total > 0 {
			score = float64(stat.intra) / float64(total)
		}
		cohesion[pkg] = score

		if len(stat.modules) > 1 && total > 0 && score < minCohesion {
			lowCohesion = append(lowCohesion, pkg)
			suggestions[pkg] = fmt.Sprintf("Move unrelated dependencies out of '%s' or split it around cohesive dependency groups", pkg)
		}
	}

	sort.Strings(lowCohesion)
	return &domain.CohesionAnalysis{
		PackageCohesion:     cohesion,
		LowCohesionPackages: lowCohesion,
		CohesionSuggestions: suggestions,
	}
}

func packageNameForModule(module string) string {
	parts := splitModuleName(module)
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

func maxSystemAnalysis(a, b int) int {
	if a > b {
		return a
	}
	return b
}
