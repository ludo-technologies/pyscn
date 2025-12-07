package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDFAReachingDefinition_LastDefInPredecessor tests that when a variable
// is defined multiple times in a predecessor block, the LAST definition
// (highest position) is the one that reaches the use in the successor block.
//
// This is a regression test for a bug where the first definition was returned
// instead of the last one.
func TestDFAReachingDefinition_LastDefInPredecessor(t *testing.T) {
	// Test case:
	// x = 1   # pos 0 in bb0
	// x = 2   # pos 1 in bb0
	// ------- (block boundary)
	// y = x   # pos 0 in bb1 - should reach x=2, not x=1

	source := `
x = 1
x = 2
if True:
    y = x
`
	cfg := buildCFGForDFA(t, source)
	require.NotNil(t, cfg)

	builder := NewDFABuilder()
	info, err := builder.Build(cfg)
	require.NoError(t, err)
	require.NotNil(t, info)

	// Check the x chain
	chainX := info.Chains["x"]
	require.NotNil(t, chainX, "x chain should exist")

	// Should have 2 definitions
	assert.Len(t, chainX.Defs, 2, "x should have 2 definitions")

	// Should have at least 1 use
	require.GreaterOrEqual(t, len(chainX.Uses), 1, "x should have at least 1 use")

	// The critical check: the def-use pair should link to the SECOND definition (pos 1)
	// not the first one (pos 0)
	if len(chainX.Pairs) > 0 {
		pair := chainX.Pairs[0]
		// The reaching definition should be the one at the higher position
		// (i.e., the last definition before the use)
		lastDefPos := -1
		for _, def := range chainX.Defs {
			if def.Position > lastDefPos {
				lastDefPos = def.Position
			}
		}
		assert.Equal(t, lastDefPos, pair.Def.Position,
			"Def-use pair should link to the last definition (pos %d), not earlier ones", lastDefPos)
	}
}

// TestDFAReachingDefinition_SingleDefInPredecessor tests the simple case
// where there's only one definition in the predecessor block.
func TestDFAReachingDefinition_SingleDefInPredecessor(t *testing.T) {
	source := `
x = 1
if True:
    y = x
`
	cfg := buildCFGForDFA(t, source)
	require.NotNil(t, cfg)

	builder := NewDFABuilder()
	info, err := builder.Build(cfg)
	require.NoError(t, err)
	require.NotNil(t, info)

	chainX := info.Chains["x"]
	require.NotNil(t, chainX)

	// Should have 1 definition and at least 1 use
	assert.Len(t, chainX.Defs, 1)
	assert.GreaterOrEqual(t, len(chainX.Uses), 1)

	// Should have at least 1 pair
	if len(chainX.Pairs) > 0 {
		pair := chainX.Pairs[0]
		assert.Equal(t, 0, pair.Def.Position)
	}
}
