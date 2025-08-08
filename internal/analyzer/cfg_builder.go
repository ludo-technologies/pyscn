package analyzer

import (
	"fmt"
	"log"
	"strings"
	"github.com/pyqol/pyqol/internal/parser"
)

// Block label constants to avoid magic strings
const (
	LabelFunctionBody   = "func_body"
	LabelClassBody      = "class_body"
	LabelUnreachable    = "unreachable"
	LabelMainModule     = "main"
	LabelEntry          = "ENTRY"
	LabelExit           = "EXIT"
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
	blockCounter uint
	
	// logger for error reporting (optional)
	logger *log.Logger
}

// NewCFGBuilder creates a new CFG builder
func NewCFGBuilder() *CFGBuilder {
	return &CFGBuilder{
		scopeStack:   []string{},
		functionCFGs: make(map[string]*CFG),
		blockCounter: 0,
		logger:       nil, // Can be set via SetLogger if needed
	}
}

// SetLogger sets an optional logger for error reporting
func (b *CFGBuilder) SetLogger(logger *log.Logger) {
	b.logger = logger
}

// logError logs an error if a logger is set
func (b *CFGBuilder) logError(format string, args ...interface{}) {
	if b.logger != nil {
		b.logger.Printf("CFGBuilder: "+format, args...)
	}
}

// Build constructs a CFG from an AST node
func (b *CFGBuilder) Build(node *parser.Node) (*CFG, error) {
	if node == nil {
		return nil, fmt.Errorf("cannot build CFG from nil node")
	}
	
	// Initialize CFG based on node type
	cfgName := LabelMainModule
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
	bodyBlock := b.createBlock(LabelFunctionBody)
	b.cfg.ConnectBlocks(b.currentBlock, bodyBlock, EdgeNormal)
	b.currentBlock = bodyBlock
	
	// Add function definition as a statement
	b.currentBlock.AddStatement(node)
	
	// Process function body
	for _, stmt := range node.Body {
		// Check for nested functions and build their CFGs
		if stmt.Type == parser.NodeFunctionDef || stmt.Type == parser.NodeAsyncFunctionDef {
			if err := b.buildNestedFunction(stmt); err != nil {
				b.logError("error in nested function: %v", err)
			}
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
	bodyBlock := b.createBlock(LabelClassBody)
	b.cfg.ConnectBlocks(b.currentBlock, bodyBlock, EdgeNormal)
	b.currentBlock = bodyBlock
	
	// Add class definition as a statement
	b.currentBlock.AddStatement(node)
	
	// Process class body (methods and attributes)
	for _, stmt := range node.Body {
		if stmt.Type == parser.NodeFunctionDef || stmt.Type == parser.NodeAsyncFunctionDef {
			// Build separate CFG for methods
			if err := b.buildNestedFunction(stmt); err != nil {
				b.logError("error building CFG for method %s: %v", stmt.Name, err)
			}
		} else {
			// Process other statements (assignments, etc.)
			b.processStatement(stmt)
		}
	}
}

// buildNestedFunction builds a separate CFG for a nested function
func (b *CFGBuilder) buildNestedFunction(node *parser.Node) error {
	// Create a new builder for the nested function
	nestedBuilder := NewCFGBuilder()
	
	// Efficiently copy scope stack
	nestedBuilder.scopeStack = make([]string, len(b.scopeStack))
	copy(nestedBuilder.scopeStack, b.scopeStack)
	
	// Copy logger if set
	nestedBuilder.logger = b.logger
	
	// Build CFG for the nested function
	funcCFG, err := nestedBuilder.Build(node)
	if err != nil {
		// Log error but don't fail the entire build
		b.logError("failed to build CFG for nested function %s: %v", node.Name, err)
		// Still add the function definition to current block
		b.currentBlock.AddStatement(node)
		return fmt.Errorf("failed to build nested function %s: %w", node.Name, err)
	}
	
	// Store the function CFG
	fullName := b.getFullScopeName(node.Name)
	b.functionCFGs[fullName] = funcCFG
	
	// Add function definition to current block
	b.currentBlock.AddStatement(node)
	return nil
}

// processStatement processes a single statement
func (b *CFGBuilder) processStatement(stmt *parser.Node) {
	if stmt == nil {
		return
	}
	
	switch stmt.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		// Build separate CFG for nested functions
		if err := b.buildNestedFunction(stmt); err != nil {
			b.logError("error in nested function: %v", err)
		}
		
	case parser.NodeClassDef:
		// Build separate scope for nested classes
		b.buildClass(stmt)
		
	case parser.NodeReturn:
		// Add return statement and connect to exit
		b.currentBlock.AddStatement(stmt)
		b.cfg.ConnectBlocks(b.currentBlock, b.cfg.Exit, EdgeReturn)
		// Create unreachable block for any code following the return statement.
		// This block will not be connected to the exit, making it truly unreachable
		// in the CFG, which helps with dead code detection in later analysis phases.
		unreachableBlock := b.createBlock(LabelUnreachable)
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
	
	// Use strings.Join for efficient concatenation
	scopePath := strings.Join(b.scopeStack, ".")
	if scopePath == "" {
		return name
	}
	return scopePath + "." + name
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