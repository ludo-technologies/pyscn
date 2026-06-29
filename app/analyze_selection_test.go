package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyAnalyzeSelection_DefaultLeavesAnalyzersEnabled(t *testing.T) {
	config := ApplyAnalyzeSelection(AnalyzeUseCaseConfig{}, nil)

	assert.False(t, config.SelectAnalysesUsed)
	assert.False(t, config.SkipComplexity)
	assert.False(t, config.SkipDeadCode)
	assert.False(t, config.SkipClones)
	assert.False(t, config.SkipSystem)
	assert.False(t, config.SkipCommunities)
}

func TestApplyAnalyzeSelection_CLIAnalyzerNames(t *testing.T) {
	config := ApplyAnalyzeSelection(AnalyzeUseCaseConfig{}, []string{"complexity", "deadcode", "clones", "deps", "communities"})

	assert.True(t, config.SelectAnalysesUsed)
	assert.False(t, config.SkipComplexity)
	assert.False(t, config.SkipDeadCode)
	assert.False(t, config.SkipClones)
	assert.False(t, config.SkipSystem)
	assert.False(t, config.SkipCommunities)
	assert.True(t, config.SkipCBO)
	assert.True(t, config.SkipLCOM)
}

func TestApplyAnalyzeSelection_MCPAnalyzerAliases(t *testing.T) {
	config := ApplyAnalyzeSelection(AnalyzeUseCaseConfig{}, []string{"dead_code", "clone", "communities"})

	assert.True(t, config.SelectAnalysesUsed)
	assert.False(t, config.SkipDeadCode)
	assert.False(t, config.SkipClones)
	assert.False(t, config.SkipCommunities)
	assert.True(t, config.SkipComplexity)
	assert.True(t, config.SkipSystem)
}
