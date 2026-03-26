package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestCalculateCognitiveComplexity_NilNode(t *testing.T) {
	result := CalculateCognitiveComplexity(nil)

	if result.Total != 0 {
		t.Errorf("Expected total 0 for nil node, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_SimpleFunction(t *testing.T) {
	// def simple():
	//     return 1
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "simple",
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

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 0 {
		t.Errorf("Expected total 0 for simple function, got %d", result.Total)
	}
	if result.FunctionName != "simple" {
		t.Errorf("Expected function name 'simple', got '%s'", result.FunctionName)
	}
}

func TestCalculateCognitiveComplexity_SingleIf(t *testing.T) {
	// def func_with_if(x):
	//     if x > 0:       # +1 (base, nesting=0)
	//         return x
	//     return 0
	// Cognitive complexity: 1
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
				Test: &parser.Node{
					Type: parser.NodeCompare,
					Location: parser.Location{
						StartLine: 2,
					},
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
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 4,
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 1 {
		t.Errorf("Expected cognitive complexity 1 for single if, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_NestedIf(t *testing.T) {
	// def nested(x, y):
	//     if x > 0:       # +1 (base, nesting=0)
	//         if y > 0:   # +2 (base + nesting=1)
	//             return x + y
	//     return 0
	// Cognitive complexity: 3
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "nested",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{
					Type: parser.NodeCompare,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeCompare,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeReturn,
								Location: parser.Location{
									StartLine: 4,
								},
							},
						},
					},
				},
			},
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 5,
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for nested if, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_IfElifElse(t *testing.T) {
	// def func(x):
	//     if x > 0:      # +1 (base, nesting=0)
	//         return 1
	//     elif x < 0:    # +1 (base, no nesting increment)
	//         return -1
	//     else:          # +1 (base)
	//         return 0
	// Cognitive complexity: 3
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   8,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{
					Type: parser.NodeCompare,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeReturn,
					},
				},
				Orelse: []*parser.Node{
					{
						Type: parser.NodeElifClause,
						Location: parser.Location{
							StartLine: 4,
						},
						Test: &parser.Node{
							Type: parser.NodeCompare,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeReturn,
							},
						},
						Orelse: []*parser.Node{
							{
								Type: parser.NodeElseClause,
								Location: parser.Location{
									StartLine: 6,
								},
								Body: []*parser.Node{
									{
										Type: parser.NodeReturn,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for if-elif-else, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_ForLoop(t *testing.T) {
	// def func(items):
	//     for item in items:  # +1 (base, nesting=0)
	//         if item > 0:    # +2 (base + nesting=1)
	//             print(item)
	// Cognitive complexity: 3
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   5,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeFor,
				Location: parser.Location{
					StartLine: 2,
				},
				Iter: &parser.Node{
					Type: parser.NodeName,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeCompare,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeExpr,
							},
						},
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for for-if, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_WhileWithBreak(t *testing.T) {
	// def func():
	//     while True:     # +1 (base, nesting=0)
	//         if done:    # +2 (base + nesting=1)
	//             break   # +1 (base for break)
	// Cognitive complexity: 4
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   5,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeWhile,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{
					Type: parser.NodeConstant,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeName,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeBreak,
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

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 4 {
		t.Errorf("Expected cognitive complexity 4 for while-if-break, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_TryExcept(t *testing.T) {
	// def func():
	//     try:
	//         do_something()
	//     except ValueError:  # +1 (base, nesting=0)
	//         handle()
	//     except:             # +1 (base, nesting=0)
	//         pass
	// Cognitive complexity: 2
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
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
						Type: parser.NodeExpr,
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
								Type: parser.NodeExpr,
							},
						},
					},
					{
						Type: parser.NodeExceptHandler,
						Location: parser.Location{
							StartLine: 6,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
							},
						},
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 2 {
		t.Errorf("Expected cognitive complexity 2 for try-except-except, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_BooleanOperator(t *testing.T) {
	// def func(a, b, c):
	//     if a and b and c:  # +1 (if, nesting=0) + +1 (bool op sequence "and")
	//         pass
	// Cognitive complexity: 2
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   4,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{
					Type: parser.NodeBoolOp,
					Op:   "and",
					Children: []*parser.Node{
						{Type: parser.NodeName, Name: "a"},
						{Type: parser.NodeName, Name: "b"},
					},
					Left:  &parser.Node{Type: parser.NodeName, Name: "a"},
					Right: &parser.Node{Type: parser.NodeName, Name: "c"},
				},
				Body: []*parser.Node{
					{
						Type: parser.NodePass,
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 2 {
		t.Errorf("Expected cognitive complexity 2 for if with 'and', got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_MixedBooleanOperators(t *testing.T) {
	// def func(a, b, c):
	//     if a and b or c:  # +1 (if) + +1 (and) + +1 (or, different from and)
	//         pass
	// Cognitive complexity: 3
	//
	// Tree structure: BoolOp(or, left=BoolOp(and, a, b), right=c)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   4,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{
					Type: parser.NodeBoolOp,
					Op:   "or",
					Left: &parser.Node{
						Type: parser.NodeBoolOp,
						Op:   "and",
						Left: &parser.Node{Type: parser.NodeName, Name: "a"},
						Right: &parser.Node{Type: parser.NodeName, Name: "b"},
					},
					Right: &parser.Node{Type: parser.NodeName, Name: "c"},
				},
				Body: []*parser.Node{
					{
						Type: parser.NodePass,
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for mixed boolean operators, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_NestedFunction(t *testing.T) {
	// def outer():
	//     def inner():            # (nesting increase)
	//         if condition:       # +2 (base + nesting=1)
	//             return True
	//     return inner()
	// Cognitive complexity: 2
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "outer",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeFunctionDef,
				Name: "inner",
				Location: parser.Location{
					StartLine: 2,
					EndLine:   4,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeName,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeReturn,
							},
						},
					},
				},
			},
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 5,
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 2 {
		t.Errorf("Expected cognitive complexity 2 for nested function with if, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_WithStatement(t *testing.T) {
	// def func():
	//     with open("f") as f:   # (nesting increase, no base increment)
	//         if f.read():       # +2 (base + nesting=1)
	//             pass
	// Cognitive complexity: 2
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   5,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeWith,
				Location: parser.Location{
					StartLine: 2,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeCall,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
							},
						},
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 2 {
		t.Errorf("Expected cognitive complexity 2 for with-if, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_Continue(t *testing.T) {
	// def func(items):
	//     for item in items:   # +1 (base, nesting=0)
	//         if not item:     # +2 (base + nesting=1)
	//             continue     # +1 (base for continue)
	//         process(item)
	// Cognitive complexity: 4
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeFor,
				Location: parser.Location{
					StartLine: 2,
				},
				Iter: &parser.Node{
					Type: parser.NodeName,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeIf,
						Location: parser.Location{
							StartLine: 3,
						},
						Test: &parser.Node{
							Type: parser.NodeUnaryOp,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeContinue,
								Location: parser.Location{
									StartLine: 4,
								},
							},
						},
					},
					{
						Type: parser.NodeExpr,
					},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 4 {
		t.Errorf("Expected cognitive complexity 4 for for-if-continue, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_DeeplyNested(t *testing.T) {
	// def func(x):
	//     if x:               # +1 (nesting=0)
	//         for i in x:     # +2 (nesting=1)
	//             while True: # +3 (nesting=2)
	//                 break   # +1
	// Cognitive complexity: 7
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeIf,
				Location: parser.Location{
					StartLine: 2,
				},
				Test: &parser.Node{Type: parser.NodeName},
				Body: []*parser.Node{
					{
						Type: parser.NodeFor,
						Location: parser.Location{
							StartLine: 3,
						},
						Iter: &parser.Node{Type: parser.NodeName},
						Body: []*parser.Node{
							{
								Type: parser.NodeWhile,
								Location: parser.Location{
									StartLine: 4,
								},
								Test: &parser.Node{Type: parser.NodeConstant},
								Body: []*parser.Node{
									{
										Type: parser.NodeBreak,
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

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 7 {
		t.Errorf("Expected cognitive complexity 7 for deeply nested, got %d", result.Total)
	}
}
