package analyzer

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
)

func TestCompleteLinkageGrouping_PreservesSeparateCliques(t *testing.T) {
	thr := 0.85

	a1 := gf("alpha_1.py", 1, 3)
	a2 := gf("alpha_2.py", 1, 3)
	a3 := gf("alpha_3.py", 1, 3)
	b1 := gf("beta_1.py", 1, 3)
	b2 := gf("beta_2.py", 1, 3)
	b3 := gf("beta_3.py", 1, 3)

	pairs := []*ClonePair{
		gp(a1, a2, 0.97),
		gp(a1, a2, 0.91), // lower duplicate should not win
		gp(a1, a3, 0.95),
		gp(a2, a3, 0.94),
		gp(b1, b2, 0.96),
		gp(b1, b3, 0.93),
		gp(b2, b3, 0.92),

		gp(a1, b1, 0.40),
		gp(a1, b2, 0.39),
		gp(a1, b3, 0.38),
		gp(a2, b1, 0.37),
		gp(a2, b2, 0.36),
		gp(a2, b3, 0.35),
		gp(a3, b1, 0.34),
		gp(a3, b2, 0.33),
		gp(a3, b3, 0.32),
	}

	groups := NewCompleteLinkageGrouping(thr).GroupClones(pairs)
	if len(groups) != 2 {
		t.Fatalf("expected two dense groups, got %d", len(groups))
	}

	got := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.Size != 3 {
			t.Fatalf("expected each group to keep three members, got size=%d", group.Size)
		}
		got = append(got, groupSignature(group))
	}
	sort.Strings(got)

	want := []string{
		"alpha_1.py,alpha_2.py,alpha_3.py",
		"beta_1.py,beta_2.py,beta_3.py",
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("unexpected groups: got %v want %v", got, want)
	}
}

func BenchmarkCompleteLinkageGroupingDenseCliques(b *testing.B) {
	thr := 0.85
	sizes := []int{20, 40, 80}

	for _, cliqueSize := range sizes {
		pairs := buildDenseCliquePairs(4, cliqueSize, 0.92, 0.40)
		b.Run(fmt.Sprintf("clique_size_%d", cliqueSize), func(b *testing.B) {
			grouping := NewCompleteLinkageGrouping(thr)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = grouping.GroupClones(pairs)
			}
		})
	}
}

func TestCompleteLinkageGrouping_PreservesScanOrderTieBreaks(t *testing.T) {
	thr := 0.75

	a := gf("a.py", 1, 3)
	b := gf("b.py", 1, 3)
	c := gf("c.py", 1, 3)
	pairs := []*ClonePair{
		gp(a, b, 0.90),
		gp(a, c, 0.90),
		gp(b, c, 0.70),
	}

	groups := NewCompleteLinkageGrouping(thr).GroupClones(pairs)
	if len(groups) != 1 {
		t.Fatalf("expected one group, got %d", len(groups))
	}

	got := generalGroupSignature(groups[0])
	want := "a.py,b.py"
	if got != want {
		t.Fatalf("expected stable tie-break group %s, got %s", want, got)
	}
}

func TestCompleteLinkageGrouping_MatchesReferenceImplementation(t *testing.T) {
	t.Parallel()

	thresholds := []float64{0.35, 0.65, 0.85}
	for _, thr := range thresholds {
		thr := thr
		t.Run(fmt.Sprintf("threshold_%0.2f", thr), func(t *testing.T) {
			t.Parallel()
			for seed := int64(1); seed <= 24; seed++ {
				pairs := buildRandomCompleteLinkagePairs(seed, 7)
				got := NewCompleteLinkageGrouping(thr).GroupClones(pairs)
				want := referenceCompleteLinkageGroups(thr, pairs)
				assertCloneGroupsEqual(t, want, got, thr, seed)
			}
		})
	}
}

func TestMajorityCloneType_TieBreaksDeterministically(t *testing.T) {
	a := gf("a.py", 1, 3)
	b := gf("b.py", 1, 3)
	c := gf("c.py", 1, 3)
	d := gf("d.py", 1, 3)
	members := []*CodeFragment{a, b, c, d}
	typeMap := map[string]CloneType{
		pairKey(a, b): Type3Clone,
		pairKey(a, c): Type1Clone,
		pairKey(a, d): Type3Clone,
		pairKey(b, c): Type1Clone,
	}

	got := majorityCloneType(typeMap, members)
	if got != Type1Clone {
		t.Fatalf("expected deterministic lower clone-type tie break, got %v", got)
	}
}

