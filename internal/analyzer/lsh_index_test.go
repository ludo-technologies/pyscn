package analyzer

import "testing"

func TestLSHIndex_FindCandidates(t *testing.T) {
	// Build a small set of signatures
	mh := NewMinHasher(128)
	fX := []string{"a", "b", "c", "d"}
	fY := []string{"a", "b", "c", "e"} // similar to X
	fZ := []string{"x", "y", "z"}      // dissimilar

	sigX := mh.ComputeSignature(fX)
	sigY := mh.ComputeSignature(fY)
	sigZ := mh.ComputeSignature(fZ)

	lsh := NewLSHIndex(32, 4)
	if err := lsh.AddFragment(1, sigX); err != nil {
		t.Fatalf("add X: %v", err)
	}
	if err := lsh.AddFragment(2, sigY); err != nil {
		t.Fatalf("add Y: %v", err)
	}
	if err := lsh.AddFragment(3, sigZ); err != nil {
		t.Fatalf("add Z: %v", err)
	}
	if err := lsh.BuildIndex(); err != nil {
		t.Fatalf("build: %v", err)
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
	mh := NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := NewLSHIndex(32, 4).WithMaxCandidates(2)
	for _, id := range []int{1, 2, 3} {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %d: %v", id, err)
		}
	}

	cands := lsh.FindCandidates(sig)
	if len(cands) != 2 {
		t.Fatalf("expected oversized bucket candidates capped at 2; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesCapsTotalCandidates(t *testing.T) {
	mh := NewMinHasher(128)
	query := mh.ComputeSignature([]string{"shared", "query", "features"})

	lsh := NewLSHIndex(32, 4).WithMaxCandidates(2)
	keys := lsh.computeBandKeys(query)
	lsh.buckets[keys[0]] = []int{1}
	lsh.buckets[keys[1]] = []int{2}
	lsh.buckets[keys[2]] = []int{3}

	cands := lsh.FindCandidates(query)
	if len(cands) > 2 {
		t.Fatalf("expected candidates to be capped at 2; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesUsesDefaultCapAtBoundary(t *testing.T) {
	mh := NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := NewLSHIndex(32, 4)
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
	mh := NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := NewLSHIndex(32, 4)
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
