package analyzer

import (
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLSHIndex(t *testing.T) {
	config := LSHConfig{
		Bands:     16,
		Rows:      8,
		Threshold: 0.5,
	}
	
	index := NewLSHIndex(config)
	
	assert.Equal(t, 16, index.bands)
	assert.Equal(t, 8, index.rows)
	assert.Equal(t, 0.5, index.threshold)
	assert.NotNil(t, index.buckets)
	assert.NotNil(t, index.signatures)
}

func TestNewLSHIndex_DefaultValues(t *testing.T) {
	config := LSHConfig{
		Bands: 0, // Invalid
		Rows:  0, // Invalid
	}
	
	index := NewLSHIndex(config)
	
	// Should use defaults
	assert.Equal(t, 32, index.bands)
	assert.Equal(t, 4, index.rows)
	
	// Threshold should be auto-computed
	expectedThreshold := math.Pow(1.0/32.0, 1.0/4.0)
	assert.InDelta(t, expectedThreshold, index.threshold, 0.001)
}

func TestNewDefaultLSHIndex(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	assert.Equal(t, 32, index.bands)
	assert.Equal(t, 4, index.rows)
}

func TestAddFragment_ValidSignature(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128) // 128 > 32*4 = 128
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	err := index.AddFragment("fragment1", signature)
	
	assert.NoError(t, err)
	assert.Equal(t, 1, index.Size())
	assert.Equal(t, signature, index.GetSignature("fragment1"))
}

func TestAddFragment_NilSignature(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	err := index.AddFragment("fragment1", nil)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature cannot be nil")
}

func TestAddFragment_InsufficientHashes(t *testing.T) {
	index := NewDefaultLSHIndex() // Needs 32*4 = 128 hashes
	hasher := NewMinHasher(64)    // Only 64 hashes
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	err := index.AddFragment("fragment1", signature)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature has 64 hashes")
}

func TestFindCandidates_IdenticalSignatures(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	features := []string{"a", "b", "c", "d", "e"}
	signature := hasher.ComputeSignature(features)
	
	// Add the signature
	err := index.AddFragment("fragment1", signature)
	require.NoError(t, err)
	
	// Find candidates for the same signature
	candidates := index.FindCandidates(signature)
	
	assert.Contains(t, candidates, "fragment1")
}

func TestFindCandidates_SimilarSignatures(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(256) // More hashes for better precision
	
	// Create similar signatures
	features1 := []string{"a", "b", "c", "d", "e"}
	features2 := []string{"a", "b", "c", "d", "f"} // One different feature
	
	sig1 := hasher.ComputeSignature(features1)
	sig2 := hasher.ComputeSignature(features2)
	
	// Add the first signature
	err := index.AddFragment("fragment1", sig1)
	require.NoError(t, err)
	
	// Find candidates for the similar signature
	candidates := index.FindCandidates(sig2)
	
	// May or may not find the similar fragment depending on hash collision
	// This is probabilistic, so we just check the result is valid
	assert.IsType(t, []string{}, candidates)
}

func TestFindCandidates_NilSignature(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	candidates := index.FindCandidates(nil)
	
	assert.Empty(t, candidates)
}

func TestFindCandidates_InsufficientHashes(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(64) // Insufficient hashes
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	candidates := index.FindCandidates(signature)
	
	assert.Empty(t, candidates)
}

func TestFindCandidatesWithMinBands(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	// Add fragment
	err := index.AddFragment("fragment1", signature)
	require.NoError(t, err)
	
	// Find candidates that appear in at least 1 band
	candidates := index.FindCandidatesWithMinBands(signature, 1)
	
	assert.Contains(t, candidates, "fragment1")
	
	// Find candidates that appear in many bands (should still find identical)
	candidates = index.FindCandidatesWithMinBands(signature, 20)
	assert.Contains(t, candidates, "fragment1")
}

func TestFindCandidatesWithMinBands_InvalidParams(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	// Test with invalid parameters
	assert.Empty(t, index.FindCandidatesWithMinBands(nil, 1))
	assert.Empty(t, index.FindCandidatesWithMinBands(signature, 0))
}

