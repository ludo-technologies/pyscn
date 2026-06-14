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
		{
			name: "class with union type annotations (Python 3.10+)",
			pythonCode: `
class Context:
    pass

class Parameter:
    pass

class Command:
    def __init__(self, ctx: Context | None = None):
        self.ctx = ctx

    def get_param(self, name: str) -> Parameter | None:
        return None

    def process(self, ctx: Context, param: Parameter | None) -> str | int:
        return 0
`,
			expectedCount: 3,
			expectedCBO:   map[string]int{"Context": 0, "Parameter": 0, "Command": 2}, // Command depends on Context and Parameter
			expectedRisk:  map[string]string{"Context": "low", "Parameter": "low", "Command": "low"},
		},
		{
			name: "class with nested union types",
			pythonCode: `
class User:
    pass

class Admin:
    pass

class Guest:
    pass

class AccessControl:
    def get_user(self) -> User | Admin | Guest | None:
        return None

    def check_access(self, user: User | Admin) -> bool:
        return True
`,
			expectedCount: 4,
			expectedCBO:   map[string]int{"User": 0, "Admin": 0, "Guest": 0, "AccessControl": 3}, // Depends on User, Admin, Guest
			expectedRisk:  map[string]string{"User": "low", "Admin": "low", "Guest": "low", "AccessControl": "low"},
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

func TestCBOAnalyzer_DependencyIdentityContract(t *testing.T) {
	pythonCode := `
from __future__ import annotations
from abc import ABC
from typing import Protocol, TypedDict

class Dependency:
    pass

class Payload(TypedDict):
    name: str

class Contract(Protocol):
    def handle(self, item: Dependency) -> Dependency:
        ...

class AbstractBase(ABC):
    pass

class Widget:
    other: Widget

    def adopt(self, owner: Widget) -> Widget:
        return Widget()

class Service:
    field: Dependency

    def set_one(self, item: Dependency) -> Dependency:
        return item

    def set_two(self, item: Dependency) -> Dependency:
        return item
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	analyzer := NewCBOAnalyzer(DefaultCBOOptions())
	results, err := analyzer.AnalyzeClasses(ast, "test.py")
	require.NoError(t, err)

	resultMap := make(map[string]*CBOResult)
	for _, result := range results {
		resultMap[result.ClassName] = result
	}

	payload := resultMap["Payload"]
	require.NotNil(t, payload)
	assert.Equal(t, 0, payload.CouplingCount)
	assert.Equal(t, 0, payload.InheritanceDependencies)
	assert.Equal(t, []string{"TypedDict"}, payload.BaseClasses)
	assert.Empty(t, payload.DependentClasses)

	contract := resultMap["Contract"]
	require.NotNil(t, contract)
	assert.Equal(t, 1, contract.CouplingCount)
	assert.Equal(t, 1, contract.TypeHintDependencies)
	assert.Equal(t, []string{"Dependency"}, contract.DependentClasses)
	assert.Equal(t, []string{"Protocol"}, contract.BaseClasses)

	abstractBase := resultMap["AbstractBase"]
	require.NotNil(t, abstractBase)
	assert.Equal(t, 0, abstractBase.CouplingCount)
	assert.Equal(t, 0, abstractBase.InheritanceDependencies)
	assert.Equal(t, []string{"ABC"}, abstractBase.BaseClasses)
	assert.Empty(t, abstractBase.DependentClasses)

	widget := resultMap["Widget"]
	require.NotNil(t, widget)
	assert.Equal(t, 0, widget.CouplingCount)
	assert.Equal(t, 0, widget.TypeHintDependencies)
	assert.Equal(t, 0, widget.InstantiationDependencies)
	assert.Empty(t, widget.DependentClasses)

	service := resultMap["Service"]
	require.NotNil(t, service)
	assert.Equal(t, 1, service.CouplingCount)
	assert.Equal(t, 1, service.TypeHintDependencies)
	assert.Equal(t, []string{"Dependency"}, service.DependentClasses)
}

func TestCBOAnalyzer_ImportedTypingNamesDoNotHideLocalClasses(t *testing.T) {
	firstFile := `
from typing import TypedDict

class Payload(TypedDict):
    name: str
`

	secondFile := `
class TypedDict:
    pass

class UsesLocal:
    payload: TypedDict
`

	analyzer := NewCBOAnalyzer(DefaultCBOOptions())

	firstAST, err := parseCode(firstFile)
	require.NoError(t, err)

	firstResults, err := analyzer.AnalyzeClasses(firstAST, "first.py")
	require.NoError(t, err)
	require.Len(t, firstResults, 1)
	assert.Equal(t, "Payload", firstResults[0].ClassName)
	assert.Equal(t, 0, firstResults[0].CouplingCount)
	assert.Empty(t, firstResults[0].DependentClasses)

	secondAST, err := parseCode(secondFile)
	require.NoError(t, err)

	secondResults, err := analyzer.AnalyzeClasses(secondAST, "second.py")
	require.NoError(t, err)

	resultMap := make(map[string]*CBOResult)
	for _, result := range secondResults {
		resultMap[result.ClassName] = result
	}

	usesLocal := resultMap["UsesLocal"]
	require.NotNil(t, usesLocal)
	assert.Equal(t, 1, usesLocal.CouplingCount)
	assert.Equal(t, 1, usesLocal.TypeHintDependencies)
	assert.Equal(t, []string{"TypedDict"}, usesLocal.DependentClasses)
}

func TestCBOAnalyzer_MultiArgumentGenericTypeHints(t *testing.T) {
	pythonCode := `
class User:
    pass

class Account:
    pass

class Service:
    cache: dict[str, User]
    pair: tuple[User, Account]
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "service.py")
	require.NoError(t, err)

	resultMap := make(map[string]*CBOResult)
	for _, result := range results {
		resultMap[result.ClassName] = result
	}

	service := resultMap["Service"]
	require.NotNil(t, service)
	assert.Equal(t, 2, service.CouplingCount)
	assert.Equal(t, 2, service.TypeHintDependencies)
	assert.Equal(t, []string{"Account", "User"}, service.DependentClasses)
}

