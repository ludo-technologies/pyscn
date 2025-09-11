package analyzer

import (
	"hash/fnv"
	"math"
	"math/rand"
	"sort"
	"time"
)

// MinHashSignature represents a MinHash signature for a set of features
type MinHashSignature struct {
	signatures []uint64 // Array of hash values
	numHashes  int      // Number of hash functions used
}

// NewMinHashSignature creates a new MinHash signature
func NewMinHashSignature(numHashes int) *MinHashSignature {
	return &MinHashSignature{
		signatures: make([]uint64, numHashes),
		numHashes:  numHashes,
	}
}

// GetSignatures returns the signature array
func (sig *MinHashSignature) GetSignatures() []uint64 {
	return sig.signatures
}

// GetNumHashes returns the number of hash functions
func (sig *MinHashSignature) GetNumHashes() int {
	return sig.numHashes
}

// HashFunc represents a hash function with parameters a and b
type HashFunc struct {
	a uint64 // Multiplier
	b uint64 // Additive constant
	p uint64 // Large prime number
}

// MinHasher implements the MinHash algorithm for estimating Jaccard similarity
type MinHasher struct {
	numHashes     int        // Number of hash functions to use
	hashFunctions []HashFunc // Pre-computed hash functions
	prime         uint64     // Large prime number for universal hashing
}

// NewMinHasher creates a new MinHasher with the specified number of hash functions
func NewMinHasher(numHashes int) *MinHasher {
	if numHashes <= 0 {
		numHashes = 128 // Default value
	}

	hasher := &MinHasher{
		numHashes:     numHashes,
		hashFunctions: make([]HashFunc, numHashes),
		prime:         2147483647, // Mersenne prime 2^31 - 1
	}

	// Initialize hash functions with random parameters
	hasher.initializeHashFunctions()

	return hasher
}

// NewMinHasherWithSeed creates a MinHasher with a specific seed for reproducible results
func NewMinHasherWithSeed(numHashes int, seed int64) *MinHasher {
	if numHashes <= 0 {
		numHashes = 128
	}

	hasher := &MinHasher{
		numHashes:     numHashes,
		hashFunctions: make([]HashFunc, numHashes),
		prime:         2147483647,
	}

	// Initialize hash functions with seeded random generator
	hasher.initializeHashFunctionsWithSeed(seed)

	return hasher
}

// initializeHashFunctions initializes hash functions with random parameters
func (mh *MinHasher) initializeHashFunctions() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	mh.initializeHashFunctionsWithRNG(rng)
}

// initializeHashFunctionsWithSeed initializes hash functions with a seeded random generator
func (mh *MinHasher) initializeHashFunctionsWithSeed(seed int64) {
	rng := rand.New(rand.NewSource(seed))
	mh.initializeHashFunctionsWithRNG(rng)
}

// initializeHashFunctionsWithRNG initializes hash functions using the provided RNG
func (mh *MinHasher) initializeHashFunctionsWithRNG(rng *rand.Rand) {
	for i := 0; i < mh.numHashes; i++ {
		// Generate random parameters for universal hashing: h(x) = ((ax + b) mod p) mod m
		// a must be non-zero and less than prime
		a := uint64(rng.Int63n(int64(mh.prime-1))) + 1
		b := uint64(rng.Int63n(int64(mh.prime)))

		mh.hashFunctions[i] = HashFunc{
			a: a,
			b: b,
			p: mh.prime,
		}
	}
}

// ComputeSignature computes the MinHash signature for a set of features
func (mh *MinHasher) ComputeSignature(features []string) *MinHashSignature {
	signature := NewMinHashSignature(mh.numHashes)

	// Initialize signature values to maximum possible value
	for i := 0; i < mh.numHashes; i++ {
		signature.signatures[i] = math.MaxUint64
	}

	// Process each feature
	for _, feature := range features {
		// Compute base hash for the feature
		baseHash := mh.computeBaseHash(feature)

		// Apply each hash function and update signature
		for i, hashFunc := range mh.hashFunctions {
			hashValue := mh.applyHashFunction(baseHash, hashFunc)
			if hashValue < signature.signatures[i] {
				signature.signatures[i] = hashValue
			}
		}
	}

	return signature
}

// computeBaseHash computes a base hash for a feature string
func (mh *MinHasher) computeBaseHash(feature string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(feature))
	return h.Sum64()
}

