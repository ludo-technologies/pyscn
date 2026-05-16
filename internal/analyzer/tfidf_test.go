package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTFIDFCalculator_IDF(t *testing.T) {
	calc := NewTFIDFCalculator()

	t.Run("standard IDF computation", func(t *testing.T) {
		// IDF = log(100 / 10) = log(10) ≈ 2.302
		idf := calc.IDF(100, 10)
		assert.InDelta(t, 2.302585, idf, 0.001)
	})

	t.Run(" IDF for feature in all documents", func(t *testing.T) {
		// Feature appears in every doc: log(100/100) = 0
		idf := calc.IDF(100, 100)
		assert.Equal(t, 0.0, idf)
	})

	t.Run(" IDF for rare feature", func(t *testing.T) {
		// Feature in 1 of 100 docs: log(100/1) = 4.605
		idf := calc.IDF(100, 1)
		assert.InDelta(t, 4.605170, idf, 0.001)
	})

	t.Run("zero guard totalDocs", func(t *testing.T) {
		assert.Equal(t, 0.0, calc.IDF(0, 10))
	})

	t.Run("zero guard docFreq", func(t *testing.T) {
		assert.Equal(t, 0.0, calc.IDF(100, 0))
	})
}

func TestTFIDFCalculator_BuildFromCorpus(t *testing.T) {
	t.Run("builds IDF cache from fragments", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		fragments := []*CodeFragment{
			{Features: []string{"if", "expr", "block"}},
			{Features: []string{"if", "expr", "while"}},
			{Features: []string{"if", "expr", "for"}},
		}

		calc.BuildFromCorpus(fragments)

		// "if" and "expr" appear in all 3 docs
		assert.InDelta(t, 0.0, calc.idfCache["if"], 0.001)
		assert.InDelta(t, 0.0, calc.idfCache["expr"], 0.001)
		// "block", "while", "for" appear in 1 doc each
		assert.InDelta(t, 1.099, calc.idfCache["block"], 0.01)
		assert.InDelta(t, 1.099, calc.idfCache["while"], 0.01)
		assert.InDelta(t, 1.099, calc.idfCache["for"], 0.01)
	})

	t.Run("empty corpus", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		calc.BuildFromCorpus([]*CodeFragment{})
		assert.Empty(t, calc.idfCache)
	})

	t.Run("nil fragments skipped", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		calc.BuildFromCorpus([]*CodeFragment{nil, {Features: []string{"fn"}}})
		assert.Contains(t, calc.idfCache, "fn")
	})

	t.Run("deduplicates feature count per document", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		fragments := []*CodeFragment{
			{Features: []string{"if", "if", "expr"}}, // "if" appears twice but counts once
		}
		calc.BuildFromCorpus(fragments)
		assert.Equal(t, 1, calc.docFreq["if"])
	})
}

func TestTFIDFCalculator_ToWeightedVector(t *testing.T) {
	t.Run("pure TF when no corpus built", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		// No BuildFromCorpus call — IDF defaults to 1.0
		vec := calc.ToWeightedVector([]string{"if", "if", "expr"})

		// tf("if")=2 → (1+log2)*1.0 ≈ 1.693, tf("expr")=1 → (1+log1)*1.0 = 1.0
		assert.InDelta(t, 1.693, vec["if"], 0.001)
		assert.InDelta(t, 1.0, vec["expr"], 0.001)
		assert.Greater(t, vec["if"], vec["expr"]) // higher TF = higher weight
	})

	t.Run("TF-IDF when corpus is built", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		fragments := []*CodeFragment{
			{Features: []string{"if", "expr", "block"}},
			{Features: []string{"if", "expr", "while"}},
			{Features: []string{"if", "expr", "for"}},
		}
		calc.BuildFromCorpus(fragments)

		vec := calc.ToWeightedVector([]string{"if", "block"})

		// "if" in 3/3 docs → IDF ≈ 0, so weight ≈ 0
		// "block" in 1/3 docs → IDF ≈ 1.099, so weight > 0
		assert.InDelta(t, 0.0, vec["if"], 0.01)
		assert.Greater(t, vec["block"], 0.0)
	})

	t.Run("empty input returns empty vector", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		vec := calc.ToWeightedVector([]string{})
		assert.Empty(t, vec)
	})

	t.Run("unknown feature uses idf=1.0", func(t *testing.T) {
		calc := NewTFIDFCalculator()
		calc.BuildFromCorpus([]*CodeFragment{{Features: []string{"if"}}})
		vec := calc.ToWeightedVector([]string{"unknown"})

		// unknown not in corpus, IDF defaults to 1.0
		tfWeight := 1.0 + 0 // count=1 → log(1)=0
		assert.InDelta(t, tfWeight, vec["unknown"], 0.001)
	})
}

func TestCosineSimilarity(t *testing.T) {
	t.Run("identical vectors", func(t *testing.T) {
		v1 := map[string]float64{"a": 1.0, "b": 0.5}
		v2 := map[string]float64{"a": 1.0, "b": 0.5}
		assert.InDelta(t, 1.0, CosineSimilarity(v1, v2), 0.0001)
	})

	t.Run("orthogonal vectors", func(t *testing.T) {
		v1 := map[string]float64{"a": 1.0}
		v2 := map[string]float64{"b": 1.0}
		assert.Equal(t, 0.0, CosineSimilarity(v1, v2))
	})

	t.Run("empty vector", func(t *testing.T) {
		assert.Equal(t, 0.0, CosineSimilarity(map[string]float64{}, map[string]float64{"a": 1.0}))
		assert.Equal(t, 0.0, CosineSimilarity(map[string]float64{"a": 1.0}, map[string]float64{}))
	})

	t.Run("partial overlap", func(t *testing.T) {
		v1 := map[string]float64{"a": 1.0, "b": 1.0}
		v2 := map[string]float64{"a": 1.0, "c": 1.0}
		// dot = 1.0, |v1|=√2, |v2|=√2, cos = 1/2 = 0.5
		assert.InDelta(t, 0.5, CosineSimilarity(v1, v2), 0.001)
	})
}

func TestTFIDFEndToEnd(t *testing.T) {
	// Full integration: build corpus → convert vectors → cosine similarity
	calc := NewTFIDFCalculator()
	corpus := []*CodeFragment{
		{Features: []string{"if", "expr", "block"}},
		{Features: []string{"if", "expr", "return"}},
		{Features: []string{"while", "expr", "block"}},
	}
	calc.BuildFromCorpus(corpus)

	// Two fragments both using "if" and "expr" should have high similarity
	// even though "if" appears in all corpus docs (low IDF) and "return"/"block" appear rarely (high IDF)
	v1 := calc.ToWeightedVector([]string{"if", "if", "expr", "return"})
	v2 := calc.ToWeightedVector([]string{"if", "expr", "return"})

	sim := CosineSimilarity(v1, v2)
	assert.Greater(t, sim, 0.5, "fragments sharing rare features should score high")

	// Fragment with only "while" should be less similar to one with "if"/"expr"
	v3 := calc.ToWeightedVector([]string{"while", "block"})
	sim2 := CosineSimilarity(v1, v3)
	assert.Less(t, sim2, sim, "different feature sets should score lower")
}