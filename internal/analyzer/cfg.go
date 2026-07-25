package analyzer

import corecfg "github.com/ludo-technologies/polyscan/core/cfg"

// CFG data structures and traversal are owned by polyscan core. Aliases keep
// pyscn's internal API stable while Python-specific construction stays local.
type (
	EdgeType   = corecfg.EdgeType
	Edge       = corecfg.Edge
	BasicBlock = corecfg.BasicBlock
	CFG        = corecfg.CFG
	CFGVisitor = corecfg.Visitor
)

const (
	EdgeNormal    = corecfg.EdgeNormal
	EdgeCondTrue  = corecfg.EdgeCondTrue
	EdgeCondFalse = corecfg.EdgeCondFalse
	EdgeException = corecfg.EdgeException
	EdgeLoop      = corecfg.EdgeLoop
	EdgeBreak     = corecfg.EdgeBreak
	EdgeContinue  = corecfg.EdgeContinue
	EdgeReturn    = corecfg.EdgeReturn
)

var (
	NewBasicBlock = corecfg.NewBasicBlock
	NewCFG        = corecfg.NewCFG
)
