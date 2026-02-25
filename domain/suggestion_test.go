package domain

import (
	"testing"
)

func TestGenerateSuggestions_EmptyResponse(t *testing.T) {
	resp := &AnalyzeResponse{}
	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 0 {
		t.Errorf("expected 0 suggestions for empty response, got %d", len(suggestions))
	}
}

func TestGenerateSuggestions_ComplexityOnly(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "high_func",
					FilePath: "a.py",
					Metrics:  ComplexityMetrics{Complexity: 25, NestingDepth: 6},
				},
				{
					Name:     "medium_func",
					FilePath: "b.py",
					Metrics:  ComplexityMetrics{Complexity: 15, NestingDepth: 2},
				},
				{
					Name:     "low_func",
					FilePath: "c.py",
					Metrics:  ComplexityMetrics{Complexity: 5},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// First suggestion should be high complexity (critical)
	if suggestions[0].Severity != SuggestionSeverityCritical {
		t.Errorf("expected critical severity for complexity=25, got %s", suggestions[0].Severity)
	}
	if suggestions[0].Effort != SuggestionEffortModerate {
		t.Errorf("expected moderate effort, got %s", suggestions[0].Effort)
	}
	if suggestions[0].Category != SuggestionCategoryComplexity {
		t.Errorf("expected complexity category, got %s", suggestions[0].Category)
	}

	// Second suggestion should be medium complexity (warning)
	if suggestions[1].Severity != SuggestionSeverityWarning {
		t.Errorf("expected warning severity for complexity=15, got %s", suggestions[1].Severity)
	}
}

func TestGenerateSuggestions_ComplexityNestingDepth(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "deep_func",
					FilePath: "a.py",
					Metrics:  ComplexityMetrics{Complexity: 25, NestingDepth: 5},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if suggestions[0].Description == "" {
		t.Error("expected non-empty description")
	}
	// Should suggest early returns for deep nesting
	if !contains(suggestions[0].Description, "early returns") {
		t.Errorf("expected 'early returns' in description for deep nesting, got: %s", suggestions[0].Description)
	}
}