func TestCBOAnalyzer_GenericInheritanceUsesBaseClassIdentity(t *testing.T) {
	pythonCode := `
class Base:
    pass

class User:
    pass

class Repo(Base[User]):
    pass
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "repo.py")
	require.NoError(t, err)

	resultMap := make(map[string]*CBOResult)
	for _, result := range results {
		resultMap[result.ClassName] = result
	}

	repo := resultMap["Repo"]
	require.NotNil(t, repo)
	assert.Equal(t, []string{"Base"}, repo.BaseClasses)
	assert.Equal(t, 1, repo.CouplingCount)
	assert.Equal(t, 1, repo.InheritanceDependencies)
	assert.Equal(t, []string{"Base"}, repo.DependentClasses)
}

func TestCBOAnalyzer_QualifiedTypeStructureDoesNotAddReceiverDependency(t *testing.T) {
	pythonCode := `
import contracts

class Service(contracts.Base):
    item: contracts.Item

    def build(self) -> contracts.Result:
        pass
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "service.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	service := results[0]
	assert.Equal(t, "Service", service.ClassName)
	assert.Equal(t, []string{"contracts.Base"}, service.BaseClasses)
	assert.Equal(t, []string{"contracts.Base", "contracts.Item", "contracts.Result"}, service.DependentClasses)
	assert.NotContains(t, service.DependentClasses, "contracts")
}

