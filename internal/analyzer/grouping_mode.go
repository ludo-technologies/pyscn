package analyzer

// GroupingMode represents the strategy for grouping clones
type GroupingMode string

const (
	GroupingModeConnected       GroupingMode = "connected"        // 現在のデフォルト（高再現率）
	GroupingModeStar            GroupingMode = "star"             // Star/Medoid（バランス型）
	GroupingModeCompleteLinkage GroupingMode = "complete_linkage" // 完全連結（高精度）
	GroupingModeKCore           GroupingMode = "k_core"           // k-core制約（スケーラブル）
	GroupingModeCentroid        GroupingMode = "centroid"         // 重心ベース（推移的問題を回避）
)

// GroupingConfig holds configuration for grouping strategies
type GroupingConfig struct {
	Mode           GroupingMode
	Threshold      float64 // Minimum similarity for group membership
	KCoreK         int     // K value for k-core mode (default: 2)
	Type1Threshold float64 // Type-1 clone threshold
	Type2Threshold float64 // Type-2 clone threshold
	Type3Threshold float64 // Type-3 clone threshold
	Type4Threshold float64 // Type-4 clone threshold
}

// CreateGroupingStrategy creates a strategy based on mode and config
func CreateGroupingStrategy(config GroupingConfig) GroupingStrategy {
	switch config.Mode {
	case GroupingModeStar:
		return NewStarMedoidGrouping(config.Threshold)
	case GroupingModeCompleteLinkage:
		return NewCompleteLinkageGrouping(config.Threshold)
	case GroupingModeKCore:
		return NewKCoreGrouping(config.Threshold, config.KCoreK)
	case GroupingModeCentroid:
		strategy := NewCentroidGrouping(config.Threshold)
		strategy.SetThresholds(config.Type1Threshold, config.Type2Threshold, config.Type3Threshold, config.Type4Threshold)
		return strategy
	case GroupingModeConnected:
		fallthrough
	default:
		return NewConnectedGrouping(config.Threshold)
	}
}
