package analyzer

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corelcom "github.com/ludo-technologies/polyscan/core/lcom"
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
			expectedExcluded: map[string]int{"ClassWithDecorators": 3},
		},
		{
			name: "abstract methods excluded from LCOM4 grouping",
			pythonCode: `
from abc import ABC, abstractmethod

class MyEndpoint(ABC):
    @property
    @abstractmethod
    def _x(self):
        raise NotImplementedError

    @property
    @abstractmethod
    def _y(self):
        raise NotImplementedError

    def get(self):
        return self._x, self._y
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"MyEndpoint": 1},
			expectedRisk:     map[string]string{"MyEndpoint": "low"},
			expectedExcluded: map[string]int{"MyEndpoint": 3},
		},
		{
			name: "dotted abstract methods excluded from LCOM4 grouping",
			pythonCode: `
import abc

class MyProvider:
    @abc.abstractmethod
    def provide(self):
        raise NotImplementedError

    def status(self):
        return self.state
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"MyProvider": 1},
			expectedRisk:     map[string]string{"MyProvider": "low"},
			expectedExcluded: map[string]int{"MyProvider": 1},
		},
		{
			name: "protocol classes are skipped",
			pythonCode: `
from typing import Protocol
import typing

class BatchCommand(Protocol):
    def setup(self) -> None:
        """Set up the command."""

    def execute(self, path) -> None: ...

    def finalize(self) -> None: ...

class Generic(typing.Protocol[T]):
    def a(self): ...
    def b(self): ...
    def c(self): ...

class Impl:
    def setup(self):
        self.ready = True

    def unrelated(self):
        return self.other
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"Impl": 2},
			expectedRisk:  map[string]string{"Impl": "low"},
		},
		{
			name: "aliased and wildcard Protocol imports are skipped",
			pythonCode: `
from typing import Protocol as P
import typing_extensions as te
from enum import *

class A(P):
    def a(self): return 1
    def b(self): return 2
    def c(self): return 3

class B(te.Protocol):
    def a(self): return 1
    def b(self): return 2
    def c(self): return 3

class C(IntEnum):
    X = 1
    def a(self): return 1
    def b(self): return 2
    def c(self): return 3
`,
			expectedCount: 0,
		},
		{
			name: "same-named classes from other modules are still analyzed",
			pythonCode: `
from twisted.internet import protocol
from mylib import Enum

class Handler(protocol.Protocol):
    def dataReceived(self, data):
        self.buffer += data

    def connectionMade(self):
        self.transport.write(self.banner)

    def timeout(self):
        return self.deadline

class Custom(Enum):
    def a(self): return self.x
    def b(self): return self.y
    def c(self): return self.z

class Local(Protocol):
    def a(self): return self.x
    def b(self): return self.y
    def c(self): return self.z
`,
			expectedCount: 3,
			expectedLCOM:  map[string]int{"Handler": 3, "Custom": 3, "Local": 3},
			expectedRisk:  map[string]string{"Handler": "medium", "Custom": "medium", "Local": "medium"},
		},
		{
			name: "enum classes are skipped",
			pythonCode: `
from enum import Enum, IntEnum
import enum

class Color(Enum):
    RED = 1
    BLUE = 2

    def is_warm(self):
        return self is Color.RED

    def label(self):
        return self.name.lower()

    def code(self):
        return self.value

class Level(enum.IntEnum):
    LOW = 1

    def a(self): return 1
    def b(self): return 2
    def c(self): return 3
`,
			expectedCount: 0,
		},
		{
			name: "empty-bodied methods excluded from LCOM4 grouping",
			pythonCode: `
class Hooks:
    def on_start(self):
        """Called before processing."""

    def on_stop(self):
        pass

    def on_error(self, exc): ...

    def render(self):
        raise NotImplementedError

    def validate(self):
        """Validate."""
        raise NotImplementedError("subclass must implement")

    def run(self):
        return self.state

    def reset(self):
        self.state = None
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"Hooks": 1},
			expectedRisk:     map[string]string{"Hooks": "low"},
			expectedExcluded: map[string]int{"Hooks": 5},
		},
		{
			name: "NotImplementedError that reads self is not an empty body",
			pythonCode: `
class Partial:
    def unsupported(self):
        raise NotImplementedError(self.description)

    def chained(self):
        raise NotImplementedError from self.cause

    def run(self):
        return self.state

    def reset(self):
        self.state = None
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"Partial": 3},
			expectedRisk:     map[string]string{"Partial": "medium"},
			expectedExcluded: map[string]int{"Partial": 0},
		},
		{
			name: "stateless raise of another exception is excluded",
			pythonCode: `
class Guard:
    def check(self):
        raise ValueError("bad")

    def run(self):
        return self.state

    def reset(self):
        self.state = None
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"Guard": 1},
			expectedRisk:     map[string]string{"Guard": "low"},
			expectedExcluded: map[string]int{"Guard": 1},
		},
		{
			name: "raising other exceptions is not an empty body",
			pythonCode: `
class Guard:
    def check(self):
        raise ValueError(self.reason)

    def run(self):
        return self.state

    def reset(self):
        self.state = None
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"Guard": 2},
			expectedRisk:     map[string]string{"Guard": "low"},
			expectedExcluded: map[string]int{"Guard": 0},
		},
		{
			name: "implicit classmethods excluded from LCOM4 grouping",
			pythonCode: `
class Registry:
    def __init_subclass__(cls, **kwargs):
        cls.registry.append(cls)

    def __class_getitem__(cls, item):
        return cls

    def run(self):
        return self.state

    def reset(self):
        self.state = None
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"Registry": 1},
			expectedRisk:     map[string]string{"Registry": "low"},
			expectedExcluded: map[string]int{"Registry": 2},
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
			name: "method consuming a property is connected via property-call edge",
			pythonCode: `
class PropertyConsumer:
    def __init__(self):
        self._x = 1

    @property
    def value(self):
        return self._x

    def use(self):
        return self.value + 1
`,
			expectedCount: 1,
			expectedLCOM:  map[string]int{"PropertyConsumer": 1},
			expectedRisk:  map[string]string{"PropertyConsumer": "low"},
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
			name: "methods without self access are excluded from the graph",
			pythonCode: `
class NoSelfAccessClass:
    def method_a(self):
        return 42

    def method_b(self):
        return 99
`,
			expectedCount:    1,
			expectedLCOM:     map[string]int{"NoSelfAccessClass": 1},
			expectedRisk:     map[string]string{"NoSelfAccessClass": "low"},
			expectedExcluded: map[string]int{"NoSelfAccessClass": 2},
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
	config := corelcom.Config{
		LowThreshold:    2,
		MediumThreshold: 5,
	}

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
			assert.Equal(t, tt.expected, string(corelcom.AssessRisk(tt.lcom4, config)))
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
    def peer(self):
        return self._sep

