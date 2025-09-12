package analyzer

import (
	"container/list"
	"fmt"
	"sort"
)

// KCoreGrouping ensures each fragment has at least k similar neighbors
type KCoreGrouping struct {
	threshold float64
	k         int
}

func NewKCoreGrouping(threshold float64, k int) *KCoreGrouping {
	if k < 2 {
		k = 2 // Minimum meaningful value
	}
	return &KCoreGrouping{threshold: threshold, k: k}
}

func (k *KCoreGrouping) GetName() string { return fmt.Sprintf("%d-Core", k.k) }

func (k *KCoreGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

	// Collect unique nodes and build adjacency with edges meeting threshold
	nodes := make([]*CodeFragment, 0)
	seen := make(map[*CodeFragment]struct{})
	adj := make(map[*CodeFragment]map[*CodeFragment]float64)
	simMap := make(map[string]float64)
	typeMap := make(map[string]CloneType)

	addNode := func(f *CodeFragment) {
		if _, ok := seen[f]; !ok {
			seen[f] = struct{}{}
			nodes = append(nodes, f)
			adj[f] = make(map[*CodeFragment]float64)
		}
	}

	for _, p := range pairs {
		if p == nil || p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		addNode(p.Fragment1)
		addNode(p.Fragment2)
		key := pairKey(p.Fragment1, p.Fragment2)
		if old, ok := simMap[key]; !ok || p.Similarity > old {
			simMap[key] = p.Similarity
			typeMap[key] = p.CloneType
		}
		if p.Similarity >= k.threshold {
			adj[p.Fragment1][p.Fragment2] = p.Similarity
			adj[p.Fragment2][p.Fragment1] = p.Similarity
		}
	}

	if len(nodes) == 0 {
		return []*CloneGroup{}
	}

	// Compute initial degrees
	degree := make(map[*CodeFragment]int, len(nodes))
	for n, nbrs := range adj {
		degree[n] = len(nbrs)
	}

	// Queue for nodes with degree < k
	q := list.New()
	inQueue := make(map[*CodeFragment]bool)
	for n, d := range degree {
		if d < k.k {
			q.PushBack(n)
			inQueue[n] = true
		}
	}

	// Iteratively remove low-degree nodes
	removed := make(map[*CodeFragment]bool)
	for q.Len() > 0 {
		e := q.Front()
		q.Remove(e)
		v := e.Value.(*CodeFragment)
		if removed[v] {
			continue
		}
		removed[v] = true
		// Decrease degree of neighbors
		for u := range adj[v] {
			if removed[u] {
				continue
			}
			degree[u]--
			delete(adj[u], v)
			if degree[u] < k.k && !inQueue[u] {
				q.PushBack(u)
				inQueue[u] = true
			}
		}
		// Clear v's adjacency
		delete(adj, v)
	}

	// Remaining nodes form the k-core subgraph (adj contains remaining nodes)
	// Now find connected components among remaining nodes
	groups := make([]*CloneGroup, 0)
	visited := make(map[*CodeFragment]bool)
	groupID := 0

	// Build deterministic order
	sort.Slice(nodes, func(i, j int) bool { return fragmentLess(nodes[i], nodes[j]) })

	for _, start := range nodes {
		if removed[start] || visited[start] || adj[start] == nil {
			continue
		}
		// BFS/DFS to collect component
		stack := []*CodeFragment{start}
		component := make([]*CodeFragment, 0)
		visited[start] = true
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			component = append(component, v)
			for u := range adj[v] {
				if !removed[u] && !visited[u] {
					visited[u] = true
					stack = append(stack, u)
				}
			}
		}
		if len(component) < 2 {
			continue
		}
		sort.Slice(component, func(i, j int) bool { return fragmentLess(component[i], component[j]) })
		g := NewCloneGroup(groupID)
		groupID++
		for _, f := range component {
			g.AddFragment(f)
		}
		g.Similarity = averageGroupSimilarity(simMap, component)
		g.CloneType = majorityCloneType(typeMap, component)
		groups = append(groups, g)
	}

	// Sort groups by similarity then size
	sort.Slice(groups, func(i, j int) bool {
		if !almostEqual(groups[i].Similarity, groups[j].Similarity) {
			return groups[i].Similarity > groups[j].Similarity
		}
		if groups[i].Size != groups[j].Size {
			return groups[i].Size > groups[j].Size
		}
		if len(groups[i].Fragments) == 0 || len(groups[j].Fragments) == 0 {
			return false
		}
		return fragmentLess(groups[i].Fragments[0], groups[j].Fragments[0])
	})

	return groups
}
