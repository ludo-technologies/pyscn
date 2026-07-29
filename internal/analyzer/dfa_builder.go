package analyzer

import (
	coredfa "github.com/ludo-technologies/polyscan/core/dfa"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// paramDefPosition places parameter definitions before all statement
// positions so they reach uses at position 0.
const paramDefPosition = -1

// pythonRefExtractor extracts Python variable definitions and uses for
// core's DFA builder. Def-use linking itself lives in core.
type pythonRefExtractor struct{}

var (
	_ coredfa.RefExtractor   = (*pythonRefExtractor)(nil)
	_ coredfa.ParamExtractor = (*pythonRefExtractor)(nil)
)

// NewDFABuilder creates a DFA builder wired to Python reference extraction
func NewDFABuilder() *DFABuilder {
	return coredfa.NewDFABuilder(&pythonRefExtractor{})
}

// ExtractDefinitions implements coredfa.RefExtractor
func (b *pythonRefExtractor) ExtractDefinitions(stmt any, block *BasicBlock, pos int) []*VarReference {
	return b.extractDefinitions(mustPythonNode(stmt), block, pos)
}

// ExtractUses implements coredfa.RefExtractor
func (b *pythonRefExtractor) ExtractUses(stmt any, block *BasicBlock, pos int) []*VarReference {
	return b.extractUses(mustPythonNode(stmt), block, pos)
}

// ExtractParameterDefs implements coredfa.ParamExtractor
func (b *pythonRefExtractor) ExtractParameterDefs(functionNode any, entry *BasicBlock) []*VarReference {
	return b.extractParameterDefs(mustPythonNode(functionNode), entry, paramDefPosition)
}

// extractDefinitions extracts all definitions from a statement
func (b *pythonRefExtractor) extractDefinitions(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
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
		// def f(...): binds f in the containing CFG. Parameters belong to the function CFG.
		defs = append(defs, b.extractNamedBindingDef(stmt, block, pos)...)

	case parser.NodeClassDef:
		// class C(...): binds C in the containing CFG.
		defs = append(defs, b.extractNamedBindingDef(stmt, block, pos)...)

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

	case parser.NodeMatchCase:
		// case x: binds x for the case body and guard.
		defs = append(defs, b.extractMatchPatternDefs(stmt.Test, block, stmt, pos)...)

	case parser.NodeNamedExpr:
		// x := value (walrus operator)
		defs = append(defs, b.extractNamedExprDefs(stmt, block, pos)...)
	}

	return defs
}

// extractUses extracts all uses from a statement
func (b *pythonRefExtractor) extractUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	if stmt == nil {
		return nil
	}

	switch stmt.Type {
	case parser.NodeAssign, parser.NodeAnnAssign:
		var uses []*VarReference
		// For assignment statements, only the right-hand side is a read.
		if valueNode, ok := stmt.Value.(*parser.Node); ok {
			uses = append(uses, b.extractUsesFromExpression(valueNode, block, stmt, pos)...)
		}
		return uses

	case parser.NodeAugAssign:
		var uses []*VarReference
		// For augmented assignment, the target is both a def and a use.
		if len(stmt.Targets) > 0 && stmt.Targets[0] != nil && stmt.Targets[0].Type == parser.NodeName {
			ref := NewVarReference(stmt.Targets[0].Name, UseKindLoad, block, stmt, pos)
			uses = append(uses, ref)
		}
		if valueNode, ok := stmt.Value.(*parser.Node); ok {
			uses = append(uses, b.extractUsesFromExpression(valueNode, block, stmt, pos)...)
		}
		return uses
	}

	return b.extractStatementHeaderUses(stmt, block, pos)
}

