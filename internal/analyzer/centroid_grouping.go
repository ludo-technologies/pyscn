package analyzer

import (
	"fmt"
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
)

// CentroidGrouping implements centroid-based grouping that avoids transitive problems
// This strategy uses BFS to grow groups while directly comparing candidates to existing members,
// avoiding the transitive similarity issue (A↔B↔C where A and C are dissimilar).
type CentroidGrouping struct {
	threshold      float64
	analyzer       *APTEDAnalyzer
	type1Threshold float64
	type2Threshold float64
	type3Threshold float64
	type4Threshold float64
}

// NewCentroidGrouping creates a new centroid-based grouping strategy
func NewCentroidGrouping(threshold float64) *CentroidGrouping {
	costModel := NewPythonCostModel()
	return &CentroidGrouping{
		threshold:      threshold,
		analyzer:       NewAPTEDAnalyzer(costModel),
		type1Threshold: domain.DefaultType1CloneThreshold,
		type2Threshold: domain.DefaultType2CloneThreshold,
		type3Threshold: domain.DefaultType3CloneThreshold,
		type4Threshold: domain.DefaultType4CloneThreshold,
	}
}

func (c *CentroidGrouping) GetName() string { return "Centroid-based" }

// SetThresholds sets the clone type thresholds for classification
func (c *CentroidGrouping) SetThresholds(type1, type2, type3, type4 float64) {
	c.type1Threshold = type1
	c.type2Threshold = type2
	c.type3Threshold = type3
	c.type4Threshold = type4
}

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

	// Create similarity index from pre-computed pairs for faster lookup
	similarityIndex := make(map[string]float64)
	for _, p := range pairs {
		if p == nil || p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		// Store both directions for quick lookup using string keys
		key1 := c.makePairKey(p.Fragment1, p.Fragment2)
		key2 := c.makePairKey(p.Fragment2, p.Fragment1)
		similarityIndex[key1] = p.Similarity
		similarityIndex[key2] = p.Similarity
	}

	// BFS-based grouping using similarity index
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

			// Performance optimization: limit group size for large groups
			const maxGroupSize = 50
			if len(group.Fragments) >= maxGroupSize {
				break
			}

			// Check all unclassified fragments
			toAdd := make([]*CodeFragment, 0)
			for candidate := range unclassified {
				// First try to use pre-computed similarity
				var similarity float64
				key := c.makePairKey(current, candidate)
				if sim, exists := similarityIndex[key]; exists {
					similarity = sim
				} else {
					// Fall back to calculating if not in index
					similarity = c.calculateSimilarity(current, candidate)
				}

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

// fragmentID returns a stable identifier for a fragment based on its location
func (c *CentroidGrouping) fragmentID(f *CodeFragment) string {
	if f == nil || f.Location == nil {
		return fmt.Sprintf("%p", f)
	}
	loc := f.Location
	return fmt.Sprintf("%s|%d|%d|%d|%d", loc.FilePath, loc.StartLine, loc.EndLine, loc.StartCol, loc.EndCol)
}

// makePairKey creates a string key for a pair of fragments to avoid memory leaks
func (c *CentroidGrouping) makePairKey(f1, f2 *CodeFragment) string {
	id1 := c.fragmentID(f1)
	id2 := c.fragmentID(f2)
	if id1 <= id2 {
		return id1 + "|" + id2
	}
	return id2 + "|" + id1
}

// calculateSimilarity computes similarity between two fragments
func (c *CentroidGrouping) calculateSimilarity(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil || f1.TreeNode == nil || f2.TreeNode == nil {
		return 0.0
	}
	return c.analyzer.ComputeSimilarityTrees(f1.TreeNode, f2.TreeNode, nil)
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

	// Classify the group based on its similarity
	c.classifyGroupBySimilarity(group, group.Similarity)
}

// classifyGroupBySimilarity classifies a group based on similarity using configured thresholds
func (c *CentroidGrouping) classifyGroupBySimilarity(group *CloneGroup, similarity float64) {
	group.Similarity = similarity

	// Determine clone type based on similarity using configured thresholds
	switch {
	case similarity >= c.type1Threshold:
		group.CloneType = Type1Clone
	case similarity >= c.type2Threshold:
		group.CloneType = Type2Clone
	case similarity >= c.type3Threshold:
		group.CloneType = Type3Clone
	case similarity >= c.type4Threshold:
		group.CloneType = Type4Clone
	default:
		group.CloneType = Type4Clone
	}
}