// applyHashFunction applies a universal hash function to a base hash
func (mh *MinHasher) applyHashFunction(baseHash uint64, hashFunc HashFunc) uint64 {
	// Universal hashing: h(x) = ((ax + b) mod p) mod m
	// Since we're working with uint64, we use modular arithmetic carefully
	return ((hashFunc.a*baseHash + hashFunc.b) % hashFunc.p)
}

// EstimateJaccardSimilarity estimates the Jaccard similarity between two signatures
func (mh *MinHasher) EstimateJaccardSimilarity(sig1, sig2 *MinHashSignature) float64 {
	if sig1 == nil || sig2 == nil {
		return 0.0
	}

	if sig1.numHashes != sig2.numHashes {
		return 0.0 // Incompatible signatures
	}

	if sig1.numHashes == 0 {
		return 0.0
	}

	// Count matching signatures
	matches := 0
	for i := 0; i < sig1.numHashes; i++ {
		if sig1.signatures[i] == sig2.signatures[i] {
			matches++
		}
	}

	return float64(matches) / float64(sig1.numHashes)
}

// ComputeJaccardSimilarity computes the exact Jaccard similarity between two feature sets
// This is used for comparison and validation purposes
func ComputeJaccardSimilarity(features1, features2 []string) float64 {
	if len(features1) == 0 && len(features2) == 0 {
		return 1.0
	}

	if len(features1) == 0 || len(features2) == 0 {
		return 0.0
	}

	// Convert to sets for efficient intersection/union computation
	set1 := make(map[string]bool)
	for _, feature := range features1 {
		set1[feature] = true
	}

	set2 := make(map[string]bool)
	for _, feature := range features2 {
		set2[feature] = true
	}

	// Compute intersection
	intersection := 0
	for feature := range set1 {
		if set2[feature] {
			intersection++
		}
	}

	// Union size = |set1| + |set2| - |intersection|
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}

// SignatureStats provides statistics about MinHash signatures
type SignatureStats struct {
	Mean     float64
	Variance float64
	Min      uint64
	Max      uint64
}

// ComputeSignatureStats computes statistics for a MinHash signature
func ComputeSignatureStats(signature *MinHashSignature) SignatureStats {
	if signature == nil || signature.numHashes == 0 {
		return SignatureStats{}
	}

	sigs := signature.signatures
	n := float64(len(sigs))

	// Find min and max
	min := sigs[0]
	max := sigs[0]
	sum := float64(sigs[0])

	for i := 1; i < len(sigs); i++ {
		val := sigs[i]
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
		sum += float64(val)
	}

	// Compute mean
	mean := sum / n

	// Compute variance
	variance := 0.0
	for _, val := range sigs {
		diff := float64(val) - mean
		variance += diff * diff
	}
	variance /= n

	return SignatureStats{
		Mean:     mean,
		Variance: variance,
		Min:      min,
		Max:      max,
	}
}

// CompareSignatures compares two MinHash signatures and returns detailed comparison info
type SignatureComparison struct {
	Similarity      float64 // Estimated Jaccard similarity
	MatchingHashes  int     // Number of matching hash values
	TotalHashes     int     // Total number of hash functions
	SignatureStats1 SignatureStats
	SignatureStats2 SignatureStats
}

// CompareSignatures performs a detailed comparison between two signatures
func (mh *MinHasher) CompareSignatures(sig1, sig2 *MinHashSignature) SignatureComparison {
	comparison := SignatureComparison{
		Similarity:      mh.EstimateJaccardSimilarity(sig1, sig2),
		SignatureStats1: ComputeSignatureStats(sig1),
		SignatureStats2: ComputeSignatureStats(sig2),
	}

	if sig1 != nil && sig2 != nil && sig1.numHashes == sig2.numHashes {
		comparison.TotalHashes = sig1.numHashes
		
		// Count matching hashes
		for i := 0; i < sig1.numHashes; i++ {
			if sig1.signatures[i] == sig2.signatures[i] {
				comparison.MatchingHashes++
			}
		}
	}

	return comparison
}

// SignatureBatch represents a batch of signatures for efficient processing
type SignatureBatch struct {
	signatures map[string]*MinHashSignature // ID to signature mapping
	ids        []string                     // Ordered list of IDs
}

// NewSignatureBatch creates a new signature batch
func NewSignatureBatch() *SignatureBatch {
	return &SignatureBatch{
		signatures: make(map[string]*MinHashSignature),
		ids:        make([]string, 0),
	}
}

// AddSignature adds a signature to the batch
func (batch *SignatureBatch) AddSignature(id string, signature *MinHashSignature) {
	if _, exists := batch.signatures[id]; !exists {
		batch.ids = append(batch.ids, id)
	}
	batch.signatures[id] = signature
}

