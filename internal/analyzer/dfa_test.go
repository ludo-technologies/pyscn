package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDFAFeatures(t *testing.T) {
	t.Run("NewDFAFeatures", func(t *testing.T) {
		features := NewDFAFeatures()

		require.NotNil(t, features)
		assert.NotNil(t, features.DefKindCounts)
		assert.NotNil(t, features.UseKindCounts)
		assert.Equal(t, 0, features.TotalDefs)
		assert.Equal(t, 0, features.TotalPairs)
	})

	t.Run("ExtractDFAFeatures_Nil", func(t *testing.T) {
		features := ExtractDFAFeatures(nil)

		require.NotNil(t, features)
		assert.Equal(t, 0, features.TotalDefs)
	})

	t.Run("ExtractDFAFeatures_Basic", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block1 := cfg.CreateBlock("body")
		block2 := cfg.CreateBlock("body2")

		// Add defs and uses
		def1 := NewVarReference("x", DefKindAssign, block1, nil, 0)
		def2 := NewVarReference("y", DefKindParam, block1, nil, 1)
		use1 := NewVarReference("x", UseKindLoad, block1, nil, 2)
		use2 := NewVarReference("x", UseKindLoad, block2, nil, 0)
		use3 := NewVarReference("y", UseKindCall, block2, nil, 1)

		info.AddDef(def1)
		info.AddDef(def2)
		info.AddUse(use1)
		info.AddUse(use2)
		info.AddUse(use3)

		// Add pairs manually for testing
		chainX := info.GetChain("x")
		chainX.AddPair(NewDefUsePair(def1, use1))
		chainX.AddPair(NewDefUsePair(def1, use2))

		chainY := info.GetChain("y")
		chainY.AddPair(NewDefUsePair(def2, use3))

		features := ExtractDFAFeatures(info)

		assert.Equal(t, 2, features.TotalDefs)
		assert.Equal(t, 3, features.TotalUses)
		assert.Equal(t, 3, features.TotalPairs)
		assert.Equal(t, 2, features.UniqueVariables)
		assert.Equal(t, 2, features.MaxChainLength) // x has 2 pairs
		assert.Equal(t, 1, features.DefKindCounts[DefKindAssign])
		assert.Equal(t, 1, features.DefKindCounts[DefKindParam])
		assert.Equal(t, 2, features.UseKindCounts[UseKindLoad])
		assert.Equal(t, 1, features.UseKindCounts[UseKindCall])
	})

	t.Run("ExtractDFAFeatures_CrossBlockPairs", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block1 := cfg.CreateBlock("body")
		block2 := cfg.CreateBlock("body2")

		def := NewVarReference("x", DefKindAssign, block1, nil, 0)
		use1 := NewVarReference("x", UseKindLoad, block1, nil, 1)
		use2 := NewVarReference("x", UseKindLoad, block2, nil, 0)

		info.AddDef(def)
		info.AddUse(use1)
		info.AddUse(use2)

		chain := info.GetChain("x")
		chain.AddPair(NewDefUsePair(def, use1)) // intra-block
		chain.AddPair(NewDefUsePair(def, use2)) // cross-block

		features := ExtractDFAFeatures(info)

		assert.Equal(t, 1, features.IntraBlockPairs)
		assert.Equal(t, 1, features.CrossBlockPairs)
	})
}
