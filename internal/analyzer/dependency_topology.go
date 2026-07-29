package analyzer

import (
	"context"
	"fmt"
)

type componentDependency struct {
	fromModule string
	toModule   string
}

type condensedDependencyGraph struct {
	componentByModule map[string]int
	componentNames    []string
	dependencies      []map[int]componentDependency
	inDegree          []int
}

// DependencyTopology is the canonical structural analysis of one dependency
// graph. MaxDepth and LongestChains share the same SCC condensation. A result
// belongs to the exact graph instance passed to AnalyzeDependencyTopology.
// Structural mutations through AddModule or AddDependency invalidate it;
// directly mutating DependencyGraph storage is unsupported.
type DependencyTopology struct {
	maxDepth      int
	longestChains []DependencyChain
	graph         *DependencyGraph
	totalModules  int
	totalEdges    int
}

// MaxDepth returns the maximum number of edges between SCCs.
func (topology *DependencyTopology) MaxDepth() int {
	return topology.maxDepth
}

// LongestChains returns a defensive copy of the ranked dependency chains.
func (topology *DependencyTopology) LongestChains() []DependencyChain {
	chains := make([]DependencyChain, len(topology.longestChains))
	copy(chains, topology.longestChains)
	for index := range chains {
		chains[index].Path = append([]string(nil), chains[index].Path...)
	}
	return chains
}

// AnalyzeDependencyTopology condenses a dependency graph and calculates its
// maximum depth and globally ranked chains from one shared SCC condensation.
func AnalyzeDependencyTopology(
	ctx context.Context,
	graph *DependencyGraph,
	chainLimit int,
) (*DependencyTopology, error) {
	if graph == nil {
		return nil, fmt.Errorf("dependency graph is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("analyze dependency topology: %w", err)
	}

	condensed, err := buildCondensedDependencyGraph(ctx, graph)
	if err != nil {
		return nil, fmt.Errorf("build condensed dependency graph: %w", err)
	}
	order, err := condensed.topologicalOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("order condensed dependency graph: %w", err)
	}
	maxDepth, err := calculateMaxDependencyDepth(ctx, condensed, order)
	if err != nil {
		return nil, fmt.Errorf("calculate dependency depth: %w", err)
	}
	longestChains, err := findLongestDependencyChains(
		ctx,
		graph,
		condensed,
		order,
		chainLimit,
	)
	if err != nil {
		return nil, fmt.Errorf("find dependency chains: %w", err)
	}

	return &DependencyTopology{
		maxDepth:      maxDepth,
		longestChains: longestChains,
		graph:         graph,
		totalModules:  graph.TotalModules,
		totalEdges:    graph.TotalEdges,
	}, nil
}

func calculateMaxDependencyDepth(
	ctx context.Context,
	condensed *condensedDependencyGraph,
	order []int,
) (int, error) {
	depth := make([]int, len(condensed.dependencies))
	maxDepth := 0
	for _, component := range order {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		for dependency := range condensed.dependencies[component] {
			candidateDepth := depth[component] + 1
			if candidateDepth > depth[dependency] {
				depth[dependency] = candidateDepth
				if candidateDepth > maxDepth {
					maxDepth = candidateDepth
				}
			}
		}
	}

	return maxDepth, nil
}

func buildCondensedDependencyGraph(
	ctx context.Context,
	graph *DependencyGraph,
) (*condensedDependencyGraph, error) {
	componentByModule, componentCount, err := dependencyComponents(ctx, graph)
	if err != nil {
		return nil, fmt.Errorf("calculate dependency components: %w", err)
	}

	condensed := &condensedDependencyGraph{
		componentByModule: componentByModule,
		componentNames:    make([]string, componentCount),
		dependencies:      make([]map[int]componentDependency, componentCount),
		inDegree:          make([]int, componentCount),
	}
	for component := range condensed.dependencies {
		condensed.dependencies[component] = make(map[int]componentDependency)
	}
	for moduleName, component := range componentByModule {
		if condensed.componentNames[component] == "" ||
			moduleName < condensed.componentNames[component] {
			condensed.componentNames[component] = moduleName
		}
	}

	for moduleName, node := range graph.Nodes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		fromComponent := componentByModule[moduleName]
		for dependency := range node.Dependencies {
			toComponent := componentByModule[dependency]
			if fromComponent == toComponent {
				continue
			}

			candidate := componentDependency{
				fromModule: moduleName,
				toModule:   dependency,
			}
			existing, exists := condensed.dependencies[fromComponent][toComponent]
			if exists {
				if dependencyEdgeLess(candidate, existing) {
					condensed.dependencies[fromComponent][toComponent] = candidate
				}
				continue
			}
			condensed.dependencies[fromComponent][toComponent] = candidate
			condensed.inDegree[toComponent]++
		}
	}

	return condensed, nil
}

