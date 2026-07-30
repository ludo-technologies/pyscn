package analyzer

import (
	"context"
	"fmt"
	"math"
	"sort"

	coregraph "github.com/ludo-technologies/polyscan/core/graph"
)

const (
	mainSequenceMaxDistance = 0.2
	zoneMinDistance         = 0.5
	lowInstability          = 0.3
	highInstability         = 0.7
	lowAbstractness         = 0.3
	highAbstractness        = 0.7
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

// CalculateMetrics calculates all metrics using topology from the same graph.
func (calc *CouplingMetricsCalculator) CalculateMetrics(
	ctx context.Context,
	topology *DependencyTopology,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("calculate coupling metrics: %w", err)
	}
	if topology == nil {
		return fmt.Errorf("calculate coupling metrics: dependency topology is required")
	}
	if topology.graph != calc.graph {
		return fmt.Errorf("calculate coupling metrics: dependency topology belongs to another graph")
	}
	if topology.graphRevision != calc.graph.topologyRevision {
		return fmt.Errorf("calculate coupling metrics: dependency graph changed after topology analysis")
	}

	config := coregraph.CouplingConfig{
		AbstractnessFunc: func(moduleName string) (float64, error) {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
			if !calc.includeAbstractness {
				return 0, nil
			}
			node := calc.graph.Nodes[moduleName]
			if node == nil {
				return 0, fmt.Errorf("calculate abstractness: module %q not found", moduleName)
			}
			return calc.calculateAbstractness(node), nil
		},
	}

	coreMetrics, err := coregraph.ComputeCouplingMetrics(calc.graph, config)
	if err != nil {
		return fmt.Errorf("calculate coupling metrics: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("calculate coupling metrics: %w", err)
	}

	moduleMetrics := make(map[string]*ModuleMetrics, len(coreMetrics))
	for moduleName, coupling := range coreMetrics {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("calculate coupling metrics: %w", err)
		}
		node := calc.graph.Nodes[moduleName]
		if node == nil {
			return fmt.Errorf("calculate module metrics: module %q not found", moduleName)
		}

		metrics := &ModuleMetrics{
			AfferentCoupling:   coupling.Ca,
			EfferentCoupling:   coupling.Ce,
			Instability:        coupling.Instability,
			Abstractness:       coupling.Abstractness,
			Distance:           coupling.Distance,
			LinesOfCode:        node.LineCount,
			ClassCount:         node.ClassCount,
			AbstractClassCount: node.AbstractClassCount,
			PublicInterface:    len(node.PublicNames),
		}
		if complexity, exists := calc.complexityData[moduleName]; exists {
			metrics.CyclomaticComplexity = complexity
		}
		moduleMetrics[moduleName] = metrics
	}

	systemMetrics, err := calc.calculateSystemMetrics(ctx, moduleMetrics, topology)
	if err != nil {
		return fmt.Errorf("calculate system metrics: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("calculate coupling metrics: %w", err)
	}

	calc.graph.ModuleMetrics = moduleMetrics
	calc.graph.SystemMetrics = systemMetrics
	return nil
}

// calculateAbstractness calculates the abstractness of a module
func (calc *CouplingMetricsCalculator) calculateAbstractness(node *ModuleNode) float64 {
	if node.ClassCount == 0 {
		return 0.0
	}

	return float64(node.AbstractClassCount) / float64(node.ClassCount)
}

// calculateSystemMetrics calculates system-wide metrics without mutating the graph.
func (calc *CouplingMetricsCalculator) calculateSystemMetrics(
	ctx context.Context,
	moduleMetrics map[string]*ModuleMetrics,
	topology *DependencyTopology,
) (*SystemMetrics, error) {
	systemMetrics := &SystemMetrics{
		TotalModules:      calc.graph.TotalModules,
		TotalDependencies: calc.graph.TotalEdges,
		PackageCount:      len(calc.graph.GetPackages()),
	}

	if systemMetrics.TotalModules == 0 {
		return systemMetrics, nil
	}

	// Aggregate metrics
	var totalFanIn, totalFanOut float64
	var totalInstability, totalAbstractness, totalDistance float64

	for _, metrics := range moduleMetrics {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
	systemMetrics.SystemComplexity = calc.calculateSystemComplexity(
		moduleMetrics,
		systemMetrics.AverageInstability,
	)

	systemMetrics.MaxDependencyDepth = topology.maxDepth

	// Identify refactoring priorities
	systemMetrics.RefactoringPriority = calc.identifyRefactoringPriorities(moduleMetrics)
	systemMetrics.StableModules = modulesMatching(moduleMetrics, isStableModule)
	systemMetrics.InstableModules = modulesMatching(moduleMetrics, isInstableModule)
	systemMetrics.ZoneOfPain = modulesMatching(moduleMetrics, isZoneOfPain)
	systemMetrics.ZoneOfUselessness = modulesMatching(moduleMetrics, isZoneOfUselessness)
	systemMetrics.MainSequence = modulesMatching(moduleMetrics, isOnMainSequence)

	return systemMetrics, nil
}

func isStableModule(metrics *ModuleMetrics) bool {
	return metrics.Instability <= lowInstability
}

func isInstableModule(metrics *ModuleMetrics) bool {
	return metrics.Instability >= highInstability
}

func isZoneOfPain(metrics *ModuleMetrics) bool {
	return metrics.Distance >= zoneMinDistance &&
		metrics.AfferentCoupling >= 2 &&
		metrics.Instability <= lowInstability &&
		metrics.Abstractness <= lowAbstractness
}

func isZoneOfUselessness(metrics *ModuleMetrics) bool {
	return metrics.Distance >= zoneMinDistance &&
		metrics.Instability >= highInstability &&
		metrics.Abstractness >= highAbstractness
}

func isOnMainSequence(metrics *ModuleMetrics) bool {
	return metrics.Distance <= mainSequenceMaxDistance
}

func modulesMatching(moduleMetrics map[string]*ModuleMetrics, match func(*ModuleMetrics) bool) []string {
	modules := make([]string, 0)
	for moduleName, metrics := range moduleMetrics {
		if metrics != nil && match(metrics) {
			modules = append(modules, moduleName)
		}
	}
	sort.Strings(modules)
	return modules
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

// calculateSystemComplexity calculates overall system complexity.
func (calc *CouplingMetricsCalculator) calculateSystemComplexity(
	moduleMetrics map[string]*ModuleMetrics,
	averageInstability float64,
) float64 {
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
	if len(moduleMetrics) > 1 {
		var sumSquaredDiffs float64

		for _, metrics := range moduleMetrics {
			diff := metrics.Instability - averageInstability
			sumSquaredDiffs += diff * diff
		}

		instabilityVariance = sumSquaredDiffs / float64(len(moduleMetrics))
	}

	couplingComplexity := math.Sqrt(instabilityVariance) * 10 // Scale to reasonable range

	// Combine complexities with weights
	return structuralComplexity*0.4 + sizeComplexity*0.3 + couplingComplexity*0.3
}

// identifyRefactoringPriorities identifies modules that need refactoring most urgently.
func (calc *CouplingMetricsCalculator) identifyRefactoringPriorities(
	moduleMetrics map[string]*ModuleMetrics,
) []string {
	type refactoringCandidate struct {
		module   string
		priority float64
	}

	var candidates []refactoringCandidate

	for moduleName, metrics := range moduleMetrics {
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
		if candidates[i].priority == candidates[j].priority {
			return candidates[i].module < candidates[j].module
		}
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

// CalculateCouplingMetrics is a convenience function for calculating metrics.
func CalculateCouplingMetrics(
	ctx context.Context,
	graph *DependencyGraph,
	topology *DependencyTopology,
	options *CouplingMetricsOptions,
) error {
	calculator := NewCouplingMetricsCalculator(graph, options)
	return calculator.CalculateMetrics(ctx, topology)
}
