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
		DependencyAnalysis:  dependencyResult,
		ArchitectureAnalysis: architectureResult,
		QualityMetrics:       qualityResult,
		GeneratedAt:         time.Now(),
		Duration:            time.Since(startTime).Milliseconds(),
		Version:             version.Version,
		Warnings:            warnings,
		Errors:              errors,
	}

	return response, nil
}

// AnalyzeDependencies performs dependency analysis only
func (s *SystemAnalysisServiceImpl) AnalyzeDependencies(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.DependencyAnalysisResult, error) {
	// Build dependency graph for all files
	graph := analyzer.NewDependencyGraph("")
	filesProcessed := 0

	for _, filePath := range req.Paths {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("dependency analysis cancelled: %w", ctx.Err())
		default:
		}

		// Parse file and extract imports
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files we can't read
		}

		// Parse the Python code
		parseResult, err := s.parser.Parse(ctx, content)
		if err != nil {
			continue // Skip files with syntax errors
		}

		// Get module name from file path
		moduleName := s.getModuleNameFromPath(filePath)
		
		// Add module to graph
		graph.AddModule(moduleName, filePath)

		// Extract dependencies using the module analyzer
		dependencies := s.extractImportsFromAST(parseResult.AST, filePath)
		
		// Debug output disabled
		// Uncomment for debugging:
		// if len(dependencies) > 0 {
		// 	fmt.Printf("DEBUG: File %s (module: %s) has %d imports\n", filePath, moduleName, len(dependencies))
		// 	for _, dep := range dependencies {
		// 		var targetModule string
		// 		if dep.IsRelative {
		// 			targetModule = s.resolveRelativeImport(dep, filePath)
		// 			fmt.Printf("  - [RELATIVE] Statement: %s -> Module: %s\n", dep.Statement, targetModule)
		// 		} else {
		// 			targetModule = s.extractModuleNameFromImport(dep)
		// 			fmt.Printf("  - Statement: %s -> Module: %s\n", dep.Statement, targetModule)
		// 		}
		// 	}
		// }

		// Add dependencies to graph
		for _, dep := range dependencies {
			// Extract module name from import statement
			var targetModule string
			if dep.IsRelative {
				// Resolve relative imports based on current module context
				targetModule = s.resolveRelativeImport(dep, filePath)
			} else {
				targetModule = s.extractModuleNameFromImport(dep)
			}
			
			if targetModule != "" && targetModule != moduleName {
				// Determine edge type
				var edgeType analyzer.DependencyEdgeType
				switch {
				case dep.IsRelative:
					edgeType = analyzer.DependencyEdgeRelative
				case strings.Contains(dep.Statement, "from"):
					edgeType = analyzer.DependencyEdgeFromImport
				default:
					edgeType = analyzer.DependencyEdgeImport
				}

				graph.AddDependency(moduleName, targetModule, edgeType, dep)
			}
		}

		filesProcessed++
	}

	// If no modules were processed, return empty result
	if filesProcessed == 0 {
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
	err := metricsCalculator.CalculateMetrics()
	if err != nil {
		return nil, err
	}
	couplingResults := s.extractCouplingResult(graph)

	// Build dependency matrix
	matrix := s.buildDependencyMatrix(graph)

	// Find longest dependency chains
	longestChains := s.findLongestChains(graph, 10) // Top 10 chains

	// Create dependency analysis result
	result := &domain.DependencyAnalysisResult{
		TotalModules:      graph.TotalModules,
		TotalDependencies: graph.TotalEdges,
		RootModules:       graph.GetRootModules(),
		LeafModules:       graph.GetLeafModules(),
		ModuleMetrics:     make(map[string]*domain.ModuleDependencyMetrics), // Mock for now
		DependencyMatrix:  matrix,
		CircularDependencies: s.convertCircularResults(circularResult),
		CouplingAnalysis:  s.convertCouplingResults(couplingResults),
		LongestChains:     longestChains,
		MaxDepth:          s.calculateMaxDepth(graph),
	}

	return result, nil
}