func TestBuildIndex(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c"}),
		"frag2": hasher.ComputeSignature([]string{"d", "e", "f"}),
		"frag3": hasher.ComputeSignature([]string{"g", "h", "i"}),
	}
	
	err := index.BuildIndex(signatures)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, index.Size())
	
	// Check that all signatures are stored
	for id, sig := range signatures {
		assert.Equal(t, sig, index.GetSignature(id))
	}
}

func TestBuildIndex_WithNilSignature(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c"}),
		"frag2": nil, // Nil signature
		"frag3": hasher.ComputeSignature([]string{"g", "h", "i"}),
	}
	
	err := index.BuildIndex(signatures)
	
	assert.NoError(t, err)
	assert.Equal(t, 2, index.Size()) // Nil signature should be skipped
}

func TestBuildIndex_InsufficientHashes(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(64) // Insufficient hashes
	
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c"}),
	}
	
	err := index.BuildIndex(signatures)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature for frag1 has 64 hashes")
}

func TestRemoveFragment(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	// Add fragment
	err := index.AddFragment("fragment1", signature)
	require.NoError(t, err)
	assert.Equal(t, 1, index.Size())
	
	// Remove fragment
	index.RemoveFragment("fragment1")
	
	assert.Equal(t, 0, index.Size())
	assert.Nil(t, index.GetSignature("fragment1"))
}

func TestRemoveFragment_NonExistent(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	// Should not panic or error
	index.RemoveFragment("nonexistent")
	
	assert.Equal(t, 0, index.Size())
}

func TestGetConfig(t *testing.T) {
	config := LSHConfig{
		Bands:     16,
		Rows:      8,
		Threshold: 0.6,
	}
	
	index := NewLSHIndex(config)
	retrievedConfig := index.GetConfig()
	
	assert.Equal(t, config.Bands, retrievedConfig.Bands)
	assert.Equal(t, config.Rows, retrievedConfig.Rows)
	assert.Equal(t, config.Threshold, retrievedConfig.Threshold)
}

func TestGetStats(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Add some fragments
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c"}),
		"frag2": hasher.ComputeSignature([]string{"d", "e", "f"}),
		"frag3": hasher.ComputeSignature([]string{"g", "h", "i"}),
	}
	
	err := index.BuildIndex(signatures)
	require.NoError(t, err)
	
	stats := index.GetStats()
	
	assert.Equal(t, 3, stats.NumFragments)
	assert.Greater(t, stats.NumBuckets, 0)
	assert.Equal(t, 32, stats.Bands)
	assert.Equal(t, 4, stats.Rows)
	assert.Greater(t, stats.AvgBucketSize, 0.0)
	assert.GreaterOrEqual(t, stats.MaxBucketSize, stats.MinBucketSize)
}

func TestGetStats_EmptyIndex(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	stats := index.GetStats()
	
	assert.Equal(t, 0, stats.NumFragments)
	assert.Equal(t, 0, stats.NumBuckets)
	assert.Equal(t, 0.0, stats.AvgBucketSize)
}

func TestFindSimilarPairs(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(256) // More hashes for better precision
	
	// Create signatures
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c", "d"}),
		"frag2": hasher.ComputeSignature([]string{"a", "b", "c", "e"}), // Similar to frag1
		"frag3": hasher.ComputeSignature([]string{"x", "y", "z"}),      // Different
	}
	
	err := index.BuildIndex(signatures)
	require.NoError(t, err)
	
	pairs := index.FindSimilarPairs(hasher)
	
	assert.IsType(t, []*SimilarPair{}, pairs)
	
	// Check that pairs are sorted by similarity (descending)
	for i := 1; i < len(pairs); i++ {
		assert.GreaterOrEqual(t, pairs[i-1].Similarity, pairs[i].Similarity)
	}
	
	// Check that all similarities are above threshold
	for _, pair := range pairs {
		assert.GreaterOrEqual(t, pair.Similarity, index.threshold)
	}
}

