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
	if err := lsh.AddFragment("X", sigX); err != nil {
		t.Fatalf("add X: %v", err)
	}
	if err := lsh.AddFragment("Y", sigY); err != nil {
		t.Fatalf("add Y: %v", err)
	}
	if err := lsh.AddFragment("Z", sigZ); err != nil {
		t.Fatalf("add Z: %v", err)
	}
	if err := lsh.BuildIndex(); err != nil {
		t.Fatalf("build: %v", err)
	}

	cands := lsh.FindCandidates(sigX)
	foundY := false
	for _, id := range cands {
		if id == "Y" {
			foundY = true
			break
		}
	}
	if !foundY {
		t.Fatalf("expected Y to be a candidate for X; got %v", cands)
	}
}

func TestLSHIndex_FindCandidatesSkipsOversizedBuckets(t *testing.T) {
	mh := NewMinHasher(128)
	sig := mh.ComputeSignature([]string{"same", "feature", "set"})

	lsh := NewLSHIndex(32, 4).WithMaxBucketSize(2)
	for _, id := range []string{"A", "B", "C"} {
		if err := lsh.AddFragment(id, sig); err != nil {
			t.Fatalf("add %s: %v", id, err)
		}
	}

	cands := lsh.FindCandidates(sig)
	if len(cands) != 0 {
		t.Fatalf("expected oversized buckets to be skipped; got %v", cands)
	}
}
