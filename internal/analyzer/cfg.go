package analyzer

import (
	"fmt"
	"github.com/pyqol/pyqol/internal/parser"
)

// EdgeType represents the type of edge between basic blocks
type EdgeType int

const (
	// EdgeNormal represents normal sequential flow
	EdgeNormal EdgeType = iota
	// EdgeCondTrue represents conditional true branch
	EdgeCondTrue
	// EdgeCondFalse represents conditional false branch
	EdgeCondFalse
	// EdgeException represents exception flow
	EdgeException
	// EdgeLoop represents loop back edge
	EdgeLoop
	// EdgeBreak represents break statement flow
	EdgeBreak
	// EdgeContinue represents continue statement flow
	EdgeContinue
	// EdgeReturn represents return statement flow
	EdgeReturn
)

// String returns string representation of EdgeType
func (e EdgeType) String() string {
	switch e {
	case EdgeNormal:
		return "normal"
	case EdgeCondTrue:
		return "true"
	case EdgeCondFalse:
		return "false"
	case EdgeException:
		return "exception"
	case EdgeLoop:
		return "loop"
	case EdgeBreak:
		return "break"
	case EdgeContinue:
		return "continue"
	case EdgeReturn:
		return "return"
	default:
		return "unknown"
	}
}

// Edge represents a directed edge between two basic blocks
type Edge struct {
	From *BasicBlock
	To   *BasicBlock
	Type EdgeType
}

// BasicBlock represents a basic block in the control flow graph
type BasicBlock struct {
	// ID is the unique identifier for this block
	ID string
	
	// Statements contains the AST nodes in this block
	Statements []*parser.Node
	
	// Predecessors are blocks that can flow into this block
	Predecessors []*Edge
	
	// Successors are blocks that this block can flow to
	Successors []*Edge
	
	// Label is an optional human-readable label
	Label string
	
	// IsEntry indicates if this is an entry block
	IsEntry bool
	
	// IsExit indicates if this is an exit block
	IsExit bool
}

// NewBasicBlock creates a new basic block with the given ID
func NewBasicBlock(id string) *BasicBlock {
	return &BasicBlock{
		ID:           id,
		Statements:   []*parser.Node{},
		Predecessors: []*Edge{},
		Successors:   []*Edge{},
	}
}

// AddStatement adds an AST node to this block
func (bb *BasicBlock) AddStatement(stmt *parser.Node) {
	if stmt != nil {
		bb.Statements = append(bb.Statements, stmt)
	}
}

// AddSuccessor adds an outgoing edge to another block
func (bb *BasicBlock) AddSuccessor(to *BasicBlock, edgeType EdgeType) *Edge {
	edge := &Edge{
		From: bb,
		To:   to,
		Type: edgeType,
	}
	bb.Successors = append(bb.Successors, edge)
	to.Predecessors = append(to.Predecessors, edge)
	return edge
}

// RemoveSuccessor removes an edge to the specified block
func (bb *BasicBlock) RemoveSuccessor(to *BasicBlock) {
	// Remove from successors
	newSuccessors := []*Edge{}
	for _, edge := range bb.Successors {
		if edge.To != to {
			newSuccessors = append(newSuccessors, edge)
		}
	}
	bb.Successors = newSuccessors
	
	// Remove from target's predecessors
	newPredecessors := []*Edge{}
	for _, edge := range to.Predecessors {
		if edge.From != bb {
			newPredecessors = append(newPredecessors, edge)
		}
	}
	to.Predecessors = newPredecessors
}

// IsEmpty returns true if the block has no statements
func (bb *BasicBlock) IsEmpty() bool {
	return len(bb.Statements) == 0
}

// String returns a string representation of the basic block
func (bb *BasicBlock) String() string {
	label := bb.Label
	if label == "" {
		label = bb.ID
	}
	
	if bb.IsEntry {
		return fmt.Sprintf("[ENTRY: %s]", label)
	}
	if bb.IsExit {
		return fmt.Sprintf("[EXIT: %s]", label)
	}
	
	return fmt.Sprintf("[%s: %d stmts]", label, len(bb.Statements))
}

// CFG represents a control flow graph
type CFG struct {
	// Entry is the entry point of the graph
	Entry *BasicBlock
	
	// Exit is the exit point of the graph
	Exit *BasicBlock
	
	// Blocks contains all blocks in the graph, indexed by ID
	Blocks map[string]*BasicBlock
	
	// Name is the name of the CFG (e.g., function name)
	Name string
	
	// nextBlockID is used to generate unique block IDs
	nextBlockID int
}

