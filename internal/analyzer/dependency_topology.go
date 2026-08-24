package analyzer

import (
	"context"
	"fmt"

	coregraph "github.com/ludo-technologies/polyscan/core/graph"
)

// DependencyTopology is the canonical load-time structural analysis of one
// dependency graph. MaxDepth and LongestChains share the same SCC condensation.
// A result belongs to the exact graph instance passed to
// AnalyzeDependencyTopology. Structural mutations through AddModule or
// AddDependency invalidate it; directly mutating DependencyGraph storage is
// unsupported.
type DependencyTopology struct {
	maxDepth      int
	longestChains []DependencyChain
	graph         *DependencyGraph
	graphRevision uint64
}

// MaxDepth returns the number of edges along the longest dependency chain,
// which is the same chain LongestChains ranks first.
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

// AnalyzeDependencyTopology condenses the graph's load-time dependencies with
// core/graph's chain finder and reads maximum depth and the chainLimit globally
// ranked chains off that one condensation. Lazy-only imports remain available
// to other graph analyses. Cancelling ctx aborts the search.
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

	finder, err := coregraph.NewChainFinder(ctx, loadTimeDependencyGraph{graph})
	if err != nil {
		return nil, fmt.Errorf("analyze dependency topology: %w", err)
	}
	paths, err := finder.LongestChains(ctx, chainLimit)
	if err != nil {
		return nil, fmt.Errorf("find dependency chains: %w", err)
	}

	// A chain of n modules is n-1 edges deep; an empty graph has no chain.
	maxDepth := max(len(finder.LongestChain())-1, 0)

	longestChains := make([]DependencyChain, 0, len(paths))
	for _, path := range paths {
		longestChains = append(longestChains, DependencyChain{
			From:   path[0],
			To:     path[len(path)-1],
			Path:   path,
			Length: len(path),
		})
	}

	return &DependencyTopology{
		maxDepth:      maxDepth,
		longestChains: longestChains,
		graph:         graph,
		graphRevision: graph.topologyRevision,
	}, nil
}
