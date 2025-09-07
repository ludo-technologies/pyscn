package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/config"
	"testing"
)

func TestComplexityResult(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		result := &ComplexityResult{
			FunctionName: "test_function",
			Complexity:   5,
			RiskLevel:    "low",
		}

		expected := "Function: test_function, Complexity: 5, Risk: low"
		if result.String() != expected {
			t.Errorf("Expected %q, got %q", expected, result.String())
		}
	})
}

func TestCalculateComplexity(t *testing.T) {
	testCases := []struct {
		name               string
		setupCFG           func() *CFG
		expectedComplexity int
		expectedRiskLevel  string
		expectedNodes      int
		expectedEdges      int
	}{
		{
			name: "NilCFG",
			setupCFG: func() *CFG {
				return nil
			},
			expectedComplexity: 0,
			expectedRiskLevel:  "low",
			expectedNodes:      0,
			expectedEdges:      0,
		},
		{
			name: "SimpleCFG",
			setupCFG: func() *CFG {
				// Simple linear function: entry -> block -> exit
				cfg := NewCFG("simple_function")
				block := cfg.CreateBlock("main")
				cfg.ConnectBlocks(cfg.Entry, block, EdgeNormal)
				cfg.ConnectBlocks(block, cfg.Exit, EdgeNormal)
				return cfg
			},
			expectedComplexity: 1,
			expectedRiskLevel:  "low",
			expectedNodes:      1, // Only main block (entry/exit excluded)
			expectedEdges:      2, // entry->main, main->exit
		},
		{
			name: "IfStatementCFG",
			setupCFG: func() *CFG {
				// if x > 0: return True else: return False
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
			expectedComplexity: 2,
			expectedRiskLevel:  "low",
			expectedNodes:      3, // condition, then, else
			expectedEdges:      5, // entry->cond, cond->then, cond->else, then->exit, else->exit
		},
		{
			name: "NestedIfCFG",
			setupCFG: func() *CFG {
				// if x > 0: if y > 0: return 1 else: return 2 else: return 3
				cfg := NewCFG("nested_if_function")

				cond1 := cfg.CreateBlock("condition1")
				cond2 := cfg.CreateBlock("condition2")
				then1 := cfg.CreateBlock("then1")
				else1 := cfg.CreateBlock("else1")
				else2 := cfg.CreateBlock("else2")

				cfg.ConnectBlocks(cfg.Entry, cond1, EdgeNormal)
				cfg.ConnectBlocks(cond1, cond2, EdgeCondTrue)
				cfg.ConnectBlocks(cond1, else2, EdgeCondFalse)
				cfg.ConnectBlocks(cond2, then1, EdgeCondTrue)
				cfg.ConnectBlocks(cond2, else1, EdgeCondFalse)
				cfg.ConnectBlocks(then1, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(else1, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(else2, cfg.Exit, EdgeNormal)

				return cfg
			},
			expectedComplexity: 3,
			expectedRiskLevel:  "low",
			expectedNodes:      5, // cond1, cond2, then1, else1, else2
			expectedEdges:      8, // All the connections above
		},
		{
			name: "LoopCFG",
			setupCFG: func() *CFG {
				// while x > 0: x -= 1
				cfg := NewCFG("loop_function")

				loopHeader := cfg.CreateBlock("loop_header")
				loopBody := cfg.CreateBlock("loop_body")

				cfg.ConnectBlocks(cfg.Entry, loopHeader, EdgeNormal)
				cfg.ConnectBlocks(loopHeader, loopBody, EdgeCondTrue)
				cfg.ConnectBlocks(loopHeader, cfg.Exit, EdgeCondFalse)
				cfg.ConnectBlocks(loopBody, loopHeader, EdgeLoop)

				return cfg
			},
			expectedComplexity: 2,
			expectedRiskLevel:  "low",
			expectedNodes:      2, // loop_header, loop_body
			expectedEdges:      4, // entry->header, header->body, header->exit, body->header
		},
		{
			name: "ComplexCFG",
			setupCFG: func() *CFG {
				// Complex function with multiple decision points
				cfg := NewCFG("complex_function")

				blocks := make([]*BasicBlock, 10)
				for i := 0; i < 10; i++ {
					blocks[i] = cfg.CreateBlock("")
				}

				// Create a complex control flow with multiple paths
				cfg.ConnectBlocks(cfg.Entry, blocks[0], EdgeNormal)
				cfg.ConnectBlocks(blocks[0], blocks[1], EdgeCondTrue)
				cfg.ConnectBlocks(blocks[0], blocks[2], EdgeCondFalse)
				cfg.ConnectBlocks(blocks[1], blocks[3], EdgeCondTrue)
				cfg.ConnectBlocks(blocks[1], blocks[4], EdgeCondFalse)
				cfg.ConnectBlocks(blocks[2], blocks[5], EdgeNormal)
				cfg.ConnectBlocks(blocks[3], blocks[6], EdgeNormal)
				cfg.ConnectBlocks(blocks[4], blocks[7], EdgeLoop)
				cfg.ConnectBlocks(blocks[7], blocks[4], EdgeLoop)
				cfg.ConnectBlocks(blocks[4], blocks[8], EdgeCondFalse)
				cfg.ConnectBlocks(blocks[5], cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(blocks[6], cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(blocks[8], cfg.Exit, EdgeNormal)

				return cfg
			},
			expectedComplexity: 4, // Multiple decision points (3 conditional blocks + 1)
			expectedRiskLevel:  "low",
			expectedNodes:      9, // All blocks except entry/exit
			expectedEdges:      13,
		},
		{
			name: "HighComplexityCFG",
			setupCFG: func() *CFG {
				// Create a high complexity function (>20)
				cfg := NewCFG("high_complexity_function")

				// Create 25 decision points to exceed high threshold
				current := cfg.Entry
				for i := 0; i < 25; i++ {
					condBlock := cfg.CreateBlock("")
					thenBlock := cfg.CreateBlock("")
					elseBlock := cfg.CreateBlock("")

					cfg.ConnectBlocks(current, condBlock, EdgeNormal)
					cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
					cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
					cfg.ConnectBlocks(thenBlock, elseBlock, EdgeNormal)

					current = elseBlock
				}
				cfg.ConnectBlocks(current, cfg.Exit, EdgeNormal)

				return cfg
			},
			expectedComplexity: 26, // Should be > 20 for high risk
			expectedRiskLevel:  "high",
		},
		{
			name: "ElifChainCFG",
			setupCFG: func() *CFG {
				// if x == 1: ... elif x == 2: ... elif x == 3: ... else: ...
				cfg := NewCFG("elif_chain_function")

				cond1 := cfg.CreateBlock("cond1") // if x == 1
				cond2 := cfg.CreateBlock("cond2") // elif x == 2
				cond3 := cfg.CreateBlock("cond3") // elif x == 3
				then1 := cfg.CreateBlock("then1")
				then2 := cfg.CreateBlock("then2")
				then3 := cfg.CreateBlock("then3")
				else_block := cfg.CreateBlock("else")

				cfg.ConnectBlocks(cfg.Entry, cond1, EdgeNormal)
				cfg.ConnectBlocks(cond1, then1, EdgeCondTrue)
				cfg.ConnectBlocks(cond1, cond2, EdgeCondFalse)
				cfg.ConnectBlocks(cond2, then2, EdgeCondTrue)
				cfg.ConnectBlocks(cond2, cond3, EdgeCondFalse)
				cfg.ConnectBlocks(cond3, then3, EdgeCondTrue)
				cfg.ConnectBlocks(cond3, else_block, EdgeCondFalse)
				cfg.ConnectBlocks(then1, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(then2, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(then3, cfg.Exit, EdgeNormal)
				cfg.ConnectBlocks(else_block, cfg.Exit, EdgeNormal)

				return cfg
			},
			expectedComplexity: 4, // 3 decision points (cond1, cond2, cond3) + 1
			expectedRiskLevel:  "low",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := tc.setupCFG()
			result := CalculateComplexity(cfg)

			if result.Complexity != tc.expectedComplexity {
				t.Errorf("Expected complexity %d, got %d", tc.expectedComplexity, result.Complexity)
			}

			if result.RiskLevel != tc.expectedRiskLevel {
				t.Errorf("Expected risk level %q, got %q", tc.expectedRiskLevel, result.RiskLevel)
			}

			if tc.expectedNodes > 0 && result.Nodes != tc.expectedNodes {
				t.Errorf("Expected %d nodes, got %d", tc.expectedNodes, result.Nodes)
			}

			if tc.expectedEdges > 0 && result.Edges != tc.expectedEdges {
				t.Errorf("Expected %d edges, got %d", tc.expectedEdges, result.Edges)
			}
		})
	}
}

func TestAssessRiskLevel(t *testing.T) {
	// Test using config.ComplexityConfig.AssessRiskLevel instead of deprecated function
	defaultConfig := config.DefaultConfig()

	testCases := []struct {
		complexity int
		expected   string
	}{
		{1, "low"},
		{5, "low"},
		{9, "low"},
		{10, "medium"},
		{15, "medium"},
		{19, "medium"},
		{20, "high"},
		{25, "high"},
		{100, "high"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := defaultConfig.Complexity.AssessRiskLevel(tc.complexity)
			if result != tc.expected {
				t.Errorf("For complexity %d, expected %q, got %q", tc.complexity, tc.expected, result)
			}
		})
	}
}

func TestCalculateFileComplexity(t *testing.T) {
	t.Run("EmptySlice", func(t *testing.T) {
		results := CalculateFileComplexity([]*CFG{})
		if len(results) != 0 {
			t.Errorf("Expected empty results, got %d", len(results))
		}
	})

	t.Run("MultipleCFGs", func(t *testing.T) {
		cfg1 := NewCFG("function1")
		block1 := cfg1.CreateBlock("main")
		cfg1.ConnectBlocks(cfg1.Entry, block1, EdgeNormal)
		cfg1.ConnectBlocks(block1, cfg1.Exit, EdgeNormal)

		cfg2 := NewCFG("function2")
		cond := cfg2.CreateBlock("condition")
		then := cfg2.CreateBlock("then")
		cfg2.ConnectBlocks(cfg2.Entry, cond, EdgeNormal)
		cfg2.ConnectBlocks(cond, then, EdgeCondTrue)
		cfg2.ConnectBlocks(cond, cfg2.Exit, EdgeCondFalse)
		cfg2.ConnectBlocks(then, cfg2.Exit, EdgeNormal)

		cfgs := []*CFG{cfg1, cfg2, nil} // Include nil to test filtering
		results := CalculateFileComplexity(cfgs)

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results[0].Complexity != 1 {
			t.Errorf("Expected complexity 1 for first function, got %d", results[0].Complexity)
		}

		if results[1].Complexity != 2 {
			t.Errorf("Expected complexity 2 for second function, got %d", results[1].Complexity)
		}
	})
}

func TestCalculateAggregateComplexity(t *testing.T) {
	t.Run("EmptyResults", func(t *testing.T) {
		agg := CalculateAggregateComplexity([]*ComplexityResult{})
		if agg.TotalFunctions != 0 {
			t.Errorf("Expected 0 total functions, got %d", agg.TotalFunctions)
		}
	})

	t.Run("MultipleResults", func(t *testing.T) {
		results := []*ComplexityResult{
			{Complexity: 1, RiskLevel: "low"},
			{Complexity: 5, RiskLevel: "low"},
			{Complexity: 15, RiskLevel: "medium"},
			{Complexity: 25, RiskLevel: "high"},
		}

		agg := CalculateAggregateComplexity(results)

		if agg.TotalFunctions != 4 {
			t.Errorf("Expected 4 total functions, got %d", agg.TotalFunctions)
		}

		expectedAverage := 11.5 // (1+5+15+25)/4
		if agg.AverageComplexity != expectedAverage {
			t.Errorf("Expected average complexity %.1f, got %.1f", expectedAverage, agg.AverageComplexity)
		}

		if agg.MinComplexity != 1 {
			t.Errorf("Expected min complexity 1, got %d", agg.MinComplexity)
		}

		if agg.MaxComplexity != 25 {
			t.Errorf("Expected max complexity 25, got %d", agg.MaxComplexity)
		}

		if agg.LowRiskCount != 2 {
			t.Errorf("Expected 2 low risk functions, got %d", agg.LowRiskCount)
		}

		if agg.MediumRiskCount != 1 {
			t.Errorf("Expected 1 medium risk function, got %d", agg.MediumRiskCount)
		}

		if agg.HighRiskCount != 1 {
			t.Errorf("Expected 1 high risk function, got %d", agg.HighRiskCount)
		}
	})
}

func TestComplexityVisitor(t *testing.T) {
	t.Run("VisitBlock", func(t *testing.T) {
		visitor := &complexityVisitor{}

		// Test with regular block
		block := NewBasicBlock("test")
		if !visitor.VisitBlock(block) {
			t.Error("VisitBlock should return true")
		}
		if visitor.nodeCount != 1 {
			t.Errorf("Expected node count 1, got %d", visitor.nodeCount)
		}

		// Test with entry block (should not increment count)
		entryBlock := NewBasicBlock("entry")
		entryBlock.IsEntry = true
		visitor.VisitBlock(entryBlock)
		if visitor.nodeCount != 1 {
			t.Errorf("Entry block should not increment node count, got %d", visitor.nodeCount)
		}

		// Test with exit block (should not increment count)
		exitBlock := NewBasicBlock("exit")
		exitBlock.IsExit = true
		visitor.VisitBlock(exitBlock)
		if visitor.nodeCount != 1 {
			t.Errorf("Exit block should not increment node count, got %d", visitor.nodeCount)
		}

		// Test with nil block
		if !visitor.VisitBlock(nil) {
			t.Error("VisitBlock should handle nil gracefully")
		}
	})

	t.Run("VisitEdge", func(t *testing.T) {
		visitor := &complexityVisitor{
			decisionPoints: make(map[*BasicBlock]int),
		}
		block1 := NewBasicBlock("b1")
		block2 := NewBasicBlock("b2")

		// Test normal edge
		edge := &Edge{From: block1, To: block2, Type: EdgeNormal}
		if !visitor.VisitEdge(edge) {
			t.Error("VisitEdge should return true")
		}
		if visitor.edgeCount != 1 {
			t.Errorf("Expected edge count 1, got %d", visitor.edgeCount)
		}

		// Test conditional edges
		condTrueEdge := &Edge{From: block1, To: block2, Type: EdgeCondTrue}
		visitor.VisitEdge(condTrueEdge)
		if len(visitor.decisionPoints) != 1 {
			t.Errorf("Expected decision point count 1, got %d", len(visitor.decisionPoints))
		}

		condFalseEdge := &Edge{From: block1, To: block2, Type: EdgeCondFalse}
		visitor.VisitEdge(condFalseEdge)
		// Should still be 1 since it's the same source block
		if len(visitor.decisionPoints) != 1 {
			t.Errorf("Expected decision point count 1 (same block), got %d", len(visitor.decisionPoints))
		}

		// Test loop edge
		loopEdge := &Edge{From: block1, To: block2, Type: EdgeLoop}
		visitor.VisitEdge(loopEdge)
		if visitor.loopStatements != 1 {
			t.Errorf("Expected loop statement count 1, got %d", visitor.loopStatements)
		}

		// Test exception edge
		exceptionEdge := &Edge{From: block1, To: block2, Type: EdgeException}
		visitor.VisitEdge(exceptionEdge)
		if visitor.exceptionHandlers != 1 {
			t.Errorf("Expected exception handler count 1, got %d", visitor.exceptionHandlers)
		}

		// Test with nil edge
		if !visitor.VisitEdge(nil) {
			t.Error("VisitEdge should handle nil gracefully")
		}
	})
}

func TestCalculateComplexityWithConfig(t *testing.T) {
	// Create a test CFG
	cfg := NewCFG("test_function")
	condBlock := cfg.CreateBlock("condition")
	thenBlock := cfg.CreateBlock("then")
	elseBlock := cfg.CreateBlock("else")

	cfg.ConnectBlocks(cfg.Entry, condBlock, EdgeNormal)
	cfg.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
	cfg.ConnectBlocks(condBlock, elseBlock, EdgeCondFalse)
	cfg.ConnectBlocks(thenBlock, cfg.Exit, EdgeNormal)
	cfg.ConnectBlocks(elseBlock, cfg.Exit, EdgeNormal)

	t.Run("CustomThresholds", func(t *testing.T) {
		// Create custom config with different thresholds
		customConfig := &config.ComplexityConfig{
			LowThreshold:    3,
			MediumThreshold: 10,
			Enabled:         true,
		}

		result := CalculateComplexityWithConfig(cfg, customConfig)

		if result.Complexity != 2 {
			t.Errorf("Expected complexity 2, got %d", result.Complexity)
		}

		// With custom thresholds, complexity 2 should be low risk
		if result.RiskLevel != "low" {
			t.Errorf("Expected low risk with custom thresholds, got %s", result.RiskLevel)
		}
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		// Test that CalculateComplexity uses default config
		result1 := CalculateComplexity(cfg)

		defaultConfig := config.DefaultConfig()
		result2 := CalculateComplexityWithConfig(cfg, &defaultConfig.Complexity)

		if result1.Complexity != result2.Complexity {
			t.Errorf("Default and explicit config should give same complexity: %d vs %d",
				result1.Complexity, result2.Complexity)
		}

		if result1.RiskLevel != result2.RiskLevel {
			t.Errorf("Default and explicit config should give same risk level: %s vs %s",
				result1.RiskLevel, result2.RiskLevel)
		}
	})
}

func TestCalculateFileComplexityWithConfig(t *testing.T) {
	// Create test CFGs
	cfg1 := NewCFG("function1")
	block1 := cfg1.CreateBlock("main")
	cfg1.ConnectBlocks(cfg1.Entry, block1, EdgeNormal)
	cfg1.ConnectBlocks(block1, cfg1.Exit, EdgeNormal)

	cfg2 := NewCFG("function2")
	condBlock := cfg2.CreateBlock("condition")
	thenBlock := cfg2.CreateBlock("then")
	cfg2.ConnectBlocks(cfg2.Entry, condBlock, EdgeNormal)
	cfg2.ConnectBlocks(condBlock, thenBlock, EdgeCondTrue)
	cfg2.ConnectBlocks(condBlock, cfg2.Exit, EdgeCondFalse)
	cfg2.ConnectBlocks(thenBlock, cfg2.Exit, EdgeNormal)

	cfgs := []*CFG{cfg1, cfg2}

	t.Run("FilterUnchanged", func(t *testing.T) {
		// Config that doesn't report unchanged (complexity 1) functions
		customConfig := &config.ComplexityConfig{
			LowThreshold:    9,
			MediumThreshold: 19,
			Enabled:         true,
			ReportUnchanged: false,
		}

		results := CalculateFileComplexityWithConfig(cfgs, customConfig)

		// Should only get function2 (complexity 2), not function1 (complexity 1)
		if len(results) != 1 {
			t.Errorf("Expected 1 result (filtered unchanged), got %d", len(results))
		}

		if results[0].FunctionName != "function2" {
			t.Errorf("Expected function2 in results, got %s", results[0].FunctionName)
		}
	})

	t.Run("IncludeUnchanged", func(t *testing.T) {
		// Config that reports unchanged functions
		customConfig := &config.ComplexityConfig{
			LowThreshold:    9,
			MediumThreshold: 19,
			Enabled:         true,
			ReportUnchanged: true,
		}

		results := CalculateFileComplexityWithConfig(cfgs, customConfig)

		// Should get both functions
		if len(results) != 2 {
			t.Errorf("Expected 2 results (include unchanged), got %d", len(results))
		}
	})

	t.Run("DisabledAnalysis", func(t *testing.T) {
		// Config with analysis disabled
		customConfig := &config.ComplexityConfig{
			Enabled: false,
		}

		results := CalculateFileComplexityWithConfig(cfgs, customConfig)

		// Should get no results when disabled
		if len(results) != 0 {
			t.Errorf("Expected 0 results (disabled), got %d", len(results))
		}
	})

	t.Run("BackwardCompatibility", func(t *testing.T) {
		// Test that old function works the same as new with default config
		results1 := CalculateFileComplexity(cfgs)

		defaultConfig := config.DefaultConfig()
		results2 := CalculateFileComplexityWithConfig(cfgs, &defaultConfig.Complexity)

		if len(results1) != len(results2) {
			t.Errorf("Backward compatibility broken: %d vs %d results",
				len(results1), len(results2))
		}

		for i := range results1 {
			if results1[i].Complexity != results2[i].Complexity {
				t.Errorf("Backward compatibility broken for complexity at index %d", i)
			}
		}
	})
}

func TestComplexityConfigIntegration(t *testing.T) {
	// Create a CFG with medium complexity
	cfg := NewCFG("medium_complexity_function")
	current := cfg.Entry

	// Create 5 decision points (complexity = 6)
	for i := 0; i < 5; i++ {
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

	t.Run("CustomRiskThresholds", func(t *testing.T) {
		testCases := []struct {
			name            string
			lowThreshold    int
			mediumThreshold int
			expectedRisk    string
		}{
			{"VeryLowThresholds", 2, 4, "high"},
			{"LowThresholds", 5, 8, "medium"},
			{"HighThresholds", 10, 20, "low"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				customConfig := &config.ComplexityConfig{
					LowThreshold:    tc.lowThreshold,
					MediumThreshold: tc.mediumThreshold,
					Enabled:         true,
				}

				result := CalculateComplexityWithConfig(cfg, customConfig)

				if result.RiskLevel != tc.expectedRisk {
					t.Errorf("With thresholds low=%d, medium=%d, expected risk %s, got %s",
						tc.lowThreshold, tc.mediumThreshold, tc.expectedRisk, result.RiskLevel)
				}
			})
		}
	})
}
