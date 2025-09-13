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
	if req.AnalyzeDependencies {
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
	if req.AnalyzeArchitecture {
		result, err := s.AnalyzeArchitecture(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Architecture analysis failed: %v", err))
		} else {
			architectureResult = result
			allResults = append(allResults, result)
		}
	}

	// Analyze quality metrics if requested
	var qualityResult *domain.QualityMetricsResult
	if req.AnalyzeQuality {
		result, err := s.AnalyzeQuality(ctx, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Quality analysis failed: %v", err))
		} else {
			qualityResult = result
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
		QualityMetrics:       qualityResult,
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
		IncludeStdLib:     req.IncludeStdLib,
		IncludeThirdParty: req.IncludeThirdParty,
		FollowRelative:    req.FollowRelative,
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

	// Create dependency analysis result
	result := &domain.DependencyAnalysisResult{
		TotalModules:         graph.TotalModules,
		TotalDependencies:    graph.TotalEdges,
		RootModules:          graph.GetRootModules(),
		LeafModules:          graph.GetLeafModules(),
		ModuleMetrics:        make(map[string]*domain.ModuleDependencyMetrics), // Mock for now
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

    // Build result
    return s.buildArchitectureResult(violations, severityCounts, layerCoupling, layerCohesion,
        problematic, layersAnalyzed, compliance, checked), nil
}

// emptyArchitectureResult returns an empty result when no rules are defined
func (s *SystemAnalysisServiceImpl) emptyArchitectureResult() *domain.ArchitectureAnalysisResult {
    return &domain.ArchitectureAnalysisResult{
        ComplianceScore:        1.0,
        TotalViolations:        0,
        TotalRules:             0,
        LayerAnalysis:          &domain.LayerAnalysis{
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

// buildArchitectureResult constructs the final result
func (s *SystemAnalysisServiceImpl) buildArchitectureResult(
    violations []domain.ArchitectureViolation,
    severityCounts map[domain.ViolationSeverity]int,
    layerCoupling map[string]map[string]int,
    layerCohesion map[string]float64,
    problematic []string,
    layersAnalyzed int,
    compliance float64,
    checked int) *domain.ArchitectureAnalysisResult {

    layerAnalysis := &domain.LayerAnalysis{
        LayersAnalyzed:    layersAnalyzed,
        LayerViolations:   s.toLayerViolations(violations),
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
        Recommendations:        []domain.ArchitectureRecommendation{},
        RefactoringTargets:     []string{},
    }
}

// buildModuleLayerMap maps each module to a layer based on ArchitectureRules.
func (s *SystemAnalysisServiceImpl) buildModuleLayerMap(graph *analyzer.DependencyGraph, rules *domain.ArchitectureRules) map[string]string {
    out := make(map[string]string)
    compiled := make(map[string][]*regexp.Regexp)
    for _, layer := range rules.Layers {
        for _, pat := range layer.Packages {
            if re := s.compileModulePattern(pat); re != nil {
                compiled[layer.Name] = append(compiled[layer.Name], re)
            }
        }
    }
    for module := range graph.Nodes {
        out[module] = s.findLayerForModule(module, compiled)
        if out[module] == "" { out[module] = "unknown" }
    }
    return out
}

// findLayerForModule returns the first matching layer for a module.
func (s *SystemAnalysisServiceImpl) findLayerForModule(module string, compiled map[string][]*regexp.Regexp) string {
    for layer, patterns := range compiled {
        for _, re := range patterns {
            if re.MatchString(module) { return layer }
        }
    }
    return ""
}

// compileModulePattern converts simple glob-like patterns to regex for module names.
// For Python modules, pattern "views" should match "views", "views.foo", "views.foo.bar", etc.
func (s *SystemAnalysisServiceImpl) compileModulePattern(glob string) *regexp.Regexp {
    // Handle wildcards
    if strings.Contains(glob, "*") {
        esc := regexp.QuoteMeta(glob)
        esc = strings.ReplaceAll(esc, "\\*", ".*")
        re, err := regexp.Compile("^" + esc + "$")
        if err != nil { return nil }
        return re
    }

    // For non-wildcard patterns, match the module and any submodules
    // Pattern "views" matches "views", "views.foo", "views.foo.bar", etc.
    esc := regexp.QuoteMeta(glob)
    pattern := "^" + esc + "(\\..+)?$"
    re, err := regexp.Compile(pattern)
    if err != nil { return nil }
    return re
}

// autoDetectArchitecture automatically detects architecture patterns from the dependency graph
func (s *SystemAnalysisServiceImpl) autoDetectArchitecture(graph *analyzer.DependencyGraph) *domain.ArchitectureRules {
    // Standard layer patterns commonly used in Python projects
    layerPatterns := map[string][]string{
        "presentation": {"api", "apis", "views", "view", "controllers", "controller", "routes", "route", "handlers", "handler", "ui", "web", "rest", "graphql", "endpoints", "endpoint"},
        "application":  {"services", "service", "use_cases", "usecase", "usecases", "workflows", "workflow", "commands", "queries", "app"},
        "domain":       {"models", "model", "entities", "entity", "domain", "domains", "core", "business", "aggregates", "valueobjects", "schemas", "schema"},
        "infrastructure": {"db", "database", "repositories", "repository", "repo", "external", "adapters", "adapter", "persistence", "storage", "cache", "clients", "client"},
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

    // Define standard layered architecture rules
    rules := []domain.LayerRule{
        // Presentation layer can access application and domain
        {From: "presentation", Allow: []string{"application", "domain"}},
        // Application layer can only access domain
        {From: "application", Allow: []string{"domain"}},
        // Domain layer should not access other layers
        {From: "domain", Deny: []string{"presentation", "application", "infrastructure"}},
        // Infrastructure can access domain
        {From: "infrastructure", Allow: []string{"domain"}},
    }

    return &domain.ArchitectureRules{
        Layers: layers,
        Rules:  rules,
        StrictMode: false, // Auto-detected rules are not strict by default
    }
}

// detectLayerFromModule determines which layer a module belongs to based on its name
func (s *SystemAnalysisServiceImpl) detectLayerFromModule(module string, patterns map[string][]string) string {
    // Split module path into parts
    parts := strings.Split(module, ".")

    // Check each part against patterns
    for _, part := range parts {
        lowerPart := strings.ToLower(part)
        for layer, layerPatterns := range patterns {
            for _, pattern := range layerPatterns {
                if lowerPart == pattern || strings.HasPrefix(lowerPart, pattern) {
                    return layer
                }
            }
        }
    }

    return ""
}

// extractPackagePrefixes extracts common package prefixes from module names
func (s *SystemAnalysisServiceImpl) extractPackagePrefixes(modules []string) []string {
    prefixMap := make(map[string]bool)

    for _, module := range modules {
        // Get the first part of the module path as the package prefix
        parts := strings.Split(module, ".")
        if len(parts) > 0 {
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
        for _, a := range layerRule.Allow { if a == toLayer { allowed = true; break } }
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
func (s *SystemAnalysisServiceImpl) toLayerViolations(vs []domain.ArchitectureViolation) []domain.LayerViolation {
    out := make([]domain.LayerViolation, 0, len(vs))
    for _, v := range vs {
        out = append(out, domain.LayerViolation{
            FromModule:  v.Module,
            ToModule:    v.Target,
            FromLayer:   "",
            ToLayer:     "",
            Rule:        v.Rule,
            Severity:    v.Severity,
            Description: v.Description,
            Suggestion:  v.Suggestion,
        })
    }
    return out
}

// AnalyzeQuality performs quality metrics analysis only
func (s *SystemAnalysisServiceImpl) AnalyzeQuality(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.QualityMetricsResult, error) {
	// Build dependency graph for quality analysis
	graph := analyzer.NewDependencyGraph("")

	// Basic file processing for quality metrics
	for _, filePath := range req.Paths {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("quality analysis cancelled: %w", ctx.Err())
		default:
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		parseResult, err := s.parser.Parse(ctx, content)
		if err != nil {
			continue
		}

		moduleName := s.getModuleNameFromPath(filePath)
		graph.AddModule(moduleName, filePath)

		// Extract basic metrics (mock implementation)
		dependencies := s.extractImportsFromAST(parseResult.AST, filePath)
		for _, dep := range dependencies {
			targetModule := s.extractModuleNameFromImport(dep)
			if targetModule != "" && targetModule != moduleName {
				graph.AddDependency(moduleName, targetModule, analyzer.DependencyEdgeImport, dep)
			}
		}
	}

	// Calculate quality metrics
	metricsCalculator := analyzer.NewCouplingMetricsCalculator(graph, analyzer.DefaultCouplingMetricsOptions())
	err := metricsCalculator.CalculateMetrics()
	if err != nil {
		return nil, err
	}
	systemMetrics := s.extractSystemMetrics(graph)

	result := &domain.QualityMetricsResult{
		OverallQuality:         s.calculateOverallQuality(systemMetrics),
		MaintainabilityIndex:   systemMetrics.MaintainabilityIndex,
		TechnicalDebtTotal:     systemMetrics.TechnicalDebtTotal,
		ModularityIndex:        systemMetrics.ModularityIndex,
		SystemInstability:      systemMetrics.AverageInstability,
		SystemAbstractness:     systemMetrics.AverageAbstractness,
		MainSequenceDistance:   systemMetrics.MainSequenceDeviation,
		SystemComplexity:       systemMetrics.SystemComplexity,
		MaxDependencyDepth:     systemMetrics.MaxDependencyDepth,
		AverageFanIn:           systemMetrics.AverageFanIn,
		AverageFanOut:          systemMetrics.AverageFanOut,
		HighQualityModules:     []string{}, // Mock for now
		ModerateQualityModules: []string{}, // Mock for now
		LowQualityModules:      systemMetrics.RefactoringPriority[:minSystemAnalysis(3, len(systemMetrics.RefactoringPriority))],
		CriticalModules:        []string{}, // Mock for now
		QualityTrends:          make(map[string]float64),
		HotSpots:               []string{},                   // Mock for now
		RefactoringTargets:     []domain.RefactoringTarget{}, // Mock for now
	}

	return result, nil
}

// Helper methods

// getModuleNameFromPath is deprecated - use ModuleAnalyzer.filePathToModuleName instead
// Keeping for backward compatibility during transition
func (s *SystemAnalysisServiceImpl) getModuleNameFromPath(filePath string) string {
	// For now, just use the simple approach without "app" hardcoding
	// Get relative path from current directory
	relPath := filePath
	if absPath, err := filepath.Abs(filePath); err == nil {
		if cwd, err := os.Getwd(); err == nil {
			if rel, err := filepath.Rel(cwd, absPath); err == nil {
				relPath = rel
			}
		}
	}

	// Remove .py extension
	relPath = strings.TrimSuffix(relPath, ".py")

	// Handle __init__.py files
	if strings.HasSuffix(relPath, "__init__") {
		relPath = filepath.Dir(relPath)
	}

	// Convert path separators to dots
	moduleName := strings.ReplaceAll(relPath, string(filepath.Separator), ".")

	// Clean up leading/trailing dots
	moduleName = strings.Trim(moduleName, ".")

	return moduleName
}

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

	return &domain.CouplingAnalysis{
		AverageCoupling:       results.AverageFanIn + results.AverageFanOut,
		AverageInstability:    results.AverageInstability,
		MainSequenceDeviation: results.MainSequenceDeviation,
		HighlyCoupledModules:  results.RefactoringPriority,
		ZoneOfPain:            s.extractZoneOfPain(results),
		ZoneOfUselessness:     s.extractZoneOfUselessness(results),
	}
}

func (s *SystemAnalysisServiceImpl) convertCircularResults(result *analyzer.CircularDependencyResult) *domain.CircularDependencyAnalysis {
	if result == nil {
		return nil
	}

	var circularDeps []domain.CircularDependency
	for _, cycle := range result.CircularDependencies {
		circularDeps = append(circularDeps, domain.CircularDependency{
			Modules:      cycle.Modules,
			Description:  cycle.Description,
			Severity:     domain.CycleSeverity(cycle.Severity),
			Size:         cycle.Size,
			Dependencies: s.convertCycleDependencies(cycle.Modules),
		})
	}

	return &domain.CircularDependencyAnalysis{
		HasCircularDependencies:  len(circularDeps) > 0,
		TotalCycles:              len(circularDeps),
		TotalModulesInCycles:     result.TotalModulesInCycles,
		CircularDependencies:     circularDeps,
		CycleBreakingSuggestions: []string{}, // Mock for now
	}
}

// Architecture analysis helper methods (simplified)

// Quality analysis helper methods

func (s *SystemAnalysisServiceImpl) calculateOverallQuality(metrics *analyzer.SystemMetrics) float64 {
	// Simple quality calculation (0-100)
	quality := 70.0 // Base quality

	// Adjust based on various factors
	if metrics.AverageInstability > 0.7 {
		quality -= 15
	}
	if metrics.CyclicDependencies > 0 {
		quality -= 20
	}
	if metrics.MainSequenceDeviation > 0.4 {
		quality -= 10
	}

	if quality < 0 {
		quality = 0
	}

	return quality
}

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

func (s *SystemAnalysisServiceImpl) convertCycleDependencies(cycle []string) []domain.DependencyPath {
	var deps []domain.DependencyPath

	for i := 0; i < len(cycle); i++ {
		next := (i + 1) % len(cycle)
		deps = append(deps, domain.DependencyPath{
			From:   cycle[i],
			To:     cycle[next],
			Length: 2,
		})
	}

	return deps
}

// extractCouplingResult extracts coupling analysis from the dependency graph
func (s *SystemAnalysisServiceImpl) extractCouplingResult(graph *analyzer.DependencyGraph) *analyzer.SystemMetrics {
	// Mock system metrics extraction from graph
	// In a real implementation, this would aggregate metrics from graph.ModuleMetrics
	return &analyzer.SystemMetrics{
		AverageFanIn:          float64(graph.TotalEdges) / float64(graph.TotalModules),
		AverageFanOut:         float64(graph.TotalEdges) / float64(graph.TotalModules),
		AverageInstability:    0.5,  // Mock value
		AverageAbstractness:   0.3,  // Mock value
		MainSequenceDeviation: 0.2,  // Mock value
		MaintainabilityIndex:  75.0, // Mock value
		TechnicalDebtTotal:    10.0, // Mock value
		ModularityIndex:       0.8,  // Mock value
		SystemComplexity:      float64(graph.TotalModules * 2),
		MaxDependencyDepth:    s.calculateMaxDepth(graph),
		CyclicDependencies:    0,          // Will be updated by circular analysis
		RefactoringPriority:   []string{}, // Mock empty
	}
}

// extractSystemMetrics extracts system-wide metrics from the dependency graph
func (s *SystemAnalysisServiceImpl) extractSystemMetrics(graph *analyzer.DependencyGraph) *analyzer.SystemMetrics {
	return s.extractCouplingResult(graph) // Same implementation for now
}

// extractImportsFromAST extracts import information from AST
// TEMPORARY: Using regex-based extraction due to AST builder issues
// extractImportsFromAST is deprecated - ModuleAnalyzer handles this internally
// Keeping for backward compatibility during transition
func (s *SystemAnalysisServiceImpl) extractImportsFromAST(ast *parser.Node, filePath string) []*analyzer.ImportInfo {
	// Read the source file directly for regex-based parsing
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Fall back to empty imports if file can't be read
		return []*analyzer.ImportInfo{}
	}

	return s.extractImportsWithRegex(string(content))
}

// NOTE: extractImportsWithRegex is deprecated - ModuleAnalyzer handles this internally
// Keeping for backward compatibility during transition
func (s *SystemAnalysisServiceImpl) extractImportsWithRegex(source string) []*analyzer.ImportInfo {
	var imports []*analyzer.ImportInfo

	// Regular expressions for different import patterns
	// Match: import module [as alias]
	importRe := regexp.MustCompile(`^\s*import\s+([\w\.]+)(?:\s+as\s+(\w+))?`)
	// Match: from module import name [as alias], name2, ...
	fromImportRe := regexp.MustCompile(`^\s*from\s+([\w\.]*?)\s+import\s+(.+)`)
	// Match relative imports: from . import ..., from .. import ...
	relativeImportRe := regexp.MustCompile(`^\s*from\s+(\.+)([\w\.]*)\s+import\s+(.+)`)

	lines := strings.Split(source, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		// Handle continuation lines (backslash at end)
		if strings.HasSuffix(line, "\\") {
			// Combine with next line(s)
			for i := lineNum + 1; i < len(lines) && strings.HasSuffix(line, "\\"); i++ {
				line = strings.TrimSuffix(line, "\\") + " " + strings.TrimSpace(lines[i])
			}
		}

		// Handle parentheses for multi-line imports
		if strings.Contains(line, "(") && !strings.Contains(line, ")") {
			// Find closing parenthesis
			for i := lineNum + 1; i < len(lines); i++ {
				line += " " + strings.TrimSpace(lines[i])
				if strings.Contains(lines[i], ")") {
					break
				}
			}
		}

		// Check for relative imports first
		if matches := relativeImportRe.FindStringSubmatch(line); matches != nil {
			dots := matches[1]
			moduleName := matches[2]
			namesStr := matches[3]

			// Parse imported names
			names := s.parseImportNames(namesStr)

			statement := fmt.Sprintf("from %s%s import %s", dots, moduleName, namesStr)

			impInfo := &analyzer.ImportInfo{
				Statement:     statement,
				ImportedNames: names,
				IsRelative:    true,
				Level:         len(dots),
				Line:          lineNum + 1,
			}
			imports = append(imports, impInfo)

		} else if matches := fromImportRe.FindStringSubmatch(line); matches != nil {
			// from module import names
			moduleName := matches[1]
			namesStr := matches[2]

			// Parse imported names
			names := s.parseImportNames(namesStr)

			statement := fmt.Sprintf("from %s import %s", moduleName, namesStr)

			impInfo := &analyzer.ImportInfo{
				Statement:     statement,
				ImportedNames: names,
				IsRelative:    false,
				Line:          lineNum + 1,
			}
			imports = append(imports, impInfo)

		} else if matches := importRe.FindStringSubmatch(line); matches != nil {
			// import module [as alias]
			moduleName := matches[1]
			alias := ""
			if len(matches) > 2 {
				alias = matches[2]
			}

			statement := fmt.Sprintf("import %s", moduleName)
			if alias != "" {
				statement = fmt.Sprintf("import %s as %s", moduleName, alias)
			}

			impInfo := &analyzer.ImportInfo{
				Statement:     statement,
				ImportedNames: []string{moduleName},
				Alias:         alias,
				IsRelative:    false,
				Line:          lineNum + 1,
			}
			imports = append(imports, impInfo)
		}
	}

	return imports
}

// parseImportNames parses comma-separated import names, handling aliases
func (s *SystemAnalysisServiceImpl) parseImportNames(namesStr string) []string {
	var names []string

	// Remove parentheses if present
	namesStr = strings.Trim(namesStr, "()")

	// Handle wildcard import
	if strings.TrimSpace(namesStr) == "*" {
		return []string{"*"}
	}

	// Split by comma and process each name
	parts := strings.Split(namesStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Handle "name as alias" format
		if strings.Contains(part, " as ") {
			nameParts := strings.Split(part, " as ")
			if len(nameParts) > 0 {
				names = append(names, strings.TrimSpace(nameParts[0]))
			}
		} else {
			names = append(names, part)
		}
	}

	return names
}

// extractModuleNameFromImport extracts the module name from import info
func (s *SystemAnalysisServiceImpl) extractModuleNameFromImport(imp *analyzer.ImportInfo) string {
	if imp == nil {
		return ""
	}

	// For relative imports, we can't determine the absolute module name without more context
	if imp.IsRelative {
		return "" // Skip relative imports for now
	}

	// For "import module" statements
	if strings.HasPrefix(imp.Statement, "import ") {
		moduleName := strings.TrimPrefix(imp.Statement, "import ")
		// Handle "import module as alias"
		if parts := strings.Split(moduleName, " as "); len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
		return strings.TrimSpace(moduleName)
	}

	// For "from module import ..." statements
	if strings.HasPrefix(imp.Statement, "from ") {
		parts := strings.Split(imp.Statement, " import ")
		if len(parts) > 0 {
			moduleName := strings.TrimPrefix(parts[0], "from ")
			moduleName = strings.TrimSpace(moduleName)
			// Remove relative dots
			moduleName = strings.TrimLeft(moduleName, ".")
			moduleName = strings.TrimSpace(moduleName)
			if moduleName != "" {
				return moduleName
			}
		}
	}

	// Fallback: use first imported name if it looks like a module
	if len(imp.ImportedNames) > 0 && imp.ImportedNames[0] != "*" {
		return imp.ImportedNames[0]
	}

	return ""
}

func minSystemAnalysis(a, b int) int {
	if a < b {
		return a
	}
	return b
}
