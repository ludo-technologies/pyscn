package service

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// SystemAnalysisServiceImpl implements the SystemAnalysisService interface
type SystemAnalysisServiceImpl struct {
	parser *parser.Parser
}

// NewSystemAnalysisService creates a new system analysis service implementation
func NewSystemAnalysisService() *SystemAnalysisServiceImpl {
	return &SystemAnalysisServiceImpl{
		parser: parser.New(),
	}
}

// Analyze performs comprehensive system analysis including dependencies, architecture, and quality metrics
func (s *SystemAnalysisServiceImpl) Analyze(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.SystemAnalysisResponse, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var allResults []interface{}
	var warnings []string
	var errors []string
	startTime := time.Now()
	analyzeDependencies := domain.BoolValue(req.AnalyzeDependencies, true)
	analyzeArchitecture := domain.BoolValue(req.AnalyzeArchitecture, true)

	var graph *analyzer.DependencyGraph
	if analyzeDependencies || analyzeArchitecture {
		var err error
		graph, err = s.buildDependencyGraph(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Module graph failed: %v", err))
		}
	}

	// Analyze dependencies if requested
	var dependencyResult *domain.DependencyAnalysisResult
	if analyzeDependencies && graph != nil {
		result, err := s.buildDependencyAnalysisResult(ctx, graph)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Dependency analysis failed: %v", err))
		} else {
			dependencyResult = result
			allResults = append(allResults, result)
		}
	}

	// Analyze architecture if requested
	var architectureResult *domain.ArchitectureAnalysisResult
	if analyzeArchitecture && graph != nil {
		result, err := s.analyzeArchitectureGraph(ctx, graph, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Architecture analysis failed: %v", err))
		} else {
			architectureResult = result
			allResults = append(allResults, result)
		}
	}

	// If all analyses failed, return an error
	if len(allResults) == 0 {
		return nil, fmt.Errorf("all requested analyses failed: %v", strings.Join(errors, "; "))
	}

	// Build comprehensive response
	response := &domain.SystemAnalysisResponse{
		DependencyAnalysis:   dependencyResult,
		ArchitectureAnalysis: architectureResult,
		Summary:              s.buildSystemAnalysisSummary(graph, dependencyResult, architectureResult),
		GeneratedAt:          time.Now(),
		Duration:             time.Since(startTime).Milliseconds(),
		Version:              version.Version,
		Warnings:             warnings,
		Errors:               errors,
	}

	return response, nil
}

func (s *SystemAnalysisServiceImpl) buildSystemAnalysisSummary(
	graph *analyzer.DependencyGraph,
	dependencyResult *domain.DependencyAnalysisResult,
	architectureResult *domain.ArchitectureAnalysisResult,
) domain.SystemAnalysisSummary {
	summary := domain.SystemAnalysisSummary{}

	if graph != nil {
		summary.ProjectRoot = graph.ProjectRoot
		summary.TotalModules = graph.TotalModules
		summary.TotalDependencies = graph.TotalEdges
		summary.TotalPackages = len(graph.GetPackages())
	}

	refactoringCandidates := make(map[string]struct{})

	if dependencyResult != nil {
		summary.TotalModules = dependencyResult.TotalModules
		summary.TotalDependencies = dependencyResult.TotalDependencies
		if packageCount := countSystemAnalysisPackages(dependencyResult.ModuleMetrics); packageCount > 0 {
			summary.TotalPackages = packageCount
		}

		if dependencyResult.CouplingAnalysis != nil {
			summary.AverageCoupling = dependencyResult.CouplingAnalysis.AverageCoupling
			summary.AverageInstability = dependencyResult.CouplingAnalysis.AverageInstability
			for _, module := range dependencyResult.CouplingAnalysis.HighlyCoupledModules {
				refactoringCandidates[module] = struct{}{}
			}
		}

		if dependencyResult.CircularDependencies != nil {
			summary.CyclicDependencies = dependencyResult.CircularDependencies.TotalModulesInCycles
			summary.CriticalIssues += countCriticalCycles(dependencyResult.CircularDependencies)
		}

		for module, metrics := range dependencyResult.ModuleMetrics {
			if metrics == nil {
				continue
			}
			if metrics.RiskLevel == domain.RiskLevelHigh {
				summary.HighRiskModules++
				refactoringCandidates[module] = struct{}{}
			}
		}
	}

	if architectureResult != nil {
		architectureViolations, criticalArchitectureViolations := countArchitectureViolations(architectureResult)
		summary.ArchitectureScore = clampSystemScore(architectureResult.ComplianceScore * 100)
		summary.ArchitectureViolations = architectureViolations
		summary.CriticalIssues += criticalArchitectureViolations
		summary.ArchitectureImprovements = len(architectureResult.Recommendations)
		for _, module := range architectureResult.RefactoringTargets {
			refactoringCandidates[module] = struct{}{}
		}
	}

	summary.RefactoringCandidates = len(refactoringCandidates)
	summary.MaintainabilityScore = s.calculateSystemMaintainabilityScore(summary, dependencyResult)
	summary.ModularityScore = s.calculateSystemModularityScore(summary, graph)
	summary.TechnicalDebtHours = s.estimateSystemTechnicalDebtHours(summary, dependencyResult)
	summary.OverallQualityScore = calculateOverallSystemQualityScore(summary, dependencyResult != nil, architectureResult != nil)

	return summary
}

func countSystemAnalysisPackages(moduleMetrics map[string]*domain.ModuleDependencyMetrics) int {
	if len(moduleMetrics) == 0 {
		return 0
	}

	packages := make(map[string]struct{})
	for _, metrics := range moduleMetrics {
		if metrics == nil || metrics.Package == "" {
			continue
		}
		packages[metrics.Package] = struct{}{}
	}
	return len(packages)
}

func countCriticalCycles(cycles *domain.CircularDependencyAnalysis) int {
	count := 0
	for _, cycle := range cycles.CircularDependencies {
		if cycle.Severity == domain.CycleSeverityCritical {
			count++
		}
	}
	return count
}

func countArchitectureViolations(result *domain.ArchitectureAnalysisResult) (int, int) {
	if result == nil {
		return 0, 0
	}

	if len(result.Violations) > 0 {
		critical := 0
		for _, violation := range result.Violations {
			if violation.Severity == domain.ViolationSeverityCritical {
				critical++
			}
		}
		return len(result.Violations), critical
	}

	total := result.TotalViolations
	critical := 0
	if result.SeverityBreakdown != nil {
		breakdownTotal := 0
		for severity, count := range result.SeverityBreakdown {
			breakdownTotal += count
			if severity == domain.ViolationSeverityCritical {
				critical = count
			}
		}
		if breakdownTotal > total {
			total = breakdownTotal
		}
	}

	return total, critical
}

func (s *SystemAnalysisServiceImpl) calculateSystemMaintainabilityScore(
	summary domain.SystemAnalysisSummary,
	dependencyResult *domain.DependencyAnalysisResult,
) float64 {
	if dependencyResult != nil {
		total := 0.0
		count := 0
		for _, metrics := range dependencyResult.ModuleMetrics {
			if metrics == nil || metrics.Maintainability <= 0 {
				continue
			}
			total += clampSystemScore(metrics.Maintainability)
			count++
		}
		if count > 0 {
			return roundSystemScore(total / float64(count))
		}
	}

	if summary.TotalModules == 0 {
		return 0
	}

	score := 100.0
	score -= math.Min(30, summary.AverageCoupling*8)
	score -= math.Min(25, moduleRatio(summary.HighRiskModules, summary.TotalModules)*25)
	score -= math.Min(20, moduleRatio(summary.CyclicDependencies, summary.TotalModules)*20)
	score -= math.Min(25, moduleRatio(summary.ArchitectureViolations, summary.TotalModules)*10)
	return roundSystemScore(clampSystemScore(score))
}

func (s *SystemAnalysisServiceImpl) calculateSystemModularityScore(
	summary domain.SystemAnalysisSummary,
	graph *analyzer.DependencyGraph,
) float64 {
	if graph != nil && graph.SystemMetrics != nil && graph.SystemMetrics.ModularityIndex > 0 {
		return roundSystemScore(clampSystemScore(graph.SystemMetrics.ModularityIndex * 100))
	}

	if summary.TotalModules == 0 {
		return 0
	}

	score := 100.0
	score -= math.Min(40, summary.AverageCoupling*10)
	score -= math.Min(40, moduleRatio(summary.CyclicDependencies, summary.TotalModules)*40)
	if summary.TotalPackages <= 1 && summary.TotalModules > 1 {
		score -= 10
	}
	return roundSystemScore(clampSystemScore(score))
}

