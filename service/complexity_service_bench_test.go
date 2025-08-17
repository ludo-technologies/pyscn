package service

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/pyqol/pyqol/domain"
)

// BenchmarkSortByComplexity benchmarks the sorting performance
func BenchmarkSortByComplexity(b *testing.B) {
	// Create test data with various sizes
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Generate random test data
			functions := generateRandomFunctions(size)
			service := &ComplexityServiceImpl{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Make a copy to avoid sorting already sorted data
				testData := make([]domain.FunctionComplexity, len(functions))
				copy(testData, functions)
				
				service.sortByComplexity(testData)
			}
		})
	}
}

// BenchmarkSortByName benchmarks name sorting performance
func BenchmarkSortByName(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			functions := generateRandomFunctions(size)
			service := &ComplexityServiceImpl{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				testData := make([]domain.FunctionComplexity, len(functions))
				copy(testData, functions)
				
				service.sortByName(testData)
			}
		})
	}
}

// BenchmarkSortByRisk benchmarks risk level sorting performance
func BenchmarkSortByRisk(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			functions := generateRandomFunctions(size)
			service := &ComplexityServiceImpl{}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				testData := make([]domain.FunctionComplexity, len(functions))
				copy(testData, functions)
				
				service.sortByRisk(testData)
			}
		})
	}
}

// generateRandomFunctions generates random test data
func generateRandomFunctions(count int) []domain.FunctionComplexity {
	// Use local random generator for reproducible benchmarks (Go 1.20+)
	rng := rand.New(rand.NewSource(42))
	
	functions := make([]domain.FunctionComplexity, count)
	riskLevels := []domain.RiskLevel{
		domain.RiskLevelLow,
		domain.RiskLevelMedium,
		domain.RiskLevelHigh,
	}

	for i := 0; i < count; i++ {
		complexity := rng.Intn(50) + 1
		riskLevel := riskLevels[rng.Intn(3)]
		
		functions[i] = domain.FunctionComplexity{
			Name:     fmt.Sprintf("function_%d", i),
			FilePath: fmt.Sprintf("/path/to/file_%d.py", i%10),
			Metrics: domain.ComplexityMetrics{
				Complexity:        complexity,
				Nodes:            complexity * 2,
				Edges:            complexity * 3,
				IfStatements:     rng.Intn(complexity),
				LoopStatements:   rng.Intn(5),
				ExceptionHandlers: rng.Intn(3),
				SwitchCases:      0,
			},
			RiskLevel: riskLevel,
		}
	}

	return functions
}

// Benchmark comparison with old bubble sort implementation (for documentation)
// The old bubble sort had O(n²) complexity:
// - 10 items: ~45 comparisons
// - 100 items: ~4,950 comparisons  
// - 1000 items: ~499,500 comparisons
// - 10000 items: ~49,995,000 comparisons
//
// The new sort.Slice has O(n log n) complexity:
// - 10 items: ~33 comparisons
// - 100 items: ~664 comparisons
// - 1000 items: ~9,965 comparisons  
// - 10000 items: ~132,877 comparisons
//
// This represents a massive performance improvement for large datasets!