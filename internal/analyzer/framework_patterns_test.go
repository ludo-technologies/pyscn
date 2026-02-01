package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBoilerplateLabel(t *testing.T) {
	tests := []struct {
		label    string
		expected bool
	}{
		// Boilerplate labels
		{"AnnAssign", true},
		{"AnnAssign(name=x)", true},
		{"Decorator", true},
		{"Decorator(name=dataclass)", true},
		{"generic_type", true},
		{"type_parameter", true},
		{"Field(default=None)", true},
		{"field(default_factory=list)", true},
		{"Factory(list)", true},
		{"attrib(default=0)", true},
		{"attr.ib()", true},

		// Non-boilerplate labels
		{"FunctionDef", false},
		{"FunctionDef(name=foo)", false},
		{"ClassDef", false},
		{"If", false},
		{"For", false},
		{"While", false},
		{"Return", false},
		{"Assign", false},
		{"Name(x)", false},
		{"Constant(42)", false},
		{"BinOp(+)", false},
		{"Call", false},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			result := IsBoilerplateLabel(tt.label)
			assert.Equal(t, tt.expected, result, "IsBoilerplateLabel(%q)", tt.label)
		})
	}
}
