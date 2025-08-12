package analyzer

import (
	"strings"
	"testing"
)

func TestCFGBuilderWithStatements(t *testing.T) {
	t.Run("SimpleWithStatement", func(t *testing.T) {
		source := `
with open("file.txt") as f:
    content = f.read()
    print(content)
print("after with")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check for with statement blocks
		hasWithSetup := false
		hasWithBody := false
		hasWithTeardown := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "with_setup") {
					hasWithSetup = true
				}
				if strings.Contains(b.Label, "with_body") {
					hasWithBody = true
				}
				if strings.Contains(b.Label, "with_teardown") {
					hasWithTeardown = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasWithSetup {
			t.Error("Missing with setup block")
		}
		if !hasWithBody {
			t.Error("Missing with body block")
		}
		if !hasWithTeardown {
			t.Error("Missing with teardown block")
		}

		// Check for exception edge from setup to teardown
		hasExceptionEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException &&
					e.From != nil && strings.Contains(e.From.Label, "with_setup") &&
					e.To != nil && strings.Contains(e.To.Label, "with_teardown") {
					hasExceptionEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasExceptionEdge {
			t.Error("Missing exception edge from setup to teardown")
		}
	})

	t.Run("MultipleWithItems", func(t *testing.T) {
		source := `
with open("input.txt") as input_file, open("output.txt", "w") as output_file:
    data = input_file.read()
    output_file.write(data.upper())
print("files processed")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should still have the basic with statement structure
		hasWithSetup := false
		hasWithBody := false
		hasWithTeardown := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "with_setup") {
					hasWithSetup = true
				}
				if strings.Contains(b.Label, "with_body") {
					hasWithBody = true
				}
				if strings.Contains(b.Label, "with_teardown") {
					hasWithTeardown = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasWithSetup {
			t.Error("Missing with setup block for multiple items")
		}
		if !hasWithBody {
			t.Error("Missing with body block for multiple items")
		}
		if !hasWithTeardown {
			t.Error("Missing with teardown block for multiple items")
		}
	})

	t.Run("NestedWithStatements", func(t *testing.T) {
		source := `
with open("outer.txt") as outer:
    with open("inner.txt") as inner:
        combined = outer.read() + inner.read()
        print(combined)
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have multiple with structures
		withSetupCount := 0
		withBodyCount := 0
		withTeardownCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "with_setup") {
					withSetupCount++
				}
				if strings.Contains(b.Label, "with_body") {
					withBodyCount++
				}
				if strings.Contains(b.Label, "with_teardown") {
					withTeardownCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if withSetupCount < 2 {
			t.Errorf("Expected at least 2 with setup blocks for nested with, got %d", withSetupCount)
		}
		if withBodyCount < 2 {
			t.Errorf("Expected at least 2 with body blocks for nested with, got %d", withBodyCount)
		}
		if withTeardownCount < 2 {
			t.Errorf("Expected at least 2 with teardown blocks for nested with, got %d", withTeardownCount)
		}
	})

	t.Run("AsyncWithStatement", func(t *testing.T) {
		source := `
async def process_async():
    async with AsyncContextManager() as manager:
        await manager.do_work()
        result = await manager.get_result()
        return result
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have the same structure as regular with
		hasWithSetup := false
		hasWithBody := false
		hasWithTeardown := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "with_setup") {
					hasWithSetup = true
				}
				if strings.Contains(b.Label, "with_body") {
					hasWithBody = true
				}
				if strings.Contains(b.Label, "with_teardown") {
					hasWithTeardown = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasWithSetup {
			t.Error("Missing async with setup block")
		}
		if !hasWithBody {
			t.Error("Missing async with body block")
		}
		if !hasWithTeardown {
			t.Error("Missing async with teardown block")
		}
	})
}

func TestCFGBuilderMatchStatements(t *testing.T) {
	t.Run("SimpleMatchStatement", func(t *testing.T) {
		source := `
def handle_value(value):
    match value:
        case 0:
            return "zero"
        case 1:
            return "one"
        case _:
            return "other"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check for match evaluation and merge blocks
		hasMatchEval := false
		hasMatchMerge := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "match_eval") {
					hasMatchEval = true
				}
				if strings.Contains(b.Label, "match_merge") {
					hasMatchMerge = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasMatchEval {
			t.Error("Missing match evaluation block")
		}
		if !hasMatchMerge {
			t.Error("Missing match merge block")
		}

		// Count match case blocks
		matchCaseCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "match_case") {
					matchCaseCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if matchCaseCount != 3 {
			t.Errorf("Expected 3 match case blocks, got %d", matchCaseCount)
		}

		// Check for conditional edges
		hasCondTrue := false
		hasCondFalse := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue {
					hasCondTrue = true
				}
				if e.Type == EdgeCondFalse {
					hasCondFalse = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasCondTrue {
			t.Error("Missing conditional true edges for match cases")
		}
		if !hasCondFalse {
			t.Error("Missing conditional false edge for default case")
		}
	})

	t.Run("ComplexMatchPatterns", func(t *testing.T) {
		source := `
def analyze_data(data):
    match data:
        case {"type": "user", "id": user_id}:
            return f"User {user_id}"
        case {"type": "admin", "permissions": perms} if perms > 5:
            return "High-privilege admin"
        case [first, *rest] if len(rest) > 0:
            return f"List starting with {first}"
        case int(x) if x > 100:
            return "Large number"
        case str(s) if s.startswith("test_"):
            return "Test string"
        case _:
            return "Unknown data type"
    print("after match")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have 6 case blocks
		matchCaseCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "match_case") {
					matchCaseCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if matchCaseCount != 6 {
			t.Errorf("Expected 6 match case blocks, got %d", matchCaseCount)
		}

		// Complex structure should have many blocks
		if cfg.Size() < 10 {
			t.Errorf("Expected at least 10 blocks for complex match, got %d", cfg.Size())
		}
	})

	t.Run("NestedMatchStatements", func(t *testing.T) {
		source := `
def nested_match(outer, inner):
    match outer:
        case "A":
            match inner:
                case 1:
                    return "A1"
                case 2:
                    return "A2"
                case _:
                    return "A_other"
        case "B":
            return "B"
        case _:
            return "other"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have multiple match structures
		matchEvalCount := 0
		matchMergeCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "match_eval") {
					matchEvalCount++
				}
				if strings.Contains(b.Label, "match_merge") {
					matchMergeCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if matchEvalCount < 2 {
			t.Errorf("Expected at least 2 match eval blocks for nested match, got %d", matchEvalCount)
		}
		if matchMergeCount < 2 {
			t.Errorf("Expected at least 2 match merge blocks for nested match, got %d", matchMergeCount)
		}
	})

	t.Run("EmptyMatchStatement", func(t *testing.T) {
		source := `
def empty_match(value):
    match value:
        case _:
            pass
    return "done"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should still create match evaluation and merge blocks
		hasMatchEval := false
		hasMatchMerge := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "match_eval") {
					hasMatchEval = true
				}
				if strings.Contains(b.Label, "match_merge") {
					hasMatchMerge = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasMatchEval {
			t.Error("Missing match evaluation block for empty match")
		}
		if !hasMatchMerge {
			t.Error("Missing match merge block for empty match")
		}
	})
}

