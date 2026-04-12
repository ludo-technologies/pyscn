package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneService_convertClonePairsToDomain_Content(t *testing.T) {
	service := NewCloneService()
	pairs := []*analyzer.ClonePair{
		{
			Fragment1: &analyzer.CodeFragment{
				Location:  &analyzer.CodeLocation{FilePath: "a.py", StartLine: 1, EndLine: 2},
				Content:   "def a():\n    return 1",
				Size:      4,
				LineCount: 2,
			},
			Fragment2: &analyzer.CodeFragment{
				Location:  &analyzer.CodeLocation{FilePath: "b.py", StartLine: 3, EndLine: 4},
				Content:   "def b():\n    return 1",
				Size:      4,
				LineCount: 2,
			},
			Similarity: 0.95,
			CloneType:  analyzer.Type1Clone,
		},
	}

	withContent := service.convertClonePairsToDomain(pairs, true)
	require.Len(t, withContent, 1)
	assert.Equal(t, "def a():\n    return 1", withContent[0].Clone1.Content)
	assert.Equal(t, "def b():\n    return 1", withContent[0].Clone2.Content)

	withoutContent := service.convertClonePairsToDomain(pairs, false)
	require.Len(t, withoutContent, 1)
	assert.Empty(t, withoutContent[0].Clone1.Content)
	assert.Empty(t, withoutContent[0].Clone2.Content)
}

func TestCloneService_convertCloneGroupsToDomain_Content(t *testing.T) {
	service := NewCloneService()
	groups := []*analyzer.CloneGroup{
		{
			ID:         1,
			CloneType:  analyzer.Type2Clone,
			Similarity: 0.91,
			Size:       2,
			Fragments: []*analyzer.CodeFragment{
				{
					Location:  &analyzer.CodeLocation{FilePath: "a.py", StartLine: 1, EndLine: 2},
					Content:   "value = 1\nprint(value)",
					Size:      3,
					LineCount: 2,
				},
				{
					Location:  &analyzer.CodeLocation{FilePath: "b.py", StartLine: 5, EndLine: 6},
					Content:   "count = 1\nprint(count)",
					Size:      3,
					LineCount: 2,
				},
			},
		},
	}

	withContent := service.convertCloneGroupsToDomain(groups, true)
	require.Len(t, withContent, 1)
	require.Len(t, withContent[0].Clones, 2)
	assert.Equal(t, "value = 1\nprint(value)", withContent[0].Clones[0].Content)
	assert.Equal(t, "count = 1\nprint(count)", withContent[0].Clones[1].Content)

	withoutContent := service.convertCloneGroupsToDomain(groups, false)
	require.Len(t, withoutContent, 1)
	require.Len(t, withoutContent[0].Clones, 2)
	assert.Empty(t, withoutContent[0].Clones[0].Content)
	assert.Empty(t, withoutContent[0].Clones[1].Content)
}
