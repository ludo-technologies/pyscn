package analyzer

import (
	"context"
	"fmt"
	"sort"
)

type componentPath struct {
	component int
	next      *componentPath
	length    int
}

// FindLongestDependencyChains returns the longest paths through the
// SCC-condensed dependency graph. The result is bounded by limit and ordered
// by component depth, then lexicographically for deterministic ties.
func FindLongestDependencyChains(
	ctx context.Context,
	graph *DependencyGraph,
	limit int,
) ([]DependencyChain, error) {
	if graph == nil {
		return nil, fmt.Errorf("dependency graph is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if limit <= 0 || len(graph.Nodes) == 0 {
		return nil, nil
	}

	condensed, err := buildCondensedDependencyGraph(ctx, graph)
	if err != nil {
		return nil, fmt.Errorf("build condensed dependency graph: %w", err)
	}
	order, err := condensed.topologicalOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("order condensed dependency graph: %w", err)
	}

	bestByComponent := make([][]*componentPath, len(condensed.dependencies))
	for index := len(order) - 1; index >= 0; index-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		component := order[index]
		candidates := make([]*componentPath, 0)
		if len(condensed.dependencies[component]) == 0 {
			candidates = append(candidates, &componentPath{
				component: component,
				length:    1,
			})
		} else {
			for dependency := range condensed.dependencies[component] {
				if err := ctx.Err(); err != nil {
					return nil, err
				}
				for _, suffix := range bestByComponent[dependency] {
					candidates = append(candidates, &componentPath{
						component: component,
						next:      suffix,
						length:    suffix.length + 1,
					})
				}
			}
		}
		sortComponentPaths(candidates, condensed.componentNames)
		if len(candidates) > limit {
			candidates = candidates[:limit]
		}
		bestByComponent[component] = candidates
	}

	candidates := make([]*componentPath, 0, len(condensed.dependencies)*limit)
	for _, paths := range bestByComponent {
		for _, path := range paths {
			if path.length > 1 {
				candidates = append(candidates, path)
			}
		}
	}
	sortComponentPaths(candidates, condensed.componentNames)
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	chains := make([]DependencyChain, 0, len(candidates))
	for _, candidate := range candidates {
		path, err := expandComponentPath(ctx, graph, condensed, candidate)
		if err != nil {
			return nil, fmt.Errorf("expand dependency chain: %w", err)
		}
		chains = append(chains, DependencyChain{
			From:   path[0],
			To:     path[len(path)-1],
			Path:   path,
			Length: len(path),
		})
	}
	return chains, nil
}

func sortComponentPaths(paths []*componentPath, componentNames []string) {
	sort.Slice(paths, func(left, right int) bool {
		return componentPathLess(paths[left], paths[right], componentNames)
	})
}

func componentPathLess(left, right *componentPath, componentNames []string) bool {
	if left.length != right.length {
		return left.length > right.length
	}
	for left != nil && right != nil {
		leftName := componentNames[left.component]
		rightName := componentNames[right.component]
		if leftName != rightName {
			return leftName < rightName
		}
		left = left.next
		right = right.next
	}
	return false
}

func expandComponentPath(
	ctx context.Context,
	graph *DependencyGraph,
	condensed *condensedDependencyGraph,
	path *componentPath,
) ([]string, error) {
	firstEdge := condensed.dependencies[path.component][path.next.component]
	modules := []string{firstEdge.fromModule, firstEdge.toModule}
	currentModule := firstEdge.toModule
	path = path.next

	for path.next != nil {
		edge := condensed.dependencies[path.component][path.next.component]
		connector, err := findComponentPath(
			ctx,
			graph,
			condensed.componentByModule,
			path.component,
			currentModule,
			edge.fromModule,
		)
		if err != nil {
			return nil, err
		}
		modules = append(modules, connector[1:]...)
		modules = append(modules, edge.toModule)
		currentModule = edge.toModule
		path = path.next
	}

	return modules, nil
}

func findComponentPath(
	ctx context.Context,
	graph *DependencyGraph,
	componentByModule map[string]int,
	component int,
	fromModule string,
	toModule string,
) ([]string, error) {
	if fromModule == toModule {
		return []string{fromModule}, nil
	}

	queue := []string{fromModule}
	previous := map[string]string{fromModule: ""}
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		current := queue[0]
		queue = queue[1:]

		dependencies := make([]string, 0, len(graph.Nodes[current].Dependencies))
		for dependency := range graph.Nodes[current].Dependencies {
			if componentByModule[dependency] == component {
				dependencies = append(dependencies, dependency)
			}
		}
		sort.Strings(dependencies)

		for _, dependency := range dependencies {
			if _, visited := previous[dependency]; visited {
				continue
			}
			previous[dependency] = current
			if dependency == toModule {
				return reconstructDependencyPath(previous, fromModule, toModule), nil
			}
			queue = append(queue, dependency)
		}
	}

	return nil, fmt.Errorf(
		"component %d has no path from %q to %q",
		component,
		fromModule,
		toModule,
	)
}

func reconstructDependencyPath(
	previous map[string]string,
	fromModule string,
	toModule string,
) []string {
	path := []string{toModule}
	for current := toModule; current != fromModule; {
		current = previous[current]
		path = append(path, current)
	}
	for left, right := 0, len(path)-1; left < right; left, right = left+1, right-1 {
		path[left], path[right] = path[right], path[left]
	}
	return path
}
