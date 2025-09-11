package analyzer

import (
	"crypto/md5"
	"fmt"
	"math"
	"sort"
	"sync"
)

// LSHIndex implements Locality Sensitive Hashing with banding technique
type LSHIndex struct {
	bands      int                            // Number of bands
	rows       int                            // Rows per band
	buckets    map[string][]string            // band_hash -> fragment_ids
	signatures map[string]*MinHashSignature   // fragment_id -> signature
	threshold  float64                        // Similarity threshold for candidates
	mutex      sync.RWMutex                   // Thread safety for concurrent access
}

// LSHConfig holds configuration parameters for LSH
type LSHConfig struct {
	Bands     int     // Number of bands (default: 32)
	Rows      int     // Rows per band (default: 4)
	Threshold float64 // Similarity threshold (default: ~0.78)
}

// NewLSHIndex creates a new LSH index with the given configuration
func NewLSHIndex(config LSHConfig) *LSHIndex {
	// Validate and set defaults
	if config.Bands <= 0 {
		config.Bands = 32
	}
	if config.Rows <= 0 {
		config.Rows = 4
	}
	if config.Threshold <= 0 || config.Threshold > 1 {
		// Compute threshold based on bands and rows: threshold ≈ (1/b)^(1/r)
		config.Threshold = math.Pow(1.0/float64(config.Bands), 1.0/float64(config.Rows))
	}

	return &LSHIndex{
		bands:      config.Bands,
		rows:       config.Rows,
		buckets:    make(map[string][]string),
		signatures: make(map[string]*MinHashSignature),
		threshold:  config.Threshold,
	}
}

// NewDefaultLSHIndex creates an LSH index with default parameters
func NewDefaultLSHIndex() *LSHIndex {
	return NewLSHIndex(LSHConfig{
		Bands: 32,
		Rows:  4,
	})
}

// AddFragment adds a fragment with its signature to the index
func (idx *LSHIndex) AddFragment(id string, signature *MinHashSignature) error {
	if signature == nil {
		return fmt.Errorf("signature cannot be nil")
	}

	if signature.GetNumHashes() < idx.bands*idx.rows {
		return fmt.Errorf("signature has %d hashes, but need at least %d (bands=%d, rows=%d)",
			signature.GetNumHashes(), idx.bands*idx.rows, idx.bands, idx.rows)
	}

	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Store the signature
	idx.signatures[id] = signature

	// Add to LSH buckets
	return idx.addToBuckets(id, signature)
}

// addToBuckets adds a fragment to the appropriate LSH buckets
func (idx *LSHIndex) addToBuckets(id string, signature *MinHashSignature) error {
	sigs := signature.GetSignatures()

	// Process each band
	for band := 0; band < idx.bands; band++ {
		// Extract rows for this band
		bandHash := idx.computeBandHash(sigs, band)
		
		// Add fragment to bucket
		bucketKey := fmt.Sprintf("band_%d_%s", band, bandHash)
		idx.buckets[bucketKey] = append(idx.buckets[bucketKey], id)
	}

	return nil
}

// computeBandHash computes a hash for a specific band
func (idx *LSHIndex) computeBandHash(signatures []uint64, bandIndex int) string {
	startIdx := bandIndex * idx.rows
	endIdx := startIdx + idx.rows

	// Create a string representation of the band
	bandData := make([]byte, 0, idx.rows*8)
	for i := startIdx; i < endIdx && i < len(signatures); i++ {
		// Convert uint64 to bytes
		sig := signatures[i]
		for j := 0; j < 8; j++ {
			bandData = append(bandData, byte(sig>>(j*8)))
		}
	}

	// Compute MD5 hash of the band
	hash := md5.Sum(bandData)
	return fmt.Sprintf("%x", hash)
}

// FindCandidates finds candidate fragments similar to the query signature
func (idx *LSHIndex) FindCandidates(querySignature *MinHashSignature) []string {
	if querySignature == nil {
		return []string{}
	}

	if querySignature.GetNumHashes() < idx.bands*idx.rows {
		return []string{}
	}

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	candidateSet := make(map[string]bool)
	sigs := querySignature.GetSignatures()

	// Check each band
	for band := 0; band < idx.bands; band++ {
		bandHash := idx.computeBandHash(sigs, band)
		bucketKey := fmt.Sprintf("band_%d_%s", band, bandHash)

		// Get all fragments in this bucket
		if fragments, exists := idx.buckets[bucketKey]; exists {
			for _, fragmentID := range fragments {
				candidateSet[fragmentID] = true
			}
		}
	}

	// Convert set to slice and sort for consistent ordering
	candidates := make([]string, 0, len(candidateSet))
	for candidate := range candidateSet {
		candidates = append(candidates, candidate)
	}
	sort.Strings(candidates)

	return candidates
}

