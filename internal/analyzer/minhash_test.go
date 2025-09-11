package analyzer

import (
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMinHashSignature(t *testing.T) {
	sig := NewMinHashSignature(128)
	
	assert.Equal(t, 128, sig.GetNumHashes())
	assert.Equal(t, 128, len(sig.GetSignatures()))
}

func TestNewMinHasher(t *testing.T) {
	hasher := NewMinHasher(64)
	
	assert.Equal(t, 64, hasher.numHashes)
	assert.Equal(t, 64, len(hasher.hashFunctions))
	assert.Equal(t, uint64(2147483647), hasher.prime)
}

func TestNewMinHasher_DefaultSize(t *testing.T) {
	hasher := NewMinHasher(0) // Should use default
	
	assert.Equal(t, 128, hasher.numHashes)
	assert.Equal(t, 128, len(hasher.hashFunctions))
}

func TestNewMinHasherWithSeed(t *testing.T) {
	hasher1 := NewMinHasherWithSeed(64, 42)
	hasher2 := NewMinHasherWithSeed(64, 42)
	hasher3 := NewMinHasherWithSeed(64, 43)
	
	// Same seed should produce same hash functions
	for i := 0; i < 64; i++ {
		assert.Equal(t, hasher1.hashFunctions[i], hasher2.hashFunctions[i])
		// Different seed should produce different hash functions (very likely)
		if hasher1.hashFunctions[i] != hasher3.hashFunctions[i] {
			return // At least one should be different
		}
	}
	t.Error("Different seeds produced identical hash functions (very unlikely)")
}

func TestComputeSignature_EmptySet(t *testing.T) {
	hasher := NewMinHasher(64)
	
	signature := hasher.ComputeSignature([]string{})
	
	// All signature values should be max uint64 (no features to minimize)
	for _, sig := range signature.GetSignatures() {
		assert.Equal(t, uint64(math.MaxUint64), sig)
	}
}

func TestComputeSignature_SingleFeature(t *testing.T) {
	hasher := NewMinHasher(64)
	features := []string{"feature1"}
	
	signature := hasher.ComputeSignature(features)
	
	// Should have valid signature values (not all max)
	maxCount := 0
	for _, sig := range signature.GetSignatures() {
		if sig == math.MaxUint64 {
			maxCount++
		}
	}
	assert.Less(t, maxCount, 64, "Not all signature values should be max")
}

func TestComputeSignature_MultipleFeatures(t *testing.T) {
	hasher := NewMinHasher(128)
	features := []string{"feature1", "feature2", "feature3", "feature4", "feature5"}
	
	signature := hasher.ComputeSignature(features)
	
	// Check that signature is properly computed
	assert.Equal(t, 128, signature.GetNumHashes())
	
	// No signature value should be max uint64 with multiple features
	for _, sig := range signature.GetSignatures() {
		assert.Less(t, sig, uint64(math.MaxUint64))
	}
}

func TestComputeSignature_Deterministic(t *testing.T) {
	hasher := NewMinHasherWithSeed(64, 42)
	features := []string{"feature1", "feature2", "feature3"}
	
	sig1 := hasher.ComputeSignature(features)
	sig2 := hasher.ComputeSignature(features)
	
	// Should produce identical signatures
	assert.Equal(t, sig1.GetSignatures(), sig2.GetSignatures())
}

func TestEstimateJaccardSimilarity_IdenticalSignatures(t *testing.T) {
	hasher := NewMinHasher(128)
	features := []string{"feature1", "feature2", "feature3"}
	
	sig1 := hasher.ComputeSignature(features)
	sig2 := hasher.ComputeSignature(features)
	
	similarity := hasher.EstimateJaccardSimilarity(sig1, sig2)
	
	assert.Equal(t, 1.0, similarity)
}

func TestEstimateJaccardSimilarity_DifferentSignatures(t *testing.T) {
	hasher := NewMinHasher(128)
	
	features1 := []string{"feature1", "feature2", "feature3"}
	features2 := []string{"feature4", "feature5", "feature6"}
	
	sig1 := hasher.ComputeSignature(features1)
	sig2 := hasher.ComputeSignature(features2)
	
	similarity := hasher.EstimateJaccardSimilarity(sig1, sig2)
	
	assert.GreaterOrEqual(t, similarity, 0.0)
	assert.LessOrEqual(t, similarity, 1.0)
	assert.Less(t, similarity, 1.0) // Should not be identical
}

func TestEstimateJaccardSimilarity_PartialOverlap(t *testing.T) {
	hasher := NewMinHasher(256) // More hashes for better estimation
	
	features1 := []string{"a", "b", "c", "d"}
	features2 := []string{"c", "d", "e", "f"}
	
	sig1 := hasher.ComputeSignature(features1)
	sig2 := hasher.ComputeSignature(features2)
	
	similarity := hasher.EstimateJaccardSimilarity(sig1, sig2)
	
	// Expected Jaccard: |{c,d}| / |{a,b,c,d,e,f}| = 2/6 = 0.333...
	expectedSimilarity := 2.0 / 6.0
	
	// Allow for estimation error (±0.2)
	assert.InDelta(t, expectedSimilarity, similarity, 0.2)
}

func TestEstimateJaccardSimilarity_NilSignatures(t *testing.T) {
	hasher := NewMinHasher(64)
	
	assert.Equal(t, 0.0, hasher.EstimateJaccardSimilarity(nil, nil))
	
	sig := hasher.ComputeSignature([]string{"test"})
	assert.Equal(t, 0.0, hasher.EstimateJaccardSimilarity(sig, nil))
	assert.Equal(t, 0.0, hasher.EstimateJaccardSimilarity(nil, sig))
}

func TestEstimateJaccardSimilarity_IncompatibleSignatures(t *testing.T) {
	hasher1 := NewMinHasher(64)
	hasher2 := NewMinHasher(128)
	
	features := []string{"feature1", "feature2"}
	sig1 := hasher1.ComputeSignature(features)
	sig2 := hasher2.ComputeSignature(features)
	
	similarity := hasher1.EstimateJaccardSimilarity(sig1, sig2)
	
	assert.Equal(t, 0.0, similarity)
}

func TestComputeJaccardSimilarity_ExactCalculation(t *testing.T) {
	tests := []struct {
		name      string
		features1 []string
		features2 []string
		expected  float64
	}{
		{
			name:      "identical sets",
			features1: []string{"a", "b", "c"},
			features2: []string{"a", "b", "c"},
			expected:  1.0,
		},
		{
			name:      "no overlap",
			features1: []string{"a", "b", "c"},
			features2: []string{"d", "e", "f"},
			expected:  0.0,
		},
		{
			name:      "partial overlap",
			features1: []string{"a", "b", "c"},
			features2: []string{"b", "c", "d"},
			expected:  2.0 / 4.0, // {b,c} / {a,b,c,d}
		},
		{
			name:      "subset",
			features1: []string{"a", "b"},
			features2: []string{"a", "b", "c"},
			expected:  2.0 / 3.0, // {a,b} / {a,b,c}
		},
		{
			name:      "empty sets",
			features1: []string{},
			features2: []string{},
			expected:  1.0,
		},
		{
			name:      "one empty",
			features1: []string{"a"},
			features2: []string{},
			expected:  0.0,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			similarity := ComputeJaccardSimilarity(test.features1, test.features2)
			assert.InDelta(t, test.expected, similarity, 0.001)
		})
	}
}

