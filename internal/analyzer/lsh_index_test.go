package analyzer

import (
	"testing"

	corelsh "github.com/ludo-technologies/polyscan/core/lsh"
)

func TestLSHIndex_FindCandidates(t *testing.T) {
	// Build a small set of signatures
	mh := corelsh.NewMinHasher(128)
	fX := []string{"a", "b", "c", "d"}
	fY := []string{"a", "b", "c", "e"} // similar to X
	fZ := []string{"x", "y", "z"}      // dissimilar

	sigX := mh.ComputeSignature(fX)
	sigY := mh.ComputeSignature(fY)
	sigZ := mh.ComputeSignature(fZ)

	lsh := newLSHCandidateIndex(32, 4, 0)
	if err := lsh.AddFragment(1, sigX); err != nil {
		t.Fatalf("add X: %v", err)
	}
	if err := lsh.AddFragment(2, sigY); err != nil {
		t.Fatalf("add Y: %v", err)
	}
	if err := lsh.AddFragment(3, sigZ); err != nil {
		t.Fatalf("add Z: %v", err)
	}

	cands := lsh.FindCandidates(sigX)
	foundY := false
	for _, id := range cands {
		if id == 2 {
			foundY = true
			break
		}
	}
	if !foundY {
		t.Fatalf("expected Y to be a candidate for X; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesKeepsOversizedBucketCandidates(t *testing.T) {
	mh := corelsh.NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := newLSHCandidateIndex(32, 4, 2)
	for _, id := range []int{1, 2, 3} {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %d: %v", id, err)
		}
	}

	cands := lsh.FindCandidates(sig)
	if len(cands) != 2 || cands[0] != 1 || cands[1] != 2 {
		t.Fatalf("expected first-traversed candidates [1 2]; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesCapKeepsTraversalOrderSelection(t *testing.T) {
	mh := corelsh.NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := newLSHCandidateIndex(32, 4, 2)
	for _, id := range []int{10, 20, 1} {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %d: %v", id, err)
		}
	}

	// The cap must keep the first candidates in band/insertion traversal
	// order ({10, 20}), not the smallest IDs ({1, 10}); output is sorted.
	cands := lsh.FindCandidates(sig)
	if len(cands) != 2 || cands[0] != 10 || cands[1] != 20 {
		t.Fatalf("expected capped selection [10 20]; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesUsesDefaultCapAtBoundary(t *testing.T) {
	mh := corelsh.NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := newLSHCandidateIndex(32, 4, 0)
	for id := 0; id <= defaultLSHMaxCandidates; id++ {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %d: %v", id, err)
		}
	}

	cands := lsh.FindCandidates(sig)
	if len(cands) != defaultLSHMaxCandidates {
		t.Fatalf("candidate count mismatch: want %d got %d", defaultLSHMaxCandidates, len(cands))
	}
	for i, id := range cands {
		if id != i {
			t.Fatalf("candidate order mismatch at %d: got %d", i, id)
		}
	}
}

func TestLSHIndex_FindCandidatesReturnsDeterministicIndexes(t *testing.T) {
	mh := corelsh.NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := newLSHCandidateIndex(32, 4, 0)
	for _, id := range []int{4, 2, 3, 1} {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %d: %v", id, err)
		}
	}

	cands := lsh.FindCandidates(sig)
	want := []int{1, 2, 3, 4}
	if len(cands) != len(want) {
		t.Fatalf("candidate count mismatch: want %v got %v", want, cands)
	}
	for i := range want {
		if cands[i] != want[i] {
			t.Fatalf("candidate order mismatch: want %v got %v", want, cands)
		}
	}
}

func TestLSHIndex_AddFragmentRejectsNegativeID(t *testing.T) {
	lsh := newLSHCandidateIndex(32, 4, 10)
	sig := corelsh.NewMinHasher(128).ComputeSignature([]string{"feature"})
	if err := lsh.AddFragment(-1, sig); err == nil {
		t.Fatal("expected negative fragment ID to be rejected")
	}
}
