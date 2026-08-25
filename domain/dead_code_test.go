package domain

import (
	"testing"

	"gopkg.in/yaml.v3"
)

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

func TestFileDeadCodeExecutionScopes(t *testing.T) {
	file := FileDeadCode{
		Functions:   []FunctionDeadCode{{Name: "build", ScopeKind: AnalysisScopeFunction}},
		ClassScopes: []FunctionDeadCode{{Name: "Config", ScopeKind: AnalysisScopeClass}},
	}

	scopes := file.ExecutionScopes()
	if len(scopes) != 2 || scopes[0].Name != "build" || scopes[1].Name != "Config" {
		t.Fatalf("ExecutionScopes() = %+v", scopes)
	}
	scopes[0].Name = "changed"
	if file.Functions[0].Name != "build" {
		t.Fatal("ExecutionScopes() must not alias file storage")
	}
}

func TestFileDeadCodeYAMLUsesCanonicalKeys(t *testing.T) {
	encoded, err := yaml.Marshal(FileDeadCode{
		FilePath: "config.py",
		ClassScopes: []FunctionDeadCode{{
			Name:      "Config",
			ScopeKind: AnalysisScopeClass,
		}},
		TotalClassScopes:    1,
		AffectedClassScopes: 1,
	})
	if err != nil {
		t.Fatalf("marshal file dead code: %v", err)
	}

	var decoded map[string]any
	if err := yaml.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode raw YAML keys: %v", err)
	}
	for _, key := range []string{"file_path", "class_scopes", "total_class_scopes", "affected_class_scopes"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing snake_case key %q in %v", key, decoded)
		}
	}
	if _, legacy := decoded["ClassScopes"]; legacy {
		t.Fatalf("unexpected legacy key in %v", decoded)
	}
	classScopes, ok := decoded["class_scopes"].([]any)
	if !ok || len(classScopes) != 1 {
		t.Fatalf("class_scopes = %#v", decoded["class_scopes"])
	}
	classScope, ok := classScopes[0].(map[string]any)
	if !ok || classScope["scope_kind"] != string(AnalysisScopeClass) {
		t.Fatalf("class scope keys = %#v", classScopes[0])
	}
}
