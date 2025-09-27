package analyzer

import (
	"math"
	"sort"
)

// CouplingMetricsCalculator calculates various coupling and quality metrics for modules
type CouplingMetricsCalculator struct {
	graph *DependencyGraph

	// Analysis options
	includeAbstractness bool
	complexityData      map[string]int     // Module name -> average complexity
	clonesData          map[string]float64 // Module name -> duplication ratio
	deadCodeData        map[string]int     // Module name -> dead code lines
}

// CouplingMetricsOptions configures metrics calculation
type CouplingMetricsOptions struct {
	IncludeAbstractness bool               // Calculate abstractness metrics
	ComplexityData      map[string]int     // Complexity data from complexity analysis
	ClonesData          map[string]float64 // Clone data from clone analysis
	DeadCodeData        map[string]int     // Dead code data from dead code analysis
}

// DefaultCouplingMetricsOptions returns default options
func DefaultCouplingMetricsOptions() *CouplingMetricsOptions {
	return &CouplingMetricsOptions{
		IncludeAbstractness: true,
		ComplexityData:      make(map[string]int),
		ClonesData:          make(map[string]float64),
		DeadCodeData:        make(map[string]int),
	}
}

// NewCouplingMetricsCalculator creates a new coupling metrics calculator
func NewCouplingMetricsCalculator(graph *DependencyGraph, options *CouplingMetricsOptions) *CouplingMetricsCalculator {
	if options == nil {
		options = DefaultCouplingMetricsOptions()
	}

	return &CouplingMetricsCalculator{
		graph:               graph,
		includeAbstractness: options.IncludeAbstractness,
		complexityData:      options.ComplexityData,
		clonesData:          options.ClonesData,
		deadCodeData:        options.DeadCodeData,
	}
}

// CalculateMetrics calculates all metrics for the dependency graph
func (calc *CouplingMetricsCalculator) CalculateMetrics() error {
	// Calculate module-level metrics
	for moduleName, node := range calc.graph.Nodes {
		metrics := calc.calculateModuleMetrics(moduleName, node)
		calc.graph.ModuleMetrics[moduleName] = metrics
	}

	// Calculate system-level metrics
	calc.calculateSystemMetrics()

	return nil
}

// calculateModuleMetrics calculates metrics for a single module
func (calc *CouplingMetricsCalculator) calculateModuleMetrics(moduleName string, node *ModuleNode) *ModuleMetrics {
	metrics := &ModuleMetrics{}

	// Basic coupling metrics (Robert Martin's metrics)
	metrics.AfferentCoupling = node.InDegree  // Ca - modules that depend on this one
	metrics.EfferentCoupling = node.OutDegree // Ce - modules this one depends on

	// Calculate Instability (I = Ce / (Ca + Ce))
	totalCoupling := metrics.AfferentCoupling + metrics.EfferentCoupling
	if totalCoupling > 0 {
		metrics.Instability = float64(metrics.EfferentCoupling) / float64(totalCoupling)
	} else {
		metrics.Instability = 0.0
	}

	// Calculate Abstractness if enabled
	if calc.includeAbstractness {
		metrics.Abstractness = calc.calculateAbstractness(node)
	}

	// Calculate Distance from Main Sequence (D = |A + I - 1|)
	metrics.Distance = math.Abs(metrics.Abstractness + metrics.Instability - 1.0)

	// Size metrics
	metrics.LinesOfCode = node.LineCount
	metrics.PublicInterface = len(node.PublicNames)

	// Quality metrics from external data
	if complexity, exists := calc.complexityData[moduleName]; exists {
		metrics.CyclomaticComplexity = complexity
	}

	return metrics
}

// calculateAbstractness calculates the abstractness of a module
func (calc *CouplingMetricsCalculator) calculateAbstractness(node *ModuleNode) float64 {
	if len(node.PublicNames) == 0 {
		return 0.0 // No public interface = not abstract
	}

	// Simple heuristic: count abstract/interface-like public names
	abstractCount := 0
	for _, name := range node.PublicNames {
		// Heuristics for abstractness:
		// - Names ending with "Interface", "Abstract", "Base"
		// - Names starting with "I" followed by uppercase (IService)
		// - Functions with "abc" decorators would need AST analysis
		if calc.isAbstractName(name) {
			abstractCount++
		}
	}

	return float64(abstractCount) / float64(len(node.PublicNames))
}

