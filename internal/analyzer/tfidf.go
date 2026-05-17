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
	idfCache  map[string]float64
	totalDocs int
	docFreq   map[string]int // how many documents contain each feature
}

// NewTFIDFCalculator creates a new TF-IDF calculator with an empty cache.
func NewTFIDFCalculator() *TFIDFCalculator {
	return &TFIDFCalculator{
		idfCache: make(map[string]float64),
		docFreq:  make(map[string]int),
	}
}

// BuildFromCorpus computes IDF values from a corpus of code fragments.
// Call this once before using ToWeightedVector with real TF-IDF weights.
// Each fragment's Features field is used to compute document frequency.
func (calc *TFIDFCalculator) BuildFromCorpus(fragments []*CodeFragment) {
	calc.totalDocs = len(fragments)
	if calc.totalDocs == 0 {
		return
	}

	// Count document frequency: how many docs contain each unique feature
	for _, frag := range fragments {
		if frag == nil || len(frag.Features) == 0 {
			continue
		}
		// Use a map to count each feature once per document
		seen := make(map[string]bool)
		for _, f := range frag.Features {
			if !seen[f] {
				calc.docFreq[f]++
				seen[f] = true
			}
		}
	}

	// Pre-compute IDF for all seen features
	for feature := range calc.docFreq {
		calc.idfCache[feature] = calc.IDF(calc.totalDocs, calc.docFreq[feature])
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
// Each feature's weight is TF * IDF, where IDF is sourced from the pre-computed
// corpus cache (call BuildFromCorpus first). Features not in the IDF cache use
// IDF=1.0 (pure TF weighting). Empty feature slices return an empty vector.
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

	// Apply TF-IDF weighting: TF * IDF
	for feature, count := range tf {
		tfWeight := 1.0 + math.Log(float64(count))
		idf := 1.0 // default to pure TF if no corpus built
		if cached, ok := calc.idfCache[feature]; ok {
			idf = cached
		}
		weighted[feature] = tfWeight * idf
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
