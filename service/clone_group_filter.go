package service

import (
	coreclone "github.com/ludo-technologies/polyscan/core/clone"
	coredomain "github.com/ludo-technologies/polyscan/core/domain"
	"github.com/ludo-technologies/pyscn/domain"
)

// cloneMetadataID adapts response-wide IDs for core's pair-metadata refresh.
// Locations are unused by that operation; no spatial grouping runs here.
type cloneMetadataID int

func (id cloneMetadataID) ItemID() int { return int(id) }
func (id cloneMetadataID) ItemLocation() coreclone.ItemLocation {
	return coreclone.ItemLocation{}
}

// filterCloneGroupsByPairs restricts each detector group to connected components
// backed by retained pairs, using core's SplitGroupsByPairs. It preserves the
// detector's grouping boundaries, recomputes type and similarity from the
// retained pairs, and never mutates input.
func filterCloneGroupsByPairs(groups []*domain.CloneGroup, pairs []*domain.ClonePair) []*domain.CloneGroup {
	clones := make(map[int]*domain.Clone)
	coreGroups := make([]*coreclone.ItemGroup[cloneMetadataID], 0, len(groups))
	for _, group := range groups {
		items := make([]cloneMetadataID, 0, len(group.Clones))
		for _, clone := range group.Clones {
			items = append(items, cloneMetadataID(clone.ID))
			clones[clone.ID] = clone
		}
		coreGroups = append(coreGroups, &coreclone.ItemGroup[cloneMetadataID]{ID: group.ID, Items: items})
	}
	corePairs := make([]*coreclone.ItemPair[cloneMetadataID], 0, len(pairs))
	for _, pair := range pairs {
		corePairs = append(corePairs, &coreclone.ItemPair[cloneMetadataID]{
			Item1: cloneMetadataID(pair.Clone1.ID), Item2: cloneMetadataID(pair.Clone2.ID),
			Similarity: pair.Similarity, PairType: coredomain.CloneType(pair.Type),
		})
	}

	var filtered []*domain.CloneGroup
	for _, split := range coreclone.SplitGroupsByPairs(coreGroups, corePairs) {
		group := &domain.CloneGroup{
			ID:         split.ID,
			Type:       domain.CloneType(split.GroupType),
			Similarity: split.Similarity,
			Size:       len(split.Items),
		}
		for _, id := range split.Items {
			member := *clones[int(id)]
			member.Type = group.Type
			group.Clones = append(group.Clones, &member)
		}
		filtered = append(filtered, group)
	}
	return filtered
}
