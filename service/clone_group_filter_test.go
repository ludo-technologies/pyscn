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
