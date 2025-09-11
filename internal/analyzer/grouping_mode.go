package analyzer

// GroupingMode represents the strategy for grouping clones
type GroupingMode string

const (
    GroupingModeConnected       GroupingMode = "connected"        // 現在のデフォルト（高再現率）
    GroupingModeStar            GroupingMode = "star"             // Star/Medoid（バランス型）
    GroupingModeCompleteLinkage GroupingMode = "complete_linkage" // 完全連結（高精度）
    GroupingModeKCore           GroupingMode = "k_core"           // k-core制約（スケーラブル）
)

// GroupingConfig holds configuration for grouping strategies
type GroupingConfig struct {
    Mode      GroupingMode
    Threshold float64 // Minimum similarity for group membership
    KCoreK    int     // K value for k-core mode (default: 2)
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
    case GroupingModeConnected:
        fallthrough
    default:
        return NewConnectedGrouping(config.Threshold)
    }
}

