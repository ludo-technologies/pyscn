package analyzer

// GroupingStrategy defines a strategy for grouping clone pairs into clone groups.
// Implementations should avoid recomputing similarities and work with provided pairs.
type GroupingStrategy interface {
	// GroupClones groups the given clone pairs into clone groups.
	GroupClones(pairs []*ClonePair) []*CloneGroup
	// GetName returns the strategy name.
	GetName() string
}
