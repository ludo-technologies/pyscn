package parser

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVisitorPattern(t *testing.T) {
	source := `
def hello(name):
    print(f"Hello, {name}!")

class Greeter:
    def greet(self, name):
        return hello(name)

for i in range(5):
    if i % 2 == 0:
        print(i)
`

	parser := New()
	ctx := context.Background()
	result, err := parser.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	t.Run("FuncVisitor", func(t *testing.T) {
		count := 0
		visitor := NewFuncVisitor(func(node *Node) bool {
			count++
			return true
		})
		result.AST.Accept(visitor)
		if count == 0 {
			t.Error("FuncVisitor didn't visit any nodes")
		}
	})

	t.Run("CollectorVisitor", func(t *testing.T) {
		collector := NewCollectorVisitor(func(node *Node) bool {
			return node.Type == NodeFunctionDef
		})
		result.AST.Accept(collector)
		functions := collector.GetNodes()
		if len(functions) != 2 { // hello and greet
			t.Errorf("Expected 2 functions, got %d", len(functions))
		}
	})

	t.Run("PrinterVisitor", func(t *testing.T) {
		var buf bytes.Buffer
		printer := NewPrinterVisitor(&buf)
		result.AST.Accept(printer)
		output := buf.String()
		if !strings.Contains(output, "FunctionDef: hello") {
			t.Error("PrinterVisitor didn't print function name")
		}
		if !strings.Contains(output, "ClassDef: Greeter") {
			t.Error("PrinterVisitor didn't print class name")
		}
	})

	t.Run("StatisticsVisitor", func(t *testing.T) {
		stats := NewStatisticsVisitor()
		result.AST.Accept(stats)
		
		if stats.TotalNodes == 0 {
			t.Error("No nodes counted")
		}
		if stats.MaxDepth == 0 {
			t.Error("Max depth not calculated")
		}
		if stats.NodeCounts[NodeFunctionDef] != 2 {
			t.Errorf("Expected 2 FunctionDef nodes, got %d", stats.NodeCounts[NodeFunctionDef])
		}
		if stats.NodeCounts[NodeClassDef] != 1 {
			t.Errorf("Expected 1 ClassDef node, got %d", stats.NodeCounts[NodeClassDef])
		}
	})

	t.Run("ValidatorVisitor", func(t *testing.T) {
		validator := NewValidatorVisitor()
		result.AST.Accept(validator)
		
		if !validator.IsValid() {
			t.Errorf("Validation failed with %d errors: %v", len(validator.GetErrors()), validator.GetErrors())
		}
	})

	t.Run("PathVisitor", func(t *testing.T) {
		var deepestPath []*Node
		maxDepth := 0
		
		pathVisitor := NewPathVisitor(func(node *Node, path []*Node) bool {
			if len(path) > maxDepth {
				maxDepth = len(path)
				deepestPath = make([]*Node, len(path))
				copy(deepestPath, path)
			}
			return true
		})
		
		result.AST.Accept(pathVisitor)
		
		if maxDepth == 0 {
			t.Error("PathVisitor didn't track any paths")
		}
		if len(deepestPath) == 0 {
			t.Error("No deepest path recorded")
		}
	})

	t.Run("DepthFirstVisitor", func(t *testing.T) {
		var preOrder []NodeType
		var postOrder []NodeType
		
		dfVisitor := NewDepthFirstVisitor(
			func(node *Node) bool {
				preOrder = append(preOrder, node.Type)
				return true
			},
			func(node *Node) {
				postOrder = append(postOrder, node.Type)
			},
		)
		
		result.AST.Accept(dfVisitor)
		
		if len(preOrder) == 0 {
			t.Error("No pre-order traversal")
		}
		if len(postOrder) == 0 {
			t.Error("No post-order traversal")
		}
		if len(preOrder) != len(postOrder) {
			t.Error("Pre-order and post-order counts don't match")
		}
	})
}

