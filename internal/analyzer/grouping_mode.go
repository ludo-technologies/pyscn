package analyzer

import (
	coreclone "github.com/ludo-technologies/polyscan/core/clone"
)

// GroupingMode represents the strategy for grouping clones. The grouping
// algorithms themselves live in core/clone; this type preserves pyscn's
// user-facing mode names (notably "star" for core's "star_medoid").
type GroupingMode string

const (
	GroupingModeConnected       GroupingMode = "connected"        // 現在のデフォルト（高再現率）
	GroupingModeStar            GroupingMode = "star"             // Star/Medoid（バランス型）
	GroupingModeCompleteLinkage GroupingMode = "complete_linkage" // 完全連結（高精度）
	GroupingModeKCore           GroupingMode = "k_core"           // k-core制約（スケーラブル）
	GroupingModeCentroid        GroupingMode = "centroid"         // 重心ベース（推移的問題を回避）
)

// coreMode translates a pyscn grouping mode to the core/clone grouping mode.
func (m GroupingMode) coreMode() coreclone.GroupingMode {
	if m == GroupingModeStar {
		return coreclone.ModeStarMedoid
	}
	return coreclone.GroupingMode(m)
}
