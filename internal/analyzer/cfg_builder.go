package analyzer

import (
	"fmt"
	"github.com/pyqol/pyqol/internal/parser"
)

// CFGBuilder builds control flow graphs from AST nodes
type CFGBuilder struct {
	// cfg is the control flow graph being built
	cfg *CFG
	
	// currentBlock is the block currently being populated
	currentBlock *BasicBlock
	
	// scopeStack tracks nested scopes (functions, classes, etc.)
	scopeStack []string
	
	// functionCFGs stores CFGs for nested functions
	functionCFGs map[string]*CFG
	
	// blockCounter for generating unique block names
	blockCounter int
}

// NewCFGBuilder creates a new CFG builder
func NewCFGBuilder() *CFGBuilder {
	return &CFGBuilder{
		scopeStack:   []string{},
		functionCFGs: make(map[string]*CFG),
		blockCounter: 0,
	}
}

// Build constructs a CFG from an AST node
func (b *CFGBuilder) Build(node *parser.Node) (*CFG, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot build CFG from nil node")
	}
	
	// Initialize CFG based on node type
	cfgName := "main"
	if node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef {
		cfgName = node.Name
	} else if node.Type == parser.NodeClassDef {
		cfgName = node.Name
	}
	
	b.cfg = NewCFG(cfgName)
	b.currentBlock = b.cfg.Entry
	
	// Build CFG based on node type
	switch node.Type {
	case parser.NodeModule:
		b.buildModule(node)
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		b.buildFunction(node)
	case parser.NodeClassDef:
		b.buildClass(node)
	default:
		// For single statements, process directly
		b.processStatement(node)
	}
	
	// Connect current block to exit if not already connected
	if b.currentBlock != nil && b.currentBlock != b.cfg.Exit && !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(b.currentBlock, b.cfg.Exit, EdgeNormal)
	}
	
	return b.cfg, nil
}

// BuildAll builds CFGs for all functions in the AST
func (b *CFGBuilder) BuildAll(node *parser.Node) (map[string]*CFG, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot build CFGs from nil node")
	}
	
	allCFGs := make(map[string]*CFG)
	
	// Build main CFG
	mainCFG, err := b.Build(node)
	if err != nil {
		return nil, err
	}
	allCFGs["__main__"] = mainCFG
	
	// Add all function CFGs (including nested ones)
	for name, cfg := range b.functionCFGs {
		allCFGs[name] = cfg
	}
	
	// Also process top-level functions directly
	if node.Type == parser.NodeModule {
		for _, stmt := range node.Body {
			if stmt.Type == parser.NodeFunctionDef || stmt.Type == parser.NodeAsyncFunctionDef {
				// Check if we already have this function
				if _, exists := allCFGs[stmt.Name]; !exists {
					funcBuilder := NewCFGBuilder()
					funcCFG, err := funcBuilder.Build(stmt)
					if err == nil {
						allCFGs[stmt.Name] = funcCFG
						// Add nested functions from this function
						for nestedName, nestedCFG := range funcBuilder.functionCFGs {
							fullName := stmt.Name + "." + nestedName
							allCFGs[fullName] = nestedCFG
						}
					}
				}
			}
		}
	}
	
	return allCFGs, nil
}

// buildModule processes a module node
func (b *CFGBuilder) buildModule(node *parser.Node) {
	// Process all statements in the module body
	for _, stmt := range node.Body {
		b.processStatement(stmt)
	}
}

// buildFunction processes a function definition
func (b *CFGBuilder) buildFunction(node *parser.Node) {
	// Enter function scope
	b.enterScope(node.Name)
	defer b.exitScope()
	
	// Create a new block for function body
	bodyBlock := b.createBlock("func_body")
	b.cfg.ConnectBlocks(b.currentBlock, bodyBlock, EdgeNormal)
	b.currentBlock = bodyBlock
	
	// Add function definition as a statement
	b.currentBlock.AddStatement(node)
	
	// Process function body
	for _, stmt := range node.Body {
		// Check for nested functions and build their CFGs
		if stmt.Type == parser.NodeFunctionDef || stmt.Type == parser.NodeAsyncFunctionDef {
			b.buildNestedFunction(stmt)
		} else {
			b.processStatement(stmt)
		}
	}
}

