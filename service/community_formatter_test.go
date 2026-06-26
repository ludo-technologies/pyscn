package service

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateCommunityGolden = flag.Bool("update", false, "update community analysis golden files")

func TestCommunityFormatter_Format_JSON_Golden(t *testing.T) {
	result := analyzeCommunityBridgeFixture(t)
	result.GeneratedAt = "2026-01-15T12:00:00Z"
	result.Version = "0.0.0-test"
	result.Config = map[string]any{
		"algorithm":           "leiden",
		"scope":               "module",
		"minCommunitySize":    2,
		"includeLazyEdges":    true,
		"reportBridgeModules": true,
		"resolution":          1.0,
		"includeStdLib":       false,
		"includeThirdParty":   true,
		"followRelative":      true,
	}

	formatter := NewCommunityFormatter()
	output, err := formatter.Format(result, domain.OutputFormatJSON)
	require.NoError(t, err)

	goldenPath := filepath.Join("testdata", "golden", "community_analysis.json")
	if *updateCommunityGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0o755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(output), 0o644))
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	assertCommunityJSONEqual(t, expected, []byte(output))
}

func TestCommunityFormatter_Format_JSON_StableOrdering(t *testing.T) {
	result := &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 2,
		Modularity:       0.21874999999999997,
		Communities: []domain.CommunityMetrics{
			{
				ID:                      "community_2",
				Modules:                 []string{"mod.d", "mod.c"},
				Packages:                []string{"mod"},
				InternalEdges:           1,
				ExternalEdges:           1,
				ExternalDependencyRatio: 0.49999999999999994,
				Size:                    2,
			},
			{
				ID:                      "community_1",
				Modules:                 []string{"mod.b", "bridge", "mod.a"},
				Packages:                nil,
				InternalEdges:           2,
				ExternalEdges:           1,
				ExternalDependencyRatio: 0.3333333333333333,
				Size:                    3,
			},
		},
		BridgeModules: []domain.BridgeModule{
			{
				Module:              "mod.c",
				Community:           "community_2",
				CrossCommunityEdges: 1,
				TargetCommunities:   []string{"community_1"},
			},
			{
				Module:              "bridge",
				Community:           "community_1",
				CrossCommunityEdges: 1,
				TargetCommunities:   []string{"community_2"},
			},
		},
		GeneratedAt: "2026-01-15T12:00:00Z",
		Version:     "0.0.0-test",
	}

	formatter := NewCommunityFormatter()
	output, err := formatter.Format(result, domain.OutputFormatJSON)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &decoded))

	communities, ok := decoded["communities"].([]any)
	require.True(t, ok)
	require.Len(t, communities, 2)
	assert.Equal(t, "community_1", communities[0].(map[string]any)["id"])
	assert.Equal(t, "community_2", communities[1].(map[string]any)["id"])

	firstModules := communities[0].(map[string]any)["modules"].([]any)
	assert.Equal(t, []any{"bridge", "mod.a", "mod.b"}, firstModules)

	firstPackages := communities[0].(map[string]any)["packages"].([]any)
	assert.Equal(t, []any{}, firstPackages)

	bridges, ok := decoded["bridge_modules"].([]any)
	require.True(t, ok)
	require.Len(t, bridges, 2)
	assert.Equal(t, "bridge", bridges[0].(map[string]any)["module"])
	assert.Equal(t, "mod.c", bridges[1].(map[string]any)["module"])
}

func TestCommunityFormatter_Write_JSON(t *testing.T) {
	result := &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 1,
		Communities:      []domain.CommunityMetrics{{ID: "community_1", Modules: []string{"mod.a"}, Size: 1}},
		BridgeModules:    []domain.BridgeModule{},
		GeneratedAt:      "2026-01-15T12:00:00Z",
		Version:          "0.0.0-test",
	}

	formatter := NewCommunityFormatter()
	var buf bytes.Buffer
	require.NoError(t, formatter.Write(result, domain.OutputFormatJSON, &buf))

	var decoded domain.CommunityAnalysisResult
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, 1, decoded.TotalCommunities)
}

func TestCommunityFormatter_Format_Text(t *testing.T) {
	formatter := NewCommunityFormatter()
	result := &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 2,
		Modularity:       0.42,
		Communities: []domain.CommunityMetrics{
			{ID: "community_1", Size: 3, InternalEdges: 2, ExternalEdges: 1},
			{ID: "community_2", Size: 2, InternalEdges: 1, ExternalEdges: 1},
		},
		BridgeModules: []domain.BridgeModule{
			{
				Module:              "bridge",
				Community:           "community_1",
				CrossCommunityEdges: 2,
				TargetCommunities:   []string{"community_2"},
			},
		},
	}

	output, err := formatter.Format(result, domain.OutputFormatText)
	require.NoError(t, err)
	assert.Contains(t, output, "Module Community Analysis")
	assert.Contains(t, output, "Total Communities")
	assert.Contains(t, output, "0.420")
	assert.Contains(t, output, "LARGEST COMMUNITIES")
	assert.Contains(t, output, "BRIDGE MODULES")
	assert.Contains(t, output, "bridge")
}