// GetSignature retrieves a signature by ID
func (batch *SignatureBatch) GetSignature(id string) *MinHashSignature {
	return batch.signatures[id]
}

// GetAllIDs returns all signature IDs
func (batch *SignatureBatch) GetAllIDs() []string {
	result := make([]string, len(batch.ids))
	copy(result, batch.ids)
	return result
}

// Size returns the number of signatures in the batch
func (batch *SignatureBatch) Size() int {
	return len(batch.ids)
}

// ComputeAllPairwiseSimilarities computes similarities for all pairs in the batch
func (mh *MinHasher) ComputeAllPairwiseSimilarities(batch *SignatureBatch) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)
	
	ids := batch.GetAllIDs()
	
	// Initialize result maps
	for _, id := range ids {
		result[id] = make(map[string]float64)
	}
	
	// Compute pairwise similarities
	for i, id1 := range ids {
		sig1 := batch.GetSignature(id1)
		
		// Set self-similarity to 1.0
		result[id1][id1] = 1.0
		
		// Compute similarities with remaining signatures
		for j := i + 1; j < len(ids); j++ {
			id2 := ids[j]
			sig2 := batch.GetSignature(id2)
			
			similarity := mh.EstimateJaccardSimilarity(sig1, sig2)
			result[id1][id2] = similarity
			result[id2][id1] = similarity // Symmetric
		}
	}
	
	return result
}

// FindSimilarSignatures finds all signatures similar to the query above a threshold
func (mh *MinHasher) FindSimilarSignatures(querySignature *MinHashSignature, batch *SignatureBatch, threshold float64) []string {
	var similar []string
	
	for _, id := range batch.GetAllIDs() {
		signature := batch.GetSignature(id)
		similarity := mh.EstimateJaccardSimilarity(querySignature, signature)
		
		if similarity >= threshold {
			similar = append(similar, id)
		}
	}
	
	// Sort by ID for consistent ordering
	sort.Strings(similar)
	
	return similar
}

// ValidateMinHashAccuracy validates MinHash accuracy against exact Jaccard computation
type AccuracyValidation struct {
	NumComparisons   int     // Number of comparisons performed
	MeanError        float64 // Mean absolute error
	MaxError         float64 // Maximum absolute error
	CorrectDirection int     // Number of comparisons with correct relative ordering
}

// ValidateAccuracy validates MinHash accuracy for a batch of feature sets
func (mh *MinHasher) ValidateAccuracy(featureSets map[string][]string) AccuracyValidation {
	validation := AccuracyValidation{}
	
	// Compute MinHash signatures
	signatures := make(map[string]*MinHashSignature)
	for id, features := range featureSets {
		signatures[id] = mh.ComputeSignature(features)
	}
	
	// Get all IDs for comparison
	var ids []string
	for id := range featureSets {
		ids = append(ids, id)
	}
	sort.Strings(ids) // Ensure consistent ordering
	
	var totalError float64
	var maxError float64
	var correctDirection int
	var comparisons int
	
	// Compare all pairs
	for i := 0; i < len(ids); i++ {
		for j := i + 1; j < len(ids); j++ {
			id1, id2 := ids[i], ids[j]
			
			// Compute exact Jaccard similarity
			exactSim := ComputeJaccardSimilarity(featureSets[id1], featureSets[id2])
			
			// Compute estimated similarity
			estimatedSim := mh.EstimateJaccardSimilarity(signatures[id1], signatures[id2])
			
			// Calculate error
			error := math.Abs(exactSim - estimatedSim)
			totalError += error
			if error > maxError {
				maxError = error
			}
			
			// Check if direction is correct (for ranking purposes)
			// Compare with another random pair
			if len(ids) > j+1 {
				id3 := ids[j+1]
				exactSim2 := ComputeJaccardSimilarity(featureSets[id1], featureSets[id3])
				estimatedSim2 := mh.EstimateJaccardSimilarity(signatures[id1], signatures[id3])
				
				if (exactSim > exactSim2 && estimatedSim > estimatedSim2) ||
				   (exactSim < exactSim2 && estimatedSim < estimatedSim2) ||
				   (exactSim == exactSim2 && estimatedSim == estimatedSim2) {
					correctDirection++
				}
			}
			
			comparisons++
		}
	}
	
	validation.NumComparisons = comparisons
	validation.MaxError = maxError
	validation.CorrectDirection = correctDirection
	
	if comparisons > 0 {
		validation.MeanError = totalError / float64(comparisons)
	}
	
	return validation
}