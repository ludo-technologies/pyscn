package analyzer

import (
	"testing"

	coresemantic "github.com/ludo-technologies/polyscan/core/semantic"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/require"
)

func TestSemanticEvidence_CallOverlap(t *testing.T) {
	tests := []struct {
		name  string
		left  string
		right string
		clone bool
	}{
		{
			name:  "shared operators cannot rescue unrelated calls",
			left:  "result = encrypt(item)\n        result = encode(result)",
			right: "result = checksum(item)\n        result = format_address(result)",
		},
		{
			name:  "one common method cannot rescue unrelated algorithms",
			left:  "result = item.bit_length()\n        result = search_block(result)\n        result = read_block(result)",
			right: "result = item.bit_length()\n        result = continued_fraction(result)\n        result = recover_key(result)",
		},
		{
			name:  "common startup methods cannot rescue unrelated work",
			left:  "result = item.start()\n        result = result.proxy()\n        result = encrypt(result)\n        result = encode(result)\n        result = transmit(result)",
			right: "result = item.start()\n        result = result.proxy()\n        result = query(result)\n        result = aggregate(result)\n        result = render(result)",
		},
		{
			name:  "renamed variables with shared operations remain clones",
			left:  "result = normalize(item)\n        result = result.digest()",
			right: "value = normalize(item)\n        result = value.digest()",
			clone: true,
		},
		{
			name:  "call-free arithmetic remains comparable",
			left:  "result = item + 1",
			right: "result = 1 + item",
			clone: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fragment := func(body string) *CodeFragment {
				return fragmentFor(t, "def process(items):\n    result = 0\n    for item in items:\n        "+body+"\n        result = result + 1\n    return result\n", parser.NodeFunctionDef)
			}
			left, right := fragment(tt.left), fragment(tt.right)
			for name, analyzer := range map[string]*SemanticSimilarityAnalyzer{
				"cfg": NewSemanticSimilarityAnalyzer(),
				"dfa": NewSemanticSimilarityAnalyzerWithDFA(),
			} {
				t.Run(name, func(t *testing.T) {
					score := analyzer.ComputeSimilarity(left, right)
					require.Equal(t, score, analyzer.ComputeSimilarity(right, left))
					if tt.clone {
						require.GreaterOrEqual(t, score, domain.DefaultType4CloneThreshold)
					} else {
						require.Less(t, score, domain.DefaultType4CloneThreshold)
					}
				})
			}
		})
	}
}

func TestSemanticOperationWeight(t *testing.T) {
	tests := []struct {
		name        string
		left, right []string
		want        float64
	}{
		{"disjoint calls despite shared operator", []string{"call:encrypt", "binop:+"}, []string{"call:checksum", "binop:+"}, 0.5},
		{"mostly shared calls", []string{"call:normalize", "method:digest"}, []string{"call:normalize", "method:digest", "call:validate"}, 0.5 + 0.5*2/3},
		{"one call-free side with shared operator", []string{"binop:+"}, []string{"call:sum_numbers", "binop:+"}, 1},
		{"disjoint call-free operators", []string{"binop:+"}, []string{"compare:>"}, 0.5},
		{"no evidence", nil, nil, 1},
		{"one empty side", nil, []string{"call:normalize"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signals := func(values []string) coresemantic.SemanticSignals {
				result := coresemantic.NewSemanticSignals()
				for _, value := range values {
					result.StrongSignals[value] = struct{}{}
				}
				return result
			}
			left, right := signals(tt.left), signals(tt.right)
			require.InDelta(t, tt.want, semanticOperationWeight(left, right), 1e-10)
			require.InDelta(t, tt.want, semanticOperationWeight(right, left), 1e-10)
		})
	}
}