// AnalyzeArchitecture performs architecture validation only
func (s *SystemAnalysisServiceImpl) AnalyzeArchitecture(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.ArchitectureAnalysisResult, error) {
	// Simple architecture analysis implementation
	layerAnalysis := &domain.LayerAnalysis{
		LayersAnalyzed:     s.countLayers(req.Paths),
		LayerViolations:    []domain.LayerViolation{}, // No violations for now
		LayerCoupling:      make(map[string]map[string]int),
		LayerCohesion:      make(map[string]float64),
		ProblematicLayers:  []string{},
	}

	result := &domain.ArchitectureAnalysisResult{
		ComplianceScore:        0.85, // Mock compliance score
		TotalViolations:        0,    // No violations detected yet
		TotalRules:             5,    // Mock number of rules
		LayerAnalysis:          layerAnalysis,
		CohesionAnalysis:       nil, // Not implemented yet
		ResponsibilityAnalysis: nil, // Not implemented yet
		Violations:             []domain.ArchitectureViolation{},
		SeverityBreakdown:      make(map[domain.ViolationSeverity]int),
		Recommendations:        []domain.ArchitectureRecommendation{},
		RefactoringTargets:     []string{},
	}

	return result, nil
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
		OverallQuality:       s.calculateOverallQuality(systemMetrics),
		MaintainabilityIndex: systemMetrics.MaintainabilityIndex,
		TechnicalDebtTotal:   systemMetrics.TechnicalDebtTotal,
		ModularityIndex:      systemMetrics.ModularityIndex,
		SystemInstability:    systemMetrics.AverageInstability,
		SystemAbstractness:   systemMetrics.AverageAbstractness,
		MainSequenceDistance: systemMetrics.MainSequenceDeviation,
		SystemComplexity:     systemMetrics.SystemComplexity,
		MaxDependencyDepth:   systemMetrics.MaxDependencyDepth,
		AverageFanIn:         systemMetrics.AverageFanIn,
		AverageFanOut:        systemMetrics.AverageFanOut,
		HighQualityModules:   []string{},     // Mock for now
		ModerateQualityModules: []string{},   // Mock for now
		LowQualityModules:    systemMetrics.RefactoringPriority[:minSystemAnalysis(3, len(systemMetrics.RefactoringPriority))],
		CriticalModules:      []string{},     // Mock for now
		QualityTrends:        make(map[string]float64),
		HotSpots:            []string{},      // Mock for now
		RefactoringTargets:  []domain.RefactoringTarget{}, // Mock for now
	}

	return result, nil
}

// Helper methods

func (s *SystemAnalysisServiceImpl) getModuleNameFromPath(filePath string) string {
	// Convert file path to Python module name
	// First, find the app directory as the root
	parts := strings.Split(filePath, string(filepath.Separator))
	
	// Find where "app" directory starts (for this project structure)
	appIndex := -1
	for i, part := range parts {
		if part == "app" {
			appIndex = i
			break
		}
	}
	
	// If we found "app", build the module path from there
	if appIndex >= 0 && appIndex < len(parts)-1 {
		moduleParts := parts[appIndex:]
		
		// Remove the file extension from the last part
		lastIndex := len(moduleParts) - 1
		moduleParts[lastIndex] = strings.TrimSuffix(moduleParts[lastIndex], filepath.Ext(moduleParts[lastIndex]))
		
		// Handle __init__.py files - use parent directory name
		if moduleParts[lastIndex] == "__init__" {
			moduleParts = moduleParts[:lastIndex]
		}
		
		// Join with dots to create Python module path
		return strings.Join(moduleParts, ".")
	}
	
	// Fallback to simple filename if we can't determine the module structure
	moduleName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	if moduleName == "__init__" {
		return filepath.Base(filepath.Dir(filePath))
	}
	
	return moduleName
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
		AverageCoupling:        results.AverageFanIn + results.AverageFanOut,
		AverageInstability:     results.AverageInstability,
		MainSequenceDeviation: results.MainSequenceDeviation,
		HighlyCoupledModules:   results.RefactoringPriority,
		ZoneOfPain:             s.extractZoneOfPain(results),
		ZoneOfUselessness:      s.extractZoneOfUselessness(results),
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
		HasCircularDependencies:   len(circularDeps) > 0,
		TotalCycles:              len(circularDeps),
		TotalModulesInCycles:     result.TotalModulesInCycles,
		CircularDependencies:     circularDeps,
		CycleBreakingSuggestions: []string{}, // Mock for now
	}
}