func buildDenseCliquePairs(groupCount, cliqueSize int, intraSim, interSim float64) []*ClonePair {
	fragments := make([][]*CodeFragment, groupCount)
	for groupIndex := 0; groupIndex < groupCount; groupIndex++ {
		groupFragments := make([]*CodeFragment, cliqueSize)
		for fragmentIndex := 0; fragmentIndex < cliqueSize; fragmentIndex++ {
			groupFragments[fragmentIndex] = gf(
				fmt.Sprintf("group_%d_fragment_%d.py", groupIndex, fragmentIndex),
				1,
				3,
			)
		}
		fragments[groupIndex] = groupFragments
	}

	pairs := make([]*ClonePair, 0, groupCount*cliqueSize*cliqueSize)
	for groupIndex := 0; groupIndex < groupCount; groupIndex++ {
		groupFragments := fragments[groupIndex]
		for i := 0; i < len(groupFragments); i++ {
			for j := i + 1; j < len(groupFragments); j++ {
				pairs = append(pairs, gp(groupFragments[i], groupFragments[j], intraSim))
			}
		}
	}

	for leftGroup := 0; leftGroup < groupCount; leftGroup++ {
		for rightGroup := leftGroup + 1; rightGroup < groupCount; rightGroup++ {
			for _, leftFragment := range fragments[leftGroup] {
				for _, rightFragment := range fragments[rightGroup] {
					pairs = append(pairs, gp(leftFragment, rightFragment, interSim))
				}
			}
		}
	}

	return pairs
}

func groupSignature(group *CloneGroup) string {
	paths := make([]string, 0, len(group.Fragments))
	for _, fragment := range group.Fragments {
		paths = append(paths, fragment.Location.FilePath)
	}
	sort.Strings(paths)
	return fmt.Sprintf("%s,%s,%s", paths[0], paths[1], paths[2])
}

func generalGroupSignature(group *CloneGroup) string {
	paths := make([]string, 0, len(group.Fragments))
	for _, fragment := range group.Fragments {
		paths = append(paths, fragment.Location.FilePath)
	}
	sort.Strings(paths)
	return strings.Join(paths, ",")
}

func buildRandomCompleteLinkagePairs(seed int64, fragmentCount int) []*ClonePair {
	rng := rand.New(rand.NewSource(seed))
	fragments := make([]*CodeFragment, fragmentCount)
	for i := 0; i < fragmentCount; i++ {
		fragments[i] = gf(fmt.Sprintf("fragment_%d.py", i), 1, 3)
	}

	pairs := make([]*ClonePair, 0, fragmentCount*fragmentCount)
	for i := 0; i < len(fragments); i++ {
		for j := i + 1; j < len(fragments); j++ {
			if rng.Float64() < 0.2 {
				continue
			}

			pairs = append(pairs, &ClonePair{
				Fragment1:  fragments[i],
				Fragment2:  fragments[j],
				Similarity: rng.Float64(),
				CloneType:  CloneType(rng.Intn(int(Type4Clone)) + 1),
			})

			if rng.Float64() < 0.35 {
				pairs = append(pairs, &ClonePair{
					Fragment1:  fragments[i],
					Fragment2:  fragments[j],
					Similarity: rng.Float64(),
					CloneType:  CloneType(rng.Intn(int(Type4Clone)) + 1),
				})
			}
		}
	}

	return pairs
}