class B:
    def __init__(self):
        self._sep = '/'
    def render(self):
        return f"x{1:{self._sep}>5}y"
    def peer(self):
        return self._sep

class C:
    def __init__(self):
        self._sep = '/'
    def render(self):
        return "x" + self._sep + "y"
    def peer(self):
        return self._sep

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
    def peer(self):
        return self._x
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
		"A": {{"peer", "render"}},
		"B": {{"peer", "render"}},
		"C": {{"peer", "render"}},
		"D": {{"m1", "m2"}},
		"E": {{"peer", "render"}},
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
		"B":             {{"__del__", "write"}},
		"K":             {{"m1", "m2"}},
		"AsyncContext":  {{"acquire", "current"}},
		"MultipleItems": {{"copy", "read_a", "read_b"}},
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
	assert.Equal(t, [][]string{{"export", "is_ready", "resize"}}, r.MethodGroups)
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
	assert.Equal(t, 1, r.LCOM4)
	assert.Equal(t, 2, r.InstanceVariables)
	assert.Equal(t, 1, r.ExcludedMethods, "utility touches no declared field")
	assert.Equal(t, [][]string{{"touches_state"}}, r.MethodGroups)
}

func TestLCOMAnalyzer_CtypesUnionAndEndianBases(t *testing.T) {
	p := parser.New()
	code := `
import ctypes
from ctypes import Union, BigEndianStructure

class Payload(Union):
    def as_int(self):
        return _AsInt(self)

    def as_float(self):
        return _AsFloat(self)

Payload._fields_ = [
    ("i", ctypes.c_int),
    ("f", ctypes.c_float),
]

class NetHeader(BigEndianStructure):
    def serialize(self):
        return _Serialize(self)

NetHeader._fields_ = [
    ("magic", ctypes.c_uint32),
    ("length", ctypes.c_uint16),
]
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 2)

	byName := map[string]*LCOMResult{}
	for _, r := range results {
		byName[r.ClassName] = r
	}
	require.Contains(t, byName, "Payload")
	require.Contains(t, byName, "NetHeader")
	assert.Equal(t, 1, byName["Payload"].LCOM4)
	assert.Equal(t, 2, byName["Payload"].InstanceVariables)
	assert.Equal(t, 2, byName["NetHeader"].InstanceVariables)
}

func TestLCOMAnalyzer_CtypesAmbiguousClassNameSkipsExternalFields(t *testing.T) {
	p := parser.New()
	// Two ctypes Structure classes share the name "Inner" at different scopes.
	// External `Inner._fields_ = ...` cannot be safely attributed to either, so
	// neither should receive the declared fields.
	code := `
import ctypes

def factory_a():
    class Inner(ctypes.Structure):
        def use(self):
            _Use(self)
    return Inner

def factory_b():
    class Inner(ctypes.Structure):
        def use(self):
            _Use(self)
    return Inner

Inner._fields_ = [
    ("x", ctypes.c_int),
]
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	for _, r := range results {
		if r.ClassName == "Inner" {
			assert.Equal(t, 0, r.InstanceVariables, "ambiguous external _fields_ must not be attributed")
		}
	}
}