func TestComputeSignatureStats(t *testing.T) {
	hasher := NewMinHasher(10)
	features := []string{"a", "b", "c", "d", "e"}
	signature := hasher.ComputeSignature(features)
	
	stats := ComputeSignatureStats(signature)
	
	assert.Greater(t, stats.Mean, 0.0)
	assert.GreaterOrEqual(t, stats.Variance, 0.0)
	assert.LessOrEqual(t, stats.Min, stats.Max)
	assert.Equal(t, 10, len(signature.GetSignatures()))
}

func TestComputeSignatureStats_NilSignature(t *testing.T) {
	stats := ComputeSignatureStats(nil)
	
	assert.Equal(t, SignatureStats{}, stats)
}

func TestCompareSignatures(t *testing.T) {
	hasher := NewMinHasher(128)
	
	features1 := []string{"a", "b", "c"}
	features2 := []string{"b", "c", "d"}
	
	sig1 := hasher.ComputeSignature(features1)
	sig2 := hasher.ComputeSignature(features2)
	
	comparison := hasher.CompareSignatures(sig1, sig2)
	
	assert.GreaterOrEqual(t, comparison.Similarity, 0.0)
	assert.LessOrEqual(t, comparison.Similarity, 1.0)
	assert.Equal(t, 128, comparison.TotalHashes)
	assert.GreaterOrEqual(t, comparison.MatchingHashes, 0)
	assert.LessOrEqual(t, comparison.MatchingHashes, 128)
}

