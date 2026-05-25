package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLCOMAnalyzer(t *testing.T) {
	tests := []struct {
		name     string
		options  *LCOMOptions
		expected *LCOMOptions
	}{
		{
			name:    "nil options should use defaults",
			options: nil,
			expected: &LCOMOptions{
				LowThreshold:    2,
				MediumThreshold: 5,
			},
		},
		{
			name: "custom options should be preserved",
			options: &LCOMOptions{
				LowThreshold:    3,
				MediumThreshold: 8,
			},
			expected: &LCOMOptions{
				LowThreshold:    3,
				MediumThreshold: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewLCOMAnalyzer(tt.options)
			assert.NotNil(t, analyzer)
			assert.Equal(t, tt.expected.LowThreshold, analyzer.options.LowThreshold)
			assert.Equal(t, tt.expected.MediumThreshold, analyzer.options.MediumThreshold)
		})
	}
}

func TestLCOMAnalyzer_AnalyzeClasses(t *testing.T) {
	tests := []struct {
		name             string
		pythonCode       string
		expectedCount    int
		expectedLCOM     map[string]int    // className -> expected LCOM4
		expectedRisk     map[string]string // className -> risk level
		expectedExcluded map[string]int    // className -> excluded methods
	}{
		{
			name: "fully cohesive class sharing one variable",
			pythonCode: `
class CohesiveClass:
    def __init__(self):
        self.value = 0

    def get_value(self):
        return self.value

    def set_value(self, v):
        self.value = v
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"CohesiveClass": 1},
			expectedRisk:  map[string]string{"CohesiveClass": "low"},
		},
		{
			name: "two disconnected method groups",
			pythonCode: `
class TwoGroupClass:
    def get_a(self):
        return self.a

    def set_a(self, v):
        self.a = v

    def get_b(self):
        return self.b

    def set_b(self, v):
        self.b = v
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"TwoGroupClass": 2},
			expectedRisk:  map[string]string{"TwoGroupClass": "low"},
		},
		{
			name: "three disconnected method groups",
			pythonCode: `
class ThreeGroupClass:
    def method_x(self):
        return self.x

    def method_y(self):
        return self.y

    def method_z(self):
        return self.z
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"ThreeGroupClass": 3},
			expectedRisk:  map[string]string{"ThreeGroupClass": "medium"},
		},
		{
			name: "class with staticmethod and classmethod excluded",
			pythonCode: `
class ClassWithDecorators:
    def __init__(self):
        self.data = []

    def add(self, item):
        self.data.append(item)

    @staticmethod
    def helper(x):
        return x * 2

    @classmethod
    def create(cls):
        return cls()
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"ClassWithDecorators": 1},
			expectedRisk:     map[string]string{"ClassWithDecorators": "low"},
			expectedExcluded: map[string]int{"ClassWithDecorators": 2},
		},
		{
			name: "class with property included",
			pythonCode: `
class ClassWithProperty:
    def __init__(self):
        self._value = 0

    @property
    def value(self):
        return self._value

    def set_value(self, v):
        self._value = v
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"ClassWithProperty": 1},
			expectedRisk:  map[string]string{"ClassWithProperty": "low"},
		},
		{
			name: "single method class is trivially cohesive",
			pythonCode: `
class SingleMethodClass:
    def do_something(self):
        self.x = 1
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"SingleMethodClass": 1},
			expectedRisk:  map[string]string{"SingleMethodClass": "low"},
		},
		{
			name: "empty class is trivially cohesive",
			pythonCode: `
class EmptyClass:
    pass
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"EmptyClass": 1},
			expectedRisk:  map[string]string{"EmptyClass": "low"},
		},
		{
			name: "methods without self access form separate components",
			pythonCode: `
class NoSelfAccessClass:
    def method_a(self):
        return 42

    def method_b(self):
        return 99
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"NoSelfAccessClass": 2},
			expectedRisk:  map[string]string{"NoSelfAccessClass": "low"},
		},
		{
			name: "magic methods sharing self.value are cohesive",
			pythonCode: `
class MagicMethodClass:
    def __init__(self, value):
        self.value = value

    def __str__(self):
        return str(self.value)

    def __repr__(self):
        return "MagicMethodClass(" + str(self.value) + ")"
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"MagicMethodClass": 1},
			expectedRisk:  map[string]string{"MagicMethodClass": "low"},
		},
		{
			name: "multiple classes in one file",
			pythonCode: `
class ClassA:
    def method(self):
        return self.x

class ClassB:
    def method1(self):
        return self.a

    def method2(self):
        return self.b
`,
			expectedCount: 2,
			expectedLCOM: map[string]int{
				"ClassA": 1,
				"ClassB": 2,
			},
			expectedRisk: map[string]string{
				"ClassA": "low",
				"ClassB": "low",
			},
		},
		{
			name: "high risk class with many disconnected groups",
			pythonCode: `
class HighLCOMClass:
    def method1(self):
        return self.a
    def method2(self):
        return self.b
    def method3(self):
        return self.c
    def method4(self):
        return self.d
    def method5(self):
        return self.e
    def method6(self):
        return self.f
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"HighLCOMClass": 6},
			expectedRisk:  map[string]string{"HighLCOMClass": "high"},
		},
		{
			name: "methods connected by intra-class calls without shared fields",
			pythonCode: `
class CallConnectedClass:
    def action_a(self):
        self.helper()

    def action_b(self):
        self.helper()

    def helper(self):
        print("work")
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"CallConnectedClass": 1},
			expectedRisk:  map[string]string{"CallConnectedClass": "low"},
		},
		{
			name: "mixed field sharing and method calls",
			pythonCode: `
class MixedConnectionClass:
    def get_data(self):
        return self.data

    def set_data(self, v):
        self.data = v

    def process(self):
        self.validate()

    def validate(self):
        pass
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"MixedConnectionClass": 2},
			expectedRisk:  map[string]string{"MixedConnectionClass": "low"},
		},
		{
			name: "chain of method calls connects all methods",
			pythonCode: `
class ChainCallClass:
    def start(self):
        self.middle()

    def middle(self):
        self.finish()

    def finish(self):
        pass
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"ChainCallClass": 1},
			expectedRisk:  map[string]string{"ChainCallClass": "low"},
		},
	}

	p := parser.New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.Parse(context.Background(), []byte(tt.pythonCode))
			require.NoError(t, err)

			analyzer := NewLCOMAnalyzer(nil)
			results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, len(results), "unexpected number of classes")

			for _, r := range results {
				if expectedLCOM, ok := tt.expectedLCOM[r.ClassName]; ok {
					assert.Equal(t, expectedLCOM, r.LCOM4,
						fmt.Sprintf("class %s: unexpected LCOM4", r.ClassName))
				}
				if expectedRisk, ok := tt.expectedRisk[r.ClassName]; ok {
					assert.Equal(t, expectedRisk, r.RiskLevel,
						fmt.Sprintf("class %s: unexpected risk level", r.ClassName))
				}
				if expectedExcluded, ok := tt.expectedExcluded[r.ClassName]; ok {
					assert.Equal(t, expectedExcluded, r.ExcludedMethods,
						fmt.Sprintf("class %s: unexpected excluded methods", r.ClassName))
				}
			}
		})
	}
}

func TestLCOMAnalyzer_NilAST(t *testing.T) {
	analyzer := NewLCOMAnalyzer(nil)
	_, err := analyzer.AnalyzeClasses(nil, "test.py")
	assert.Error(t, err)
}

func TestLCOMAnalyzer_RiskLevels(t *testing.T) {
	analyzer := NewLCOMAnalyzer(&LCOMOptions{
		LowThreshold:    2,
		MediumThreshold: 5,
	})

	tests := []struct {
		lcom4    int
		expected string
	}{
		{1, "low"},
		{2, "low"},
		{3, "medium"},
		{5, "medium"},
		{6, "high"},
		{10, "high"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("LCOM4=%d", tt.lcom4), func(t *testing.T) {
			assert.Equal(t, tt.expected, analyzer.assessRiskLevel(tt.lcom4))
		})
	}
}

func TestCalculateLCOM(t *testing.T) {
	p := parser.New()
	code := `
class SimpleClass:
    def __init__(self):
        self.value = 0
    def get_value(self):
        return self.value
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	results, err := CalculateLCOM(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 1, results[0].LCOM4)
	assert.Equal(t, "SimpleClass", results[0].ClassName)
}

