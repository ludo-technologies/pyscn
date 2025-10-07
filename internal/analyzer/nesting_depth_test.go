package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestCalculateMaxNestingDepth_NilNode(t *testing.T) {
	result := CalculateMaxNestingDepth(nil)

	if result.MaxDepth != 0 {
		t.Errorf("Expected max depth 0 for nil node, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_SimpleFunction(t *testing.T) {
	// Create a simple function with no nesting
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "simple_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   3,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 2,
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 0 {
		t.Errorf("Expected max depth 0 for simple function, got %d", result.MaxDepth)
	}
	if result.FunctionName != "simple_func" {
		t.Errorf("Expected function name 'simple_func', got '%s'", result.FunctionName)
	}
}

func TestCalculateMaxNestingDepth_SingleIfStatement(t *testing.T) {
	// Create a function with a single if statement (depth 1)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func_with_if",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   5,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeReturn,
						Location: parser.Location{
							StartLine: 3,
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 1 {
		t.Errorf("Expected max depth 1 for single if statement, got %d", result.MaxDepth)
	}
	if result.DeepestNestingLine != 2 {
		t.Errorf("Expected deepest nesting at line 2, got %d", result.DeepestNestingLine)
	}
}

func TestCalculateMaxNestingDepth_NestedIfStatements(t *testing.T) {
	// Create a function with nested if statements (depth 3)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "nested_ifs",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   10,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeIf,
								Location: parser.Location{
									StartLine: 4,
								},
								Body: []*parser.Node{
									{
										Type: parser.NodeReturn,
										Location: parser.Location{
											StartLine: 5,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 3 {
		t.Errorf("Expected max depth 3 for nested if statements, got %d", result.MaxDepth)
	}
	if result.DeepestNestingLine != 4 {
		t.Errorf("Expected deepest nesting at line 4, got %d", result.DeepestNestingLine)
	}
}

func TestCalculateMaxNestingDepth_ForAndWhileLoop(t *testing.T) {
	// Create a function with for loop nested inside while loop (depth 2)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "nested_loops",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   8,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeWhile,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeFor,
						Location: parser.Location{
							StartLine: 3,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
								Location: parser.Location{
									StartLine: 4,
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 2 {
		t.Errorf("Expected max depth 2 for nested loops, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_TryExcept(t *testing.T) {
	// Create a function with try-except (depth 2: try + except handler)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "try_except_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   8,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeTry,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodePass,
						Location: parser.Location{
							StartLine: 3,
						},
					},
				},
				Handlers: []*parser.Node{
					{
						Type: parser.NodeExceptHandler,
						Location: parser.Location{
							StartLine: 4,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
								Location: parser.Location{
									StartLine: 5,
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	// Try statement is depth 1, except handler is depth 2
	if result.MaxDepth != 2 {
		t.Errorf("Expected max depth 2 for try-except, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_WithStatement(t *testing.T) {
	// Create a function with nested with statements (depth 2)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "with_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeWith,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeWith,
						Location: parser.Location{
							StartLine: 3,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
								Location: parser.Location{
									StartLine: 4,
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 2 {
		t.Errorf("Expected max depth 2 for nested with statements, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_ComplexNesting(t *testing.T) {
	// Create a function with complex nesting: if > for > while > try (depth 4)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "complex_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   15,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeFor,
						Location: parser.Location{
							StartLine: 3,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeWhile,
								Location: parser.Location{
									StartLine: 4,
								},
								Body: []*parser.Node{
									{
										Type: parser.NodeTry,
										Location: parser.Location{
											StartLine: 5,
										},
										Body: []*parser.Node{
											{
												Type: parser.NodePass,
												Location: parser.Location{
													StartLine: 6,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 4 {
		t.Errorf("Expected max depth 4 for complex nesting, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_ListComprehension(t *testing.T) {
	// Create a function with list comprehension (depth 1)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "comprehension_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   3,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 2,
				},
				Children: []*parser.Node{
					{
						Type: parser.NodeListComp,
						Location: parser.Location{
							StartLine: 2,
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 1 {
		t.Errorf("Expected max depth 1 for list comprehension, got %d", result.MaxDepth)
	}
}

func TestCalculateMaxNestingDepth_LambdaExpression(t *testing.T) {
	// Create a function with lambda expression (depth 1)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "lambda_func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   3,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 2,
				},
				Children: []*parser.Node{
					{
						Type: parser.NodeLambda,
						Location: parser.Location{
							StartLine: 2,
						},
					},
				},
			},
		},
	}

	result := CalculateMaxNestingDepth(funcNode)

	if result.MaxDepth != 1 {
		t.Errorf("Expected max depth 1 for lambda expression, got %d", result.MaxDepth)
	}
}

func TestIsNestingNode(t *testing.T) {
	tests := []struct {
		name     string
		nodeType parser.NodeType
		expected bool
	}{
		{"If statement", parser.NodeIf, true},
		{"For loop", parser.NodeFor, true},
		{"While loop", parser.NodeWhile, true},
		{"With statement", parser.NodeWith, true},
		{"Try statement", parser.NodeTry, true},
		{"Except handler", parser.NodeExceptHandler, true},
		{"Lambda", parser.NodeLambda, true},
		{"List comprehension", parser.NodeListComp, true},
		{"Dict comprehension", parser.NodeDictComp, true},
		{"Nested function", parser.NodeFunctionDef, true},
		{"Nested class", parser.NodeClassDef, true},
		{"Return statement", parser.NodeReturn, false},
		{"Pass statement", parser.NodePass, false},
		{"Assign statement", parser.NodeAssign, false},
		{"Else clause", parser.NodeElseClause, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &parser.Node{Type: tt.nodeType}
			result := isNestingNode(node)
			if result != tt.expected {
				t.Errorf("isNestingNode(%s) = %v, expected %v", tt.nodeType, result, tt.expected)
			}
		})
	}
}