// TestLCOMAnalyzer_Issue627Repro pins the two classes cited in
// https://github.com/ludo-technologies/pyscn/issues/627, which claimed
// InstanceVariables is always 0. Both already reported correct non-zero
// values on this code path (the report did not reproduce); this guards
// against a future regression on these real-world shapes.
func TestLCOMAnalyzer_Issue627Repro(t *testing.T) {
	t.Run("Point", func(t *testing.T) {
		p := parser.New()
		code := `
def my_decorator(func):
    return func

class Point:
    def __init__(self, x: int, y: int) -> None:
        self.x = x
        self.y = y

    def abs(self) -> 'Point':
        return Point(abs(self.x), abs(self.y))

    def add(self, other: 'Point'):
        self.x += other.x
        self.y += other.y

    def to_origin(self):
        self.x = 0
        self.y = 0

    def ignored(self):
        self.foo = 'bar'

    def __len__(self):
        return 0

    @staticmethod
    def from_coords(coords) -> 'Point':
        return Point(coords[0], coords[1])

    @property
    def coords(self):
        return self.x, self.y

    @staticmethod
    def skip_static_decorator_pragma(a: int, b: int) -> int:
        return a + b * 2

    @classmethod
    def skip_class_decorator_pragma(cls, value: int) -> "Point":
        return cls(value + 1, value * 2)

    def skip_instance_method_pragma(self) -> int:
        return self.x + self.y * 2

    @staticmethod
    def pragma_on_staticmethod_decorator(a: int, b: int) -> int:
        return a + b * 2

    @classmethod
    def pragma_on_classmethod_decorator(cls, value: int) -> "Point":
        return cls(value + 1, value * 2)

    @my_decorator
    @classmethod
    def skip_multi_decorator(cls, value: int) -> "Point":
        return cls(value + 1, value * 2)
`
		result, err := p.Parse(context.Background(), []byte(code))
		require.NoError(t, err)

		analyzer := NewLCOMAnalyzer(nil)
		results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
		require.NoError(t, err)
		require.Len(t, results, 1)

		r := results[0]
		assert.Equal(t, "Point", r.ClassName)
		assert.Equal(t, 14, r.TotalMethods)
		assert.Equal(t, 8, r.ExcludedMethods, "decorators, __init__, and stateless __len__")
		assert.Equal(t, 3, r.InstanceVariables, "self.x, self.y, self.foo")
		assert.Equal(t, 2, r.LCOM4, "x/y methods and ignored; __len__ is stateless")
	})

	t.Run("HammettRunner", func(t *testing.T) {
		p := parser.New()
		code := `
class TestRunner:
    pass

class HammettRunner(TestRunner):
    def __init__(self, config) -> None:
        self.hammett_kwargs = None

    def other_method(self):
        return self.hammett_kwargs
`
		result, err := p.Parse(context.Background(), []byte(code))
		require.NoError(t, err)

		analyzer := NewLCOMAnalyzer(nil)
		results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
		require.NoError(t, err)

		byName := make(map[string]*LCOMResult, len(results))
		for _, res := range results {
			byName[res.ClassName] = res
		}

		r, ok := byName["HammettRunner"]
		require.True(t, ok, "missing class HammettRunner")
		assert.Equal(t, 1, r.InstanceVariables, "self.hammett_kwargs")
		assert.Equal(t, [][]string{{"other_method"}}, r.MethodGroups)
	})
}

