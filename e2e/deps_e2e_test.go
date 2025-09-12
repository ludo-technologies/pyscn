package e2e

import (
    "bytes"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// TestDepsE2EBasic verifies text output for deps command
func TestDepsE2EBasic(t *testing.T) {
    binaryPath := buildPyscnBinary(t)
    defer os.Remove(binaryPath)

    testDir := t.TempDir()
    // Create simple package with one import
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "a.py", "import pkg.b\n")
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "b.py", "# noop\n")

    cmd := exec.Command(binaryPath, "deps", testDir)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
    }
    out := stdout.String()
    if !strings.Contains(out, "Dependency Analysis") || !strings.Contains(out, "Edges:") {
        t.Fatalf("Unexpected deps text output: %s", out)
    }
}

// TestDepsE2EJSONOutput verifies JSON file generation
func TestDepsE2EJSONOutput(t *testing.T) {
    binaryPath := buildPyscnBinary(t)
    defer os.Remove(binaryPath)

    testDir := t.TempDir()
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "a.py", "import pkg.b\n")
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "b.py", "# noop\n")

    outputDir := t.TempDir()
    createTestConfigFile(t, testDir, outputDir)

    cmd := exec.Command(binaryPath, "deps", "--json", testDir)
    cmd.Dir = testDir
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        t.Fatalf("Command failed: %v\nStderr: %s", err, stderr.String())
    }

    files, _ := filepath.Glob(filepath.Join(outputDir, "deps_*.json"))
    if len(files) == 0 {
        list, _ := os.ReadDir(outputDir)
        var names []string
        for _, f := range list { names = append(names, f.Name()) }
        t.Fatalf("No deps JSON file generated in %s, files: %v", outputDir, names)
    }

    content, err := os.ReadFile(files[0])
    if err != nil { t.Fatalf("read json: %v", err) }
    var obj map[string]interface{}
    if err := json.Unmarshal(content, &obj); err != nil {
        t.Fatalf("invalid json: %v\ncontent: %s", err, string(content))
    }
    if _, ok := obj["edges"]; !ok {
        t.Fatalf("json should contain 'edges'")
    }
}

// TestDepsE2EHTMLOutput verifies HTML file generation
func TestDepsE2EHTMLOutput(t *testing.T) {
    binaryPath := buildPyscnBinary(t)
    defer os.Remove(binaryPath)

    testDir := t.TempDir()
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "a.py", "import pkg.b\n")
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "b.py", "# noop\n")

    outputDir := t.TempDir()
    createTestConfigFile(t, testDir, outputDir)

    cmd := exec.Command(binaryPath, "deps", "--html", "--no-open", testDir)
    cmd.Dir = testDir
    if err := cmd.Run(); err != nil { t.Fatalf("deps html failed: %v", err) }
    files, _ := filepath.Glob(filepath.Join(outputDir, "deps_*.html"))
    if len(files) == 0 { t.Fatalf("no deps html generated in %s", outputDir) }
    data, _ := os.ReadFile(files[0])
    if !strings.Contains(string(data), "<html") { t.Fatalf("invalid html content: %s", string(data)) }
}

// TestDepsE2ECSVOutput verifies CSV edges output
func TestDepsE2ECSVOutput(t *testing.T) {
    binaryPath := buildPyscnBinary(t)
    defer os.Remove(binaryPath)

    testDir := t.TempDir()
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "a.py", "import pkg.b\n")
    createTestPythonFile(t, filepath.Join(testDir, "pkg"), "b.py", "# noop\n")
    outputDir := t.TempDir()
    createTestConfigFile(t, testDir, outputDir)

    cmd := exec.Command(binaryPath, "deps", "--csv", testDir)
    cmd.Dir = testDir
    if err := cmd.Run(); err != nil { t.Fatalf("deps csv failed: %v", err) }
    files, _ := filepath.Glob(filepath.Join(outputDir, "deps_*.csv"))
    if len(files) == 0 { t.Fatalf("no deps csv generated in %s", outputDir) }
    data, _ := os.ReadFile(files[0])
    if !strings.HasPrefix(string(data), "from,to") { t.Fatalf("unexpected csv header: %s", string(data)) }
}

// TestDepsE2ELayerViolations verifies layer rule violations are detected
func TestDepsE2ELayerViolations(t *testing.T) {
    binaryPath := buildPyscnBinary(t)
    defer os.Remove(binaryPath)

    testDir := t.TempDir()
    // presentation -> domain (violation)
    createTestPythonFile(t, filepath.Join(testDir, "pkg", "presentation"), "controller.py", "import pkg.domain.model\n")
    createTestPythonFile(t, filepath.Join(testDir, "pkg", "domain"), "model.py", "# domain\n")
    outputDir := t.TempDir()
    // Write architecture config into testDir
    archCfg := `
[output]
directory = "%s"

[architecture]
enabled = true

[[architecture.layers]]
name = "presentation"
packages = ["pkg.presentation.**"]

[[architecture.layers]]
name = "domain"
packages = ["pkg.domain.**"]

[[architecture.rules]]
from = "presentation"
allow = []
`
    if err := os.WriteFile(filepath.Join(testDir, ".pyscn.toml"), []byte(fmt.Sprintf(archCfg, outputDir)), 0644); err != nil {
        t.Fatalf("write cfg: %v", err)
    }

    cmd := exec.Command(binaryPath, "deps", "--json", testDir)
    cmd.Dir = testDir
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil { t.Fatalf("deps json failed: %v, stderr: %s", err, stderr.String()) }
    files, _ := filepath.Glob(filepath.Join(outputDir, "deps_*.json"))
    if len(files) == 0 { t.Fatalf("no deps json generated for arch in %s", outputDir) }
    data, _ := os.ReadFile(files[0])
    var obj map[string]interface{}
    if err := json.Unmarshal(data, &obj); err != nil { t.Fatalf("invalid json: %v", err) }
    summary := obj["summary"].(map[string]interface{})
    if summary["layer_violations"].(float64) < 1 {
        t.Fatalf("expected layer violations in summary, got: %v", summary)
    }
}
