package parser

import (
	"context"
	"strings"
	"testing"

	sitter "github.com/smacker/go-tree-sitter"
)

func TestNew(t *testing.T) {
	parser := New()
	if parser == nil {
		t.Fatal("New() returned nil")
	}
	if parser.parser == nil {
		t.Fatal("parser field is nil")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantErr bool
	}{
		{
			name: "simple function",
			source: `def hello():
    print("Hello, World!")`,
			wantErr: false,
		},
		{
			name: "class definition",
			source: `class MyClass:
    def __init__(self):
        self.value = 42`,
			wantErr: false,
		},
		{
			name: "complex code",
			source: `import sys

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

class Calculator:
    def add(self, a, b):
        return a + b
    
    def subtract(self, a, b):
        return a - b

if __name__ == "__main__":
    calc = Calculator()
    print(calc.add(10, 5))`,
			wantErr: false,
		},
		{
			name:    "empty source",
			source:  "",
			wantErr: false,
		},
		{
			name: "syntax error",
			source: `def broken(:
    pass`,
			wantErr: true,
		},
		{
			name: "incomplete code",
			source: `def incomplete(
`,
			wantErr: true,
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
			if result.Tree == nil {
				t.Fatal("ParseResult.Tree is nil")
			}
			if result.RootNode == nil {
				t.Fatal("ParseResult.RootNode is nil")
			}
			if string(result.SourceCode) != tt.source {
				t.Errorf("ParseResult.SourceCode mismatch: got %q, want %q",
					string(result.SourceCode), tt.source)
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid Python code",
			content: "print('Hello')",
			wantErr: false,
		},
		{
			name:    "invalid syntax",
			content: "print('Hello'",
			wantErr: true,
		},
	}

	parser := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			result, err := parser.ParseFile(ctx, reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFile() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFile() unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("ParseFile() returned nil result")
			}
		})
	}
}

func TestGetNodeText(t *testing.T) {
	parser := New()
	ctx := context.Background()
	source := []byte("def hello(): pass")

	result, err := parser.Parse(ctx, source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	text := parser.GetNodeText(result.RootNode, source)
	if text != string(source) {
		t.Errorf("GetNodeText() = %q, want %q", text, string(source))
	}
}

func TestWalkTree(t *testing.T) {
	parser := New()
	ctx := context.Background()
	source := []byte(`def func1():
    pass

def func2():
    pass`)

	result, err := parser.Parse(ctx, source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	nodeCount := 0
	err = parser.WalkTree(result.RootNode, func(node *sitter.Node) error {
		nodeCount++
		return nil
	})

	if err != nil {
		t.Errorf("WalkTree() error: %v", err)
	}

	if nodeCount == 0 {
		t.Error("WalkTree() visited 0 nodes")
	}
}

func TestFindNodes(t *testing.T) {
	parser := New()
	ctx := context.Background()
	source := []byte(`def func1():
    pass

def func2():
    pass

class MyClass:
    pass`)

	result, err := parser.Parse(ctx, source)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	tests := []struct {
		nodeType string
		minCount int
	}{
		{"function_definition", 2},
		{"class_definition", 1},
		{"pass_statement", 3},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			nodes := parser.FindNodes(result.RootNode, tt.nodeType)
			if len(nodes) < tt.minCount {
				t.Errorf("FindNodes(%q) found %d nodes, want at least %d",
					tt.nodeType, len(nodes), tt.minCount)
			}
		})
	}
}

func TestHasSyntaxErrors(t *testing.T) {
	parser := New()
	ctx := context.Background()

	tests := []struct {
		name      string
		source    string
		hasErrors bool
	}{
		{
			name:      "valid code",
			source:    "def hello(): pass",
			hasErrors: false,
		},
		{
			name:      "syntax error",
			source:    "def broken(: pass",
			hasErrors: true,
		},
		{
			name:      "incomplete code",
			source:    "if True",
			hasErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, _ := parser.parser.ParseCtx(ctx, nil, []byte(tt.source))
			rootNode := tree.RootNode()

			hasErrors := parser.HasSyntaxErrors(rootNode)
			if hasErrors != tt.hasErrors {
				t.Errorf("HasSyntaxErrors() = %v, want %v", hasErrors, tt.hasErrors)
			}
		})
	}
}

func BenchmarkParse(b *testing.B) {
	parser := New()
	ctx := context.Background()
	source := []byte(`import sys

def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

class Calculator:
    def add(self, a, b):
        return a + b
    
    def subtract(self, a, b):
        return a - b

if __name__ == "__main__":
    calc = Calculator()
    print(calc.add(10, 5))`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(ctx, source)
	}
}

func BenchmarkWalkTree(b *testing.B) {
	parser := New()
	ctx := context.Background()
	source := []byte(`def func1():
    x = 1
    y = 2
    return x + y

def func2():
    for i in range(10):
        print(i)`)

	result, _ := parser.Parse(ctx, source)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parser.WalkTree(result.RootNode, func(node *sitter.Node) error {
			return nil
		})
	}
}
