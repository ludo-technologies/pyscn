package analyzer

import (
	"fmt"
	"github.com/pyqol/pyqol/internal/parser"
	"log"
	"strings"
)

// Block label constants to avoid magic strings
const (
	LabelFunctionBody = "func_body"
	LabelClassBody    = "class_body"
	LabelUnreachable  = "unreachable"
	LabelMainModule   = "main"
	LabelEntry        = "ENTRY"
	LabelExit         = "EXIT"

	// Loop-related labels
	LabelLoopHeader = "loop_header"
	LabelLoopBody   = "loop_body"
	LabelLoopExit   = "loop_exit"
	LabelLoopElse   = "loop_else"

	// Exception-related labels
	LabelTryBlock     = "try_block"
	LabelExceptBlock  = "except_block"
	LabelFinallyBlock = "finally_block"
	LabelTryElse      = "try_else"

	// Advanced construct labels
	LabelWithSetup    = "with_setup"
	LabelWithBody     = "with_body"
	LabelWithTeardown = "with_teardown"
	LabelMatchCase    = "match_case"
	LabelMatchMerge   = "match_merge"
)

// loopContext tracks the context of a loop for break/continue handling
type loopContext struct {
	headerBlock *BasicBlock // Loop condition/iterator block
	exitBlock   *BasicBlock // Loop exit point
	elseBlock   *BasicBlock // Loop else clause (optional)
	loopType    string      // "for" or "while"
}

// exceptionContext tracks the context of a try block for exception handling
type exceptionContext struct {
	tryBlock     *BasicBlock   // Try block
	finallyBlock *BasicBlock   // Finally block (optional)
	handlers     []*BasicBlock // Exception handler blocks
	elseBlock    *BasicBlock   // Try else clause (optional)
}

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

	// loopStack tracks nested loops for break/continue handling
	loopStack []*loopContext

	// exceptionStack tracks nested try blocks for exception handling
	exceptionStack []*exceptionContext
}

// NewCFGBuilder creates a new CFG builder
func NewCFGBuilder() *CFGBuilder {
	return &CFGBuilder{
		scopeStack:     []string{},
		functionCFGs:   make(map[string]*CFG),
		blockCounter:   0,
		logger:         nil, // Can be set via SetLogger if needed
		loopStack:      []*loopContext{},
		exceptionStack: []*exceptionContext{},
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

	// Process all statements in the function body
	for _, stmt := range node.Body {
		b.processStatement(stmt)
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

	case parser.NodeIf:
		// Handle if/elif/else statements
		b.processIfStatement(stmt)

	case parser.NodeIfExp:
		// Handle conditional expressions (ternary operators)
		// For now, treat as a simple expression
		b.currentBlock.AddStatement(stmt)

	case parser.NodeFor, parser.NodeAsyncFor:
		// Handle for and async for loops
		b.processForStatement(stmt)

	case parser.NodeWhile:
		// Handle while loops
		b.processWhileStatement(stmt)

	case parser.NodeBreak:
		// Handle break statements
		b.processBreakStatement(stmt)

	case parser.NodeContinue:
		// Handle continue statements
		b.processContinueStatement(stmt)

	case parser.NodeTry:
		// Handle try/except/else/finally statements
		b.processTryStatement(stmt)

	case parser.NodeRaise:
		// Handle raise statements
		b.processRaiseStatement(stmt)

	case parser.NodeWith, parser.NodeAsyncWith:
		// Handle with and async with statements
		b.processWithStatement(stmt)

	case parser.NodeMatch:
		// Handle match statements (Python 3.10+)
		b.processMatchStatement(stmt)

	case parser.NodeAwait:
		// Handle await expressions - treat as expression in current block
		b.currentBlock.AddStatement(stmt)

	case parser.NodeYield, parser.NodeYieldFrom:
		// Handle yield expressions - treat as expression (potential suspend point)
		b.currentBlock.AddStatement(stmt)

	default:
		// For other statements, just add them to current block
		b.currentBlock.AddStatement(stmt)
	}
}

// processIfStatement handles if/elif/else statements
func (b *CFGBuilder) processIfStatement(stmt *parser.Node) {
	// Save the condition block (current block where the test happens)
	conditionBlock := b.currentBlock

	// Add the test condition to the current block
	conditionBlock.AddStatement(stmt)

	// Create blocks for the then branch
	thenBlock := b.createBlock("if_then")

	// Create merge block only for the outermost if
	// elif chains will share the same final merge block
	mergeBlock := b.createBlock("if_merge")

	// Connect condition block to then block (true branch)
	b.cfg.ConnectBlocks(conditionBlock, thenBlock, EdgeCondTrue)

	// Process the then branch
	b.currentBlock = thenBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}

	// Save the end of then branch
	thenEndBlock := b.currentBlock

	// Process else/elif branches
	if len(stmt.Orelse) > 0 {
		// Check if the else clause is another if (elif)
		if len(stmt.Orelse) == 1 && stmt.Orelse[0].Type == parser.NodeIf {
			// This is an elif - handle it specially
			elifBlock := b.createBlock("elif")
			b.cfg.ConnectBlocks(conditionBlock, elifBlock, EdgeCondFalse)
			b.currentBlock = elifBlock

			// Process elif recursively - it will create its own test, then, and possibly else
			b.processIfStatementElif(stmt.Orelse[0], mergeBlock)

			// Connect then branch to merge if not already connected to exit
			if !b.hasSuccessor(thenEndBlock, b.cfg.Exit) {
				b.cfg.ConnectBlocks(thenEndBlock, mergeBlock, EdgeNormal)
			}
		} else {
			// This is a regular else clause
			elseBlock := b.createBlock("if_else")
			b.cfg.ConnectBlocks(conditionBlock, elseBlock, EdgeCondFalse)

			// Process else branch
			b.currentBlock = elseBlock
			for _, elseStmt := range stmt.Orelse {
				b.processStatement(elseStmt)
			}

			// Connect both branches to merge if not already connected to exit
			if !b.hasSuccessor(thenEndBlock, b.cfg.Exit) {
				b.cfg.ConnectBlocks(thenEndBlock, mergeBlock, EdgeNormal)
			}
			if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
				b.cfg.ConnectBlocks(b.currentBlock, mergeBlock, EdgeNormal)
			}
		}
	} else {
		// No else clause - connect false branch directly to merge
		b.cfg.ConnectBlocks(conditionBlock, mergeBlock, EdgeCondFalse)

		// Connect then branch to merge if not already connected to exit
		if !b.hasSuccessor(thenEndBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(thenEndBlock, mergeBlock, EdgeNormal)
		}
	}

	// Continue with merge block
	b.currentBlock = mergeBlock
}