func TestSignatureBatch(t *testing.T) {
	batch := NewSignatureBatch()
	hasher := NewMinHasher(64)
	
	// Add some signatures
	sig1 := hasher.ComputeSignature([]string{"a", "b"})
	sig2 := hasher.ComputeSignature([]string{"c", "d"})
	
	batch.AddSignature("id1", sig1)
	batch.AddSignature("id2", sig2)
	
	assert.Equal(t, 2, batch.Size())
	assert.Equal(t, sig1, batch.GetSignature("id1"))
	assert.Equal(t, sig2, batch.GetSignature("id2"))
	
	ids := batch.GetAllIDs()
	sort.Strings(ids) // Sort for consistent comparison
	assert.Equal(t, []string{"id1", "id2"}, ids)
}

func TestSignatureBatch_DuplicateID(t *testing.T) {
	batch := NewSignatureBatch()
	hasher := NewMinHasher(64)
	
	sig1 := hasher.ComputeSignature([]string{"a", "b"})
	sig2 := hasher.ComputeSignature([]string{"c", "d"})
	
	batch.AddSignature("id1", sig1)
	batch.AddSignature("id1", sig2) // Same ID
	
	assert.Equal(t, 1, batch.Size())
	assert.Equal(t, sig2, batch.GetSignature("id1")) // Should be updated
}

func TestComputeAllPairwiseSimilarities(t *testing.T) {
	hasher := NewMinHasher(128)
	batch := NewSignatureBatch()
	
	// Add signatures
	batch.AddSignature("id1", hasher.ComputeSignature([]string{"a", "b", "c"}))
	batch.AddSignature("id2", hasher.ComputeSignature([]string{"b", "c", "d"}))
	batch.AddSignature("id3", hasher.ComputeSignature([]string{"a", "b", "c"})) // Same as id1
	
	similarities := hasher.ComputeAllPairwiseSimilarities(batch)
	
	assert.Equal(t, 3, len(similarities))
	
	// Check self-similarities
	assert.Equal(t, 1.0, similarities["id1"]["id1"])
	assert.Equal(t, 1.0, similarities["id2"]["id2"])
	assert.Equal(t, 1.0, similarities["id3"]["id3"])
	
	// Check symmetry
	assert.Equal(t, similarities["id1"]["id2"], similarities["id2"]["id1"])
	assert.Equal(t, similarities["id1"]["id3"], similarities["id3"]["id1"])
	assert.Equal(t, similarities["id2"]["id3"], similarities["id3"]["id2"])
	
	// id1 and id3 should be identical
	assert.Equal(t, 1.0, similarities["id1"]["id3"])
}

func TestFindSimilarSignatures(t *testing.T) {
	hasher := NewMinHasher(256) // More hashes for better precision
	batch := NewSignatureBatch()
	
	queryFeatures := []string{"a", "b", "c"}
	querySignature := hasher.ComputeSignature(queryFeatures)
	
	// Add signatures with different similarities
	batch.AddSignature("identical", hasher.ComputeSignature([]string{"a", "b", "c"}))
	batch.AddSignature("similar", hasher.ComputeSignature([]string{"a", "b", "d"}))
	batch.AddSignature("different", hasher.ComputeSignature([]string{"x", "y", "z"}))
	
	similar := hasher.FindSimilarSignatures(querySignature, batch, 0.5)
	
	// Should find at least the identical one
	assert.Contains(t, similar, "identical")
	// Might or might not find "similar" depending on hash variance
	// Should not find "different" (very unlikely with 256 hashes)
}

func TestValidateAccuracy(t *testing.T) {
	hasher := NewMinHasher(256)
	
	featureSets := map[string][]string{
		"set1": {"a", "b", "c", "d"},
		"set2": {"b", "c", "d", "e"},
		"set3": {"c", "d", "e", "f"},
		"set4": {"x", "y", "z"},
	}
	
	validation := hasher.ValidateAccuracy(featureSets)
	
	assert.Greater(t, validation.NumComparisons, 0)
	assert.GreaterOrEqual(t, validation.MeanError, 0.0)
	assert.LessOrEqual(t, validation.MeanError, 1.0)
	assert.GreaterOrEqual(t, validation.MaxError, validation.MeanError)
	assert.GreaterOrEqual(t, validation.CorrectDirection, 0)
}

