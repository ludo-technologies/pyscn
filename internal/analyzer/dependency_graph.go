package analyzer

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ModuleNode represents a module in the dependency graph
type ModuleNode struct {
	// Module identification
	Name         string // Module name (e.g., "mypackage.submodule")
	FilePath     string // Absolute path to the Python file
	RelativePath string // Relative path from project root
	Package      string // Package name (e.g., "mypackage")
	IsPackage    bool   // True if this represents a package (__init__.py)

	// Dependencies
	Imports      []string        // Direct imports from this module
	ImportedBy   []string        // Modules that import this module
	Dependencies map[string]bool // Set of modules this module depends on
	Dependents   map[string]bool // Set of modules that depend on this module

	// Metrics
	InDegree  int // Number of incoming dependencies
	OutDegree int // Number of outgoing dependencies

	// Module information
	LineCount     int      // Total lines in the module
	FunctionCount int      // Number of functions defined
	ClassCount    int      // Number of classes defined
	PublicNames   []string // Public names exported by this module
}

// DependencyEdge represents a dependency relationship between modules
type DependencyEdge struct {
	From       string             // Source module name
	To         string             // Target module name
	EdgeType   DependencyEdgeType // Type of dependency
	ImportInfo *ImportInfo        // Details about the import
}

// DependencyEdgeType represents the type of dependency relationship
type DependencyEdgeType string

const (
	DependencyEdgeImport     DependencyEdgeType = "import"      // Direct import (import module)
	DependencyEdgeFromImport DependencyEdgeType = "from_import" // From import (from module import name)
	DependencyEdgeRelative   DependencyEdgeType = "relative"    // Relative import
	DependencyEdgeImplicit   DependencyEdgeType = "implicit"    // Implicit dependency
)

// ImportInfo contains details about an import statement
type ImportInfo struct {
	Statement      string   // Original import statement
	ImportedNames  []string // Names imported (for from imports)
	Alias          string   // Alias used (if any)
	IsRelative     bool     // True for relative imports
	Level          int      // Level for relative imports (number of dots)
	Line           int      // Line number where import occurs
	IsTypeChecking bool     // True if import is inside a TYPE_CHECKING block
}

// DependencyGraph represents the complete module dependency graph
type DependencyGraph struct {
	// Graph structure
	Nodes map[string]*ModuleNode // Module name -> ModuleNode
	Edges []*DependencyEdge      // All dependency relationships

	// Graph metadata
	TotalModules int      // Total number of modules
	TotalEdges   int      // Total number of dependencies
	RootModules  []string // Modules with no dependencies
	LeafModules  []string // Modules with no dependents
	ProjectRoot  string   // Project root directory

	// Analysis results
	CyclicGroups  [][]string                // Strongly connected components (cycles)
	ModuleMetrics map[string]*ModuleMetrics // Module-level metrics
	SystemMetrics *SystemMetrics            // System-wide metrics
}

// ModuleMetrics contains metrics for a single module
type ModuleMetrics struct {
	// Coupling metrics
	AfferentCoupling int     // Ca - Number of modules that depend on this module
	EfferentCoupling int     // Ce - Number of modules this module depends on
	Instability      float64 // I = Ce / (Ca + Ce) - Measure of instability
	Abstractness     float64 // A - Measure of abstractness (0-1)
	Distance         float64 // D - Distance from main sequence

	// Size metrics
	LinesOfCode     int // Total lines of code
	PublicInterface int // Number of public functions/classes

	// Quality metrics
	CyclomaticComplexity int // Average complexity of functions
}