// processIfStatementElif handles elif chains specially to maintain proper structure
func (b *CFGBuilder) processIfStatementElif(stmt *parser.Node, finalMerge *BasicBlock) {
	// Current block is the elif block
	conditionBlock := b.currentBlock

	// Add the test condition
	conditionBlock.AddStatement(stmt)

	// Create then block for this elif
	thenBlock := b.createBlock("elif_then")
	b.cfg.ConnectBlocks(conditionBlock, thenBlock, EdgeCondTrue)

	// Process the then branch
	b.currentBlock = thenBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}
	thenEndBlock := b.currentBlock

	// Process else/elif branches
	if len(stmt.Orelse) > 0 {
		// Check if this is another elif
		if len(stmt.Orelse) == 1 && stmt.Orelse[0].Type == parser.NodeIf {
			// Another elif - recurse
			elifBlock := b.createBlock("elif")
			b.cfg.ConnectBlocks(conditionBlock, elifBlock, EdgeCondFalse)
			b.currentBlock = elifBlock
			b.processIfStatementElif(stmt.Orelse[0], finalMerge)
		} else {
			// Final else clause
			elseBlock := b.createBlock("elif_else")
			b.cfg.ConnectBlocks(conditionBlock, elseBlock, EdgeCondFalse)

			b.currentBlock = elseBlock
			for _, elseStmt := range stmt.Orelse {
				b.processStatement(elseStmt)
			}

			// Connect else to final merge
			if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
				b.cfg.ConnectBlocks(b.currentBlock, finalMerge, EdgeNormal)
			}
		}
	} else {
		// No more elif/else - connect false branch to final merge
		b.cfg.ConnectBlocks(conditionBlock, finalMerge, EdgeCondFalse)
	}

	// Connect then branch to final merge
	if !b.hasSuccessor(thenEndBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(thenEndBlock, finalMerge, EdgeNormal)
	}

	// Set current to final merge for any subsequent processing
	b.currentBlock = finalMerge
}

