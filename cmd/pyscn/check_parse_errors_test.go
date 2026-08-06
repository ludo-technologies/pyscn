package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var errQualityIssuesFixture = errors.New("found 3 quality issue(s)")

// writeMixedSyntaxPackage writes a package holding one clean module and one
// module CPython refuses to parse, and returns its directory.
func writeMixedSyntaxPackage(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	good := "def tidy(n):\n    if n > 0:\n        return n\n    return -n\n"
	// `return 0 objs` is a hard SyntaxError, so the whole module is dropped.
	broken := "def loader(objs):\n    total = 0\n    for o in objs:\n        total += o\n    return 0 objs\n"

	for name, source := range map[string]string{"good.py": good, "broken.py": broken} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(source), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

// A file that fails to parse contributes no functions, so it clears every
// threshold by contributing nothing. The gate must fail on it instead of
// reporting a pass (issue #690).
func TestCheckFailsOnUnparseableFile(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "complexity", writeMixedSyntaxPackage(t)})

	err := cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected the check to fail on a file that cannot be parsed")
	}
	if got := exitCodeFor(err); got != exitCodeAnalysisError {
		t.Errorf("expected exit code %d (analysis error), got %d", exitCodeAnalysisError, got)
	}
	if !strings.Contains(stderr.String(), "broken.py") {
		t.Errorf("expected the unparseable file to be named, got: %q", stderr.String())
	}
	if strings.Contains(stderr.String(), "Code quality check passed") {
		t.Errorf("unparseable input must not report a pass, got: %q", stderr.String())
	}
}

func TestCheckAllowParseErrorsWaivesTheGate(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "complexity", "--allow-parse-errors", writeMixedSyntaxPackage(t)})

	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("--allow-parse-errors should waive the failure, got: %v", err)
	}
	if !strings.Contains(stderr.String(), "broken.py") {
		t.Errorf("waived parse errors must still be reported, got: %q", stderr.String())
	}
}

func TestCheckPassesOnFullyParseablePackage(t *testing.T) {
	dir := t.TempDir()
	source := "def tidy(n):\n    if n > 0:\n        return n\n    return -n\n"
	if err := os.WriteFile(filepath.Join(dir, "good.py"), []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "complexity", dir})

	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("clean input should pass, got: %v", err)
	}
}

func TestExitCodeForQualityIssues(t *testing.T) {
	if got := exitCodeFor(errQualityIssuesFixture); got != exitCodeQualityIssues {
		t.Errorf("expected exit code %d for a quality verdict, got %d", exitCodeQualityIssues, got)
	}
}