// TestLCOMAnalyzer_ConstructorsExcludedFromGraph pins the repro from
// https://github.com/ludo-technologies/pyscn/issues/698. A constructor that
// initializes every attribute unions all of the class's responsibility
// clusters, so leaving it in the graph reported LCOM4=1 for a class with two
// plainly separate concerns.
func TestLCOMAnalyzer_ConstructorsExcludedFromGraph(t *testing.T) {
	p := parser.New()
	code := `
class Incohesive:
    def __init__(self):
        self.a = 1
        self.b = 2
    def uses_a(self):
        return self.a + 1
    def also_a(self):
        return self.a * 2
    def uses_b(self):
        return self.b - 1
    def also_b(self):
        return self.b / 2
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 2, r.LCOM4, "the two concerns must stay separate")
	assert.Equal(t, [][]string{{"also_a", "uses_a"}, {"also_b", "uses_b"}}, r.MethodGroups)
	assert.Equal(t, 5, r.TotalMethods)
	assert.Equal(t, 1, r.ExcludedMethods, "__init__")
	assert.Equal(t, 2, r.InstanceVariables, "self.a, self.b stay counted")
}

// TestLCOMAnalyzer_ConstructorVariantsExcluded covers the other two shapes that
// initialize an instance: __new__ and the dataclass __post_init__ hook.
func TestLCOMAnalyzer_ConstructorVariantsExcluded(t *testing.T) {
	p := parser.New()
	code := `
class ViaNew:
    def __new__(cls):
        self = super().__new__(cls)
        self.a = 1
        self.b = 2
        return self
    def uses_a(self):
        return self.a
    def uses_b(self):
        return self.b

class ViaPostInit:
    def __post_init__(self):
        self.a = 1
        self.b = 2
    def uses_a(self):
        return self.a
    def uses_b(self):
        return self.b
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)

	byName := make(map[string]*LCOMResult, len(results))
	for _, res := range results {
		byName[res.ClassName] = res
	}

	for _, className := range []string{"ViaNew", "ViaPostInit"} {
		t.Run(className, func(t *testing.T) {
			r, ok := byName[className]
			require.True(t, ok, "missing class %s", className)
			assert.Equal(t, 2, r.LCOM4)
			assert.Equal(t, [][]string{{"uses_a"}, {"uses_b"}}, r.MethodGroups)
			assert.Equal(t, 1, r.ExcludedMethods)
		})
	}
}

// TestLCOMAnalyzer_ConstructorOnlyVariablesStillCounted guards the
// InstanceVariables statistic: an attribute that only the constructor touches
// must survive the constructor leaving the graph.
func TestLCOMAnalyzer_ConstructorOnlyVariablesStillCounted(t *testing.T) {
	p := parser.New()
	code := `
class Holder:
    def __init__(self):
        self.used = 1
        self.only_here = 2
    def read(self):
        return self.used
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 2, r.InstanceVariables, "self.used and self.only_here")
	assert.Equal(t, 1, r.LCOM4)
	assert.Equal(t, [][]string{{"read"}}, r.MethodGroups)
}

// TestLCOMAnalyzer_ConstructorOnlyClass keeps a class whose only method is a
// constructor trivially cohesive rather than reporting an empty graph.
func TestLCOMAnalyzer_ConstructorOnlyClass(t *testing.T) {
	p := parser.New()
	code := `
class OnlyCtor:
    def __init__(self):
        self.a = 1
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 1, r.LCOM4)
	assert.Equal(t, 1, r.TotalMethods)
	assert.Equal(t, 1, r.ExcludedMethods)
	assert.Equal(t, 1, r.InstanceVariables)
	assert.Equal(t, "low", r.RiskLevel)
}

