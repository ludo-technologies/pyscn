package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeLongFunctionFile writes a module holding one flat function of the given
// body length (McCabe 1) plus a short one, and returns its path.
func writeLongFunctionFile(t *testing.T, bodyLines int) string {
	t.Helper()

	var source strings.Builder
	source.WriteString("def build_table():\n    rows = []\n")
	for i := 0; i < bodyLines; i++ {
		fmt.Fprintf(&source, "    rows.append(%d)\n", i)
	}
	source.WriteString("    return rows\n\n\ndef short():\n    return 1\n")

	path := filepath.Join(t.TempDir(), "long_function.py")
	if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return path
}

func TestCheckComplexityReportsLongFunctions(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)

	// 120 straight-line statements: far past the critical threshold of 100,
	// while McCabe stays at 1 and never trips the complexity gate.
	path := writeLongFunctionFile(t, 120)

	cobraCmd.SetArgs([]string{"--select", "complexity", path})
	if err := cobraCmd.Execute(); err == nil {
		t.Fatal("expected the long function to fail the quality gate")
	}

	output := stderr.String()
	// def + rows = [] + 120 appends + return
	if !strings.Contains(output, "build_table is too long (123 SLOC > 100)") {
		t.Errorf("expected a long-function diagnostic, got: %s", output)
	}
	if !strings.Contains(output, path+":1:1:") {
		t.Errorf("expected the diagnostic to carry file:line:col, got: %s", output)
	}
	if strings.Contains(output, "short is too long") {
		t.Errorf("short function must not be reported, got: %s", output)
	}
	if strings.Contains(output, "<module> is too long") {
		t.Errorf("module scope must not be reported, got: %s", output)
	}
}

func TestCheckComplexityStaysSilentBelowThreshold(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)

	// 60 statements: long enough to warn in the report, below the check gate.
	path := writeLongFunctionFile(t, 60)

	cobraCmd.SetArgs([]string{"--select", "complexity", "--quiet", path})
	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("expected no issues below the critical threshold, got %v", err)
	}
	if output := stderr.String(); output != "" {
		t.Errorf("expected no diagnostics, got: %s", output)
	}
}

func TestCheckComplexityReportsClassExecutionScope(t *testing.T) {
	var source strings.Builder
	source.WriteString("class Config:\n")
	for i := 0; i < 11; i++ {
		fmt.Fprintf(&source, "    if flag_%d:\n        value = %d\n", i, i)
	}

	path := filepath.Join(t.TempDir(), "config.py")
	if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)

	cobraCmd.SetArgs([]string{"--select", "complexity", path})
	if err := cobraCmd.Execute(); err == nil {
		t.Fatalf("expected the class-scope issue to fail the gate: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "class scope Config is too complex (12 > 10)") {
		t.Fatalf("expected a class-scope diagnostic, got: %s", stderr.String())
	}
}

func TestCheckComplexityGateIgnoresReportFilters(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "filtered.py")
	source := `def filtered_from_report(value):
    if value > 0: value += 1
    if value > 1: value += 1
    if value > 2: value += 1
    if value > 3: value += 1
    if value > 4: value += 1
    if value > 5: value += 1
    if value > 6: value += 1
    if value > 7: value += 1
    if value > 8: value += 1
    if value > 9: value += 1
    return value
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}

	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte("[complexity]\nmin_complexity = 12\n"), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	checkCmd := NewCheckCommand()
	checkCmd.configFile = configPath
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)

	cobraCmd.SetArgs([]string{"--select", "complexity", path})
	if err := cobraCmd.Execute(); err == nil {
		t.Fatalf("expected the filtered CC-11 scope to fail the CC-10 gate: %s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "filtered_from_report is too complex (11 > 10)") {
		t.Fatalf("expected the complete analyzed population to drive the gate, got: %s", stderr.String())
	}
}

func TestCheckComplexityGatesFunctionsHiddenByDisplayFilters(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	var stderr bytes.Buffer
	cobraCmd.SetErr(&stderr)

	path := writeLongFunctionFile(t, 120)
	configPath := filepath.Join(t.TempDir(), ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte("[complexity]\nreport_unchanged = false\n"), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	checkCmd.configFile = configPath

	cobraCmd.SetArgs([]string{"--select", "complexity", path})
	if err := cobraCmd.Execute(); err == nil {
		t.Fatal("expected the hidden long function to fail the gate")
	}
	if output := stderr.String(); !strings.Contains(output, "build_table is too long (123 SLOC > 100)") {
		t.Errorf("expected a long-function diagnostic, got: %s", output)
	}
}
