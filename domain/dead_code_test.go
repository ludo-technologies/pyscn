package domain

import "testing"

func TestFunctionDeadCodeScopeLabel(t *testing.T) {
	tests := []struct {
		name  string
		scope FunctionDeadCode
		want  string
	}{
		{name: "function", scope: FunctionDeadCode{Name: "build", ScopeKind: AnalysisScopeFunction}, want: "build"},
		{name: "class", scope: FunctionDeadCode{Name: "Config", ScopeKind: AnalysisScopeClass}, want: "class scope Config"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.scope.ScopeLabel(); got != test.want {
				t.Fatalf("ScopeLabel() = %q, want %q", got, test.want)
			}
		})
	}
}
