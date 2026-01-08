// Package mockdetector provides heuristics-based detection of mock data in Python code.
package mockdetector

import (
	"regexp"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// Heuristics contains the detection rules and patterns for mock data.
type Heuristics struct {
	// Keywords that indicate mock data in identifiers and strings
	Keywords []string

	// Domains that indicate mock/test data
	Domains []string

	// Compiled regex patterns for efficient matching
	emailPattern     *regexp.Regexp
	phonePattern     *regexp.Regexp
	uuidPattern      *regexp.Regexp
	repetitivePattern *regexp.Regexp
	placeholderPattern *regexp.Regexp
	credentialPattern *regexp.Regexp
}

// NewHeuristics creates a new Heuristics instance with default or custom patterns.
func NewHeuristics(keywords, domains []string) *Heuristics {
	if len(keywords) == 0 {
		keywords = domain.DefaultMockDataKeywords()
	}
	if len(domains) == 0 {
		domains = domain.DefaultMockDataDomains()
	}

	return &Heuristics{
		Keywords:           keywords,
		Domains:            domains,
		emailPattern:       compileEmailPattern(),
		phonePattern:       compilePhonePattern(),
		uuidPattern:        compileUUIDPattern(),
		repetitivePattern:  compileRepetitivePattern(),
		placeholderPattern: compilePlaceholderPattern(),
		credentialPattern:  compileCredentialPattern(),
	}
}

// compileEmailPattern creates a regex for detecting test/mock email addresses.
func compileEmailPattern() *regexp.Regexp {
	return regexp.MustCompile(`(?i)[a-z0-9._%+-]+@(example\.(com|org|net)|test\.(com|org|net)|localhost|invalid|foo\.com|bar\.com)`)
}

// compilePhonePattern creates a regex for detecting placeholder phone numbers.
// Matches patterns like: 000-0000-0000, 123-456-7890, (000) 000-0000
func compilePhonePattern() *regexp.Regexp {
	return regexp.MustCompile(`(?:^|[^0-9])(\(?(?:0{3}|1{3}|123)\)?[-.\s]?(?:0{3,4}|1{3,4}|456)[-.\s]?(?:0{4}|1{4}|7890))(?:[^0-9]|$)`)
}

// compileUUIDPattern creates a regex for detecting low-entropy UUIDs.
// Matches UUIDs with repeated characters like 00000000-0000-0000-0000-000000000000
func compileUUIDPattern() *regexp.Regexp {
	return regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
}

// compileRepetitivePattern returns nil as Go's regexp doesn't support backreferences.
// We'll use a custom function instead.
func compileRepetitivePattern() *regexp.Regexp {
	return nil // Handled by custom function isRepetitive
}

// compilePlaceholderPattern creates a regex for detecting placeholder comments.
func compilePlaceholderPattern() *regexp.Regexp {
	return regexp.MustCompile(`(?i)\b(TODO|FIXME|XXX|HACK|BUG|NOTE)[\s:]+.*(?:replace|change|update|mock|fake|dummy|placeholder|temporary|temp)`)
}

// compileCredentialPattern creates a regex for detecting test credentials.
func compileCredentialPattern() *regexp.Regexp {
	return regexp.MustCompile(`(?i)^(password|secret|api[_-]?key|token|credential|auth)[0-9]*$|^(password|secret)123$|^test(password|secret|key)$`)
}

// Match represents a detection match result.
type Match struct {
	Value       string
	Type        domain.MockDataType
	Severity    domain.MockDataSeverity
	Description string
	Rationale   string
	StartIndex  int
	EndIndex    int
}

// CheckString checks a string value for mock data patterns.
func (h *Heuristics) CheckString(value string) []Match {
	var matches []Match

	// Check for keyword matches in the string
	if match := h.checkKeywords(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for domain matches
	if match := h.checkDomains(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for email patterns
	if match := h.checkEmail(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for phone patterns
	if match := h.checkPhone(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for low-entropy UUIDs
	if match := h.checkUUID(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for repetitive patterns
	if match := h.checkRepetitive(value); match != nil {
		matches = append(matches, *match)
	}

	// Check for test credentials
	if match := h.checkCredential(value); match != nil {
		matches = append(matches, *match)
	}

	return matches
}

// CheckIdentifier checks an identifier (variable/function name) for mock data patterns.
func (h *Heuristics) CheckIdentifier(name string) []Match {
	var matches []Match

	lower := strings.ToLower(name)
	for _, keyword := range h.Keywords {
		if containsWordBoundary(lower, keyword) {
			matches = append(matches, Match{
				Value:       name,
				Type:        domain.MockDataTypeKeyword,
				Severity:    domain.MockDataSeverityWarning,
				Description: "Identifier contains mock data keyword",
				Rationale:   "Contains keyword: " + keyword,
			})
			break
		}
	}

	return matches
}

// checkKeywords checks for keyword patterns in a string.
func (h *Heuristics) checkKeywords(value string) *Match {
	lower := strings.ToLower(value)
	for _, keyword := range h.Keywords {
		if containsWordBoundary(lower, keyword) {
			return &Match{
				Value:       value,
				Type:        domain.MockDataTypeKeyword,
				Severity:    domain.MockDataSeverityInfo,
				Description: "String contains mock data keyword",
				Rationale:   "Contains keyword: " + keyword,
			}
		}
	}
	return nil
}

// containsWordBoundary checks if the keyword appears in the string at word boundaries.
// A word boundary is defined as: start/end of string, or non-alphanumeric characters like '_', '-', '.', etc.
// This prevents false positives like "contest" matching "test".
func containsWordBoundary(s, keyword string) bool {
	idx := strings.Index(s, keyword)
	for idx != -1 {
		// Check left boundary
		leftOK := idx == 0 || !isAlphaNum(s[idx-1])
		// Check right boundary
		endIdx := idx + len(keyword)
		rightOK := endIdx == len(s) || !isAlphaNum(s[endIdx])

		if leftOK && rightOK {
			return true
		}

		// Search for next occurrence
		if endIdx < len(s) {
			nextIdx := strings.Index(s[endIdx:], keyword)
			if nextIdx == -1 {
				break
			}
			idx = endIdx + nextIdx
		} else {
			break
		}
	}
	return false
}

// isAlphaNum returns true if the byte is alphanumeric (a-z, A-Z, 0-9).
func isAlphaNum(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// checkDomains checks for test/mock domain patterns.
func (h *Heuristics) checkDomains(value string) *Match {
	lower := strings.ToLower(value)
	for _, d := range h.Domains {
		if strings.Contains(lower, d) {
			return &Match{
				Value:       value,
				Type:        domain.MockDataTypeDomain,
				Severity:    domain.MockDataSeverityWarning,
				Description: "String contains test/mock domain",
				Rationale:   "Contains domain: " + d,
			}
		}
	}
	return nil
}

// checkEmail checks for test/mock email addresses.
func (h *Heuristics) checkEmail(value string) *Match {
	if h.emailPattern.MatchString(value) {
		return &Match{
			Value:       value,
			Type:        domain.MockDataTypeEmail,
			Severity:    domain.MockDataSeverityWarning,
			Description: "Mock email address detected",
			Rationale:   "Email uses test/example domain",
		}
	}
	return nil
}

// checkPhone checks for placeholder phone numbers.
func (h *Heuristics) checkPhone(value string) *Match {
	if h.phonePattern.MatchString(value) {
		return &Match{
			Value:       value,
			Type:        domain.MockDataTypePhone,
			Severity:    domain.MockDataSeverityWarning,
			Description: "Placeholder phone number detected",
			Rationale:   "Phone number has repetitive or sequential pattern",
		}
	}
	return nil
}

// checkUUID checks for low-entropy UUIDs.
func (h *Heuristics) checkUUID(value string) *Match {
	if !h.uuidPattern.MatchString(value) {
		return nil
	}

	// Check for low entropy (repeated characters)
	cleaned := strings.ReplaceAll(value, "-", "")
	if isLowEntropyUUID(cleaned) {
		return &Match{
			Value:       value,
			Type:        domain.MockDataTypeUUID,
			Severity:    domain.MockDataSeverityWarning,
			Description: "Low-entropy UUID detected",
			Rationale:   "UUID has low randomness (placeholder pattern)",
		}
	}
	return nil
}

// isLowEntropyUUID checks if a UUID string (without dashes) has low entropy.
func isLowEntropyUUID(cleaned string) bool {
	// Check for all zeros, all ones, or sequential patterns
	allSame := true
	firstChar := cleaned[0]
	for _, c := range cleaned[1:] {
		if byte(c) != firstChar {
			allSame = false
			break
		}
	}
	if allSame {
		return true
	}

	// Check for simple sequential pattern
	uniqueChars := make(map[rune]int)
	for _, c := range cleaned {
		uniqueChars[c]++
	}

	// If only 1-2 unique characters, it's low entropy
	if len(uniqueChars) <= 2 {
		return true
	}

	// Check for common placeholder UUIDs
	placeholderPatterns := []string{
		"00000000000000000000000000000000",
		"11111111111111111111111111111111",
		"ffffffffffffffffffffffffffffffff",
		"12345678123456781234567812345678",
	}
	for _, pattern := range placeholderPatterns {
		if strings.EqualFold(cleaned, pattern) {
			return true
		}
	}

	return false
}

// checkRepetitive checks for repetitive character patterns.
func (h *Heuristics) checkRepetitive(value string) *Match {
	// Only check short strings that could be placeholder values
	if len(value) < 4 || len(value) > 20 {
		return nil
	}

	if isRepetitive(value) {
		return &Match{
			Value:       value,
			Type:        domain.MockDataTypeRepetitive,
			Severity:    domain.MockDataSeverityInfo,
			Description: "Repetitive pattern detected",
			Rationale:   "String has repetitive characters",
		}
	}
	return nil
}

// isRepetitive checks if a string consists of repetitive characters.
func isRepetitive(s string) bool {
	if len(s) < 4 {
		return false
	}

	// Check for single character repetition (aaaa, 1111)
	firstChar := s[0]
	allSame := true
	for i := 1; i < len(s); i++ {
		if s[i] != firstChar {
			allSame = false
			break
		}
	}
	if allSame {
		return true
	}

	// Check for two-character repetition (abab, 1212)
	if len(s) >= 4 && len(s)%2 == 0 {
		pattern := s[:2]
		isPattern := true
		for i := 2; i < len(s); i += 2 {
			if i+1 < len(s) && (s[i] != pattern[0] || s[i+1] != pattern[1]) {
				isPattern = false
				break
			}
		}
		if isPattern {
			return true
		}
	}

	return false
}

// checkCredential checks for test credentials.
func (h *Heuristics) checkCredential(value string) *Match {
	if h.credentialPattern.MatchString(value) {
		return &Match{
			Value:       value,
			Type:        domain.MockDataTypeTestCredential,
			Severity:    domain.MockDataSeverityError,
			Description: "Test credential detected",
			Rationale:   "Value appears to be a test password or API key",
		}
	}
	return nil
}

// CheckComment checks a comment for placeholder markers.
func (h *Heuristics) CheckComment(comment string) *Match {
	if h.placeholderPattern.MatchString(comment) {
		return &Match{
			Value:       comment,
			Type:        domain.MockDataTypePlaceholder,
			Severity:    domain.MockDataSeverityInfo,
			Description: "Placeholder comment detected",
			Rationale:   "Comment mentions mock/placeholder data",
		}
	}
	return nil
}
