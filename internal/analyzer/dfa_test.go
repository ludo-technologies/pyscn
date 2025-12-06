package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefUseKind(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		tests := []struct {
			kind     DefUseKind
			expected string
		}{
			{DefKindAssign, "assign"},
			{DefKindParameter, "parameter"},
			{DefKindForTarget, "for_target"},
			{DefKindImport, "import"},
			{DefKindWithTarget, "with_target"},
			{DefKindExceptTarget, "except_target"},
			{DefKindAugmented, "augmented"},
			{UseKindRead, "read"},
			{UseKindCall, "call_arg"},
			{UseKindAttribute, "attribute"},
			{UseKindSubscript, "subscript"},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.expected, tt.kind.String())
		}
	})

	t.Run("IsDef", func(t *testing.T) {
		defKinds := []DefUseKind{
			DefKindAssign, DefKindParameter, DefKindForTarget,
			DefKindImport, DefKindWithTarget, DefKindExceptTarget, DefKindAugmented,
		}
		for _, k := range defKinds {
			assert.True(t, k.IsDef(), "expected %s to be a definition", k.String())
		}

		useKinds := []DefUseKind{UseKindRead, UseKindCall, UseKindAttribute, UseKindSubscript}
		for _, k := range useKinds {
			assert.False(t, k.IsDef(), "expected %s not to be a definition", k.String())
		}
	})

	t.Run("IsUse", func(t *testing.T) {
		useKinds := []DefUseKind{UseKindRead, UseKindCall, UseKindAttribute, UseKindSubscript, DefKindAugmented}
		for _, k := range useKinds {
			assert.True(t, k.IsUse(), "expected %s to be a use", k.String())
		}

		nonUseKinds := []DefUseKind{DefKindAssign, DefKindParameter, DefKindForTarget, DefKindImport}
		for _, k := range nonUseKinds {
			assert.False(t, k.IsUse(), "expected %s not to be a use", k.String())
		}
	})
}

func TestVarReference(t *testing.T) {
	t.Run("NewVarReference", func(t *testing.T) {
		block := NewBasicBlock("bb0")
		ref := NewVarReference("x", DefKindAssign, block, nil, 0)

		require.NotNil(t, ref)
		assert.Equal(t, "x", ref.Name)
		assert.Equal(t, DefKindAssign, ref.Kind)
		assert.Equal(t, block, ref.Block)
		assert.Equal(t, 0, ref.Position)
	})
}

func TestDefUsePair(t *testing.T) {
	t.Run("NewDefUsePair", func(t *testing.T) {
		block := NewBasicBlock("bb0")
		def := NewVarReference("x", DefKindAssign, block, nil, 0)
		use := NewVarReference("x", UseKindRead, block, nil, 1)

		pair := NewDefUsePair(def, use)

		require.NotNil(t, pair)
		assert.Equal(t, def, pair.Def)
		assert.Equal(t, use, pair.Use)
	})

	t.Run("IsCrossBlock_SameBlock", func(t *testing.T) {
		block := NewBasicBlock("bb0")
		def := NewVarReference("x", DefKindAssign, block, nil, 0)
		use := NewVarReference("x", UseKindRead, block, nil, 1)
		pair := NewDefUsePair(def, use)

		assert.False(t, pair.IsCrossBlock())
	})

	t.Run("IsCrossBlock_DifferentBlocks", func(t *testing.T) {
		block1 := NewBasicBlock("bb0")
		block2 := NewBasicBlock("bb1")
		def := NewVarReference("x", DefKindAssign, block1, nil, 0)
		use := NewVarReference("x", UseKindRead, block2, nil, 0)
		pair := NewDefUsePair(def, use)

		assert.True(t, pair.IsCrossBlock())
	})

	t.Run("IsCrossBlock_NilRefs", func(t *testing.T) {
		pair := NewDefUsePair(nil, nil)
		assert.False(t, pair.IsCrossBlock())
	})
}

