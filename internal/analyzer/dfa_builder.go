package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// DFABuilder constructs def-use chain information from a CFG
type DFABuilder struct {
	// cfg is the control flow graph to analyze
	cfg *CFG

	// info is the DFA information being built
	info *DFAInfo
}

// NewDFABuilder creates a new DFA builder
func NewDFABuilder() *DFABuilder {
	return &DFABuilder{}
}

// Build creates DFA information for the given CFG
func (b *DFABuilder) Build(cfg *CFG) (*DFAInfo, error) {
	if cfg == nil {
		return nil, nil
	}

	b.cfg = cfg
	b.info = NewDFAInfo(cfg)

	// Step 1: Collect all definitions
	b.collectDefinitions()

	// Step 2: Collect all uses
	b.collectUses()

	// Step 3: Link definitions to uses
	b.linkDefUse()

	return b.info, nil
}

// collectDefinitions walks the CFG to find all variable definitions
func (b *DFABuilder) collectDefinitions() {
	for _, block := range b.cfg.Blocks {
		// Skip exit block (never has statements)
		// Entry block may have statements (module-level code)
		if block.IsExit {
			continue
		}

		for pos, stmt := range block.Statements {
			defs := b.extractDefinitions(stmt, block, pos)
			for _, def := range defs {
				b.info.AddDef(def)
			}
		}
	}
}

// collectUses walks the CFG to find all variable uses
func (b *DFABuilder) collectUses() {
	for _, block := range b.cfg.Blocks {
		// Skip exit block (never has statements)
		// Entry block may have statements (module-level code)
		if block.IsExit {
			continue
		}

		for pos, stmt := range block.Statements {
			uses := b.extractUses(stmt, block, pos)
			for _, use := range uses {
				b.info.AddUse(use)
			}
		}
	}
}

// linkDefUse connects definitions to their uses
// Uses simplified reaching-definitions approximation:
// - Within same block: def at position i reaches use at position j if i < j and no intervening def
// - Cross-block: use CFG reachability with simple forward analysis
func (b *DFABuilder) linkDefUse() {
	for varName, chain := range b.info.Chains {
		// Sort definitions and uses by block order (approximation)
		// Link each use to the most recent definition that can reach it

		for _, use := range chain.Uses {
			def := b.findReachingDef(varName, use)
			if def != nil {
				pair := NewDefUsePair(def, use)
				chain.AddPair(pair)
			}
		}
	}
}

// findReachingDef finds the definition that reaches this use
func (b *DFABuilder) findReachingDef(varName string, use *VarReference) *VarReference {
	if use == nil || use.Block == nil {
		return nil
	}

	chain := b.info.Chains[varName]
	if chain == nil {
		return nil
	}

	// First, check for definitions in the same block before this use
	sameBlockDef := b.findDefInBlockBefore(chain.Defs, use.Block.ID, use.Position)
	if sameBlockDef != nil {
		return sameBlockDef
	}

	// Then, look for definitions in predecessor blocks using BFS
	return b.findDefInPredecessors(chain.Defs, use.Block)
}

// findDefInBlockBefore finds the last definition in the same block before the given position
func (b *DFABuilder) findDefInBlockBefore(defs []*VarReference, blockID string, usePos int) *VarReference {
	var lastDef *VarReference
	for _, def := range defs {
		if def.Block != nil && def.Block.ID == blockID && def.Position < usePos {
			if lastDef == nil || def.Position > lastDef.Position {
				lastDef = def
			}
		}
	}
	return lastDef
}

// findDefInPredecessors finds a definition in predecessor blocks using BFS
func (b *DFABuilder) findDefInPredecessors(defs []*VarReference, startBlock *BasicBlock) *VarReference {
	if startBlock == nil {
		return nil
	}

	// BFS to find nearest definition in predecessors
	visited := make(map[string]bool)
	queue := []*BasicBlock{}

	// Add all predecessors to queue
	for _, edge := range startBlock.Predecessors {
		if edge.From != nil && !visited[edge.From.ID] {
			queue = append(queue, edge.From)
			visited[edge.From.ID] = true
		}
	}

	for len(queue) > 0 {
		block := queue[0]
		queue = queue[1:]

		// Find the last definition in this block
		var lastDef *VarReference
		for _, def := range defs {
			if def.Block != nil && def.Block.ID == block.ID {
				// We want the definition with the highest position in this block
				if lastDef == nil || def.Position > lastDef.Position {
					lastDef = def
				}
			}
		}

		if lastDef != nil {
			return lastDef
		}

		// Add predecessors to queue
		for _, edge := range block.Predecessors {
			if edge.From != nil && !visited[edge.From.ID] {
				queue = append(queue, edge.From)
				visited[edge.From.ID] = true
			}
		}
	}

	return nil
}