// buildClass processes a class definition
func (b *CFGBuilder) buildClass(node *parser.Node) {
	// Enter class scope
	b.enterScope(node.Name)
	defer b.exitScope()
	
	// Create a new block for class body
	bodyBlock := b.createBlock("class_body")
	b.cfg.ConnectBlocks(b.currentBlock, bodyBlock, EdgeNormal)
	b.currentBlock = bodyBlock
	
	// Add class definition as a statement
	b.currentBlock.AddStatement(node)
	
	// Process class body (methods and attributes)
	for _, stmt := range node.Body {
		if stmt.Type == parser.NodeFunctionDef || stmt.Type == parser.NodeAsyncFunctionDef {
			// Build separate CFG for methods
			b.buildNestedFunction(stmt)
		} else {
			// Process other statements (assignments, etc.)
			b.processStatement(stmt)
		}
	}
}

// buildNestedFunction builds a separate CFG for a nested function
func (b *CFGBuilder) buildNestedFunction(node *parser.Node) {
	// Create a new builder for the nested function
	nestedBuilder := NewCFGBuilder()
	nestedBuilder.scopeStack = append([]string{}, b.scopeStack...)
	
	// Build CFG for the nested function
	funcCFG, err := nestedBuilder.Build(node)
	if err == nil {
		// Store the function CFG
		fullName := b.getFullScopeName(node.Name)
		b.functionCFGs[fullName] = funcCFG
	}
	
	// Add function definition to current block
	b.currentBlock.AddStatement(node)
}

// processStatement processes a single statement
func (b *CFGBuilder) processStatement(stmt *parser.Node) {
	if stmt == nil {
		return
	}
	
	switch stmt.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		// Build separate CFG for nested functions
		b.buildNestedFunction(stmt)
		
	case parser.NodeClassDef:
		// Build separate scope for nested classes
		b.buildClass(stmt)
		
	case parser.NodeReturn:
		// Add return statement and connect to exit
		b.currentBlock.AddStatement(stmt)
		b.cfg.ConnectBlocks(b.currentBlock, b.cfg.Exit, EdgeReturn)
		// Create unreachable block for any following code
		unreachableBlock := b.createBlock("unreachable")
		b.currentBlock = unreachableBlock
		
	case parser.NodePass:
		// Pass statement - just add to current block
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeAssign, parser.NodeAugAssign, parser.NodeAnnAssign:
		// Assignment statements
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeExpr:
		// Expression statements
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeImport, parser.NodeImportFrom:
		// Import statements
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeGlobal, parser.NodeNonlocal:
		// Global/nonlocal declarations
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeDelete:
		// Delete statements
		b.currentBlock.AddStatement(stmt)
		
	case parser.NodeAssert:
		// Assert statements (for now, treat as sequential)
		b.currentBlock.AddStatement(stmt)
		
	default:
		// For control flow statements (if, for, while, etc.)
		// These will be handled in the next issues
		// For now, just add them as statements
		b.currentBlock.AddStatement(stmt)
	}
}

// createBlock creates a new basic block
func (b *CFGBuilder) createBlock(label string) *BasicBlock {
	b.blockCounter++
	blockLabel := fmt.Sprintf("%s_%d", label, b.blockCounter)
	return b.cfg.CreateBlock(blockLabel)
}

// enterScope enters a new scope
func (b *CFGBuilder) enterScope(name string) {
	b.scopeStack = append(b.scopeStack, name)
}

// exitScope exits the current scope
func (b *CFGBuilder) exitScope() {
	if len(b.scopeStack) > 0 {
		b.scopeStack = b.scopeStack[:len(b.scopeStack)-1]
	}
}

// getFullScopeName returns the fully qualified scope name
func (b *CFGBuilder) getFullScopeName(name string) string {
	if len(b.scopeStack) == 0 {
		return name
	}
	
	fullName := ""
	for _, scope := range b.scopeStack {
		if fullName == "" {
			fullName = scope
		} else {
			fullName = fullName + "." + scope
		}
	}
	
	if fullName == "" {
		return name
	}
	return fullName + "." + name
}

// hasSuccessor checks if a block has a specific successor
func (b *CFGBuilder) hasSuccessor(from, to *BasicBlock) bool {
	for _, edge := range from.Successors {
		if edge.To == to {
			return true
		}
	}
	return false
}

// getCurrentScope returns the current scope name
func (b *CFGBuilder) getCurrentScope() string {
	if len(b.scopeStack) == 0 {
		return ""
	}
	return b.scopeStack[len(b.scopeStack)-1]
}