// TestLCOMAnalyzer_ConstructorPropertyReadsAreNotVariables guards the property
// reclassification on the constructor path. A bare `self.<prop>` read invokes
// the getter through the descriptor protocol, so it must not reach
// InstanceVariables just because the read happens inside a constructor.
func TestLCOMAnalyzer_ConstructorPropertyReadsAreNotVariables(t *testing.T) {
	p := parser.New()
	code := `
class Cached:
    def __init__(self):
        self._width = 1
        self.cached = self.doubled
    @property
    def doubled(self):
        return self._width * 2
    def show(self):
        return self._width
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 2, r.InstanceVariables, "self._width and self.cached; self.doubled is a getter call")
	assert.Equal(t, 1, r.LCOM4)
	assert.Equal(t, [][]string{{"doubled", "show"}}, r.MethodGroups)
}

// TestLCOMAnalyzer_ConstructorOnlyPropertyReadAddsNoVariable pins the narrow
// case where a property is the only thing the constructor reads, so a leak
// would be visible as a variable no method actually stores.
func TestLCOMAnalyzer_ConstructorOnlyPropertyReadAddsNoVariable(t *testing.T) {
	p := parser.New()
	code := `
class Probe:
    def __init__(self):
        print(self.ready)
    @property
    def ready(self):
        return True
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 0, r.InstanceVariables, "the class stores no instance state")
	assert.Equal(t, 2, r.ExcludedMethods, "__init__ and the stateless property")
}

// TestLCOMAnalyzer_MethodReferencesAreNotVariables pins issue #678: a
// `self.<method>` reference names a method defined in the class, not
// instance state, whether it is called or passed along as a callback.
func TestLCOMAnalyzer_MethodReferencesAreNotVariables(t *testing.T) {
	p := parser.New()
	code := `
class C:
    def __init__(self):
        self._a = 1
        self._b = 2
        self._setup()
    def _setup(self):
        register(self.helper)
    def helper(self):
        return self._a
    def run(self):
        self.helper()
        self._callback()
        return self._b
`
	result, err := p.Parse(context.Background(), []byte(code))
	require.NoError(t, err)

	analyzer := NewLCOMAnalyzer(nil)
	results, err := analyzer.AnalyzeClasses(result.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	r := results[0]
	assert.Equal(t, 3, r.InstanceVariables, "self._a, self._b and self._callback; _setup and helper are methods")
	assert.Equal(t, 1, r.LCOM4)
}

func TestLCOMAnalyzer_IterationProtocolConnectsMethods(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Iterable:
    def __iter__(self):
        return iter(self.items)

    def via_list(self):
        return list(self)

    def via_iter(self):
        return iter(self)

    def via_loop(self):
        for item in self:
            return item

    def via_comprehension(self):
        return [item for item in self]

    def unrelated(self, other):
        return list(other)
`)

	require.Equal(t, 1, r.LCOM4)
	assert.ElementsMatch(t, [][]string{
		{"__iter__", "via_comprehension", "via_iter", "via_list", "via_loop"},
	}, r.MethodGroups)
}

func TestLCOMAnalyzer_IterationPrefersDeclaredIterator(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Both:
    def __iter__(self):
        return iter(self.iter_state)

    def __getitem__(self, index):
        return self.item_state[index]

    def consume(self):
        return list(self)
`)

	require.Equal(t, 2, r.LCOM4)
	assert.ElementsMatch(t, [][]string{
		{"__iter__", "consume"},
		{"__getitem__"},
	}, r.MethodGroups)
}

func TestLCOMAnalyzer_SubscriptProtocolsRespectAccessContext(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Mapping:
    def __getitem__(self, key):
        return self.read_store[key]

    def __setitem__(self, key, value):
        self.write_store[key] = value

    def __delitem__(self, key):
        del self.delete_store[key]

    def read(self):
        return self[0]

    def write(self):
        self[0] = 1

    def remove(self):
        del self[0]

    def index_of_foreign_target(self, other):
        other[self[0]] = 1

    def foreign_write(self, other):
        other[0] = 1
`)

	require.Equal(t, 3, r.LCOM4)
	assert.ElementsMatch(t, [][]string{
		{"__getitem__", "index_of_foreign_target", "read"},
		{"__setitem__", "write"},
		{"__delitem__", "remove"},
	}, r.MethodGroups)
}

func TestLCOMAnalyzer_AugmentedSubscriptReadsAndWrites(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Counter:
    def __getitem__(self, key):
        return self.read_store[key]

    def __setitem__(self, key, value):
        self.write_store[key] = value

    def increment(self):
        self[0] += 1
`)

	require.Equal(t, 1, r.LCOM4)
	assert.Equal(t, [][]string{{"__getitem__", "__setitem__", "increment"}}, r.MethodGroups)
}