// extractDefinitions extracts all definitions from a statement
func (b *DFABuilder) extractDefinitions(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	if stmt == nil {
		return nil
	}

	var defs []*VarReference

	switch stmt.Type {
	case parser.NodeAssign:
		// x = value, a, b = value
		defs = append(defs, b.extractAssignmentDefs(stmt, block, pos)...)

	case parser.NodeAugAssign:
		// x += value (both def and use, but we record as def here)
		defs = append(defs, b.extractAugAssignDefs(stmt, block, pos)...)

	case parser.NodeAnnAssign:
		// x: int = value
		defs = append(defs, b.extractAnnAssignDefs(stmt, block, pos)...)

	case parser.NodeFor, parser.NodeAsyncFor:
		// for x in ...: (loop target is a definition)
		defs = append(defs, b.extractForTargetDefs(stmt, block, pos)...)

	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		// def f(x, y): (parameters are definitions)
		defs = append(defs, b.extractParameterDefs(stmt, block, pos)...)

	case parser.NodeImport:
		// import x, y
		defs = append(defs, b.extractImportDefs(stmt, block, pos)...)

	case parser.NodeImportFrom:
		// from m import x, y
		defs = append(defs, b.extractImportFromDefs(stmt, block, pos)...)

	case parser.NodeWith, parser.NodeAsyncWith:
		// with ... as x:
		defs = append(defs, b.extractWithTargetDefs(stmt, block, pos)...)

	case parser.NodeExceptHandler:
		// except E as x:
		defs = append(defs, b.extractExceptTargetDefs(stmt, block, pos)...)

	case parser.NodeNamedExpr:
		// x := value (walrus operator)
		defs = append(defs, b.extractNamedExprDefs(stmt, block, pos)...)
	}

	return defs
}

// extractUses extracts all uses from a statement
func (b *DFABuilder) extractUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	if stmt == nil {
		return nil
	}

	var uses []*VarReference

	// For assignment statements, we need to extract uses from the Value field
	// since Walk doesn't include it
	if stmt.Type == parser.NodeAssign || stmt.Type == parser.NodeAnnAssign {
		if valueNode, ok := stmt.Value.(*parser.Node); ok {
			uses = append(uses, b.extractUsesFromExpression(valueNode, block, stmt, pos)...)
		}
		return uses
	}

	// For augmented assignment, the target is both a def and a use
	if stmt.Type == parser.NodeAugAssign {
		if len(stmt.Targets) > 0 && stmt.Targets[0] != nil && stmt.Targets[0].Type == parser.NodeName {
			ref := NewVarReference(stmt.Targets[0].Name, UseKindRead, block, stmt, pos)
			uses = append(uses, ref)
		}
		// Also extract uses from the right-hand side
		if valueNode, ok := stmt.Value.(*parser.Node); ok {
			uses = append(uses, b.extractUsesFromExpression(valueNode, block, stmt, pos)...)
		}
		return uses
	}

	// For other statements, walk the entire statement tree
	uses = append(uses, b.extractUsesFromExpression(stmt, block, stmt, pos)...)

	return uses
}