// NewCFG creates a new control flow graph
func NewCFG(name string) *CFG {
	cfg := &CFG{
		Name:        name,
		Blocks:      make(map[string]*BasicBlock),
		nextBlockID: 0,
	}
	
	// Create entry and exit blocks
	cfg.Entry = cfg.CreateBlock("entry")
	cfg.Entry.IsEntry = true
	cfg.Entry.Label = "ENTRY"
	
	cfg.Exit = cfg.CreateBlock("exit")
	cfg.Exit.IsExit = true
	cfg.Exit.Label = "EXIT"
	
	return cfg
}

// CreateBlock creates a new basic block and adds it to the graph
func (cfg *CFG) CreateBlock(label string) *BasicBlock {
	id := fmt.Sprintf("bb%d", cfg.nextBlockID)
	cfg.nextBlockID++
	
	block := NewBasicBlock(id)
	if label != "" {
		block.Label = label
	}
	
	cfg.Blocks[id] = block
	return block
}

// AddBlock adds an existing block to the graph
func (cfg *CFG) AddBlock(block *BasicBlock) {
	if block != nil {
		cfg.Blocks[block.ID] = block
	}
}

// RemoveBlock removes a block from the graph
func (cfg *CFG) RemoveBlock(block *BasicBlock) {
	if block == nil || block.IsEntry || block.IsExit {
		return
	}
	
	// Remove all edges to and from this block
	for _, pred := range block.Predecessors {
		pred.From.RemoveSuccessor(block)
	}
	for _, succ := range block.Successors {
		block.RemoveSuccessor(succ.To)
	}
	
	// Remove from blocks map
	delete(cfg.Blocks, block.ID)
}

// ConnectBlocks creates an edge between two blocks
func (cfg *CFG) ConnectBlocks(from, to *BasicBlock, edgeType EdgeType) *Edge {
	if from == nil || to == nil {
		return nil
	}
	return from.AddSuccessor(to, edgeType)
}

// GetBlock retrieves a block by its ID
func (cfg *CFG) GetBlock(id string) *BasicBlock {
	return cfg.Blocks[id]
}

// Size returns the number of blocks in the graph
func (cfg *CFG) Size() int {
	return len(cfg.Blocks)
}

// CFGVisitor defines the interface for visiting CFG nodes
type CFGVisitor interface {
	// VisitBlock is called for each basic block
	// Returns false to stop traversal
	VisitBlock(block *BasicBlock) bool
	
	// VisitEdge is called for each edge
	// Returns false to stop traversal
	VisitEdge(edge *Edge) bool
}

// Walk performs a depth-first traversal of the CFG
func (cfg *CFG) Walk(visitor CFGVisitor) {
	if cfg.Entry == nil {
		return
	}
	
	visited := make(map[string]bool)
	cfg.walkBlock(cfg.Entry, visitor, visited)
}

// walkBlock recursively visits blocks in depth-first order
func (cfg *CFG) walkBlock(block *BasicBlock, visitor CFGVisitor, visited map[string]bool) {
	if block == nil || visited[block.ID] {
		return
	}
	
	visited[block.ID] = true
	
	// Visit the block
	if !visitor.VisitBlock(block) {
		return
	}
	
	// Visit edges and successors
	for _, edge := range block.Successors {
		if !visitor.VisitEdge(edge) {
			return
		}
		cfg.walkBlock(edge.To, visitor, visited)
	}
}

// BreadthFirstWalk performs a breadth-first traversal of the CFG
func (cfg *CFG) BreadthFirstWalk(visitor CFGVisitor) {
	if cfg.Entry == nil {
		return
	}
	
	visited := make(map[string]bool)
	queue := []*BasicBlock{cfg.Entry}
	
	for len(queue) > 0 {
		block := queue[0]
		queue = queue[1:]
		
		if visited[block.ID] {
			continue
		}
		visited[block.ID] = true
		
		// Visit the block
		if !visitor.VisitBlock(block) {
			return
		}
		
		// Visit edges and add successors to queue
		for _, edge := range block.Successors {
			if !visitor.VisitEdge(edge) {
				return
			}
			if !visited[edge.To.ID] {
				queue = append(queue, edge.To)
			}
		}
	}
}

// String returns a string representation of the CFG
func (cfg *CFG) String() string {
	return fmt.Sprintf("CFG(%s): %d blocks", cfg.Name, cfg.Size())
}