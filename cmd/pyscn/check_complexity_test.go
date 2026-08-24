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
