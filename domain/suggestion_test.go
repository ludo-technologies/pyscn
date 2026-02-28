package domain

import (
	"strings"
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
					Name:      "deep_func",
					FilePath:  "a.py",
					StartLine: 42,
					Metrics:   ComplexityMetrics{Complexity: 25, NestingDepth: 5},
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
	if !strings.Contains(suggestions[0].Description, "early returns") {
		t.Errorf("expected 'early returns' in description for deep nesting, got: %s", suggestions[0].Description)
	}
	// Steps should mention guard clauses
	if len(suggestions[0].Steps) < 2 || len(suggestions[0].Steps) > 3 {
		t.Fatalf("expected 2-3 steps for complexity nesting, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[1], "guard clauses") {
		t.Errorf("expected 'guard clauses' in steps for deep nesting, got: %v", suggestions[0].Steps)
	}
}

func TestGenerateSuggestions_ComplexityDefault(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "wide_func",
					FilePath: "a.py",
					Metrics:  ComplexityMetrics{Complexity: 15, NestingDepth: 2},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if len(suggestions[0].Steps) < 2 || len(suggestions[0].Steps) > 3 {
		t.Fatalf("expected 2-3 steps for default complexity, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[0], "smaller functions") {
		t.Errorf("expected 'smaller functions' in steps, got: %v", suggestions[0].Steps)
	}
}

func TestGenerateSuggestions_ComplexityLoopStatements(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "loopy_func",
					FilePath: "a.py",
					Metrics:  ComplexityMetrics{Complexity: 15, NestingDepth: 2, LoopStatements: 4},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if !strings.Contains(suggestions[0].Steps[0], "loop bodies") {
		t.Errorf("expected 'loop bodies' in steps for loop-heavy function, got: %v", suggestions[0].Steps)
	}
}

func TestGenerateSuggestions_ComplexityExceptionHandlers(t *testing.T) {
	resp := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "error_func",
					FilePath: "a.py",
					Metrics:  ComplexityMetrics{Complexity: 15, NestingDepth: 2, ExceptionHandlers: 4},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if !strings.Contains(suggestions[0].Steps[0], "exception handlers") {
		t.Errorf("expected 'exception handlers' in steps, got: %v", suggestions[0].Steps)
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
	// Steps should contain "Delete lines"
	if len(afterReturn.Steps) < 2 || len(afterReturn.Steps) > 3 {
		t.Fatalf("expected 2-3 steps for dead code after_return, got %d", len(afterReturn.Steps))
	}
	if !strings.Contains(afterReturn.Steps[0], "Delete lines") {
		t.Errorf("expected 'Delete lines' in steps for after_return, got: %v", afterReturn.Steps)
	}

	// unreachable_branch should be moderate effort
	branch := findSuggestionByFunction(suggestions, "process")
	if branch == nil {
		t.Fatal("expected suggestion for 'process'")
	}
	if branch.Effort != SuggestionEffortModerate {
		t.Errorf("expected moderate effort for unreachable_branch, got %s", branch.Effort)
	}
	// Steps should mention "condition"
	if len(branch.Steps) < 2 || len(branch.Steps) > 3 {
		t.Fatalf("expected 2-3 steps for unreachable_branch, got %d", len(branch.Steps))
	}
	if !strings.Contains(branch.Steps[0], "condition") {
		t.Errorf("expected 'condition' in steps for unreachable_branch, got: %v", branch.Steps)
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
	// Steps should contain "shared function"
	if len(suggestions[0].Steps) != 3 {
		t.Fatalf("expected 3 steps for Type-1 clone, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[0], "shared function") {
		t.Errorf("expected 'shared function' in steps for Type-1, got: %v", suggestions[0].Steps)
	}

	// Type-4 with 2 members should be warning + hard
	if suggestions[1].Severity != SuggestionSeverityWarning {
		t.Errorf("expected warning severity for 2-member group, got %s", suggestions[1].Severity)
	}
	if suggestions[1].Effort != SuggestionEffortHard {
		t.Errorf("expected hard effort for Type-4, got %s", suggestions[1].Effort)
	}
	// Type-4 steps should mention "serve the same purpose"
	if len(suggestions[1].Steps) != 3 {
		t.Fatalf("expected 3 steps for Type-4 clone, got %d", len(suggestions[1].Steps))
	}
	if !strings.Contains(suggestions[1].Steps[0], "serve the same purpose") {
		t.Errorf("expected 'serve the same purpose' in Type-4 steps, got: %v", suggestions[1].Steps)
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
	// Steps should mention "interfaces"
	if len(high.Steps) != 3 {
		t.Fatalf("expected 3 steps for CBO>7, got %d", len(high.Steps))
	}
	if !strings.Contains(high.Steps[1], "interfaces") {
		t.Errorf("expected 'interfaces' in steps for CBO>7, got: %v", high.Steps)
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
	// Medium CBO steps should have 2 items
	if len(med.Steps) != 2 {
		t.Fatalf("expected 2 steps for CBO 4-7, got %d", len(med.Steps))
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
	if !strings.Contains(suggestions[0].Description, "Method groups:") {
		t.Errorf("expected method groups in description, got: %s", suggestions[0].Description)
	}
	// Steps should mention "Split into"
	if len(suggestions[0].Steps) != 3 {
		t.Fatalf("expected 3 steps for LCOM4>5, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[1], "Split into") {
		t.Errorf("expected 'Split into' in steps for LCOM4>5, got: %v", suggestions[0].Steps)
	}
}

func TestGenerateSuggestions_LCOMMedium(t *testing.T) {
	resp := &AnalyzeResponse{
		LCOM: &LCOMResponse{
			Classes: []ClassCohesion{
				{
					Name:     "MedClass",
					FilePath: "a.py",
					Metrics: LCOMMetrics{
						LCOM4:        4,
						MethodGroups: [][]string{{"m1"}, {"m2"}, {"m3"}, {"m4"}},
					},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	// Medium LCOM steps should have 2 items
	if len(suggestions[0].Steps) != 2 {
		t.Fatalf("expected 2 steps for LCOM 3-5, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[0], "method groups") {
		t.Errorf("expected 'method groups' in steps for LCOM 3-5, got: %v", suggestions[0].Steps)
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
		// System suggestions should not have Steps (Description is already specific enough)
		if len(s.Steps) != 0 {
			t.Errorf("expected 0 steps for system cycle breaking (avoid duplication), got %d", len(s.Steps))
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

	// System suggestions should not have Steps (Description is already specific enough)
	for _, s := range suggestions {
		if len(s.Steps) != 0 {
			t.Errorf("expected 0 steps for architecture violation (avoid duplication), got %d", len(s.Steps))
		}
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

	// All suggestions should have Steps
	for i, s := range suggestions {
		if len(s.Steps) == 0 {
			t.Errorf("suggestion %d (%s) has no steps", i, s.Title)
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

func TestGenerateSuggestions_ClonePrioritySurvivesLimit(t *testing.T) {
	// Create 12 clone groups: first 11 are Type-3 with 2 members (warning),
	// 12th is Type-1 with 5 members (critical+easy).
	// The critical group must survive the 10-item cap.
	groups := make([]*CloneGroup, 12)
	for i := 0; i < 11; i++ {
		groups[i] = &CloneGroup{
			ID:         i + 1,
			Type:       Type3Clone,
			Similarity: 0.80,
			Clones:     make([]*Clone, 2), // warning
		}
	}
	groups[11] = &CloneGroup{
		ID:         12,
		Type:       Type1Clone,
		Similarity: 1.0,
		Clones:     make([]*Clone, 5), // critical+easy
	}

	resp := &AnalyzeResponse{
		Clone: &CloneResponse{CloneGroups: groups},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != maxSuggestionsPerCategory {
		t.Fatalf("expected %d suggestions, got %d", maxSuggestionsPerCategory, len(suggestions))
	}

	// The critical+easy group (originally 12th) must be in the result
	found := false
	for _, s := range suggestions {
		if s.Severity == SuggestionSeverityCritical && s.Effort == SuggestionEffortEasy {
			found = true
			break
		}
	}
	if !found {
		t.Error("critical+easy clone suggestion was dropped by the limit; expected it to survive via priority sort")
	}

	// It should also be sorted first
	if suggestions[0].Severity != SuggestionSeverityCritical {
		t.Errorf("first suggestion should be critical, got %s", suggestions[0].Severity)
	}
}

func TestGenerateSuggestions_SystemIndependentLimits(t *testing.T) {
	// 12 cycle-breaking suggestions + 3 architecture violations.
	// Architecture violations must NOT be blocked by the dependency limit.
	cycleStrings := make([]string, 12)
	for i := range cycleStrings {
		cycleStrings[i] = "Break cycle " + string(rune('A'+i))
	}

	resp := &AnalyzeResponse{
		System: &SystemAnalysisResponse{
			DependencyAnalysis: &DependencyAnalysisResult{
				CircularDependencies: &CircularDependencyAnalysis{
					CycleBreakingSuggestions: cycleStrings,
				},
			},
			ArchitectureAnalysis: &ArchitectureAnalysisResult{
				Violations: []ArchitectureViolation{
					{Module: "mod_a", Severity: ViolationSeverityCritical, Suggestion: "Fix A"},
					{Module: "mod_b", Severity: ViolationSeverityWarning, Suggestion: "Fix B"},
					{Module: "mod_c", Severity: ViolationSeverityInfo, Suggestion: "Fix C"},
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)

	depCount := 0
	archCount := 0
	for _, s := range suggestions {
		switch s.Category {
		case SuggestionCategoryDependency:
			depCount++
		case SuggestionCategoryArchitecture:
			archCount++
		}
	}

	if depCount != maxSuggestionsPerCategory {
		t.Errorf("expected %d dependency suggestions, got %d", maxSuggestionsPerCategory, depCount)
	}
	if archCount != 3 {
		t.Errorf("expected 3 architecture suggestions (independent of dependency limit), got %d", archCount)
	}
}

func TestGenerateSuggestions_CloneType2Steps(t *testing.T) {
	resp := &AnalyzeResponse{
		Clone: &CloneResponse{
			CloneGroups: []*CloneGroup{
				{
					ID:         1,
					Type:       Type2Clone,
					Similarity: 0.95,
					Clones:     make([]*Clone, 3),
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if len(suggestions[0].Steps) != 3 {
		t.Fatalf("expected 3 steps for Type-2, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[0], "parameterized function") {
		t.Errorf("expected 'parameterized function' in Type-2 steps, got: %v", suggestions[0].Steps)
	}
}

func TestGenerateSuggestions_CloneType3Steps(t *testing.T) {
	resp := &AnalyzeResponse{
		Clone: &CloneResponse{
			CloneGroups: []*CloneGroup{
				{
					ID:         1,
					Type:       Type3Clone,
					Similarity: 0.80,
					Clones:     make([]*Clone, 2),
				},
			},
		},
	}

	suggestions := GenerateSuggestions(resp)
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}
	if len(suggestions[0].Steps) != 3 {
		t.Fatalf("expected 3 steps for Type-3, got %d", len(suggestions[0].Steps))
	}
	if !strings.Contains(suggestions[0].Steps[0], "common structure") {
		t.Errorf("expected 'common structure' in Type-3 steps, got: %v", suggestions[0].Steps)
	}
}

// helpers

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
