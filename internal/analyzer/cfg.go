package analyzer

import (
	corecfg "github.com/ludo-technologies/polyscan/core/cfg"
	"github.com/ludo-technologies/pyscn/domain"
)

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

// CFGScope is the canonical identity of one Python execution scope. Name is
// user-facing and may repeat; source location and kind keep identities distinct
// without leaking encoded map keys into reports.
type CFGScope struct {
	Kind        domain.AnalysisScopeKind
	Name        string
	StartLine   int
	StartColumn int
}

// ScopedCFG binds a control-flow graph to the execution scope that owns it.
type ScopedCFG struct {
	Scope CFGScope
	Graph *CFG
}

// ControlFlowGraphs preserves source traversal order and permits same-named
// scopes. A string-keyed map cannot represent those contracts without making
// an internal disambiguation key part of the public function name.
type ControlFlowGraphs []ScopedCFG
