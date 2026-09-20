package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneGroupsRetainOnlyConnectedEnabledMembers(t *testing.T) {
	s := NewCloneService()
	clones := []*domain.Clone{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}}
	pairs := []*domain.ClonePair{
		{Clone1: clones[0], Clone2: clones[1], Type: domain.Type4Clone, Similarity: 0.85},
		{Clone1: clones[1], Clone2: clones[2], Type: domain.Type3Clone, Similarity: 0.75},
		{Clone1: clones[2], Clone2: clones[3], Type: domain.Type4Clone, Similarity: 0.8},
	}
	for _, tt := range []struct {
		name  string
		pairs []*domain.ClonePair
		types []domain.CloneType
		want  [][]int
	}{
		{"orphan", pairs[:2], domain.DefaultEnabledCloneTypes, [][]int{{1, 2}}},
		{"disabled bridge", pairs, domain.DefaultEnabledCloneTypes, [][]int{{1, 2}, {3, 4}}},
		{"all enabled", pairs, []domain.CloneType{domain.Type3Clone, domain.Type4Clone}, [][]int{{1, 2, 3, 4}}},
		{"all disabled", pairs, []domain.CloneType{domain.Type1Clone}, nil},
		{"external edge", []*domain.ClonePair{{Clone1: clones[0], Clone2: &domain.Clone{ID: 5}, Type: domain.Type4Clone}}, domain.DefaultEnabledCloneTypes, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			group := &domain.CloneGroup{ID: 7, Clones: clones, Size: 4, Type: domain.Type4Clone, Similarity: 0.8}
			req := newDefaultCloneRequest()
			req.CloneTypes = tt.types
			retained := s.filterClonePairsByType(tt.pairs, req)
			groups := filterCloneGroupsByPairs([]*domain.CloneGroup{group}, retained)
			require.Len(t, groups, len(tt.want))
			ids := make(map[int]bool)
			for i, got := range groups {
				var members []int
				for _, clone := range got.Clones {
					members = append(members, clone.ID)
				}
				assert.Equal(t, tt.want[i], members)
				assert.Equal(t, len(members), got.Size)
				assert.False(t, ids[got.ID], "group IDs must be unique")
				ids[got.ID] = true
			}
			assert.Len(t, group.Clones, 4, "filtering must not mutate the scored population")
			assert.Equal(t, 4, group.Size)
		})
	}
}

func TestCloneGroupMetadataUsesRetainedPairs(t *testing.T) {
	clones := []*domain.Clone{{ID: 1, Type: domain.Type3Clone}, {ID: 2}, {ID: 3}, {ID: 4}}
	parent := &domain.CloneGroup{ID: 7, Clones: clones, Size: 4, Type: domain.Type3Clone, Similarity: 0.82}
	pairs := []*domain.ClonePair{
		{Clone1: clones[0], Clone2: clones[1], Type: domain.Type1Clone, Similarity: 1.0},
		{Clone1: clones[2], Clone2: clones[3], Type: domain.Type4Clone, Similarity: 0.70},
	}
	groups := filterCloneGroupsByPairs([]*domain.CloneGroup{parent}, pairs)
	require.Len(t, groups, 2)
	assert.Equal(t, domain.Type1Clone, groups[0].Type)
	assert.Equal(t, 1.0, groups[0].Similarity)
	assert.Equal(t, domain.Type4Clone, groups[1].Type)
	assert.Equal(t, 0.70, groups[1].Similarity)
	assert.Equal(t, domain.Type1Clone, groups[0].Clones[0].Type)
	assert.Equal(t, domain.Type3Clone, parent.Type)
	assert.Equal(t, 0.82, parent.Similarity)
	assert.Equal(t, domain.Type3Clone, clones[0].Type)

	s := NewCloneService()
	for _, tt := range []struct {
		name     string
		min, max float64
		wantType domain.CloneType
	}{
		{"high similarity", 0.9, 1.0, domain.Type1Clone},
		{"low similarity", 0.6, 0.75, domain.Type4Clone},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := newDefaultCloneRequest()
			req.MinSimilarity, req.MaxSimilarity = tt.min, tt.max
			visible := s.filterCloneGroupsBySimilarity(groups, req)
			require.Len(t, visible, 1)
			assert.Equal(t, tt.wantType, visible[0].Type)
		})
	}
}

func TestCloneGroupMetadataRefreshesWithoutMemberRemoval(t *testing.T) {
	clones := []*domain.Clone{{ID: 1}, {ID: 2}, {ID: 3}}
	parent := &domain.CloneGroup{Clones: clones, Size: 3, Type: domain.Type3Clone, Similarity: 0.82}
	pairs := []*domain.ClonePair{
		{Clone1: clones[0], Clone2: clones[1], Type: domain.Type1Clone, Similarity: 1.0},
		{Clone1: clones[1], Clone2: clones[2], Type: domain.Type4Clone, Similarity: 0.70},
	}
	groups := filterCloneGroupsByPairs([]*domain.CloneGroup{parent}, pairs)
	require.Len(t, groups, 1)
	assert.InDelta(t, 0.85, groups[0].Similarity, 1e-9)
	assert.Equal(t, domain.Type1Clone, groups[0].Type, "type comes from the highest-similarity pair")

	// Keep the full scored group independent of the displayed group's refresh.
	visible := filterCloneGroupsByPairs(groups, pairs[1:])
	require.Len(t, visible, 1)
	assert.Equal(t, domain.Type4Clone, visible[0].Type)
	assert.Equal(t, 0.70, visible[0].Similarity)
	assert.Equal(t, domain.Type1Clone, groups[0].Type)
	assert.InDelta(t, 0.85, groups[0].Similarity, 1e-9)
	assert.Len(t, groups[0].Clones, 3)
}
