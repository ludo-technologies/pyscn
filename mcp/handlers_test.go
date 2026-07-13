package mcp_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/mcp"
	"github.com/ludo-technologies/pyscn/service"
	mcplib "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type args struct {
	arguments interface{}
	setupFS   func(t *testing.T) string
}

func setupConfig(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, "test-config")
	err := os.WriteFile(configFile, []byte(""), 0o644)
	require.NoError(t, err)
	return configFile
}

func setupTestFile(t *testing.T, filename string) string {
	t.Helper()
	tmp := t.TempDir()
	rootDir, err := os.Getwd()
	require.NoError(t, err)
	parentDir := filepath.Dir(rootDir)
	src := filepath.Join(parentDir, "testdata", "python", "simple", filename)
	dst := filepath.Join(tmp, filename)
	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(dst, data, 0o644))
	return dst
}

func setupNestedTestProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.Mkdir(nested, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "root.py"), []byte("def root():\n    return 1\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(nested, "child.py"), []byte("def child():\n    return 2\n"), 0o644))
	return root
}

func runToolTest(
	t *testing.T,
	setupFS func(t *testing.T) string,
	arguments interface{},
	handlerFunc func(*mcp.HandlerSet, context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error),
) *mcplib.CallToolResult {

	t.Helper()
	configFile := setupConfig(t)
	return runToolTestWithConfig(t, setupFS, arguments, configFile, handlerFunc)
}

func runToolTestWithConfig(
	t *testing.T,
	setupFS func(t *testing.T) string,
	arguments interface{},
	configFile string,
	handlerFunc func(*mcp.HandlerSet, context.Context, mcplib.CallToolRequest) (*mcplib.CallToolResult, error),
) *mcplib.CallToolResult {

	t.Helper()
	deps := mcp.NewTestDependencies(service.NewFileReader(), nil, configFile)
	h := mcp.NewHandlerSet(deps)

	var filePath string
	if setupFS != nil {
		filePath = setupFS(t)
	}

	if filePath != "" {
		if m, ok := arguments.(map[string]interface{}); ok {
			m["path"] = filePath
		}
	}

	req := mcplib.CallToolRequest{
		Params: mcplib.CallToolParams{
			Arguments: arguments,
		},
	}

	res, err := handlerFunc(h, context.Background(), req)
	require.NoError(t, err)

	return res
}

func setupCommunitiesDisabledConfig(t *testing.T) string {
	t.Helper()
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, ".pyscn.toml")
	err := os.WriteFile(configFile, []byte("[communities]\nenabled = false\n"), 0o644)
	require.NoError(t, err)
	return configFile
}

