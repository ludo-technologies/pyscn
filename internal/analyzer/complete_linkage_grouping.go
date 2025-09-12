package analyzer

import (
	"sort"
)

// CompleteLinkageGrouping ensures all pairs in a group meet the threshold
type CompleteLinkageGrouping struct {
	threshold float64
}

func NewCompleteLinkageGrouping(threshold float64) *CompleteLinkageGrouping {
	return &CompleteLinkageGrouping{threshold: threshold}
}

func (c *CompleteLinkageGrouping) GetName() string { return "Complete Linkage" }

func (c *CompleteLinkageGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

	// Collect unique fragments and build a similarity cache from pairs
	fragments := make([]*CodeFragment, 0)
	seen := make(map[*CodeFragment]struct{})
	sims := make(map[string]float64)
	types := make(map[string]CloneType)
	for _, p := range pairs {
		if p == nil || p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		if _, ok := seen[p.Fragment1]; !ok {
			seen[p.Fragment1] = struct{}{}
			fragments = append(fragments, p.Fragment1)
		}
		if _, ok := seen[p.Fragment2]; !ok {
			seen[p.Fragment2] = struct{}{}
			fragments = append(fragments, p.Fragment2)
		}
		k := pairKey(p.Fragment1, p.Fragment2)
		if old, ok := sims[k]; !ok || p.Similarity > old {
			sims[k] = p.Similarity
			types[k] = p.CloneType
		}
	}

	n := len(fragments)
	if n < 2 {
		return []*CloneGroup{}
	}

	// Initialize clusters: each fragment in its own cluster
	clusters := make([][]*CodeFragment, n)
	for i, f := range fragments {
		clusters[i] = []*CodeFragment{f}
	}

	// Helper to compute complete-linkage similarity between two clusters
	clusterSim := func(a, b []*CodeFragment) float64 {
		// minimum similarity between any pair across clusters
		// Early exit: if any pair falls below threshold, this cluster pair is invalid.
		minSim := 1.0
		hasPair := false
		for _, x := range a {
			for _, y := range b {
				s := similarity(sims, x, y) // 0 if missing
				if s < c.threshold {
					// Early rejection for complete linkage
					return 0.0
				}
				if s < minSim {
					minSim = s
				}
				hasPair = true
			}
		}
		if !hasPair {
			return 0.0
		}
		return minSim
	}

	// Repeatedly merge the best pair whose complete-linkage sim >= threshold
	for {
		bestI, bestJ := -1, -1
		bestScore := -1.0
		// Find best pair
		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				s := clusterSim(clusters[i], clusters[j])
				if s >= c.threshold {
					if s > bestScore {
						bestScore = s
						bestI, bestJ = i, j
					}
				}
			}
		}
		if bestI == -1 || bestJ == -1 {
			break // no more merges possible
		}
		// Merge bestJ into bestI
		merged := append(clusters[bestI], clusters[bestJ]...)
		clusters[bestI] = merged
		// Remove bestJ
		clusters = append(clusters[:bestJ], clusters[bestJ+1:]...)
	}

	// Build groups from clusters where all intra-pairs meet threshold
	groups := make([]*CloneGroup, 0)
	groupID := 0
	for _, cl := range clusters {
		if len(cl) < 2 {
			continue
		}
		// Verify complete linkage property within cluster
		ok := true
		for i := 0; i < len(cl) && ok; i++ {
			for j := i + 1; j < len(cl); j++ {
				if similarity(sims, cl[i], cl[j]) < c.threshold {
					ok = false
					break
				}
			}
		}
		if !ok {
			continue
		}

		sort.Slice(cl, func(i, j int) bool { return fragmentLess(cl[i], cl[j]) })
		g := NewCloneGroup(groupID)
		groupID++
		for _, f := range cl {
			g.AddFragment(f)
		}
		g.Similarity = averageGroupSimilarity(sims, cl)
		g.CloneType = majorityCloneType(types, cl)
		groups = append(groups, g)
	}

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
