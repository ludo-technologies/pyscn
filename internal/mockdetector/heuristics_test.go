package mockdetector

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestHeuristics_CheckString_Keywords(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name     string
		input    string
		wantType domain.MockDataType
		wantHit  bool
	}{
		{"mock keyword", "mock_user", domain.MockDataTypeKeyword, true},
		{"fake keyword", "fake_data", domain.MockDataTypeKeyword, true},
		{"dummy keyword", "dummy_value", domain.MockDataTypeKeyword, true},
		{"test keyword", "test_data", domain.MockDataTypeKeyword, true},
		{"example keyword", "example_name", domain.MockDataTypeKeyword, true},
		{"foo keyword", "foo", domain.MockDataTypeKeyword, true},
		{"bar keyword", "bar", domain.MockDataTypeKeyword, true},
		{"lorem keyword", "lorem ipsum", domain.MockDataTypeKeyword, true},
		{"normal string", "production_data", domain.MockDataTypeKeyword, false},
		{"empty string", "", domain.MockDataTypeKeyword, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			if tt.wantHit {
				if len(matches) == 0 {
					t.Errorf("expected matches for %q, got none", tt.input)
					return
				}
				found := false
				for _, m := range matches {
					if m.Type == tt.wantType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected type %s for %q, got different types", tt.wantType, tt.input)
				}
			} else {
				for _, m := range matches {
					if m.Type == tt.wantType {
						t.Errorf("did not expect type %s for %q, but got it", tt.wantType, tt.input)
					}
				}
			}
		})
	}
}

func TestHeuristics_CheckString_Domains(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"example.com", "user@example.com", true},
		{"test.com", "http://test.com/api", true},
		{"localhost", "http://localhost:8080", true},
		{"invalid domain", "http://invalid/path", true},
		{"real domain", "user@gmail.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			found := false
			for _, m := range matches {
				if m.Type == domain.MockDataTypeDomain || m.Type == domain.MockDataTypeEmail {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Errorf("expected domain match for %q, got none", tt.input)
			} else if !tt.wantHit && found {
				t.Errorf("did not expect domain match for %q, but got one", tt.input)
			}
		})
	}
}

func TestHeuristics_CheckString_Email(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"example email", "test@example.com", true},
		{"test email", "user@test.org", true},
		{"localhost email", "admin@localhost", true},
		{"real email", "user@gmail.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			found := false
			for _, m := range matches {
				if m.Type == domain.MockDataTypeEmail {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Errorf("expected email match for %q, got none", tt.input)
			} else if !tt.wantHit && found {
				t.Errorf("did not expect email match for %q, but got one", tt.input)
			}
		})
	}
}

func TestHeuristics_CheckString_Phone(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"all zeros", "000-0000-0000", true},
		{"sequential", "123-456-7890", true},
		{"real phone", "555-867-5309", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			found := false
			for _, m := range matches {
				if m.Type == domain.MockDataTypePhone {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Errorf("expected phone match for %q, got none", tt.input)
			} else if !tt.wantHit && found {
				t.Errorf("did not expect phone match for %q, but got one", tt.input)
			}
		})
	}
}

func TestHeuristics_CheckString_UUID(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"all zeros uuid", "00000000-0000-0000-0000-000000000000", true},
		{"all ones uuid", "11111111-1111-1111-1111-111111111111", true},
		{"random uuid", "550e8400-e29b-41d4-a716-446655440000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			found := false
			for _, m := range matches {
				if m.Type == domain.MockDataTypeUUID {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Errorf("expected UUID match for %q, got none", tt.input)
			} else if !tt.wantHit && found {
				t.Errorf("did not expect UUID match for %q, but got one", tt.input)
			}
		})
	}
}