func (s *SystemAnalysisServiceImpl) estimateSystemTechnicalDebtHours(
	summary domain.SystemAnalysisSummary,
	dependencyResult *domain.DependencyAnalysisResult,
) float64 {
	if dependencyResult != nil {
		total := 0.0
		for _, metrics := range dependencyResult.ModuleMetrics {
			if metrics != nil {
				total += metrics.TechnicalDebt
			}
		}
		if total > 0 {
			return roundSystemScore(total)
		}
	}

	debt := float64(summary.HighRiskModules*4 +
		summary.CyclicDependencies*2 +
		summary.ArchitectureViolations +
		summary.RefactoringCandidates*2 +
		summary.CriticalIssues*4)
	return roundSystemScore(debt)
}

func calculateOverallSystemQualityScore(
	summary domain.SystemAnalysisSummary,
	hasDependencyData bool,
	hasArchitectureData bool,
) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	if hasDependencyData {
		totalScore += summary.MaintainabilityScore * 0.4
		totalScore += summary.ModularityScore * 0.3
		totalWeight += 0.7
	}
	if hasArchitectureData {
		totalScore += summary.ArchitectureScore * 0.3
		totalWeight += 0.3
	}
	if totalWeight == 0 {
		return 0
	}
	return roundSystemScore(clampSystemScore(totalScore / totalWeight))
}

func moduleRatio(count, total int) float64 {
	if total <= 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func clampSystemScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func roundSystemScore(score float64) float64 {
	return math.Round(score*100) / 100
}

// AnalyzeDependencies performs dependency analysis only
func (s *SystemAnalysisServiceImpl) AnalyzeDependencies(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.DependencyAnalysisResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	graph, err := s.buildDependencyGraph(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.buildDependencyAnalysisResult(ctx, graph)
}

func (s *SystemAnalysisServiceImpl) buildDependencyAnalysisResult(ctx context.Context, graph *analyzer.DependencyGraph) (*domain.DependencyAnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Check if any modules were processed
	if graph.TotalModules == 0 {
		return &domain.DependencyAnalysisResult{
			TotalModules:      0,
			TotalDependencies: 0,
			RootModules:       []string{},
			LeafModules:       []string{},
			MaxDepth:          0,
		}, nil
	}

	// Detect circular dependencies
	circularDetector := analyzer.NewCircularDependencyDetector(graph)
	circularResult := circularDetector.DetectCircularDependencies()
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Calculate coupling metrics
	metricsCalculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
	if err := metricsCalculator.CalculateMetrics(); err != nil {
		return nil, err
	}
	couplingResults := s.extractCouplingResult(graph)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Build dependency matrix
	matrix := s.buildDependencyMatrix(graph)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Find longest dependency chains
	longestChains := s.findLongestChains(graph, 10) // Top 10 chains
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Extract module metrics
	moduleMetrics := s.extractModuleMetrics(graph)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("dependency analysis cancelled: %w", err)
	}

	// Create dependency analysis result
	result := &domain.DependencyAnalysisResult{
		TotalModules:         graph.TotalModules,
		TotalDependencies:    graph.TotalEdges,
		RootModules:          graph.GetRootModules(),
		LeafModules:          graph.GetLeafModules(),
		ModuleMetrics:        moduleMetrics,
		DependencyMatrix:     matrix,
		CircularDependencies: s.convertCircularResults(circularResult),
		CouplingAnalysis:     s.convertCouplingResults(couplingResults),
		LongestChains:        longestChains,
		MaxDepth:             s.calculateMaxDepth(graph),
	}

	return result, nil
}

// AnalyzeArchitecture performs architecture validation only
func (s *SystemAnalysisServiceImpl) AnalyzeArchitecture(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.ArchitectureAnalysisResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	graph, err := s.buildDependencyGraph(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.analyzeArchitectureGraph(ctx, graph, req)
}

func (s *SystemAnalysisServiceImpl) analyzeArchitectureGraph(ctx context.Context, graph *analyzer.DependencyGraph, req domain.SystemAnalysisRequest) (*domain.ArchitectureAnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("architecture analysis cancelled: %w", err)
	}

	responsibilityAnalysis, cohesionAnalysis, responsibilityViolations, responsibilityChecks := s.analyzeResponsibilityForRequest(graph, req)
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("architecture analysis cancelled: %w", err)
	}

	// Clone ArchitectureRules before modifying to avoid mutating the caller's object
	// (the pointer is shared even though SystemAnalysisRequest is passed by value).
	rules := s.resolveArchitectureRules(graph, req.ArchitectureRules)
	if rules == nil || len(rules.Layers) == 0 {
		if responsibilityAnalysis == nil && cohesionAnalysis == nil {
			return s.emptyArchitectureResult(), nil
		}
		severityCounts := responsibilitySeverityCounts(responsibilityViolations)
		checked := responsibilityChecks
		errorCount := severityCounts[domain.ViolationSeverityError]
		warningCount := severityCounts[domain.ViolationSeverityWarning]
		compliance, weighted := s.calculateComplianceWeighted(errorCount, warningCount, checked)
		recommendations := s.generateArchitectureRecommendations(responsibilityViolations, map[string]float64{}, nil, compliance)
		refactoringTargets := s.identifyArchitectureRefactoringTargets(responsibilityViolations, map[string]string{})
		return s.buildArchitectureResultWithRecommendations(
			responsibilityViolations,
			severityCounts,
			map[string]map[string]int{},
			map[string]float64{},
			nil,
			0,
			compliance,
			weighted,
			checked,
			map[string]string{},
			recommendations,
			refactoringTargets,
			cohesionAnalysis,
			responsibilityAnalysis,
		), nil
	}
	req.ArchitectureRules = rules

	// Map modules to layers
	moduleToLayer := s.buildModuleLayerMap(graph, req.ArchitectureRules)

	// Evaluate layer rules and collect violations
	violations, severityCounts, layerCoupling, checked := s.evaluateLayerRules(ctx, graph, moduleToLayer, req.ArchitectureRules)
	if violations == nil {
		// Check if context was cancelled
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("architecture analysis cancelled: %w", ctx.Err())
		default:
			// If not cancelled, return empty result (no violations)
			return s.emptyArchitectureResult(), nil
		}
	}

	// Calculate metrics
	layerCohesion, problematic, layersAnalyzed := s.calculateLayerMetrics(layerCoupling)
	for _, violation := range responsibilityViolations {
		violations = append(violations, violation)
		severityCounts[violation.Severity]++
	}
	checked += responsibilityChecks
	errorCount := severityCounts[domain.ViolationSeverityError]
	warningCount := severityCounts[domain.ViolationSeverityWarning]
	compliance, weighted := s.calculateComplianceWeighted(errorCount, warningCount, checked)

	// Generate architecture recommendations
	recommendations := s.generateArchitectureRecommendations(violations, layerCohesion, problematic, compliance)

	// Identify refactoring targets based on violations
	refactoringTargets := s.identifyArchitectureRefactoringTargets(violations, moduleToLayer)

	// Build result
	return s.buildArchitectureResultWithRecommendations(violations, severityCounts, layerCoupling, layerCohesion,
		problematic, layersAnalyzed, compliance, weighted, checked, moduleToLayer, recommendations, refactoringTargets,
		cohesionAnalysis, responsibilityAnalysis), nil
}

// emptyArchitectureResult returns an empty result when no rules are defined
func (s *SystemAnalysisServiceImpl) emptyArchitectureResult() *domain.ArchitectureAnalysisResult {
	return &domain.ArchitectureAnalysisResult{
		ComplianceScore: 1.0,
		TotalViolations: 0,
		TotalRules:      0,
		LayerAnalysis: &domain.LayerAnalysis{
			LayersAnalyzed:    0,
			LayerViolations:   []domain.LayerViolation{},
			LayerCoupling:     make(map[string]map[string]int),
			LayerCohesion:     make(map[string]float64),
			ProblematicLayers: []string{},
		},
		CohesionAnalysis:       nil,
		ResponsibilityAnalysis: nil,
		Violations:             []domain.ArchitectureViolation{},
		SeverityBreakdown:      make(map[domain.ViolationSeverity]int),
		Recommendations:        []domain.ArchitectureRecommendation{},
		RefactoringTargets:     []string{},
	}
}

