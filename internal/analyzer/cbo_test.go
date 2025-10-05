package analyzer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCBOAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		options  *CBOOptions
		expected *CBOOptions
	}{
		{
			name:    "nil options should use defaults",
			options: nil,
			expected: &CBOOptions{
				IncludeBuiltins:   false,
				IncludeImports:    true,
				PublicClassesOnly: false,
				LowThreshold:      3, // Updated to industry standard
				MediumThreshold:   7, // Updated to industry standard
			},
		},
		{
			name: "custom options should be preserved",
			options: &CBOOptions{
				IncludeBuiltins: true,
				IncludeImports:  false,
				LowThreshold:    3,
				MediumThreshold: 8,
			},
			expected: &CBOOptions{
				IncludeBuiltins: true,
				IncludeImports:  false,
				LowThreshold:    3,
				MediumThreshold: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewCBOAnalyzer(tt.options)
			assert.NotNil(t, analyzer)
			assert.Equal(t, tt.expected.IncludeBuiltins, analyzer.options.IncludeBuiltins)
			assert.Equal(t, tt.expected.IncludeImports, analyzer.options.IncludeImports)
			assert.Equal(t, tt.expected.LowThreshold, analyzer.options.LowThreshold)
			assert.Equal(t, tt.expected.MediumThreshold, analyzer.options.MediumThreshold)
		})
	}
}

func TestCBOAnalyzer_AnalyzeClasses(t *testing.T) {
	tests := []struct {
		name          string
		pythonCode    string
		options       *CBOOptions
		expectedCount int
		expectedCBO   map[string]int    // className -> expected CBO count
		expectedRisk  map[string]string // className -> risk level
	}{
		{
			name: "simple class with no dependencies",
			pythonCode: `
class SimpleClass:
    def __init__(self):
        self.value = 42
    
    def get_value(self):
        return self.value
`,
			expectedCount: 1,
			expectedCBO:   map[string]int{"SimpleClass": 0},
			expectedRisk:  map[string]string{"SimpleClass": "low"},
		},
		{
			name: "class with inheritance",
			pythonCode: `
class BaseClass:
    pass

class DerivedClass(BaseClass):
    def __init__(self):
        super().__init__()
`,
			expectedCount: 2,
			expectedCBO:   map[string]int{"BaseClass": 0, "DerivedClass": 1},
			expectedRisk:  map[string]string{"BaseClass": "low", "DerivedClass": "low"},
		},
		{
			name: "class with multiple inheritance",
			pythonCode: `
class MixinA:
    pass

class MixinB:
    pass

class MultipleInheritance(MixinA, MixinB):
    pass
`,
			expectedCount: 3,
			expectedCBO:   map[string]int{"MixinA": 0, "MixinB": 0, "MultipleInheritance": 2},
			expectedRisk:  map[string]string{"MixinA": "low", "MixinB": "low", "MultipleInheritance": "low"},
		},
		{
			name: "class with type annotations",
			pythonCode: `
from typing import List, Dict

class User:
    pass

class UserManager:
    def __init__(self):
        self.users: List[User] = []
        self.metadata: Dict[str, str] = {}
    
    def add_user(self, user: User) -> None:
        self.users.append(user)
`,
			expectedCount: 2,
			expectedCBO:   map[string]int{"User": 0, "UserManager": 1}, // UserManager depends on User
			expectedRisk:  map[string]string{"User": "low", "UserManager": "low"},
		},
		{
			name: "class with object instantiation",
			pythonCode: `
class Logger:
    def log(self, message):
        print(message)

class Service:
    def __init__(self):
        self.logger = Logger()
    
    def do_work(self):
        self.logger.log("Working...")
`,
			expectedCount: 2,
			expectedCBO:   map[string]int{"Logger": 0, "Service": 1},
			expectedRisk:  map[string]string{"Logger": "low", "Service": "low"},
		},
		{
			name: "high coupling class",
			pythonCode: `
class A: pass
class B: pass  
class C: pass
class D: pass
class E: pass
class F: pass

class HighlyCoupled(A):
    def __init__(self):
        self.b = B()
        self.c = C()
        self.d = D()
        self.e = E()
        self.f = F()
`,
			options: &CBOOptions{
				LowThreshold:    2,
				MediumThreshold: 5,
			},
			expectedCount: 7,
			expectedCBO:   map[string]int{"HighlyCoupled": 6}, // Inherits from A + instantiates B,C,D,E,F
			expectedRisk:  map[string]string{"HighlyCoupled": "high"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the Python code
			ast, err := parseCode(tt.pythonCode)
			require.NoError(t, err, "Failed to parse Python code")

			// Create analyzer with options
			options := tt.options
			if options == nil {
				options = DefaultCBOOptions()
			}
			analyzer := NewCBOAnalyzer(options)

			// Analyze classes
			results, err := analyzer.AnalyzeClasses(ast, "test.py")
			require.NoError(t, err, "CBO analysis failed")

			// Check number of classes found
			if tt.expectedCount > 0 {
				assert.Len(t, results, tt.expectedCount, "Unexpected number of classes found")
			}

			// Check CBO values for specific classes
			resultMap := make(map[string]*CBOResult)
			for _, result := range results {
				resultMap[result.ClassName] = result
			}

			for className, expectedCBO := range tt.expectedCBO {
				result, found := resultMap[className]
				require.True(t, found, "Class %s not found in results", className)
				assert.Equal(t, expectedCBO, result.CouplingCount, "Unexpected CBO for class %s", className)
			}

			// Check risk levels
			for className, expectedRisk := range tt.expectedRisk {
				result, found := resultMap[className]
				require.True(t, found, "Class %s not found in results", className)
				assert.Equal(t, expectedRisk, result.RiskLevel, "Unexpected risk level for class %s", className)
			}
		})
	}
}