func TestCommunityFormatter_Format_HTML_PackageMismatch(t *testing.T) {
	score := 0.0
	result := &domain.CommunityAnalysisResult{
		Algorithm:             "leiden",
		Scope:                 "module",
		TotalCommunities:      2,
		Modularity:            0.42,
		PackageAlignmentScore: &score,
		SplitPackages:         []string{"mod"},
		MixedCommunities:      []string{"community_2"},
		Communities: []domain.CommunityMetrics{
			{
				ID:                          "community_1",
				Size:                        3,
				InternalEdges:               2,
				ExternalEdges:               1,
				IncomingCrossCommunityEdges: 0,
				OutgoingCrossCommunityEdges: 1,
				DominantPackage:             "mod",
				PackageCount:                1,
				PackageAlignment:            1.0,
			},
			{
				ID:                          "community_2",
				Size:                        2,
				InternalEdges:               1,
				ExternalEdges:               1,
				IncomingCrossCommunityEdges: 1,
				OutgoingCrossCommunityEdges: 0,
				DominantPackage:             "billing",
				PackageCount:                2,
				PackageAlignment:            0.5,
			},
		},
	}

	formatter := NewCommunityFormatter()
	output, err := formatter.Format(result, domain.OutputFormatHTML)
	require.NoError(t, err)
	assert.Contains(t, output, "Package Alignment")
	assert.Contains(t, output, "Package Mismatch")
	assert.Contains(t, output, "<th>Dominant Package</th>")
	assert.Contains(t, output, "<th>Packages</th>")
	assert.Contains(t, output, "<th>Package Alignment</th>")
	assert.Contains(t, output, ">mod<")
	assert.Contains(t, output, ">billing<")
	assert.Contains(t, output, ">1.000<")
	assert.Contains(t, output, ">0.500<")
	assert.Contains(t, output, "Split packages:")
	assert.Contains(t, output, "Mixed communities:")
}

func TestCommunityFormatter_Format_Text_PackageMismatch(t *testing.T) {
	score := 0.0
	result := &domain.CommunityAnalysisResult{
		Algorithm:             "leiden",
		Scope:                 "module",
		TotalCommunities:      2,
		Modularity:            0.42,
		PackageAlignmentScore: &score,
		SplitPackages:         []string{"mod"},
		Communities: []domain.CommunityMetrics{
			{
				ID:               "community_1",
				Size:             3,
				DominantPackage:  "mod",
				PackageCount:     1,
				PackageAlignment: 1.0,
			},
		},
	}

	formatter := NewCommunityFormatter()
	output, err := formatter.Format(result, domain.OutputFormatText)
	require.NoError(t, err)
	assert.Contains(t, output, "Package Alignment")
	assert.Contains(t, output, "PACKAGE MISMATCH")
	assert.Contains(t, output, "Split Packages")
	assert.Contains(t, output, "pkg-align")
}

func TestCommunityFormatter_Format_AllFormats_BridgeFixture(t *testing.T) {
	result := analyzeCommunityBridgeFixture(t)
	result.GeneratedAt = "2026-01-15T12:00:00Z"
	result.Version = "0.0.0-test"

	formatter := NewCommunityFormatter()

	text, err := formatter.Format(result, domain.OutputFormatText)
	require.NoError(t, err)
	assert.Contains(t, text, "2")
	assert.Contains(t, text, "BRIDGE MODULES")
	assert.NotContains(t, text, "mod.a")

	yamlOutput, err := formatter.Format(result, domain.OutputFormatYAML)
	require.NoError(t, err)
	assert.Contains(t, yamlOutput, "total_communities: 2")
	assert.Contains(t, yamlOutput, "bridge_modules:")

	csvOutput, err := formatter.Format(result, domain.OutputFormatCSV)
	require.NoError(t, err)
	assert.Contains(t, csvOutput, "Summary,Total Communities,2")
	assert.Contains(t, csvOutput, "Bridge,bridge")

	htmlOutput, err := formatter.Format(result, domain.OutputFormatHTML)
	require.NoError(t, err)
	assert.Contains(t, htmlOutput, "Community Analysis Report")
	assert.Contains(t, htmlOutput, "Bridge Modules")
	assert.Contains(t, htmlOutput, "bridge")
	assert.Contains(t, htmlOutput, "<th>Dominant Package</th>")
	assert.Contains(t, htmlOutput, "<th>Package Alignment</th>")
	assert.Contains(t, htmlOutput, ">mod<")
	assert.Contains(t, htmlOutput, "Split packages:")
	assert.NotContains(t, htmlOutput, "<nil>")

	dotOutput, err := formatter.Format(result, domain.OutputFormatDOT)
	require.NoError(t, err)
	assert.Contains(t, dotOutput, "digraph ModuleCommunities")
	assert.Contains(t, dotOutput, "bridge")
	assert.Contains(t, dotOutput, "->")
}

func analyzeCommunityBridgeFixture(t *testing.T) *domain.CommunityAnalysisResult {
	t.Helper()

	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.TotalCommunities)
	require.NotEmpty(t, result.BridgeModules)

	return result
}

func assertCommunityJSONEqual(t *testing.T, expected, actual []byte) {
	t.Helper()

	var exp map[string]any
	var act map[string]any
	require.NoError(t, json.Unmarshal(expected, &exp))
	require.NoError(t, json.Unmarshal(actual, &act))
	assert.Equal(t, exp, act)
}
