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

func TestCalculateCognitiveComplexity_TraversesStatementValue(t *testing.T) {
	tests := []struct {
		name string
		stmt *parser.Node
		want int
	}{
		{
			name: "return ternary",
			stmt: &parser.Node{
				Type: parser.NodeReturn,
				Value: &parser.Node{
					Type: parser.NodeIfExp,
					Test: &parser.Node{Type: parser.NodeName, Name: "flag"},
					Body: []*parser.Node{{Type: parser.NodeName, Name: "left"}},
					Orelse: []*parser.Node{
						{Type: parser.NodeName, Name: "right"},
					},
				},
			},
			want: 1,
		},
		{
			name: "assignment boolean expression",
			stmt: &parser.Node{
				Type: parser.NodeAssign,
				Value: &parser.Node{
					Type: parser.NodeBoolOp,
					Op:   "and",
					Children: []*parser.Node{
						{Type: parser.NodeName, Name: "left"},
						{Type: parser.NodeName, Name: "right"},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			funcNode := &parser.Node{
				Type: parser.NodeFunctionDef,
				Name: "value_expr",
				Body: []*parser.Node{tt.stmt},
			}

			result := CalculateCognitiveComplexity(funcNode)
			if result.Total != tt.want {
				t.Fatalf("Expected cognitive complexity %d, got %d", tt.want, result.Total)
			}
		})
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
						{Type: parser.NodeName, Name: "c"},
					},
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
					Children: []*parser.Node{
						{
							Type: parser.NodeBoolOp,
							Op:   "and",
							Children: []*parser.Node{
								{Type: parser.NodeName, Name: "a"},
								{Type: parser.NodeName, Name: "b"},
							},
						},
						{Type: parser.NodeName, Name: "c"},
					},
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
	//     def inner():            # scope boundary, not traversed
	//         if condition:       # NOT counted against outer
	//             return True
	//     return inner()
	// Cognitive complexity: 0 (outer's own body has no control flow)
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

	if result.Total != 0 {
		t.Errorf("Expected cognitive complexity 0 for outer function (nested scopes are boundaries), got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_NestedScopeBoundary(t *testing.T) {
	// def outer():
	//     for i in items:            # +1 (loop)
	//         if i > 0:              # +2 (base + nesting=1 from loop)
	//             pass
	//     def inner():               # scope boundary
	//         for j in range(10):    # NOT counted against outer
	//             pass
	//     return
	// Cognitive complexity: 3 (only outer's own control flow)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "outer",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   8,
		},
		Body: []*parser.Node{
			{
				Type:     parser.NodeFor,
				Location: parser.Location{StartLine: 2},
				Iter: &parser.Node{
					Type: parser.NodeName,
				},
				Body: []*parser.Node{
					{
						Type:     parser.NodeIf,
						Location: parser.Location{StartLine: 3},
						Test: &parser.Node{
							Type: parser.NodeCompare,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodePass,
							},
						},
					},
				},
			},
			{
				Type: parser.NodeFunctionDef,
				Name: "inner",
				Location: parser.Location{
					StartLine: 5,
					EndLine:   7,
				},
				Body: []*parser.Node{
					{
						Type:     parser.NodeFor,
						Location: parser.Location{StartLine: 6},
						Iter: &parser.Node{
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
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 8,
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for outer function (nested scopes not traversed), got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_NestedClassScopeBoundary(t *testing.T) {
	// def outer():
	//     for i in items:          # +1 (loop)
	//         pass
	//     class InnerClass:        # scope boundary
	//         def method():        # NOT counted against outer
	//             if True:        # NOT counted against outer
	//                 pass
	//     return
	// Cognitive complexity: 1 (only outer's own loop)
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "outer",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   8,
		},
		Body: []*parser.Node{
			{
				Type:     parser.NodeFor,
				Location: parser.Location{StartLine: 2},
				Iter: &parser.Node{
					Type: parser.NodeName,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodePass,
					},
				},
			},
			{
				Type: parser.NodeClassDef,
				Name: "InnerClass",
				Location: parser.Location{
					StartLine: 4,
					EndLine:   7,
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeFunctionDef,
						Name: "method",
						Body: []*parser.Node{
							{
								Type: parser.NodeIf,
								Test: &parser.Node{
									Type: parser.NodeConstant,
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
			},
			{
				Type: parser.NodeReturn,
				Location: parser.Location{
					StartLine: 8,
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 1 {
		t.Errorf("Expected cognitive complexity 1 for outer function (class scope not traversed), got %d", result.Total)
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

func TestCalculateCognitiveComplexity_Lambda(t *testing.T) {
	// def func():
	//     f = lambda x: x + 1   # lambda: nesting increase only, no base increment
	//     return f
	// Cognitive complexity: 0
	funcNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "func",
		Location: parser.Location{
			StartLine: 1,
			EndLine:   4,
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeAssign,
				Children: []*parser.Node{
					{
						Type: parser.NodeLambda,
						Location: parser.Location{
							StartLine: 2,
						},
						Body: []*parser.Node{
							{
								Type: parser.NodeBinOp,
							},
						},
					},
				},
			},
			{
				Type: parser.NodeReturn,
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 0 {
		t.Errorf("Expected cognitive complexity 0 for lambda (no base increment), got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_LambdaWithNesting(t *testing.T) {
	// def func():
	//     f = lambda x: x if x > 0 else -x  # lambda: nesting+1, IfExp: +1 base + nesting=1 = +2
	//     return f
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
				Type: parser.NodeAssign,
				Children: []*parser.Node{
					{
						Type: parser.NodeLambda,
						Location: parser.Location{
							StartLine: 2,
						},
						Children: []*parser.Node{
							{
								Type: parser.NodeIfExp,
								Location: parser.Location{
									StartLine: 2,
								},
								Test: &parser.Node{Type: parser.NodeCompare},
							},
						},
					},
				},
			},
			{
				Type: parser.NodeReturn,
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	if result.Total != 2 {
		t.Errorf("Expected cognitive complexity 2 for lambda with ternary, got %d", result.Total)
	}
}

func TestCalculateCognitiveComplexity_BoolOpWithNestedIfExp(t *testing.T) {
	// def func(a, cond, x, y):
	//     if a and (x if cond else y):  # if: +1(base) + 0(nesting)
	//                                   # "and": +1(bool op)
	//                                   # IfExp: +1(base) + 1(nesting from if)
	//                                   # Total inside if's test is at nesting=1
	//         pass
	// Cognitive complexity: 3
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
						{
							Type: parser.NodeIfExp,
							Location: parser.Location{
								StartLine: 2,
							},
							Test: &parser.Node{Type: parser.NodeName, Name: "cond"},
						},
					},
				},
				Body: []*parser.Node{
					{Type: parser.NodePass},
				},
			},
		},
	}

	result := CalculateCognitiveComplexity(funcNode)

	// if: +1, and: +1, IfExp at nesting=1 (inherited from if's test context): +1+1 = +2
	// But BoolOp's children are traversed with nestingLevel from the caller (which is
	// the if's test traversal at nestingLevel=0, not nestingLevel+1).
	// The if condition's Test is traversed at nestingLevel (0), so IfExp gets +1+0 = +1.
	// Total: 1(if) + 1(and) + 1(IfExp at nesting=0) = 3
	if result.Total != 3 {
		t.Errorf("Expected cognitive complexity 3 for BoolOp with nested IfExp, got %d", result.Total)
	}
}