// processForStatement handles for and async for loops
func (b *CFGBuilder) processForStatement(stmt *parser.Node) {
	// Create loop header block (iterator evaluation and condition check)
	headerBlock := b.createBlock(LabelLoopHeader)
	b.cfg.ConnectBlocks(b.currentBlock, headerBlock, EdgeNormal)

	// Add the for statement (iterator setup) to header
	headerBlock.AddStatement(stmt)

	// Create loop body block
	bodyBlock := b.createBlock(LabelLoopBody)

	// Create loop exit block (for normal loop completion)
	exitBlock := b.createBlock(LabelLoopExit)

	// Create else block if present
	var elseBlock *BasicBlock
	if len(stmt.Orelse) > 0 {
		elseBlock = b.createBlock(LabelLoopElse)
	}

	// Create loop context for break/continue handling
	loopCtx := &loopContext{
		headerBlock: headerBlock,
		exitBlock:   exitBlock,
		elseBlock:   elseBlock,
		loopType:    "for",
	}
	b.pushLoopContext(loopCtx)
	defer b.popLoopContext()

	// Connect header to body (loop condition true - has more items)
	b.cfg.ConnectBlocks(headerBlock, bodyBlock, EdgeCondTrue)

	// Connect header to exit/else (loop condition false - iterator exhausted)
	if elseBlock != nil {
		b.cfg.ConnectBlocks(headerBlock, elseBlock, EdgeCondFalse)
	} else {
		b.cfg.ConnectBlocks(headerBlock, exitBlock, EdgeCondFalse)
	}

	// Process loop body
	b.currentBlock = bodyBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}

	// Connect body back to header (loop back edge)
	if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(b.currentBlock, headerBlock, EdgeLoop)
	}

	// Process else clause if present
	if elseBlock != nil {
		b.currentBlock = elseBlock
		for _, elseStmt := range stmt.Orelse {
			b.processStatement(elseStmt)
		}
		// Connect else to exit
		if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(b.currentBlock, exitBlock, EdgeNormal)
		}
	}

	// Continue with exit block
	b.currentBlock = exitBlock
}

// processWhileStatement handles while loops
func (b *CFGBuilder) processWhileStatement(stmt *parser.Node) {
	// Create loop header block (condition evaluation)
	headerBlock := b.createBlock(LabelLoopHeader)
	b.cfg.ConnectBlocks(b.currentBlock, headerBlock, EdgeNormal)

	// Add the while statement (condition) to header
	headerBlock.AddStatement(stmt)

	// Create loop body block
	bodyBlock := b.createBlock(LabelLoopBody)

	// Create loop exit block
	exitBlock := b.createBlock(LabelLoopExit)

	// Create else block if present
	var elseBlock *BasicBlock
	if len(stmt.Orelse) > 0 {
		elseBlock = b.createBlock(LabelLoopElse)
	}

	// Create loop context for break/continue handling
	loopCtx := &loopContext{
		headerBlock: headerBlock,
		exitBlock:   exitBlock,
		elseBlock:   elseBlock,
		loopType:    "while",
	}
	b.pushLoopContext(loopCtx)
	defer b.popLoopContext()

	// Connect header to body (condition true)
	b.cfg.ConnectBlocks(headerBlock, bodyBlock, EdgeCondTrue)

	// Connect header to exit/else (condition false)
	if elseBlock != nil {
		b.cfg.ConnectBlocks(headerBlock, elseBlock, EdgeCondFalse)
	} else {
		b.cfg.ConnectBlocks(headerBlock, exitBlock, EdgeCondFalse)
	}

	// Process loop body
	b.currentBlock = bodyBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}

	// Connect body back to header (loop back edge)
	if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(b.currentBlock, headerBlock, EdgeLoop)
	}

	// Process else clause if present
	if elseBlock != nil {
		b.currentBlock = elseBlock
		for _, elseStmt := range stmt.Orelse {
			b.processStatement(elseStmt)
		}
		// Connect else to exit
		if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(b.currentBlock, exitBlock, EdgeNormal)
		}
	}

	// Continue with exit block
	b.currentBlock = exitBlock
}

// processBreakStatement handles break statements
func (b *CFGBuilder) processBreakStatement(stmt *parser.Node) {
	// Add break statement to current block
	b.currentBlock.AddStatement(stmt)

	// Get the innermost loop context
	if len(b.loopStack) == 0 {
		b.logError("break statement outside of loop")
		return
	}

	loopCtx := b.loopStack[len(b.loopStack)-1]

	// Connect to loop exit with break edge
	b.cfg.ConnectBlocks(b.currentBlock, loopCtx.exitBlock, EdgeBreak)

	// Create unreachable block for any code after break
	unreachableBlock := b.createBlock(LabelUnreachable)
	b.currentBlock = unreachableBlock
}

