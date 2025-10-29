package version_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/version"
)

func TestShort(t *testing.T) {
	result := version.Short()

	if result == "" {
		t.Error("Short() should return non-empty string")
	}
}

func TestInfo(t *testing.T) {

	info := version.Info()

	// Verify version info contains expected components
	if !strings.Contains(info, "pyscn") {
		t.Error("Info() should contain 'pyscn'")
	}

	// Verify Go version is included
	if !strings.Contains(info, runtime.Version()) {
		t.Errorf("Info() should contain Go version %s", runtime.Version())
	}

	// Verify OS/Arch is included
	expectedArch := runtime.GOOS + "/" + runtime.GOARCH
	if !strings.Contains(info, expectedArch) {
		t.Errorf("Info() should contain OS/Arch %s", expectedArch)
	}

	// Verify format contains expected fields
	requiredFields := []string{"Commit:", "Built:", "Go:", "OS/Arch:"}
	for _, field := range requiredFields {
		if !strings.Contains(info, field) {
			t.Errorf("Info() should contain %s field", field)
		}
	}

}

func TestInfoFormat(t *testing.T) {
	info := version.Info()
	lines := strings.Split(info, "\n")

	// Verify the number of lines in the returned output
	if len(lines) < 5 {
		t.Errorf("Info() should contain 5 lines, got %d", len(lines))
	}

	// Verify each line starts with the expected prefix
	expectedPrefixes := []string{"pyscn ", "Commit:", "Built:", "Go:", "OS/Arch:"}

	for i, prefix := range expectedPrefixes {
		if !strings.HasPrefix(lines[i], prefix) {
			t.Errorf("line %d should start with %q, got %q", i+1, prefix, lines[i])
		}
	}
}

func TestInfoIncludesBuildMetadata(t *testing.T) {
	info := version.Info()

	// Verify all expected metadata values are included in the output
	metadataFields := map[string]string{
		"pyscn":  version.Version,
		"Commit": version.Commit,
		"Built":  version.Date,
	}

	for name, val := range metadataFields {
		if val == "" {
			t.Fatalf("%s should not be empty", name)
		}

		var expected string
		if name == "pyscn" {
			expected = fmt.Sprintf("%s %s", name, val)
		} else {
			expected = fmt.Sprintf("%s: %s", name, val)
		}

		if !strings.Contains(info, expected) {
			t.Errorf("Info() output missing %q", expected)
		}
	}
}