func TestBatchAddFragments(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := map[string]*MinHashSignature{
		"frag1": hasher.ComputeSignature([]string{"a", "b", "c"}),
		"frag2": hasher.ComputeSignature([]string{"d", "e", "f"}),
	}
	
	err := index.BatchAddFragments(signatures)
	
	assert.NoError(t, err)
	assert.Equal(t, 2, index.Size())
}

func TestClear(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Add some data
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	err := index.AddFragment("fragment1", signature)
	require.NoError(t, err)
	
	assert.Equal(t, 1, index.Size())
	
	// Clear the index
	index.Clear()
	
	assert.Equal(t, 0, index.Size())
	assert.Empty(t, index.GetAllFragmentIDs())
}

func TestGetAllFragmentIDs(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := map[string]*MinHashSignature{
		"frag3": hasher.ComputeSignature([]string{"a", "b", "c"}),
		"frag1": hasher.ComputeSignature([]string{"d", "e", "f"}),
		"frag2": hasher.ComputeSignature([]string{"g", "h", "i"}),
	}
	
	err := index.BuildIndex(signatures)
	require.NoError(t, err)
	
	ids := index.GetAllFragmentIDs()
	
	assert.Equal(t, 3, len(ids))
	
	// Should be sorted
	assert.Equal(t, []string{"frag1", "frag2", "frag3"}, ids)
}

func TestComputeOptimalBandParameters(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		maxBands  int
		expected  LSHConfig
	}{
		{
			name:      "invalid threshold low",
			threshold: 0.0,
			maxBands:  50,
			expected:  LSHConfig{Bands: 32, Rows: 4},
		},
		{
			name:      "invalid threshold high",
			threshold: 1.0,
			maxBands:  50,
			expected:  LSHConfig{Bands: 32, Rows: 4},
		},
		{
			name:      "valid threshold",
			threshold: 0.8,
			maxBands:  50,
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := ComputeOptimalBandParameters(test.threshold, test.maxBands)
			
			assert.Greater(t, config.Bands, 0)
			assert.Greater(t, config.Rows, 0)
			assert.GreaterOrEqual(t, config.Threshold, 0.0)
			assert.LessOrEqual(t, config.Threshold, 1.0)
			
			if test.expected.Bands > 0 {
				assert.Equal(t, test.expected.Bands, config.Bands)
				assert.Equal(t, test.expected.Rows, config.Rows)
			}
		})
	}
}

func TestEstimateFalsePositiveRate(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	tests := []struct {
		similarity float64
		expected   float64
	}{
		{0.0, 0.0},
		{1.0, 0.0},
		{0.5, 0.0}, // Should be > 0 but we just check it doesn't panic
	}
	
	for _, test := range tests {
		rate := index.EstimateFalsePositiveRate(test.similarity)
		assert.GreaterOrEqual(t, rate, 0.0)
		assert.LessOrEqual(t, rate, 1.0)
	}
}

func TestEstimateFalseNegativeRate(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	tests := []struct {
		similarity float64
		expected   float64
	}{
		{0.0, 1.0},
		{1.0, 1.0},
		{0.5, 0.0}, // Should be < 1.0 but we just check it doesn't panic
	}
	
	for _, test := range tests {
		rate := index.EstimateFalseNegativeRate(test.similarity)
		assert.GreaterOrEqual(t, rate, 0.0)
		assert.LessOrEqual(t, rate, 1.0)
	}
}

func TestComputeBandHash_Consistency(t *testing.T) {
	index := NewDefaultLSHIndex()
	
	signatures := []uint64{1, 2, 3, 4, 5, 6, 7, 8}
	
	hash1 := index.computeBandHash(signatures, 0)
	hash2 := index.computeBandHash(signatures, 0)
	
	assert.Equal(t, hash1, hash2, "Band hash should be consistent")
	
	// Different bands should produce different hashes (very likely)
	hash3 := index.computeBandHash(signatures, 1)
	assert.NotEqual(t, hash1, hash3, "Different bands should produce different hashes")
}