// Architecture analysis helper methods (simplified)

func (s *SystemAnalysisServiceImpl) countArchitecturalLayers(paths []string) int {
	return s.countLayers(paths)
}

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

func (s *SystemAnalysisServiceImpl) countLayers(paths []string) int {
	layers := make(map[string]bool)
	
	for _, path := range paths {
		// Extract layer from directory structure
		parts := strings.Split(filepath.Dir(path), string(filepath.Separator))
		for _, part := range parts {
			if s.isArchitecturalLayer(part) {
				layers[part] = true
			}
		}
	}
	
	return len(layers)
}

func (s *SystemAnalysisServiceImpl) calculateCodeHealth(metrics *analyzer.SystemMetrics) float64 {
	// Simple health calculation based on multiple factors
	health := 1.0
	
	// Penalize high instability
	if metrics.AverageInstability > 0.8 {
		health -= 0.2
	}
	
	// Penalize high main sequence deviation
	if metrics.MainSequenceDeviation > 0.5 {
		health -= 0.3
	}
	
	// Penalize cyclic dependencies
	if metrics.CyclicDependencies > 0 {
		health -= 0.4
	}
	
	if health < 0 {
		health = 0
	}
	
	return health
}

// Removed evaluateQualityGate - QualityGate is not part of the QualityMetricsResult domain

func (s *SystemAnalysisServiceImpl) generateRecommendations(metrics *analyzer.SystemMetrics) []string {
	var recommendations []string
	
	if metrics.CyclicDependencies > 0 {
		recommendations = append(recommendations, "Break circular dependencies to improve maintainability")
	}
	
	if metrics.AverageInstability > 0.7 {
		recommendations = append(recommendations, "Reduce coupling between modules")
	}
	
	if metrics.MainSequenceDeviation > 0.4 {
		recommendations = append(recommendations, "Balance abstractness and instability of modules")
	}
	
	if len(metrics.RefactoringPriority) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Consider refactoring high-priority modules: %v", metrics.RefactoringPriority[:minSystemAnalysis(3, len(metrics.RefactoringPriority))]))
	}
	
	return recommendations
}

// Additional helper methods

func (s *SystemAnalysisServiceImpl) isArchitecturalLayer(name string) bool {
	layers := []string{"presentation", "application", "domain", "infrastructure", "data", "service", "controller", "model", "view"}
	for _, layer := range layers {
		if strings.Contains(strings.ToLower(name), layer) {
			return true
		}
	}
	return false
}

func (s *SystemAnalysisServiceImpl) getLayerDescription(name string) string {
	descriptions := map[string]string{
		"presentation":   "User interface and presentation logic",
		"application":    "Application services and use cases",
		"domain":         "Business logic and domain entities",
		"infrastructure": "External concerns and technical implementation",
		"data":           "Data access and persistence",
		"service":        "Service layer implementation",
		"controller":     "Request handling and routing",
		"model":          "Data models and entities",
		"view":           "User interface templates and views",
	}
	
	for key, desc := range descriptions {
		if strings.Contains(strings.ToLower(name), key) {
			return desc
		}
	}
	
	return "Application component"
}

func (s *SystemAnalysisServiceImpl) getLayerLevel(name string) int {
	levels := map[string]int{
		"presentation":   1,
		"controller":     1,
		"view":           1,
		"application":    2,
		"service":        2,
		"domain":         3,
		"model":          3,
		"infrastructure": 4,
		"data":           4,
	}
	
	for key, level := range levels {
		if strings.Contains(strings.ToLower(name), key) {
			return level
		}
	}
	
	return 2 // Default level
}

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

