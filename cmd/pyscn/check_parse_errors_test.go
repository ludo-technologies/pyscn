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

// Clone detection failing on its own stays informational, but an unparseable
// file is a problem with the input and must fail the gate like everywhere else.
func TestCheckSelectClonesFailsOnUnparseableFile(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "clones", writeMixedSyntaxPackage(t)})

	err := cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected --select clones to fail on a file that cannot be parsed")
	}
	if got := exitCodeFor(err); got != exitCodeAnalysisError {
		t.Errorf("expected exit code %d (analysis error), got %d", exitCodeAnalysisError, got)
	}
	if !strings.Contains(stderr.String(), "broken.py") {
		t.Errorf("expected the unparseable file to be named, got: %q", stderr.String())
	}
}

func TestCheckSelectDependenciesFailsOnUnparseableFile(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "deps", writeMixedSyntaxPackage(t)})

	err := cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected --select deps to fail on a file that cannot be parsed")
	}
	if got := exitCodeFor(err); got != exitCodeAnalysisError {
		t.Errorf("expected exit code %d (analysis error), got %d", exitCodeAnalysisError, got)
	}
	if !strings.Contains(stderr.String(), "broken.py") {
		t.Errorf("expected the unparseable file to be named, got: %q", stderr.String())
	}
}

func TestCheckSelectDIFailsOnUnparseableFile(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "di", writeMixedSyntaxPackage(t)})

	err := cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected --select di to fail on a file that cannot be parsed")
	}
	if got := exitCodeFor(err); got != exitCodeAnalysisError {
		t.Errorf("expected exit code %d (analysis error), got %d", exitCodeAnalysisError, got)
	}
	if !strings.Contains(stderr.String(), "broken.py") {
		t.Errorf("expected the unparseable file to be named, got: %q", stderr.String())
	}
}

func TestCheckSelectDependenciesOverridesDisabledConfig(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "..", "testdata", "python", "circular_deps_test"))
	if err != nil {
		t.Fatalf("resolve circular dependency fixture: %v", err)
	}
	configPath := filepath.Join(t.TempDir(), ".pyscn.toml")
	config := "[system_analysis]\nenabled = false\n\n[dependencies]\nenabled = false\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--config", configPath, "--select", "deps", fixtureRoot})

	err = cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected the explicitly selected dependency gate to report the fixture cycle")
	}
	if got := exitCodeFor(err); got != exitCodeQualityIssues {
		t.Fatalf("expected quality exit code %d, got %d: %s", exitCodeQualityIssues, got, stderr.String())
	}
	if !strings.Contains(stderr.String(), "circular dependency detected") {
		t.Fatalf("expected circular dependency diagnostic, got: %s", stderr.String())
	}
}

// An unusable invocation is invalid input, which the documented contract puts
// at exit 2 — not exit 1, which would read as a quality failure.
func TestCheckInvalidInvocationExitsAsAnalysisError(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown flag", args: []string{"--no-such-flag", dir}},
		{name: "invalid select value", args: []string{"--select", "nonsense", dir}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkCmd := NewCheckCommand()
			cobraCmd := checkCmd.CreateCobraCommand()
			cobraCmd.SetErr(&bytes.Buffer{})
			cobraCmd.SetOut(&bytes.Buffer{})
			cobraCmd.SetArgs(tt.args)

			err := cobraCmd.Execute()
			if err == nil {
				t.Fatal("expected an unusable invocation to fail")
			}
			if got := exitCodeFor(err); got != exitCodeAnalysisError {
				t.Errorf("expected exit code %d (analysis error), got %d", exitCodeAnalysisError, got)
			}
		})
	}
}
