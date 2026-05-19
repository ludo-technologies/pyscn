package analyzer

import (
	"math"
	"testing"
)

func TestNewTFIDFCalculator(t *testing.T) {
	calc := NewTFIDFCalculator()
	if calc == nil {
		t.Fatal("expected non-nil calculator")
	}
	if calc.idfCache == nil {
		t.Error("expected non-nil IDF cache")
	}
}

func TestIDF(t *testing.T) {
	calc := NewTFIDFCalculator()

	tests := []struct {
		name      string
		totalDocs int
		docFreq   int
		want      float64
	}{
		{"valid docs", 100, 10, 2.302585092994046}, // log(10)
		{"single doc", 10, 1, 2.302585092994046},
		{"doc in all docs", 100, 100, 0.0},
		{"zero total docs", 0, 10, 0.0},
		{"zero doc freq", 100, 0, 0.0},
		{"negative total docs", -10, 5, 0.0},
		{"negative doc freq", 10, -5, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.IDF(tt.totalDocs, tt.docFreq)
			if got != tt.want {
				t.Errorf("IDF(%d, %d) = %v, want %v", tt.totalDocs, tt.docFreq, got, tt.want)
			}
		})
	}
}

func TestComputeIDF_CacheHit(t *testing.T) {
	calc := NewTFIDFCalculator()

	// First call computes and caches
	got1 := calc.ComputeIDF("feature", 100, 10)
	// Second call should return cached value
	got2 := calc.ComputeIDF("feature", 100, 10)

	if got1 != got2 {
		t.Errorf("expected cache hit, got %v vs %v", got1, got2)
	}
}

func TestComputeIDF_EmptyCache(t *testing.T) {
	calc := NewTFIDFCalculator()

	// Empty cache should compute IDF
	got := calc.ComputeIDF("feature", 100, 10)
	want := calc.IDF(100, 10)

	if got != want {
		t.Errorf("ComputeIDF = %v, want %v", got, want)
	}

	// Verify it was cached
	if _, ok := calc.idfCache["feature"]; !ok {
		t.Error("expected feature to be cached")
	}
}

func TestToWeightedVector_Empty(t *testing.T) {
	calc := NewTFIDFCalculator()

	// Empty features
	vec := calc.ToWeightedVector([]string{})
	if len(vec) != 0 {
		t.Errorf("expected empty vector, got %v", vec)
	}

	// Nil features
	vec = calc.ToWeightedVector(nil)
	if len(vec) != 0 {
		t.Errorf("expected empty vector for nil, got %v", vec)
	}
}

func TestToWeightedVector_SingleFeature(t *testing.T) {
	calc := NewTFIDFCalculator()

	vec := calc.ToWeightedVector([]string{"if_stmt"})
	if len(vec) != 1 {
		t.Fatalf("expected 1 key, got %d", len(vec))
	}
	// TF = 1 + log(1) = 1 + 0 = 1
	if vec["if_stmt"] != 1.0 {
		t.Errorf("expected TF weight 1.0, got %v", vec["if_stmt"])
	}
}

func TestToWeightedVector_MultipleFeatures(t *testing.T) {
	calc := NewTFIDFCalculator()

	vec := calc.ToWeightedVector([]string{"if_stmt", "if_stmt", "return"})

	if math.Abs(vec["if_stmt"]-1.6931471805599452) > 1e-10 {
		t.Errorf("expected TF weight ~1.693, got %v", vec["if_stmt"])
	}
	if vec["return"] != 1.0 { // 1 + log(1)
		t.Errorf("expected TF weight 1.0, got %v", vec["return"])
	}
}

func TestToWeightedVector_UniqueFeatures(t *testing.T) {
	calc := NewTFIDFCalculator()

	vec := calc.ToWeightedVector([]string{"if_stmt", "return", "expr"})

	// All single occurrence
	for key, val := range vec {
		if val != 1.0 {
			t.Errorf("expected TF weight 1.0 for %s, got %v", key, val)
		}
	}
	if len(vec) != 3 {
		t.Errorf("expected 3 unique features, got %d", len(vec))
	}
}

func TestCosineSimilarity_EmptyVectors(t *testing.T) {
	// Both empty
	vec := CosineSimilarity(map[string]float64{}, map[string]float64{})
	if vec != 0.0 {
		t.Errorf("expected 0 for empty vectors, got %v", vec)
	}

	// One empty
	vec = CosineSimilarity(map[string]float64{"a": 1.0}, map[string]float64{})
	if vec != 0.0 {
		t.Errorf("expected 0 when one vec empty, got %v", vec)
	}

	vec = CosineSimilarity(map[string]float64{}, map[string]float64{"a": 1.0})
	if vec != 0.0 {
		t.Errorf("expected 0 when one vec empty, got %v", vec)
	}
}

func TestCosineSimilarity_IdenticalVectors(t *testing.T) {
	v1 := map[string]float64{"if_stmt": 1.0, "return": 1.0}
	v2 := map[string]float64{"if_stmt": 1.0, "return": 1.0}

	sim := CosineSimilarity(v1, v2)
	if math.Abs(sim-1.0) > 1e-9 {
		t.Errorf("expected ~1.0 for identical vectors, got %v", sim)
	}
}

func TestCosineSimilarity_OrthogonalVectors(t *testing.T) {
	v1 := map[string]float64{"if_stmt": 1.0}
	v2 := map[string]float64{"return": 1.0}

	sim := CosineSimilarity(v1, v2)
	if sim != 0.0 {
		t.Errorf("expected 0.0 for orthogonal vectors, got %v", sim)
	}
}

func TestCosineSimilarity_PartialOverlap(t *testing.T) {
	v1 := map[string]float64{"if_stmt": 1.0, "return": 1.0}
	v2 := map[string]float64{"if_stmt": 1.0, "expr": 1.0}

	sim := CosineSimilarity(v1, v2)
	// cos(angle) = dot / (||v1|| * ||v2||) = 1 / (sqrt(2) * sqrt(2)) = 0.5
	if math.Abs(sim-0.5) > 1e-9 {
		t.Errorf("expected ~0.5 for partial overlap, got %v", sim)
	}
}

func TestCosineSimilarity_ZeroNormVector(t *testing.T) {
	v1 := map[string]float64{}
	v2 := map[string]float64{"a": 1.0}

	sim := CosineSimilarity(v1, v2)
	if sim != 0.0 {
		t.Errorf("expected 0.0 when one norm is zero, got %v", sim)
	}
}

func TestCosineSimilarity_UnequalNorms(t *testing.T) {
	v1 := map[string]float64{"a": 2.0}
	v2 := map[string]float64{"a": 1.0}

	sim := CosineSimilarity(v1, v2)
	// dot = 2, norm1 = 2, norm2 = 1
	// cos = 2 / (2 * 1) = 1.0
	if sim != 1.0 {
		t.Errorf("expected 1.0 for same direction different magnitude, got %v", sim)
	}
}