func (b *pythonRefExtractor) extractStatementHeaderUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var uses []*VarReference

	switch stmt.Type {
	case parser.NodeIf, parser.NodeElifClause, parser.NodeWhile:
		uses = append(uses, b.extractUsesFromExpression(stmt.Test, block, stmt, pos)...)

	case parser.NodeFor, parser.NodeAsyncFor:
		uses = append(uses, b.extractUsesFromExpression(stmt.Iter, block, stmt, pos)...)

	case parser.NodeWith, parser.NodeAsyncWith:
		uses = append(uses, b.extractWithContextUses(stmt, block, pos)...)

	case parser.NodeMatch:
		uses = append(uses, b.extractUsesFromExpression(stmt.Test, block, stmt, pos)...)

	case parser.NodeMatchCase:
		uses = append(uses, b.extractMatchPatternUses(stmt.Test, block, stmt, pos)...)
		if guard := nodeValue(stmt); guard != nil {
			uses = append(uses, b.extractUsesFromExpression(guard, block, stmt, pos)...)
		}

	case parser.NodeExceptHandler:
		if exceptionType := nodeValue(stmt); exceptionType != nil {
			uses = append(uses, b.extractUsesFromExpression(exceptionType, block, stmt, pos)...)
		}

	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		uses = append(uses, b.extractDecoratorUses(stmt, block, pos)...)
		uses = append(uses, b.extractFunctionDefaultUses(stmt, block, pos)...)

	case parser.NodeClassDef:
		uses = append(uses, b.extractDecoratorUses(stmt, block, pos)...)
		for _, base := range stmt.Bases {
			uses = append(uses, b.extractUsesFromExpression(base, block, stmt, pos)...)
		}

	default:
		uses = append(uses, b.extractUsesFromExpression(stmt, block, stmt, pos)...)
	}

	return uses
}

func (b *pythonRefExtractor) extractDecoratorUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var uses []*VarReference
	for _, decorator := range stmt.Decorator {
		uses = append(uses, b.extractUsesFromExpression(decorator, block, stmt, pos)...)
	}
	return uses
}

func (b *pythonRefExtractor) extractFunctionDefaultUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var uses []*VarReference
	for _, arg := range stmt.Args {
		if arg == nil {
			continue
		}
		if defaultExpr, ok := arg.Value.(*parser.Node); ok {
			uses = append(uses, b.extractUsesFromExpression(defaultExpr, block, stmt, pos)...)
		}
	}
	return uses
}

func (b *pythonRefExtractor) extractWithContextUses(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var uses []*VarReference
	for _, child := range stmt.Children {
		if child == nil || child.Type != parser.NodeWithItem {
			continue
		}
		if contextExpr := nodeValue(child); contextExpr != nil {
			uses = append(uses, b.extractUsesFromExpression(contextExpr, block, stmt, pos)...)
		}
	}
	return uses
}

func (b *pythonRefExtractor) extractMatchPatternUses(pattern *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if pattern == nil {
		return nil
	}

	var uses []*VarReference
	switch pattern.Type {
	case parser.NodeName, parser.NodeConstant, parser.NodeMatchSingleton, parser.NodeMatchStar:
		return nil
	case parser.NodeAttribute, parser.NodeSubscript, parser.NodeMatchValue:
		uses = append(uses, b.extractPatternValueUses(pattern, block, stmt, pos)...)
	case parser.NodeCall, parser.NodeMatchClass:
		if funcNode := nodeValue(pattern); funcNode != nil {
			uses = append(uses, b.extractClassPatternHeadUses(funcNode, block, stmt, pos)...)
		}
		uses = append(uses, b.extractMatchPatternUsesFromNodes(pattern.Args, block, stmt, pos)...)
		for _, keyword := range pattern.Keywords {
			if keyword == nil {
				continue
			}
			if value := nodeValue(keyword); value != nil {
				uses = append(uses, b.extractMatchPatternUses(value, block, stmt, pos)...)
			}
		}
	case parser.NodeTuple, parser.NodeList, parser.NodeDict, parser.NodeStarred,
		parser.NodeMatchSequence, parser.NodeMatchMapping, parser.NodeMatchAs,
		parser.NodeMatchOr:
		uses = append(uses, b.extractMatchPatternUsesFromNodes(pattern.GetChildren(), block, stmt, pos)...)
	default:
		switch string(pattern.Type) {
		case "attribute":
			uses = append(uses, b.extractDottedPatternUses(pattern, block, stmt, pos)...)
		case "dotted_name":
			if len(pattern.Children) > 1 {
				uses = append(uses, b.extractDottedPatternUses(pattern, block, stmt, pos)...)
			}
		case "class_pattern":
			if len(pattern.Children) > 0 {
				uses = append(uses, b.extractClassPatternHeadUses(pattern.Children[0], block, stmt, pos)...)
				uses = append(uses, b.extractMatchPatternUsesFromNodes(pattern.Children[1:], block, stmt, pos)...)
			}
		case "value_pattern":
			uses = append(uses, b.extractPatternValueUsesFromNodes(pattern.GetChildren(), block, stmt, pos)...)
		case "keyword_pattern":
			if len(pattern.Children) > 0 {
				uses = append(uses, b.extractMatchPatternUses(pattern.Children[len(pattern.Children)-1], block, stmt, pos)...)
			}
		case "wildcard_pattern", "_":
			return nil
		default:
			uses = append(uses, b.extractMatchPatternUsesFromNodes(pattern.GetChildren(), block, stmt, pos)...)
		}
	}

	return uses
}

