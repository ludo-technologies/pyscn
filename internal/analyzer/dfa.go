package analyzer

import coredfa "github.com/ludo-technologies/polyscan/core/dfa"

// Def-use chain data structures and the DFA builder are owned by polyscan
// core. Aliases keep pyscn's internal API stable while Python-specific
// reference extraction stays local (see dfa_builder.go).
type (
	DefUseKind   = coredfa.DefUseKind
	VarReference = coredfa.VarReference
	DefUsePair   = coredfa.DefUsePair
	DefUseChain  = coredfa.DefUseChain
	DFAInfo      = coredfa.DFAInfo
	DFABuilder   = coredfa.DFABuilder
)

const (
	DefKindAssign    = coredfa.DefKindAssign
	DefKindAugAssign = coredfa.DefKindAugAssign
	DefKindParam     = coredfa.DefKindParam
	DefKindImport    = coredfa.DefKindImport
	DefKindFor       = coredfa.DefKindFor
	DefKindWith      = coredfa.DefKindWith
	DefKindExcept    = coredfa.DefKindExcept
	DefKindPattern   = coredfa.DefKindPattern
	UseKindLoad      = coredfa.UseKindLoad
	UseKindCall      = coredfa.UseKindCall
	UseKindAttribute = coredfa.UseKindAttribute
	UseKindSubscript = coredfa.UseKindSubscript
)

var (
	NewVarReference = coredfa.NewVarReference
	NewDefUsePair   = coredfa.NewDefUsePair
	NewDefUseChain  = coredfa.NewDefUseChain
	NewDFAInfo      = coredfa.NewDFAInfo
)

// DFAFeatures captures data flow characteristics for clone comparison.
// pyscn keeps its own feature shape: AvgChainLength is pairs per definition,
// which differs from core's per-chain average, and Type-4 similarity scores
// depend on it.
type DFAFeatures struct {
	TotalDefs       int // Total number of definitions
	TotalUses       int // Total number of uses
	TotalPairs      int // Total number of def-use pairs
	UniqueVariables int // Number of unique variables

	AvgChainLength  float64 // Average uses per definition
	MaxChainLength  int     // Maximum def-use chain length
	CrossBlockPairs int     // Def-use pairs spanning blocks
	IntraBlockPairs int     // Def-use pairs within same block

	DefKindCounts map[DefUseKind]int // Distribution of definition kinds
	UseKindCounts map[DefUseKind]int // Distribution of use kinds
}

// NewDFAFeatures creates a new DFA features instance
func NewDFAFeatures() *DFAFeatures {
	return &DFAFeatures{
		DefKindCounts: make(map[DefUseKind]int),
		UseKindCounts: make(map[DefUseKind]int),
	}
}

// ExtractDFAFeatures extracts DFA features from DFAInfo
func ExtractDFAFeatures(info *DFAInfo) *DFAFeatures {
	if info == nil {
		return NewDFAFeatures()
	}

	features := NewDFAFeatures()
	features.TotalDefs = info.TotalDefs()
	features.TotalUses = info.TotalUses()
	features.TotalPairs = info.TotalPairs()
	features.UniqueVariables = info.UniqueVariables()

	// Calculate chain metrics and count pairs by type
	totalChainLength := 0
	for _, chain := range info.Chains {
		chainLength := len(chain.Pairs)
		totalChainLength += chainLength
		if chainLength > features.MaxChainLength {
			features.MaxChainLength = chainLength
		}

		// Count cross-block vs intra-block pairs
		for _, pair := range chain.Pairs {
			if pair.IsCrossBlock() {
				features.CrossBlockPairs++
			} else {
				features.IntraBlockPairs++
			}
		}

		// Count definition kinds
		for _, def := range chain.Defs {
			features.DefKindCounts[def.Kind]++
		}

		// Count use kinds
		for _, use := range chain.Uses {
			features.UseKindCounts[use.Kind]++
		}
	}

	// Calculate average chain length
	if features.TotalDefs > 0 {
		features.AvgChainLength = float64(totalChainLength) / float64(features.TotalDefs)
	}

	return features
}
