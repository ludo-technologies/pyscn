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
		TotalCommunities: 2,
		Modularity:       0.42,
	}

	output, err := formatter.Format(result, domain.OutputFormatText)
	require.NoError(t, err)
	assert.Contains(t, output, "2 communities")
	assert.Contains(t, output, "0.420")
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
