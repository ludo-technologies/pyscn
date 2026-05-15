package parser

import (
	"context"
	"testing"
)

func TestASTBuilder(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		wantErr    bool
		checkNodes func(*testing.T, *Node)
	}{
		{
			name: "simple function",
			source: `def hello():
    print("Hello, World!")`,
			wantErr: false,
			checkNodes: func(t *testing.T, ast *Node) {
				if ast.Type != NodeModule {
					t.Errorf("Expected Module node, got %s", ast.Type)
				}
				if len(ast.Body) != 1 {
					t.Errorf("Expected 1 statement in body, got %d", len(ast.Body))
				}
				if ast.Body[0].Type != NodeFunctionDef {
					t.Errorf("Expected FunctionDef, got %s", ast.Body[0].Type)
				}
				if ast.Body[0].Name != "hello" {
					t.Errorf("Expected function name 'hello', got %s", ast.Body[0].Name)
				}
			},
		},
		{
			name: "class definition",
			source: `class MyClass:
    def __init__(self):
        self.value = 42`,
			wantErr: false,
			checkNodes: func(t *testing.T, ast *Node) {
				if ast.Type != NodeModule {
					t.Errorf("Expected Module node, got %s", ast.Type)
				}
				if len(ast.Body) != 1 {
					t.Errorf("Expected 1 statement in body, got %d", len(ast.Body))
				}
				if ast.Body[0].Type != NodeClassDef {
					t.Errorf("Expected ClassDef, got %s", ast.Body[0].Type)
				}
				if ast.Body[0].Name != "MyClass" {
					t.Errorf("Expected class name 'MyClass', got %s", ast.Body[0].Name)
				}
			},
		},
		{
			name: "if statement",
			source: `if x > 0:
    print("positive")
else:
    print("non-positive")`,
			wantErr: false,
			checkNodes: func(t *testing.T, ast *Node) {
				if ast.Type != NodeModule {
					t.Errorf("Expected Module node, got %s", ast.Type)
				}
				if len(ast.Body) != 1 {
					t.Errorf("Expected 1 statement in body, got %d", len(ast.Body))
				}
				if ast.Body[0].Type != NodeIf {
					t.Errorf("Expected If node, got %s", ast.Body[0].Type)
				}
				ifNode := ast.Body[0]
				if ifNode.Test == nil {
					t.Error("Expected If node to have a test condition")
				}
				if len(ifNode.Body) == 0 {
					t.Error("Expected If node to have a body")
				}
				if len(ifNode.Orelse) == 0 {
					t.Error("Expected If node to have an else clause")
				}
			},
		},
		{
			name: "for loop",
			source: `for i in range(10):
    print(i)`,
			wantErr: false,
			checkNodes: func(t *testing.T, ast *Node) {
				if ast.Type != NodeModule {
					t.Errorf("Expected Module node, got %s", ast.Type)
				}
				if len(ast.Body) != 1 {
					t.Errorf("Expected 1 statement in body, got %d", len(ast.Body))
				}
				if ast.Body[0].Type != NodeFor {
					t.Errorf("Expected For node, got %s", ast.Body[0].Type)
				}
				forNode := ast.Body[0]
				if len(forNode.Targets) == 0 {
					t.Error("Expected For node to have a target")
				}
				if forNode.Iter == nil {
					t.Error("Expected For node to have an iterator")
				}
				if len(forNode.Body) == 0 {
					t.Error("Expected For node to have a body")
				}
			},
		},
		{
			name: "import statements",
			source: `import os
from sys import path
from collections import defaultdict`,
			wantErr: false,
			checkNodes: func(t *testing.T, ast *Node) {
				if ast.Type != NodeModule {
					t.Errorf("Expected Module node, got %s", ast.Type)
				}
				if len(ast.Body) != 3 {
					t.Errorf("Expected 3 statements in body, got %d", len(ast.Body))
				}
				if ast.Body[0].Type != NodeImport {
					t.Errorf("Expected Import node, got %s", ast.Body[0].Type)
				}
				if ast.Body[1].Type != NodeImportFrom {
					t.Errorf("Expected ImportFrom node, got %s", ast.Body[1].Type)
				}
				if ast.Body[2].Type != NodeImportFrom {
					t.Errorf("Expected ImportFrom node, got %s", ast.Body[2].Type)
				}
			},
		},
	}

	parser := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, []byte(tt.source))

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("Parse() returned nil result")
			}

			if result.AST == nil {
				t.Fatal("ParseResult.AST is nil")
			}

			if tt.checkNodes != nil {
				tt.checkNodes(t, result.AST)
			}
		})
	}
}

