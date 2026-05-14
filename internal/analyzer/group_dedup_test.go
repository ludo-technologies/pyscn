package analyzer

import (
	"sort"
	"testing"
)

// makeGroup builds a CloneGroup populated from the given fragments via the
// public constructor + AddFragment to mirror real strategy output.
func makeGroup(id int, frags ...*CodeFragment) *CloneGroup {
	g := NewCloneGroup(id)
	for _, f := range frags {
		g.AddFragment(f)
	}
	return g
}

func dedupeGroups(groups []*CloneGroup, pairs ...*ClonePair) []*CloneGroup {
	return dedupeStrictSubsetGroupMembers(groups, pairs).groups
}

func gpType(a, b *CodeFragment, sim float64, cloneType CloneType) *ClonePair {
	return &ClonePair{Fragment1: a, Fragment2: b, Similarity: sim, CloneType: cloneType}
}

// rangesOf returns a sorted slice of "file:start-end" labels for a group's
// fragments, providing order-independent comparison in assertions.
func rangesOf(g *CloneGroup) []string {
	out := make([]string, 0, len(g.Fragments))
	for _, f := range g.Fragments {
		out = append(out, f.Location.String())
	}
	sort.Strings(out)
	return out
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDedupe_AutofeatRepro reproduces the exact shape from issue #416:
// a same-file pair {512-542, 515-542} (one a strict subset of the other) and
// a legitimate cross-method clone {548-575}. The 515-542 member must be
// suppressed.
func TestDedupe_AutofeatRepro(t *testing.T) {
	a := gf("x.py", 512, 542)
	b := gf("x.py", 515, 542)
	c := gf("x.py", 548, 575)
	g := makeGroup(1, a, b, c)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 1 {
		t.Fatalf("expected 1 group, got %d", len(out))
	}
	if out[0].Size != 2 {
		t.Fatalf("expected Size=2 after dedup, got %d", out[0].Size)
	}
	got := rangesOf(out[0])
	want := []string{a.Location.String(), c.Location.String()}
	sort.Strings(want)
	if !equalStringSlices(got, want) {
		t.Fatalf("expected members %v, got %v", want, got)
	}
}

// TestDedupe_EqualRangeKeepsFirst verifies that exactly-equal duplicates
// (same file, same start/end) collapse to a single member with a deterministic
// tiebreak (first occurrence wins).
func TestDedupe_EqualRangeKeepsFirst(t *testing.T) {
	a := gf("x.py", 1, 10)
	aDup := gf("x.py", 1, 10)
	c := gf("y.py", 1, 10)
	g := makeGroup(1, a, aDup, c)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 1 {
		t.Fatalf("expected 1 group, got %d", len(out))
	}
	if out[0].Size != 2 {
		t.Fatalf("expected Size=2 after dedup, got %d", out[0].Size)
	}
	// First a (index 0) and c (different file) survive; aDup (index 1) drops.
	for _, f := range out[0].Fragments {
		if f == aDup {
			t.Fatalf("duplicate-range second occurrence should have been suppressed")
		}
	}
	foundA, foundC := false, false
	for _, f := range out[0].Fragments {
		if f == a {
			foundA = true
		}
		if f == c {
			foundC = true
		}
	}
	if !foundA || !foundC {
		t.Fatalf("expected first-occurrence a and c to remain, got fragments %+v", out[0].Fragments)
	}
}

// TestDedupe_ChainCollapses verifies that a chain m3 ⊂ m2 ⊂ m1 in the same
// file collapses to {m1}, plus an unrelated cross-file member.
func TestDedupe_ChainCollapses(t *testing.T) {
	a := gf("x.py", 1, 100)
	b := gf("x.py", 10, 90)
	c := gf("x.py", 20, 80)
	d := gf("y.py", 1, 50)
	g := makeGroup(1, a, b, c, d)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 1 || out[0].Size != 2 {
		t.Fatalf("expected 1 group of size 2, got %+v", out)
	}
	for _, f := range out[0].Fragments {
		if f == b || f == c {
			t.Fatalf("inner chain members b/c should have been suppressed, got %+v", out[0].Fragments)
		}
	}
}

// TestDedupe_DifferentFilesUntouched verifies that fragments in different
// files are never suppressed regardless of line-range relationship.
func TestDedupe_DifferentFilesUntouched(t *testing.T) {
	a := gf("x.py", 1, 10)
	b := gf("y.py", 1, 10) // identical range but different file
	c := gf("z.py", 1, 10)
	g := makeGroup(1, a, b, c)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 1 || out[0].Size != 3 {
		t.Fatalf("expected 1 group of size 3, got %+v", out)
	}
}

// TestDedupe_GroupShrinksToSingleton verifies a group with only a covering
// pair from one file (no cross-file neighbor) is dropped entirely.
func TestDedupe_GroupShrinksToSingleton(t *testing.T) {
	a := gf("x.py", 1, 100)
	b := gf("x.py", 10, 90)
	g := makeGroup(1, a, b)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 0 {
		t.Fatalf("expected 0 groups after singleton drop, got %d", len(out))
	}
}

// TestDedupe_DisjointSameFileBothKept verifies that disjoint same-file
// fragments (e.g., two methods in the same module) are not suppressed.
func TestDedupe_DisjointSameFileBothKept(t *testing.T) {
	a := gf("x.py", 1, 10)
	b := gf("x.py", 20, 30)
	g := makeGroup(1, a, b)

	out := dedupeGroups([]*CloneGroup{g})

	if len(out) != 1 || out[0].Size != 2 {
		t.Fatalf("expected 1 group of size 2, got %+v", out)
	}
}

// TestDedupe_EmptyAndNilGroups verifies passthrough behavior for degenerate
// inputs.
func TestDedupe_EmptyAndNilGroups(t *testing.T) {
	if got := dedupeGroups(nil); len(got) != 0 {
		t.Fatalf("nil input should return empty result, got %d groups", len(got))
	}
	if got := dedupeGroups([]*CloneGroup{}); len(got) != 0 {
		t.Fatalf("empty input should return empty result, got %d groups", len(got))
	}
	// A nil group entry should be skipped without panic.
	out := dedupeGroups([]*CloneGroup{nil})
	if len(out) != 0 {
		t.Fatalf("nil group entry should be skipped, got %d groups", len(out))
	}
}

// TestDedupe_RecomputesMetadataAfterSuppression verifies group-level metadata
// reflects the members that remain after pruning, not the larger pre-dedup set.
func TestDedupe_RecomputesMetadataAfterSuppression(t *testing.T) {
	a := gf("x.py", 1, 10)
	b := gf("x.py", 2, 10)
	c := gf("y.py", 1, 10)
	g := makeGroup(1, a, b, c)
	g.Similarity = 0.99
	g.CloneType = Type1Clone
	pairs := []*ClonePair{
		gpType(a, c, 0.70, Type4Clone),
		gpType(b, c, 0.99, Type1Clone),
	}

	out := dedupeStrictSubsetGroupMembers([]*CloneGroup{g}, pairs).groups

	if len(out) != 1 {
		t.Fatalf("expected 1 group, got %d", len(out))
	}
	if !almostEqual(out[0].Similarity, 0.70) {
		t.Fatalf("expected recomputed similarity 0.70, got %.3f", out[0].Similarity)
	}
	if out[0].CloneType != Type4Clone {
		t.Fatalf("expected recomputed clone type Type4, got %v", out[0].CloneType)
	}
}

// TestGroupClonesWithStrategy_FiltersSuppressedClonePairs verifies pair output
// cannot keep reporting a strict-subset member hidden from clone groups.
func TestGroupClonesWithStrategy_FiltersSuppressedClonePairs(t *testing.T) {
	a := gf("x.py", 1, 10)
	b := gf("x.py", 2, 10)
	c := gf("y.py", 1, 10)
	cd := &CloneDetector{
		clonePairs: []*ClonePair{
			gpType(a, c, 0.95, Type2Clone),
			gpType(b, c, 0.99, Type1Clone),
		},
	}

	cd.groupClonesWithStrategy(NewConnectedGrouping(0.85))

	if len(cd.cloneGroups) != 1 || cd.cloneGroups[0].Size != 2 {
		t.Fatalf("expected one deduped group of size 2, got %+v", cd.cloneGroups)
	}
	if len(cd.clonePairs) != 1 {
		t.Fatalf("expected suppressed-member pair to be filtered, got %d pairs", len(cd.clonePairs))
	}
	if cd.clonePairs[0].Fragment1 == b || cd.clonePairs[0].Fragment2 == b {
		t.Fatalf("clonePairs still contains suppressed fragment")
	}
}

// TestDedupe_E2EThroughConnectedGrouping is the end-to-end integration test:
// pairs (A, C) and (B, C) where A and B are overlapping same-file fragments
// — isOverlappingLocation prevents a direct (A, B) pair, but Union-Find unites
// {A, B, C} via C. After dedup the group must contain {A, C}.
func TestDedupe_E2EThroughConnectedGrouping(t *testing.T) {
	A := gf("x.py", 512, 542)
	B := gf("x.py", 515, 542)
	C := gf("y.py", 1, 30)
	pairs := []*ClonePair{
		gp(A, C, 0.95),
		gp(B, C, 0.95),
	}

	groups := NewConnectedGrouping(0.85).GroupClones(pairs)
	out := dedupeGroups(groups, pairs...)

	if len(out) != 1 {
		t.Fatalf("expected 1 group end-to-end, got %d", len(out))
	}
	if out[0].Size != 2 {
		t.Fatalf("expected Size=2 after dedup, got %d", out[0].Size)
	}
	for _, f := range out[0].Fragments {
		if f == B {
			t.Fatalf("B (strict subset of A in same file) should have been suppressed")
		}
	}
}
