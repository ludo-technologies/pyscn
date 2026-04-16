package analyzer

import (
	"fmt"
	"sort"
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
