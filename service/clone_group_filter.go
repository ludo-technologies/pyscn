package service

import "github.com/ludo-technologies/pyscn/domain"

// filterCloneGroupsByPairs restricts each detector group to connected components
// backed by retained pairs. It preserves the detector's grouping boundaries and
// strategy-specific metadata, and never mutates the input population.
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
	return filtered
}