func TestCBOAnalyzer_IncludeImportsControlsStandardLibraryDependencies(t *testing.T) {
	pythonCode := `
from pathlib import Path as FilePath

class Reader:
    path: FilePath

    def make_path(self) -> FilePath:
        return FilePath("data.txt")
`

	tests := []struct {
		name                     string
		includeImports           bool
		expectedCBO              int
		expectedImportDeps       int
		expectedDependentClasses []string
	}{
		{
			name:                     "include stdlib imports",
			includeImports:           true,
			expectedCBO:              1,
			expectedImportDeps:       1,
			expectedDependentClasses: []string{"FilePath"},
		},
		{
			name:                     "exclude stdlib imports",
			includeImports:           false,
			expectedCBO:              0,
			expectedImportDeps:       0,
			expectedDependentClasses: []string{},
		},
	}

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := DefaultCBOOptions()
			options.IncludeImports = tt.includeImports

			results, err := NewCBOAnalyzer(options).AnalyzeClasses(ast, "reader.py")
			require.NoError(t, err)
			require.Len(t, results, 1)

			reader := results[0]
			assert.Equal(t, "Reader", reader.ClassName)
			assert.Equal(t, tt.expectedCBO, reader.CouplingCount)
			assert.Equal(t, tt.expectedImportDeps, reader.ImportDependencies)
			assert.Equal(t, tt.expectedDependentClasses, reader.DependentClasses)
		})
	}
}

func TestCBOAnalyzer_QualifiedSameNameDependencyIsNotSelfReference(t *testing.T) {
	pythonCode := `
import models

class Widget:
    parent: models.Widget
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "widget.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	widget := results[0]
	assert.Equal(t, "Widget", widget.ClassName)
	assert.Equal(t, 1, widget.CouplingCount)
	assert.Equal(t, 1, widget.ImportDependencies)
	assert.Equal(t, []string{"models.Widget"}, widget.DependentClasses)
}

func TestCBOAnalyzer_ImportedSameNameDependencyIsSelfReference(t *testing.T) {
	pythonCode := `
from pkg import Widget

class Widget:
    parent: Widget

    def clone(self) -> Widget:
        return Widget()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "widget.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	widget := results[0]
	assert.Equal(t, "Widget", widget.ClassName)
	assert.Equal(t, 0, widget.CouplingCount)
	assert.Equal(t, 0, widget.ImportDependencies)
	assert.Empty(t, widget.DependentClasses)
}