// SystemMetrics contains system-wide quality metrics
type SystemMetrics struct {
	// Overall structure
	TotalModules      int // Total number of modules
	TotalDependencies int // Total number of dependencies
	PackageCount      int // Number of packages

	// Dependency metrics
	AverageFanIn    float64 // Average number of incoming dependencies
	AverageFanOut   float64 // Average number of outgoing dependencies
	DependencyRatio float64 // Total dependencies / Total modules

	// Coupling and cohesion
	AverageInstability    float64 // System average instability
	AverageAbstractness   float64 // System average abstractness
	MainSequenceDeviation float64 // Average distance from main sequence

	// Modularity
	ModularityIndex float64 // Measure of system decomposition quality
	ComponentRatio  float64 // Ratio of strongly connected components

	// Quality indicators
	CyclicDependencies int     // Number of modules in cycles
	MaxDependencyDepth int     // Maximum dependency chain length
	SystemComplexity   float64 // Overall system complexity score

	// Refactoring
	RefactoringPriority []string // Modules needing refactoring (highest priority first)
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph(projectRoot string) *DependencyGraph {
	return &DependencyGraph{
		Nodes:         make(map[string]*ModuleNode),
		Edges:         make([]*DependencyEdge, 0),
		ModuleMetrics: make(map[string]*ModuleMetrics),
		ProjectRoot:   projectRoot,
		SystemMetrics: &SystemMetrics{},
	}
}

// AddModule adds a module to the graph
func (g *DependencyGraph) AddModule(moduleName, filePath string) *ModuleNode {
	if node, exists := g.Nodes[moduleName]; exists {
		return node
	}

	// Calculate relative path
	relativePath, _ := filepath.Rel(g.ProjectRoot, filePath)

	// Determine package name
	packageName := g.extractPackageName(moduleName)

	// Check if this is a package (__init__.py)
	isPackage := strings.HasSuffix(filePath, "__init__.py")

	node := &ModuleNode{
		Name:         moduleName,
		FilePath:     filePath,
		RelativePath: relativePath,
		Package:      packageName,
		IsPackage:    isPackage,
		Dependencies: make(map[string]bool),
		Dependents:   make(map[string]bool),
		Imports:      make([]string, 0),
		ImportedBy:   make([]string, 0),
		PublicNames:  make([]string, 0),
	}

	g.Nodes[moduleName] = node
	g.TotalModules++
	return node
}

// AddDependency adds a dependency edge between two modules
func (g *DependencyGraph) AddDependency(from, to string, edgeType DependencyEdgeType, importInfo *ImportInfo) {
	// Ensure both nodes exist
	fromNode := g.Nodes[from]
	toNode := g.Nodes[to]

	if fromNode == nil || toNode == nil {
		return // Skip invalid dependencies
	}

	// Avoid self-dependencies
	if from == to {
		return
	}

	// Check if dependency already exists
	if fromNode.Dependencies[to] {
		return
	}

	// Add to graph structure
	edge := &DependencyEdge{
		From:       from,
		To:         to,
		EdgeType:   edgeType,
		ImportInfo: importInfo,
	}
	g.Edges = append(g.Edges, edge)
	g.TotalEdges++

	// Update node relationships
	fromNode.Dependencies[to] = true
	fromNode.Imports = append(fromNode.Imports, to)
	fromNode.OutDegree++

	toNode.Dependents[from] = true
	toNode.ImportedBy = append(toNode.ImportedBy, from)
	toNode.InDegree++
}

// GetModule retrieves a module node by name
func (g *DependencyGraph) GetModule(moduleName string) *ModuleNode {
	return g.Nodes[moduleName]
}

// GetDependencies returns all modules that the given module depends on
func (g *DependencyGraph) GetDependencies(moduleName string) []string {
	if node := g.Nodes[moduleName]; node != nil {
		return node.Imports
	}
	return nil
}

// GetDependents returns all modules that depend on the given module
func (g *DependencyGraph) GetDependents(moduleName string) []string {
	if node := g.Nodes[moduleName]; node != nil {
		return node.ImportedBy
	}
	return nil
}

// GetModuleNames returns all module names in the graph
func (g *DependencyGraph) GetModuleNames() []string {
	names := make([]string, 0, len(g.Nodes))
	for name := range g.Nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetPackages returns all unique package names
func (g *DependencyGraph) GetPackages() []string {
	packages := make(map[string]bool)
	for _, node := range g.Nodes {
		if node.Package != "" {
			packages[node.Package] = true
		}
	}

	result := make([]string, 0, len(packages))
	for pkg := range packages {
		result = append(result, pkg)
	}
	sort.Strings(result)
	return result
}

// HasCycle checks if the graph contains any cycles
func (g *DependencyGraph) HasCycle() bool {
	return len(g.CyclicGroups) > 0
}

// GetModulesInCycles returns all modules that are part of dependency cycles
func (g *DependencyGraph) GetModulesInCycles() []string {
	cyclic := make(map[string]bool)
	for _, group := range g.CyclicGroups {
		for _, module := range group {
			cyclic[module] = true
		}
	}

	result := make([]string, 0, len(cyclic))
	for module := range cyclic {
		result = append(result, module)
	}
	sort.Strings(result)
	return result
}

// GetRootModules returns modules with no dependencies (entry points)
func (g *DependencyGraph) GetRootModules() []string {
	if g.RootModules == nil {
		g.calculateRootAndLeafModules()
	}
	return g.RootModules
}

// GetLeafModules returns modules with no dependents (utilities)
func (g *DependencyGraph) GetLeafModules() []string {
	if g.LeafModules == nil {
		g.calculateRootAndLeafModules()
	}
	return g.LeafModules
}

// calculateRootAndLeafModules identifies root and leaf modules
func (g *DependencyGraph) calculateRootAndLeafModules() {
	var roots, leaves []string

	for name, node := range g.Nodes {
		if node.OutDegree == 0 {
			roots = append(roots, name)
		}
		if node.InDegree == 0 {
			leaves = append(leaves, name)
		}
	}

	sort.Strings(roots)
	sort.Strings(leaves)

	g.RootModules = roots
	g.LeafModules = leaves
}

// extractPackageName extracts the package name from a module name
func (g *DependencyGraph) extractPackageName(moduleName string) string {
	parts := strings.Split(moduleName, ".")
	if len(parts) > 1 {
		return parts[0]
	}
	return ""
}

// GetDependencyChain finds the dependency path between two modules
func (g *DependencyGraph) GetDependencyChain(from, to string) []string {
	// Simple BFS to find shortest path
	if from == to {
		return []string{from}
	}

	queue := [][]string{{from}}
	visited := make(map[string]bool)
	visited[from] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		current := path[len(path)-1]

		if node := g.Nodes[current]; node != nil {
			for dep := range node.Dependencies {
				if dep == to {
					return append(path, dep)
				}

				if !visited[dep] {
					visited[dep] = true
					newPath := make([]string, len(path)+1)
					copy(newPath, path)
					newPath[len(path)] = dep
					queue = append(queue, newPath)
				}
			}
		}
	}

	return nil // No path found
}

// String returns a string representation of the graph
func (g *DependencyGraph) String() string {
	return fmt.Sprintf("DependencyGraph{modules=%d, edges=%d, cycles=%d}",
		g.TotalModules, g.TotalEdges, len(g.CyclicGroups))
}

// Validate checks the graph for consistency
func (g *DependencyGraph) Validate() error {
	// Check that all edge endpoints exist as nodes
	for _, edge := range g.Edges {
		if g.Nodes[edge.From] == nil {
			return fmt.Errorf("edge references non-existent source module: %s", edge.From)
		}
		if g.Nodes[edge.To] == nil {
			return fmt.Errorf("edge references non-existent target module: %s", edge.To)
		}
	}

	// Check degree consistency
	for name, node := range g.Nodes {
		if len(node.Dependencies) != node.OutDegree {
			return fmt.Errorf("module %s: dependency count mismatch", name)
		}
		if len(node.Dependents) != node.InDegree {
			return fmt.Errorf("module %s: dependent count mismatch", name)
		}
	}

	return nil
}

// Clone creates a deep copy of the dependency graph
func (g *DependencyGraph) Clone() *DependencyGraph {
	clone := NewDependencyGraph(g.ProjectRoot)

	// Copy nodes
	for name, node := range g.Nodes {
		newNode := &ModuleNode{
			Name:          node.Name,
			FilePath:      node.FilePath,
			RelativePath:  node.RelativePath,
			Package:       node.Package,
			IsPackage:     node.IsPackage,
			Dependencies:  make(map[string]bool),
			Dependents:    make(map[string]bool),
			Imports:       make([]string, len(node.Imports)),
			ImportedBy:    make([]string, len(node.ImportedBy)),
			InDegree:      node.InDegree,
			OutDegree:     node.OutDegree,
			LineCount:     node.LineCount,
			FunctionCount: node.FunctionCount,
			ClassCount:    node.ClassCount,
			PublicNames:   make([]string, len(node.PublicNames)),
		}

		// Copy maps and slices
		for dep := range node.Dependencies {
			newNode.Dependencies[dep] = true
		}
		for dep := range node.Dependents {
			newNode.Dependents[dep] = true
		}
		copy(newNode.Imports, node.Imports)
		copy(newNode.ImportedBy, node.ImportedBy)
		copy(newNode.PublicNames, node.PublicNames)

		clone.Nodes[name] = newNode
	}

	// Copy edges
	for _, edge := range g.Edges {
		newImportInfo := &ImportInfo{}
		if edge.ImportInfo != nil {
			*newImportInfo = *edge.ImportInfo
		}

		newEdge := &DependencyEdge{
			From:       edge.From,
			To:         edge.To,
			EdgeType:   edge.EdgeType,
			ImportInfo: newImportInfo,
		}
		clone.Edges = append(clone.Edges, newEdge)
	}

	clone.TotalModules = g.TotalModules
	clone.TotalEdges = g.TotalEdges

	return clone
}
