package analyzer

import (
	"fmt"
	"sort"
	"strings"
)

// CircularDependencyDetector detects circular dependencies using Tarjan's algorithm
type CircularDependencyDetector struct {
	graph *DependencyGraph

	// Tarjan's algorithm state
	index    int
	stack    []string
	inStack  map[string]bool
	indices  map[string]int
	lowLinks map[string]int

	// Results
	components [][]string // Strongly connected components
}

// CircularDependency represents a circular dependency relationship
type CircularDependency struct {
	Modules      []string          // Modules involved in the cycle
	Dependencies []DependencyChain // The dependency chains that form the cycle
	Severity     CycleSeverity     // Severity level of this cycle
	Size         int               // Number of modules in the cycle
	Description  string            // Human-readable description
}

// DependencyChain represents a chain of dependencies
type DependencyChain struct {
	From   string   // Starting module
	To     string   // Ending module
	Path   []string // Complete dependency path
	Length int      // Length of the chain
}

// CycleSeverity represents the severity level of a circular dependency
type CycleSeverity string

const (
	CycleSeverityLow      CycleSeverity = "low"      // Simple 2-module cycles
	CycleSeverityMedium   CycleSeverity = "medium"   // 3-5 module cycles
	CycleSeverityHigh     CycleSeverity = "high"     // 6-10 module cycles
	CycleSeverityCritical CycleSeverity = "critical" // 10+ module cycles or core infrastructure
)

// CircularDependencyResult contains the results of circular dependency analysis
type CircularDependencyResult struct {
	HasCircularDependencies bool                  // True if any cycles were found
	TotalCycles             int                   // Total number of cycles detected
	TotalModulesInCycles    int                   // Total number of modules involved in cycles
	CircularDependencies    []*CircularDependency // All detected circular dependencies

	// Severity breakdown
	LowSeverityCycles      int // Number of low severity cycles
	MediumSeverityCycles   int // Number of medium severity cycles
	HighSeverityCycles     int // Number of high severity cycles
	CriticalSeverityCycles int // Number of critical severity cycles

	// Most problematic cycles
	LargestCycle       *CircularDependency // Cycle with most modules
	MostComplexCycle   *CircularDependency // Cycle with most dependency chains
	CoreInfrastructure []string            // Modules that appear in multiple cycles
}

// NewCircularDependencyDetector creates a new circular dependency detector
func NewCircularDependencyDetector(graph *DependencyGraph) *CircularDependencyDetector {
	return &CircularDependencyDetector{
		graph:      graph,
		inStack:    make(map[string]bool),
		indices:    make(map[string]int),
		lowLinks:   make(map[string]int),
		components: make([][]string, 0),
	}
}

// DetectCircularDependencies detects all circular dependencies in the graph
func (cdd *CircularDependencyDetector) DetectCircularDependencies() *CircularDependencyResult {
	// Reset state
	cdd.resetState()

	// Run Tarjan's algorithm to find strongly connected components
	cdd.findStronglyConnectedComponents()

	// Process components to identify circular dependencies
	circularDeps := cdd.processComponents()

	// Create result
	result := &CircularDependencyResult{
		HasCircularDependencies: len(circularDeps) > 0,
		TotalCycles:             len(circularDeps),
		CircularDependencies:    circularDeps,
	}

	// Calculate statistics
	cdd.calculateStatistics(result)

	// Update graph with results
	cdd.updateGraphWithCycles()

	return result
}

// resetState resets the detector state for a new analysis
func (cdd *CircularDependencyDetector) resetState() {
	cdd.index = 0
	cdd.stack = make([]string, 0)
	cdd.inStack = make(map[string]bool)
	cdd.indices = make(map[string]int)
	cdd.lowLinks = make(map[string]int)
	cdd.components = make([][]string, 0)
}

// findStronglyConnectedComponents implements Tarjan's algorithm
func (cdd *CircularDependencyDetector) findStronglyConnectedComponents() {
	// Run Tarjan's algorithm on each unvisited node
	for moduleName := range cdd.graph.Nodes {
		if _, visited := cdd.indices[moduleName]; !visited {
			cdd.strongConnect(moduleName)
		}
	}
}

// strongConnect is the core of Tarjan's algorithm
func (cdd *CircularDependencyDetector) strongConnect(module string) {
	// Set the depth index for this node to the smallest unused index
	cdd.indices[module] = cdd.index
	cdd.lowLinks[module] = cdd.index
	cdd.index++

	// Push the node onto the stack
	cdd.stack = append(cdd.stack, module)
	cdd.inStack[module] = true

	// Consider successors of the current module
	if node := cdd.graph.Nodes[module]; node != nil {
		for dependency := range node.Dependencies {
			if _, visited := cdd.indices[dependency]; !visited {
				// Successor has not yet been visited; recurse on it
				cdd.strongConnect(dependency)
				cdd.lowLinks[module] = minLowLink(cdd.lowLinks[module], cdd.lowLinks[dependency])
			} else if cdd.inStack[dependency] {
				// Successor is in stack and hence in the current SCC
				cdd.lowLinks[module] = minLowLink(cdd.lowLinks[module], cdd.indices[dependency])
			}
		}
	}

	// If this is a root node, pop the stack and create an SCC
	if cdd.lowLinks[module] == cdd.indices[module] {
		var component []string
		for {
			// Pop from stack
			top := cdd.stack[len(cdd.stack)-1]
			cdd.stack = cdd.stack[:len(cdd.stack)-1]
			cdd.inStack[top] = false
			component = append(component, top)

			if top == module {
				break
			}
		}

		// Only add components with more than one module (cycles)
		if len(component) > 1 {
			// Sort component for consistent ordering
			sort.Strings(component)
			cdd.components = append(cdd.components, component)
		}
	}
}

