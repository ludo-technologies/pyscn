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
				AverageComplexity:         7.0,  // -6
				CodeDuplication:           15.0, // -12 (Medium-High range)
				CBOClasses:                10,
				HighCouplingClasses:       2, // 20% ratio: -12 (new: Medium penalty)
				DepsEnabled:               true,
				DepsMainSequenceDeviation: 1.0, // -2
				ArchEnabled:               true,
				ArchCompliance:            0.125, // -7
			},
			expectedScore:             61,  // Updated: 100-6-12-12-2-7 = 61
			expectedGrade:             "C", // 61 is in C range (55-69)
			expectError:               false,
			expectedComplexityScore:   70,  // 100 - (6/20)*100 = 70
			expectedDeadCodeScore:     100, // No dead code
			expectedDuplicationScore:  40,  // 100 - (12/20)*100 = 40
			expectedCouplingScore:     40,  // 100 - (12/20)*100 = 40 (new penalty)
			expectedDependencyScore:   83,  // 100 - (2/12)*100 = 83 (rounded)
			expectedArchitectureScore: 12,  // 100 - (7/8)*100 = 12 (rounded)
		},
		{
			name: "moderate complexity and duplication",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 12.0, // -12
				CodeDuplication:   30.0, // -20 (new: >20% = High penalty)
				ArchEnabled:       false,
				DepsEnabled:       false,
			},
			expectedScore: 68,  // Updated: 100-12-20 = 68
			expectedGrade: "C", // Updated from B
			expectError:   false,
		},
		{
			name: "high complexity",
			summary: domain.AnalyzeSummary{
				AverageComplexity: 25.0, // -20
				CodeDuplication:   5.0,  // -6 (new: 3-10% = Low penalty)
			},
			expectedScore: 74,  // Updated: 100-20-6 = 74
			expectedGrade: "B", // 74 is in B range (70-84)
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
				AverageComplexity:   25.0, // -20
				CodeDuplication:     50.0, // -20
				CBOClasses:          10,
				HighCouplingClasses: 6, // 60% ratio: -20 (new: High penalty)
				DeadCodeCount:       100,
				CriticalDeadCode:    50, // -20 (capped)
				DepsEnabled:         true,
				DepsTotalModules:    10,
				DepsModulesInCycles: 10, // -8
				ArchEnabled:         true,
				ArchCompliance:      0.0, // -8
			},
			expectedScore: 10, // Floor at 10 (actual: 100-20-20-20-20-8-8 = 4, floored to 10)
			expectedGrade: "F",
			expectError:   false,
		},
		{
			name: "grade A threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   4.0, // -0 (low complexity)
				CodeDuplication:     2.0, // -0 (below 3% threshold)
				CBOClasses:          20,
				HighCouplingClasses: 2, // 10% ratio: -6 (new: Low penalty)
				DepsEnabled:         true,
				DepsTotalModules:    10,
				DepsMaxDepth:        4, // Expected ~4, so no penalty
				ArchEnabled:         true,
				ArchCompliance:      0.9, // -1 (rounded)
			},
			expectedScore: 93, // Updated: 100-0-0-6-0-1 = 93
			expectedGrade: "A",
			expectError:   false,
		},
		{
			name: "grade C threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   15.0, // -12
				CodeDuplication:     25.0, // -20 (>20% = High penalty)
				CBOClasses:          20,
				HighCouplingClasses: 2, // 10% ratio: -6 (new: Low penalty, 10% > 5%)
				DeadCodeCount:       5,
				CriticalDeadCode:    0, // No critical issues, so no dead code penalty
				TotalFiles:          1,
			},
			expectedScore: 62,  // Updated: 100-12-20-6 = 62
			expectedGrade: "C", // 62 is in C range (55-69)
			expectError:   false,
		},
		{
			name: "grade D threshold",
			summary: domain.AnalyzeSummary{
				AverageComplexity:   22.0, // -20
				CodeDuplication:     45.0, // -20
				CBOClasses:          10,
				HighCouplingClasses: 6, // 60% ratio: -20 (new: High penalty)
			},
			expectedScore: 40,  // Updated: 100-20-20-20 = 40
			expectedGrade: "D", // 40 is exact D threshold
			expectError:   false,
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
		{"grade A boundary", 85, "A"},
		{"grade A above boundary", 90, "A"},
		{"grade B boundary", 70, "B"},
		{"grade B above boundary", 75, "B"},
		{"grade B below A", 84, "B"},
		{"grade C boundary", 55, "C"},
		{"grade C above boundary", 60, "C"},
		{"grade C below B", 69, "C"},
		{"grade D boundary", 40, "D"},
		{"grade D above boundary", 45, "D"},
		{"grade D below C", 54, "D"},
		{"grade F high", 39, "F"},
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
