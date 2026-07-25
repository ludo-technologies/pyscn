package analyzer

import (
	"testing"

	coreapted "github.com/ludo-technologies/polyscan/core/apted"
	"github.com/stretchr/testify/assert"
)

func TestPythonCostModel(t *testing.T) {
	costModel := NewPythonCostModel()
	node1 := coreapted.NewTreeNode(1, "FunctionDef(test)")
	node2 := coreapted.NewTreeNode(2, "FunctionDef(test)")

	assert.Positive(t, costModel.Insert(node1))
	assert.Positive(t, costModel.Delete(node1))
	assert.Zero(t, costModel.Rename(node1, node2))

	structuralCost := costModel.Insert(coreapted.NewTreeNode(3, "FunctionDef(test)"))
	expressionCost := costModel.Insert(coreapted.NewTreeNode(4, "BinOp(+)"))
	assert.Greater(t, structuralCost, expressionCost)
}

func TestPythonCostModelWithConfig(t *testing.T) {
	leftName := coreapted.NewTreeNode(1, "Name(left)")
	rightName := coreapted.NewTreeNode(2, "Name(right)")
	leftLiteral := coreapted.NewTreeNode(3, "Constant(1)")
	rightLiteral := coreapted.NewTreeNode(4, "Constant(2)")

	strict := NewPythonCostModelWithConfig(false, false)
	assert.Positive(t, strict.Rename(leftName, rightName))
	assert.Positive(t, strict.Rename(leftLiteral, rightLiteral))

	ignoring := NewPythonCostModelWithConfig(true, true)
	assert.Zero(t, ignoring.Rename(leftName, rightName))
	assert.Zero(t, ignoring.Rename(leftLiteral, rightLiteral))
}

func TestCalculateLabelSimilarityTopLevelDefinitions(t *testing.T) {
	costModel := NewPythonCostModel()

	assert.Equal(t, 0.0, costModel.calculateLabelSimilarity("ClassDef(UserProfile)", "ClassDef(ProductInventory)"))
	assert.Equal(t, 0.0, costModel.calculateLabelSimilarity("FunctionDef(foo)", "FunctionDef(bar)"))
	assert.Equal(t, 0.0, costModel.calculateLabelSimilarity("AsyncFunctionDef(async_foo)", "AsyncFunctionDef(async_bar)"))
	assert.Equal(t, 0.3, costModel.calculateLabelSimilarity("ClassDef(UserProfile)", "ClassDef(UserProfile)"))
	assert.Equal(t, 0.3, costModel.calculateLabelSimilarity("Name(x)", "Name(y)"))
	assert.Equal(t, 0.3, costModel.calculateLabelSimilarity("Constant(1)", "Constant(2)"))
}

func TestExtractNameFromLabel(t *testing.T) {
	costModel := NewPythonCostModel()
	tests := map[string]string{
		"ClassDef(MyClass)":        "MyClass",
		"FunctionDef(my_function)": "my_function",
		"Name(x)":                  "x",
		"Constant(42)":             "42",
		"If":                       "",
		"":                         "",
	}

	for label, expected := range tests {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, expected, costModel.extractNameFromLabel(label))
		})
	}
}

func TestIsTopLevelDefinition(t *testing.T) {
	costModel := NewPythonCostModel()

	assert.True(t, costModel.isTopLevelDefinition("ClassDef"))
	assert.True(t, costModel.isTopLevelDefinition("FunctionDef"))
	assert.True(t, costModel.isTopLevelDefinition("AsyncFunctionDef"))
	assert.False(t, costModel.isTopLevelDefinition("Name"))
	assert.False(t, costModel.isTopLevelDefinition("Constant"))
	assert.False(t, costModel.isTopLevelDefinition("If"))
}
