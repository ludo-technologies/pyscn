package domain_test

import (
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestAnalyzeSummary_Validate(t *testing.T) {
	tests := []struct {
		name    string
		summary domain.AnalyzeSummary
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid summary",
			summary: domain.AnalyzeSummary{
				ArchEnabled:               true,
				ArchCompliance:            0.8,
				DepsEnabled:               true,
				DepsMainSequenceDeviation: 0.3,
				CodeDuplication:           15.5,
				AverageComplexity:         5.0,
			},
			wantErr: false,
		},
		{
			name: "invalid arch compliance over 1",
			summary: domain.AnalyzeSummary{
				ArchEnabled:    true,
				ArchCompliance: 100.0, // Should be 0-1, not 0-100
			},
			wantErr: true,
			errMsg:  "ArchCompliance must be 0-1",
		},
		{
			name: "invalid arch compliance negative",
			summary: domain.AnalyzeSummary{
				ArchEnabled:    true,
				ArchCompliance: -0.5,
			},
			wantErr: true,
			errMsg:  "ArchCompliance must be 0-1",
		},
		{
			name: "invalid code duplication over 100",
			summary: domain.AnalyzeSummary{
				CodeDuplication: 150.0, // Over 100%
			},
			wantErr: true,
			errMsg:  "CodeDuplication must be 0-100",
		},
		{
			name: "invalid code duplication negative",
			summary: domain.AnalyzeSummary{
				CodeDuplication: -10.0,
			},
			wantErr: true,
			errMsg:  "CodeDuplication must be 0-100",
		},
		{
			name: "disabled arch analysis with invalid value is OK",
			summary: domain.AnalyzeSummary{
				ArchEnabled:    false,
				ArchCompliance: 100.0, // Invalid but ignored when disabled
			},
			wantErr: false,
		},
		{
			name: "negative average complexity",
			summary: domain.AnalyzeSummary{
				AverageComplexity: -5.0,
			},
			wantErr: true,
			errMsg:  "AverageComplexity cannot be negative",
		},
		{
			name: "invalid deps main sequence deviation",
			summary: domain.AnalyzeSummary{
				DepsEnabled:               true,
				DepsMainSequenceDeviation: 1.5, // Over 1.0
			},
			wantErr: true,
			errMsg:  "DepsMainSequenceDeviation must be 0-1",
		},
		{
			name: "invalid deps modules in cycles",
			summary: domain.AnalyzeSummary{
				DepsEnabled:         true,
				DepsTotalModules:    10,
				DepsModulesInCycles: 15, // More than total
			},
			wantErr: true,
			errMsg:  "DepsModulesInCycles",
		},
		{
			name: "invalid high coupling classes",
			summary: domain.AnalyzeSummary{
				CBOClasses:          10,
				HighCouplingClasses: 15, // More than total
			},
			wantErr: true,
			errMsg:  "HighCouplingClasses",
		},
		{
			name: "disabled deps with invalid values is OK",
			summary: domain.AnalyzeSummary{
				DepsEnabled:               false,
				DepsMainSequenceDeviation: 2.0, // Invalid but ignored
				DepsTotalModules:          5,
				DepsModulesInCycles:       10, // Invalid but ignored
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.summary.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message should contain %q, got %q", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestAnalyzeSummary_CalculateHealthScore(t *testing.T) {
	tests := []struct {
		name                      string
		summary                   domain.AnalyzeSummary
		expectedScore             int
		expectedGrade             string
		expectError               bool
		expectedComplexityScore   int
		expectedDeadCodeScore     int
		expectedDuplicationScore  int
		expectedCouplingScore     int
		expectedDependencyScore   int
		expectedArchitectureScore int
	}{
		{
			name: "perfect score",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 2.0,
				CodeDuplication:   0.0,
				ArchEnabled:       true,
				ArchCompliance:    1.0,
			},
			expectedScore:             100,
			expectedGrade:             "A",
			expectError:               false,
			expectedComplexityScore:   100,
			expectedDeadCodeScore:     100,
			expectedDuplicationScore:  100,
			expectedCouplingScore:     100,
			expectedDependencyScore:   100,
			expectedArchitectureScore: 100,
		},
		{
			name: "typical 74 score case",
			summary: domain.AnalyzeSummary{
				AverageComplexity:         7.0,  // Continuous: (7-2)/13*20 = 7.69 → 8
				CodeDuplication:           15.0, // Continuous: 15/10*20 = 30 → 20 (capped, 0-10% scale)
				CBOClasses:                10,
				HighCouplingClasses:       2, // 20% ratio: 0.20/0.12*20 = 33.33 → 20 (capped)
				DepsEnabled:               true,
				DepsMainSequenceDeviation: 1.0, // 1.0*3 = 3 (new max MSD)
				ArchEnabled:               true,
				ArchCompliance:            0.125, // (1-0.125)*12 = 10.5 → 11 (new max arch)
			},
			expectedScore:             38,  // Updated: 100-8-20-20-3-11 = 38 (0-10% duplication scale, capped)
			expectedGrade:             "F", // 38 < 45 = F
			expectError:               false,
			expectedComplexityScore:   60,  // 100 - (8/20)*100 = 60
			expectedDeadCodeScore:     100, // No dead code
			expectedDuplicationScore:  0,   // 100 - (20/20)*100 = 0 (0-10% scale: 20 penalty capped)
			expectedCouplingScore:     0,   // 100 - (20/20)*100 = 0
			expectedDependencyScore:   80,  // Normalized: (3/16)*20 = 3.75 → 4, Score: 100 - (4/20)*100 = 80
			expectedArchitectureScore: 13,  // Compliance 0.125 * 100 = 12.5 → 13
		},
		{
			name: "moderate complexity and duplication",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 12.0, // Continuous: (12-2)/13*20 = 15.38 → 15
				CodeDuplication:   30.0, // Continuous: (30-1)/7*20 = 82.86 → 20 (capped)
				ArchEnabled:       false,
				DepsEnabled:       false,
			},
			expectedScore: 65,  // Updated: 100-15-20 = 65
			expectedGrade: "C", // 60 ≤ 65 < 75 = C
			expectError:   false,
		},
		{
			name: "high complexity",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 25.0, // Continuous: (25-2)/13*20 = 35.38 → 20 (capped)
				CodeDuplication:   5.0,  // 5/10*20 = 10 penalty (0-10% scale)
			},
			expectedScore: 70,  // Updated: 100-20-10 = 70 (0-10% duplication scale)
			expectedGrade: "C", // 60 ≤ 70 < 75 = C
			expectError:   false,
		},
		{
			name: "invalid data - negative complexity",
			summary: domain.AnalyzeSummary{
				ArchEnabled:       true,
				ArchCompliance:    0.5,
				AverageComplexity: -5.0, // Invalid
			},
			expectError: true,
		},
		{
			name: "invalid data - arch compliance over 1",
			summary: domain.AnalyzeSummary{
				ArchEnabled:    true,
				ArchCompliance: 1.5, // Invalid
			},
			expectError: true,
		},
		{
			name: "minimum score floor",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   25.0, // Capped at 20
				CodeDuplication:     50.0, // Capped at 20
				CBOClasses:          10,
				HighCouplingClasses: 6, // 60% ratio: 0.60/0.30*20 = capped at 20
				DeadCodeCount:       100,
				CriticalDeadCode:    50, // 50/1.0 = capped at 20
				TotalFiles:          10, // For normalization
				DepsEnabled:         true,
				DepsTotalModules:    10,
				DepsModulesInCycles: 10, // 1.0*10 = 10 (new max cycles)
				ArchEnabled:         true,
				ArchCompliance:      0.0, // (1-0)*12 = 12 (new max arch)
			},
			expectedScore: 0, // Floor now 0 (actual: 100-20-20-20-20-10-12 = -2, clamped to 0)
			expectedGrade: "F",
			expectError:   false,
		},
		{
			name: "grade A threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   4.0, // Continuous: (4-2)/13*20 = 3.08 → 3
				CodeDuplication:     2.0, // 2/10*20 = 4 penalty (0-10% scale)
				CBOClasses:          20,
				HighCouplingClasses: 2, // 10% ratio: 0.10/0.12*20 = 16.67 → 17
				DepsEnabled:         true,
				DepsTotalModules:    10,
				DepsMaxDepth:        4, // Expected = max(3, ceil(log2(11))+1) = 5, excess = 4-5 = 0
				ArchEnabled:         true,
				ArchCompliance:      0.9, // (1-0.9)*12 = 1.2 → 1
			},
			expectedScore: 75,  // Updated: 100-3-4-17-0-1 = 75 (0-10% duplication scale)
			expectedGrade: "B", // 75 ≤ 75 < 90 = B
			expectError:   false,
		},
		{
			name: "grade C threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   15.0, // Continuous: (15-2)/13*20 = 20 (capped)
				CodeDuplication:     25.0, // 25/20*20 = 25 → capped at 20 penalty (0-20% scale)
				CBOClasses:          20,
				HighCouplingClasses: 2, // 10% ratio: 0.10/0.12*20 = 16.67 → 17
				DeadCodeCount:       5,
				CriticalDeadCode:    0, // No critical issues, so no dead code penalty
				TotalFiles:          1,
			},
			expectedScore: 43,  // Updated: 100-20-20-17 = 43 (new 0-20% duplication scale, capped)
			expectedGrade: "F", // 43 < 45 = F
			expectError:   false,
		},
		{
			name: "grade D threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   22.0, // Capped at 20
				CodeDuplication:     45.0, // Capped at 20
				CBOClasses:          10,
				HighCouplingClasses: 6, // 60% ratio: 0.60/0.30*20 = capped at 20
			},
			expectedScore: 40,  // Updated: 100-20-20-20 = 40
			expectedGrade: "F", // 40 < 45 = F (stricter grade D threshold)
			expectError:   false,
		},
		{
			name: "edge case - complexity boundary at 2.0",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 2.0, // Exactly at boundary, should result in 0 penalty
			},
			expectedScore:           100,
			expectedGrade:           "A",
			expectError:             false,
			expectedComplexityScore: 100, // 0 penalty
		},
		{
			name: "edge case - duplication at 1.0%",
			summary: domain.AnalyzeSummary{
				CodeDuplication: 1.0, // 1% duplication = 2 penalty (0-10% scale: 1/10*20=2)
			},
			expectedScore:            98,
			expectedGrade:            "A",
			expectError:              false,
			expectedDuplicationScore: 90, // penalty=2, score=100-(2/20)*100=90
		},
		{
			name: "edge case - small weighted dead code",
			summary: domain.AnalyzeSummary{
				CriticalDeadCode: 0,
				WarningDeadCode:  0,
				InfoDeadCode:     1, // Weighted = 0.2
				TotalFiles:       10,
			},
			expectedScore:         100, // Very small weighted value, normalized to 0 penalty
			expectedGrade:         "A",
			expectError:           false,
			expectedDeadCodeScore: 100, // Should be 100 due to very small penalty
		},
		{
			name: "edge case - worst dependency score (score should be 0)",
			summary: domain.AnalyzeSummary{
				DepsEnabled:               true,
				DepsTotalModules:          10,
				DepsModulesInCycles:       10,  // 100% in cycles: 10 penalty
				DepsMaxDepth:              20,  // Very deep: 3 penalty (capped)
				DepsMainSequenceDeviation: 1.0, // Max deviation: 3 penalty
			},
			expectedScore:           84, // 100 - 16 = 84
			expectedGrade:           "B",
			expectError:             false,
			expectedDependencyScore: 0, // Normalized: (16/16)*20 = 20, Score: 100 - (20/20)*100 = 0
		},
		{
			name: "edge case - worst architecture score (score should be 0)",
			summary: domain.AnalyzeSummary{
				ArchEnabled:    true,
				ArchCompliance: 0.0, // 0% compliance: 12 penalty
			},
			expectedScore:             88, // 100 - 12 = 88
			expectedGrade:             "B",
			expectError:               false,
			expectedArchitectureScore: 0, // Normalized: (12/12)*20 = 20, Score: 100 - (20/20)*100 = 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.summary.CalculateHealthScore()

			if (err != nil) != tt.expectError {
				t.Errorf("CalculateHealthScore() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError {
				if tt.summary.HealthScore != tt.expectedScore {
					t.Errorf("HealthScore = %d, want %d", tt.summary.HealthScore, tt.expectedScore)
				}
				if tt.summary.Grade != tt.expectedGrade {
					t.Errorf("Grade = %s, want %s", tt.summary.Grade, tt.expectedGrade)
				}

				// Check individual scores if specified
				if tt.expectedComplexityScore > 0 {
					if tt.summary.ComplexityScore != tt.expectedComplexityScore {
						t.Errorf("ComplexityScore = %d, want %d", tt.summary.ComplexityScore, tt.expectedComplexityScore)
					}
				}
				if tt.expectedDeadCodeScore > 0 {
					if tt.summary.DeadCodeScore != tt.expectedDeadCodeScore {
						t.Errorf("DeadCodeScore = %d, want %d", tt.summary.DeadCodeScore, tt.expectedDeadCodeScore)
					}
				}
				if tt.expectedDuplicationScore > 0 {
					if tt.summary.DuplicationScore != tt.expectedDuplicationScore {
						t.Errorf("DuplicationScore = %d, want %d", tt.summary.DuplicationScore, tt.expectedDuplicationScore)
					}
				}
				if tt.expectedCouplingScore > 0 {
					if tt.summary.CouplingScore != tt.expectedCouplingScore {
						t.Errorf("CouplingScore = %d, want %d", tt.summary.CouplingScore, tt.expectedCouplingScore)
					}
				}
				if tt.expectedDependencyScore > 0 {
					if tt.summary.DependencyScore != tt.expectedDependencyScore {
						t.Errorf("DependencyScore = %d, want %d", tt.summary.DependencyScore, tt.expectedDependencyScore)
					}
				}
				if tt.expectedArchitectureScore > 0 {
					if tt.summary.ArchitectureScore != tt.expectedArchitectureScore {
						t.Errorf("ArchitectureScore = %d, want %d", tt.summary.ArchitectureScore, tt.expectedArchitectureScore)
					}
				}
			} else {
				// When error occurs, check default values
				if tt.summary.HealthScore != 0 {
					t.Errorf("HealthScore should be 0 on error, got %d", tt.summary.HealthScore)
				}
				if tt.summary.Grade != "N/A" {
					t.Errorf("Grade should be N/A on error, got %s", tt.summary.Grade)
				}
				// All scores should be 0 on error
				if tt.summary.ComplexityScore != 0 {
					t.Errorf("ComplexityScore should be 0 on error, got %d", tt.summary.ComplexityScore)
				}
			}
		})
	}
}