// extractUsesFromExpression recursively extracts variable uses from an expression
func (b *DFABuilder) extractUsesFromExpression(expr *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if expr == nil {
		return nil
	}

	var uses []*VarReference

	switch expr.Type {
	case parser.NodeName:
		ref := NewVarReference(expr.Name, UseKindRead, block, stmt, pos)
		uses = append(uses, ref)

	case parser.NodeAttribute:
		// x.attr - the base (x) is a use
		// The base object is stored in Value field, not Children
		if baseNode, ok := expr.Value.(*parser.Node); ok {
			if baseNode.Type == parser.NodeName {
				ref := NewVarReference(baseNode.Name, UseKindAttribute, block, stmt, pos)
				uses = append(uses, ref)
			} else {
				// Recurse for chained attributes like a.b.c
				uses = append(uses, b.extractUsesFromExpression(baseNode, block, stmt, pos)...)
			}
		} else if len(expr.Children) > 0 {
			// Fallback to checking Children for other AST structures
			base := expr.Children[0]
			if base != nil && base.Type == parser.NodeName {
				ref := NewVarReference(base.Name, UseKindAttribute, block, stmt, pos)
				uses = append(uses, ref)
			} else if base != nil {
				uses = append(uses, b.extractUsesFromExpression(base, block, stmt, pos)...)
			}
		}

	case parser.NodeSubscript:
		// x[i] - the base (x) is a use, and the index might contain uses too
		if len(expr.Children) > 0 {
			base := expr.Children[0]
			if base != nil && base.Type == parser.NodeName {
				ref := NewVarReference(base.Name, UseKindSubscript, block, stmt, pos)
				uses = append(uses, ref)
			} else if base != nil {
				uses = append(uses, b.extractUsesFromExpression(base, block, stmt, pos)...)
			}
		}
		// Also check the subscript index
		if len(expr.Children) > 1 {
			uses = append(uses, b.extractUsesFromExpression(expr.Children[1], block, stmt, pos)...)
		}

	case parser.NodeCall:
		// f(x) - the function and arguments are uses
		if len(expr.Children) > 0 {
			funcNode := expr.Children[0]
			if funcNode != nil && funcNode.Type == parser.NodeName {
				ref := NewVarReference(funcNode.Name, UseKindCall, block, stmt, pos)
				uses = append(uses, ref)
			} else if funcNode != nil {
				uses = append(uses, b.extractUsesFromExpression(funcNode, block, stmt, pos)...)
			}
		}
		// Process arguments
		for _, arg := range expr.Args {
			uses = append(uses, b.extractUsesFromExpression(arg, block, stmt, pos)...)
		}
		// Process keyword arguments
		for _, kw := range expr.Keywords {
			if len(kw.Children) > 0 {
				uses = append(uses, b.extractUsesFromExpression(kw.Children[0], block, stmt, pos)...)
			}
		}

	case parser.NodeBinOp:
		// a + b - both operands are uses
		uses = append(uses, b.extractUsesFromExpression(expr.Left, block, stmt, pos)...)
		uses = append(uses, b.extractUsesFromExpression(expr.Right, block, stmt, pos)...)

	case parser.NodeUnaryOp:
		// -x, not x - the operand is a use
		if len(expr.Children) > 0 {
			uses = append(uses, b.extractUsesFromExpression(expr.Children[0], block, stmt, pos)...)
		}

	case parser.NodeCompare:
		// a < b - both operands are uses
		uses = append(uses, b.extractUsesFromExpression(expr.Left, block, stmt, pos)...)
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeTuple, parser.NodeList, parser.NodeSet:
		// Collection literals - elements are uses
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeDict:
		// Dict literal - keys and values are uses
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeIfExp:
		// Ternary expression: a if b else c
		uses = append(uses, b.extractUsesFromExpression(expr.Test, block, stmt, pos)...)
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeLambda:
		// Lambda: lambda x: x + 1 - body contains uses, but parameters are local
		// For simplicity, we skip lambda internals for now

	case parser.NodeBoolOp:
		// and/or - operands are uses
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	default:
		// For other expression types, recursively process children
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}
	}

	return uses
}

// extractAssignmentDefs extracts definitions from an assignment statement
func (b *DFABuilder) extractAssignmentDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	for _, target := range stmt.Targets {
		defs = append(defs, b.extractNamesFromTarget(target, DefKindAssign, block, stmt, pos)...)
	}

	return defs
}