func TestNodeMethods(t *testing.T) {
	// Create a simple AST structure for testing
	module := NewNode(NodeModule)

	funcDef := NewNode(NodeFunctionDef)
	funcDef.Name = "test_func"
	module.AddToBody(funcDef)

	ifStmt := NewNode(NodeIf)
	funcDef.AddToBody(ifStmt)

	returnStmt := NewNode(NodeReturn)
	ifStmt.AddToBody(returnStmt)

	// Test IsStatement
	if !funcDef.IsStatement() {
		t.Error("FunctionDef should be a statement")
	}
	if module.IsStatement() {
		t.Error("Module should not be a statement")
	}

	// Test IsControlFlow
	if !ifStmt.IsControlFlow() {
		t.Error("If should be control flow")
	}
	if funcDef.IsControlFlow() {
		t.Error("FunctionDef should not be control flow")
	}

	// Test FindByType
	functions := module.FindByType(NodeFunctionDef)
	if len(functions) != 1 {
		t.Errorf("Expected 1 function, found %d", len(functions))
	}

	returns := module.FindByType(NodeReturn)
	if len(returns) != 1 {
		t.Errorf("Expected 1 return statement, found %d", len(returns))
	}

	// Test GetParentOfType
	parent := returnStmt.GetParentOfType(NodeFunctionDef)
	if parent != funcDef {
		t.Error("Expected to find parent FunctionDef")
	}

	// Test Walk
	nodeCount := 0
	module.Walk(func(n *Node) bool {
		nodeCount++
		return true
	})
	if nodeCount != 4 { // module, funcDef, ifStmt, returnStmt
		t.Errorf("Expected 4 nodes in walk, got %d", nodeCount)
	}

	// Test Copy
	copied := module.Copy()
	if copied == module {
		t.Error("Copy should create a new instance")
	}
	if copied.Type != module.Type {
		t.Error("Copied node should have same type")
	}
	if len(copied.Body) != len(module.Body) {
		t.Error("Copied node should have same body length")
	}
}

func TestCanonicalTraversalIncludesValueNodes(t *testing.T) {
	source := `
async def load(client, items, idx):
    result = client.fetch(items[idx])
    await client.commit(result)
    yield result
    return client.done(result)
`

	result, err := New().Parse(context.Background(), []byte(source))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	calls := result.AST.FindByType(NodeCall)
	callByAttr := make(map[string]*Node)
	for _, call := range calls {
		callee, ok := call.Value.(*Node)
		if ok && callee.Type == NodeAttribute {
			callByAttr[callee.Name] = call
			if callee.Parent != call {
				t.Fatalf("Callee %s parent = %v, want call", callee.Name, callee.Parent)
			}
		}
	}
	for _, name := range []string{"fetch", "commit", "done"} {
		if callByAttr[name] == nil {
			t.Fatalf("FindByType(NodeCall) did not find call to client.%s", name)
		}
	}

	if returnNode := callByAttr["done"].GetParentOfType(NodeReturn); returnNode == nil {
		t.Fatal("Call stored in Return.Value should resolve its return parent")
	}

	subscripts := result.AST.FindByType(NodeSubscript)
	if len(subscripts) != 1 {
		t.Fatalf("Expected 1 subscript, got %d", len(subscripts))
	}
	base, ok := subscripts[0].Value.(*Node)
	if !ok || base.Type != NodeName || base.Name != "items" {
		t.Fatalf("Subscript base = %T %v, want Name(items)", subscripts[0].Value, subscripts[0].Value)
	}
	if base.Parent != subscripts[0] {
		t.Fatal("Subscript base parent should point to the subscript node")
	}

	awaits := result.AST.FindByType(NodeAwait)
	if len(awaits) != 1 {
		t.Fatalf("Expected 1 await node, got %d", len(awaits))
	}
	if valueNode, ok := awaits[0].Value.(*Node); !ok || valueNode.Parent != awaits[0] {
		t.Fatal("Await value should be traversable and parented")
	}

	yields := result.AST.FindByType(NodeYield)
	if len(yields) != 1 {
		t.Fatalf("Expected 1 yield node, got %d", len(yields))
	}
	if valueNode, ok := yields[0].Value.(*Node); !ok || valueNode.Parent != yields[0] {
		t.Fatal("Yield value should be traversable and parented")
	}
}

