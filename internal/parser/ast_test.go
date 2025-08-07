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