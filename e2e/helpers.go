package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// createTestConfigFile creates a temporary .pyscn.yaml config file for testing
// that directs output to the specified output directory
func createTestConfigFile(t *testing.T, testDir, outputDir string) {
	t.Helper()
	configFile := filepath.Join(testDir, ".pyscn.yaml")
	configContent := fmt.Sprintf("output:\n  directory: \"%s\"\n", outputDir)
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}
}