func TestHashFunctionProperties(t *testing.T) {
	hasher := NewMinHasherWithSeed(64, 42)
	
	// All hash functions should have valid parameters
	for i, hashFunc := range hasher.hashFunctions {
		assert.Greater(t, hashFunc.a, uint64(0), "Hash function %d should have a > 0", i)
		assert.Less(t, hashFunc.a, hasher.prime, "Hash function %d should have a < prime", i)
		assert.Less(t, hashFunc.b, hasher.prime, "Hash function %d should have b < prime", i)
		assert.Equal(t, hasher.prime, hashFunc.p, "Hash function %d should have correct prime", i)
	}
}

func TestApplyHashFunction(t *testing.T) {
	hasher := NewMinHasher(1)
	baseHash := uint64(12345)
	
	// Test with the first hash function
	hashFunc := hasher.hashFunctions[0]
	result := hasher.applyHashFunction(baseHash, hashFunc)
	
	// Result should be within expected range
	assert.Less(t, result, hasher.prime)
	
	// Same input should give same output
	result2 := hasher.applyHashFunction(baseHash, hashFunc)
	assert.Equal(t, result, result2)
}

func TestBaseHashConsistency(t *testing.T) {
	hasher := NewMinHasher(64)
	
	feature := "test_feature"
	hash1 := hasher.computeBaseHash(feature)
	hash2 := hasher.computeBaseHash(feature)
	
	assert.Equal(t, hash1, hash2, "Base hash should be consistent")
	
	// Different features should have different hashes (very likely)
	hash3 := hasher.computeBaseHash("different_feature")
	assert.NotEqual(t, hash1, hash3, "Different features should have different hashes")
}

func BenchmarkComputeSignature(b *testing.B) {
	hasher := NewMinHasher(128)
	features := []string{
		"feature1", "feature2", "feature3", "feature4", "feature5",
		"feature6", "feature7", "feature8", "feature9", "feature10",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.ComputeSignature(features)
	}
}

func BenchmarkEstimateJaccardSimilarity(b *testing.B) {
	hasher := NewMinHasher(128)
	
	sig1 := hasher.ComputeSignature([]string{"a", "b", "c", "d", "e"})
	sig2 := hasher.ComputeSignature([]string{"c", "d", "e", "f", "g"})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.EstimateJaccardSimilarity(sig1, sig2)
	}
}

func BenchmarkComputeJaccardSimilarity(b *testing.B) {
	features1 := []string{"a", "b", "c", "d", "e"}
	features2 := []string{"c", "d", "e", "f", "g"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeJaccardSimilarity(features1, features2)
	}
}

// Test MinHash accuracy with various set sizes and overlaps
func TestMinHashAccuracy(t *testing.T) {
	hasher := NewMinHasher(512) // Large number of hashes for better accuracy
	
	testCases := []struct {
		name       string
		set1       []string
		set2       []string
		tolerance  float64
	}{
		{
			name:      "high similarity",
			set1:      []string{"a", "b", "c", "d", "e"},
			set2:      []string{"a", "b", "c", "d", "f"},
			tolerance: 0.15,
		},
		{
			name:      "medium similarity",
			set1:      []string{"a", "b", "c", "d", "e", "f"},
			set2:      []string{"c", "d", "e", "f", "g", "h"},
			tolerance: 0.15,
		},
		{
			name:      "low similarity",
			set1:      []string{"a", "b", "c", "d", "e", "f", "g", "h"},
			set2:      []string{"g", "h", "i", "j", "k", "l", "m", "n"},
			tolerance: 0.2,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			exactSimilarity := ComputeJaccardSimilarity(tc.set1, tc.set2)
			
			sig1 := hasher.ComputeSignature(tc.set1)
			sig2 := hasher.ComputeSignature(tc.set2)
			estimatedSimilarity := hasher.EstimateJaccardSimilarity(sig1, sig2)
			
			assert.InDelta(t, exactSimilarity, estimatedSimilarity, tc.tolerance,
				"MinHash estimate should be close to exact Jaccard similarity")
		})
	}
}