func TestHandleAnalyzeCode(t *testing.T) {
	type want struct {
		isError      *bool
		expectPrefix string
		check        func(t *testing.T, res *mcplib.CallToolResult)
	}
	errTrue := true
	errFalse := false
	tests := map[string]struct {
		args args
		want want
	}{
		"invalid_arguments_format": {
			args: args{
				arguments: "not-a-map",
			},
			want: want{
				isError:      &errTrue,
				expectPrefix: "invalid arguments format",
			},
		},
		"path_missing": {
			args: args{
				arguments: map[string]interface{}{},
			},
			want: want{
				isError: &errTrue,
			},
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/path",
				},
			},
			want: want{
				isError:      &errTrue,
				expectPrefix: "path does not exist",
			},
		},
		"success": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{},
			},
			want: want{
				isError: nil,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					require.Greater(t, len(res.Content), 0)
					text := mcplib.GetTextFromContent(res.Content[0])
					require.NotEmpty(t, text)
					var result map[string]interface{}
					require.NoError(t, json.Unmarshal([]byte(text), &result))
					assert.Contains(t, result, "health_score")

				},
			},
		},
		"success_full_output": {
			args: args{
				setupFS: func(t *testing.T) string { return setupTestFile(t, "classes.py") },
				arguments: map[string]interface{}{
					"output_mode": "full",
				},
			},
			want: want{
				isError: &errFalse,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					text := mcplib.GetTextFromContent(res.Content[0])
					require.NotEmpty(t, text)
				},
			},
		},
		"recursive_false_limits_directory_to_root": {
			args: args{
				setupFS: setupNestedTestProject,
				arguments: map[string]interface{}{
					"analyses":    []interface{}{"complexity"},
					"output_mode": "full",
					"recursive":   false,
				},
			},
			want: want{
				isError: &errFalse,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					text := mcplib.GetTextFromContent(res.Content[0])
					var result struct {
						Summary struct {
							TotalFiles int `json:"total_files"`
						} `json:"summary"`
					}
					require.NoError(t, json.Unmarshal([]byte(text), &result))
					assert.Equal(t, 1, result.Summary.TotalFiles)
				},
			},
		},
		"recursive_true_includes_nested_files": {
			args: args{
				setupFS: setupNestedTestProject,
				arguments: map[string]interface{}{
					"analyses":    []interface{}{"complexity"},
					"output_mode": "full",
					"recursive":   true,
				},
			},
			want: want{
				isError: &errFalse,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					text := mcplib.GetTextFromContent(res.Content[0])
					var result struct {
						Summary struct {
							TotalFiles int `json:"total_files"`
						} `json:"summary"`
					}
					require.NoError(t, json.Unmarshal([]byte(text), &result))
					assert.Equal(t, 2, result.Summary.TotalFiles)
				},
			},
		},
		"recursive_rejects_non_boolean": {
			args: args{
				setupFS: setupNestedTestProject,
				arguments: map[string]interface{}{
					"recursive": "false",
				},
			},
			want: want{
				isError:      &errTrue,
				expectPrefix: "recursive parameter must be a boolean",
			},
		},
		"communities_full_context_map": {
			args: args{
				setupFS: func(t *testing.T) string {
					rootDir, err := os.Getwd()
					require.NoError(t, err)
					return filepath.Join(filepath.Dir(rootDir), "testdata", "python", "community_bridge")
				},
				arguments: map[string]interface{}{
					"analyses":    []interface{}{"communities"},
					"output_mode": "full",
				},
			},
			want: want{
				isError: &errFalse,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					text := mcplib.GetTextFromContent(res.Content[0])
					var result map[string]interface{}
					require.NoError(t, json.Unmarshal([]byte(text), &result))

					community, ok := result["community_analysis"].(map[string]interface{})
					require.True(t, ok, "expected community_analysis in full output")

					contextMap, ok := community["community_context_map"].(map[string]interface{})
					require.True(t, ok, "expected community_context_map in community_analysis")
					assert.Equal(t, float64(1), contextMap["version"])
					bundles, ok := contextMap["bundles"].([]interface{})
					require.True(t, ok)
					assert.NotEmpty(t, bundles)
				},
			},
		},
		"analyses_complexity_only": {
			args: args{
				setupFS: func(t *testing.T) string { return setupTestFile(t, "classes.py") },
				arguments: map[string]interface{}{
					"analyses": []interface{}{"complexity"},
				},
			},
			want: want{
				isError: nil,
				check: func(t *testing.T, res *mcplib.CallToolResult) {
					text := mcplib.GetTextFromContent(res.Content[0])
					var result map[string]interface{}
					require.NoError(t, json.Unmarshal([]byte(text), &result))
					assert.Contains(t, result, "health_score")
				},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,

				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleAnalyzeCode,
			)

			if tc.want.isError != nil && *tc.want.isError != res.IsError {
				t.Fatalf("IsError = %v, want %v", res.IsError, *tc.want.isError)
			}
			if tc.want.expectPrefix != "" && len(res.Content) > 0 {
				text := mcplib.GetTextFromContent(res.Content[0])
				if !strings.HasPrefix(text, tc.want.expectPrefix) {
					t.Fatalf("error text %q does not start with %q", text, tc.want.expectPrefix)
				}
			}
			if tc.want.check != nil && len(res.Content) > 0 {
				tc.want.check(t, res)
			}
		})
	}
}