func (graph *condensedDependencyGraph) topologicalOrder(ctx context.Context) ([]int, error) {
	inDegree := append([]int(nil), graph.inDegree...)
	queue := make([]int, 0, len(inDegree))
	for component, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, component)
		}
	}

	order := make([]int, 0, len(inDegree))
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		component := queue[0]
		queue = queue[1:]
		order = append(order, component)

		for dependency := range graph.dependencies[component] {
			inDegree[dependency]--
			if inDegree[dependency] == 0 {
				queue = append(queue, dependency)
			}
		}
	}

	if len(order) != len(inDegree) {
		return nil, fmt.Errorf("condensed dependency graph contains a cycle")
	}
	return order, nil
}

func dependencyEdgeLess(left, right componentDependency) bool {
	if left.fromModule != right.fromModule {
		return left.fromModule < right.fromModule
	}
	return left.toModule < right.toModule
}

// dependencyComponents owns topology SCC construction because the core cycle
// detector reports only non-singleton cycles and cannot observe cancellation.
// Topology requires a complete, cancellable partition including singletons.
func dependencyComponents(
	ctx context.Context,
	graph *DependencyGraph,
) (map[string]int, int, error) {
	componentByModule := make(map[string]int, len(graph.Nodes))
	indexByModule := make(map[string]int, len(graph.Nodes))
	lowLinkByModule := make(map[string]int, len(graph.Nodes))
	onStack := make(map[string]bool, len(graph.Nodes))
	stack := make([]string, 0, len(graph.Nodes))
	nextIndex := 0
	componentCount := 0

	var visit func(string) error
	visit = func(moduleName string) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		node := graph.Nodes[moduleName]
		if node == nil {
			return fmt.Errorf("module %q has no graph node", moduleName)
		}

		indexByModule[moduleName] = nextIndex
		lowLinkByModule[moduleName] = nextIndex
		nextIndex++
		stack = append(stack, moduleName)
		onStack[moduleName] = true

		for dependency := range node.Dependencies {
			if err := ctx.Err(); err != nil {
				return err
			}
			if _, exists := graph.Nodes[dependency]; !exists {
				return fmt.Errorf(
					"module %q depends on unknown module %q",
					moduleName,
					dependency,
				)
			}

			dependencyIndex, visited := indexByModule[dependency]
			if !visited {
				if err := visit(dependency); err != nil {
					return err
				}
				if lowLinkByModule[dependency] < lowLinkByModule[moduleName] {
					lowLinkByModule[moduleName] = lowLinkByModule[dependency]
				}
			} else if onStack[dependency] && dependencyIndex < lowLinkByModule[moduleName] {
				lowLinkByModule[moduleName] = dependencyIndex
			}
		}

		if lowLinkByModule[moduleName] != indexByModule[moduleName] {
			return nil
		}

		for {
			stackIndex := len(stack) - 1
			member := stack[stackIndex]
			stack = stack[:stackIndex]
			onStack[member] = false
			componentByModule[member] = componentCount
			if member == moduleName {
				break
			}
		}
		componentCount++
		return nil
	}

	for moduleName := range graph.Nodes {
		if _, visited := indexByModule[moduleName]; visited {
			continue
		}
		if err := visit(moduleName); err != nil {
			return nil, 0, err
		}
	}

	return componentByModule, componentCount, nil
}
