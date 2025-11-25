package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	var allResults []interface{}
	var warnings []string
	var errors []string
	startTime := time.Now()

	// Analyze dependencies if requested
	var dependencyResult *domain.DependencyAnalysisResult
	if domain.BoolValue(req.AnalyzeDependencies, true) {
		result, err := s.AnalyzeDependencies(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Dependency analysis failed: %v", err))
		} else {
			dependencyResult = result
			allResults = append(allResults, result)
		}
	}

	// Analyze architecture if requested
	var architectureResult *domain.ArchitectureAnalysisResult
	if domain.BoolValue(req.AnalyzeArchitecture, true) {
		result, err := s.AnalyzeArchitecture(ctx, req)
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
		GeneratedAt:          time.Now(),
		Duration:             time.Since(startTime).Milliseconds(),
		Version:              version.Version,
		Warnings:             warnings,
		Errors:               errors,
	}

	return response, nil
}

// AnalyzeDependencies performs dependency analysis only
func (s *SystemAnalysisServiceImpl) AnalyzeDependencies(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.DependencyAnalysisResult, error) {
	// Determine project root from common parent of paths
	projectRoot := s.findProjectRoot(req.Paths)

	// Create module analyzer with options
	options := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		IncludeStdLib:     domain.BoolValue(req.IncludeStdLib, false),
		IncludeThirdParty: domain.BoolValue(req.IncludeThirdParty, true),
		FollowRelative:    domain.BoolValue(req.FollowRelative, true),
		IncludePatterns:   req.IncludePatterns,
		ExcludePatterns:   req.ExcludePatterns,
	}

	moduleAnalyzer, err := analyzer.NewModuleAnalyzer(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create module analyzer: %w", err)
	}

	// Analyze files using ModuleAnalyzer
	graph, err := moduleAnalyzer.AnalyzeFiles(req.Paths)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
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

	// Calculate coupling metrics
	metricsCalculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
	if err = metricsCalculator.CalculateMetrics(); err != nil {
		return nil, err
	}
	couplingResults := s.extractCouplingResult(graph)

	// Build dependency matrix
	matrix := s.buildDependencyMatrix(graph)

	// Find longest dependency chains
	longestChains := s.findLongestChains(graph, 10) // Top 10 chains

	// Extract module metrics
	moduleMetrics := s.extractModuleMetrics(graph)

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
	// Build dependency graph
	graph, err := s.buildDependencyGraph(req)
	if err != nil {
		return nil, err
	}

	// Auto-detect architecture if no rules are defined
	if req.ArchitectureRules == nil || (len(req.ArchitectureRules.Layers) == 0 && len(req.ArchitectureRules.Rules) == 0) {
		req.ArchitectureRules = s.autoDetectArchitecture(graph)
		// If auto-detection found no recognizable patterns, return empty result
		if req.ArchitectureRules == nil || len(req.ArchitectureRules.Layers) == 0 {
			return s.emptyArchitectureResult(), nil
		}
	}

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
	compliance := s.calculateCompliance(len(violations), checked)

	// Generate architecture recommendations
	recommendations := s.generateArchitectureRecommendations(violations, layerCohesion, problematic, compliance)

	// Identify refactoring targets based on violations
	refactoringTargets := s.identifyArchitectureRefactoringTargets(violations, moduleToLayer)

	// Build result
	return s.buildArchitectureResultWithRecommendations(violations, severityCounts, layerCoupling, layerCohesion,
		problematic, layersAnalyzed, compliance, checked, moduleToLayer, recommendations, refactoringTargets), nil
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
func (s *SystemAnalysisServiceImpl) buildDependencyGraph(req domain.SystemAnalysisRequest) (*analyzer.DependencyGraph, error) {
	projectRoot := s.findProjectRoot(req.Paths)
	options := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		IncludeStdLib:     domain.BoolValue(req.IncludeStdLib, false),
		IncludeThirdParty: domain.BoolValue(req.IncludeThirdParty, true),
		FollowRelative:    domain.BoolValue(req.FollowRelative, true),
		IncludePatterns:   req.IncludePatterns,
		ExcludePatterns:   req.ExcludePatterns,
	}
	ma, err := analyzer.NewModuleAnalyzer(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create module analyzer: %w", err)
	}
	graph, err := ma.AnalyzeFiles(req.Paths)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze architecture dependencies: %w", err)
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

	return layerCohesion, problematic, layersAnalyzed
}

