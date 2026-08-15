package analyzer

import (
	coreclone "github.com/ludo-technologies/polyscan/core/clone"
)

// GroupingMode represents the strategy for grouping clones. The grouping
// algorithms themselves live in core/clone; this type preserves pyscn's
// user-facing mode names (notably "star" for core's "star_medoid").
type GroupingMode string

const (
	GroupingModeConnected       GroupingMode = "connected"        // Current default (high recall)
	GroupingModeStar            GroupingMode = "star"             // Star/medoid (balanced)
	GroupingModeCompleteLinkage GroupingMode = "complete_linkage" // Complete linkage (high precision)
	GroupingModeKCore           GroupingMode = "k_core"           // k-core constrained (scalable)
	GroupingModeCentroid        GroupingMode = "centroid"         // Centroid based (avoids transitivity issues)
)

// coreMode translates a pyscn grouping mode to the core/clone grouping mode.
func (m GroupingMode) coreMode() coreclone.GroupingMode {
	if m == GroupingModeStar {
		return coreclone.ModeStarMedoid
	}
	return coreclone.GroupingMode(m)
}