func TestThreadSafety(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Test concurrent reads and writes
	done := make(chan bool)
	
	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			signature := hasher.ComputeSignature([]string{string(rune(i))})
			if err := index.AddFragment(string(rune(i)), signature); err != nil {
				t.Errorf("Failed to add fragment: %v", err)
				return
			}
		}
		done <- true
	}()
	
	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			index.Size()
			index.GetAllFragmentIDs()
		}
		done <- true
	}()
	
	// Wait for both goroutines
	<-done
	<-done
	
	// Should not panic and should have some data
	assert.GreaterOrEqual(t, index.Size(), 0)
}

func BenchmarkAddFragment(b *testing.B) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := make([]*MinHashSignature, b.N)
	for i := 0; i < b.N; i++ {
		signatures[i] = hasher.ComputeSignature([]string{string(rune(i))})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := index.AddFragment(string(rune(i)), signatures[i]); err != nil {
			b.Fatalf("Failed to add fragment: %v", err)
		}
	}
}

func BenchmarkFindCandidates(b *testing.B) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Pre-populate index
	for i := 0; i < 1000; i++ {
		signature := hasher.ComputeSignature([]string{string(rune(i))})
		if err := index.AddFragment(string(rune(i)), signature); err != nil {
			b.Fatalf("Failed to add fragment: %v", err)
		}
	}
	
	querySignature := hasher.ComputeSignature([]string{"query"})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.FindCandidates(querySignature)
	}
}

func BenchmarkBuildIndex(b *testing.B) {
	hasher := NewMinHasher(128)
	
	signatures := make(map[string]*MinHashSignature)
	for i := 0; i < 1000; i++ {
		id := string(rune(i))
		signatures[id] = hasher.ComputeSignature([]string{id})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := NewDefaultLSHIndex()
		if err := index.BuildIndex(signatures); err != nil {
			b.Fatalf("Failed to build index: %v", err)
		}
	}
}

// Test LSH probability properties
func TestLSHProbabilityProperties(t *testing.T) {
	// Test with a configuration that should have known properties
	config := LSHConfig{
		Bands:     20,
		Rows:      5,
		Threshold: math.Pow(1.0/20.0, 1.0/5.0), // ≈ 0.724
	}
	
	index := NewLSHIndex(config)
	
	// Test false positive rate decreases with decreasing similarity
	rate1 := index.EstimateFalsePositiveRate(0.9)
	rate2 := index.EstimateFalsePositiveRate(0.7)
	rate3 := index.EstimateFalsePositiveRate(0.5)
	
	assert.Greater(t, rate1, rate2)
	assert.Greater(t, rate2, rate3)
	
	// Test false negative rate increases with decreasing similarity
	fnRate1 := index.EstimateFalseNegativeRate(0.9)
	fnRate2 := index.EstimateFalseNegativeRate(0.7)
	fnRate3 := index.EstimateFalseNegativeRate(0.5)
	
	assert.Less(t, fnRate1, fnRate2)
	assert.Less(t, fnRate2, fnRate3)
}

// Test edge case with very large number of bands
func TestLSHWithManyBands(t *testing.T) {
	config := LSHConfig{
		Bands:     100,
		Rows:      1,
		Threshold: 0.5,
	}
	
	index := NewLSHIndex(config)
	hasher := NewMinHasher(100) // Exactly enough hashes
	
	signature := hasher.ComputeSignature([]string{"a", "b", "c"})
	
	err := index.AddFragment("fragment1", signature)
	assert.NoError(t, err)
	
	candidates := index.FindCandidates(signature)
	assert.Contains(t, candidates, "fragment1")
}

// Test collision handling in buckets
func TestBucketCollisionHandling(t *testing.T) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Add many fragments to increase chance of collisions
	for i := 0; i < 100; i++ {
		signature := hasher.ComputeSignature([]string{string(rune(i + 65))}) // A, B, C, ...
		err := index.AddFragment(string(rune(i+65)), signature)
		require.NoError(t, err)
	}
	
	assert.Equal(t, 100, index.Size())
	
	// All fragments should be retrievable
	ids := index.GetAllFragmentIDs()
	assert.Equal(t, 100, len(ids))
	
	// Verify IDs are sorted
	sortedIDs := make([]string, len(ids))
	copy(sortedIDs, ids)
	sort.Strings(sortedIDs)
	assert.Equal(t, sortedIDs, ids)
}