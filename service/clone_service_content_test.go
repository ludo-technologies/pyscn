package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
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

	withContent := service.convertCloneGroupsToDomain(groups, true, nil)
	require.Len(t, withContent, 1)
	require.Len(t, withContent[0].Clones, 2)
	assert.Equal(t, "value = 1\nprint(value)", withContent[0].Clones[0].Content)
	assert.Equal(t, "count = 1\nprint(count)", withContent[0].Clones[1].Content)

	withoutContent := service.convertCloneGroupsToDomain(groups, false, nil)
	require.Len(t, withoutContent, 1)
	require.Len(t, withoutContent[0].Clones, 2)
	assert.Empty(t, withoutContent[0].Clones[0].Content)
	assert.Empty(t, withoutContent[0].Clones[1].Content)
}

// Regression for #488 (fragment id/type were always 0) plus the #503 review:
// a group fragment must surface its response-wide id (the same id it carries in
// top-level clones[]) instead of a per-group counter, so ids stay unique across
// the response and link back to clones[].
func TestCloneService_convertCloneGroupsToDomain_FragmentIDAndType(t *testing.T) {
	service := NewCloneService()
	frag0 := &analyzer.CodeFragment{
		Location:  &analyzer.CodeLocation{FilePath: "a.py", StartLine: 1, EndLine: 2},
		Content:   "value = 1\nprint(value)",
		Size:      3,
		LineCount: 2,
	}
	frag1 := &analyzer.CodeFragment{
		Location:  &analyzer.CodeLocation{FilePath: "b.py", StartLine: 5, EndLine: 6},
		Content:   "count = 1\nprint(count)",
		Size:      3,
		LineCount: 2,
	}
	groups := []*analyzer.CloneGroup{
		{
			ID:         1,
			CloneType:  analyzer.Type2Clone,
			Similarity: 0.91,
			Size:       2,
			Fragments:  []*analyzer.CodeFragment{frag0, frag1},
		},
	}

	// Non-sequential, response-wide ids (as if these were clones #3 and #7 in
	// the top-level clones[]). The group must surface these exact ids, not a
	// per-group 1,2 reset.
	fragmentIDs := map[*analyzer.CodeFragment]int{frag0: 3, frag1: 7}

	result := service.convertCloneGroupsToDomain(groups, false, fragmentIDs)
	require.Len(t, result, 1)
	require.Len(t, result[0].Clones, 2)

	// Fragment ids match their response-wide ids (link to top-level clones[]),
	// not a per-group reset.
	assert.Equal(t, 3, result[0].Clones[0].ID)
	assert.Equal(t, 7, result[0].Clones[1].ID)

	// Each fragment inherits the group's clone type instead of the zero value.
	for _, clone := range result[0].Clones {
		assert.Equal(t, domain.Type2Clone, clone.Type, "fragment type should match the group clone type")
	}
}
