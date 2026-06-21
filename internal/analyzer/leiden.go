package analyzer

import "sort"

// LeidenOptions configures the Leiden community detection algorithm.
type LeidenOptions struct {
	// Resolution scales the null-model term in the modularity quality
	// function (default 1.0). Higher values favour smaller communities.
	Resolution float64

	// MinCommunitySize merges communities smaller than this threshold into
	// a neighbouring community after detection (default 1, no merging).
	MinCommunitySize int

	// MaxIterations bounds local-moving sweeps per phase (default 64).
	MaxIterations int

	// MaxPasses bounds Leiden passes before stopping (default 16).
	MaxPasses int
}

// DefaultLeidenOptions returns the default Leiden parameters.
func DefaultLeidenOptions() *LeidenOptions {
	return &LeidenOptions{
		Resolution:       1.0,
		MinCommunitySize: 1,
		MaxIterations:    64,
		MaxPasses:        16,
	}
}

// LeidenResult holds the output of Leiden community detection.
type LeidenResult struct {
	// Membership maps node index to community id in [0, NumCommunities).
	Membership []int

	// Modularity is the final modularity Q of the partition.
	Modularity float64

	// NumCommunities is the number of distinct communities.
	NumCommunities int
}

// DetectCommunitiesLeiden runs the Traag-Waltman-van Eck Leiden algorithm on
// a CommunityGraph. The implementation is pure Leiden (singleton start, no
// Louvain-style warm-up pass) with deterministic node iteration in sorted
// index order and lowest-community-id tie-breaking on equal modularity gain.
func DetectCommunitiesLeiden(cg *CommunityGraph, opts *LeidenOptions) *LeidenResult {
	if cg == nil || cg.NodeCount == 0 {
		return &LeidenResult{}
	}

	if opts == nil {
		opts = DefaultLeidenOptions()
	}
	if opts.Resolution <= 0 {
		opts.Resolution = 1.0
	}
	if opts.MinCommunitySize <= 0 {
		opts.MinCommunitySize = 1
	}
	if opts.MaxIterations <= 0 {
		opts.MaxIterations = 64
	}
	if opts.MaxPasses <= 0 {
		opts.MaxPasses = 16
	}

	original := newLeidenGraph(cg)
	if original.n == 0 {
		return &LeidenResult{}
	}

	g := original
	comm := make([]int, g.n)
	for i := range comm {
		comm[i] = i
	}

	project := func(graph *leidenGraph, partition []int) []int {
		membership := make([]int, cg.NodeCount)
		for i := 0; i < cg.NodeCount; i++ {
			membership[i] = partition[graph.lifted[i]]
		}
		return membership
	}

	bestMembership := project(g, comm)
	bestQ := original.modularity(bestMembership, opts.Resolution)

	prevQ := g.modularity(comm, opts.Resolution)
	for pass := 0; pass < opts.MaxPasses; pass++ {
		moved := g.localMove(comm, opts)
		refined := g.refine(comm, opts)
		nextG, nextComm := g.aggregate(comm, refined)
		nextQ := nextG.modularity(nextComm, opts.Resolution)

		projected := project(nextG, nextComm)
		projected = relabelDense(projected)
		if q := original.modularity(projected, opts.Resolution); q > bestQ {
			bestQ = q
			bestMembership = projected
		}

		if !moved && nextQ <= prevQ {
			break
		}
		g = nextG
		comm = nextComm
		prevQ = nextQ
	}

	membership := bestMembership
	membership, numCommunities := compactCommunityIDs(membership)
	if opts.MinCommunitySize > 1 {
		membership, numCommunities = mergeSmallCommunities(original, membership, opts.MinCommunitySize)
	}

	modularity := original.modularity(membership, opts.Resolution)
	return &LeidenResult{
		Membership:     membership,
		Modularity:     modularity,
		NumCommunities: numCommunities,
	}
}

// leidenGraph is a weighted undirected graph used by the Leiden inner loop.
type leidenGraph struct {
	n      int
	adj    [][]WeightedNeighbor
	deg    []float64
	loop   []float64
	m2     float64
	lifted []int
}

func newLeidenGraph(cg *CommunityGraph) *leidenGraph {
	n := cg.NodeCount
	adj := make([][]WeightedNeighbor, n)
	deg := make([]float64, n)
	var m2 float64

	for i := 0; i < n; i++ {
		neighbors := make([]WeightedNeighbor, len(cg.UndirectedAdj[i]))
		copy(neighbors, cg.UndirectedAdj[i])
		adj[i] = neighbors
		for _, nb := range neighbors {
			deg[i] += nb.Weight
		}
		m2 += deg[i]
	}

	lifted := make([]int, n)
	for i := range lifted {
		lifted[i] = i
	}

	return &leidenGraph{
		n:      n,
		adj:    adj,
		deg:    deg,
		loop:   make([]float64, n),
		m2:     m2,
		lifted: lifted,
	}
}