func TestLCOMAnalyzer_DictConversionUsesMappingProtocol(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Mapping:
    def __iter__(self):
        return iter(self.iter_store)

    def __getitem__(self, key):
        return self.item_store[key]

    def keys(self):
        return self.key_store

    def as_dict(self):
        return dict(self)

    def foreign_dict(self, other):
        return dict(other)
`)

	require.Equal(t, 2, r.LCOM4)
	assert.ElementsMatch(t, [][]string{
		{"__getitem__", "as_dict", "keys"},
		{"__iter__"},
	}, r.MethodGroups)
}

func TestLCOMAnalyzer_ComparisonProtocolsRespectReceiver(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Comparisons:
    def __eq__(self, other):
        return self.equality_state == other

    def equal(self, other):
        return self == other

    def foreign_left(self, other):
        return other == self

    def mixed_chain(self, other):
        return self < other == 3
`)

	require.Equal(t, 1, r.LCOM4)
	assert.ElementsMatch(t, [][]string{
		{"__eq__", "equal"},
	}, r.MethodGroups)
}

func TestLCOMAnalyzer_ProtocolOperationsRespectReceiver(t *testing.T) {
	tests := []struct {
		name, method, expression string
	}{
		{"length", "__len__", "len(self)"},
		{"membership", "__contains__", "other in self"},
		{"negated membership", "__contains__", "other not in self"},
		{"inequality", "__ne__", "self != other"},
		{"less than", "__lt__", "self < other"},
		{"less or equal", "__le__", "self <= other"},
		{"greater than", "__gt__", "self > other"},
		{"greater or equal", "__ge__", "self >= other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeLCOMClass(t, `
class Subject:
    def `+tt.method+`(self, other=None):
        return self.state

    def check(self, other):
        return `+tt.expression+`

    def unrelated(self, other):
        return `+strings.ReplaceAll(tt.expression, "self", "other")+`
`)
			assert.ElementsMatch(t, [][]string{{tt.method, "check"}}, r.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_ProtocolFallbacksRespectPrecedence(t *testing.T) {
	tests := []struct {
		name, expression, methods string
		groups                    [][]string
	}{
		{
			name:       "inequality uses equality when no inequality override exists",
			expression: "self != other",
			methods: `
    def __eq__(self, other):
        return self.equality_state == other
`,
			groups: [][]string{{"__eq__", "check"}},
		},
		{
			name:       "membership override precedes iteration",
			expression: "other in self",
			methods: `
    def __contains__(self, other):
        return other in self.members

    def __iter__(self):
        return iter(self.iter_state)
`,
			groups: [][]string{{"__contains__", "check"}, {"__iter__"}},
		},
		{
			name:       "membership falls back to iteration",
			expression: "other not in self",
			methods: `
    def __iter__(self):
        return iter(self.iter_state)

    def __getitem__(self, index):
        return self.item_state[index]
`,
			groups: [][]string{{"__iter__", "check"}, {"__getitem__"}},
		},
		{
			name:       "membership falls back to indexed access",
			expression: "other in self",
			methods: `
    def __getitem__(self, index):
        return self.item_state[index]
`,
			groups: [][]string{{"__getitem__", "check"}},
		},
		{
			name:       "dict consumes pairs without a keys method",
			expression: "dict(self)",
			methods: `
    def __iter__(self):
        return iter(self.pairs)
`,
			groups: [][]string{{"__iter__", "check"}},
		},
		{
			name:       "iteration falls back to indexed access",
			expression: "list(self)",
			methods: `
    def __getitem__(self, index):
        return self.item_state[index]
`,
			groups: [][]string{{"__getitem__", "check"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeLCOMClass(t, "class Subject:\n"+tt.methods+`
    def check(self, other):
        return `+tt.expression+"\n")
			assert.ElementsMatch(t, tt.groups, r.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_InheritedProtocolsDoNotCreateLocalEdges(t *testing.T) {
	tests := []struct {
		name   string
		source string
		groups [][]string
	}{
		{
			name: "inherited iterator prevents assumed item fallback",
			source: `
class Derived(Base):
    def __getitem__(self, key):
        return self.items[key]

    def consume(self):
        return list(self)
`,
			groups: [][]string{{"__getitem__"}},
		},
		{
			name: "inherited keys prevents assumed pair iteration",
			source: `
class Derived(Base):
    def __iter__(self):
        return iter(self.iter_state)

    def __getitem__(self, key):
        return self.item_state[key]

    def as_dict(self):
        return dict(self)
`,
			groups: [][]string{{"__iter__"}, {"__getitem__"}},
		},
		{
			name: "inherited membership prevents assumed iteration",
			source: `
class Derived(Base):
    def __iter__(self):
        return iter(self.items)

    def has(self, item):
        return item in self
`,
			groups: [][]string{{"__iter__"}},
		},
		{
			name: "inherited inequality prevents assumed equality",
			source: `
class Derived(Base):
    def __eq__(self, other):
        return self.state == other

    def differs(self, other):
        return self != other
`,
			groups: [][]string{{"__eq__"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeLCOMClass(t, tt.source)
			assert.ElementsMatch(t, tt.groups, r.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_DisabledProtocolsStopFallback(t *testing.T) {
	tests := []struct {
		name, disabled, method, expression string
	}{
		{"membership", "__contains__", "__iter__", "other in self"},
		{"iteration", "__iter__", "__getitem__", "list(self)"},
		{"inequality", "__ne__", "__eq__", "self != other"},
		{"mapping", "keys", "__getitem__", "dict(self)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeLCOMClass(t, `
class Subject:
    `+tt.disabled+` = None

    def `+tt.method+`(self, other=None):
        return self.state

    def check(self, other):
        return `+tt.expression+"\n")
			assert.ElementsMatch(t, [][]string{{tt.method}}, r.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_ProtocolMethodReplacesDisabledBinding(t *testing.T) {
	r := analyzeLCOMClass(t, `
class Container:
    __contains__ = None

    def __contains__(self, item):
        return item in self.items

    def __iter__(self):
        return iter(self.iter_state)

    def has(self, item):
        return item in self
`)

	assert.ElementsMatch(t, [][]string{{"__contains__", "has"}, {"__iter__"}}, r.MethodGroups)
}

func TestLCOMAnalyzer_AssignmentReplacesProtocolMethod(t *testing.T) {
	tests := []struct {
		name, method, expression, extraMethod string
	}{
		{"equality", "__eq__", "self == other", ""},
		{"length", "__len__", "len(self)", ""},
		{"list length", "__len__", "list(self)", ""},
		{"mapping item", "__getitem__", "dict(self)", "keys"},
		{"subscript", "__getitem__", "self[other]", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := `
class Subject:
    def ` + tt.method + `(self, other=None):
        return self.state

    ` + tt.method + ` = None

    def check(self, other):
        return ` + tt.expression + "\n"
			var groups [][]string
			if tt.extraMethod != "" {
				source += "\n    def " + tt.extraMethod + "(self):\n        return self.keys_state\n"
				groups = [][]string{{tt.method}, {"check", tt.extraMethod}}
			} else {
				groups = [][]string{{tt.method}}
			}
			r := analyzeLCOMClass(t, source)
			assert.ElementsMatch(t, groups, r.MethodGroups)
		})
	}
}

func TestLCOMAnalyzer_ListUsesDeclaredLength(t *testing.T) {
	r := analyzeLCOMClass(t, `
class SizedIterable:
    def __iter__(self):
        return iter(self.items)

    def __len__(self):
        return self.size

    def consume(self):
        return list(self)
`)

	assert.Equal(t, [][]string{{"__iter__", "__len__", "consume"}}, r.MethodGroups)
}

func TestLCOMAnalyzer_OrderedMappingProtocolRepro(t *testing.T) {
	r := analyzeLCOMClass(t, `
class OrderedDictLike:
    def __init__(self):
        self.__map = {}

    def __setitem__(self, key, value):
        self.__map[key] = value

    def __iter__(self):
        return iter(self.__map)

    def __getitem__(self, key):
        return self.__map[key]

    def keys(self):
        return list(self)

    def values(self):
        return [self[key] for key in self]

    def __eq__(self, other):
        return dict(self) == dict(other)

    def __ne__(self, other):
        return not self == other

    def update(self, other):
        for key in other:
            self[key] = other[key]
`)

	require.Equal(t, 1, r.LCOM4)
	assert.Equal(t, 9, r.TotalMethods)
	assert.Equal(t, 1, r.ExcludedMethods)
	assert.Equal(t, 1, r.InstanceVariables)
	assert.Equal(t, [][]string{{
		"__eq__", "__getitem__", "__iter__", "__ne__", "__setitem__", "keys", "update", "values",
	}}, r.MethodGroups)
}

// TestLCOMAnalyzer_StatelessMethodsExcludedFromGraph pins
// https://github.com/ludo-technologies/pyscn/issues/810. A method that touches
// no instance state and is not connected to any other method by a call can
// only be its own component, so it inflates LCOM4 without naming anything to
// split. Those methods are left out of the graph the same way abstract methods
// are.
func TestLCOMAnalyzer_StatelessMethodsExcludedFromGraph(t *testing.T) {
	t.Run("closure dunders in a nested class", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
def make_operator(name, module):
    class OperatorImplementation:
        def __call__(self, *args):
            return f"{module}.{name}(...)"

        def __repr__(self):
            return f"{module}.{name}"

        def __str__(self):
            return f"{name}"

    return OperatorImplementation()
`)
		assert.Equal(t, "OperatorImplementation", r.ClassName)
		assert.Equal(t, 1, r.LCOM4)
		assert.Equal(t, 3, r.TotalMethods)
		assert.Equal(t, 3, r.ExcludedMethods)
		assert.Equal(t, 0, r.InstanceVariables)
		assert.Empty(t, r.MethodGroups)
	})

	t.Run("tzinfo overrides returning constants", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
class UTC(tzinfo):
    def utcoffset(self, dt):
        return timedelta(0)

    def dst(self, dt):
        return timedelta(0)

    def tzname(self, dt):
        return "UTC"
`)
		assert.Equal(t, 1, r.LCOM4)
		assert.Equal(t, 3, r.TotalMethods)
		assert.Equal(t, 3, r.ExcludedMethods)
		assert.Equal(t, 0, r.InstanceVariables)
		assert.Empty(t, r.MethodGroups)
	})

	t.Run("operator dunder passing self to a free function", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
class Sequence:
    def __add__(self, other):
        return chain(self, other)

    def append(self, item):
        self.items.append(item)

    def extend(self, items):
        self.items.extend(items)
`)
		assert.Equal(t, 1, r.LCOM4)
		assert.Equal(t, 3, r.TotalMethods)
		assert.Equal(t, 1, r.ExcludedMethods)
		assert.Equal(t, 1, r.InstanceVariables)
		assert.Equal(t, [][]string{{"append", "extend"}}, r.MethodGroups)
	})

	t.Run("stateless helper called by a stateful method stays in the graph", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
class Viewer:
    def _fmt(self):
        return "x"

    def show(self):
        return self._fmt() + self.name
`)
		assert.Equal(t, 1, r.LCOM4)
		assert.Equal(t, 2, r.TotalMethods)
		assert.Equal(t, 0, r.ExcludedMethods)
		assert.Equal(t, [][]string{{"_fmt", "show"}}, r.MethodGroups)
	})

	t.Run("stateless method that calls a sibling stays in the graph", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
class Label:
    def __repr__(self):
        return self.__str__()

    def __str__(self):
        return self.name
`)
		assert.Equal(t, 1, r.LCOM4)
		assert.Equal(t, 2, r.TotalMethods)
		assert.Equal(t, 0, r.ExcludedMethods)
		assert.Equal(t, [][]string{{"__repr__", "__str__"}}, r.MethodGroups)
	})

	t.Run("disjoint state stays split when a stateless method is excluded", func(t *testing.T) {
		r := analyzeLCOMClass(t, `
class Split:
    def read_a(self):
        return self.a

    def write_a(self, v):
        self.a = v

    def read_b(self):
        return self.b

    def label(self):
        return "split"
`)
		assert.Equal(t, 2, r.LCOM4)
		assert.Equal(t, 4, r.TotalMethods)
		assert.Equal(t, 1, r.ExcludedMethods)
		assert.Equal(t, 2, r.InstanceVariables)
		assert.Equal(t, [][]string{{"read_a", "write_a"}, {"read_b"}}, r.MethodGroups)
	})
}

func analyzeLCOMClass(t *testing.T, source string) *LCOMResult {
	t.Helper()
	parsed, err := parser.New().Parse(context.Background(), []byte(source))
	require.NoError(t, err)
	results, err := NewLCOMAnalyzer(nil).AnalyzeClasses(parsed.AST, "test.py")
	require.NoError(t, err)
	require.Len(t, results, 1)
	return results[0]
}