func TestDefUseChain(t *testing.T) {
	t.Run("NewDefUseChain", func(t *testing.T) {
		chain := NewDefUseChain("x")

		require.NotNil(t, chain)
		assert.Equal(t, "x", chain.Variable)
		assert.Empty(t, chain.Defs)
		assert.Empty(t, chain.Uses)
		assert.Empty(t, chain.Pairs)
	})

	t.Run("AddDef", func(t *testing.T) {
		chain := NewDefUseChain("x")
		block := NewBasicBlock("bb0")
		def := NewVarReference("x", DefKindAssign, block, nil, 0)

		chain.AddDef(def)

		assert.Len(t, chain.Defs, 1)
		assert.Equal(t, def, chain.Defs[0])
	})

	t.Run("AddUse", func(t *testing.T) {
		chain := NewDefUseChain("x")
		block := NewBasicBlock("bb0")
		use := NewVarReference("x", UseKindRead, block, nil, 0)

		chain.AddUse(use)

		assert.Len(t, chain.Uses, 1)
		assert.Equal(t, use, chain.Uses[0])
	})

	t.Run("AddPair", func(t *testing.T) {
		chain := NewDefUseChain("x")
		block := NewBasicBlock("bb0")
		def := NewVarReference("x", DefKindAssign, block, nil, 0)
		use := NewVarReference("x", UseKindRead, block, nil, 1)
		pair := NewDefUsePair(def, use)

		chain.AddPair(pair)

		assert.Len(t, chain.Pairs, 1)
		assert.Equal(t, pair, chain.Pairs[0])
	})

	t.Run("AddNil", func(t *testing.T) {
		chain := NewDefUseChain("x")
		chain.AddDef(nil)
		chain.AddUse(nil)
		chain.AddPair(nil)

		assert.Empty(t, chain.Defs)
		assert.Empty(t, chain.Uses)
		assert.Empty(t, chain.Pairs)
	})
}

func TestDFAInfo(t *testing.T) {
	t.Run("NewDFAInfo", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)

		require.NotNil(t, info)
		assert.Equal(t, cfg, info.CFG)
		assert.NotNil(t, info.Chains)
		assert.NotNil(t, info.BlockDefs)
		assert.NotNil(t, info.BlockUses)
	})

	t.Run("GetChain", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)

		chain1 := info.GetChain("x")
		chain2 := info.GetChain("x")
		chain3 := info.GetChain("y")

		assert.Same(t, chain1, chain2) // Same chain for same variable
		assert.NotSame(t, chain1, chain3)
		assert.Equal(t, "x", chain1.Variable)
		assert.Equal(t, "y", chain3.Variable)
	})

	t.Run("AddDef", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block := cfg.CreateBlock("body")
		def := NewVarReference("x", DefKindAssign, block, nil, 0)

		info.AddDef(def)

		assert.Equal(t, 1, info.TotalDefs())
		assert.Len(t, info.BlockDefs[block.ID], 1)
		assert.Len(t, info.Chains["x"].Defs, 1)
	})

	t.Run("AddUse", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block := cfg.CreateBlock("body")
		use := NewVarReference("x", UseKindRead, block, nil, 0)

		info.AddUse(use)

		assert.Equal(t, 1, info.TotalUses())
		assert.Len(t, info.BlockUses[block.ID], 1)
		assert.Len(t, info.Chains["x"].Uses, 1)
	})

	t.Run("Metrics", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block := cfg.CreateBlock("body")

		// Add definitions and uses for two variables
		info.AddDef(NewVarReference("x", DefKindAssign, block, nil, 0))
		info.AddDef(NewVarReference("y", DefKindAssign, block, nil, 1))
		info.AddUse(NewVarReference("x", UseKindRead, block, nil, 2))
		info.AddUse(NewVarReference("y", UseKindRead, block, nil, 3))
		info.AddUse(NewVarReference("x", UseKindRead, block, nil, 4))

		assert.Equal(t, 2, info.TotalDefs())
		assert.Equal(t, 3, info.TotalUses())
		assert.Equal(t, 2, info.UniqueVariables())
	})
}

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
		def2 := NewVarReference("y", DefKindParameter, block1, nil, 1)
		use1 := NewVarReference("x", UseKindRead, block1, nil, 2)
		use2 := NewVarReference("x", UseKindRead, block2, nil, 0)
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
		assert.Equal(t, 1, features.DefKindCounts[DefKindParameter])
		assert.Equal(t, 2, features.UseKindCounts[UseKindRead])
		assert.Equal(t, 1, features.UseKindCounts[UseKindCall])
	})

	t.Run("ExtractDFAFeatures_CrossBlockPairs", func(t *testing.T) {
		cfg := NewCFG("test")
		info := NewDFAInfo(cfg)
		block1 := cfg.CreateBlock("body")
		block2 := cfg.CreateBlock("body2")

		def := NewVarReference("x", DefKindAssign, block1, nil, 0)
		use1 := NewVarReference("x", UseKindRead, block1, nil, 1)
		use2 := NewVarReference("x", UseKindRead, block2, nil, 0)

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