func (g *leidenGraph) modularity(comm []int, resolution float64) float64 {
	if g.m2 == 0 {
		return 0
	}

	cMax := 0
	for _, c := range comm {
		if c+1 > cMax {
			cMax = c + 1
		}
	}

	sigmaIn := make([]float64, cMax)
	sigmaTot := make([]float64, cMax)
	for v := 0; v < g.n; v++ {
		c := comm[v]
		sigmaTot[c] += g.deg[v]
		sigmaIn[c] += g.loop[v]
		for _, nb := range g.adj[v] {
			if comm[nb.Index] == c {
				sigmaIn[c] += nb.Weight
			}
		}
	}

	var q float64
	invM2 := 1.0 / g.m2
	for c := 0; c < cMax; c++ {
		q += sigmaIn[c]*invM2 - resolution*(sigmaTot[c]*invM2)*(sigmaTot[c]*invM2)
	}
	return q
}

func (g *leidenGraph) localMove(comm []int, opts *LeidenOptions) bool {
	if g.m2 == 0 {
		return false
	}

	cMax := maxCommunityID(comm) + 1
	sigmaTot := make([]float64, cMax)
	for v := 0; v < g.n; v++ {
		sigmaTot[comm[v]] += g.deg[v]
	}

	kvcArr := make([]float64, cMax)
	touched := make([]int, 0, 32)
	anyMoved := false

	for iter := 0; iter < opts.MaxIterations; iter++ {
		moved := false
		for v := 0; v < g.n; v++ {
			cv := comm[v]
			touched = touched[:0]

			for _, nb := range g.adj[v] {
				if nb.Index == v {
					continue
				}
				cu := comm[nb.Index]
				if kvcArr[cu] == 0 {
					touched = append(touched, cu)
				}
				kvcArr[cu] += nb.Weight
			}

			sigmaTot[cv] -= g.deg[v]
			bestC := cv
			bestDelta := 0.0
			invM2 := 1.0 / g.m2
			twoInvM2 := 2 * invM2
			vFactor := opts.Resolution * 2 * g.deg[v] * invM2 * invM2

			for _, c := range touched {
				delta := twoInvM2*kvcArr[c] - vFactor*sigmaTot[c]
				if delta > bestDelta || (delta == bestDelta && c < bestC) {
					bestDelta = delta
					bestC = c
				}
			}
			for _, c := range touched {
				kvcArr[c] = 0
			}

			sigmaTot[bestC] += g.deg[v]
			if bestC != cv {
				comm[v] = bestC
				moved = true
				anyMoved = true
			}
		}
		if !moved {
			break
		}
	}
	return anyMoved
}

func (g *leidenGraph) refine(parent []int, opts *LeidenOptions) []int {
	if g.m2 == 0 {
		out := make([]int, len(parent))
		copy(out, parent)
		return out
	}

	refined := make([]int, g.n)
	for i := range refined {
		refined[i] = i
	}

	sigmaTot := make([]float64, g.n)
	for v := 0; v < g.n; v++ {
		sigmaTot[v] += g.deg[v]
	}

	kvcArr := make([]float64, g.n)
	touched := make([]int, 0, 32)

	for iter := 0; iter < opts.MaxIterations; iter++ {
		moved := false
		for v := 0; v < g.n; v++ {
			parentV := parent[v]
			cv := refined[v]
			touched = touched[:0]

			for _, nb := range g.adj[v] {
				if nb.Index == v {
					continue
				}
				if parent[nb.Index] != parentV {
					continue
				}
				ru := refined[nb.Index]
				if kvcArr[ru] == 0 {
					touched = append(touched, ru)
				}
				kvcArr[ru] += nb.Weight
			}

			sigmaTot[cv] -= g.deg[v]
			bestC := cv
			bestDelta := 0.0
			invM2 := 1.0 / g.m2
			twoInvM2 := 2 * invM2
			vFactor := opts.Resolution * 2 * g.deg[v] * invM2 * invM2

			for _, c := range touched {
				delta := twoInvM2*kvcArr[c] - vFactor*sigmaTot[c]
				if delta > bestDelta || (delta == bestDelta && c < bestC) {
					bestDelta = delta
					bestC = c
				}
			}
			for _, c := range touched {
				kvcArr[c] = 0
			}

			sigmaTot[bestC] += g.deg[v]
			if bestC != cv {
				refined[v] = bestC
				moved = true
			}
		}
		if !moved {
			break
		}
	}
	return refined
}

