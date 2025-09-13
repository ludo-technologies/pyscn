package analyzer

import (
	"fmt"
	"testing"
)

// Helpers for building fragments and pairs
func gf(path string, sl, el int) *CodeFragment {
	return &CodeFragment{Location: &CodeLocation{FilePath: path, StartLine: sl, EndLine: el}}
}
func gp(a, b *CodeFragment, sim float64) *ClonePair {
	return &ClonePair{Fragment1: a, Fragment2: b, Similarity: sim}
}

func TestGroupingStrategiesComparison(t *testing.T) {
	// Chain: A-B-C-D; only adjacent pairs are strong
	A := gf("A.py", 1, 3)
	B := gf("B.py", 1, 3)
	C := gf("C.py", 1, 3)
	D := gf("D.py", 1, 3)
	chainPairs := []*ClonePair{
		gp(A, B, 0.90),
		gp(B, C, 0.90),
		gp(C, D, 0.90),
		gp(A, C, 0.50), // weak
		gp(B, D, 0.50), // weak
		gp(A, D, 0.30), // weak
	}

	// Star: S is center with strong edges; leaves weak between each other
	S := gf("S.py", 1, 3)
	L1 := gf("L1.py", 1, 3)
	L2 := gf("L2.py", 1, 3)
	L3 := gf("L3.py", 1, 3)
	starPairs := []*ClonePair{
		gp(S, L1, 0.92), gp(S, L2, 0.91), gp(S, L3, 0.90),
		gp(L1, L2, 0.10), gp(L2, L3, 0.10), gp(L1, L3, 0.10),
	}

	// Clique: all strong
	X1 := gf("X1.py", 1, 3)
	X2 := gf("X2.py", 1, 3)
	X3 := gf("X3.py", 1, 3)
	cliquePairs := []*ClonePair{
		gp(X1, X2, 0.95), gp(X2, X3, 0.96), gp(X1, X3, 0.97),
	}

	thr := 0.85

	// Connected should allow chaining
	conn := NewConnectedGrouping(thr)
	g1 := conn.GroupClones(chainPairs)
	if len(g1) == 0 || g1[0].Size < 3 {
		t.Fatalf("Connected should allow chain grouping (expected size>=3), got %+v", sizes(g1))
	}

	// Complete-linkage should avoid chaining; groups should be small (<=2)
	comp := NewCompleteLinkageGrouping(thr)
	g2 := comp.GroupClones(chainPairs)
	for _, g := range g2 {
		if g.Size > 2 {
			t.Fatalf("Complete-linkage should not form large chain groups: got size=%d", g.Size)
		}
	}

	// K-core with k=2 on chain should produce no groups (ends drop degrees)
	kc := NewKCoreGrouping(thr, 2)
	g3 := kc.GroupClones(chainPairs)
	if len(g3) != 0 {
		t.Fatalf("K-core k=2 should eliminate chain components, got %v", sizes(g3))
	}

	// Star-medoid should yield a star group centered at S
	star := NewStarMedoidGrouping(thr)
	g4 := star.GroupClones(starPairs)
	if len(g4) == 0 || g4[0].Size < 3 {
		t.Fatalf("Star/Medoid should form a star group, got %+v", sizes(g4))
	}

	// Complete-linkage on star should not create big groups (>2)
	g5 := comp.GroupClones(starPairs)
	for _, g := range g5 {
		if g.Size > 2 {
			t.Fatalf("Complete-linkage on star should limit groups to pairs, got size=%d", g.Size)
		}
	}

	// All strategies on clique should form one group of size 3
	if sz := NewConnectedGrouping(thr).GroupClones(cliquePairs); !(len(sz) == 1 && sz[0].Size == 3) {
		t.Fatalf("Connected on clique should form one group of 3, got %+v", sizes(sz))
	}
	if sz := NewCompleteLinkageGrouping(thr).GroupClones(cliquePairs); !(len(sz) == 1 && sz[0].Size == 3) {
		t.Fatalf("Complete-linkage on clique should form one group of 3, got %+v", sizes(sz))
	}
	if sz := NewKCoreGrouping(thr, 2).GroupClones(cliquePairs); !(len(sz) == 1 && sz[0].Size == 3) {
		t.Fatalf("K-core on clique should form one group of 3, got %+v", sizes(sz))
	}
	if sz := NewStarMedoidGrouping(thr).GroupClones(cliquePairs); !(len(sz) == 1 && sz[0].Size == 3) {
		t.Fatalf("Star/Medoid on clique should form one group of 3, got %+v", sizes(sz))
	}
}

func sizes(gs []*CloneGroup) []int {
	out := make([]int, 0, len(gs))
	for _, g := range gs {
		out = append(out, g.Size)
	}
	return out
}

func BenchmarkGroupingStrategies(b *testing.B) {
	sizes := []int{100, 500, 1000}
	thr := 0.85
	for _, n := range sizes {
		// Build synthetic data: m clusters each of ~10 with center-leaf pairs
		per := 10
		clusters := n / per
		frags := make([]*CodeFragment, 0, clusters*per)
		for i := 0; i < clusters*per; i++ {
			frags = append(frags, gf(fmt.Sprintf("F_%d.py", i), 1, 3))
		}
		pairs := make([]*ClonePair, 0, clusters*(per-1))
		for c := 0; c < clusters; c++ {
			base := c * per
			center := frags[base]
			for j := 1; j < per; j++ {
				pairs = append(pairs, gp(center, frags[base+j], 0.9))
			}
		}

		b.Run(fmt.Sprintf("connected_%d", n), func(b *testing.B) {
			g := NewConnectedGrouping(thr)
			for i := 0; i < b.N; i++ {
				_ = g.GroupClones(pairs)
			}
		})
		b.Run(fmt.Sprintf("kcore_%d", n), func(b *testing.B) {
			g := NewKCoreGrouping(thr, 2)
			for i := 0; i < b.N; i++ {
				_ = g.GroupClones(pairs)
			}
		})
		b.Run(fmt.Sprintf("completelink_%d", n), func(b *testing.B) {
			g := NewCompleteLinkageGrouping(thr)
			for i := 0; i < b.N; i++ {
				_ = g.GroupClones(pairs)
			}
		})
		b.Run(fmt.Sprintf("star_%d", n), func(b *testing.B) {
			g := NewStarMedoidGrouping(thr)
			for i := 0; i < b.N; i++ {
				_ = g.GroupClones(pairs)
			}
		})
	}
}