// processContinueStatement handles continue statements
func (b *CFGBuilder) processContinueStatement(stmt *parser.Node) {
	// Add continue statement to current block
	b.currentBlock.AddStatement(stmt)

	// Get the innermost loop context
	if len(b.loopStack) == 0 {
		b.logError("continue statement outside of loop")
		return
	}

	loopCtx := b.loopStack[len(b.loopStack)-1]

	// Connect to loop header with continue edge
	b.cfg.ConnectBlocks(b.currentBlock, loopCtx.headerBlock, EdgeContinue)

	// Create unreachable block for any code after continue
	unreachableBlock := b.createBlock(LabelUnreachable)
	b.currentBlock = unreachableBlock
}

// pushLoopContext pushes a loop context onto the stack
func (b *CFGBuilder) pushLoopContext(ctx *loopContext) {
	b.loopStack = append(b.loopStack, ctx)
}

// popLoopContext pops the top loop context from the stack
func (b *CFGBuilder) popLoopContext() {
	if len(b.loopStack) > 0 {
		b.loopStack = b.loopStack[:len(b.loopStack)-1]
	}
}

// processTryStatement handles try/except/else/finally blocks
func (b *CFGBuilder) processTryStatement(stmt *parser.Node) {
	// Create try block
	tryBlock := b.createBlock(LabelTryBlock)
	b.cfg.ConnectBlocks(b.currentBlock, tryBlock, EdgeNormal)

	// Create exit block (final convergence point)
	exitBlock := b.createBlock("try_exit")

	// Create finally block if present
	var finallyBlock *BasicBlock
	if len(stmt.Finalbody) > 0 {
		finallyBlock = b.createBlock(LabelFinallyBlock)
	}

	// Create else block if present
	var elseBlock *BasicBlock
	if len(stmt.Orelse) > 0 {
		elseBlock = b.createBlock(LabelTryElse)
	}

	// Create handler blocks for each except clause
	var handlers []*BasicBlock
	for i := range stmt.Handlers {
		handlerBlock := b.createBlock(fmt.Sprintf("%s_%d", LabelExceptBlock, i+1))
		handlers = append(handlers, handlerBlock)
	}

	// Create exception context
	exceptionCtx := &exceptionContext{
		tryBlock:     tryBlock,
		finallyBlock: finallyBlock,
		handlers:     handlers,
		elseBlock:    elseBlock,
	}
	b.pushExceptionContext(exceptionCtx)
	defer b.popExceptionContext()

	// Process try body
	b.currentBlock = tryBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}
	tryEndBlock := b.currentBlock

	// If no exception, try flows to else (if present) or finally/exit
	var nextAfterTry *BasicBlock
	if elseBlock != nil {
		nextAfterTry = elseBlock
	} else if finallyBlock != nil {
		nextAfterTry = finallyBlock
	} else {
		nextAfterTry = exitBlock
	}

	if !b.hasSuccessor(tryEndBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(tryEndBlock, nextAfterTry, EdgeNormal)
	}

	// Connect try block to all exception handlers
	for _, handlerBlock := range handlers {
		b.cfg.ConnectBlocks(tryBlock, handlerBlock, EdgeException)
	}

	// Process exception handlers
	for i, handler := range stmt.Handlers {
		handlerBlock := handlers[i]
		b.currentBlock = handlerBlock

		// Add the exception handler node itself
		handlerBlock.AddStatement(handler)

		// Process handler body
		for _, handlerStmt := range handler.Body {
			b.processStatement(handlerStmt)
		}

		// Handler flows to finally (if present) or exit
		var nextAfterHandler *BasicBlock
		if finallyBlock != nil {
			nextAfterHandler = finallyBlock
		} else {
			nextAfterHandler = exitBlock
		}

		if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(b.currentBlock, nextAfterHandler, EdgeNormal)
		}
	}

	// Process else block if present
	if elseBlock != nil {
		b.currentBlock = elseBlock
		for _, elseStmt := range stmt.Orelse {
			b.processStatement(elseStmt)
		}

		// Else flows to finally (if present) or exit
		var nextAfterElse *BasicBlock
		if finallyBlock != nil {
			nextAfterElse = finallyBlock
		} else {
			nextAfterElse = exitBlock
		}

		if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(b.currentBlock, nextAfterElse, EdgeNormal)
		}
	}

	// Process finally block if present
	if finallyBlock != nil {
		b.currentBlock = finallyBlock
		for _, finallyStmt := range stmt.Finalbody {
			b.processStatement(finallyStmt)
		}

		// Finally always flows to exit
		if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
			b.cfg.ConnectBlocks(b.currentBlock, exitBlock, EdgeNormal)
		}
	}

	// Continue with exit block
	b.currentBlock = exitBlock
}