func referenceCompleteLinkageGroups(threshold float64, pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

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
		key := pairKey(p.Fragment1, p.Fragment2)
		if old, ok := sims[key]; !ok || p.Similarity > old {
			sims[key] = p.Similarity
			types[key] = p.CloneType
		}
	}

	if len(fragments) < 2 {
		return []*CloneGroup{}
	}

	clusters := make([][]*CodeFragment, len(fragments))
	for i, fragment := range fragments {
		clusters[i] = []*CodeFragment{fragment}
	}

	clusterSim := func(firstCluster, secondCluster []*CodeFragment) float64 {
		minSim := 1.0
		hasPair := false
		for _, firstFragment := range firstCluster {
			for _, secondFragment := range secondCluster {
				sim := similarity(sims, firstFragment, secondFragment)
				if sim < threshold {
					return 0.0
				}
				if sim < minSim {
					minSim = sim
				}
				hasPair = true
			}
		}
		if !hasPair {
			return 0.0
		}
		return minSim
	}

	for {
		bestI, bestJ := -1, -1
		bestScore := -1.0
		for i := 0; i < len(clusters); i++ {
			for j := i + 1; j < len(clusters); j++ {
				sim := clusterSim(clusters[i], clusters[j])
				if sim >= threshold && sim > bestScore {
					bestScore = sim
					bestI = i
					bestJ = j
				}
			}
		}
		if bestI == -1 || bestJ == -1 {
			break
		}

		clusters[bestI] = append(clusters[bestI], clusters[bestJ]...)
		clusters = append(clusters[:bestJ], clusters[bestJ+1:]...)
	}

	groups := make([]*CloneGroup, 0)
	groupID := 0
	for _, cluster := range clusters {
		if len(cluster) < 2 {
			continue
		}

		valid := true
		for i := 0; i < len(cluster) && valid; i++ {
			for j := i + 1; j < len(cluster); j++ {
				if similarity(sims, cluster[i], cluster[j]) < threshold {
					valid = false
					break
				}
			}
		}
		if !valid {
			continue
		}

		sort.Slice(cluster, func(i, j int) bool { return fragmentLess(cluster[i], cluster[j]) })
		group := NewCloneGroup(groupID)
		groupID++
		for _, fragment := range cluster {
			group.AddFragment(fragment)
		}
		group.Similarity = averageGroupSimilarity(sims, cluster)
		group.CloneType = majorityCloneType(types, cluster)
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

func assertCloneGroupsEqual(t *testing.T, want, got []*CloneGroup, threshold float64, seed int64) {
	t.Helper()

	wantSnapshots := snapshotCloneGroups(want)
	gotSnapshots := snapshotCloneGroups(got)
	if len(wantSnapshots) != len(gotSnapshots) {
		t.Fatalf("threshold %.2f seed %d: group count mismatch: want %d got %d", threshold, seed, len(wantSnapshots), len(gotSnapshots))
	}

	for i := range wantSnapshots {
		if wantSnapshots[i].members != gotSnapshots[i].members {
			t.Fatalf("threshold %.2f seed %d: members mismatch: want %s got %s", threshold, seed, wantSnapshots[i].members, gotSnapshots[i].members)
		}
		if wantSnapshots[i].cloneType != gotSnapshots[i].cloneType {
			t.Fatalf("threshold %.2f seed %d: clone type mismatch for %s: want %v got %v", threshold, seed, wantSnapshots[i].members, wantSnapshots[i].cloneType, gotSnapshots[i].cloneType)
		}
		if !almostEqual(wantSnapshots[i].similarity, gotSnapshots[i].similarity) {
			t.Fatalf("threshold %.2f seed %d: similarity mismatch for %s: want %f got %f", threshold, seed, wantSnapshots[i].members, wantSnapshots[i].similarity, gotSnapshots[i].similarity)
		}
	}
}

type cloneGroupSnapshot struct {
	members    string
	similarity float64
	cloneType  CloneType
}

func snapshotCloneGroups(groups []*CloneGroup) []cloneGroupSnapshot {
	snapshots := make([]cloneGroupSnapshot, 0, len(groups))
	for _, group := range groups {
		memberIDs := make([]string, 0, len(group.Fragments))
		for _, fragment := range group.Fragments {
			memberIDs = append(memberIDs, fragmentID(fragment))
		}
		sort.Strings(memberIDs)
		snapshots = append(snapshots, cloneGroupSnapshot{
			members:    strings.Join(memberIDs, "||"),
			similarity: group.Similarity,
			cloneType:  group.CloneType,
		})
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].members != snapshots[j].members {
			return snapshots[i].members < snapshots[j].members
		}
		if !almostEqual(snapshots[i].similarity, snapshots[j].similarity) {
			return snapshots[i].similarity < snapshots[j].similarity
		}
		return snapshots[i].cloneType < snapshots[j].cloneType
	})

	return snapshots
}