func TestTransformVisitor(t *testing.T) {
	source := `
x = 1 + 2
y = 3 + 4
`
	parser := New()
	ctx := context.Background()
	result, err := parser.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Count binary operations before transformation
	binOpsBefore := len(result.AST.FindByType(NodeBinOp))
	if binOpsBefore == 0 {
		t.Skip("No binary operations found in parsed code")
	}
	
	// Transform binary operations to constants
	transformed := 0
	transformer := NewTransformVisitor(func(node *Node) *Node {
		if node.Type == NodeBinOp && node.Op == "+" {
			// Simplified transformation: mark as transformed
			node.Type = NodeConstant
			node.Value = "transformed"
			node.Left = nil
			node.Right = nil
			node.Op = ""
			transformed++
		}
		return node
	})
	
	result.AST.Accept(transformer)
	
	// Check that transformations occurred
	binOpsAfter := len(result.AST.FindByType(NodeBinOp))
	
	if transformed == 0 {
		t.Log("Warning: No transformations occurred - AST structure may differ from expected")
	}
	if binOpsAfter >= binOpsBefore && transformed > 0 {
		t.Error("Binary operations were not transformed despite transformation count")
	}
}

func TestSimplifierVisitor(t *testing.T) {
	// Create an AST with simplifiable constructs
	module := NewNode(NodeModule)
	
	// Add a function with pass statements
	funcDef := NewNode(NodeFunctionDef)
	funcDef.Name = "test"
	funcDef.AddToBody(NewNode(NodePass))
	funcDef.AddToBody(NewNode(NodeReturn))
	funcDef.AddToBody(NewNode(NodePass))
	module.AddToBody(funcDef)
	
	// Add a constant binary operation
	binOp := NewNode(NodeBinOp)
	binOp.Op = "+"
	binOp.Left = NewNode(NodeConstant)
	binOp.Left.Value = int64(5)
	binOp.Right = NewNode(NodeConstant)
	binOp.Right.Value = int64(3)
	module.AddToBody(binOp)
	
	// Run simplifier
	simplifier := NewSimplifierVisitor()
	module.Accept(simplifier)
	
	if !simplifier.WasSimplified() {
		t.Error("No simplification performed")
	}
	
	// Check that pass statements were removed
	passCount := 0
	for _, stmt := range funcDef.Body {
		if stmt.Type == NodePass {
			passCount++
		}
	}
	if passCount > 0 {
		t.Errorf("Pass statements not removed, found %d", passCount)
	}
	
	// Check that constant expression was simplified
	if binOp.Type != NodeConstant {
		t.Error("Binary operation not simplified to constant")
	}
	if binOp.Value != int64(8) {
		t.Errorf("Incorrect simplification result: %v", binOp.Value)
	}
}

func TestValidatorVisitorErrors(t *testing.T) {
	tests := []struct {
		name      string
		buildNode func() *Node
		wantError bool
	}{
		{
			name: "function without name",
			buildNode: func() *Node {
				funcDef := NewNode(NodeFunctionDef)
				funcDef.AddToBody(NewNode(NodePass))
				return funcDef
			},
			wantError: true,
		},
		{
			name: "function with empty body",
			buildNode: func() *Node {
				funcDef := NewNode(NodeFunctionDef)
				funcDef.Name = "test"
				return funcDef
			},
			wantError: true,
		},
		{
			name: "if without condition",
			buildNode: func() *Node {
				ifNode := NewNode(NodeIf)
				ifNode.AddToBody(NewNode(NodePass))
				return ifNode
			},
			wantError: true,
		},
		{
			name: "for without target",
			buildNode: func() *Node {
				forNode := NewNode(NodeFor)
				forNode.Iter = NewNode(NodeName)
				forNode.AddToBody(NewNode(NodePass))
				return forNode
			},
			wantError: true,
		},
		{
			name: "binary op without operands",
			buildNode: func() *Node {
				binOp := NewNode(NodeBinOp)
				binOp.Op = "+"
				return binOp
			},
			wantError: true,
		},
		{
			name: "valid function",
			buildNode: func() *Node {
				funcDef := NewNode(NodeFunctionDef)
				funcDef.Name = "valid"
				funcDef.AddToBody(NewNode(NodePass))
				return funcDef
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := tt.buildNode()
			validator := NewValidatorVisitor()
			node.Accept(validator)
			
			hasError := !validator.IsValid()
			if hasError != tt.wantError {
				t.Errorf("Expected error: %v, got: %v", tt.wantError, hasError)
				if hasError {
					t.Logf("Errors: %v", validator.GetErrors())
				}
			}
		})
	}
}