func TestCFGBuilderAsyncConstructs(t *testing.T) {
	t.Run("AwaitExpressions", func(t *testing.T) {
		source := `
async def async_function():
    result1 = await async_call1()
    result2 = await async_call2(result1)
    return await final_call(result2)
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Await expressions should be treated as regular expressions
		// Check that statements are properly added to blocks
		totalStatements := countStatements(cfg)
		if totalStatements < 4 { // function def + 3 assignments/return
			t.Errorf("Expected at least 4 statements, got %d", totalStatements)
		}

		// Should have normal flow structure
		hasNormalEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeNormal {
					hasNormalEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasNormalEdge {
			t.Error("Missing normal edges for await expressions")
		}
	})

	t.Run("YieldExpressions", func(t *testing.T) {
		source := `
def generator_function():
    for i in range(3):
        yield i * 2
    yield from other_generator()
    return "done"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Yield expressions should be treated as regular statements
		// Should have loop structure + yield statements
		hasLoopHeader := false
		hasLoopBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "loop_body") {
					hasLoopBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasLoopHeader {
			t.Error("Missing loop header for generator with yield")
		}
		if !hasLoopBody {
			t.Error("Missing loop body for generator with yield")
		}

		// Check for return edge
		hasReturnEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeReturn {
					hasReturnEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasReturnEdge {
			t.Error("Missing return edge in generator")
		}
	})

	t.Run("AsyncGeneratorFunction", func(t *testing.T) {
		source := `
async def async_generator():
    async for item in async_iterable():
        processed = await process_item(item)
        yield processed
    return "async generator done"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should combine async for loop structure with yield
		hasLoopHeader := false
		hasLoopBody := false
		hasReturnEdge := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "loop_body") {
					hasLoopBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeReturn {
					hasReturnEdge = true
				}
				return true
			},
		})

		if !hasLoopHeader {
			t.Error("Missing loop header for async generator")
		}
		if !hasLoopBody {
			t.Error("Missing loop body for async generator")
		}
		if !hasReturnEdge {
			t.Error("Missing return edge for async generator")
		}
	})
}

func TestCFGBuilderAdvancedIntegration(t *testing.T) {
	t.Run("CombinedAdvancedConstructs", func(t *testing.T) {
		source := `
async def complex_processing(data_source):
    try:
        async with AsyncContextManager() as manager:
            async for item in data_source:
                match item.type:
                    case "critical":
                        await manager.process_critical(item)
                    case "normal":
                        await manager.process_normal(item)
                    case _:
                        continue
                yield await manager.get_result()
    except ProcessingError:
        raise RuntimeError("Processing failed")
    return "completed"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should be a complex CFG with many blocks
		if cfg.Size() < 15 {
			t.Errorf("Expected at least 15 blocks for complex integration, got %d", cfg.Size())
		}

		// Check for presence of different constructs
		hasTryBlock := false
		hasExceptBlock := false
		hasWithSetup := false
		hasLoopHeader := false
		hasMatchEval := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "except_block") {
					hasExceptBlock = true
				}
				if strings.Contains(b.Label, "with_setup") {
					hasWithSetup = true
				}
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "match_eval") {
					hasMatchEval = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasTryBlock {
			t.Error("Missing try block in complex integration")
		}
		if !hasExceptBlock {
			t.Error("Missing except block in complex integration")
		}
		if !hasWithSetup {
			t.Error("Missing with setup in complex integration")
		}
		if !hasLoopHeader {
			t.Error("Missing loop header in complex integration")
		}
		if !hasMatchEval {
			t.Error("Missing match evaluation in complex integration")
		}

		// Check for different edge types
		edgeTypes := make(map[EdgeType]bool)
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				edgeTypes[e.Type] = true
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		expectedEdgeTypes := []EdgeType{
			EdgeNormal, EdgeCondTrue, EdgeCondFalse,
			EdgeException, EdgeReturn, EdgeLoop, EdgeContinue,
		}

		for _, expectedType := range expectedEdgeTypes {
			if !edgeTypes[expectedType] {
				t.Errorf("Missing edge type %s in complex integration", expectedType.String())
			}
		}
	})
}