// calculateCompliance calculates the compliance score
func (s *SystemAnalysisServiceImpl) calculateCompliance(violations, checked int) float64 {
	compliance := 1.0
	if checked > 0 {
		compliance = 1.0 - (float64(violations) / float64(checked))
		if compliance < 0 {
			compliance = 0
		}
	}
	return compliance
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
	checked int,
	moduleToLayer map[string]string,
	recommendations []domain.ArchitectureRecommendation,
	refactoringTargets []string) *domain.ArchitectureAnalysisResult {

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
		TotalRules:             checked,
		LayerAnalysis:          layerAnalysis,
		CohesionAnalysis:       nil,
		ResponsibilityAnalysis: nil,
		Violations:             violations,
		SeverityBreakdown:      severityCounts,
		Recommendations:        recommendations,
		RefactoringTargets:     refactoringTargets,
	}
}

// buildModuleLayerMap maps each module to a layer based on ArchitectureRules.
// compiledPattern keeps the compiled regex and its original pattern with simple specificity info.
type compiledPattern struct {
	re          *regexp.Regexp
	original    string
	specificity int // number of dots in original pattern; higher = more specific
}

func (s *SystemAnalysisServiceImpl) buildModuleLayerMap(graph *analyzer.DependencyGraph, rules *domain.ArchitectureRules) map[string]string {
	out := make(map[string]string)
	compiled := make(map[string][]compiledPattern)
	for _, layer := range rules.Layers {
		for _, pat := range layer.Packages {
			if re := s.compileModulePattern(pat); re != nil {
				cp := compiledPattern{re: re, original: pat, specificity: strings.Count(pat, ".")}
				compiled[layer.Name] = append(compiled[layer.Name], cp)
			}
		}
	}
	for module := range graph.Nodes {
		out[module] = s.findLayerForModule(module, compiled)
		if out[module] == "" {
			out[module] = "unknown"
		}
	}
	return out
}

// findLayerForModule returns the most specific matching layer for a module.
func (s *SystemAnalysisServiceImpl) findLayerForModule(module string, compiled map[string][]compiledPattern) string {
	// Find all matching patterns with their specificity
	type match struct {
		layer       string
		pattern     string
		specificity int
	}

	var matches []match
	for layer, patterns := range compiled {
		for _, cp := range patterns {
			if cp.re.MatchString(module) {
				matches = append(matches, match{layer: layer, pattern: cp.original, specificity: cp.specificity})
			}
		}
	}

	// Return the most specific match
	if len(matches) > 0 {
		best := matches[0]
		for _, m := range matches[1:] {
			if m.specificity > best.specificity {
				best = m
			} else if m.specificity == best.specificity {
				// tie-breaker: prefer longer original pattern
				if len(m.pattern) > len(best.pattern) {
					best = m
				}
			}
		}
		return best.layer
	}

	return ""
}

// compileModulePattern converts simple glob-like patterns to regex for module names.
// For Python modules, pattern "views" should match "views", "views.foo", "views.foo.bar", etc.
func (s *SystemAnalysisServiceImpl) compileModulePattern(glob string) *regexp.Regexp {
	if glob == "" {
		return nil
	}

	escaped := regexp.QuoteMeta(glob)
	hasWildcard := strings.Contains(glob, "*")
	if hasWildcard {
		escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	} else {
		// Allow matching the module segment plus any nested submodules
		escaped = fmt.Sprintf("%s(?:\\..+)?", escaped)
	}

	// Match either at the beginning of the module path or as a later segment.
	pattern := fmt.Sprintf("^(?:%s|.+\\.%s)$", escaped, escaped)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil
	}
	return re
}

// autoDetectArchitecture automatically detects architecture patterns from the dependency graph
func (s *SystemAnalysisServiceImpl) autoDetectArchitecture(graph *analyzer.DependencyGraph) *domain.ArchitectureRules {
	// Load layer patterns and rules from the embedded default config
	defaultConfig, err := config.LoadDefaultConfig()
	if err != nil {
		// This should never happen since the config is embedded at compile time
		// But if it does, return nil to indicate no auto-detection is possible
		return nil
	}

	// Build layer patterns map from default config
	layerPatterns := make(map[string][]string)
	for _, layer := range defaultConfig.Architecture.Layers {
		layerPatterns[layer.Name] = layer.Packages
	}

	// Detect which modules belong to which layer
	moduleToLayer := make(map[string]string)
	layerModules := make(map[string][]string)

	for module := range graph.Nodes {
		layer := s.detectLayerFromModule(module, layerPatterns)
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
		StrictMode: false, // Auto-detected rules are not strict by default
	}
}