func (g *leidenGraph) aggregate(parent, refined []int) (*leidenGraph, []int) {
	refinedRemap := make([]int, g.n)
	nNew := 0
	seen := make(map[int]int, g.n)
	for v := 0; v < g.n; v++ {
		r := refined[v]
		if id, ok := seen[r]; ok {
			refinedRemap[v] = id
			continue
		}
		seen[r] = nNew
		refinedRemap[v] = nNew
		nNew++
	}

	parentRemap := make(map[int]int)
	newComm := make([]int, nNew)
	for v := 0; v < g.n; v++ {
		super := refinedRemap[v]
		p := parent[v]
		if _, ok := parentRemap[p]; !ok {
			parentRemap[p] = len(parentRemap)
		}
		newComm[super] = parentRemap[p]
	}

	type edgeKey struct {
		from int
		to   int
	}
	edgeWeights := make(map[edgeKey]float64)
	loop := make([]float64, nNew)

	for v := 0; v < g.n; v++ {
		sv := refinedRemap[v]
		for _, nb := range g.adj[v] {
			u := nb.Index
			if v >= u {
				continue
			}
			su := refinedRemap[u]
			if sv == su {
				loop[sv] += nb.Weight
				continue
			}
			from, to := sv, su
			if from > to {
				from, to = to, from
			}
			edgeWeights[edgeKey{from, to}] += nb.Weight
		}
	}

	adj := make([][]WeightedNeighbor, nNew)
	deg := make([]float64, nNew)
	for key, weight := range edgeWeights {
		adj[key.from] = append(adj[key.from], WeightedNeighbor{Index: key.to, Weight: weight})
		adj[key.to] = append(adj[key.to], WeightedNeighbor{Index: key.from, Weight: weight})
		deg[key.from] += weight
		deg[key.to] += weight
	}

	for i := range adj {
		sort.Slice(adj[i], func(a, b int) bool {
			return adj[i][a].Index < adj[i][b].Index
		})
	}

	var m2 float64
	for i := 0; i < nNew; i++ {
		m2 += deg[i] + 2*loop[i]
	}

	nextLifted := make([]int, len(g.lifted))
	for v := 0; v < len(g.lifted); v++ {
		nextLifted[v] = refinedRemap[g.lifted[v]]
	}

	return &leidenGraph{
		n:      nNew,
		adj:    adj,
		deg:    deg,
		loop:   loop,
		m2:     m2,
		lifted: nextLifted,
	}, newComm
}

func maxCommunityID(comm []int) int {
	maxID := 0
	for _, c := range comm {
		if c > maxID {
			maxID = c
		}
	}
	return maxID
}

func relabelDense(membership []int) []int {
	out, _ := compactCommunityIDs(append([]int(nil), membership...))
	return out
}

func compactCommunityIDs(membership []int) ([]int, int) {
	remap := make(map[int]int)
	next := 0
	for _, c := range membership {
		if _, ok := remap[c]; !ok {
			remap[c] = next
			next++
		}
	}
	out := make([]int, len(membership))
	for i, c := range membership {
		out[i] = remap[c]
	}
	return out, next
}

func mergeSmallCommunities(g *leidenGraph, membership []int, minSize int) ([]int, int) {
	if minSize <= 1 || g.m2 == 0 {
		return membership, countCommunities(membership)
	}

	out := append([]int(nil), membership...)
	changed := true
	for changed {
		changed = false
		sizes := communitySizes(out)
		for commID, size := range sizes {
			if size >= minSize {
				continue
			}
			target := bestMergeTarget(g, out, commID)
			if target < 0 {
				continue
			}
			for i, c := range out {
				if c == commID {
					out[i] = target
				}
			}
			changed = true
		}
	}
	return compactCommunityIDs(out)
}

func communitySizes(membership []int) map[int]int {
	sizes := make(map[int]int)
	for _, c := range membership {
		sizes[c]++
	}
	return sizes
}

func countCommunities(membership []int) int {
	seen := make(map[int]struct{})
	for _, c := range membership {
		seen[c] = struct{}{}
	}
	return len(seen)
}

func bestMergeTarget(g *leidenGraph, membership []int, commID int) int {
	weights := make(map[int]float64)
	for v := 0; v < g.n; v++ {
		if membership[v] != commID {
			continue
		}
		for _, nb := range g.adj[v] {
			other := membership[nb.Index]
			if other == commID {
				continue
			}
			weights[other] += nb.Weight
		}
	}
	if len(weights) == 0 {
		return -1
	}

	best := -1
	bestWeight := -1.0
	for target, weight := range weights {
		if weight > bestWeight || (weight == bestWeight && (best < 0 || target < best)) {
			bestWeight = weight
			best = target
		}
	}
	return best
}