func TestCBOAnalyzer_BuiltinTypes(t *testing.T) {
	pythonCode := `
class MyClass:
    def __init__(self):
        self.data: list = []
        self.count: int = 0
    
    def process(self, items: list) -> dict:
        result = dict()
        for item in items:
            result[str(item)] = len(item)
        return result
`

	tests := []struct {
		name            string
		includeBuiltins bool
		expectedCBO     int
	}{
		{
			name:            "exclude builtins",
			includeBuiltins: false,
			expectedCBO:     0,
		},
		{
			name:            "include builtins",
			includeBuiltins: true,
			expectedCBO:     3, // list, int, dict (str and len are functions, not types)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parseCode(pythonCode)
			require.NoError(t, err)

			options := DefaultCBOOptions()
			options.IncludeBuiltins = tt.includeBuiltins

			analyzer := NewCBOAnalyzer(options)
			results, err := analyzer.AnalyzeClasses(ast, "test.py")
			require.NoError(t, err)

			require.Len(t, results, 1)
			assert.Equal(t, tt.expectedCBO, results[0].CouplingCount)
		})
	}
}

func TestCBOAnalyzer_ExcludePatterns(t *testing.T) {
	pythonCode := `
class TestClass:
    pass

class _PrivateClass:
    pass

class MyTestHelper:
    pass

class NormalClass(_PrivateClass):
    def __init__(self):
        self.helper = MyTestHelper()
`

	tests := []struct {
		name              string
		excludePatterns   []string
		publicClassesOnly bool
		expectedClasses   []string
	}{
		{
			name:            "no exclusions",
			excludePatterns: []string{},
			expectedClasses: []string{"TestClass", "_PrivateClass", "MyTestHelper", "NormalClass"},
		},
		{
			name:            "exclude test classes",
			excludePatterns: []string{"Test*", "*Test*"},
			expectedClasses: []string{"_PrivateClass", "NormalClass"},
		},
		{
			name:              "public classes only",
			publicClassesOnly: true,
			expectedClasses:   []string{"TestClass", "MyTestHelper", "NormalClass"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ast, err := parseCode(pythonCode)
			require.NoError(t, err)

			options := DefaultCBOOptions()
			options.ExcludePatterns = tt.excludePatterns
			options.PublicClassesOnly = tt.publicClassesOnly

			analyzer := NewCBOAnalyzer(options)
			results, err := analyzer.AnalyzeClasses(ast, "test.py")
			require.NoError(t, err)

			var actualClasses []string
			for _, result := range results {
				actualClasses = append(actualClasses, result.ClassName)
			}

			assert.ElementsMatch(t, tt.expectedClasses, actualClasses)
		})
	}
}

func TestCBOAnalyzer_RiskAssessment(t *testing.T) {
	tests := []struct {
		name            string
		cbo             int
		lowThreshold    int
		mediumThreshold int
		expectedRisk    string
	}{
		{"low risk", 3, 5, 10, "low"},
		{"low risk boundary", 5, 5, 10, "low"},
		{"medium risk", 7, 5, 10, "medium"},
		{"medium risk boundary", 10, 5, 10, "medium"},
		{"high risk", 15, 5, 10, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := &CBOOptions{
				LowThreshold:    tt.lowThreshold,
				MediumThreshold: tt.mediumThreshold,
			}
			analyzer := NewCBOAnalyzer(options)

			risk := analyzer.assessRiskLevel(tt.cbo)
			assert.Equal(t, tt.expectedRisk, risk)
		})
	}
}

