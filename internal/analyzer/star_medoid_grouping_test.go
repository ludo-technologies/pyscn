package analyzer

import (
    "fmt"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
)

func newFrag(path string, sl, el int) *CodeFragment {
    return &CodeFragment{Location: &CodeLocation{FilePath: path, StartLine: sl, EndLine: el}}
}

func newPair(a, b *CodeFragment, sim float64) *ClonePair {
    return &ClonePair{Fragment1: a, Fragment2: b, Similarity: sim}
}

func TestStarMedoidGrouping_SimpleCase(t *testing.T) {
    // A-B: 0.9, B-C: 0.8, A-C: 0.3
    // threshold = 0.85 -> expect only {A,B}
    A := newFrag("A.py", 1, 5)
    B := newFrag("B.py", 1, 5)
    C := newFrag("C.py", 1, 5)

    pairs := []*ClonePair{
        newPair(A, B, 0.90),
        newPair(B, C, 0.80),
        newPair(A, C, 0.30),
    }

    g := NewStarMedoidGrouping(0.85)
    groups := g.GroupClones(pairs)

    if !assert.Len(t, groups, 1, "Should form exactly one group") {
        t.Fatalf("unexpected groups: %+v", groups)
    }
    grp := groups[0]
    assert.Equal(t, 2, grp.Size)
    // Check that A and B are members, C excluded
    hasA := false
    hasB := false
    for _, f := range grp.Fragments {
        if f == A {
            hasA = true
        }
        if f == B {
            hasB = true
        }
        if f == C {
            t.Fatalf("C should be excluded from the group")
        }
    }
    assert.True(t, hasA && hasB, "Group should contain A and B")
}

func TestStarMedoidGrouping_MultipleGroups(t *testing.T) {
    // Two independent groups and a singleton (excluded)
    A1 := newFrag("g1_a1.py", 1, 3)
    A2 := newFrag("g1_a2.py", 1, 3)
    A3 := newFrag("g1_a3.py", 1, 3)
    B1 := newFrag("g2_b1.py", 1, 3)
    B2 := newFrag("g2_b2.py", 1, 3)
    Z := newFrag("z.py", 1, 3)

    pairs := []*ClonePair{
        // Group A
        newPair(A1, A2, 0.90),
        newPair(A1, A3, 0.88),
        newPair(A2, A3, 0.86),
        // Group B
        newPair(B1, B2, 0.92),
        // Cross-group noise (below threshold)
        newPair(A1, B1, 0.20),
        newPair(A2, B2, 0.25),
        // Z has no high-similarity links
        newPair(Z, A1, 0.10),
    }

    g := NewStarMedoidGrouping(0.85)
    groups := g.GroupClones(pairs)

    // Expect two groups: sizes 3 and 2
    if !assert.Len(t, groups, 2, "Should form two groups") {
        t.Fatalf("unexpected groups len: %d", len(groups))
    }
    sizes := []int{groups[0].Size, groups[1].Size}
    if sizes[0] > sizes[1] {
        // ok
    } else {
        sizes[0], sizes[1] = sizes[1], sizes[0]
    }
    assert.Equal(t, []int{3, 2}, sizes)

    // Ensure Z not included in any group
    for _, grp := range groups {
        for _, f := range grp.Fragments {
            if f == Z {
                t.Fatalf("singleton Z should be excluded")
            }
        }
    }
}

func TestStarMedoidGrouping_NoGroups(t *testing.T) {
    A := newFrag("a.py", 1, 2)
    B := newFrag("b.py", 1, 2)
    C := newFrag("c.py", 1, 2)

    pairs := []*ClonePair{
        newPair(A, B, 0.60),
        newPair(B, C, 0.65),
        newPair(A, C, 0.55),
    }

    g := NewStarMedoidGrouping(0.80)
    groups := g.GroupClones(pairs)
    assert.Len(t, groups, 0, "All similarities below threshold; no groups expected")
}

func TestStarMedoidGrouping_LargeScale(t *testing.T) {
    // Build 10 clusters of 12 fragments (total 120+)
    clusterCount := 10
    perCluster := 12
    total := clusterCount * perCluster
    frags := make([]*CodeFragment, 0, total)
    for i := 0; i < total; i++ {
        frags = append(frags, newFrag(fmt.Sprintf("file_%03d.py", i), 1, 3))
    }

    // Build star-like pairs within each cluster: center is index base
    basePairs := make([]*ClonePair, 0)
    for c := 0; c < clusterCount; c++ {
        base := c * perCluster
        center := frags[base]
        for j := 1; j < perCluster; j++ {
            basePairs = append(basePairs, newPair(center, frags[base+j], 0.90))
        }
    }

    g := NewStarMedoidGrouping(0.85)
    start := time.Now()
    groups := g.GroupClones(basePairs)
    elapsed := time.Since(start)

    // Expect clusterCount groups
    assert.Len(t, groups, clusterCount, "Should form one group per cluster")
    // Each should have perCluster members
    for _, grp := range groups {
        assert.Equal(t, perCluster, grp.Size)
    }
    // Sanity runtime guard (should be fast on 120 frags)
    if elapsed > 2*time.Second {
        t.Fatalf("Large-scale grouping took too long: %s", elapsed)
    }
}