// extractAugAssignDefs extracts definitions from an augmented assignment
func (b *DFABuilder) extractAugAssignDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	if len(stmt.Targets) > 0 && stmt.Targets[0] != nil {
		target := stmt.Targets[0]
		if target.Type == parser.NodeName {
			ref := NewVarReference(target.Name, DefKindAugmented, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

// extractAnnAssignDefs extracts definitions from an annotated assignment
func (b *DFABuilder) extractAnnAssignDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// AnnAssign: target is usually in Targets[0] or Children[0]
	if len(stmt.Targets) > 0 && stmt.Targets[0] != nil {
		target := stmt.Targets[0]
		if target.Type == parser.NodeName {
			ref := NewVarReference(target.Name, DefKindAssign, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

// extractForTargetDefs extracts definitions from a for loop target
func (b *DFABuilder) extractForTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// For loop target is stored in Targets
	for _, target := range stmt.Targets {
		defs = append(defs, b.extractNamesFromTarget(target, DefKindForTarget, block, stmt, pos)...)
	}

	return defs
}

// extractParameterDefs extracts parameter definitions from a function
func (b *DFABuilder) extractParameterDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// Find Arguments node in children
	for _, child := range stmt.Children {
		if child != nil && child.Type == parser.NodeArguments {
			// Extract parameter names from arguments
			for _, arg := range child.Children {
				if arg != nil && arg.Type == parser.NodeArg && arg.Name != "" {
					ref := NewVarReference(arg.Name, DefKindParameter, block, stmt, pos)
					defs = append(defs, ref)
				}
			}
			break
		}
	}

	return defs
}

// extractImportDefs extracts definitions from an import statement
func (b *DFABuilder) extractImportDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// Import names from Names field or Alias children
	for _, name := range stmt.Names {
		ref := NewVarReference(name, DefKindImport, block, stmt, pos)
		defs = append(defs, ref)
	}

	// Also check Alias children
	for _, child := range stmt.Children {
		if child != nil && child.Type == parser.NodeAlias {
			name := child.Name
			if name == "" && len(child.Names) > 0 {
				name = child.Names[0]
			}
			if name != "" {
				ref := NewVarReference(name, DefKindImport, block, stmt, pos)
				defs = append(defs, ref)
			}
		}
	}

	return defs
}

// extractImportFromDefs extracts definitions from a from-import statement
func (b *DFABuilder) extractImportFromDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// Import names
	for _, name := range stmt.Names {
		ref := NewVarReference(name, DefKindImport, block, stmt, pos)
		defs = append(defs, ref)
	}

	// Also check Alias children for as-names
	for _, child := range stmt.Children {
		if child != nil && child.Type == parser.NodeAlias {
			name := child.Name
			if name == "" && len(child.Names) > 0 {
				name = child.Names[0]
			}
			if name != "" {
				ref := NewVarReference(name, DefKindImport, block, stmt, pos)
				defs = append(defs, ref)
			}
		}
	}

	return defs
}

// extractWithTargetDefs extracts definitions from a with statement
func (b *DFABuilder) extractWithTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// Look for WithItem children
	for _, child := range stmt.Children {
		if child != nil && child.Type == parser.NodeWithItem {
			// The second child of WithItem is the optional target (as x)
			if len(child.Children) > 1 && child.Children[1] != nil {
				target := child.Children[1]
				defs = append(defs, b.extractNamesFromTarget(target, DefKindWithTarget, block, stmt, pos)...)
			}
		}
	}

	return defs
}

// extractExceptTargetDefs extracts definitions from an except handler
func (b *DFABuilder) extractExceptTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// ExceptHandler has a name field for "as x"
	if stmt.Name != "" {
		ref := NewVarReference(stmt.Name, DefKindExceptTarget, block, stmt, pos)
		defs = append(defs, ref)
	}

	return defs
}

// extractNamedExprDefs extracts definitions from a named expression (walrus operator)
func (b *DFABuilder) extractNamedExprDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// First child is the target name
	if len(stmt.Children) > 0 && stmt.Children[0] != nil {
		target := stmt.Children[0]
		if target.Type == parser.NodeName {
			ref := NewVarReference(target.Name, DefKindAssign, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

// extractNamesFromTarget extracts all names from an assignment target (handles tuples)
func (b *DFABuilder) extractNamesFromTarget(target *parser.Node, kind DefUseKind, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if target == nil {
		return nil
	}

	var defs []*VarReference

	switch target.Type {
	case parser.NodeName:
		ref := NewVarReference(target.Name, kind, block, stmt, pos)
		defs = append(defs, ref)

	case parser.NodeTuple, parser.NodeList:
		// Unpacking: a, b = ... or [a, b] = ...
		for _, elem := range target.Children {
			defs = append(defs, b.extractNamesFromTarget(elem, kind, block, stmt, pos)...)
		}

	case parser.NodeStarred:
		// *rest in unpacking
		if len(target.Children) > 0 {
			defs = append(defs, b.extractNamesFromTarget(target.Children[0], kind, block, stmt, pos)...)
		}

	default:
		// Handle pattern_list and other tree-sitter specific types
		// pattern_list is used for tuple unpacking like: a, b = 1, 2
		if string(target.Type) == "pattern_list" {
			for _, elem := range target.Children {
				// Skip comma separators
				if elem != nil && elem.Type != "," && string(elem.Type) != "," {
					defs = append(defs, b.extractNamesFromTarget(elem, kind, block, stmt, pos)...)
				}
			}
		}
	}

	return defs
}
