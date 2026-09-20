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
// backed by retained pairs. It preserves the detector's grouping boundaries,
// refreshes metadata using the detector's core routine, and never mutates input.
func filterCloneGroupsByPairs(groups []*domain.CloneGroup, pairs []*domain.ClonePair) []*domain.CloneGroup {
	adjacency := make(map[int][]int)
	for _, pair := range pairs {
		a, b := pair.Clone1.ID, pair.Clone2.ID
		adjacency[a] = append(adjacency[a], b)
		adjacency[b] = append(adjacency[b], a)
	}
	nextID := 0
	for _, group := range groups {
		if group.ID >= nextID {
			nextID = group.ID + 1
		}
	}
	var filtered []*domain.CloneGroup
	for _, group := range groups {
		members := make(map[int]bool, len(group.Clones))
		for _, clone := range group.Clones {
			members[clone.ID] = true
		}
		components := make(map[int]int, len(members))
		var buckets [][]*domain.Clone
		for _, clone := range group.Clones {
			if _, seen := components[clone.ID]; seen {
				continue
			}
			component := len(buckets)
			buckets = append(buckets, nil)
			components[clone.ID] = component
			queue := []int{clone.ID}
			for i := 0; i < len(queue); i++ {
				for _, neighbor := range adjacency[queue[i]] {
					if !members[neighbor] {
						continue
					}
					if _, seen := components[neighbor]; !seen {
						components[neighbor] = component
						queue = append(queue, neighbor)
					}
				}
			}
		}
		// Keep the detector's order within each component.
		for _, clone := range group.Clones {
			component := components[clone.ID]
			buckets[component] = append(buckets[component], clone)
		}
		first := true
		for _, bucket := range buckets {
			if len(bucket) < 2 {
				continue
			}
			retained := *group
			retained.Clones = bucket
			retained.Size = len(bucket)
			if !first {
				retained.ID = nextID
				nextID++
			}
			first = false
			filtered = append(filtered, &retained)
		}
	}
	return refreshCloneGroupMetadata(filtered, pairs)
}

func refreshCloneGroupMetadata(groups []*domain.CloneGroup, pairs []*domain.ClonePair) []*domain.CloneGroup {
	corePairs := make([]*coreclone.ItemPair[cloneMetadataID], 0, len(pairs))
	for _, pair := range pairs {
		corePairs = append(corePairs, &coreclone.ItemPair[cloneMetadataID]{
			Item1: cloneMetadataID(pair.Clone1.ID), Item2: cloneMetadataID(pair.Clone2.ID),
			Similarity: pair.Similarity, PairType: coredomain.CloneType(pair.Type),
		})
	}
	coreGroups := make([]*coreclone.ItemGroup[cloneMetadataID], 0, len(groups))
	byID := make(map[int]*domain.CloneGroup, len(groups))
	for _, group := range groups {
		items := make([]cloneMetadataID, 0, len(group.Clones))
		for _, clone := range group.Clones {
			items = append(items, cloneMetadataID(clone.ID))
		}
		coreGroups = append(coreGroups, &coreclone.ItemGroup[cloneMetadataID]{ID: group.ID, Items: items})
		byID[group.ID] = group
	}
	var refreshed []*domain.CloneGroup
	for _, metadata := range coreclone.FilterGroupsWithoutBackingPairs(coreGroups, corePairs) {
		group := byID[metadata.ID]
		group.Type = domain.CloneType(metadata.GroupType)
		group.Similarity = metadata.Similarity
		for i, clone := range group.Clones {
			member := *clone
			member.Type = group.Type
			group.Clones[i] = &member
		}
		refreshed = append(refreshed, group)
	}
	return refreshed
}