func TestHeuristics_CheckString_Credential(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"password123", "password123", true},
		{"secret123", "secret123", true},
		{"testpassword", "testpassword", true},
		{"api_key", "api_key", true},
		{"real password", "xK9#mP2$vL8@nQ4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckString(tt.input)

			found := false
			for _, m := range matches {
				if m.Type == domain.MockDataTypeTestCredential {
					found = true
					break
				}
			}

			if tt.wantHit && !found {
				t.Errorf("expected credential match for %q, got none", tt.input)
			} else if !tt.wantHit && found {
				t.Errorf("did not expect credential match for %q, but got one", tt.input)
			}
		})
	}
}

func TestHeuristics_CheckIdentifier(t *testing.T) {
	h := NewHeuristics(nil, nil)

	tests := []struct {
		name    string
		input   string
		wantHit bool
	}{
		{"mock variable", "mock_user", true},
		{"fake variable", "fake_response", true},
		{"test variable", "test_data", true},
		{"normal variable", "user_count", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := h.CheckIdentifier(tt.input)

			if tt.wantHit && len(matches) == 0 {
				t.Errorf("expected identifier match for %q, got none", tt.input)
			} else if !tt.wantHit && len(matches) > 0 {
				t.Errorf("did not expect identifier match for %q, but got one", tt.input)
			}
		})
	}
}

func TestNewHeuristics_CustomKeywords(t *testing.T) {
	customKeywords := []string{"custom", "special"}
	h := NewHeuristics(customKeywords, nil)

	matches := h.CheckString("custom_value")
	if len(matches) == 0 {
		t.Error("expected match for custom keyword")
	}

	matches = h.CheckString("mock_value")
	if len(matches) > 0 {
		t.Error("did not expect match for non-custom keyword")
	}
}

func TestNewHeuristics_CustomDomains(t *testing.T) {
	customDomains := []string{"custom.test"}
	h := NewHeuristics(nil, customDomains)

	matches := h.CheckString("http://custom.test/api")
	if len(matches) == 0 {
		t.Error("expected match for custom domain")
	}
}

func TestContainsWordBoundary(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		keyword string
		want    bool
	}{
		// True positives - should match
		{"standalone word", "test", "test", true},
		{"prefix with underscore", "test_user", "test", true},
		{"suffix with underscore", "user_test", "test", true},
		{"prefix with hyphen", "test-user", "test", true},
		{"suffix with hyphen", "user-test", "test", true},
		{"middle with underscores", "my_test_data", "test", true},
		{"space separated", "my test data", "test", true},
		{"dot separated", "test.py", "test", true},

		// False positives - should NOT match
		{"contest", "contest", "test", false},
		{"attest", "attest", "test", false},
		{"latest", "latest", "test", false},
		{"testify", "testify", "test", false},
		{"protest", "protest", "test", false},
		{"detestable", "detestable", "test", false},

		// Edge cases
		{"empty string", "", "test", false},
		{"keyword only at end", "mytest", "test", false},
		{"keyword only at start", "testing", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsWordBoundary(tt.s, tt.keyword)
			if got != tt.want {
				t.Errorf("containsWordBoundary(%q, %q) = %v, want %v", tt.s, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestHeuristics_CheckString_Keywords_WordBoundary(t *testing.T) {
	h := NewHeuristics(nil, nil)

	// Should NOT match - false positives prevented by word boundary
	falsePositives := []string{
		"contest",
		"attest",
		"latest",
		"testify",
		"protest",
	}

	for _, fp := range falsePositives {
		matches := h.CheckString(fp)
		for _, m := range matches {
			if m.Type == domain.MockDataTypeKeyword {
				t.Errorf("should not match keyword in %q, but got match: %s", fp, m.Rationale)
			}
		}
	}

	// Should match - true positives with proper word boundaries
	truePositives := []string{
		"test_user",
		"user_test",
		"my_test_data",
		"test-data",
	}

	for _, tp := range truePositives {
		matches := h.CheckString(tp)
		found := false
		for _, m := range matches {
			if m.Type == domain.MockDataTypeKeyword {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("should match keyword in %q, but got no match", tp)
		}
	}
}