func (b *pythonRefExtractor) extractMatchPatternUsesFromNodes(nodes []*parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	var uses []*VarReference
	for _, node := range nodes {
		uses = append(uses, b.extractMatchPatternUses(node, block, stmt, pos)...)
	}
	return uses
}

func (b *pythonRefExtractor) extractPatternValueUsesFromNodes(nodes []*parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	var uses []*VarReference
	for _, node := range nodes {
		uses = append(uses, b.extractPatternValueUses(node, block, stmt, pos)...)
	}
	return uses
}

func (b *pythonRefExtractor) extractPatternValueUses(expr *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if expr == nil {
		return nil
	}
	if expr.Type == parser.NodeName {
		return []*VarReference{NewVarReference(expr.Name, UseKindLoad, block, stmt, pos)}
	}
	if expr.Type == parser.NodeAttribute || string(expr.Type) == "attribute" || string(expr.Type) == "dotted_name" {
		return b.extractDottedPatternUses(expr, block, stmt, pos)
	}
	return b.extractUsesFromExpression(expr, block, stmt, pos)
}

func (b *pythonRefExtractor) extractClassPatternHeadUses(expr *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if expr == nil {
		return nil
	}
	if expr.Type == parser.NodeName {
		return []*VarReference{NewVarReference(expr.Name, UseKindCall, block, stmt, pos)}
	}
	return b.extractPatternValueUses(expr, block, stmt, pos)
}

func (b *pythonRefExtractor) extractDottedPatternUses(expr *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if expr == nil {
		return nil
	}
	if expr.Type == parser.NodeAttribute {
		return b.extractUsesFromExpression(expr, block, stmt, pos)
	}
	for _, child := range expr.GetChildren() {
		if child == nil {
			continue
		}
		if child.Type == parser.NodeName {
			return []*VarReference{NewVarReference(child.Name, UseKindAttribute, block, stmt, pos)}
		}
		if uses := b.extractDottedPatternUses(child, block, stmt, pos); len(uses) > 0 {
			return uses
		}
	}
	return nil
}