// buildDependencyGraph creates the dependency graph using ModuleAnalyzer
func (s *SystemAnalysisServiceImpl) buildDependencyGraph(ctx context.Context, req domain.SystemAnalysisRequest) (*analyzer.DependencyGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("module graph cancelled: %w", err)
	}

	projectRoot := FindProjectRoot(req.Paths)
	options := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		IncludeStdLib:     req.IncludeStdLib,
		IncludeThirdParty: req.IncludeThirdParty,
		FollowRelative:    req.FollowRelative,
		IncludePatterns:   req.IncludePatterns,
		ExcludePatterns:   req.ExcludePatterns,
	}
	ma, err := analyzer.NewModuleAnalyzer(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create module analyzer: %w", err)
	}
	graph, err := ma.AnalyzeFiles(req.Paths)
	if err != nil {
		return nil, fmt.Errorf("failed to build module graph: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("module graph cancelled: %w", err)
	}
	return graph, nil
}

// evaluateLayerRules evaluates all edges against layer rules
func (s *SystemAnalysisServiceImpl) evaluateLayerRules(ctx context.Context, graph *analyzer.DependencyGraph,
	moduleToLayer map[string]string, rules *domain.ArchitectureRules) ([]domain.ArchitectureViolation,
	map[domain.ViolationSeverity]int, map[string]map[string]int, int) {

	layerCoupling := make(map[string]map[string]int)
	violations := make([]domain.ArchitectureViolation, 0)
	severityCounts := make(map[domain.ViolationSeverity]int)
	checked := 0

	for _, edge := range graph.Edges {
		select {
		case <-ctx.Done():
			return nil, nil, nil, 0
		default:
		}
		fromLayer := moduleToLayer[edge.From]
		toLayer := moduleToLayer[edge.To]

		if layerCoupling[fromLayer] == nil {
			layerCoupling[fromLayer] = make(map[string]int)
		}
		layerCoupling[fromLayer][toLayer]++

		if v := s.evaluateLayerEdge(rules, edge.From, edge.To, fromLayer, toLayer); v != nil {
			violations = append(violations, *v)
			severityCounts[v.Severity]++
		}
		checked++
	}

	return violations, severityCounts, layerCoupling, checked
}

// calculateLayerMetrics calculates cohesion and identifies problematic layers
func (s *SystemAnalysisServiceImpl) calculateLayerMetrics(layerCoupling map[string]map[string]int) (
	map[string]float64, []string, int) {

	layerCohesion := make(map[string]float64)
	problematic := make([]string, 0)
	uniqueLayers := make(map[string]bool)

	for layer, targets := range layerCoupling {
		uniqueLayers[layer] = true
		total, intra := 0, 0
		for to, cnt := range targets {
			total += cnt
			if to == layer {
				intra += cnt
			}
		}
		if total > 0 {
			layerCohesion[layer] = float64(intra) / float64(total)
		} else {
			layerCohesion[layer] = 1.0
		}
		if layerCohesion[layer] < 0.5 {
			problematic = append(problematic, layer)
		}
	}

	layersAnalyzed := 0
	for l := range uniqueLayers {
		if l != "" {
			layersAnalyzed++
		}
	}

	// Sort problematic layers for deterministic results
	sort.Strings(problematic)

	return layerCohesion, problematic, layersAnalyzed
}

// calculateComplianceWeighted calculates compliance with severity weights.
// Error = 5 points, Warning = 1 point. Returns the compliance score and the
// weighted violation count used as its numerator, so callers can expose both
// in the result struct and keep ComplianceScore reproducible from JSON.
func (s *SystemAnalysisServiceImpl) calculateComplianceWeighted(errorCount, warningCount, checked int) (float64, int) {
	weighted := errorCount*5 + warningCount*1
	if checked == 0 {
		return 1.0, weighted
	}
	compliance := 1.0 - (float64(weighted) / float64(checked))
	if compliance < 0 {
		compliance = 0
	}
	return compliance, weighted
}

// buildArchitectureResultWithRecommendations constructs the final result with recommendations
func (s *SystemAnalysisServiceImpl) buildArchitectureResultWithRecommendations(
	violations []domain.ArchitectureViolation,
	severityCounts map[domain.ViolationSeverity]int,
	layerCoupling map[string]map[string]int,
	layerCohesion map[string]float64,
	problematic []string,
	layersAnalyzed int,
	compliance float64,
	weightedViolations int,
	checked int,
	moduleToLayer map[string]string,
	recommendations []domain.ArchitectureRecommendation,
	refactoringTargets []string,
	cohesionAnalysis *domain.CohesionAnalysis,
	responsibilityAnalysis *domain.ResponsibilityAnalysis) *domain.ArchitectureAnalysisResult {

	layerAnalysis := &domain.LayerAnalysis{
		LayersAnalyzed:    layersAnalyzed,
		LayerViolations:   s.toLayerViolations(violations, moduleToLayer),
		LayerCoupling:     layerCoupling,
		LayerCohesion:     layerCohesion,
		ProblematicLayers: problematic,
	}

	return &domain.ArchitectureAnalysisResult{
		ComplianceScore:        compliance,
		TotalViolations:        len(violations),
		WeightedViolations:     weightedViolations,
		TotalRules:             checked,
		LayerAnalysis:          layerAnalysis,
		CohesionAnalysis:       cohesionAnalysis,
		ResponsibilityAnalysis: responsibilityAnalysis,
		Violations:             violations,
		SeverityBreakdown:      severityCounts,
		Recommendations:        recommendations,
		RefactoringTargets:     refactoringTargets,
	}
}

// BuildModuleLayerMap maps each module to a layer based on ArchitectureRules.
func (s *SystemAnalysisServiceImpl) BuildModuleLayerMap(graph *analyzer.DependencyGraph, rules *domain.ArchitectureRules) map[string]string {
	return s.buildModuleLayerMap(graph, rules)
}

// ResolveArchitectureRules returns resolved architecture rules with style presets applied.
func (s *SystemAnalysisServiceImpl) ResolveArchitectureRules(graph *analyzer.DependencyGraph, rules *domain.ArchitectureRules) *domain.ArchitectureRules {
	return s.resolveArchitectureRules(graph, rules)
}

// buildModuleLayerMap maps each module to a layer based on ArchitectureRules.
// compiledPattern keeps the compiled regexes and its original pattern with simple specificity info.
// It uses two regexes to distinguish prefix (position 0) and suffix (position 1+) matches.
type compiledPattern struct {
	prefixRe    *regexp.Regexp // matches when the pattern appears at the start of the module path
	suffixRe    *regexp.Regexp // matches when the pattern appears after a dot separator
	original    string
	specificity int // number of dots in original pattern; higher = more specific
}

type modulePatternMatch struct {
	matched       bool
	isPrefix      bool
	boundaryScore int // dot/exact matches outrank underscore-boundary matches
	matchPos      int // byte offset in the module where the match starts; lower = better
}

// matchModule returns whether the pattern matches the module and the match position.
// position 0 means prefix match (higher priority), position 1 means suffix match.
func (cp *compiledPattern) matchModule(module string) modulePatternMatch {
	if cp.prefixRe != nil && cp.prefixRe.MatchString(module) {
		return modulePatternMatch{matched: true, isPrefix: true, boundaryScore: 2, matchPos: 0}
	}
	if cp.suffixRe != nil && cp.suffixRe.MatchString(module) {
		// The suffix regex is anchored (^.+\.<pattern>$), so we locate the
		// matched segment by searching for ".<base>" at a dot boundary.
		// Strip wildcard suffixes so "api.*" searches for ".api".
		base := strings.TrimRight(strings.ToLower(cp.original), "*.")
		if base == "" {
			base = strings.ToLower(cp.original)
		}
		pos := strings.Index(strings.ToLower(module), "."+base)
		if pos >= 0 {
			pos++ // skip the dot to point at the pattern itself
		} else {
			pos = 0
		}
		return modulePatternMatch{matched: true, isPrefix: false, boundaryScore: 2, matchPos: pos}
	}
	if strings.Contains(cp.original, "*") {
		return modulePatternMatch{}
	}

	lowerOriginal := strings.ToLower(cp.original)
	parts := strings.Split(module, ".")
	offset := 0
	for _, part := range parts {
		lowerPart := strings.ToLower(part)
		isPrefix := offset == 0
		if strings.HasPrefix(lowerPart, lowerOriginal+"_") {
			return modulePatternMatch{matched: true, isPrefix: isPrefix, boundaryScore: 1, matchPos: offset}
		}
		if strings.HasSuffix(lowerPart, "_"+lowerOriginal) {
			pos := offset + len(part) - len(cp.original)
			return modulePatternMatch{matched: true, isPrefix: false, boundaryScore: 1, matchPos: pos}
		}
		offset += len(part) + 1
	}

	return modulePatternMatch{}
}

