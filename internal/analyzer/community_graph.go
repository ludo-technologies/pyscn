package analyzer

import "sort"

// CommunityGraphBuildOptions configures construction of a CommunityGraph.
type CommunityGraphBuildOptions struct {
	// ExcludeLazyEdges omits lazy (function-body) import edges when true.
	// Default false (lazy edges included). Exclusion uses the same criterion
	// as circular dependency detection (issue #460).
	ExcludeLazyEdges bool

	// EdgeWeightFunc assigns a weight to each dependency edge. When nil, every
	// edge has weight 1.0. Reserved for future weighting by import type,
	// frequency, or laziness.
	EdgeWeightFunc func(edge *DependencyEdge) float64
}

// DefaultCommunityGraphBuildOptions returns the default build options.
func DefaultCommunityGraphBuildOptions() *CommunityGraphBuildOptions {
	return &CommunityGraphBuildOptions{}
}

// CommunityGraph is a compact, integer-indexed graph derived from a
// DependencyGraph for community detection (e.g. Leiden) and cross-community
// metrics.
type CommunityGraph struct {
	NodeCount   int
	NodeNames   []string
	NameToIndex map[string]int

	// UndirectedAdj holds weighted undirected adjacency lists for clustering.
	// Each neighbor list is sorted by index for deterministic iteration.
	UndirectedAdj [][]WeightedNeighbor

	// DirectedEdges preserves directed import edges for cross-community
	// in/out counts. Not collapsed into undirected form.
	DirectedEdges []CommunityDirectedEdge

	// TotalUndirectedWeight is the sum of undirected edge weights (each
	// undirected pair counted once). Used as a modularity denominator hook.
	TotalUndirectedWeight float64
}

// WeightedNeighbor is a neighbor in the undirected adjacency list.
type WeightedNeighbor struct {
	Index  int
	Weight float64
}

// CommunityDirectedEdge is a directed dependency edge between indexed nodes.
type CommunityDirectedEdge struct {
	FromIndex int
	ToIndex   int
	Weight    float64
	IsLazy    bool
}

// BuildCommunityGraph constructs a CommunityGraph from a DependencyGraph.
// Returns an empty graph when graph is nil or has no modules.
func BuildCommunityGraph(graph *DependencyGraph, opts *CommunityGraphBuildOptions) *CommunityGraph {
	if graph == nil || len(graph.Nodes) == 0 {
		return &CommunityGraph{
			NameToIndex: make(map[string]int),
		}
	}

	if opts == nil {
		opts = DefaultCommunityGraphBuildOptions()
	}

	nodeNames := graph.GetModuleNames()
	nodeCount := len(nodeNames)
	nameToIndex := make(map[string]int, nodeCount)
	for i, name := range nodeNames {
		nameToIndex[name] = i
	}

	weightFunc := opts.EdgeWeightFunc
	if weightFunc == nil {
		weightFunc = func(*DependencyEdge) float64 { return 1.0 }
	}

	directedEdges := make([]CommunityDirectedEdge, 0, len(graph.Edges))
	undirectedWeights := make(map[undirectedPairKey]float64)

	for _, edge := range graph.Edges {
		if edge == nil {
			continue
		}
		if edge.From == edge.To {
			continue
		}

		fromIdx, fromOK := nameToIndex[edge.From]
		toIdx, toOK := nameToIndex[edge.To]
		if !fromOK || !toOK {
			continue
		}

		if opts.ExcludeLazyEdges && edge.IsLazy {
			continue
		}

		weight := weightFunc(edge)
		if weight <= 0 {
			continue
		}

		directedEdges = append(directedEdges, CommunityDirectedEdge{
			FromIndex: fromIdx,
			ToIndex:   toIdx,
			Weight:    weight,
			IsLazy:    edge.IsLazy,
		})

		key := undirectedPairKey{min(fromIdx, toIdx), max(fromIdx, toIdx)}
		undirectedWeights[key] += weight
	}

	undirectedAdj := make([][]WeightedNeighbor, nodeCount)
	var totalUndirectedWeight float64
	for key, weight := range undirectedWeights {
		undirectedAdj[key.low] = append(undirectedAdj[key.low], WeightedNeighbor{Index: key.high, Weight: weight})
		undirectedAdj[key.high] = append(undirectedAdj[key.high], WeightedNeighbor{Index: key.low, Weight: weight})
		totalUndirectedWeight += weight
	}

	for i := range undirectedAdj {
		sort.Slice(undirectedAdj[i], func(a, b int) bool {
			return undirectedAdj[i][a].Index < undirectedAdj[i][b].Index
		})
	}

	return &CommunityGraph{
		NodeCount:             nodeCount,
		NodeNames:             nodeNames,
		NameToIndex:           nameToIndex,
		UndirectedAdj:         undirectedAdj,
		DirectedEdges:         directedEdges,
		TotalUndirectedWeight: totalUndirectedWeight,
	}
}

type undirectedPairKey struct {
	low  int
	high int
}
