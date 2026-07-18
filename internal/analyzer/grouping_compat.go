package analyzer

import (
	"sort"

	coreclone "github.com/ludo-technologies/polyscan/core/clone"
	coredomain "github.com/ludo-technologies/polyscan/core/domain"
)

type fragmentPairID struct {
	first  int
	second int
}

func newFragmentPairID(first, second int) fragmentPairID {
	if first > second {
		first, second = second, first
	}
	return fragmentPairID{first: first, second: second}
}

// filterStarGroupsByMedoid preserves pyscn's star-mode contract: every
// non-medoid member must meet the configured similarity threshold to the medoid.
func filterStarGroupsByMedoid(
	groups []*coreclone.ItemGroup[*CodeFragment],
	pairs []*coreclone.ItemPair[*CodeFragment],
	threshold float64,
) []*coreclone.ItemGroup[*CodeFragment] {
	similarities := make(map[fragmentPairID]float64, len(pairs))
	for _, pair := range pairs {
		if pair == nil || pair.Item1 == nil || pair.Item2 == nil {
			continue
		}
		key := newFragmentPairID(pair.Item1.ItemID(), pair.Item2.ItemID())
		if pair.Similarity > similarities[key] {
			similarities[key] = pair.Similarity
		}
	}

	filtered := make([]*coreclone.ItemGroup[*CodeFragment], 0, len(groups))
	for _, group := range groups {
		if group == nil || len(group.Items) < 2 {
			continue
		}
		medoid := group.Items[0]
		bestAverage := -1.0
		for _, candidate := range group.Items {
			total := 0.0
			for _, other := range group.Items {
				if candidate != other {
					total += similarities[newFragmentPairID(candidate.ItemID(), other.ItemID())]
				}
			}
			average := total / float64(len(group.Items)-1)
			if average > bestAverage || average == bestAverage && fragmentLocationLess(candidate, medoid) {
				medoid = candidate
				bestAverage = average
			}
		}

		members := make([]*CodeFragment, 0, len(group.Items))
		for _, member := range group.Items {
			if member == medoid || similarities[newFragmentPairID(member.ItemID(), medoid.ItemID())] >= threshold {
				members = append(members, member)
			}
		}
		if len(members) < 2 {
			continue
		}
		group.Items = members
		filtered = append(filtered, group)
	}
	return filtered
}

func fragmentLocationLess(first, second *CodeFragment) bool {
	if first == second {
		return false
	}
	if first == nil || first.Location == nil {
		return second != nil && second.Location != nil
	}
	if second == nil || second.Location == nil {
		return false
	}
	left, right := first.Location, second.Location
	if left.FilePath != right.FilePath {
		return left.FilePath < right.FilePath
	}
	if left.StartLine != right.StartLine {
		return left.StartLine < right.StartLine
	}
	if left.StartCol != right.StartCol {
		return left.StartCol < right.StartCol
	}
	if left.EndLine != right.EndLine {
		return left.EndLine < right.EndLine
	}
	return left.EndCol < right.EndCol
}

type centroidCompatibilityStrategy struct {
	threshold float64
	analyzer  *APTEDAnalyzer
}

func newCentroidCompatibilityStrategy(cd *CloneDetector, threshold float64) *centroidCompatibilityStrategy {
	return &centroidCompatibilityStrategy{
		threshold: threshold,
		analyzer:  NewAPTEDAnalyzer(buildCloneCostModel(&cd.cloneDetectorConfig)),
	}
}

func (s *centroidCompatibilityStrategy) Name() string { return string(coreclone.ModeCentroid) }

// GroupItems keeps centroid's historical BFS expansion, including computing
// similarities that were not retained in the reported pair set.
func (s *centroidCompatibilityStrategy) GroupItems(pairs []*coreclone.ItemPair[*CodeFragment]) []*coreclone.ItemGroup[*CodeFragment] {
	itemsByID := make(map[int]*CodeFragment)
	similarities := make(map[fragmentPairID]float64, len(pairs))
	for _, pair := range pairs {
		if pair == nil || pair.Item1 == nil || pair.Item2 == nil {
			continue
		}
		itemsByID[pair.Item1.ItemID()] = pair.Item1
		itemsByID[pair.Item2.ItemID()] = pair.Item2
		similarities[newFragmentPairID(pair.Item1.ItemID(), pair.Item2.ItemID())] = pair.Similarity
	}
	items := make([]*CodeFragment, 0, len(itemsByID))
	for _, item := range itemsByID {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return fragmentLocationLess(items[i], items[j]) })

	assigned := make(map[int]bool, len(items))
	groups := make([]*coreclone.ItemGroup[*CodeFragment], 0)
	for _, seed := range items {
		if assigned[seed.ItemID()] {
			continue
		}
		assigned[seed.ItemID()] = true
		members := []*CodeFragment{seed}
		queue := []*CodeFragment{seed}
		for len(queue) > 0 && len(members) < 50 {
			current := queue[0]
			queue = queue[1:]
			for _, candidate := range items {
				if assigned[candidate.ItemID()] {
					continue
				}
				key := newFragmentPairID(current.ItemID(), candidate.ItemID())
				similarity, exists := similarities[key]
				if !exists {
					similarity = s.analyzer.ComputeSimilarity(current.TreeNode, candidate.TreeNode)
					similarities[key] = similarity
				}
				if similarity >= s.threshold {
					assigned[candidate.ItemID()] = true
					members = append(members, candidate)
					queue = append(queue, candidate)
				}
			}
		}
		if len(members) >= 2 {
			groups = append(groups, &coreclone.ItemGroup[*CodeFragment]{
				ID:    len(groups),
				Items: members,
			})
		}
	}
	return groups
}

func (cd *CloneDetector) refreshCentroidGroupMetadata(groups []*coreclone.ItemGroup[*CodeFragment]) {
	analyzer := NewAPTEDAnalyzer(buildCloneCostModel(&cd.cloneDetectorConfig))
	for _, group := range groups {
		total := 0.0
		count := 0
		for i := 0; i < len(group.Items); i++ {
			for j := i + 1; j < len(group.Items); j++ {
				total += analyzer.ComputeSimilarity(group.Items[i].TreeNode, group.Items[j].TreeNode)
				count++
			}
		}
		if count == 0 {
			continue
		}
		group.Similarity = total / float64(count)
		switch {
		case group.Similarity >= cd.cloneDetectorConfig.Type1Threshold:
			group.GroupType = coredomain.Type1Clone
		case group.Similarity >= cd.cloneDetectorConfig.Type2Threshold:
			group.GroupType = coredomain.Type2Clone
		case group.Similarity >= cd.cloneDetectorConfig.Type3Threshold:
			group.GroupType = coredomain.Type3Clone
		default:
			group.GroupType = coredomain.Type4Clone
		}
	}
}