func TestCalculateLCOMWithConfig(t *testing.T) {
	p := parser.New()
	code := `
class TestClass:
    def method_a(self):
        return self.x
    def method_b(self):
        return self.y
    def method_c(self):
        return self.z
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	options := &LCOMOptions{
		LowThreshold:    1,
		MediumThreshold: 2,
	}
	results, err := CalculateLCOMWithConfig(result.AST, "test.py", options)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 3, results[0].LCOM4)
	assert.Equal(t, "high", results[0].RiskLevel) // 3 > MediumThreshold(2) → high
}

func TestLCOMAnalyzer_MethodGroups(t *testing.T) {
	p := parser.New()
	code := `
class TwoGroupClass:
    def get_a(self):
        return self.a
    def set_a(self, v):
        self.a = v
    def get_b(self):
        return self.b
    def set_b(self, v):
        self.b = v
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 2, r.LCOM4)
	assert.Equal(t, 2, len(r.MethodGroups))
	assert.Equal(t, 2, r.InstanceVariables)

	// Groups should be sorted deterministically
	assert.Contains(t, r.MethodGroups[0], "get_a")
	assert.Contains(t, r.MethodGroups[0], "set_a")
	assert.Contains(t, r.MethodGroups[1], "get_b")
	assert.Contains(t, r.MethodGroups[1], "set_b")
}

