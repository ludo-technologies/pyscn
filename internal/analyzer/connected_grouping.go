package analyzer

import (
	"sort"
)

// ConnectedGrouping wraps the existing transitive grouping logic using Union-Find
type ConnectedGrouping struct {
	threshold float64
}

func NewConnectedGrouping(threshold float64) *ConnectedGrouping {
	return &ConnectedGrouping{threshold: threshold}
}

func (c *ConnectedGrouping) GetName() string { return "Connected Components" }

func (c *ConnectedGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

	// Build set of fragments and adjacency filtered by threshold
	fragments := make([]*CodeFragment, 0)
	seen := make(map[*CodeFragment]struct{})
	// Similarity map for later group similarity calculation
	simMap := make(map[string]float64)
	typeMap := make(map[string]CloneType)

	for _, p := range pairs {
		if p == nil || p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		// Track fragments
		if _, ok := seen[p.Fragment1]; !ok {
			seen[p.Fragment1] = struct{}{}
			fragments = append(fragments, p.Fragment1)
		}
		if _, ok := seen[p.Fragment2]; !ok {
			seen[p.Fragment2] = struct{}{}
			fragments = append(fragments, p.Fragment2)
		}

		// Cache similarity and type for existing pair
		k := pairKey(p.Fragment1, p.Fragment2)
		if old, ok := simMap[k]; !ok || p.Similarity > old {
			simMap[k] = p.Similarity
			typeMap[k] = p.CloneType
		}
	}

	if len(fragments) == 0 {
		return []*CloneGroup{}
	}

	// Union-Find across edges with similarity >= threshold
	parent := make(map[*CodeFragment]*CodeFragment, len(fragments))
	rank := make(map[*CodeFragment]int, len(fragments))

	var find func(*CodeFragment) *CodeFragment
	find = func(x *CodeFragment) *CodeFragment {
		if parent[x] != x {
			parent[x] = find(parent[x])
		}
		return parent[x]
	}
	union := func(a, b *CodeFragment) {
		ra := find(a)
		rb := find(b)
		if ra == rb {
			return
		}
		if rank[ra] < rank[rb] {
			parent[ra] = rb
		} else if rank[ra] > rank[rb] {
			parent[rb] = ra
		} else {
			parent[rb] = ra
			rank[ra]++
		}
	}
	for _, f := range fragments {
		parent[f] = f
		rank[f] = 0
	}

	// Union only for edges meeting threshold
	for _, p := range pairs {
		if p == nil || p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		if p.Similarity >= c.threshold {
			union(p.Fragment1, p.Fragment2)
		}
	}

	// Build components
	comp := make(map[*CodeFragment][]*CodeFragment)
	for _, f := range fragments {
		r := find(f)
		comp[r] = append(comp[r], f)
	}

	// Convert to groups, exclude singletons
	groups := make([]*CloneGroup, 0, len(comp))
	groupID := 0
	for _, members := range comp {
		if len(members) < 2 {
			continue
		}
		sort.Slice(members, func(i, j int) bool { return fragmentLess(members[i], members[j]) })
		g := NewCloneGroup(groupID)
		groupID++
		for _, f := range members {
			g.AddFragment(f)
		}
		// Compute average similarity using cached pairs among members
		g.Similarity = averageGroupSimilarity(simMap, members)
		// Determine predominant clone type from within-group available pairs
		g.CloneType = majorityCloneType(typeMap, members)
		groups = append(groups, g)
	}

	// Sort groups by decreasing similarity then size
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

// majorityCloneType chooses the most frequent CloneType among all pair edges in members.
func majorityCloneType(typeMap map[string]CloneType, members []*CodeFragment) CloneType {
	counts := make(map[CloneType]int)
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			key := pairKey(members[i], members[j])
			if t, ok := typeMap[key]; ok {
				counts[t]++
			}
		}
	}
	var best CloneType
	maxC := -1
	for t, c := range counts {
		if c > maxC {
			maxC = c
			best = t
		}
	}
	if maxC < 0 {
		return Type3Clone // fallback reasonable default
	}
	return best
}