// isAbstractName checks if a name suggests abstractness
func (calc *CouplingMetricsCalculator) isAbstractName(name string) bool {
	abstractPrefixes := []string{"I"} // Interface naming convention
	abstractSuffixes := []string{"Interface", "Abstract", "Base", "ABC"}

	// Check suffixes
	for _, suffix := range abstractSuffixes {
		if len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix {
			return true
		}
	}

	// Check prefixes (IService, IRepository, etc.)
	for _, prefix := range abstractPrefixes {
		if len(name) > len(prefix)+1 &&
			name[:len(prefix)] == prefix &&
			name[len(prefix)] >= 'A' && name[len(prefix)] <= 'Z' {
			return true
		}
	}

	return false
}

// calculateSystemMetrics calculates system-wide metrics
func (calc *CouplingMetricsCalculator) calculateSystemMetrics() {
	systemMetrics := calc.graph.SystemMetrics

	// Basic counts
	systemMetrics.TotalModules = calc.graph.TotalModules
	systemMetrics.TotalDependencies = calc.graph.TotalEdges
	systemMetrics.PackageCount = len(calc.graph.GetPackages())

	if systemMetrics.TotalModules == 0 {
		return
	}

	// Aggregate metrics
	var totalFanIn, totalFanOut float64
	var totalInstability, totalAbstractness, totalDistance float64

	for _, metrics := range calc.graph.ModuleMetrics {
		totalFanIn += float64(metrics.AfferentCoupling)
		totalFanOut += float64(metrics.EfferentCoupling)
		totalInstability += metrics.Instability
		totalAbstractness += metrics.Abstractness
		totalDistance += metrics.Distance
	}

	moduleCount := float64(systemMetrics.TotalModules)

	// Calculate averages
	systemMetrics.AverageFanIn = totalFanIn / moduleCount
	systemMetrics.AverageFanOut = totalFanOut / moduleCount
	systemMetrics.DependencyRatio = float64(systemMetrics.TotalDependencies) / moduleCount
	systemMetrics.AverageInstability = totalInstability / moduleCount
	systemMetrics.AverageAbstractness = totalAbstractness / moduleCount
	systemMetrics.MainSequenceDeviation = totalDistance / moduleCount

	// Calculate modularity index
	systemMetrics.ModularityIndex = calc.calculateModularityIndex()

	// Count cyclic dependencies
	systemMetrics.CyclicDependencies = len(calc.graph.GetModulesInCycles())

	// Calculate system complexity
	systemMetrics.SystemComplexity = calc.calculateSystemComplexity()

	// Calculate max dependency depth
	systemMetrics.MaxDependencyDepth = calc.calculateMaxDependencyDepth()

	// Identify refactoring priorities
	systemMetrics.RefactoringPriority = calc.identifyRefactoringPriorities()
}

// calculateModularityIndex calculates the modularity index of the system
func (calc *CouplingMetricsCalculator) calculateModularityIndex() float64 {
	if calc.graph.TotalModules == 0 {
		return 0.0
	}

	// Modularity index based on:
	// - Package cohesion (modules within packages should be related)
	// - Inter-package coupling (should be minimized)
	// - Cycle count (should be minimized)

	packages := calc.graph.GetPackages()
	if len(packages) <= 1 {
		return 0.5 // Single package system has moderate modularity
	}

	// Calculate intra-package vs inter-package dependencies
	intraPackageDeps := 0
	interPackageDeps := 0

	for _, edge := range calc.graph.Edges {
		fromNode := calc.graph.Nodes[edge.From]
		toNode := calc.graph.Nodes[edge.To]

		if fromNode != nil && toNode != nil {
			if fromNode.Package == toNode.Package {
				intraPackageDeps++
			} else {
				interPackageDeps++
			}
		}
	}

	totalDeps := intraPackageDeps + interPackageDeps
	if totalDeps == 0 {
		return 1.0
	}

	// Good modularity has high intra-package coupling, low inter-package coupling
	cohesionRatio := float64(intraPackageDeps) / float64(totalDeps)

	// Apply penalty for cycles
	cyclePenalty := 1.0
	if len(calc.graph.CyclicGroups) > 0 {
		cyclicModules := len(calc.graph.GetModulesInCycles())
		cyclePenalty = 1.0 - (float64(cyclicModules) / float64(calc.graph.TotalModules) * 0.5)
	}

	return cohesionRatio * cyclePenalty
}

