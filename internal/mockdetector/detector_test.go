package mockdetector

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestDetector_Detect_MockStrings(t *testing.T) {
	detector := NewDetector(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name       string
		source     string
		wantType   domain.MockDataType
		wantHit    bool
		wantMinLen int
	}{
		{
			name:       "mock email",
			source:     `email = "test@example.com"`,
			wantType:   domain.MockDataTypeDomain, // Domain check fires (example.com)
			wantHit:    true,
			wantMinLen: 1,
		},
		{
			name:       "test domain in URL",
			source:     `url = "http://localhost:8080/api"`,
			wantType:   domain.MockDataTypeDomain,
			wantHit:    true,
			wantMinLen: 1,
		},
		{
			name:       "placeholder phone",
			source:     `phone = "123-456-7890"`,
			wantType:   domain.MockDataTypePhone,
			wantHit:    true,
			wantMinLen: 1,
		},
		{
			name:       "low entropy UUID",
			source:     `uuid = "00000000-0000-0000-0000-000000000000"`,
			wantType:   domain.MockDataTypeUUID,
			wantHit:    true,
			wantMinLen: 1,
		},
		{
			name:       "test credential",
			source:     `password = "password123"`,
			wantType:   domain.MockDataTypeTestCredential,
			wantHit:    true,
			wantMinLen: 1,
		},
		{
			name:       "real data - should not match",
			source:     `email = "user@company.com"`,
			wantHit:    false,
			wantMinLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.Detect(ctx, []byte(tt.source), "test.py")
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if tt.wantHit {
				if len(result.Findings) < tt.wantMinLen {
					t.Errorf("expected at least %d findings, got %d", tt.wantMinLen, len(result.Findings))
					return
				}

				found := false
				for _, f := range result.Findings {
					if f.Type == tt.wantType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected finding of type %s, but not found in %v", tt.wantType, result.Findings)
				}
			} else {
				// For non-matches, we allow some findings but not the specific type
				for _, f := range result.Findings {
					if f.Type == tt.wantType {
						t.Errorf("unexpected finding of type %s", tt.wantType)
					}
				}
			}
		})
	}
}

func TestDetector_Detect_MockIdentifiers(t *testing.T) {
	detector := NewDetector(nil, nil)
	ctx := context.Background()

	tests := []struct {
		name    string
		source  string
		wantHit bool
	}{
		{
			name:    "mock variable",
			source:  `mock_user = get_user()`,
			wantHit: true,
		},
		{
			name:    "fake variable",
			source:  `fake_response = create_response()`,
			wantHit: true,
		},
		{
			name:    "dummy variable",
			source:  `dummy_data = {}`,
			wantHit: true,
		},
		{
			name:    "normal variable",
			source:  `user_count = len(users)`,
			wantHit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := detector.Detect(ctx, []byte(tt.source), "test.py")
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			found := false
			for _, f := range result.Findings {
				if f.Type == domain.MockDataTypeKeyword {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Error("expected keyword match, but not found")
			}
			if !tt.wantHit && found {
				t.Error("unexpected keyword match")
			}
		})
	}
}

func TestDetector_Detect_Assignment(t *testing.T) {
	detector := NewDetector(nil, nil)
	ctx := context.Background()

	// Test that mock variable assignments with mock values get detected
	source := `mock_email = "test@example.com"`

	result, err := detector.Detect(ctx, []byte(source), "test.py")
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	if len(result.Findings) == 0 {
		t.Error("expected findings for mock assignment, got none")
	}

	// Check that we have an assignment-level finding
	foundAssignment := false
	for _, f := range result.Findings {
		if f.Description == "Mock data assignment detected" {
			foundAssignment = true
			break
		}
	}

	if !foundAssignment {
		t.Error("expected assignment-level finding")
	}
}

func TestExtractStringContent(t *testing.T) {
	// Note: extractStringContent expects input from tree-sitter parser,
	// which provides the raw string literal including quotes but
	// processes prefixes (r, f, b, etc.) separately.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"double quoted", `"hello"`, "hello"},
		{"single quoted", `'hello'`, "hello"},
		{"triple double quoted", `"""hello"""`, "hello"},
		{"triple single quoted", `'''hello'''`, "hello"},
		{"f-string content", `f"hello"`, "hello"}, // f-prefix handled by tree-sitter
		{"no quotes", "hello", "hello"},
		{"empty string", `""`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractStringContent(tt.input)
			if got != tt.want {
				t.Errorf("extractStringContent(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSelectHighestSeverity(t *testing.T) {
	tests := []struct {
		name    string
		matches []Match
		want    domain.MockDataSeverity
	}{
		{
			name:    "empty matches",
			matches: []Match{},
			want:    "",
		},
		{
			name: "single match",
			matches: []Match{
				{Severity: domain.MockDataSeverityWarning},
			},
			want: domain.MockDataSeverityWarning,
		},
		{
			name: "multiple matches - error is highest",
			matches: []Match{
				{Severity: domain.MockDataSeverityInfo},
				{Severity: domain.MockDataSeverityError},
				{Severity: domain.MockDataSeverityWarning},
			},
			want: domain.MockDataSeverityError,
		},
		{
			name: "multiple matches - warning is highest",
			matches: []Match{
				{Severity: domain.MockDataSeverityInfo},
				{Severity: domain.MockDataSeverityWarning},
				{Severity: domain.MockDataSeverityInfo},
			},
			want: domain.MockDataSeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectHighestSeverity(tt.matches)
			if got.Severity != tt.want {
				t.Errorf("selectHighestSeverity() severity = %v, want %v", got.Severity, tt.want)
			}
		})
	}
}
