package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestUnknownTomlKeys_PyscnToml(t *testing.T) {
	data := []byte(`
[clones]
grouping_mode = "complete_linkage"
totally_made_up_key = "xyz"

[clone]
group_mode = "complete_linkage"

[nonsense_section]
bogus_key = 42
`)

	unknown := UnknownTomlKeys(data, PyscnTomlConfig{})

	want := []string{"clone", "clones.totally_made_up_key", "nonsense_section"}
	if len(unknown) != len(want) {
		t.Fatalf("unknown keys = %v, want %v", unknown, want)
	}
	for i, key := range want {
		if unknown[i] != key {
			t.Errorf("unknown[%d] = %q, want %q", i, unknown[i], key)
		}
	}
}

func TestUnknownTomlKeys_AcceptsKnownKeys(t *testing.T) {
	data := []byte(`
project_root = "src"

[clones]
grouping_mode = "star"
min_lines = 5

[complexity]
low_threshold = 9

[[architecture.layers]]
name = "domain"
packages = ["domain"]
`)

	if unknown := UnknownTomlKeys(data, PyscnTomlConfig{}); len(unknown) != 0 {
		t.Errorf("unexpected unknown keys: %v", unknown)
	}
}

func TestUnknownTomlKeys_ArrayOfTablesReportsUnknownField(t *testing.T) {
	data := []byte(`
[[architecture.layers]]
name = "domain"
packagez = ["domain"]
`)

	unknown := UnknownTomlKeys(data, PyscnTomlConfig{})
	if len(unknown) != 1 || unknown[0] != "architecture.layers.packagez" {
		t.Errorf("unknown keys = %v, want [architecture.layers.packagez]", unknown)
	}
}

func TestUnknownTomlKeys_PyprojectIgnoresOtherTools(t *testing.T) {
	data := []byte(`
[project]
name = "demo"

[tool.ruff]
line-length = 120

[tool.pyscn.clones]
grouping_mode = "connected"
group_mode = "connected"
`)

	unknown := UnknownTomlKeys(data, PyprojectPyscnSection{}, "tool", "pyscn")
	if len(unknown) != 1 || unknown[0] != "tool.pyscn.clones.group_mode" {
		t.Errorf("unknown keys = %v, want [tool.pyscn.clones.group_mode]", unknown)
	}
}

func TestLoadFromPyscnToml_WarnsOnUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, pyscnTomlFileName)
	content := "[clone]\ngroup_mode = \"complete_linkage\"\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	var buf bytes.Buffer
	original := configWarningWriter
	configWarningWriter = &buf
	defer func() { configWarningWriter = original }()

	loader := NewTomlConfigLoader()
	if _, err := loader.LoadConfig(configPath); err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if got := buf.String(); got != "Warning: "+configPath+": unknown configuration key \"clone\"\n" {
		t.Errorf("warning output = %q", got)
	}
}
