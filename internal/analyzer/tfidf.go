package analyzer

import (
	"math"
)

// TFIDFCalculator computes TF-IDF (Term Frequency-Inverse Document Frequency)
// values for code clone detection. This enables cosine similarity comparison
// of code fragments based on their AST feature distributions rather than
// simple Jaccard similarity.
//
// TF-IDF weights features by their importance across the corpus of fragments,
// so rare distinguishing features contribute more to similarity than common
// boilerplate features.
type TFIDFCalculator struct {
	idfCache map[string]float64
}

// NewTFIDFCalculator creates a new TF-IDF calculator with an empty cache.
func NewTFIDFCalculator() *TFIDFCalculator {
	return &TFIDFCalculator{
		idfCache: make(map[string]float64),
	}
}

// ComputeIDF computes the IDF (Inverse Document Frequency) for a given feature.
// It uses the cached value if available, otherwise computes and caches it.
func (calc *TFIDFCalculator) ComputeIDF(feature string, totalDocs int, docFreq int) float64 {
	cacheKey := feature
	if cached, ok := calc.idfCache[cacheKey]; ok {
		return cached
	}

	idf := calc.IDF(totalDocs, docFreq)
	calc.idfCache[cacheKey] = idf
	return idf
}

// IDF computes the IDF value using the standard log formula:
// IDF = log(total_documents / document_frequency)
// Returns 0 for invalid input to avoid division errors.
func (calc *TFIDFCalculator) IDF(totalDocs int, docFreq int) float64 {
	if totalDocs <= 0 || docFreq <= 0 {
		return 0.0
	}
	return math.Log(float64(totalDocs) / float64(docFreq))
}

// ToWeightedVector converts a slice of features to a TF-IDF weighted vector.
// Each feature's TF is its frequency in the document, weighted by its IDF.
// Empty feature slices return an empty vector.
func (calc *TFIDFCalculator) ToWeightedVector(features []string) map[string]float64 {
	weighted := make(map[string]float64)
	if len(features) == 0 {
		return weighted
	}

	// Count term frequencies
	tf := make(map[string]int)
	for _, f := range features {
		tf[f]++
	}

	// Apply TF-IDF weighting
	for feature, count := range tf {
		// Use 1 + log(tf) for term frequency to dampen high frequencies
		tfWeight := 1.0 + math.Log(float64(count))
		// IDF is set to 1.0 during vector conversion since corpus stats
		// are not available at this stage; caller should use ComputeIDF
		// with corpus-wide statistics for proper IDF weighting
		weighted[feature] = tfWeight
	}

	return weighted
}

// CosineSimilarity computes the cosine similarity between two weighted vectors.
// Returns 0 if either vector is empty to avoid division by zero.
//
// Cosine similarity measures the angle between two vectors, treating them
// as directions in high-dimensional space. Values range from 0 (orthogonal)
// to 1 (identical direction).
func CosineSimilarity(v1, v2 map[string]float64) float64 {
	if len(v1) == 0 || len(v2) == 0 {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64
	for key, val1 := range v1 {
		if val2, ok := v2[key]; ok {
			dotProduct += val1 * val2
		}
	}
	for _, val := range v1 {
		norm1 += val * val
	}
	for _, val := range v2 {
		norm2 += val * val
	}

	denom := math.Sqrt(norm1) * math.Sqrt(norm2)
	if denom == 0 {
		return 0.0
	}

	return dotProduct / denom
}