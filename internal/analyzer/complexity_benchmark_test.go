package analyzer

import (
	"testing"
)

// BenchmarkComplexityCalculation benchmarks the complexity calculation performance
func BenchmarkComplexityCalculation(b *testing.B) {
	testCases := []struct {
		name     string
		setupCFG func() *CFG
	}{
		{
			name: "SimpleCFG",
			setupCFG: func() *CFG {
				cfg := NewCFG("simple_function")
				block := cfg.CreateBlock("main")
				cfg.ConnectBlocks(cfg.Entry, block, EdgeNormal)
				cfg.ConnectBlocks(block, cfg.Exit, EdgeNormal)
				return cfg
			},
		},
		{
			name: "IfStatementCFG",
			setupCFG: func() *CFG {
				cfg := NewCFG("if_function")
				condBlock := cfg.CreateBlock("condition")
				thenBlock := cfg.CreateBlock("then")
				elseBlock := cfg.CreateBlock("else")

				cfg.ConnectBlocks(cfg.Entry, condBlock, EdgeNormal)
				cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
				cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
				cfg.ConnectBlocks(thenBlock, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(elseBlock, cfg.Exit, EdgeNormal)

				return cfg
			},
		},
		{
			name: "ComplexCFG",
			setupCFG: func() *CFG {
				cfg := NewCFG("complex_function")

				// Create 10 decision points for medium complexity
				current := cfg.Entry
				for i := 0; i < 10; i++ {
					condBlock := cfg.CreateBlock("condition")
					thenBlock := cfg.CreateBlock("then")
					elseBlock := cfg.CreateBlock("else")

					cfg.ConnectBlocks(current, condBlock, EdgeNormal)
					cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
					cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
					cfg.ConnectBlocks(thenBlock, elseBlock, EdgeNormal)

					current = elseBlock
				}
				cfg.ConnectBlocks(current, cfg.Exit, EdgeNormal)

				return cfg
			},
		},
		{
			name: "HighComplexityCFG",
			setupCFG: func() *CFG {
				cfg := NewCFG("high_complexity_function")

				// Create 50 decision points for high complexity
				current := cfg.Entry
				for i := 0; i < 50; i++ {
					condBlock := cfg.CreateBlock("condition")
					thenBlock := cfg.CreateBlock("then")
					elseBlock := cfg.CreateBlock("else")

					cfg.ConnectBlocks(current, condBlock, EdgeNormal)
					cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
					cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
					cfg.ConnectBlocks(thenBlock, elseBlock, EdgeNormal)

					current = elseBlock
				}
				cfg.ConnectBlocks(current, cfg.Exit, EdgeNormal)

				return cfg
			},
		},
	}

	for _, tc := range testCases {
		cfg := tc.setupCFG()
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = CalculateComplexity(cfg)
			}
		})
	}
}

// BenchmarkFileComplexityCalculation benchmarks calculating complexity for multiple CFGs
func BenchmarkFileComplexityCalculation(b *testing.B) {
	// Create a slice of CFGs representing functions in a file
	cfgs := make([]*CFG, 20)
	for i := 0; i < 20; i++ {
		cfg := NewCFG("function")

		// Create varying complexity functions
		numDecisions := (i % 5) + 1 // 1-5 decision points
		current := cfg.Entry

		for j := 0; j < numDecisions; j++ {
			condBlock := cfg.CreateBlock("condition")
			thenBlock := cfg.CreateBlock("then")
			elseBlock := cfg.CreateBlock("else")

			cfg.ConnectBlocks(current, condBlock, EdgeNormal)
			cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
			cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
			cfg.ConnectBlocks(thenBlock, elseBlock, EdgeNormal)

			current = elseBlock
		}
		cfg.ConnectBlocks(current, cfg.Exit, EdgeNormal)

		cfgs[i] = cfg
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateFileComplexity(cfgs)
	}
}

// BenchmarkAggregateComplexityCalculation benchmarks aggregate metrics calculation
func BenchmarkAggregateComplexityCalculation(b *testing.B) {
	// Create complexity results
	results := make([]*ComplexityResult, 100)
	for i := 0; i < 100; i++ {
		complexity := (i % 25) + 1 // 1-25 complexity range
		results[i] = &ComplexityResult{
			Complexity:   complexity,
			FunctionName: "test_function",
			RiskLevel:    assessRiskLevel(complexity),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateAggregateComplexity(results)
	}
}

// BenchmarkComplexityMemoryUsage benchmarks memory allocation during complexity calculation
func BenchmarkComplexityMemoryUsage(b *testing.B) {
	testCases := []struct {
		name     string
		cfgCount int
		avgNodes int
	}{
		{"SmallFile", 5, 10},
		{"MediumFile", 20, 25},
		{"LargeFile", 100, 50},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Create CFGs for the benchmark
			cfgs := make([]*CFG, tc.cfgCount)
			for i := 0; i < tc.cfgCount; i++ {
				cfg := NewCFG("function")

				// Create CFG with specified number of nodes
				current := cfg.Entry
				for j := 0; j < tc.avgNodes; j++ {
					block := cfg.CreateBlock("block")
					cfg.ConnectBlocks(current, block, EdgeNormal)
					current = block
				}
				cfg.ConnectBlocks(current, cfg.Exit, EdgeNormal)

				cfgs[i] = cfg
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				results := CalculateFileComplexity(cfgs)
				_ = CalculateAggregateComplexity(results)
			}
		})
	}
}