// processComponents converts strongly connected components to circular dependencies
func (cdd *CircularDependencyDetector) processComponents() []*CircularDependency {
	var circularDeps []*CircularDependency

	for _, component := range cdd.components {
		if len(component) <= 1 {
			continue // Skip non-circular components
		}

		// Create circular dependency
		circularDep := &CircularDependency{
			Modules: component,
			Size:    len(component),
		}

		// Find all dependency chains within the component
		circularDep.Dependencies = cdd.findDependencyChains(component)

		// Assess severity
		circularDep.Severity = cdd.assessCycleSeverity(circularDep)

		// Generate description
		circularDep.Description = cdd.generateCycleDescription(circularDep)

		circularDeps = append(circularDeps, circularDep)
	}

	// Sort by severity and size
	sort.Slice(circularDeps, func(i, j int) bool {
		if circularDeps[i].Severity != circularDeps[j].Severity {
			return cdd.severityOrder(circularDeps[i].Severity) > cdd.severityOrder(circularDeps[j].Severity)
		}
		return circularDeps[i].Size > circularDeps[j].Size
	})

	return circularDeps
}

// findDependencyChains finds all dependency chains within a circular component
func (cdd *CircularDependencyDetector) findDependencyChains(modules []string) []DependencyChain {
	var chains []DependencyChain
	moduleSet := make(map[string]bool)

	// Create module set for quick lookup
	for _, module := range modules {
		moduleSet[module] = true
	}

	// Find direct dependencies between modules in the component
	for _, from := range modules {
		if node := cdd.graph.Nodes[from]; node != nil {
			for to := range node.Dependencies {
				if moduleSet[to] {
					// Find the shortest path from 'from' to 'to' within the component
					path := cdd.findPathInComponent(from, to, moduleSet)
					if len(path) > 0 {
						chain := DependencyChain{
							From:   from,
							To:     to,
							Path:   path,
							Length: len(path) - 1, // Number of edges
						}
						chains = append(chains, chain)
					}
				}
			}
		}
	}

	return chains
}

// findPathInComponent finds a path between two modules within a component
func (cdd *CircularDependencyDetector) findPathInComponent(from, to string, moduleSet map[string]bool) []string {
	if from == to {
		return []string{from}
	}

	// Simple BFS within the component
	queue := [][]string{{from}}
	visited := make(map[string]bool)
	visited[from] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		current := path[len(path)-1]

		if node := cdd.graph.Nodes[current]; node != nil {
			for dependency := range node.Dependencies {
				if !moduleSet[dependency] {
					continue // Skip modules outside the component
				}

				if dependency == to {
					return append(path, dependency)
				}

				if !visited[dependency] {
					visited[dependency] = true
					newPath := make([]string, len(path)+1)
					copy(newPath, path)
					newPath[len(path)] = dependency
					queue = append(queue, newPath)
				}
			}
		}
	}

	return nil // No path found
}

// assessCycleSeverity determines the severity of a circular dependency
func (cdd *CircularDependencyDetector) assessCycleSeverity(cycle *CircularDependency) CycleSeverity {
	size := cycle.Size

	// Check if any modules are core infrastructure (high fan-in)
	hasCore := false
	for _, module := range cycle.Modules {
		if node := cdd.graph.Nodes[module]; node != nil {
			if node.InDegree > 10 { // High fan-in indicates core infrastructure
				hasCore = true
				break
			}
		}
	}

	if hasCore || size >= 10 {
		return CycleSeverityCritical
	} else if size >= 6 {
		return CycleSeverityHigh
	} else if size >= 3 {
		return CycleSeverityMedium
	}

	return CycleSeverityLow
}

// generateCycleDescription generates a human-readable description
func (cdd *CircularDependencyDetector) generateCycleDescription(cycle *CircularDependency) string {
	if len(cycle.Modules) == 2 {
		return fmt.Sprintf("Direct circular dependency between %s and %s",
			cycle.Modules[0], cycle.Modules[1])
	}

	return fmt.Sprintf("Circular dependency involving %d modules: %s",
		len(cycle.Modules), strings.Join(cycle.Modules, " → "))
}