func (s *SystemAnalysisServiceImpl) calculateCycleSeverity(size int) string {
	switch {
	case size >= 5:
		return "Critical"
	case size >= 3:
		return "Warning"
	default:
		return "Info"
	}
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
		AverageFanIn:            float64(graph.TotalEdges) / float64(graph.TotalModules),
		AverageFanOut:           float64(graph.TotalEdges) / float64(graph.TotalModules), 
		AverageInstability:      0.5, // Mock value
		AverageAbstractness:     0.3, // Mock value
		MainSequenceDeviation:  0.2, // Mock value
		MaintainabilityIndex:    75.0, // Mock value
		TechnicalDebtTotal:      10.0, // Mock value
		ModularityIndex:         0.8, // Mock value
		SystemComplexity:        float64(graph.TotalModules * 2),
		MaxDependencyDepth:      s.calculateMaxDepth(graph),
		CyclicDependencies:      0, // Will be updated by circular analysis
		RefactoringPriority:     []string{}, // Mock empty
	}
}

// extractSystemMetrics extracts system-wide metrics from the dependency graph
func (s *SystemAnalysisServiceImpl) extractSystemMetrics(graph *analyzer.DependencyGraph) *analyzer.SystemMetrics {
	return s.extractCouplingResult(graph) // Same implementation for now
}

// extractImportsFromAST extracts import information from AST
// TEMPORARY: Using regex-based extraction due to AST builder issues
func (s *SystemAnalysisServiceImpl) extractImportsFromAST(ast *parser.Node, filePath string) []*analyzer.ImportInfo {
	// Read the source file directly for regex-based parsing
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Fall back to empty imports if file can't be read
		return []*analyzer.ImportInfo{}
	}
	
	return s.extractImportsWithRegex(string(content))
}

// extractImportsWithRegex uses regex to extract imports from Python source
// This is a temporary solution until the AST builder is fixed
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

// walkAST walks the AST recursively
func (s *SystemAnalysisServiceImpl) walkAST(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil || !visitor(node) {
		return
	}
	
	for _, child := range node.Children {
		s.walkAST(child, visitor)
	}
	
	if node.Body != nil {
		for _, stmt := range node.Body {
			s.walkAST(stmt, visitor)
		}
	}
}

// resolveRelativeImport resolves a relative import to an absolute module name
func (s *SystemAnalysisServiceImpl) resolveRelativeImport(imp *analyzer.ImportInfo, filePath string) string {
	if imp == nil || !imp.IsRelative {
		return ""
	}
	
	// Get the current module path
	moduleName := s.getModuleNameFromPath(filePath)
	
	// Split module path into parts
	parts := strings.Split(moduleName, ".")
	
	// For level 1 (from . import x), use current package
	// For level 2 (from .. import x), go up one level, etc.
	if imp.Level > 0 && imp.Level < len(parts) {
		// Go up 'Level' directories
		baseParts := parts[:len(parts)-imp.Level]
		
		// Extract the module name from the statement
		modulePart := ""
		if strings.Contains(imp.Statement, " import ") {
			statementParts := strings.Split(imp.Statement, " import ")
			if len(statementParts) > 0 {
				fromPart := strings.TrimPrefix(statementParts[0], "from ")
				fromPart = strings.TrimSpace(fromPart)
				// Remove the dots
				fromPart = strings.TrimLeft(fromPart, ".")
				fromPart = strings.TrimSpace(fromPart)
				if fromPart != "" {
					modulePart = fromPart
				}
			}
		}
		
		// Combine base path with module part
		if modulePart != "" {
			return strings.Join(append(baseParts, modulePart), ".")
		}
		
		// If importing directly from parent (from . import x), 
		// we're importing from the same package
		if imp.Level == 1 && len(imp.ImportedNames) > 0 {
			// Return the package path + imported name as module
			// This helps track intra-package dependencies
			if len(baseParts) > 0 {
				return strings.Join(baseParts, ".")
			}
		}
		
		// For higher levels, just return the parent package
		if len(baseParts) > 0 {
			return strings.Join(baseParts, ".")
		}
	}
	
	return ""
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