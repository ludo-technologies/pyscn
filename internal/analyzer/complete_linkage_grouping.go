package analyzer

import "sort"

// CompleteLinkageGrouping ensures all pairs in a group meet the threshold.
type CompleteLinkageGrouping struct {
	threshold float64
}

func NewCompleteLinkageGrouping(threshold float64) *CompleteLinkageGrouping {
	return &CompleteLinkageGrouping{threshold: threshold}
}

func (c *CompleteLinkageGrouping) GetName() string { return "Complete Linkage" }

func (c *CompleteLinkageGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	input := c.collectInput(pairs)
	if len(input.fragments) < 2 {
		return []*CloneGroup{}
	}

	clusterer := newCompleteLinkageClusterer(input.fragments, input.edges)
	clusterer.mergeUntilStable()

	return c.buildGroups(clusterer.activeClusters(), input.similarities, input.types)
}

type completeLinkageInput struct {
	fragments    []*CodeFragment
	similarities map[string]float64
	types        map[string]CloneType
	edges        []completeLinkageEdge
}

type completeLinkagePairRecord struct {
	left       *CodeFragment
	right      *CodeFragment
	similarity float64
	cloneType  CloneType
}

type completeLinkageEdge struct {
	leftID  int
	rightID int
	score   float64
}

func (c *CompleteLinkageGrouping) collectInput(pairs []*ClonePair) completeLinkageInput {
	input := completeLinkageInput{
		fragments:    make([]*CodeFragment, 0),
		similarities: make(map[string]float64),
		types:        make(map[string]CloneType),
	}

	seen := make(map[*CodeFragment]struct{})
	pairRecords := make(map[string]completeLinkagePairRecord)
	for _, pair := range pairs {
		if pair == nil || pair.Fragment1 == nil || pair.Fragment2 == nil {
			continue
		}

		if _, ok := seen[pair.Fragment1]; !ok {
			seen[pair.Fragment1] = struct{}{}
			input.fragments = append(input.fragments, pair.Fragment1)
		}
		if _, ok := seen[pair.Fragment2]; !ok {
			seen[pair.Fragment2] = struct{}{}
			input.fragments = append(input.fragments, pair.Fragment2)
		}

		key := pairKey(pair.Fragment1, pair.Fragment2)
		record, ok := pairRecords[key]
		if !ok || pair.Similarity > record.similarity {
			pairRecords[key] = completeLinkagePairRecord{
				left:       pair.Fragment1,
				right:      pair.Fragment2,
				similarity: pair.Similarity,
				cloneType:  pair.CloneType,
			}
			input.similarities[key] = pair.Similarity
			input.types[key] = pair.CloneType
		}
	}

	fragmentIDs := make(map[*CodeFragment]int, len(input.fragments))
	for fragmentID, fragment := range input.fragments {
		fragmentIDs[fragment] = fragmentID
	}

	input.edges = make([]completeLinkageEdge, 0, len(pairRecords))
	for _, record := range pairRecords {
		if record.similarity < c.threshold {
			continue
		}

		leftID, rightID := orderClusterIDs(fragmentIDs[record.left], fragmentIDs[record.right])
		input.edges = append(input.edges, completeLinkageEdge{
			leftID:  leftID,
			rightID: rightID,
			score:   record.similarity,
		})
	}

	return input
}

func (c *CompleteLinkageGrouping) buildGroups(activeClusters []*completeLinkageCluster, similarities map[string]float64, types map[string]CloneType) []*CloneGroup {
	groups := make([]*CloneGroup, 0, len(activeClusters))
	groupID := 0
	for _, cluster := range activeClusters {
		members := cluster.members
		if len(members) < 2 {
			continue
		}

		// Keep a final safety check so the optimized clusterer cannot return a
		// non-clique even if an internal update regresses later.
		valid := true
		for i := 0; i < len(members) && valid; i++ {
			for j := i + 1; j < len(members); j++ {
				if similarity(similarities, members[i], members[j]) < c.threshold {
					valid = false
					break
				}
			}
		}
		if !valid {
			continue
		}

		sortedMembers := append([]*CodeFragment(nil), members...)
		sort.Slice(sortedMembers, func(i, j int) bool { return fragmentLess(sortedMembers[i], sortedMembers[j]) })

		group := NewCloneGroup(groupID)
		groupID++
		for _, fragment := range sortedMembers {
			group.AddFragment(fragment)
		}
		group.Similarity = averageGroupSimilarity(similarities, sortedMembers)
		group.CloneType = majorityCloneType(types, sortedMembers)
		groups = append(groups, group)
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
