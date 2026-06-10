package analyzer

import "testing"

// TestCalculateComplexityElifChain pins cyclomatic complexity for elif chains.
//
// Each if/elif condition is a decision point, so "if + N elif [+ else]" has
// cyclomatic complexity N+2 (N+1 conditions plus the base of 1). Regression
// guard for #511: the AST builder kept only the last elif of a chain, so every
// elif but the last was dropped and complexity was undercounted on long
// dispatch chains. The existing MultipleElif CFG test only asserted
// cfg.Size() >= 5, so it stayed green while branches went missing.
func TestCalculateComplexityElifChain(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{
			name: "if_else",
			src:  "\nif x > 1:\n    print(1)\nelse:\n    print(0)\n",
			want: 2,
		},
		{
			name: "if_one_elif_else",
			src:  "\nif x > 1:\n    print(1)\nelif x > 2:\n    print(2)\nelse:\n    print(0)\n",
			want: 3,
		},
		{
			name: "if_three_elif_else",
			src:  "\nif x > 100:\n    print(1)\nelif x > 50:\n    print(2)\nelif x > 10:\n    print(3)\nelif x > 0:\n    print(4)\nelse:\n    print(0)\n",
			want: 5,
		},
		{
			name: "if_five_elif_no_else",
			src:  "\nif x == 1:\n    print(1)\nelif x == 2:\n    print(2)\nelif x == 3:\n    print(3)\nelif x == 4:\n    print(4)\nelif x == 5:\n    print(5)\nelif x == 6:\n    print(6)\n",
			want: 7,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ast := parseSource(t, c.src)
			cfg, err := NewCFGBuilder().Build(ast)
			if err != nil {
				t.Fatalf("Build CFG: %v", err)
			}
			got := CalculateComplexity(cfg).Complexity
			if got != c.want {
				t.Errorf("cyclomatic complexity = %d, want %d", got, c.want)
			}
		})
	}
}