func TestLCOMAnalyzer_FStringInstanceVariableAccesses(t *testing.T) {
	p := parser.New()
	code := `
class A:
    def __init__(self):
        self._sep = '/'
    def render(self):
        return f"x{self._sep}y"

class B:
    def __init__(self):
        self._sep = '/'
    def render(self):
        return f"x{1:{self._sep}>5}y"

class C:
    def __init__(self):
        self._sep = '/'
    def render(self):
        return "x" + self._sep + "y"

class D:
    def __init__(self):
        self._x = 1
    def m1(self):
        return self._x
    def m2(self):
        return f"{self._x}"

class E:
    def __init__(self):
        self._x = 1
    def render(self):
        return f"{f'{self._x}'}"
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 5)

	byName := make(map[string]*LCOMResult, len(results))
	for _, result := range results {
		byName[result.ClassName] = result
	}

	expectedGroups := map[string][][]string{
		"A": {{"__init__", "render"}},
		"B": {{"__init__", "render"}},
		"C": {{"__init__", "render"}},
		"D": {{"__init__", "m1", "m2"}},
		"E": {{"__init__", "render"}},
	}

	for className, groups := range expectedGroups {
		t.Run(className, func(t *testing.T) {
			result, ok := byName[className]
			require.True(t, ok, "missing class %s", className)
			assert.Equal(t, 1, result.LCOM4)
			assert.Equal(t, 1, result.InstanceVariables)
			assert.Equal(t, groups, result.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_WithContextInstanceVariableAccesses(t *testing.T) {
	p := parser.New()
	code := `
class B:
    def __init__(self):
        self._fname = 'x'
    def write(self, t):
        with open(self._fname, 'w') as f:
            f.write(t)
    def __del__(self):
        with open(self._fname, 'w') as f:
            f.write('done')

class K:
    def __init__(self):
        self._x = 1
    def m1(self):
        with open(self._x, 'r') as f:
            return f.read()
    def m2(self):
        return self._x

class AsyncContext:
    def __init__(self):
        self._resource = None
    async def acquire(self):
        async with self._resource as ctx:
            return ctx
    def current(self):
        return self._resource

class MultipleItems:
    def __init__(self):
        self._a = 'a'
        self._b = 'b'
    def copy(self):
        with open(self._a, 'r') as src, open(self._b, 'w') as dst:
            dst.write(src.read())
    def read_a(self):
        return self._a
    def read_b(self):
        return self._b
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 4)

	byName := make(map[string]*LCOMResult, len(results))
	for _, result := range results {
		byName[result.ClassName] = result
	}

	expectedGroups := map[string][][]string{
		"B":             {{"__del__", "__init__", "write"}},
		"K":             {{"__init__", "m1", "m2"}},
		"AsyncContext":  {{"__init__", "acquire", "current"}},
		"MultipleItems": {{"__init__", "copy", "read_a", "read_b"}},
	}

	for className, groups := range expectedGroups {
		t.Run(className, func(t *testing.T) {
			result, ok := byName[className]
			require.True(t, ok, "missing class %s", className)
			assert.Equal(t, 1, result.LCOM4)
			assert.Equal(t, groups, result.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_CtypesFieldsAssignedOutsideClass(t *testing.T) {
	p := parser.New()
	code := `
import ctypes

class Image(ctypes.Structure):
    def __init__(self, data=None, width=None, height=None, mipmaps=None, format_=None):
        super(Image, self).__init__(
            data,
            width or 0,
            height or 0,
            mipmaps or 1,
            format_ or 7,
        )

    @property
    def is_ready(self):
        return _IsImageReady(self)

    def resize(self, width, height):
        _ImageResize(self, width, height)

    def export(self, file_name):
        return _ExportImage(self, file_name)

Image._fields_ = [
    ("data", ctypes.c_void_p),
    ("width", ctypes.c_int),
    ("height", ctypes.c_int),
    ("mipmaps", ctypes.c_int),
    ("format", ctypes.c_int),
]
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "Image", r.ClassName)
	assert.Equal(t, 1, r.LCOM4)
	assert.Equal(t, 5, r.InstanceVariables)
	assert.Equal(t, [][]string{{"__init__", "export", "is_ready", "resize"}}, r.MethodGroups)
}

func TestLCOMAnalyzer_CtypesFieldsDoNotTreatSelfParameterAsFieldAccess(t *testing.T) {
	p := parser.New()
	code := `
import ctypes

class Packet(ctypes.Structure):
    def touches_state(self):
        _TouchPacket(self)

    def utility(self):
        return 42

Packet._fields_ = [
    ("kind", ctypes.c_int),
    ("size", ctypes.c_int),
]
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, "Packet", r.ClassName)
	assert.Equal(t, 2, r.LCOM4)
	assert.Equal(t, 2, r.InstanceVariables)
	assert.Equal(t, [][]string{{"touches_state"}, {"utility"}}, r.MethodGroups)
}