func TestCopyDeepCopiesValueNodes(t *testing.T) {
	source := `
def call():
    return foo()
`

	result, err := New().Parse(context.Background(), []byte(source))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	copied := result.AST.Copy()
	originalCall := firstCallWithName(t, result.AST, "foo")
	copiedCall := firstCallWithName(t, copied, "foo")

	copiedCallee := copiedCall.Value.(*Node)
	copiedCallee.Name = "bar"

	originalCallee := originalCall.Value.(*Node)
	if originalCallee.Name != "foo" {
		t.Fatalf("Original callee mutated to %q", originalCallee.Name)
	}
	if copiedCallee.Parent != copiedCall {
		t.Fatal("Copied value node should point to copied parent")
	}
}

func firstCallWithName(t *testing.T, ast *Node, name string) *Node {
	t.Helper()

	for _, call := range ast.FindByType(NodeCall) {
		callee, ok := call.Value.(*Node)
		if ok && callee.Type == NodeName && callee.Name == name {
			return call
		}
	}
	t.Fatalf("Call %s() not found", name)
	return nil
}

func TestASTBuilderComplexCode(t *testing.T) {
	source := `
import sys
from typing import List, Optional

class Calculator:
    """A simple calculator class."""
    
    def __init__(self, initial: float = 0):
        self.value = initial
    
    def add(self, x: float) -> float:
        """Add a value."""
        self.value += x
        return self.value
    
    def multiply(self, x: float) -> float:
        """Multiply by a value."""
        self.value *= x
        return self.value

def fibonacci(n: int) -> int:
    """Calculate fibonacci number."""
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

async def fetch_data(url: str) -> dict:
    """Fetch data from URL."""
    async with session.get(url) as response:
        return await response.json()

# Main execution
if __name__ == "__main__":
    calc = Calculator(10)
    result = calc.add(5)
    print(f"Result: {result}")
    
    for i in range(5):
        print(f"Fib({i}) = {fibonacci(i)}")
`

	parser := New()
	ctx := context.Background()

	result, err := parser.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse complex code: %v", err)
	}

	if result.AST == nil {
		t.Fatal("AST is nil")
	}

	// Check module structure
	if result.AST.Type != NodeModule {
		t.Errorf("Expected Module, got %s", result.AST.Type)
	}

	// Find all function definitions
	functions := result.AST.FindByType(NodeFunctionDef)
	asyncFunctions := result.AST.FindByType(NodeAsyncFunctionDef)
	totalFunctions := len(functions) + len(asyncFunctions)

	if totalFunctions < 3 { // __init__, add, multiply, fibonacci, fetch_data
		t.Errorf("Expected at least 3 functions, found %d", totalFunctions)
	}

	// Find class definitions
	classes := result.AST.FindByType(NodeClassDef)
	if len(classes) != 1 {
		t.Errorf("Expected 1 class, found %d", len(classes))
	}

	// Find import statements
	imports := result.AST.FindByType(NodeImport)
	importFroms := result.AST.FindByType(NodeImportFrom)
	if len(imports)+len(importFroms) < 2 {
		t.Errorf("Expected at least 2 import statements, found %d", len(imports)+len(importFroms))
	}

	// Find if statements
	ifStmts := result.AST.FindByType(NodeIf)
	if len(ifStmts) < 2 { // One in fibonacci, one for __main__
		t.Errorf("Expected at least 2 if statements, found %d", len(ifStmts))
	}

	// Find for loops
	forLoops := result.AST.FindByType(NodeFor)
	if len(forLoops) < 1 {
		t.Errorf("Expected at least 1 for loop, found %d", len(forLoops))
	}
}

