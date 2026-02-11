package analyzer

import "math"

// handles the weight calculations for code features
type TFIDFCalculator struct {
	DocumentFrequency map[string]int // feature → how many code blocks contain it
	TotalDocuments    int            // total number of code blocks scanned
}

// NewTFIDFCalculator creates a new TF-IDF calculator
func NewTFIDFCalculator() *TFIDFCalculator {
	return &TFIDFCalculator{
		DocumentFrequency: make(map[string]int),
	}
}

// ComputeIDF pre-calculates the counts for IDF from a set of code fragments
func (c *TFIDFCalculator) ComputeIDF(fragments []*CodeFragment) {
	if len(fragments) == 0 {
		return
	}
	c.TotalDocuments = len(fragments)
	extractor := NewASTFeatureExtractor() // Use default options
	for _, f := range fragments {
		if f.TreeNode != nil {
			features, _ := extractor.ExtractFeatures(f.TreeNode)
			// Deduplicate features per "document" for DF calculation
			uniqueFeatures := make(map[string]struct{})
			for _, feat := range features {
				uniqueFeatures[feat] = struct{}{}
			}
			for feat := range uniqueFeatures {
				c.DocumentFrequency[feat]++
			}
		}
	}
}

// calculates the Inverse Document Frequency
func (c *TFIDFCalculator) IDF(feature string) float64 {
	df := c.DocumentFrequency[feature]
	if df == 0 {
		return 0
	}
	// log(N/df) dampens the weight of very common terms
	return math.Log(float64(c.TotalDocuments) / float64(df))
}

// converts features into TF-IDF weighted values
func (c *TFIDFCalculator) ToWeightedVector(features []string) map[string]float64 {
	tf := make(map[string]int)
	for _, f := range features {
		tf[f]++
	}

	vector := make(map[string]float64)
	for f, count := range tf {
		vector[f] = float64(count) * c.IDF(f)
	}
	return vector
}

func CosineSimilarity(v1, v2 map[string]float64) float64 {
	var dotProduct, norm1, norm2 float64
	for k, val1 := range v1 {
		if val2, ok := v2[k]; ok {
			dotProduct += val1 * val2
		}
		norm1 += val1 * val1
	}
	for _, val2 := range v2 {
		norm2 += val2 * val2
	}
	if norm1 == 0 || norm2 == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}