// processRaiseStatement handles raise statements
func (b *CFGBuilder) processRaiseStatement(stmt *parser.Node) {
	// Add raise statement to current block
	b.currentBlock.AddStatement(stmt)

	// Get the innermost exception context
	if len(b.exceptionStack) > 0 {
		exceptionCtx := b.exceptionStack[len(b.exceptionStack)-1]

		// Connect to all exception handlers in the current context
		for _, handler := range exceptionCtx.handlers {
			b.cfg.ConnectBlocks(b.currentBlock, handler, EdgeException)
		}
	} else {
		// No exception context - connect to exit (unhandled exception)
		b.cfg.ConnectBlocks(b.currentBlock, b.cfg.Exit, EdgeException)
	}

	// Create unreachable block for any code after raise
	unreachableBlock := b.createBlock(LabelUnreachable)
	b.currentBlock = unreachableBlock
}

// pushExceptionContext pushes an exception context onto the stack
func (b *CFGBuilder) pushExceptionContext(ctx *exceptionContext) {
	b.exceptionStack = append(b.exceptionStack, ctx)
}

// popExceptionContext pops the top exception context from the stack
func (b *CFGBuilder) popExceptionContext() {
	if len(b.exceptionStack) > 0 {
		b.exceptionStack = b.exceptionStack[:len(b.exceptionStack)-1]
	}
}

// processWithStatement handles with and async with statements
func (b *CFGBuilder) processWithStatement(stmt *parser.Node) {
	// Create setup block (context manager entry)
	setupBlock := b.createBlock(LabelWithSetup)
	b.cfg.ConnectBlocks(b.currentBlock, setupBlock, EdgeNormal)

	// Add the with statement (context manager setup) to setup block
	setupBlock.AddStatement(stmt)

	// Create body block
	bodyBlock := b.createBlock(LabelWithBody)

	// Create teardown block (context manager exit - always executed)
	teardownBlock := b.createBlock(LabelWithTeardown)

	// Create exit block
	exitBlock := b.createBlock("with_exit")

	// Connect setup to body
	b.cfg.ConnectBlocks(setupBlock, bodyBlock, EdgeNormal)

	// Process with body
	b.currentBlock = bodyBlock
	for _, bodyStmt := range stmt.Body {
		b.processStatement(bodyStmt)
	}

	// Connect body to teardown (normal flow)
	if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
		b.cfg.ConnectBlocks(b.currentBlock, teardownBlock, EdgeNormal)
	}

	// Connect setup to teardown (exception flow - context manager should clean up)
	b.cfg.ConnectBlocks(setupBlock, teardownBlock, EdgeException)

	// Process teardown (context manager exit)
	b.currentBlock = teardownBlock

	// Connect teardown to exit
	b.cfg.ConnectBlocks(teardownBlock, exitBlock, EdgeNormal)

	// Continue with exit block
	b.currentBlock = exitBlock
}

// processMatchStatement handles match statements (Python 3.10+)
func (b *CFGBuilder) processMatchStatement(stmt *parser.Node) {
	// Create match evaluation block
	matchBlock := b.createBlock("match_eval")
	b.cfg.ConnectBlocks(b.currentBlock, matchBlock, EdgeNormal)

	// Add the match statement (subject evaluation) to match block
	matchBlock.AddStatement(stmt)

	// Create exit/merge block
	mergeBlock := b.createBlock(LabelMatchMerge)

	// Process match cases
	if len(stmt.Body) > 0 {
		for i, caseNode := range stmt.Body {
			// Create case block
			caseBlock := b.createBlock(fmt.Sprintf("%s_%d", LabelMatchCase, i+1))

			// Connect match evaluation to this case (conditional edge)
			b.cfg.ConnectBlocks(matchBlock, caseBlock, EdgeCondTrue)

			// Process case body
			b.currentBlock = caseBlock

			// Add the case node itself
			caseBlock.AddStatement(caseNode)

			// Process case body statements
			for _, caseStmt := range caseNode.Body {
				b.processStatement(caseStmt)
			}

			// Connect case to merge (if not already connected to exit)
			if !b.hasSuccessor(b.currentBlock, b.cfg.Exit) {
				b.cfg.ConnectBlocks(b.currentBlock, mergeBlock, EdgeNormal)
			}
		}

		// If no case matches, connect match to merge (default case)
		b.cfg.ConnectBlocks(matchBlock, mergeBlock, EdgeCondFalse)
	} else {
		// No cases - connect directly to merge
		b.cfg.ConnectBlocks(matchBlock, mergeBlock, EdgeNormal)
	}

	// Continue with merge block
	b.currentBlock = mergeBlock
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
	if from == nil || to == nil {
		return false
	}
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
