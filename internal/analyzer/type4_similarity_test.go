package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestType4SimilarityScores(t *testing.T) {
	// Read test files
	testDir := "../../testdata/python/clones/type4"

	files := []string{
		"sum_iterative.py",
		"sum_recursive.py",
		"find_max_a.py",
		"find_max_b.py",
	}

	p := parser.New()
	ctx := context.Background()

	// Parse all files and extract functions
	type funcInfo struct {
		file string
		name string
		ast  *parser.Node
	}
	var functions []funcInfo

	for _, file := range files {
		path := filepath.Join(testDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Skipping %s: %v", file, err)
			continue
		}

		result, err := p.Parse(ctx, content)
		if err != nil {
			t.Logf("Skipping %s: parse error: %v", file, err)
			continue
		}

		// Extract function definitions
		for _, node := range result.AST.Body {
			if node.Type == parser.NodeFunctionDef {
				functions = append(functions, funcInfo{
					file: file,
					name: node.Name,
					ast:  node,
				})
			}
		}
	}

	t.Logf("Found %d functions", len(functions))

	// Create semantic analyzer with DFA
	analyzer := NewSemanticSimilarityAnalyzerWithDFA()

	// Compare all pairs
	t.Logf("\nSimilarity Matrix (with DFA):")
	for i := 0; i < len(functions); i++ {
		for j := i + 1; j < len(functions); j++ {
			f1 := &CodeFragment{ASTNode: functions[i].ast}
			f2 := &CodeFragment{ASTNode: functions[j].ast}

			sim := analyzer.ComputeSimilarity(f1, f2)
			t.Logf("  %s:%s vs %s:%s = %.3f",
				functions[i].file, functions[i].name,
				functions[j].file, functions[j].name,
				sim)
		}
	}

	// Also test without DFA
	analyzerNoDFA := NewSemanticSimilarityAnalyzer()
	t.Logf("\nSimilarity Matrix (CFG only):")
	for i := 0; i < len(functions); i++ {
		for j := i + 1; j < len(functions); j++ {
			f1 := &CodeFragment{ASTNode: functions[i].ast}
			f2 := &CodeFragment{ASTNode: functions[j].ast}

			sim := analyzerNoDFA.ComputeSimilarity(f1, f2)
			t.Logf("  %s:%s vs %s:%s = %.3f",
				functions[i].file, functions[i].name,
				functions[j].file, functions[j].name,
				sim)
		}
	}
}