func TestAnalyzeSummary_IsHealthy(t *testing.T) {
	tests := []struct {
		name        string
		healthScore int
		want        bool
	}{
		{"score 100", 100, true},
		{"score 85", 85, true},
		{"score 70", 70, true},
		{"score 69", 69, false},
		{"score 50", 50, false},
		{"score 0", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &domain.AnalyzeSummary{
				HealthScore: tt.healthScore,
			}
			if got := s.IsHealthy(); got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGradeFromScore(t *testing.T) {
	tests := []struct {
		name  string
		score int
		want  string
	}{
		{"perfect score", 100, "A"},
		{"grade A boundary", 90, "A"}, // Updated: 85→90
		{"grade A above boundary", 95, "A"},
		{"grade B boundary", 75, "B"}, // Updated: 70→75
		{"grade B above boundary", 80, "B"},
		{"grade B below A", 89, "B"},  // Updated: 84→89
		{"grade C boundary", 60, "C"}, // Updated: 55→60
		{"grade C above boundary", 65, "C"},
		{"grade C below B", 74, "C"},  // Updated: 69→74
		{"grade D boundary", 45, "D"}, // Updated: 40→45
		{"grade D above boundary", 50, "D"},
		{"grade D below C", 59, "D"}, // Updated: 54→59
		{"grade F high", 44, "F"},    // Updated: 39→44
		{"grade F low", 10, "F"},
		{"grade F zero", 0, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.GetGradeFromScore(tt.score); got != tt.want {
				t.Errorf("GetGradeFromScore(%d) = %v, want %v", tt.score, got, tt.want)
			}
		})
	}
}

func TestAnalyzeSummary_CalculateFallbackScore(t *testing.T) {
	tests := []struct {
		name    string
		summary domain.AnalyzeSummary
		want    int
	}{
		{
			name:    "perfect score",
			summary: domain.AnalyzeSummary{},
			want:    100,
		},
		{
			name: "with average complexity above threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 15.0,
			},
			want: 90,
		},
		{
			name: "with dead code",
			summary: domain.AnalyzeSummary{
				DeadCodeCount: 5,
			},
			want: 95,
		},
		{
			name: "with high complexity count",
			summary: domain.AnalyzeSummary{
				HighComplexityCount: 3,
			},
			want: 95,
		},
		{
			name: "with all issues",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   15.0,
				DeadCodeCount:       5,
				HighComplexityCount: 3,
			},
			want: 80,
		},
		{
			name: "minimum score floor",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   50.0,
				DeadCodeCount:       100,
				HighComplexityCount: 50,
			},
			want: 80, // Only 20 points deducted due to simple fallback logic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.summary.CalculateFallbackScore(); got != tt.want {
				t.Errorf("CalculateFallbackScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzeSummary_HasIssues(t *testing.T) {
	tests := []struct {
		name    string
		summary domain.AnalyzeSummary
		want    bool
	}{
		{
			name:    "no issues",
			summary: domain.AnalyzeSummary{},
			want:    false,
		},
		{
			name: "has high complexity",
			summary: domain.AnalyzeSummary{
				HighComplexityCount: 5,
			},
			want: true,
		},
		{
			name: "has dead code",
			summary: domain.AnalyzeSummary{
				DeadCodeCount: 10,
			},
			want: true,
		},
		{
			name: "has clone pairs",
			summary: domain.AnalyzeSummary{
				ClonePairs: 3,
			},
			want: true,
		},
		{
			name: "has high coupling",
			summary: domain.AnalyzeSummary{
				HighCouplingClasses: 2,
			},
			want: true,
		},
		{
			name: "multiple issues",
			summary: domain.AnalyzeSummary{
				HighComplexityCount: 5,
				DeadCodeCount:       10,
				ClonePairs:          3,
				HighCouplingClasses: 2,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.summary.HasIssues(); got != tt.want {
				t.Errorf("HasIssues() = %v, want %v", got, tt.want)
			}
		})
	}
}
