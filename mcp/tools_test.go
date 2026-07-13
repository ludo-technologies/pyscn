package mcp

import (
	"encoding/json"
	"testing"

	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterTools_AdvertisesImplementedOutputControls(t *testing.T) {
	s := server.NewMCPServer("test", "test")
	RegisterTools(s, &HandlerSet{})

	tests := map[string][]string{
		"check_complexity": {"output_mode", "max_results"},
		"detect_clones":    {"output_mode", "max_results"},
		"check_coupling":   {"min_cbo", "output_mode", "max_results"},
		"check_cohesion":   {"output_mode", "max_results"},
		"find_dead_code":   {"output_mode", "max_results"},
	}

	for toolName, propertyNames := range tests {
		toolName := toolName
		propertyNames := propertyNames
		t.Run(toolName, func(t *testing.T) {
			registered := s.GetTool(toolName)
			require.NotNil(t, registered)

			properties := schemaProperties(t, registered.Tool.InputSchema)
			for _, propertyName := range propertyNames {
				assert.Contains(t, properties, propertyName)
			}

			outputMode := properties["output_mode"].(map[string]interface{})
			assert.ElementsMatch(t, []interface{}{"summary", "detailed", "full"}, outputMode["enum"])

			maxResults := properties["max_results"].(map[string]interface{})
			assert.Equal(t, float64(0), maxResults["minimum"])
		})
	}
}

func schemaProperties(t *testing.T, schema mcplib.ToolInputSchema) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(schema)
	require.NoError(t, err)

	var decoded struct {
		Properties map[string]interface{} `json:"properties"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded))
	return decoded.Properties
}
