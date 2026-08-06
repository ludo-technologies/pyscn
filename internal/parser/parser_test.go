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
		{
			name: "try finally statement",
			source: `
try:
    print("try")
finally:
    print("finally")
`,
			wantErr: false,
		},
		{
			name: "try except finally",
			source: `
try:
    risky_operation()
except ValueError:
    handle_error()
finally:
    cleanup()
`,
			wantErr: false,
		},
		{
			name: "try except else finally",
			source: `
try:
    operation()
except Exception:
    handle()
else:
    success()
finally:
    cleanup()
`,
			wantErr: false,
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

func TestCheckSyntax(t *testing.T) {
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

			err := parser.CheckSyntax(rootNode, []byte(tt.source))
			if (err != nil) != tt.hasErrors {
				t.Errorf("CheckSyntax() error = %v, want error = %v", err, tt.hasErrors)
			}
		})
	}
}

// The tree-sitter Python grammar accepts these legacy constructs without
// producing an ERROR node, so they used to be analyzed as if they were valid.
// Every case here is a SyntaxError on both Python 3.13 and 3.14.
func TestParseRejectsSyntaxInvalidInEveryPython3(t *testing.T) {
	parser := New()
	ctx := context.Background()

	tests := []struct {
		name   string
		source string
	}{
		{"print statement", "print 'hello'\n"},
		{"print statement with a name", "x = 1\nprint x\n"},
		{"print statement with a tuple", "print 1, 2\n"},
		{"exec statement", "exec 'code'\n"},
		{"exec statement with in", "exec 'code' in {}\n"},
		{
			name:   "exception list bound with as",
			source: "try:\n    pass\nexcept OSError, TypeError as e:\n    pass\n",
		},
		{"raise with argument list", "raise ValueError, 'msg'\n"},
		{"raise with traceback", "raise ValueError, 'msg', tb\n"},
		{"backtick repr", "x = `y`\n"},
		{"bare octal literal", "mode = 0777\n"},
		{"leading zero decimal", "n = 08\n"},
		{"long literal", "n = 10L\n"},
		{"<> operator", "if a <> b:\n    pass\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(ctx, []byte(tt.source))
			if err == nil {
				t.Fatalf("Parse() succeeded for invalid source, want error (got %d AST children)", len(result.AST.Body))
			}
			if !strings.Contains(err.Error(), "not valid Python 3") {
				t.Errorf("Parse() error = %q, want it to mention it is not valid Python 3", err)
			}
		})
	}
}

// Every case here compiles on Python 3.13 and/or 3.14, so pyscn must analyze it
// rather than reject it. The unparenthesized `except A, B:` list became valid in
// Python 3.14 (PEP 758), and `print >>f, x` parses as a shift plus a tuple.
func TestParseAcceptsValidPython3(t *testing.T) {
	parser := New()
	ctx := context.Background()

	tests := []struct {
		name   string
		source string
	}{
		{"print function", "print('hello')\n"},
		{"print function with keywords", "import sys\nprint('a', file=sys.stderr, sep='')\n"},
		{"print as an identifier", "print = 5\nvalue = print\n"},
		{"print shifted into a stream", "import sys\nprint >>sys.stderr, 'oops'\n"},
		{"print subscripted", "print [1]\n"},
		{"print negated", "print -1\n"},
		{"exec function", "exec('code')\n"},
		{"exec subscripted", "exec [1]\n"},
		{
			name:   "unparenthesized exception list is valid since PEP 758",
			source: "def f(x):\n    try:\n        return x\n    except OSError, TypeError:\n        return None\n",
		},
		{"parenthesized exception list", "try:\n    pass\nexcept (OSError, TypeError):\n    pass\n"},
		{"parenthesized exception list with as", "try:\n    pass\nexcept (OSError, TypeError) as e:\n    pass\n"},
		{"except with as", "try:\n    pass\nexcept OSError as e:\n    pass\n"},
		{"except group", "try:\n    pass\nexcept* OSError:\n    pass\n"},
		{"raise from", "raise ValueError('x') from err\n"},
		{"raise with call arguments", "raise SomeExc(1, 2)\n"},
		{"bare raise", "try:\n    pass\nexcept OSError:\n    raise\n"},
		{"zero literals", "x = 0\ny = 00\nz = 0_0\n"},
		{"prefixed literals", "o = 0o777\nb = 0b1010\nh = 0xFF\n"},
		{"complex and float literals", "c = 10j\nz = 0j\nf = 1e10\n"},
		{"not equal operator", "if a != b:\n    pass\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parser.Parse(ctx, []byte(tt.source)); err != nil {
				t.Errorf("Parse() error = %v, want valid Python 3 to parse", err)
			}
		})
	}
}

func TestParseRejectsIncompleteCode(t *testing.T) {
	parser := New()

	_, err := parser.Parse(context.Background(), []byte("def broken(: pass"))
	if err == nil {
		t.Fatal("expected Parse to reject incomplete code")
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
