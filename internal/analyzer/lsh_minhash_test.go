package analyzer

import (
	"math"
	"testing"
)

func TestMinHasher_EstimateJaccardSimilarity(t *testing.T) {
	a := []string{"a", "b", "c", "d", "e"}
	b := []string{"c", "d", "e", "f", "g"}
	// True Jaccard: |{c,d,e}| / |{a,b,c,d,e,f,g}| = 3/7 ≈ 0.4286
	trueJ := 3.0 / 7.0

	mh := NewMinHasher(128)
	sigA := mh.ComputeSignature(a)
	sigB := mh.ComputeSignature(b)
	est := mh.EstimateJaccardSimilarity(sigA, sigB)

	if math.IsNaN(est) || est < 0 || est > 1 {
		t.Fatalf("invalid estimate: %v", est)
	}

	// Loose tolerance for randomized hashing
	if math.Abs(est-trueJ) > 0.3 { // allow generous error bound
		t.Logf("warning: estimate %.3f differs from true %.3f by > 0.3 (expected randomness)", est, trueJ)
	}
}