func TestCBOAnalyzer_TypeSystemImportRootExcludesTypingNames(t *testing.T) {
	pythonCode := `
import typing
from abc import ABCMeta
from typing_extensions import NotRequired

class Contract(ABCMeta):
    mapping: typing.MutableMapping
    task: typing.Awaitable
    coro: typing.Coroutine
    gen: typing.Generator
    annotated: typing.Annotated
    missing: NotRequired
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "contract.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	contract := results[0]
	assert.Equal(t, "Contract", contract.ClassName)
	assert.Equal(t, 0, contract.CouplingCount)
	assert.Equal(t, 0, contract.InheritanceDependencies)
	assert.Equal(t, 0, contract.TypeHintDependencies)
	assert.Equal(t, []string{"ABCMeta"}, contract.BaseClasses)
	assert.Empty(t, contract.DependentClasses)
}

func TestCBOAnalyzer_QualifiedProjectTypeSystemNameStillCounts(t *testing.T) {
	pythonCode := `
import models

class Service:
    contract: models.Protocol
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "service.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	service := results[0]
	assert.Equal(t, "Service", service.ClassName)
	assert.Equal(t, 1, service.CouplingCount)
	assert.Equal(t, 1, service.ImportDependencies)
	assert.Equal(t, []string{"models.Protocol"}, service.DependentClasses)
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

func TestCBOAnalyzer_UsesCanonicalASTChildren(t *testing.T) {
	pythonCode := `
class Logger:
    pass

class Service:
    def build(self, flag):
        if flag:
            return None
        else:
            return Logger()
`
	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := CalculateCBO(ast, "test.py")
	require.NoError(t, err)

	var serviceResult *CBOResult
	for _, result := range results {
		if result.ClassName == "Service" {
			serviceResult = result
			break
		}
	}

	require.NotNil(t, serviceResult)
	assert.Equal(t, 1, serviceResult.CouplingCount)
	assert.Equal(t, 1, serviceResult.InstantiationDependencies)
	assert.Contains(t, serviceResult.DependentClasses, "Logger")
}

func TestCBOAnalyzer_NamespaceImportMembersAreGrouped(t *testing.T) {
	pythonCode := `
import libcst as cst
import libcst.matchers as m

class Foo:
    a = cst.Name
    b = cst.Call
    c = cst.Arg
    d = m.Comparison
    e = m.ComparisonTarget
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	t.Run("group namespace imports", func(t *testing.T) {
		options := DefaultCBOOptions()
		options.GroupNamespaceImports = true

		results, err := NewCBOAnalyzer(options).AnalyzeClasses(ast, "foo.py")
		require.NoError(t, err)
		require.Len(t, results, 1)

		foo := results[0]
		assert.Equal(t, 2, foo.CouplingCount, "cst.* and m.* should collapse to two edges")
		assert.Equal(t, []string{"cst", "m"}, foo.DependentClasses)
		assert.Equal(t, 2, foo.ImportDependencies)
	})

	t.Run("keep per-member edges when disabled", func(t *testing.T) {
		options := DefaultCBOOptions()
		options.GroupNamespaceImports = false

		results, err := NewCBOAnalyzer(options).AnalyzeClasses(ast, "foo.py")
		require.NoError(t, err)
		require.Len(t, results, 1)

		foo := results[0]
		assert.Equal(t, 5, foo.CouplingCount, "each alias.Member should be a separate edge")
		assert.Equal(t, []string{"cst.Arg", "cst.Call", "cst.Name", "m.Comparison", "m.ComparisonTarget"}, foo.DependentClasses)
	})
}

func TestCBOAnalyzer_NamespaceImportCallsAreGrouped(t *testing.T) {
	pythonCode := `
import libcst as cst

class Builder:
    def build(self):
        name = cst.Name()
        call = cst.Call()
        arg = cst.Arg()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	options := DefaultCBOOptions()
	options.GroupNamespaceImports = true

	results, err := NewCBOAnalyzer(options).AnalyzeClasses(ast, "builder.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	builder := results[0]
	assert.Equal(t, 1, builder.CouplingCount)
	assert.Equal(t, []string{"cst"}, builder.DependentClasses)
	assert.Equal(t, 1, builder.ImportDependencies)
}

func TestCBOAnalyzer_NonAliasedModuleMembersAreNotGrouped(t *testing.T) {
	pythonCode := `
import datetime

class Scheduler:
    def now(self):
        return datetime.datetime.now()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	options := DefaultCBOOptions()
	options.GroupNamespaceImports = true

	results, err := NewCBOAnalyzer(options).AnalyzeClasses(ast, "scheduler.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	scheduler := results[0]
	assert.Equal(t, []string{"datetime.datetime"}, scheduler.DependentClasses)
	assert.Equal(t, 1, scheduler.CouplingCount)
}

func TestCBOAnalyzer_FunctionCallsAreNotClassDependencies(t *testing.T) {
	// Regression test for #494: imported function calls (os.getcwd, suppress,
	// escape) must not be counted as class dependencies.
	pythonCode := `
import os
from contextlib import suppress
from rich.markup import escape


class Widget:
    pass


class MyThing:
    def run(self):
        cwd = os.getcwd()
        path = os.path.realpath(cwd)
        with suppress(KeyError):
            x = escape("hi")
        return Widget()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "mything.py")
	require.NoError(t, err)

	resultMap := make(map[string]*CBOResult)
	for _, result := range results {
		resultMap[result.ClassName] = result
	}

	myThing, found := resultMap["MyThing"]
	require.True(t, found, "MyThing not found in results")
	assert.Equal(t, []string{"Widget"}, myThing.DependentClasses)
	assert.Equal(t, 1, myThing.CouplingCount)
	assert.Equal(t, "low", myThing.RiskLevel)
}

func TestCBOAnalyzer_LowercaseLocalClassInstantiationStillCounts(t *testing.T) {
	// Locally defined classes are known to be classes regardless of naming.
	pythonCode := `
class widget_factory:
    pass

class Consumer:
    def build(self):
        return widget_factory()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "consumer.py")
	require.NoError(t, err)

	var consumer *CBOResult
	for _, result := range results {
		if result.ClassName == "Consumer" {
			consumer = result
			break
		}
	}

	require.NotNil(t, consumer)
	assert.Equal(t, []string{"widget_factory"}, consumer.DependentClasses)
	assert.Equal(t, 1, consumer.CouplingCount)
}

func TestCBOAnalyzer_KnownLowercaseStdlibClassesStillCount(t *testing.T) {
	// Well-known stdlib classes that break the CapWords convention remain
	// counted via the explicit allowlist.
	pythonCode := `
import datetime
from collections import deque

class Scheduler:
    def setup(self):
        self.queue = deque()
        self.started = datetime.datetime(2024, 1, 1)
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "scheduler.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	scheduler := results[0]
	assert.Equal(t, "Scheduler", scheduler.ClassName)
	assert.Contains(t, scheduler.DependentClasses, "deque")
	assert.Contains(t, scheduler.DependentClasses, "datetime.datetime")
}

func TestCBOAnalyzer_CallHeuristicJudgesOriginalImportedName(t *testing.T) {
	// Aliases resolve to the original imported name before applying the
	// CapWords heuristic: a function aliased to CapWords stays excluded and a
	// class aliased to snake_case stays counted.
	pythonCode := `
from helpers import make_widget as WidgetFactory
from models import Widget as widget_cls

class Consumer:
    def build(self):
        a = WidgetFactory()
        return widget_cls()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "consumer.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	consumer := results[0]
	assert.Equal(t, []string{"widget_cls"}, consumer.DependentClasses)
	assert.Equal(t, 1, consumer.CouplingCount)
}

func TestCBOAnalyzer_ImportedClassMethodCallsCountClassCoupling(t *testing.T) {
	// Class-method/static-method calls on an imported class are real class
	// coupling: the dependency is recorded as the class part of the dotted
	// name, not the method and not nothing.
	pythonCode := `
from pathlib import Path
import datetime

class Worker:
    def run(self):
        here = Path.cwd()
        now = datetime.datetime.now()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "worker.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	worker := results[0]
	assert.Equal(t, "Worker", worker.ClassName)
	assert.Equal(t, []string{"Path", "datetime.datetime"}, worker.DependentClasses)
	assert.Equal(t, 2, worker.CouplingCount)
}

func TestCBOAnalyzer_LocalClassMethodCallsCountClassCoupling(t *testing.T) {
	pythonCode := `
class Widget:
    @classmethod
    def create(cls):
        return cls()

class Consumer:
    def build(self):
        return Widget.create()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "consumer.py")
	require.NoError(t, err)

	var consumer *CBOResult
	for _, result := range results {
		if result.ClassName == "Consumer" {
			consumer = result
			break
		}
	}

	require.NotNil(t, consumer)
	assert.Equal(t, []string{"Widget"}, consumer.DependentClasses)
	assert.Equal(t, 1, consumer.CouplingCount)
}

func TestCBOAnalyzer_OwnClassMethodCallIsNotSelfCoupling(t *testing.T) {
	pythonCode := `
from models import Widget

class Widget:
    def clone(self):
        return Widget.create()
`

	ast, err := parseCode(pythonCode)
	require.NoError(t, err)

	results, err := NewCBOAnalyzer(DefaultCBOOptions()).AnalyzeClasses(ast, "widget.py")
	require.NoError(t, err)
	require.Len(t, results, 1)

	widget := results[0]
	assert.Equal(t, "Widget", widget.ClassName)
	assert.Empty(t, widget.DependentClasses)
	assert.Equal(t, 0, widget.CouplingCount)
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