func TestASTBuilderFStringInterpolationsPreserveExpressions(t *testing.T) {
	source := `
class FStringExamples:
    def embedded(self):
        return f"x{self._sep}y"

    def simple(self):
        return f"{self._value}"

    def format_spec(self):
        return f"x{1:{self._width}>5}y"

    def right_aligned(self):
        return f"{self._value:>5}"

    def left_aligned(self):
        return f"{self._value:<5}"

    def conversion(self):
        return f"{self._value!r}"

    def debug_value(self):
        return f"{self._value=}"

    def nested(self):
        return f"{f'{self._nested}'}"

    def plain_string(self):
        return "x{self._plain}y"
`

	result, err := New().Parse(context.Background(), []byte(source))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	returns := result.AST.FindByType(NodeReturn)
	if len(returns) != 9 {
		t.Fatalf("Expected 9 return statements, got %d", len(returns))
	}

	tests := []struct {
		name string
		idx  int
		attr string
	}{
		{name: "embedded interpolation", idx: 0, attr: "_sep"},
		{name: "simple interpolation", idx: 1, attr: "_value"},
		{name: "format spec expression", idx: 2, attr: "_width"},
		{name: "right aligned static format spec", idx: 3, attr: "_value"},
		{name: "left aligned static format spec", idx: 4, attr: "_value"},
		{name: "conversion marker", idx: 5, attr: "_value"},
		{name: "debug marker", idx: 6, attr: "_value"},
		{name: "nested f-string expression", idx: 7, attr: "_nested"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := returns[tt.idx].Value.(*Node)
			if !ok {
				t.Fatalf("Return value is %T, want *Node", returns[tt.idx].Value)
			}
			if value.Type != NodeJoinedStr {
				t.Fatalf("Return value type = %s, want %s", value.Type, NodeJoinedStr)
			}

			formattedValues := 0
			foundAttr := false
			value.WalkDeep(func(node *Node) bool {
				if node.Type == NodeFormattedValue {
					formattedValues++
				}
				if node.Type == NodeAttribute && node.Name == tt.attr {
					foundAttr = true
				}
				return true
			})

			if formattedValues == 0 {
				t.Fatal("Expected at least one formatted value")
			}
			if !foundAttr {
				t.Fatalf("Expected deep traversal to find self.%s", tt.attr)
			}
		})
	}

	for _, tt := range []struct {
		name string
		idx  int
		want string
	}{
		{name: "dynamic format spec keeps static suffix", idx: 2, want: ">5"},
		{name: "right aligned static format spec", idx: 3, want: ":>5"},
		{name: "left aligned static format spec", idx: 4, want: ":<5"},
		{name: "conversion marker", idx: 5, want: "!r"},
		{name: "debug marker", idx: 6, want: "="},
	} {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := returns[tt.idx].Value.(*Node)
			if !ok {
				t.Fatalf("Return value is %T, want *Node", returns[tt.idx].Value)
			}

			foundLiteral := false
			value.WalkDeep(func(node *Node) bool {
				if node.Type == NodeConstant && node.Value == tt.want {
					foundLiteral = true
				}
				return true
			})

			if !foundLiteral {
				t.Fatalf("Expected deep traversal to find format spec literal %q", tt.want)
			}
		})
	}

	for _, tt := range []struct {
		name string
		idx  int
		dup  string
	}{
		{name: "simple interpolation expression is not duplicated", idx: 1, dup: "self._value"},
		{name: "static format spec expression is not duplicated", idx: 3, dup: "self._value"},
		{name: "dynamic format spec outer expression is not duplicated", idx: 2, dup: "1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := returns[tt.idx].Value.(*Node)
			if !ok {
				t.Fatalf("Return value is %T, want *Node", returns[tt.idx].Value)
			}

			if countStringConstants(value, tt.dup) != 0 {
				t.Fatalf("Expression literal %q was duplicated as a string constant", tt.dup)
			}
		})
	}

	plainValue, ok := returns[8].Value.(*Node)
	if !ok {
		t.Fatalf("Plain string return value is %T, want *Node", returns[8].Value)
	}
	if plainValue.Type != NodeConstant {
		t.Fatalf("Plain string type = %s, want %s", plainValue.Type, NodeConstant)
	}
}

