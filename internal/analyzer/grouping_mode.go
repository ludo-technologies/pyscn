package analyzer

import (
	coreclone "github.com/ludo-technologies/polyscan/core/clone"

	"github.com/ludo-technologies/pyscn/domain"
)

// GroupingMode represents the strategy for grouping clones. The grouping
// algorithms themselves live in core/clone; this type preserves pyscn's
// user-facing mode names (notably "star" for core's "star_medoid").
type GroupingMode string

const (
	GroupingModeConnected       GroupingMode = domain.CloneGroupModeConnected       // Single linkage (high recall, chains unrelated clones)
	GroupingModeStar            GroupingMode = domain.CloneGroupModeStar            // Star/medoid (balanced)
	GroupingModeCompleteLinkage GroupingMode = domain.CloneGroupModeCompleteLinkage // Complete linkage (default, high precision)
	GroupingModeKCore           GroupingMode = domain.CloneGroupModeKCore           // k-core constrained (scalable)
	GroupingModeCentroid        GroupingMode = domain.CloneGroupModeCentroid        // Centroid based (avoids transitivity issues)
)

// coreMode translates a pyscn grouping mode to the core/clone grouping mode.
func (m GroupingMode) coreMode() coreclone.GroupingMode {
	if m == GroupingModeStar {
		return coreclone.ModeStarMedoid
	}
	return coreclone.GroupingMode(m)
}