// calculateStatistics calculates result statistics
func (cdd *CircularDependencyDetector) calculateStatistics(result *CircularDependencyResult) {
	if len(result.CircularDependencies) == 0 {
		return
	}

	// Count modules in cycles
	modulesInCycles := make(map[string]int)
	var largest *CircularDependency
	var mostComplex *CircularDependency

	maxSize := 0
	maxChains := 0

	// Process each circular dependency
	for _, cycle := range result.CircularDependencies {
		// Count by severity
		switch cycle.Severity {
		case CycleSeverityLow:
			result.LowSeverityCycles++
		case CycleSeverityMedium:
			result.MediumSeverityCycles++
		case CycleSeverityHigh:
			result.HighSeverityCycles++
		case CycleSeverityCritical:
			result.CriticalSeverityCycles++
		}

		// Track modules
		for _, module := range cycle.Modules {
			modulesInCycles[module]++
		}

		// Find largest cycle
		if cycle.Size > maxSize {
			maxSize = cycle.Size
			largest = cycle
		}

		// Find most complex cycle (most dependency chains)
		chainCount := len(cycle.Dependencies)
		if chainCount > maxChains {
			maxChains = chainCount
			mostComplex = cycle
		}
	}

	result.TotalModulesInCycles = len(modulesInCycles)
	result.LargestCycle = largest
	result.MostComplexCycle = mostComplex

	// Identify core infrastructure (modules in multiple cycles)
	var coreInfra []string
	for module, count := range modulesInCycles {
		if count > 1 {
			coreInfra = append(coreInfra, module)
		}
	}
	sort.Strings(coreInfra)
	result.CoreInfrastructure = coreInfra
}

// updateGraphWithCycles updates the graph with cycle information
func (cdd *CircularDependencyDetector) updateGraphWithCycles() {
	// Clear existing cycle information
	cdd.graph.CyclicGroups = make([][]string, 0)

	// Add detected cycles
	for _, component := range cdd.components {
		if len(component) > 1 {
			cdd.graph.CyclicGroups = append(cdd.graph.CyclicGroups, component)
		}
	}
}

// severityOrder returns numeric order for severity comparison
func (cdd *CircularDependencyDetector) severityOrder(severity CycleSeverity) int {
	switch severity {
	case CycleSeverityCritical:
		return 4
	case CycleSeverityHigh:
		return 3
	case CycleSeverityMedium:
		return 2
	case CycleSeverityLow:
		return 1
	default:
		return 0
	}
}

// Utility functions

// minLowLink returns the minimum of two integers for low-link calculation
func minLowLink(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// DetectCircularDependencies is a convenience function for detecting cycles in a graph
func DetectCircularDependencies(graph *DependencyGraph) *CircularDependencyResult {
	detector := NewCircularDependencyDetector(graph)
	return detector.DetectCircularDependencies()
}

// HasCircularDependencies quickly checks if a graph has any circular dependencies
func HasCircularDependencies(graph *DependencyGraph) bool {
	detector := NewCircularDependencyDetector(graph)
	result := detector.DetectCircularDependencies()
	return result.HasCircularDependencies
}

// FindSimpleCycles finds all simple cycles (2-module cycles) in the graph
func FindSimpleCycles(graph *DependencyGraph) []*CircularDependency {
	var simpleCycles []*CircularDependency

	// Check each pair of modules for mutual dependencies
	modules := graph.GetModuleNames()
	for i, moduleA := range modules {
		nodeA := graph.Nodes[moduleA]
		if nodeA == nil {
			continue
		}

		for j := i + 1; j < len(modules); j++ {
			moduleB := modules[j]
			nodeB := graph.Nodes[moduleB]
			if nodeB == nil {
				continue
			}

			// Check if A depends on B and B depends on A
			if nodeA.Dependencies[moduleB] && nodeB.Dependencies[moduleA] {
				cycle := &CircularDependency{
					Modules:     []string{moduleA, moduleB},
					Size:        2,
					Severity:    CycleSeverityLow,
					Description: fmt.Sprintf("Direct circular dependency between %s and %s", moduleA, moduleB),
					Dependencies: []DependencyChain{
						{From: moduleA, To: moduleB, Path: []string{moduleA, moduleB}, Length: 1},
						{From: moduleB, To: moduleA, Path: []string{moduleB, moduleA}, Length: 1},
					},
				}
				simpleCycles = append(simpleCycles, cycle)
			}
		}
	}

	return simpleCycles
}

// GetCycleBreakingSuggestions suggests module refactoring to break cycles
func GetCycleBreakingSuggestions(result *CircularDependencyResult) []string {
	if !result.HasCircularDependencies {
		return nil
	}

	var suggestions []string

	// Suggest breaking the largest cycle first
	if result.LargestCycle != nil {
		suggestions = append(suggestions,
			fmt.Sprintf("Break the largest cycle (%d modules) by extracting common functionality into a separate module",
				result.LargestCycle.Size))
	}

	// Suggest refactoring core infrastructure
	if len(result.CoreInfrastructure) > 0 {
		suggestions = append(suggestions,
			fmt.Sprintf("Refactor core modules that appear in multiple cycles: %s",
				strings.Join(result.CoreInfrastructure, ", ")))
	}

	// Suggest dependency inversion for critical cycles
	if result.CriticalSeverityCycles > 0 {
		suggestions = append(suggestions,
			"Consider applying Dependency Inversion Principle to critical cycles")
	}

	return suggestions
}