func TestHandleAnalyzeCode_RecursiveOmittedUsesProjectConfig(t *testing.T) {
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, ".pyscn.toml")
	require.NoError(t, os.WriteFile(configFile, []byte("[analysis]\nrecursive = false\n"), 0o644))

	arguments := map[string]interface{}{
		"analyses":    []interface{}{"complexity"},
		"output_mode": "full",
	}
	res := runToolTestWithConfig(
		t,
		setupNestedTestProject,
		arguments,
		configFile,
		(*mcp.HandlerSet).HandleAnalyzeCode,
	)
	require.False(t, res.IsError)

	text := mcplib.GetTextFromContent(res.Content[0])
	var result struct {
		Summary struct {
			TotalFiles int `json:"total_files"`
		} `json:"summary"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &result))
	assert.Equal(t, 1, result.Summary.TotalFiles)
}

func TestHandleAnalyzeCode_RecursiveOverrideUpdatesFullResponse(t *testing.T) {
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, ".pyscn.toml")
	require.NoError(t, os.WriteFile(configFile, []byte("[analysis]\nrecursive = false\n"), 0o644))

	arguments := map[string]interface{}{
		"analyses":    []interface{}{"complexity"},
		"output_mode": "full",
		"recursive":   true,
	}
	res := runToolTestWithConfig(
		t,
		setupNestedTestProject,
		arguments,
		configFile,
		(*mcp.HandlerSet).HandleAnalyzeCode,
	)
	require.False(t, res.IsError)

	text := mcplib.GetTextFromContent(res.Content[0])
	var result struct {
		Complexity struct {
			Request struct {
				Recursive bool `json:"recursive"`
			} `json:"request"`
		} `json:"complexity"`
		Summary struct {
			TotalFiles int `json:"total_files"`
		} `json:"summary"`
	}
	require.NoError(t, json.Unmarshal([]byte(text), &result))
	assert.Equal(t, 2, result.Summary.TotalFiles)
	assert.True(t, result.Complexity.Request.Recursive)
}

func TestHandleAnalyzeCodeExplicitCommunitiesOverrideDisabledConfig(t *testing.T) {
	t.Parallel()

	fixtureRoot := func(t *testing.T) string {
		rootDir, err := os.Getwd()
		require.NoError(t, err)
		return filepath.Join(filepath.Dir(rootDir), "testdata", "python", "community_bridge")
	}
	arguments := map[string]interface{}{
		"analyses":    []interface{}{"communities"},
		"output_mode": "full",
	}

	res := runToolTestWithConfig(
		t,
		fixtureRoot,
		arguments,
		setupCommunitiesDisabledConfig(t),
		(*mcp.HandlerSet).HandleAnalyzeCode,
	)

	require.False(t, res.IsError)
	require.Greater(t, len(res.Content), 0)
	text := mcplib.GetTextFromContent(res.Content[0])
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(text), &result))
	require.Contains(t, result, "community_analysis")
}

func TestHandleCheckComplexity(t *testing.T) {

	errTrue := true

	tests := map[string]struct {
		args    args
		isError *bool
	}{
		"invalid_arguments": {
			args:    args{arguments: "not-a-map"},
			isError: &errTrue,
		},
		"path_missing": {
			args:    args{arguments: map[string]interface{}{}},
			isError: &errTrue,
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/file.py",
				},
			},
			isError: &errTrue,
		},
		"success_single_file": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,
				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleCheckComplexity,
			)

			if tc.isError != nil {
				require.Equal(t, *tc.isError, res.IsError)
				return
			}

			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)

		})
	}
}

func TestHandleCheckCoupling(t *testing.T) {
	errTrue := true

	tests := map[string]struct {
		args    args
		isError *bool
	}{
		"happy_path_summary": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{
					"output_mode": "summary",
				},
			},
		},
		"happy_path_full": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{
					"output_mode": "full",
				},
			},
		},
		"invalid_arguments": {
			args: args{
				arguments: "bad",
			},
			isError: &errTrue,
		},
		"path_missing": {
			args: args{
				arguments: map[string]interface{}{},
			},
			isError: &errTrue,
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/file.py",
				},
			},
			isError: &errTrue,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,
				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleCheckCoupling,
			)

			if tc.isError != nil {
				require.Equal(t, *tc.isError, res.IsError)
				return
			}

			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)

			outText := mcplib.GetTextFromContent(res.Content[0])

			if argsMap, ok := tc.args.arguments.(map[string]interface{}); ok {
				if argsMap["output_mode"] == "summary" {
					var out map[string]interface{}
					require.NoError(t, json.Unmarshal([]byte(outText), &out))
					assert.Contains(t, out, "summary")
				}
			}
		})
	}

}
func TestHandleDetectClones(t *testing.T) {

	errTrue := true

	tests := map[string]struct {
		args    args
		isError *bool
	}{
		"invalid_arguments": {
			args:    args{arguments: "bad"},
			isError: &errTrue,
		},
		"path_missing": {
			args:    args{arguments: map[string]interface{}{}},
			isError: &errTrue,
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/file.py",
				},
			},
			isError: &errTrue,
		},
		"success_summary": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{},
			},
		},
		"success_detailed": {
			args: args{
				arguments: map[string]interface{}{
					"output_mode": "detailed",
				},
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
			},
		},
		"success_clone_only": {
			args: args{
				arguments: map[string]interface{}{
					"analyses": []interface{}{"clone"},
				},
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
			},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,
				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleDetectClones,
			)

			if tc.isError != nil {
				require.Equal(t, *tc.isError, res.IsError)
				return
			}

			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)
		})
	}

}

func TestHandleFindDeadCode(t *testing.T) {

	errTrue := true

	tests := map[string]struct {
		args    args
		isError *bool
	}{
		"happy_path_warning": {
			args: args{
				arguments: map[string]interface{}{
					"min_severity": "warning",
				},
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
			},
		},
		"happy_path_info": {
			args: args{
				arguments: map[string]interface{}{
					"min_severity": "info",
				},
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
			},
		},
		"happy_path_critical": {
			args: args{
				arguments: map[string]interface{}{
					"min_severity": "critical",
				},
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
			},
		},
		"happy_path_default_severity": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{},
			},
		},
		"invalid_arguments": {
			args: args{
				arguments: "bad",
			},
			isError: &errTrue,
		},
		"path_missing": {
			args: args{
				arguments: map[string]interface{}{},
			},
			isError: &errTrue,
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/file.py",
				},
			},
			isError: &errTrue,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,
				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleFindDeadCode,
			)

			if tc.isError != nil {
				require.Equal(t, *tc.isError, res.IsError)
				return
			}

			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)

			text := mcplib.GetTextFromContent(res.Content[0])
			var out map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(text), &out))
			assert.Contains(t, out, "summary")
		})
	}
}
func TestHandleGetHealthScore(t *testing.T) {

	errTrue := true

	tests := map[string]struct {
		args    args
		isError *bool
	}{
		"happy_path_single_file": {
			args: args{
				setupFS: func(t *testing.T) string {
					return setupTestFile(t, "classes.py")
				},
				arguments: map[string]interface{}{},
			},
		},
		"invalid_arguments": {
			args: args{
				arguments: "bad",
			},
			isError: &errTrue,
		},
		"path_missing": {
			args: args{
				arguments: map[string]interface{}{},
			},
			isError: &errTrue,
		},
		"path_not_exist": {
			args: args{
				arguments: map[string]interface{}{
					"path": "/non/existing/file.py",
				},
			},
			isError: &errTrue,
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			res := runToolTest(
				t,
				tc.args.setupFS,
				tc.args.arguments,
				(*mcp.HandlerSet).HandleGetHealthScore,
			)

			if tc.isError != nil {
				require.Equal(t, *tc.isError, res.IsError)
				return
			}

			if res.IsError && len(res.Content) > 0 {
				t.Logf("unexpected error content: %s", mcplib.GetTextFromContent(res.Content[0]))
			}
			require.False(t, res.IsError)
			require.NotEmpty(t, res.Content)

			text := mcplib.GetTextFromContent(res.Content[0])
			var out map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(text), &out))

			assert.Contains(t, out, "health_score")
			assert.Contains(t, out, "grade")
			assert.Contains(t, out, "category_scores")
		})
	}
}
