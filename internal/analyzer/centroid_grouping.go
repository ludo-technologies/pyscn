package analyzer

import (
	"sort"
)

// CentroidGrouping implements centroid-based grouping that avoids transitive problems
// Note: This is the standard GroupingStrategy implementation that works with pre-computed pairs.
// For performance optimization, clone_detector.go has a direct implementation (detectClonesWithCentroid)
// that avoids pre-computing all pairs.
type CentroidGrouping struct {
	threshold float64
	analyzer  *APTEDAnalyzer
}

// NewCentroidGrouping creates a new centroid-based grouping strategy
func NewCentroidGrouping(threshold float64) *CentroidGrouping {
	costModel := NewPythonCostModel()
	return &CentroidGrouping{
		threshold: threshold,
		analyzer:  NewAPTEDAnalyzer(costModel),
	}
}

func (c *CentroidGrouping) GetName() string { return "Centroid-based" }

// GroupClones groups clones using centroid-based approach
func (c *CentroidGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

	// Collect unique fragments
	fragments := make([]*CodeFragment, 0)
	seen := make(map[*CodeFragment]struct{})
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
	}

	if len(fragments) == 0 {
		return []*CloneGroup{}
	}

	// BFS-based grouping without using pre-computed pairs
	groups := make([]*CloneGroup, 0)
	unclassified := make(map[*CodeFragment]bool)
	for _, f := range fragments {
		unclassified[f] = true
	}

	groupID := 0
	for len(unclassified) > 0 {
		// Pick first unclassified fragment as seed
		var seed *CodeFragment
		for f := range unclassified {
			seed = f
			break
		}
		delete(unclassified, seed)

		// Start new group
		group := NewCloneGroup(groupID)
		groupID++
		group.AddFragment(seed)

		// BFS queue
		queue := []*CodeFragment{seed}

		// Process queue
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			// Check all unclassified fragments
			toAdd := make([]*CodeFragment, 0)
			for candidate := range unclassified {
				// Calculate similarity between candidate and current group member
				similarity := c.calculateSimilarity(current, candidate)

				if similarity >= c.threshold {
					toAdd = append(toAdd, candidate)
				}
			}

			// Add qualifying fragments to group and queue
			for _, f := range toAdd {
				group.AddFragment(f)
				queue = append(queue, f)
				delete(unclassified, f)
			}
		}

		// Only keep groups with 2+ members
		if len(group.Fragments) >= 2 {
			// Calculate average similarity within group
			c.calculateGroupSimilarity(group)
			groups = append(groups, group)
		}
	}

	// Sort groups by similarity and size
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Similarity != groups[j].Similarity {
			return groups[i].Similarity > groups[j].Similarity
		}
		return groups[i].Size > groups[j].Size
	})

	return groups
}

// calculateSimilarity computes similarity between two fragments
func (c *CentroidGrouping) calculateSimilarity(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil || f1.TreeNode == nil || f2.TreeNode == nil {
		return 0.0
	}
	return c.analyzer.ComputeSimilarity(f1.TreeNode, f2.TreeNode)
}

// calculateGroupSimilarity calculates average similarity within a group
func (c *CentroidGrouping) calculateGroupSimilarity(group *CloneGroup) {
	if len(group.Fragments) < 2 {
		group.Similarity = 1.0
		return
	}

	totalSimilarity := 0.0
	pairCount := 0

	// Calculate pairwise similarities
	for i := 0; i < len(group.Fragments); i++ {
		for j := i + 1; j < len(group.Fragments); j++ {
			sim := c.calculateSimilarity(group.Fragments[i], group.Fragments[j])
			totalSimilarity += sim
			pairCount++
		}
	}

	if pairCount > 0 {
		group.Similarity = totalSimilarity / float64(pairCount)
	} else {
		group.Similarity = 0.0
	}

	// Determine clone type based on average similarity
	switch {
	case group.Similarity >= 0.95:
		group.CloneType = Type1Clone
	case group.Similarity >= 0.85:
		group.CloneType = Type2Clone
	case group.Similarity >= 0.75:
		group.CloneType = Type3Clone
	default:
		group.CloneType = Type4Clone
	}
}