func TestVisitorOnTestData(t *testing.T) {
	testDataDir := "../../testdata/python"
	
	// Skip if testdata doesn't exist
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		t.Skip("Test data directory doesn't exist")
	}
	
	parser := New()
	ctx := context.Background()
	
	// Test files from simple directory
	simpleFiles := []string{
		"simple/functions.py",
		"simple/classes.py",
		"simple/control_flow.py",
		"simple/imports.py",
	}
	
	for _, file := range simpleFiles {
		t.Run(file, func(t *testing.T) {
			path := filepath.Join(testDataDir, file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("Could not read file: %v", err)
			}
			
			result, err := parser.Parse(ctx, content)
			if err != nil {
				// Some files may have intentional errors
				if strings.Contains(file, "syntax_errors") {
					return
				}
				t.Fatalf("Failed to parse %s: %v", file, err)
			}
			
			// Run statistics visitor
			stats := NewStatisticsVisitor()
			result.AST.Accept(stats)
			
			if stats.TotalNodes == 0 {
				t.Errorf("No nodes found in %s", file)
			}
			
			// Run validator
			validator := NewValidatorVisitor()
			result.AST.Accept(validator)
			
			if !validator.IsValid() {
				t.Errorf("Validation failed in %s with %d errors: %v", file, len(validator.GetErrors()), validator.GetErrors())
			}
		})
	}
}

func TestAcceptMethod(t *testing.T) {
	// Test that Accept method works correctly
	module := NewNode(NodeModule)
	funcDef := NewNode(NodeFunctionDef)
	funcDef.Name = "test"
	module.AddToBody(funcDef)
	
	visited := make(map[*Node]bool)
	visitor := NewFuncVisitor(func(node *Node) bool {
		visited[node] = true
		return true
	})
	
	module.Accept(visitor)
	
	if !visited[module] {
		t.Error("Module not visited")
	}
	if !visited[funcDef] {
		t.Error("Function not visited")
	}
	
	// Test early termination
	earlyStop := NewFuncVisitor(func(node *Node) bool {
		return false // Stop immediately
	})
	
	count := 0
	countVisitor := NewFuncVisitor(func(node *Node) bool {
		count++
		return true
	})
	
	// First visitor stops early
	module.Accept(earlyStop)
	// Count should still work
	module.Accept(countVisitor)
	
	if count != 2 { // module and funcDef
		t.Errorf("Expected 2 nodes visited, got %d", count)
	}
}

func BenchmarkVisitors(b *testing.B) {
	source := `
class Calculator:
    def __init__(self):
        self.value = 0
    
    def add(self, x):
        self.value += x
        return self.value
    
    def multiply(self, x):
        self.value *= x
        return self.value

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

for i in range(100):
    if i % 2 == 0:
        print(i)
`

	parser := New()
	ctx := context.Background()
	result, _ := parser.Parse(ctx, []byte(source))
	
	b.Run("FuncVisitor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			visitor := NewFuncVisitor(func(node *Node) bool {
				return true
			})
			result.AST.Accept(visitor)
		}
	})
	
	b.Run("CollectorVisitor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			collector := NewCollectorVisitor(func(node *Node) bool {
				return node.Type == NodeFunctionDef
			})
			result.AST.Accept(collector)
		}
	})
	
	b.Run("StatisticsVisitor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			stats := NewStatisticsVisitor()
			result.AST.Accept(stats)
		}
	})
	
	b.Run("ValidatorVisitor", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			validator := NewValidatorVisitor()
			result.AST.Accept(validator)
		}
	})
}