func TestCBOAnalyzer_matchesPattern(t *testing.T) {
	analyzer := NewCBOAnalyzer(nil)

	tests := []struct {
		str     string
		pattern string
		matches bool
	}{
		{"TestClass", "Test*", true},
		{"MyTest", "*Test", true},
		{"TestHelper", "*Test*", true},
		{"NormalClass", "Test*", false},
		{"Helper", "*Test*", false},
		{"exact", "exact", true},
		{"different", "exact", false},
		{"anything", "*", true},
	}

	for _, tt := range tests {
		t.Run(tt.str+"_vs_"+tt.pattern, func(t *testing.T) {
			result := analyzer.matchesPattern(tt.str, tt.pattern)
			assert.Equal(t, tt.matches, result)
		})
	}
}

func TestCalculateCBO(t *testing.T) {
	pythonCode := `
class SimpleClass:
    pass

class DerivedClass(SimpleClass):
    def __init__(self):
        super().__init__()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := CalculateCBO(ast, "test.py")
	require.NoError(t, err)

	assert.Len(t, results, 2)

	// Find results by class name
	var simpleResult, derivedResult *CBOResult
	for _, result := range results {
		switch result.ClassName {
		case "SimpleClass":
			simpleResult = result
		case "DerivedClass":
			derivedResult = result
		}
	}

	require.NotNil(t, simpleResult)
	require.NotNil(t, derivedResult)

	assert.Equal(t, 0, simpleResult.CouplingCount)
	assert.Equal(t, 1, derivedResult.CouplingCount) // Depends on SimpleClass
	assert.Contains(t, derivedResult.DependentClasses, "SimpleClass")
}

// Helper function to parse Python code into AST
func parseCode(code string) (*parser.Node, error) {
	p := parser.New()
	ctx := context.Background()

	// Remove leading/trailing whitespace and ensure proper indentation
	code = strings.TrimSpace(code)

	result, err := p.Parse(ctx, []byte(code))
	if err != nil {
		return nil, err
	}

	return result.AST, nil
}

// Benchmark tests
func BenchmarkCBOAnalysis(b *testing.B) {
	// Create a moderately complex class structure
	pythonCode := `
from typing import List, Dict, Optional

class BaseClass:
    def base_method(self):
        pass

class MixinA:
    def mixin_a_method(self):
        pass

class MixinB:
    def mixin_b_method(self):
        pass

class ComplexClass(BaseClass, MixinA, MixinB):
    def __init__(self):
        self.data: List[str] = []
        self.lookup: Dict[str, int] = {}
        self.cache: Optional[Dict] = None
        self.helper = HelperClass()
        self.processor = DataProcessor()
    
    def process(self, items: List[str]) -> Dict[str, int]:
        processor = DataProcessor()
        return processor.process(items)

class HelperClass:
    def help(self):
        return "helping"

class DataProcessor:
    def process(self, data: List) -> Dict:
        return {}
`

	ast, err := parseCode(pythonCode)
	if err != nil {
		b.Fatalf("Failed to parse code: %v", err)
	}

	analyzer := NewCBOAnalyzer(DefaultCBOOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeClasses(ast, "benchmark.py")
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}

func BenchmarkLargeClassHierarchy(b *testing.B) {
	// Generate a larger class hierarchy for performance testing
	var codeBuilder strings.Builder

	// Create base classes
	for i := 0; i < 20; i++ {
		codeBuilder.WriteString(fmt.Sprintf("class Base%d:\n    pass\n\n", i))
	}

	// Create derived classes with multiple dependencies
	for i := 0; i < 50; i++ {
		codeBuilder.WriteString(fmt.Sprintf(`class Derived%d(Base%d):
    def __init__(self):
        self.helper1 = Base%d()
        self.helper2 = Base%d()
        self.helper3 = Base%d()

`, i, i%20, (i+1)%20, (i+2)%20, (i+3)%20))
	}

	ast, err := parseCode(codeBuilder.String())
	if err != nil {
		b.Fatalf("Failed to parse large code: %v", err)
	}

	analyzer := NewCBOAnalyzer(DefaultCBOOptions())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeClasses(ast, "large.py")
		if err != nil {
			b.Fatalf("Analysis failed: %v", err)
		}
	}
}
