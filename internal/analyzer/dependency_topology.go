package analyzer

import (
	"context"
	"fmt"
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

	componentByModule, componentCount, err := dependencyComponents(ctx, graph)
	if err != nil {
		return 0, fmt.Errorf("calculate dependency components: %w", err)
	}

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