func TestASTBuilderWithItemsPreserveContextExpressions(t *testing.T) {
	source := `
def sync_copy(path_a, path_b, lock):
    with open(path_a) as src, open(path_b, "w") as dst:
        return dst.write(src.read())
    with lock:
        return None

async def async_acquire(resource):
    async with resource as ctx:
        return ctx
`

	result, err := New().Parse(context.Background(), []byte(source))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	withStmts := result.AST.FindByType(NodeWith)
	if len(withStmts) != 2 {
		t.Fatalf("Expected 2 with statements, got %d", len(withStmts))
	}

	if len(withStmts[0].Children) != 2 {
		t.Fatalf("Expected first with statement to have 2 items, got %d", len(withStmts[0].Children))
	}
	for i, wantName := range []string{"src", "dst"} {
		item := withStmts[0].Children[i]
		if item.Type != NodeWithItem {
			t.Fatalf("With child %d type = %s, want %s", i, item.Type, NodeWithItem)
		}
		if item.Name != wantName {
			t.Fatalf("With item %d alias = %q, want %q", i, item.Name, wantName)
		}
		value, ok := item.Value.(*Node)
		if !ok {
			t.Fatalf("With item %d value is %T, want *Node", i, item.Value)
		}
		if value.Type != NodeCall {
			t.Fatalf("With item %d value type = %s, want %s", i, value.Type, NodeCall)
		}
	}

	if len(withStmts[1].Children) != 1 {
		t.Fatalf("Expected second with statement to have 1 item, got %d", len(withStmts[1].Children))
	}
	lockValue, ok := withStmts[1].Children[0].Value.(*Node)
	if !ok {
		t.Fatalf("Direct with item value is %T, want *Node", withStmts[1].Children[0].Value)
	}
	if lockValue.Type != NodeName || lockValue.Name != "lock" {
		t.Fatalf("Direct with item value = %s(%s), want Name(lock)", lockValue.Type, lockValue.Name)
	}

	asyncWithStmts := result.AST.FindByType(NodeAsyncWith)
	if len(asyncWithStmts) != 1 {
		t.Fatalf("Expected 1 async with statement, got %d", len(asyncWithStmts))
	}
	if len(asyncWithStmts[0].Children) != 1 {
		t.Fatalf("Expected async with statement to have 1 item, got %d", len(asyncWithStmts[0].Children))
	}
	asyncItem := asyncWithStmts[0].Children[0]
	if asyncItem.Name != "ctx" {
		t.Fatalf("Async with item alias = %q, want %q", asyncItem.Name, "ctx")
	}
	asyncValue, ok := asyncItem.Value.(*Node)
	if !ok {
		t.Fatalf("Async with item value is %T, want *Node", asyncItem.Value)
	}
	if asyncValue.Type != NodeName || asyncValue.Name != "resource" {
		t.Fatalf("Async with item value = %s(%s), want Name(resource)", asyncValue.Type, asyncValue.Name)
	}
}

func TestASTBuilderWithItemTargetField(t *testing.T) {
	source := `
def unpack(cm):
    with cm() as f:
        pass
    with cm() as (a, b):
        pass
    with cm() as [c, *rest]:
        pass
`

	result, err := New().Parse(context.Background(), []byte(source))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}

	withStmts := result.AST.FindByType(NodeWith)
	if len(withStmts) != 3 {
		t.Fatalf("Expected 3 with statements, got %d", len(withStmts))
	}

	// 1) simple identifier alias → Name is populated, Target stays nil
	// (the alias is already represented by Name and the apted label; setting Target
	// would duplicate it in generic AST traversal and inflate clone fingerprints).
	itemSimple := withStmts[0].Children[0]
	if itemSimple.Name != "f" {
		t.Fatalf("Simple alias Name = %q, want %q", itemSimple.Name, "f")
	}
	if itemSimple.Target != nil {
		t.Fatalf("Simple alias should not populate Target, got %s(%s)", itemSimple.Target.Type, itemSimple.Target.Name)
	}

	// 2) tuple alias → Target is a Tuple node containing two Name children
	itemTuple := withStmts[1].Children[0]
	if itemTuple.Target == nil {
		t.Fatalf("Expected Target on tuple alias")
	}
	if itemTuple.Target.Type != NodeTuple {
		t.Fatalf("Tuple alias Target type = %s, want %s", itemTuple.Target.Type, NodeTuple)
	}
	tupleNames := collectNameLeaves(itemTuple.Target)
	if got, want := tupleNames, []string{"a", "b"}; !equalStringSlices(got, want) {
		t.Fatalf("Tuple alias names = %v, want %v", got, want)
	}
	// Name should NOT be populated for compound aliases — that was the silent-drop case.
	if itemTuple.Name != "" {
		t.Fatalf("Compound alias should not populate Name, got %q", itemTuple.Name)
	}

	// 3) list alias with starred → Target is a List node; *rest is captured
	itemList := withStmts[2].Children[0]
	if itemList.Target == nil {
		t.Fatalf("Expected Target on list alias")
	}
	if itemList.Target.Type != NodeList {
		t.Fatalf("List alias Target type = %s, want %s", itemList.Target.Type, NodeList)
	}
	listNames := collectNameLeaves(itemList.Target)
	if got, want := listNames, []string{"c", "rest"}; !equalStringSlices(got, want) {
		t.Fatalf("List alias names = %v, want %v", got, want)
	}
}

func collectNameLeaves(n *Node) []string {
	var out []string
	n.WalkDeep(func(child *Node) bool {
		if child.Type == NodeName {
			out = append(out, child.Name)
		}
		return true
	})
	return out
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func countStringConstants(node *Node, value string) int {
	count := 0
	node.WalkDeep(func(child *Node) bool {
		if child.Type == NodeConstant && child.Value == value {
			count++
		}
		return true
	})
	return count
}
