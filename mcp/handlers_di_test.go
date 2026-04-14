package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	mcptypes "github.com/mark3labs/mcp-go/mcp"
)

func TestHandleDetectDIAntipatterns_RespectsConfigThreshold(t *testing.T) {
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "service.py")
	source := `class Service:
    def __init__(self, a, b, c, d, e, f):
        pass
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	configPath := filepath.Join(tempDir, ".pyscn.toml")
	config := `[di]
constructor_param_threshold = 10
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	handlers := NewHandlerSet(NewDependencies(nil, configPath))
	request := mcptypes.CallToolRequest{
		Params: mcptypes.CallToolParams{
			Name: "detect_di_antipatterns",
			Arguments: map[string]interface{}{
				"path": tempDir,
			},
		},
	}

	result, err := handlers.HandleDetectDIAntipatterns(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected successful MCP tool result, got error result: %+v", result.Content)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected tool result content")
	}

	textContent, ok := result.Content[0].(mcptypes.TextContent)
	if !ok {
		t.Fatalf("expected text content, got %T", result.Content[0])
	}

	var response domain.DIAntipatternResponse
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("failed to unmarshal DI response: %v", err)
	}

	if response.Summary.TotalFindings != 0 {
		t.Fatalf("expected no DI findings with threshold from config, got %d", response.Summary.TotalFindings)
	}
}