// FindCandidatesWithMinBands finds candidates that appear in at least minBands bands
func (idx *LSHIndex) FindCandidatesWithMinBands(querySignature *MinHashSignature, minBands int) []string {
	if querySignature == nil || minBands <= 0 {
		return []string{}
	}

	if querySignature.GetNumHashes() < idx.bands*idx.rows {
		return []string{}
	}

	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	candidateCounts := make(map[string]int)
	sigs := querySignature.GetSignatures()

	// Check each band
	for band := 0; band < idx.bands; band++ {
		bandHash := idx.computeBandHash(sigs, band)
		bucketKey := fmt.Sprintf("band_%d_%s", band, bandHash)

		// Get all fragments in this bucket
		if fragments, exists := idx.buckets[bucketKey]; exists {
			for _, fragmentID := range fragments {
				candidateCounts[fragmentID]++
			}
		}
	}

	// Filter candidates that appear in at least minBands bands
	var candidates []string
	for candidate, count := range candidateCounts {
		if count >= minBands {
			candidates = append(candidates, candidate)
		}
	}

	sort.Strings(candidates)
	return candidates
}

// BuildIndex builds the index from a batch of signatures
func (idx *LSHIndex) BuildIndex(signatures map[string]*MinHashSignature) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Clear existing index
	idx.buckets = make(map[string][]string)
	idx.signatures = make(map[string]*MinHashSignature)

	// Add all signatures
	for id, signature := range signatures {
		if signature == nil {
			continue
		}

		if signature.GetNumHashes() < idx.bands*idx.rows {
			return fmt.Errorf("signature for %s has %d hashes, but need at least %d",
				id, signature.GetNumHashes(), idx.bands*idx.rows)
		}

		// Store signature
		idx.signatures[id] = signature

		// Add to buckets
		if err := idx.addToBuckets(id, signature); err != nil {
			return fmt.Errorf("failed to add fragment %s to buckets: %v", id, err)
		}
	}

	return nil
}

// RemoveFragment removes a fragment from the index
func (idx *LSHIndex) RemoveFragment(id string) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	signature, exists := idx.signatures[id]
	if !exists {
		return
	}

	// Remove from signatures
	delete(idx.signatures, id)

	// Remove from buckets
	idx.removeFromBuckets(id, signature)
}

// removeFromBuckets removes a fragment from all its buckets
func (idx *LSHIndex) removeFromBuckets(id string, signature *MinHashSignature) {
	sigs := signature.GetSignatures()

	// Process each band
	for band := 0; band < idx.bands; band++ {
		bandHash := idx.computeBandHash(sigs, band)
		bucketKey := fmt.Sprintf("band_%d_%s", band, bandHash)

		// Remove from bucket
		if fragments, exists := idx.buckets[bucketKey]; exists {
			// Find and remove the fragment
			for i, fragmentID := range fragments {
				if fragmentID == id {
					// Remove by swapping with last element and truncating
					fragments[i] = fragments[len(fragments)-1]
					idx.buckets[bucketKey] = fragments[:len(fragments)-1]
					break
				}
			}

			// Remove bucket if empty
			if len(idx.buckets[bucketKey]) == 0 {
				delete(idx.buckets, bucketKey)
			}
		}
	}
}

// GetSignature retrieves the stored signature for a fragment
func (idx *LSHIndex) GetSignature(id string) *MinHashSignature {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return idx.signatures[id]
}

// Size returns the number of fragments in the index
func (idx *LSHIndex) Size() int {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return len(idx.signatures)
}

// GetConfig returns the LSH configuration
func (idx *LSHIndex) GetConfig() LSHConfig {
	return LSHConfig{
		Bands:     idx.bands,
		Rows:      idx.rows,
		Threshold: idx.threshold,
	}
}

// GetStats returns statistics about the index
func (idx *LSHIndex) GetStats() LSHIndexStats {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	stats := LSHIndexStats{
		NumFragments: len(idx.signatures),
		NumBuckets:   len(idx.buckets),
		Bands:        idx.bands,
		Rows:         idx.rows,
		Threshold:    idx.threshold,
	}

	// Compute bucket size statistics
	if len(idx.buckets) > 0 {
		bucketSizes := make([]int, 0, len(idx.buckets))
		totalFragments := 0

		for _, fragments := range idx.buckets {
			size := len(fragments)
			bucketSizes = append(bucketSizes, size)
			totalFragments += size
		}

		sort.Ints(bucketSizes)

		stats.MinBucketSize = bucketSizes[0]
		stats.MaxBucketSize = bucketSizes[len(bucketSizes)-1]
		stats.AvgBucketSize = float64(totalFragments) / float64(len(idx.buckets))

		// Median bucket size
		if len(bucketSizes)%2 == 0 {
			mid := len(bucketSizes) / 2
			stats.MedianBucketSize = float64(bucketSizes[mid-1]+bucketSizes[mid]) / 2.0
		} else {
			stats.MedianBucketSize = float64(bucketSizes[len(bucketSizes)/2])
		}
	}

	return stats
}

