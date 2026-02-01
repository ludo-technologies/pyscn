package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestFrameworkPattern_String(t *testing.T) {
	tests := []struct {
		pattern  FrameworkPattern
		expected string
	}{
		{PatternNone, "none"},
		{PatternDataclass, "dataclass"},
		{PatternPydanticModel, "pydantic"},
		{PatternNamedTuple, "namedtuple"},
		{PatternTypedDict, "typeddict"},
		{PatternAttrs, "attrs"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.pattern.String())
		})
	}
}

func TestNewPatternDetector(t *testing.T) {
	pd := NewPatternDetector()

	assert.True(t, pd.EnableDataclassDetection)
	assert.True(t, pd.EnablePydanticDetection)
	assert.True(t, pd.EnableNamedTupleDetection)
	assert.True(t, pd.EnableTypedDictDetection)
	assert.True(t, pd.EnableAttrsDetection)
}

func TestPatternDetector_DetectPattern_NilNode(t *testing.T) {
	pd := NewPatternDetector()
	assert.Equal(t, PatternNone, pd.DetectPattern(nil))
}

func TestPatternDetector_DetectPattern_NonClassNode(t *testing.T) {
	pd := NewPatternDetector()

	// Non-class nodes should return PatternNone
	funcNode := &parser.Node{Type: parser.NodeFunctionDef}
	assert.Equal(t, PatternNone, pd.DetectPattern(funcNode))
}

func TestPatternDetector_IsBoilerplateNode(t *testing.T) {
	pd := NewPatternDetector()

	tests := []struct {
		name     string
		node     *parser.Node
		expected bool
	}{
		{"nil node", nil, false},
		{"AnnAssign is boilerplate", &parser.Node{Type: parser.NodeAnnAssign}, true},
		{"Decorator is boilerplate", &parser.Node{Type: parser.NodeDecorator}, true},
		{"Assign is not boilerplate", &parser.Node{Type: parser.NodeAssign}, false},
		{"FunctionDef is not boilerplate", &parser.Node{Type: parser.NodeFunctionDef}, false},
		{"If is not boilerplate", &parser.Node{Type: parser.NodeIf}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, pd.IsBoilerplateNode(tt.node))
		})
	}
}

func TestPatternDetector_IsBoilerplateLabel(t *testing.T) {
	pd := NewPatternDetector()

	tests := []struct {
		label    string
		expected bool
	}{
		{"AnnAssign", true},
		{"AnnAssign(name=x)", true},
		{"Decorator", true},
		{"Decorator(name=dataclass)", true},
		{"generic_type", true},
		{"type_parameter", true},
		{"FunctionDef", false},
		{"If", false},
		{"For", false},
		{"ClassDef", false},
		{"Name(x)", false},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			assert.Equal(t, tt.expected, pd.IsBoilerplateLabel(tt.label))
		})
	}
}

func TestPatternDetector_CountBoilerplateNodes(t *testing.T) {
	pd := NewPatternDetector()

	t.Run("nil node", func(t *testing.T) {
		b, total := pd.CountBoilerplateNodes(nil)
		assert.Equal(t, 0, b)
		assert.Equal(t, 0, total)
	})

	t.Run("single boilerplate node", func(t *testing.T) {
		node := &parser.Node{Type: parser.NodeAnnAssign}
		b, total := pd.CountBoilerplateNodes(node)
		assert.Equal(t, 1, b)
		assert.Equal(t, 1, total)
	})

	t.Run("single non-boilerplate node", func(t *testing.T) {
		node := &parser.Node{Type: parser.NodeFunctionDef}
		b, total := pd.CountBoilerplateNodes(node)
		assert.Equal(t, 0, b)
		assert.Equal(t, 1, total)
	})

	t.Run("mixed nodes", func(t *testing.T) {
		parent := &parser.Node{
			Type: parser.NodeClassDef,
			Body: []*parser.Node{
				{Type: parser.NodeAnnAssign},
				{Type: parser.NodeAnnAssign},
				{Type: parser.NodeFunctionDef},
			},
		}
		b, total := pd.CountBoilerplateNodes(parent)
		assert.Equal(t, 2, b)
		assert.Equal(t, 4, total)
	})
}

func TestPatternDetector_CalculateSemanticContentRatio(t *testing.T) {
	pd := NewPatternDetector()

	t.Run("nil node returns 1.0", func(t *testing.T) {
		ratio := pd.CalculateSemanticContentRatio(nil)
		assert.Equal(t, 1.0, ratio)
	})

	t.Run("all boilerplate returns 0.0", func(t *testing.T) {
		node := &parser.Node{Type: parser.NodeAnnAssign}
		ratio := pd.CalculateSemanticContentRatio(node)
		assert.Equal(t, 0.0, ratio)
	})

	t.Run("no boilerplate returns 1.0", func(t *testing.T) {
		node := &parser.Node{Type: parser.NodeFunctionDef}
		ratio := pd.CalculateSemanticContentRatio(node)
		assert.Equal(t, 1.0, ratio)
	})

	t.Run("mixed returns partial ratio", func(t *testing.T) {
		parent := &parser.Node{
			Type: parser.NodeClassDef,
			Body: []*parser.Node{
				{Type: parser.NodeAnnAssign},
				{Type: parser.NodeFunctionDef},
			},
		}
		ratio := pd.CalculateSemanticContentRatio(parent)
		// 1 boilerplate, 3 total (parent + 2 children) -> 2/3 semantic
		assert.Equal(t, 2.0/3.0, ratio)
	})
}

func TestPatternDetector_CountBoilerplateInTreeNode(t *testing.T) {
	pd := NewPatternDetector()

	t.Run("nil node", func(t *testing.T) {
		b, total := pd.CountBoilerplateInTreeNode(nil)
		assert.Equal(t, 0, b)
		assert.Equal(t, 0, total)
	})

	t.Run("boilerplate label", func(t *testing.T) {
		node := &TreeNode{Label: "AnnAssign"}
		b, total := pd.CountBoilerplateInTreeNode(node)
		assert.Equal(t, 1, b)
		assert.Equal(t, 1, total)
	})

	t.Run("non-boilerplate label", func(t *testing.T) {
		node := &TreeNode{Label: "FunctionDef"}
		b, total := pd.CountBoilerplateInTreeNode(node)
		assert.Equal(t, 0, b)
		assert.Equal(t, 1, total)
	})

	t.Run("tree with children", func(t *testing.T) {
		node := &TreeNode{
			Label: "ClassDef",
			Children: []*TreeNode{
				{Label: "AnnAssign"},
				{Label: "AnnAssign"},
				{Label: "FunctionDef"},
			},
		}
		b, total := pd.CountBoilerplateInTreeNode(node)
		assert.Equal(t, 2, b)
		assert.Equal(t, 4, total)
	})
}

func TestPatternDetector_CalculateSemanticContentRatioFromTree(t *testing.T) {
	pd := NewPatternDetector()

	t.Run("nil node returns 1.0", func(t *testing.T) {
		ratio := pd.CalculateSemanticContentRatioFromTree(nil)
		assert.Equal(t, 1.0, ratio)
	})

	t.Run("calculates ratio correctly", func(t *testing.T) {
		node := &TreeNode{
			Label: "ClassDef",
			Children: []*TreeNode{
				{Label: "AnnAssign"},
				{Label: "FunctionDef"},
			},
		}
		ratio := pd.CalculateSemanticContentRatioFromTree(node)
		// 1 boilerplate, 3 total -> 2/3 semantic
		assert.Equal(t, 2.0/3.0, ratio)
	})
}