// compileLayerPatterns compiles the package patterns for each layer into
// position-aware regexes for use with findLayerForModule.
func (s *SystemAnalysisServiceImpl) compileLayerPatterns(layers []domain.Layer) map[string][]compiledPattern {
	compiled := make(map[string][]compiledPattern)
	for _, layer := range layers {
		for _, pat := range layer.Packages {
			if cp := s.compileModulePatterns(pat); cp != nil {
				compiled[layer.Name] = append(compiled[layer.Name], *cp)
			}
		}
	}
	return compiled
}

func (s *SystemAnalysisServiceImpl) buildModuleLayerMap(graph *analyzer.DependencyGraph, rules *domain.ArchitectureRules) map[string]string {
	out := make(map[string]string)
	compiled := s.compileLayerPatterns(rules.Layers)
	for module := range graph.Nodes {
		if s.isTestModule(module) {
			out[module] = "unknown"
			continue
		}
		// Strip the first matching neutral prefix before layer matching
		stripped := module
		for _, prefix := range rules.NeutralPrefixes {
			if strings.HasPrefix(stripped, prefix+".") {
				stripped = stripped[len(prefix)+1:]
				break
			}
		}
		out[module] = s.findLayerForModule(stripped, compiled)
		if out[module] == "" {
			out[module] = "unknown"
		}
	}
	return out
}

// findLayerForModule returns the most specific matching layer for a module.
// Tie-breaking priority:
//  1. Higher specificity (more dots in pattern) wins
//  2. Prefix match wins over suffix match (among equal specificity)
//  3. Higher boundary strength wins (dot-segment match > underscore-boundary fallback)
//  4. Earlier match position wins (leftmost segment carries more context)
//  5. Longer pattern string wins
//  6. Alphabetical layer name for determinism
func (s *SystemAnalysisServiceImpl) findLayerForModule(module string, compiled map[string][]compiledPattern) string {
	type match struct {
		layer       string
		pattern     string
		specificity int
		isPrefix    bool // true = prefix match (higher priority)
		boundary    int
		matchPos    int // byte offset where the pattern matched in the module
	}

	var matches []match
	for layer, patterns := range compiled {
		for _, cp := range patterns {
			if m := cp.matchModule(module); m.matched {
				matches = append(matches, match{
					layer:       layer,
					pattern:     cp.original,
					specificity: cp.specificity,
					isPrefix:    m.isPrefix,
					boundary:    m.boundaryScore,
					matchPos:    m.matchPos,
				})
			}
		}
	}

	if len(matches) == 0 {
		return ""
	}

	// Sort matches by priority: specificity > prefix > boundary strength > match position > pattern length > layer name
	// Specificity (dot count) is the primary signal so that "api.v1" always beats
	// "foo" even when "foo" matches at prefix position. Among equal-specificity
	// matches, prefer prefix over suffix to resolve ambiguities like
	// "domain.routers" → domain (prefix) over presentation (suffix "routers").
	sort.Slice(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		// 1. Higher specificity wins
		if a.specificity != b.specificity {
			return a.specificity > b.specificity
		}
		// 2. Prefix match wins (among equal specificity)
		if a.isPrefix != b.isPrefix {
			return a.isPrefix
		}
		// 3. Dot/exact segment matches beat underscore-boundary fallbacks
		if a.boundary != b.boundary {
			return a.boundary > b.boundary
		}
		// 4. Earlier match position wins (leftmost segment carries more context)
		if a.matchPos != b.matchPos {
			return a.matchPos < b.matchPos
		}
		// 5. Longer pattern wins
		if len(a.pattern) != len(b.pattern) {
			return len(a.pattern) > len(b.pattern)
		}
		// 6. Alphabetical layer name
		return a.layer < b.layer
	})

	return matches[0].layer
}

// compileModulePatterns converts simple glob-like patterns to a compiledPattern
// with separate prefix and suffix regexes for position-aware matching.
// For Python modules, pattern "views" should match "views", "views.foo", "views.foo.bar", etc.
func (s *SystemAnalysisServiceImpl) compileModulePatterns(glob string) *compiledPattern {
	if glob == "" {
		return nil
	}

	escaped := regexp.QuoteMeta(glob)
	hasWildcard := strings.Contains(glob, "*")

	var prefixPattern, suffixPattern string
	if hasWildcard {
		wild := strings.ReplaceAll(escaped, "\\*", ".*")
		prefixPattern = fmt.Sprintf("(?i)^%s$", wild)
		suffixPattern = fmt.Sprintf("(?i)^.+\\.%s$", wild)
	} else {
		segment := fmt.Sprintf("%s(?:\\..+)?", escaped)
		prefixPattern = fmt.Sprintf("(?i)^%s$", segment)
		suffixPattern = fmt.Sprintf("(?i)^.+\\.%s$", segment)
	}

	prefixRe, err1 := regexp.Compile(prefixPattern)
	suffixRe, err2 := regexp.Compile(suffixPattern)
	if err1 != nil && err2 != nil {
		return nil
	}
	if err1 != nil {
		prefixRe = nil
	}
	if err2 != nil {
		suffixRe = nil
	}

	return &compiledPattern{
		prefixRe:    prefixRe,
		suffixRe:    suffixRe,
		original:    glob,
		specificity: strings.Count(glob, "."),
	}
}