// detectLayerFromModule determines which layer a module belongs to based on its name
func (s *SystemAnalysisServiceImpl) detectLayerFromModule(module string, patterns map[string][]string) string {
	// Split module path into parts
	parts := strings.Split(module, ".")

	if len(parts) == 0 {
		return ""
	}

	// Priority 1: Check the LAST part of the module path (most specific)
	// This is usually the most accurate indicator of the layer
	// e.g., "app.api.v1.admin.companies.router" -> "router" indicates presentation layer
	lastPart := strings.ToLower(parts[len(parts)-1])

	// Define priority order for last part matching
	// Presentation and Infrastructure are most specific when they're the last part
	lastPartPriority := []string{"presentation", "infrastructure", "application", "domain"}
	for _, layer := range lastPartPriority {
		for _, pattern := range patterns[layer] {
			// Exact match or variations with underscores/prefixes/suffixes
			if lastPart == pattern ||
				strings.HasPrefix(lastPart, pattern+"_") ||
				strings.HasSuffix(lastPart, "_"+pattern) ||
				(len(lastPart) > len(pattern) && strings.Contains(lastPart, pattern)) {
				return layer
			}
		}
	}

	// Priority 2: Check the second-to-last part if it exists
	// This handles cases like "services.py" where the last part is just "py"
	if len(parts) > 1 {
		secondToLast := strings.ToLower(parts[len(parts)-2])
		for _, layer := range lastPartPriority {
			for _, pattern := range patterns[layer] {
				if secondToLast == pattern || strings.HasPrefix(secondToLast, pattern+"_") || strings.HasSuffix(secondToLast, "_"+pattern) {
					return layer
				}
			}
		}
	}

	// Priority 3: Check all parts from beginning to end
	// This catches cases where the layer indicator is in the middle of the path
	for _, part := range parts {
		lowerPart := strings.ToLower(part)
		for _, layer := range lastPartPriority {
			for _, pattern := range patterns[layer] {
				if lowerPart == pattern {
					return layer
				}
			}
		}
	}

	return ""
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

	// Convert map to slice
	prefixes := make([]string, 0, len(prefixMap))
	for prefix := range prefixMap {
		prefixes = append(prefixes, prefix)
	}

	return prefixes
}

// evaluateLayerEdge evaluates a single dependency edge against rules and returns a violation if any.
func (s *SystemAnalysisServiceImpl) evaluateLayerEdge(rules *domain.ArchitectureRules, fromModule, toModule, fromLayer, toLayer string) *domain.ArchitectureViolation {
	if (fromLayer == "unknown" || toLayer == "unknown") && rules.StrictMode {
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

// findProjectRoot finds the common parent directory of all given paths
func (s *SystemAnalysisServiceImpl) findProjectRoot(paths []string) string {
	if len(paths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	// Get absolute paths
	absPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		// If it's a file, get its directory
		info, err := os.Stat(absPath)
		if err == nil && !info.IsDir() {
			absPath = filepath.Dir(absPath)
		}

		absPaths = append(absPaths, absPath)
	}

	if len(absPaths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	// Find common parent
	commonParent := absPaths[0]
	for _, path := range absPaths[1:] {
		for !strings.HasPrefix(path, commonParent) {
			commonParent = filepath.Dir(commonParent)
			if commonParent == "/" || commonParent == "." {
				break
			}
		}
	}

	// If common parent has __init__.py, it's a Python package root
	// Otherwise, look for common markers like setup.py, pyproject.toml, etc.
	for {
		// Check for Python project markers
		markers := []string{"setup.py", "pyproject.toml", "setup.cfg", ".git", "requirements.txt"}
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(commonParent, marker)); err == nil {
				return commonParent
			}
		}

		// Check if we've reached the root
		parent := filepath.Dir(commonParent)
		if parent == commonParent || parent == "/" || parent == "." {
			break
		}

		// Don't go above the original common parent too much
		if !strings.HasPrefix(absPaths[0], parent) {
			break
		}

		commonParent = parent
	}

	return commonParent
}

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

	// Sort by length (descending)
	sort.Slice(chains, func(i, j int) bool {
		return chains[i].Length > chains[j].Length
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
		ZoneOfPain:            s.extractZoneOfPain(results),
		ZoneOfUselessness:     s.extractZoneOfUselessness(results),
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

func (s *SystemAnalysisServiceImpl) extractZoneOfPain(metrics *analyzer.SystemMetrics) []string {
	// Zone of pain: high coupling, low abstractness
	// For now, return refactoring priorities as a proxy
	if len(metrics.RefactoringPriority) > 3 {
		return metrics.RefactoringPriority[:3]
	}
	return metrics.RefactoringPriority
}

func (s *SystemAnalysisServiceImpl) extractZoneOfUselessness(metrics *analyzer.SystemMetrics) []string {
	// Zone of uselessness: low coupling, high abstractness
	// This would require more detailed analysis of individual modules
	return []string{}
}

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
			LinesOfCode:     node.LineCount,
			FunctionCount:   node.FunctionCount,
			ClassCount:      node.ClassCount,
			PublicInterface: node.PublicNames,

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
		return modules[i].count > modules[j].count
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