// LSHIndexStats provides statistics about the LSH index
type LSHIndexStats struct {
	NumFragments     int     // Number of fragments indexed
	NumBuckets       int     // Number of hash buckets
	Bands            int     // Number of bands
	Rows             int     // Rows per band
	Threshold        float64 // Similarity threshold
	MinBucketSize    int     // Minimum bucket size
	MaxBucketSize    int     // Maximum bucket size
	AvgBucketSize    float64 // Average bucket size
	MedianBucketSize float64 // Median bucket size
}

// FindSimilarPairs finds all similar pairs using LSH
func (idx *LSHIndex) FindSimilarPairs(minHasher *MinHasher) []*SimilarPair {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	var pairs []*SimilarPair
	processed := make(map[string]bool)

	// Get all fragment IDs
	var fragmentIDs []string
	for id := range idx.signatures {
		fragmentIDs = append(fragmentIDs, id)
	}

	// For each fragment, find its candidates
	for _, id := range fragmentIDs {
		if processed[id] {
			continue
		}

		signature := idx.signatures[id]
		candidates := idx.FindCandidates(signature)

		// Compare with candidates
		for _, candidateID := range candidates {
			if candidateID <= id || processed[candidateID] {
				continue // Avoid duplicates and self-comparison
			}

			candidateSignature := idx.signatures[candidateID]
			if candidateSignature != nil {
				similarity := minHasher.EstimateJaccardSimilarity(signature, candidateSignature)
				if similarity >= idx.threshold {
					pairs = append(pairs, &SimilarPair{
						ID1:        id,
						ID2:        candidateID,
						Similarity: similarity,
					})
				}
			}
		}

		processed[id] = true
	}

	// Sort pairs by similarity (descending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Similarity > pairs[j].Similarity
	})

	return pairs
}

// SimilarPair represents a pair of similar fragments
type SimilarPair struct {
	ID1        string  // First fragment ID
	ID2        string  // Second fragment ID
	Similarity float64 // Estimated similarity
}

// BatchAddFragments adds multiple fragments efficiently
func (idx *LSHIndex) BatchAddFragments(signatures map[string]*MinHashSignature) error {
	return idx.BuildIndex(signatures)
}

// Clear clears all data from the index
func (idx *LSHIndex) Clear() {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	idx.buckets = make(map[string][]string)
	idx.signatures = make(map[string]*MinHashSignature)
}

// GetAllFragmentIDs returns all fragment IDs in the index
func (idx *LSHIndex) GetAllFragmentIDs() []string {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	ids := make([]string, 0, len(idx.signatures))
	for id := range idx.signatures {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// ComputeOptimalBandParameters computes optimal band parameters for a target threshold
func ComputeOptimalBandParameters(targetThreshold float64, maxBands int) LSHConfig {
	if targetThreshold <= 0 || targetThreshold >= 1 {
		return LSHConfig{Bands: 32, Rows: 4, Threshold: 0.78}
	}

	bestConfig := LSHConfig{Bands: 32, Rows: 4}
	bestError := math.Inf(1)

	// Try different combinations of bands and rows
	for bands := 1; bands <= maxBands; bands++ {
		for rows := 1; rows <= 8; rows++ {
			// Compute threshold for this configuration
			threshold := math.Pow(1.0/float64(bands), 1.0/float64(rows))
			error := math.Abs(threshold - targetThreshold)

			if error < bestError {
				bestError = error
				bestConfig = LSHConfig{
					Bands:     bands,
					Rows:      rows,
					Threshold: threshold,
				}
			}
		}
	}

	return bestConfig
}

// EstimateFalsePositiveRate estimates the false positive rate for the current configuration
func (idx *LSHIndex) EstimateFalsePositiveRate(trueSimilarity float64) float64 {
	if trueSimilarity <= 0 || trueSimilarity >= 1 {
		return 0.0
	}

	// Probability that at least one band matches
	// P(false positive) = 1 - (1 - s^r)^b
	// where s is true similarity, r is rows per band, b is number of bands
	probBandMatches := math.Pow(trueSimilarity, float64(idx.rows))
	probNoMatches := math.Pow(1.0-probBandMatches, float64(idx.bands))
	return 1.0 - probNoMatches
}

// EstimateFalseNegativeRate estimates the false negative rate for the current configuration
func (idx *LSHIndex) EstimateFalseNegativeRate(trueSimilarity float64) float64 {
	if trueSimilarity <= 0 || trueSimilarity >= 1 {
		return 1.0
	}

	// Probability that no bands match
	// P(false negative) = (1 - s^r)^b
	probBandMatches := math.Pow(trueSimilarity, float64(idx.rows))
	return math.Pow(1.0-probBandMatches, float64(idx.bands))
}