// compileModulePattern is a convenience wrapper that returns a combined regex
// matching both prefix and suffix positions. Used by autoDetectArchitecture tests.
func (s *SystemAnalysisServiceImpl) compileModulePattern(glob string) *regexp.Regexp {
	cp := s.compileModulePatterns(glob)
	if cp == nil {
		return nil
	}
	// Build a combined regex for backward compatibility
	escaped := regexp.QuoteMeta(glob)
	hasWildcard := strings.Contains(glob, "*")
	var pattern string
	if hasWildcard {
		escaped = strings.ReplaceAll(escaped, "\\*", ".*")
		pattern = fmt.Sprintf("(?i)^(?:%s|.+\\.%s)$", escaped, escaped)
	} else {
		segmentTail := "(?:$|[._].+)"
		prefixPattern := fmt.Sprintf("%s%s", escaped, segmentTail)
		suffixPattern := fmt.Sprintf(".+[._]%s%s", escaped, segmentTail)
		pattern = fmt.Sprintf("(?i)^(?:%s|%s)$", prefixPattern, suffixPattern)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

// autoDetectArchitecture automatically detects architecture patterns from the dependency graph
func (s *SystemAnalysisServiceImpl) autoDetectArchitecture(graph *analyzer.DependencyGraph) *domain.ArchitectureRules {
	// Load layer patterns and rules from the embedded default config
	defaultConfig, err := config.LoadDefaultConfigFromTOML()
	if err != nil {
		// This should never happen since the config is embedded at compile time
		// But if it does, return nil to indicate no auto-detection is possible
		return nil
	}

	// Convert config layer definitions to domain layers and compile patterns
	domainLayers := make([]domain.Layer, 0, len(defaultConfig.Architecture.Layers))
	for _, l := range defaultConfig.Architecture.Layers {
		domainLayers = append(domainLayers, domain.Layer{
			Name:     l.Name,
			Packages: l.Packages,
		})
	}
	compiled := s.compileLayerPatterns(domainLayers)

	// Detect which modules belong to which layer
	moduleToLayer := make(map[string]string)
	layerModules := make(map[string][]string)

	for module := range graph.Nodes {
		if s.isTestModule(module) {
			continue
		}
		layer := s.findLayerForModule(module, compiled)
		if layer != "" {
			moduleToLayer[module] = layer
			layerModules[layer] = append(layerModules[layer], module)
		}
	}

	// If no standard patterns found, return nil
	if len(layerModules) == 0 {
		return nil
	}

	// Build layers configuration
	layers := make([]domain.Layer, 0)
	for layerName, modules := range layerModules {
		// Extract unique package prefixes from modules
		packagePrefixes := s.extractPackagePrefixes(modules)
		if len(packagePrefixes) > 0 {
			layers = append(layers, domain.Layer{
				Name:        layerName,
				Description: fmt.Sprintf("Auto-detected %s layer", layerName),
				Packages:    packagePrefixes,
			})
		}
	}

	// Convert config.LayerRule to domain.LayerRule
	rules := make([]domain.LayerRule, 0, len(defaultConfig.Architecture.Rules))
	for _, rule := range defaultConfig.Architecture.Rules {
		rules = append(rules, domain.LayerRule{
			From:  rule.From,
			Allow: rule.Allow,
			Deny:  rule.Deny,
		})
	}

	return &domain.ArchitectureRules{
		Layers:     layers,
		Rules:      rules,
		StrictMode: false,
	}
}

// resolveArchitectureRules returns a self-contained ArchitectureRules ready for
// evaluation. It clones the caller's rules (if any) to avoid mutation, then fills
// in missing layers or rules from auto-detection / embedded defaults.
func (s *SystemAnalysisServiceImpl) resolveArchitectureRules(graph *analyzer.DependencyGraph, orig *domain.ArchitectureRules) *domain.ArchitectureRules {
	// No user config at all — fully auto-detect
	if orig == nil {
		return s.autoDetectArchitecture(graph)
	}

	// Clone to avoid mutating the caller's object
	resolved := &domain.ArchitectureRules{
		Style:             orig.Style,
		Layers:            append([]domain.Layer(nil), orig.Layers...),
		Rules:             append([]domain.LayerRule(nil), orig.Rules...),
		NeutralPrefixes:   append([]string(nil), orig.NeutralPrefixes...),
		StrictMode:        orig.StrictMode,
		AllowedPatterns:   orig.AllowedPatterns,
		ForbiddenPatterns: orig.ForbiddenPatterns,
	}

	// Settings such as strict_mode can be configured without defining layers or
	// rules. Auto-detect the missing structure while preserving those settings.
	if len(resolved.Layers) == 0 && len(resolved.Rules) == 0 && resolved.Style == "" {
		if autoDetected := s.autoDetectArchitecture(graph); autoDetected != nil {
			resolved.Layers = autoDetected.Layers
			resolved.Rules = autoDetected.Rules
		}
		return resolved
	}

	// A style preset is selected — apply its layers/rules. Explicit user-defined
	// layers/rules take precedence over the preset. An unrecognized style yields
	// no preset; fall through to the explicit-config / auto-detect handling below.
	if resolved.Style != "" {
		presetLayers, presetRules := config.ArchitectureStylePreset(resolved.Style)
		if presetLayers != nil || presetRules != nil {
			presetDomainRules := convertLayerRules(presetRules)
			if len(resolved.Layers) == 0 {
				resolved.Layers = convertLayerDefinitions(presetLayers)
			} else {
				presetDomainRules = s.filterLayerRulesForLayers(presetDomainRules, resolved.Layers)
			}
			if len(resolved.Rules) == 0 {
				resolved.Rules = presetDomainRules
			} else {
				// User provided some rules — merge on top of the preset; user
				// rules win for any matching From value.
				resolved.Rules = s.mergeLayerRules(presetDomainRules, resolved.Rules)
			}
			return resolved
		}
		// Unknown style with no explicit config — fall back to auto-detection.
		if len(resolved.Layers) == 0 && len(resolved.Rules) == 0 {
			return s.autoDetectArchitecture(graph)
		}
	}

	if len(resolved.Layers) > 0 && len(resolved.Rules) == 0 {
		// User provided layers but no rules — load default rules filtered to
		// user-defined layer names.
		resolved.Rules = s.loadDefaultRulesForLayers(resolved.Layers)
	} else if len(resolved.Rules) > 0 && len(resolved.Layers) == 0 {
		// User provided rules but no layers — auto-detect layers, then merge
		// auto-detected default rules with user rules (user rules take precedence).
		autoDetected := s.autoDetectArchitecture(graph)
		if autoDetected != nil {
			resolved.Layers = autoDetected.Layers
			resolved.Rules = s.mergeLayerRules(autoDetected.Rules, resolved.Rules)
		}
	}

	return resolved
}

// mergeLayerRules merges base rules with override rules. For any From value
// that appears in overrides, the override replaces the base rule entirely.
func (s *SystemAnalysisServiceImpl) mergeLayerRules(base, overrides []domain.LayerRule) []domain.LayerRule {
	// Build set of From values that the user explicitly defined
	overrideFroms := make(map[string]struct{}, len(overrides))
	for _, r := range overrides {
		overrideFroms[r.From] = struct{}{}
	}

	// Start with base rules that are NOT overridden by the user
	var merged []domain.LayerRule
	for _, r := range base {
		if _, overridden := overrideFroms[r.From]; !overridden {
			merged = append(merged, r)
		}
	}
	// Append all user overrides
	merged = append(merged, overrides...)
	return merged
}

// filterLayerRulesForLayers returns only rules whose From layer exists in layers.
func (s *SystemAnalysisServiceImpl) filterLayerRulesForLayers(rules []domain.LayerRule, layers []domain.Layer) []domain.LayerRule {
	layerNames := make(map[string]struct{}, len(layers))
	for _, l := range layers {
		layerNames[l.Name] = struct{}{}
	}

	filtered := make([]domain.LayerRule, 0, len(rules))
	for _, rule := range rules {
		if _, ok := layerNames[rule.From]; ok {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// loadDefaultRulesForLayers loads the embedded default layer rules and returns
// only those whose From field matches one of the given layer names. This avoids
// injecting rules that reference built-in layer names (e.g. "presentation") when
// the user's layers use custom names (e.g. "api").
func (s *SystemAnalysisServiceImpl) loadDefaultRulesForLayers(layers []domain.Layer) []domain.LayerRule {
	defaultConfig, err := config.LoadDefaultConfigFromTOML()
	if err != nil {
		return nil
	}

	defaultRules := make([]domain.LayerRule, 0, len(defaultConfig.Architecture.Rules))
	for _, rule := range defaultConfig.Architecture.Rules {
		defaultRules = append(defaultRules, domain.LayerRule{
			From:  rule.From,
			Allow: rule.Allow,
			Deny:  rule.Deny,
			Warn:  rule.Warn,
		})
	}
	return s.filterLayerRulesForLayers(defaultRules, layers)
}

// isTestModule checks if a module represents test code
func (s *SystemAnalysisServiceImpl) isTestModule(module string) bool {
	parts := strings.Split(module, ".")
	for _, part := range parts {
		lowerPart := strings.ToLower(part)
		// Check for test directory/package names
		if lowerPart == "tests" || lowerPart == "test" || lowerPart == "testing" || lowerPart == "conftest" {
			return true
		}
	}
	// Check if the last part indicates a test file
	if len(parts) > 0 {
		lastPart := strings.ToLower(parts[len(parts)-1])
		if strings.HasPrefix(lastPart, "test_") || strings.HasSuffix(lastPart, "_test") {
			return true
		}
	}
	return false
}

// isArchitecturalComponent checks if a module part represents an architectural component
func (s *SystemAnalysisServiceImpl) isArchitecturalComponent(part string) bool {
	architecturalKeywords := []string{
		// Presentation layer
		"api", "apis", "views", "view", "controllers", "controller", "routes", "route",
		"handlers", "handler", "ui", "web", "rest", "graphql", "endpoints", "endpoint",
		"routers", "router",
		// Application layer
		"services", "service", "use_cases", "usecase", "usecases", "workflows", "workflow",
		"commands", "queries",
		// Domain layer
		"models", "model", "entities", "entity", "domain", "domains", "core", "business",
		"aggregates", "valueobjects", "schemas", "schema",
		// Infrastructure layer
		"db", "database", "repositories", "repository", "repo", "external", "adapters",
		"adapter", "persistence", "storage", "cache", "clients", "client",
		// Other common architectural components
		"utils", "util", "helpers", "helper", "common", "shared", "lib", "libs",
	}

	lowerPart := strings.ToLower(part)
	for _, keyword := range architecturalKeywords {
		if lowerPart == keyword || strings.HasPrefix(lowerPart, keyword) {
			return true
		}
	}
	return false
}

// extractPackagePrefixes extracts common package prefixes from module names
func (s *SystemAnalysisServiceImpl) extractPackagePrefixes(modules []string) []string {
	prefixMap := make(map[string]bool)

	for _, module := range modules {
		// For auto-detection, use more specific prefixes to avoid conflicts
		// e.g., "app.api.v1.admin" should produce "app.api" not just "app"
		parts := strings.Split(module, ".")

		if len(parts) >= 2 {
			// Check if the second part is a meaningful architectural component
			secondPart := strings.ToLower(parts[1])
			if s.isArchitecturalComponent(secondPart) {
				// Use first two parts as prefix (e.g., "app.api")
				prefixMap[parts[0]+"."+parts[1]] = true
			} else {
				// Use just the first part only if it's not too generic
				// Avoid using "app" alone if there are more specific prefixes
				prefixMap[parts[0]] = true
			}
		} else if len(parts) > 0 {
			prefixMap[parts[0]] = true
		}
	}

	// Convert map to slice and sort for deterministic results
	prefixes := make([]string, 0, len(prefixMap))
	for prefix := range prefixMap {
		prefixes = append(prefixes, prefix)
	}
	sort.Strings(prefixes)

	return prefixes
}

// evaluateLayerEdge evaluates a single dependency edge against rules and returns a violation if any.
func (s *SystemAnalysisServiceImpl) evaluateLayerEdge(rules *domain.ArchitectureRules, fromModule, toModule, fromLayer, toLayer string) *domain.ArchitectureViolation {
	if fromLayer == "unknown" || toLayer == "unknown" {
		if rules.StrictMode {
			return &domain.ArchitectureViolation{
				Type:        domain.ViolationTypeLayer,
				Severity:    domain.ViolationSeverityWarning,
				Module:      fromModule,
				Target:      toModule,
				Rule:        "strict_mode",
				Description: "Dependency involves unknown layer(s)",
				Suggestion:  "Assign modules to defined layers or relax strict_mode",
			}
		}
		return nil
	}
	// Find rule for fromLayer
	var layerRule *domain.LayerRule
	for i := range rules.Rules {
		if rules.Rules[i].From == fromLayer {
			layerRule = &rules.Rules[i]
			break
		}
	}
	if layerRule == nil {
		if rules.StrictMode {
			return &domain.ArchitectureViolation{
				Type:        domain.ViolationTypeLayer,
				Severity:    domain.ViolationSeverityWarning,
				Module:      fromModule,
				Target:      toModule,
				Rule:        "no_rule",
				Description: fmt.Sprintf("No rule defined for layer '%s'", fromLayer),
				Suggestion:  "Define allow/deny rules for this layer",
			}
		}
		return nil
	}
	// Deny takes precedence
	for _, d := range layerRule.Deny {
		if d == toLayer {
			return &domain.ArchitectureViolation{
				Type:        domain.ViolationTypeLayer,
				Severity:    domain.ViolationSeverityError,
				Module:      fromModule,
				Target:      toModule,
				Rule:        fmt.Sprintf("%s !> %s", fromLayer, toLayer),
				Description: fmt.Sprintf("Dependency from '%s' to '%s' is denied by rule", fromLayer, toLayer),
				Suggestion:  "Introduce an interface or move code to respect layer boundaries",
			}
		}
	}
	// Warn: the dependency is permitted but discouraged. Emit a warning and
	// stop before the allow-list check so it is not also reported as an error.
	for _, w := range layerRule.Warn {
		if w == toLayer {
			return &domain.ArchitectureViolation{
				Type:        domain.ViolationTypeLayer,
				Severity:    domain.ViolationSeverityWarning,
				Module:      fromModule,
				Target:      toModule,
				Rule:        fmt.Sprintf("%s ~> %s", fromLayer, toLayer),
				Description: fmt.Sprintf("Dependency from '%s' to '%s' is discouraged", fromLayer, toLayer),
				Suggestion:  "Route this dependency through an intermediary layer if possible",
			}
		}
	}
	if len(layerRule.Allow) > 0 {
		allowed := false
		for _, a := range layerRule.Allow {
			if a == toLayer {
				allowed = true
				break
			}
		}
		if !allowed {
			return &domain.ArchitectureViolation{
				Type:        domain.ViolationTypeLayer,
				Severity:    domain.ViolationSeverityError,
				Module:      fromModule,
				Target:      toModule,
				Rule:        fmt.Sprintf("%s -> {%s}", fromLayer, strings.Join(layerRule.Allow, ",")),
				Description: fmt.Sprintf("Dependency from '%s' to '%s' not allowed", fromLayer, toLayer),
				Suggestion:  "Refactor dependency or update architecture rules if intentional",
			}
		}
	}
	return nil
}

// toLayerViolations converts ArchitectureViolation list to LayerViolation list for summary.
func (s *SystemAnalysisServiceImpl) toLayerViolations(vs []domain.ArchitectureViolation, moduleToLayer map[string]string) []domain.LayerViolation {
	out := make([]domain.LayerViolation, 0, len(vs))
	for _, v := range vs {
		if v.Type != domain.ViolationTypeLayer {
			continue
		}
		out = append(out, domain.LayerViolation{
			FromModule:  v.Module,
			ToModule:    v.Target,
			FromLayer:   moduleToLayer[v.Module],
			ToLayer:     moduleToLayer[v.Target],
			Rule:        v.Rule,
			Severity:    v.Severity,
			Description: v.Description,
			Suggestion:  v.Suggestion,
		})
	}
	return out
}

// Helper methods

func (s *SystemAnalysisServiceImpl) buildDependencyMatrix(graph *analyzer.DependencyGraph) map[string]map[string]bool {
	matrix := make(map[string]map[string]bool)

	for moduleName := range graph.Nodes {
		matrix[moduleName] = make(map[string]bool)
		node := graph.Nodes[moduleName]

		for dep := range node.Dependencies {
			matrix[moduleName][dep] = true
		}
	}

	return matrix
}

func (s *SystemAnalysisServiceImpl) findLongestChains(graph *analyzer.DependencyGraph, limit int) []domain.DependencyPath {
	var chains []domain.DependencyPath

	// Find all paths using simple DFS
	for moduleName := range graph.Nodes {
		paths := s.findPathsFromModule(graph, moduleName, make(map[string]bool), []string{moduleName}, limit)
		chains = append(chains, paths...)
	}

	// Sort by length (descending), then by first module name for deterministic results
	sort.Slice(chains, func(i, j int) bool {
		if chains[i].Length != chains[j].Length {
			return chains[i].Length > chains[j].Length
		}
		// Tie-breaker: compare first module name for deterministic results
		return chains[i].Path[0] < chains[j].Path[0]
	})

	// Return top chains
	if len(chains) > limit {
		chains = chains[:limit]
	}

	return chains
}

func (s *SystemAnalysisServiceImpl) findPathsFromModule(graph *analyzer.DependencyGraph, current string, visited map[string]bool, path []string, maxPaths int) []domain.DependencyPath {
	var paths []domain.DependencyPath

	if len(paths) >= maxPaths {
		return paths
	}

	visited[current] = true
	defer delete(visited, current)

	node := graph.Nodes[current]
	if node == nil {
		return paths
	}

	for dep := range node.Dependencies {
		if !visited[dep] {
			newPath := append([]string{}, path...)
			newPath = append(newPath, dep)

			// Add this path
			if len(newPath) >= 2 {
				paths = append(paths, domain.DependencyPath{
					From:   newPath[0],
					To:     dep,
					Path:   newPath,
					Length: len(newPath),
				})
			}

			// Continue searching
			subPaths := s.findPathsFromModule(graph, dep, visited, newPath, maxPaths-len(paths))
			paths = append(paths, subPaths...)

			if len(paths) >= maxPaths {
				break
			}
		}
	}

	return paths
}

func (s *SystemAnalysisServiceImpl) calculateMaxDepth(graph *analyzer.DependencyGraph) int {
	maxDepth := 0

	for moduleName := range graph.Nodes {
		depth := s.calculateDepthFromModule(graph, moduleName, make(map[string]bool), 0)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

func (s *SystemAnalysisServiceImpl) calculateDepthFromModule(graph *analyzer.DependencyGraph, current string, visited map[string]bool, currentDepth int) int {
	if visited[current] {
		return currentDepth
	}

	visited[current] = true
	defer delete(visited, current)

	maxSubDepth := currentDepth
	node := graph.Nodes[current]

	if node != nil {
		for dep := range node.Dependencies {
			depth := s.calculateDepthFromModule(graph, dep, visited, currentDepth+1)
			if depth > maxSubDepth {
				maxSubDepth = depth
			}
		}
	}

	return maxSubDepth
}

func (s *SystemAnalysisServiceImpl) convertCouplingResults(results *analyzer.SystemMetrics) *domain.CouplingAnalysis {
	if results == nil {
		return nil
	}

	// Only consider modules as highly coupled if they actually have high coupling
	var highlyCoupled []string
	if len(results.RefactoringPriority) > 0 {
		// Only include in highly coupled if there's actual coupling
		if results.AverageFanIn+results.AverageFanOut > 0.5 {
			highlyCoupled = results.RefactoringPriority
		}
	}

	return &domain.CouplingAnalysis{
		AverageCoupling:       results.AverageFanIn + results.AverageFanOut,
		AverageInstability:    results.AverageInstability,
		MainSequenceDeviation: results.MainSequenceDeviation,
		HighlyCoupledModules:  highlyCoupled,
		StableModules:         results.StableModules,
		InstableModules:       results.InstableModules,
		ZoneOfPain:            results.ZoneOfPain,
		ZoneOfUselessness:     results.ZoneOfUselessness,
		MainSequence:          results.MainSequence,
	}
}

func (s *SystemAnalysisServiceImpl) convertCircularResults(result *analyzer.CircularDependencyResult) *domain.CircularDependencyAnalysis {
	if result == nil {
		return nil
	}

	var circularDeps []domain.CircularDependency
	coreModules := make(map[string]int) // Track modules appearing in multiple cycles

	for _, cycle := range result.CircularDependencies {
		circularDeps = append(circularDeps, domain.CircularDependency{
			Modules:      cycle.Modules,
			Description:  cycle.Description,
			Severity:     domain.CycleSeverity(cycle.Severity),
			Size:         cycle.Size,
			Dependencies: s.convertDependencyChains(cycle.Dependencies),
		})

		// Count occurrences for core infrastructure identification
		for _, module := range cycle.Modules {
			coreModules[module]++
		}
	}

	// Identify core infrastructure (modules in multiple cycles)
	var coreInfrastructure []string
	for module, count := range coreModules {
		if count > 1 {
			coreInfrastructure = append(coreInfrastructure, module)
		}
	}
	sort.Strings(coreInfrastructure)

	// Generate cycle breaking suggestions
	suggestions := s.generateCycleBreakingSuggestions(circularDeps, coreInfrastructure)

	return &domain.CircularDependencyAnalysis{
		HasCircularDependencies:  len(circularDeps) > 0,
		TotalCycles:              len(circularDeps),
		TotalModulesInCycles:     result.TotalModulesInCycles,
		CircularDependencies:     circularDeps,
		CycleBreakingSuggestions: suggestions,
		CoreInfrastructure:       coreInfrastructure,
	}
}

// Architecture analysis helper methods (simplified)

// Removed legacy helpers for ad-hoc layer counting.

// Removed helper methods that used undefined domain types

// convertDependencyChains converts analyzer.DependencyChain to domain.DependencyPath
func (s *SystemAnalysisServiceImpl) convertDependencyChains(chains []analyzer.DependencyChain) []domain.DependencyPath {
	var deps []domain.DependencyPath

	for _, chain := range chains {
		deps = append(deps, domain.DependencyPath{
			From:   chain.From,
			To:     chain.To,
			Path:   chain.Path,
			Length: chain.Length,
		})
	}

	return deps
}

// extractCouplingResult extracts coupling analysis from the dependency graph
func (s *SystemAnalysisServiceImpl) extractCouplingResult(graph *analyzer.DependencyGraph) *analyzer.SystemMetrics {
	// If SystemMetrics is already calculated in the graph, use it
	if graph.SystemMetrics != nil && graph.SystemMetrics.RefactoringPriority != nil {
		return graph.SystemMetrics
	}

	// Otherwise, calculate basic metrics
	metrics := &analyzer.SystemMetrics{
		TotalModules:      graph.TotalModules,
		TotalDependencies: graph.TotalEdges,
	}

	if graph.TotalModules > 0 {
		// Calculate averages from module metrics
		var totalFanIn, totalFanOut float64
		var totalInstability, totalAbstractness, totalDistance float64
		var refactoringCandidates []string

		if graph.ModuleMetrics != nil {
			for moduleName, moduleMetrics := range graph.ModuleMetrics {
				totalFanIn += float64(moduleMetrics.AfferentCoupling)
				totalFanOut += float64(moduleMetrics.EfferentCoupling)
				totalInstability += moduleMetrics.Instability
				totalAbstractness += moduleMetrics.Abstractness
				totalDistance += moduleMetrics.Distance

				// Identify refactoring priorities based on distance
				if moduleMetrics.Distance > 0.5 {
					refactoringCandidates = append(refactoringCandidates, moduleName)
				}
			}
		}

		moduleCount := float64(graph.TotalModules)
		metrics.AverageFanIn = totalFanIn / moduleCount
		metrics.AverageFanOut = totalFanOut / moduleCount
		metrics.AverageInstability = totalInstability / moduleCount
		metrics.AverageAbstractness = totalAbstractness / moduleCount
		metrics.MainSequenceDeviation = totalDistance / moduleCount
		metrics.SystemComplexity = float64(graph.TotalModules * 2)
		metrics.MaxDependencyDepth = s.calculateMaxDepth(graph)
		metrics.RefactoringPriority = refactoringCandidates

		// Modularity index approximation
		if graph.TotalEdges > 0 {
			metrics.ModularityIndex = 1.0 - (float64(graph.TotalEdges) / float64(graph.TotalModules*graph.TotalModules))
			if metrics.ModularityIndex < 0 {
				metrics.ModularityIndex = 0
			}
		}
	}

	return metrics
}

func minSystemAnalysis(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// extractModuleMetrics extracts module dependency metrics from the graph
func (s *SystemAnalysisServiceImpl) extractModuleMetrics(graph *analyzer.DependencyGraph) map[string]*domain.ModuleDependencyMetrics {
	result := make(map[string]*domain.ModuleDependencyMetrics)

	for moduleName, node := range graph.Nodes {
		// Get analyzer metrics if available
		var analyzerMetrics *analyzer.ModuleMetrics
		var hasMetrics bool
		if graph.ModuleMetrics != nil {
			analyzerMetrics, hasMetrics = graph.ModuleMetrics[moduleName]
		}

		metrics := &domain.ModuleDependencyMetrics{
			ModuleName: moduleName,
			FilePath:   node.FilePath,
			IsPackage:  node.IsPackage,
			Package:    node.Package,

			// Size metrics from node
			LinesOfCode:        node.LineCount,
			FunctionCount:      node.FunctionCount,
			ClassCount:         node.ClassCount,
			AbstractClassCount: node.AbstractClassCount,
			PublicInterface:    node.PublicNames,

			// Dependencies
			DirectDependencies:     s.getDirectDependencies(moduleName, node),
			TransitiveDependencies: s.getTransitiveDependencies(graph, moduleName),
			Dependents:             s.getDependents(graph, moduleName),
		}

		// If analyzer metrics are available, use them
		if hasMetrics && analyzerMetrics != nil {
			metrics.AfferentCoupling = analyzerMetrics.AfferentCoupling
			metrics.EfferentCoupling = analyzerMetrics.EfferentCoupling
			metrics.Instability = analyzerMetrics.Instability
			metrics.Abstractness = analyzerMetrics.Abstractness
			metrics.Distance = analyzerMetrics.Distance
			metrics.AbstractClassCount = analyzerMetrics.AbstractClassCount

			// Determine risk level based on distance
			if analyzerMetrics.Distance > 0.7 {
				metrics.RiskLevel = domain.RiskLevelHigh
			} else if analyzerMetrics.Distance > 0.4 {
				metrics.RiskLevel = domain.RiskLevelMedium
			} else {
				metrics.RiskLevel = domain.RiskLevelLow
			}
		} else {
			// Fallback to basic metrics
			metrics.AfferentCoupling = node.InDegree
			metrics.EfferentCoupling = node.OutDegree
			if (node.InDegree + node.OutDegree) > 0 {
				metrics.Instability = float64(node.OutDegree) / float64(node.InDegree+node.OutDegree)
			}
			metrics.RiskLevel = domain.RiskLevelLow
		}

		result[moduleName] = metrics
	}

	return result
}

// getDirectDependencies returns the direct dependencies of a module
func (s *SystemAnalysisServiceImpl) getDirectDependencies(moduleName string, node *analyzer.ModuleNode) []string {
	var deps []string
	for dep := range node.Dependencies {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	return deps
}

// getTransitiveDependencies returns all transitive dependencies of a module
func (s *SystemAnalysisServiceImpl) getTransitiveDependencies(graph *analyzer.DependencyGraph, moduleName string) []string {
	visited := make(map[string]bool)
	s.collectTransitiveDependencies(graph, moduleName, visited)

	// Remove the module itself from visited
	delete(visited, moduleName)

	var deps []string
	for dep := range visited {
		deps = append(deps, dep)
	}
	sort.Strings(deps)
	return deps
}

// collectTransitiveDependencies recursively collects all transitive dependencies
func (s *SystemAnalysisServiceImpl) collectTransitiveDependencies(graph *analyzer.DependencyGraph, moduleName string, visited map[string]bool) {
	if visited[moduleName] {
		return
	}
	visited[moduleName] = true

	node, exists := graph.Nodes[moduleName]
	if !exists {
		return
	}

	for dep := range node.Dependencies {
		s.collectTransitiveDependencies(graph, dep, visited)
	}
}

// getDependents returns the modules that depend on the given module
func (s *SystemAnalysisServiceImpl) getDependents(graph *analyzer.DependencyGraph, moduleName string) []string {
	var dependents []string
	for otherModule, otherNode := range graph.Nodes {
		if _, depends := otherNode.Dependencies[moduleName]; depends {
			dependents = append(dependents, otherModule)
		}
	}
	sort.Strings(dependents)
	return dependents
}

// generateCycleBreakingSuggestions generates suggestions for breaking circular dependencies
func (s *SystemAnalysisServiceImpl) generateCycleBreakingSuggestions(cycles []domain.CircularDependency, coreInfrastructure []string) []string {
	var suggestions []string

	if len(cycles) == 0 {
		return suggestions
	}

	// General suggestions
	suggestions = append(suggestions, "Consider introducing interfaces or abstract base classes to invert dependencies")

	// Suggest refactoring core infrastructure
	if len(coreInfrastructure) > 0 {
		suggestions = append(suggestions, fmt.Sprintf("Modules %v appear in multiple cycles - consider extracting shared functionality to a separate module", coreInfrastructure))
	}

	// Analyze cycle patterns
	for i, cycle := range cycles {
		if i >= 3 { // Limit detailed suggestions to first 3 cycles
			break
		}

		if cycle.Size == 2 {
			// For simple two-module cycles
			suggestions = append(suggestions, fmt.Sprintf("Break cycle between %s and %s by introducing a third module or using dependency injection", cycle.Modules[0], cycle.Modules[1]))
		} else if cycle.Size <= 4 {
			// For small cycles
			suggestions = append(suggestions, fmt.Sprintf("Cycle involving %v - identify the least coupled module and extract its dependencies", cycle.Modules))
		}
	}

	// Architecture-based suggestions
	suggestions = append(suggestions, "Review your architecture to ensure proper layer separation (e.g., presentation → application → domain → infrastructure)")
	suggestions = append(suggestions, "Consider using event-driven patterns to decouple tightly coupled modules")

	return suggestions
}

// generateArchitectureRecommendations generates architecture improvement recommendations
func (s *SystemAnalysisServiceImpl) generateArchitectureRecommendations(
	violations []domain.ArchitectureViolation,
	layerCohesion map[string]float64,
	problematicLayers []string,
	compliance float64) []domain.ArchitectureRecommendation {

	var recommendations []domain.ArchitectureRecommendation

	// Recommend based on compliance score
	if compliance < 0.6 {
		recommendations = append(recommendations, domain.ArchitectureRecommendation{
			Type:        domain.RecommendationTypeRestructure,
			Priority:    domain.RecommendationPriorityCritical,
			Title:       "Major Architecture Restructuring Required",
			Description: fmt.Sprintf("With %.1f%% compliance, your architecture has significant structural issues", compliance*100),
			Benefits: []string{
				"Improved maintainability and testability",
				"Clear separation of concerns",
				"Reduced coupling between layers",
			},
			Effort: domain.EstimatedEffortLarge,
			Steps: []string{
				"Identify and document current architecture patterns",
				"Define clear layer boundaries and responsibilities",
				"Create interfaces to decouple layers",
				"Gradually refactor violations starting with critical ones",
			},
		})
	}

	// Recommend fixing layer violations
	if len(violations) > 10 {
		violationModules := make(map[string]int)
		for _, v := range violations {
			violationModules[v.Module]++
		}

		var topViolators []string
		for module, count := range violationModules {
			if count > 2 {
				topViolators = append(topViolators, module)
			}
		}
		// Sort for deterministic results
		sort.Strings(topViolators)

		if len(topViolators) > 0 {
			recommendations = append(recommendations, domain.ArchitectureRecommendation{
				Type:        domain.RecommendationTypeRefactor,
				Priority:    domain.RecommendationPriorityHigh,
				Title:       "Address Frequent Architecture Violators",
				Description: fmt.Sprintf("Modules with multiple violations need refactoring: %v", topViolators[:minSystemAnalysis(3, len(topViolators))]),
				Benefits: []string{
					"Reduced architecture violations",
					"Better adherence to design principles",
					"Improved system structure",
				},
				Effort:  domain.EstimatedEffortMedium,
				Modules: topViolators,
				Steps: []string{
					"Review dependencies of violating modules",
					"Identify improper layer crossings",
					"Introduce abstractions or move code to appropriate layers",
				},
			})
		}
	}

	// Recommend improving layer cohesion
	for _, layer := range problematicLayers {
		if cohesion, exists := layerCohesion[layer]; exists && cohesion < 0.5 {
			recommendations = append(recommendations, domain.ArchitectureRecommendation{
				Type:        domain.RecommendationTypeRestructure,
				Priority:    domain.RecommendationPriorityMedium,
				Title:       fmt.Sprintf("Improve Cohesion in %s Layer", layer),
				Description: fmt.Sprintf("Layer '%s' has low cohesion (%.2f), indicating mixed responsibilities", layer, cohesion),
				Benefits: []string{
					"Better separation of concerns",
					"Increased code reusability",
					"Easier to understand and maintain",
				},
				Effort: domain.EstimatedEffortMedium,
				Steps: []string{
					"Review the responsibilities of modules in this layer",
					"Group related functionality together",
					"Consider splitting the layer if it has multiple distinct responsibilities",
				},
			})
		}
	}

	// General recommendations for good architecture
	if len(recommendations) < 3 {
		recommendations = append(recommendations, domain.ArchitectureRecommendation{
			Type:        domain.RecommendationTypeInterface,
			Priority:    domain.RecommendationPriorityLow,
			Title:       "Introduce Dependency Injection",
			Description: "Use dependency injection to reduce coupling between layers",
			Benefits: []string{
				"Better testability with mock dependencies",
				"Reduced coupling between components",
				"More flexible and maintainable code",
			},
			Effort: domain.EstimatedEffortMedium,
			Steps: []string{
				"Identify tightly coupled components",
				"Create interfaces for dependencies",
				"Inject dependencies rather than creating them directly",
			},
		})
	}

	return recommendations
}

// identifyArchitectureRefactoringTargets identifies modules that need refactoring based on violations
func (s *SystemAnalysisServiceImpl) identifyArchitectureRefactoringTargets(
	violations []domain.ArchitectureViolation,
	moduleToLayer map[string]string) []string {

	// Count violations per module
	violationCount := make(map[string]int)
	for _, v := range violations {
		violationCount[v.Module]++
	}

	// Sort modules by violation count
	type moduleViolation struct {
		module string
		count  int
	}

	var modules []moduleViolation
	for module, count := range violationCount {
		modules = append(modules, moduleViolation{module: module, count: count})
	}

	sort.Slice(modules, func(i, j int) bool {
		if modules[i].count != modules[j].count {
			return modules[i].count > modules[j].count
		}
		// Tie-breaker: alphabetical order by module name for deterministic results
		return modules[i].module < modules[j].module
	})

	// Return top refactoring targets
	var targets []string
	for i, m := range modules {
		if i >= 10 { // Limit to top 10
			break
		}
		targets = append(targets, m.module)
	}

	return targets
}