// calculateSystemComplexity calculates overall system complexity
func (calc *CouplingMetricsCalculator) calculateSystemComplexity() float64 {
	if calc.graph.TotalModules == 0 {
		return 0.0
	}

	// System complexity is a composite metric:
	// - Structural complexity (dependencies, cycles)
	// - Size complexity (number of modules)
	// - Coupling complexity (instability variance)

	// Structural complexity
	depRatio := float64(calc.graph.TotalEdges) / float64(calc.graph.TotalModules)
	structuralComplexity := math.Log2(1 + depRatio)

	// Size complexity
	sizeComplexity := math.Log2(1 + float64(calc.graph.TotalModules))

	// Coupling complexity (variance in instability)
	var instabilityVariance float64
	if len(calc.graph.ModuleMetrics) > 1 {
		mean := calc.graph.SystemMetrics.AverageInstability
		var sumSquaredDiffs float64

		for _, metrics := range calc.graph.ModuleMetrics {
			diff := metrics.Instability - mean
			sumSquaredDiffs += diff * diff
		}

		instabilityVariance = sumSquaredDiffs / float64(len(calc.graph.ModuleMetrics))
	}

	couplingComplexity := math.Sqrt(instabilityVariance) * 10 // Scale to reasonable range

	// Combine complexities with weights
	return structuralComplexity*0.4 + sizeComplexity*0.3 + couplingComplexity*0.3
}

// calculateMaxDependencyDepth calculates the maximum depth of dependency chains
func (calc *CouplingMetricsCalculator) calculateMaxDependencyDepth() int {
	maxDepth := 0

	// Start from root modules (no dependencies) and find longest paths
	rootModules := calc.graph.GetRootModules()

	for _, root := range rootModules {
		depth := calc.findMaxDepthFrom(root, make(map[string]bool))
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

// findMaxDepthFrom finds the maximum depth from a given module
func (calc *CouplingMetricsCalculator) findMaxDepthFrom(module string, visited map[string]bool) int {
	if visited[module] {
		return 0 // Avoid infinite recursion in cycles
	}

	visited[module] = true
	defer func() { visited[module] = false }()

	node := calc.graph.Nodes[module]
	if node == nil {
		return 0
	}

	maxChildDepth := 0
	for dependency := range node.Dependencies {
		childDepth := calc.findMaxDepthFrom(dependency, visited)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return maxChildDepth + 1
}

// identifyRefactoringPriorities identifies modules that need refactoring most urgently
func (calc *CouplingMetricsCalculator) identifyRefactoringPriorities() []string {
	type refactoringCandidate struct {
		module   string
		priority float64
	}

	var candidates []refactoringCandidate

	for moduleName, metrics := range calc.graph.ModuleMetrics {
		priority := 0.0

		// High priority for poor architectural position
		if metrics.Distance > 0.5 {
			priority += metrics.Distance * 50
		}

		// High priority for modules in cycles
		if calc.isModuleInCycle(moduleName) {
			priority += 30
		}

		// High priority for excessive complexity
		if metrics.CyclomaticComplexity > 20 {
			priority += float64(metrics.CyclomaticComplexity-20) * 2
		}

		if priority > 10 { // Threshold for inclusion
			candidates = append(candidates, refactoringCandidate{
				module:   moduleName,
				priority: priority,
			})
		}
	}

	// Sort by priority (highest first)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].priority > candidates[j].priority
	})

	// Return top 10 candidates
	var result []string
	maxResults := 10
	if len(candidates) < maxResults {
		maxResults = len(candidates)
	}

	for i := 0; i < maxResults; i++ {
		result = append(result, candidates[i].module)
	}

	return result
}

// isModuleInCycle checks if a module is part of any circular dependency
func (calc *CouplingMetricsCalculator) isModuleInCycle(moduleName string) bool {
	for _, cycle := range calc.graph.CyclicGroups {
		for _, module := range cycle {
			if module == moduleName {
				return true
			}
		}
	}
	return false
}

// CalculateCouplingMetrics is a convenience function for calculating metrics
func CalculateCouplingMetrics(graph *DependencyGraph, options *CouplingMetricsOptions) error {
	calculator := NewCouplingMetricsCalculator(graph, options)
	return calculator.CalculateMetrics()
}