func TestGenerateSuggestions_DeadCodeOnly(t *testing.T) {
	resp := &AnalyzeResponse{
		DeadCode: &DeadCodeResponse{
			Files: []FileDeadCode{
				{
					Functions: []FunctionDeadCode{
						{
							Findings: []DeadCodeFinding{
								{
									Location:     DeadCodeLocation{FilePath: "a.py", StartLine: 10, EndLine: 15},
									FunctionName: "validate",
									Reason:       "after_return",
									Severity:     DeadCodeSeverityCritical,
									Description:  "Code after return is unreachable",
								},
								{
									Location:     DeadCodeLocation{FilePath: "b.py", StartLine: 20, EndLine: 30},
									FunctionName: "process",
									Reason:       "unreachable_branch",
									Severity:     DeadCodeSeverityWarning,
									Description:  "Branch is never reached",
								},
							},
						},
					},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// after_return should be easy effort
	afterReturn := findSuggestionByFunction(suggestions, "validate")
	if afterReturn == nil {
		t.Fatal("expected suggestion for 'validate'")
	}
	if afterReturn.Effort != SuggestionEffortEasy {
		t.Errorf("expected easy effort for after_return, got %s", afterReturn.Effort)
	}
	if afterReturn.Severity != SuggestionSeverityCritical {
		t.Errorf("expected critical severity, got %s", afterReturn.Severity)
	}

	// unreachable_branch should be moderate effort
	branch := findSuggestionByFunction(suggestions, "process")
	if branch == nil {
		t.Fatal("expected suggestion for 'process'")
	}
	if branch.Effort != SuggestionEffortModerate {
		t.Errorf("expected moderate effort for unreachable_branch, got %s", branch.Effort)
	}
}

func TestGenerateSuggestions_CloneGroups(t *testing.T) {
	resp := &AnalyzeResponse{
		Clone: &CloneResponse{
			CloneGroups: []*CloneGroup{
				{
					ID:         1,
					Type:       Type1Clone,
					Similarity: 1.0,
					Clones:     make([]*Clone, 5), // 5 members -> critical
				},
				{
					ID:         2,
					Type:       Type4Clone,
					Similarity: 0.75,
					Clones:     make([]*Clone, 2), // 2 members -> warning
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	// Type-1 with 5 members should be critical + easy
	if suggestions[0].Severity != SuggestionSeverityCritical {
		t.Errorf("expected critical severity for 5-member group, got %s", suggestions[0].Severity)
	}
	if suggestions[0].Effort != SuggestionEffortEasy {
		t.Errorf("expected easy effort for Type-1, got %s", suggestions[0].Effort)
	}

	// Type-4 with 2 members should be warning + hard
	if suggestions[1].Severity != SuggestionSeverityWarning {
		t.Errorf("expected warning severity for 2-member group, got %s", suggestions[1].Severity)
	}
	if suggestions[1].Effort != SuggestionEffortHard {
		t.Errorf("expected hard effort for Type-4, got %s", suggestions[1].Effort)
	}
}

func TestGenerateSuggestions_CBOOnly(t *testing.T) {
	resp := &AnalyzeResponse{
		CBO: &CBOResponse{
			Classes: []ClassCoupling{
				{
					Name:     "HighClass",
					FilePath: "a.py",
					Metrics:  CBOMetrics{CouplingCount: 10, DependentClasses: []string{"A", "B"}},
				},
				{
					Name:     "MedClass",
					FilePath: "b.py",
					Metrics:  CBOMetrics{CouplingCount: 5, DependentClasses: []string{"C"}},
				},
				{
					Name:     "LowClass",
					FilePath: "c.py",
					Metrics:  CBOMetrics{CouplingCount: 2},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions (LowClass excluded), got %d", len(suggestions))
	}

	high := findSuggestionByClass(suggestions, "HighClass")
	if high == nil {
		t.Fatal("expected suggestion for HighClass")
	}
	if high.Severity != SuggestionSeverityCritical {
		t.Errorf("expected critical for CBO=10, got %s", high.Severity)
	}
	if high.Effort != SuggestionEffortHard {
		t.Errorf("expected hard effort for CBO>7, got %s", high.Effort)
	}

	med := findSuggestionByClass(suggestions, "MedClass")
	if med == nil {
		t.Fatal("expected suggestion for MedClass")
	}
	if med.Severity != SuggestionSeverityWarning {
		t.Errorf("expected warning for CBO=5, got %s", med.Severity)
	}
	if med.Effort != SuggestionEffortModerate {
		t.Errorf("expected moderate effort for CBO 4-7, got %s", med.Effort)
	}
}

func TestGenerateSuggestions_LCOMOnly(t *testing.T) {
	resp := &AnalyzeResponse{
		LCOM: &LCOMResponse{
			Classes: []ClassCohesion{
				{
					Name:     "BadClass",
					FilePath: "a.py",
					Metrics: LCOMMetrics{
						LCOM4:        7,
						MethodGroups: [][]string{{"m1", "m2"}, {"m3"}, {"m4"}, {"m5"}, {"m6"}, {"m7"}, {"m8"}},
					},
				},
				{
					Name:     "OkClass",
					FilePath: "b.py",
					Metrics:  LCOMMetrics{LCOM4: 1},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion (OkClass excluded), got %d", len(suggestions))
	}

	if suggestions[0].Severity != SuggestionSeverityCritical {
		t.Errorf("expected critical for LCOM4=7, got %s", suggestions[0].Severity)
	}
	if suggestions[0].Effort != SuggestionEffortHard {
		t.Errorf("expected hard effort for LCOM4>5, got %s", suggestions[0].Effort)
	}
	if !contains(suggestions[0].Description, "Method groups:") {
		t.Errorf("expected method groups in description, got: %s", suggestions[0].Description)
	}
}

func TestGenerateSuggestions_SystemCycleBreaking(t *testing.T) {
	resp := &AnalyzeResponse{
		System: &SystemAnalysisResponse{
			DependencyAnalysis: &DependencyAnalysisResult{
				CircularDependencies: &CircularDependencyAnalysis{
					CycleBreakingSuggestions: []string{
						"Move shared logic from module_a to a new utility module",
						"Introduce an interface to break the cycle between X and Y",
					},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}

	for _, s := range suggestions {
		if s.Category != SuggestionCategoryDependency {
			t.Errorf("expected dependency category, got %s", s.Category)
		}
		if s.Severity != SuggestionSeverityCritical {
			t.Errorf("expected critical severity for cycle breaking, got %s", s.Severity)
		}
		if s.Effort != SuggestionEffortHard {
			t.Errorf("expected hard effort for cycle breaking, got %s", s.Effort)
		}
	}
}

func TestGenerateSuggestions_SystemArchViolation(t *testing.T) {
	resp := &AnalyzeResponse{
		System: &SystemAnalysisResponse{
			ArchitectureAnalysis: &ArchitectureAnalysisResult{
				Violations: []ArchitectureViolation{
					{
						Module:     "service.users",
						Severity:   ViolationSeverityCritical,
						Suggestion: "Move database access out of the service layer",
					},
					{
						Module:     "utils.helpers",
						Severity:   ViolationSeverityWarning,
						Suggestion: "Avoid direct imports from presentation layer",
					},
					{
						Module:     "core.model",
						Severity:   ViolationSeverityInfo,
						Suggestion: "", // empty suggestion should be skipped
					},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions (empty suggestion skipped), got %d", len(suggestions))
	}
}

func TestGenerateSuggestions_AllAnalysesEnabled_SortOrder(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{Name: "complex_func", Metrics: ComplexityMetrics{Complexity: 25}},
			},
		},
		DeadCode: &DeadCodeResponse{
			Files: []FileDeadCode{
				{
					Functions: []FunctionDeadCode{
						{
							Findings: []DeadCodeFinding{
								{
									Location:     DeadCodeLocation{FilePath: "a.py", StartLine: 1, EndLine: 5},
									FunctionName: "dead_func",
									Reason:       "after_return",
									Severity:     DeadCodeSeverityCritical,
								},
							},
						},
					},
				},
			},
		},
		CBO: &CBOResponse{
			Classes: []ClassCoupling{
				{Name: "CoupledClass", Metrics: CBOMetrics{CouplingCount: 10, DependentClasses: []string{"A"}}},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) < 3 {
		t.Fatalf("expected at least 3 suggestions, got %d", len(suggestions))
	}

	// Critical+Easy should come first (dead code after_return)
	if suggestions[0].Effort != SuggestionEffortEasy {
		t.Errorf("expected easy effort first (critical+easy), got %s with effort %s", suggestions[0].Title, suggestions[0].Effort)
	}

	// Verify ordering: critical+easy before critical+moderate before critical+hard
	for i := 1; i < len(suggestions); i++ {
		prevPri := suggestionPriority(suggestions[i-1])
		currPri := suggestionPriority(suggestions[i])
		if prevPri > currPri {
			t.Errorf("suggestion %d (priority %d) should not come before suggestion %d (priority %d): %s vs %s",
				i-1, prevPri, i, currPri, suggestions[i-1].Title, suggestions[i].Title)
		}
	}
}

func TestGenerateSuggestions_CategoryLimit(t *testing.T) {
	// Create 20 high-complexity functions; only 10 should become suggestions
	funcs := make([]FunctionComplexity, 20)
	for i := range funcs {
		funcs[i] = FunctionComplexity{
			Name:    "func_" + string(rune('a'+i)),
			Metrics: ComplexityMetrics{Complexity: 25},
		}
	}

	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{Functions: funcs},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != maxSuggestionsPerCategory {
		t.Errorf("expected %d suggestions (category limit), got %d", maxSuggestionsPerCategory, len(suggestions))
	}
}

func TestSuggestionPriority_Ordering(t *testing.T) {
	tests := []struct {
		name     string
		severity SuggestionSeverity
		effort   SuggestionEffort
		wantPri  int
	}{
		{"critical+easy", SuggestionSeverityCritical, SuggestionEffortEasy, 0},
		{"critical+moderate", SuggestionSeverityCritical, SuggestionEffortModerate, 1},
		{"warning+easy", SuggestionSeverityWarning, SuggestionEffortEasy, 2},
		{"warning+moderate", SuggestionSeverityWarning, SuggestionEffortModerate, 3},
		{"critical+hard", SuggestionSeverityCritical, SuggestionEffortHard, 4},
		{"warning+hard", SuggestionSeverityWarning, SuggestionEffortHard, 5},
		{"info+easy", SuggestionSeverityInfo, SuggestionEffortEasy, 6},
		{"info+moderate", SuggestionSeverityInfo, SuggestionEffortModerate, 7},
		{"info+hard", SuggestionSeverityInfo, SuggestionEffortHard, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Suggestion{Severity: tt.severity, Effort: tt.effort}
			got := suggestionPriority(s)
			if got != tt.wantPri {
				t.Errorf("suggestionPriority(%s) = %d, want %d", tt.name, got, tt.wantPri)
			}
		})
	}

	// Verify ordering is strictly increasing
	for i := 1; i < len(tests); i++ {
		prev := tests[i-1].wantPri
		curr := tests[i].wantPri
		if prev >= curr {
			t.Errorf("priority of %s (%d) should be less than %s (%d)",
				tests[i-1].name, prev, tests[i].name, curr)
		}
	}
}

func TestSuggestionSeverityIcon(t *testing.T) {
	tests := []struct {
		severity SuggestionSeverity
		want     string
	}{
		{SuggestionSeverityCritical, "\U0001F534"},
		{SuggestionSeverityWarning, "\U0001F7E1"},
		{SuggestionSeverityInfo, "\U0001F535"},
	}
	for _, tt := range tests {
		s := Suggestion{Severity: tt.severity}
		if got := s.SeverityIcon(); got != tt.want {
			t.Errorf("SeverityIcon() for %s = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

// helpers

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func findSuggestionByFunction(suggestions []Suggestion, fn string) *Suggestion {
	for i := range suggestions {
		if suggestions[i].Function == fn {
			return &suggestions[i]
		}
	}
	return nil
}

func findSuggestionByClass(suggestions []Suggestion, cls string) *Suggestion {
	for i := range suggestions {
		if suggestions[i].ClassName == cls {
			return &suggestions[i]
		}
	}
	return nil
}
