package analyzer

import (
	"context"
	"fmt"

	coregraph "github.com/ludo-technologies/polyscan/core/graph"
)

// CalculateMaxDependencyDepth returns the maximum number of edges between
// strongly connected components in the dependency graph. Modules in a cycle
// form one component because circular dependencies are reported separately.
func CalculateMaxDependencyDepth(ctx context.Context, graph *DependencyGraph) (int, error) {
	if graph == nil {
		return 0, fmt.Errorf("dependency graph is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if len(graph.Nodes) == 0 {
		return 0, nil
	}

	cycles := coregraph.NewCycleDetector().DetectCycles(graph).Cycles
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	componentByModule, componentCount := dependencyComponents(graph, cycles)
	dependencies := make([]map[int]struct{}, componentCount)
	inDegree := make([]int, componentCount)
	for index := range dependencies {
		dependencies[index] = make(map[int]struct{})
	}

	for moduleName, node := range graph.Nodes {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		fromComponent := componentByModule[moduleName]
		for dependency := range node.Dependencies {
			toComponent := componentByModule[dependency]
			if fromComponent == toComponent {
				continue
			}
			if _, exists := dependencies[fromComponent][toComponent]; exists {
				continue
			}
			dependencies[fromComponent][toComponent] = struct{}{}
			inDegree[toComponent]++
		}
	}

	queue := make([]int, 0, componentCount)
	for component, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, component)
		}
	}

	depth := make([]int, componentCount)
	maxDepth := 0
	processed := 0
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return 0, err
		}

		component := queue[0]
		queue = queue[1:]
		processed++

		for dependency := range dependencies[component] {
			candidateDepth := depth[component] + 1
			if candidateDepth > depth[dependency] {
				depth[dependency] = candidateDepth
				if candidateDepth > maxDepth {
					maxDepth = candidateDepth
				}
			}
			inDegree[dependency]--
			if inDegree[dependency] == 0 {
				queue = append(queue, dependency)
			}
		}
	}

	if processed != componentCount {
		return 0, fmt.Errorf("condensed dependency graph contains a cycle")
	}
	return maxDepth, nil
}

func dependencyComponents(graph *DependencyGraph, cycles [][]string) (map[string]int, int) {
	componentByModule := make(map[string]int, len(graph.Nodes))
	componentCount := 0

	for _, cycle := range cycles {
		for _, moduleName := range cycle {
			componentByModule[moduleName] = componentCount
		}
		componentCount++
	}

	for moduleName := range graph.Nodes {
		if _, assigned := componentByModule[moduleName]; assigned {
			continue
		}
		componentByModule[moduleName] = componentCount
		componentCount++
	}

	return componentByModule, componentCount
}