// extractUsesFromExpression recursively extracts variable uses from an expression
func (b *pythonRefExtractor) extractUsesFromExpression(expr *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	if expr == nil {
		return nil
	}

	var uses []*VarReference

	switch expr.Type {
	case parser.NodeName:
		ref := NewVarReference(expr.Name, UseKindLoad, block, stmt, pos)
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
		if base := nodeValue(expr); base != nil {
			if base.Type == parser.NodeName {
				ref := NewVarReference(base.Name, UseKindSubscript, block, stmt, pos)
				uses = append(uses, ref)
			} else {
				uses = append(uses, b.extractUsesFromExpression(base, block, stmt, pos)...)
			}
		}
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeCall:
		// f(x) - the function and arguments are uses
		if funcNode := nodeValue(expr); funcNode != nil {
			if funcNode.Type == parser.NodeName {
				ref := NewVarReference(funcNode.Name, UseKindCall, block, stmt, pos)
				uses = append(uses, ref)
			} else {
				uses = append(uses, b.extractUsesFromExpression(funcNode, block, stmt, pos)...)
			}
		}
		// Process arguments
		for _, arg := range expr.Args {
			uses = append(uses, b.extractUsesFromExpression(arg, block, stmt, pos)...)
		}
		// Process keyword arguments
		for _, kw := range expr.Keywords {
			for _, child := range kw.GetChildren() {
				uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
			}
		}

	case parser.NodeBinOp:
		// a + b - both operands are uses
		uses = append(uses, b.extractUsesFromExpression(expr.Left, block, stmt, pos)...)
		uses = append(uses, b.extractUsesFromExpression(expr.Right, block, stmt, pos)...)

	case parser.NodeUnaryOp:
		// -x, not x - the operand is a use
		if value := nodeValue(expr); value != nil {
			uses = append(uses, b.extractUsesFromExpression(value, block, stmt, pos)...)
		}

	case parser.NodeCompare:
		// a < b - both operands are uses
		uses = append(uses, b.extractUsesFromExpression(expr.Left, block, stmt, pos)...)
		for _, child := range expr.Children {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeTuple, parser.NodeList, parser.NodeSet:
		// Collection literals - elements are uses
		for _, child := range expr.GetChildren() {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeDict:
		// Dict literal - keys and values are uses
		for _, child := range expr.GetChildren() {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeIfExp:
		// Ternary expression: a if b else c
		for _, child := range expr.GetChildren() {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeLambda:
		// Lambda internals have local parameter scope.
		return uses

	case parser.NodeBoolOp:
		// and/or - operands are uses
		for _, child := range expr.GetChildren() {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}

	case parser.NodeNamedExpr:
		// x := value defines x and reads only the value side.
		if value := nodeValue(expr); value != nil {
			uses = append(uses, b.extractUsesFromExpression(value, block, stmt, pos)...)
		}

	default:
		// For other expression types, recursively process children
		for _, child := range expr.GetChildren() {
			uses = append(uses, b.extractUsesFromExpression(child, block, stmt, pos)...)
		}
	}

	return uses
}

// extractAssignmentDefs extracts definitions from an assignment statement
func (b *pythonRefExtractor) extractAssignmentDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	for _, target := range stmt.Targets {
		defs = append(defs, b.extractNamesFromTarget(target, DefKindAssign, block, stmt, pos)...)
	}

	return defs
}

// extractAugAssignDefs extracts definitions from an augmented assignment
func (b *pythonRefExtractor) extractAugAssignDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	if len(stmt.Targets) > 0 && stmt.Targets[0] != nil {
		target := stmt.Targets[0]
		if target.Type == parser.NodeName {
			ref := NewVarReference(target.Name, DefKindAugAssign, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

// extractAnnAssignDefs extracts definitions from an annotated assignment
func (b *pythonRefExtractor) extractAnnAssignDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
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
func (b *pythonRefExtractor) extractForTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// For loop target is stored in Targets
	for _, target := range stmt.Targets {
		defs = append(defs, b.extractNamesFromTarget(target, DefKindFor, block, stmt, pos)...)
	}

	return defs
}

func (b *pythonRefExtractor) extractNamedBindingDef(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	if stmt.Name == "" {
		return nil
	}
	return []*VarReference{NewVarReference(stmt.Name, DefKindAssign, block, stmt, pos)}
}

// extractParameterDefs extracts parameter definitions from a function
func (b *pythonRefExtractor) extractParameterDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	for _, arg := range stmt.Args {
		if arg != nil && arg.Type == parser.NodeArg && arg.Name != "" {
			ref := NewVarReference(arg.Name, DefKindParam, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

func (b *pythonRefExtractor) extractMatchPatternDefs(pattern *parser.Node, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
	var defs []*VarReference
	b.collectMatchPatternDefs(pattern, block, stmt, pos, &defs)
	return defs
}

func (b *pythonRefExtractor) collectMatchPatternDefs(pattern *parser.Node, block *BasicBlock, stmt *parser.Node, pos int, defs *[]*VarReference) {
	if pattern == nil {
		return
	}

	switch pattern.Type {
	case parser.NodeName:
		if isPatternCaptureName(pattern.Name) {
			*defs = append(*defs, NewVarReference(pattern.Name, DefKindPattern, block, stmt, pos))
		}
	case parser.NodeAttribute, parser.NodeSubscript, parser.NodeMatchValue, parser.NodeMatchSingleton:
		return
	case parser.NodeCall:
		b.collectMatchPatternDefsFromNodes(pattern.Args, block, stmt, pos, defs)
		for _, keyword := range pattern.Keywords {
			if keyword == nil {
				continue
			}
			if value := nodeValue(keyword); value != nil {
				b.collectMatchPatternDefs(value, block, stmt, pos, defs)
			}
		}
	case parser.NodeTuple, parser.NodeList, parser.NodeDict, parser.NodeStarred,
		parser.NodeMatchSequence, parser.NodeMatchMapping, parser.NodeMatchAs,
		parser.NodeMatchOr, parser.NodeMatchStar:
		b.collectMatchPatternDefsFromNodes(pattern.GetChildren(), block, stmt, pos, defs)
	default:
		switch string(pattern.Type) {
		case "attribute", "wildcard_pattern", "_":
			return
		case "dotted_name":
			if len(pattern.Children) == 1 {
				b.collectMatchPatternDefs(pattern.Children[0], block, stmt, pos, defs)
			}
		case "class_pattern":
			if len(pattern.Children) > 1 {
				b.collectMatchPatternDefsFromNodes(pattern.Children[1:], block, stmt, pos, defs)
			}
		case "keyword_pattern":
			if len(pattern.Children) > 0 {
				b.collectMatchPatternDefs(pattern.Children[len(pattern.Children)-1], block, stmt, pos, defs)
			}
		default:
			b.collectMatchPatternDefsFromNodes(pattern.GetChildren(), block, stmt, pos, defs)
		}
	}
}

func (b *pythonRefExtractor) collectMatchPatternDefsFromNodes(nodes []*parser.Node, block *BasicBlock, stmt *parser.Node, pos int, defs *[]*VarReference) {
	for _, node := range nodes {
		b.collectMatchPatternDefs(node, block, stmt, pos, defs)
	}
}

func isPatternCaptureName(name string) bool {
	return name != "" && name != "_"
}

func nodeValue(node *parser.Node) *parser.Node {
	if node == nil {
		return nil
	}
	value, _ := node.Value.(*parser.Node)
	return value
}

// extractImportDefs extracts definitions from an import statement
func (b *pythonRefExtractor) extractImportDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
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
func (b *pythonRefExtractor) extractImportFromDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
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
func (b *pythonRefExtractor) extractWithTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// Look for WithItem children
	for _, child := range stmt.Children {
		if child == nil || child.Type != parser.NodeWithItem {
			continue
		}
		if child.Target != nil {
			defs = append(defs, b.extractNamesFromTarget(child.Target, DefKindWith, block, stmt, pos)...)
		} else if child.Name != "" {
			ref := NewVarReference(child.Name, DefKindWith, block, stmt, pos)
			defs = append(defs, ref)
		}
	}

	return defs
}

// extractExceptTargetDefs extracts definitions from an except handler
func (b *pythonRefExtractor) extractExceptTargetDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
	var defs []*VarReference

	// ExceptHandler has a name field for "as x"
	if stmt.Name != "" {
		ref := NewVarReference(stmt.Name, DefKindExcept, block, stmt, pos)
		defs = append(defs, ref)
	}

	return defs
}

// extractNamedExprDefs extracts definitions from a named expression (walrus operator)
func (b *pythonRefExtractor) extractNamedExprDefs(stmt *parser.Node, block *BasicBlock, pos int) []*VarReference {
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
func (b *pythonRefExtractor) extractNamesFromTarget(target *parser.Node, kind DefUseKind, block *BasicBlock, stmt *parser.Node, pos int) []*VarReference {
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
		// Handle tree-sitter pattern node types that buildNode falls through as generic nodes:
		//   - pattern_list:        a, b = 1, 2
		//   - tuple_pattern:       with cm() as (a, b):
		//   - list_pattern:        with cm() as [a, b]:
		//   - list_splat_pattern:  *rest inside the above patterns
		switch string(target.Type) {
		case "pattern_list", "tuple_pattern", "list_pattern":
			for _, elem := range target.Children {
				if elem != nil && elem.Type != "," && string(elem.Type) != "," {
					defs = append(defs, b.extractNamesFromTarget(elem, kind, block, stmt, pos)...)
				}
			}
		case "list_splat_pattern", "list_splat":
			for _, elem := range target.Children {
				if elem != nil && elem.Type != "*" && string(elem.Type) != "*" {
					defs = append(defs, b.extractNamesFromTarget(elem, kind, block, stmt, pos)...)
				}
			}
		}
	}

	return defs
}
