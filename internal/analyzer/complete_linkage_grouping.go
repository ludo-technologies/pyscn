package analyzer

import "sort"

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

	fragments, sims, types := c.collectFragmentsAndSimilarities(pairs)
	n := len(fragments)
	if n < 2 {
		return []*CloneGroup{}
	}

	clusters := make([][]*CodeFragment, n)
	active := make([]bool, n)
	versions := make([]int, n)
	for i, f := range fragments {
		clusters[i] = []*CodeFragment{f}
		active[i] = true
	}

	// Cache complete-linkage similarities so merges only need O(k) updates via
	// sim(AB, C) = min(sim(A, C), sim(B, C)).
	clusterSims := c.buildClusterSimilarityMatrix(fragments, sims)
	candidateHeap := c.buildCandidateHeap(clusterSims, versions)

	for {
		bestPair, ok := c.popBestClusterPair(candidateHeap, active, versions)
		if !ok {
			break
		}

		targetID := bestPair.leftID
		sourceID := bestPair.rightID
		clusters[targetID] = append(clusters[targetID], clusters[sourceID]...)
		active[sourceID] = false
		versions[targetID]++
		c.updateMergedClusterSimilarities(active, clusterSims, versions, candidateHeap, targetID, sourceID)
		clusters[sourceID] = nil
	}

	return c.buildGroups(active, clusters, sims, types)
}

func (c *CompleteLinkageGrouping) collectFragmentsAndSimilarities(pairs []*ClonePair) ([]*CodeFragment, map[string]float64, map[string]CloneType) {
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

	return fragments, sims, types
}

func (c *CompleteLinkageGrouping) buildClusterSimilarityMatrix(fragments []*CodeFragment, sims map[string]float64) [][]float64 {
	clusterSims := make([][]float64, len(fragments))
	for i := range clusterSims {
		clusterSims[i] = make([]float64, len(fragments))
		clusterSims[i][i] = 1.0
	}

	for i := 0; i < len(fragments); i++ {
		for j := i + 1; j < len(fragments); j++ {
			s := similarity(sims, fragments[i], fragments[j])
			clusterSims[i][j] = s
			clusterSims[j][i] = s
		}
	}

	return clusterSims
}

func (c *CompleteLinkageGrouping) buildCandidateHeap(clusterSims [][]float64, versions []int) *completeLinkageCandidateHeap {
	pairHeap := &completeLinkageCandidateHeap{}
	for i := 0; i < len(clusterSims); i++ {
		for j := i + 1; j < len(clusterSims); j++ {
			c.pushCandidate(pairHeap, versions, i, j, clusterSims[i][j])
		}
	}
	return pairHeap
}

func (c *CompleteLinkageGrouping) popBestClusterPair(pairHeap *completeLinkageCandidateHeap, active []bool, versions []int) (completeLinkageCandidate, bool) {
	for pairHeap.Len() > 0 {
		bestPair := pairHeap.pop()
		if !active[bestPair.leftID] || !active[bestPair.rightID] {
			continue
		}
		if versions[bestPair.leftID] != bestPair.leftVersion || versions[bestPair.rightID] != bestPair.rightVersion {
			continue
		}

		return bestPair, true
	}

	return completeLinkageCandidate{}, false
}

func (c *CompleteLinkageGrouping) pushCandidate(pairHeap *completeLinkageCandidateHeap, versions []int, firstID, secondID int, score float64) {
	if score < c.threshold {
		return
	}

	leftID, rightID := orderClusterIDs(firstID, secondID)
	pairHeap.push(completeLinkageCandidate{
		leftID:       leftID,
		rightID:      rightID,
		score:        score,
		leftVersion:  versions[leftID],
		rightVersion: versions[rightID],
	})
}

func (c *CompleteLinkageGrouping) updateMergedClusterSimilarities(active []bool, clusterSims [][]float64, versions []int, pairHeap *completeLinkageCandidateHeap, targetID, sourceID int) {
	clusterSims[targetID][sourceID] = 0.0
	clusterSims[sourceID][targetID] = 0.0
	clusterSims[sourceID][sourceID] = 0.0

	for otherID, isActive := range active {
		if !isActive || otherID == targetID {
			continue
		}

		mergedSim := clusterSims[targetID][otherID]
		if sourceSim := clusterSims[sourceID][otherID]; sourceSim < mergedSim {
			mergedSim = sourceSim
		}
		clusterSims[targetID][otherID] = mergedSim
		clusterSims[otherID][targetID] = mergedSim
		c.pushCandidate(pairHeap, versions, targetID, otherID, mergedSim)
	}

	clusterSims[targetID][targetID] = 1.0
}

func (c *CompleteLinkageGrouping) buildGroups(active []bool, clusters [][]*CodeFragment, sims map[string]float64, types map[string]CloneType) []*CloneGroup {
	groups := make([]*CloneGroup, 0)
	groupID := 0
	for clusterID, isActive := range active {
		if !isActive {
			continue
		}

		cl := clusters[clusterID]
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
