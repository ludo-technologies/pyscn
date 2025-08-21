package domain

import (
	"testing"

	"github.com/pyqol/pyqol/internal/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneType_String(t *testing.T) {
	tests := []struct {
		cloneType CloneType
		expected  string
	}{
		{Type1Clone, "Type-1"},
		{Type2Clone, "Type-2"},
		{Type3Clone, "Type-3"},
		{Type4Clone, "Type-4"},
		{CloneType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.cloneType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneLocation_String(t *testing.T) {
	location := &CloneLocation{
		FilePath:  "/path/to/file.py",
		StartLine: 10,
		EndLine:   20,
		StartCol:  5,
		EndCol:    15,
	}

	expected := "/path/to/file.py:10:5-20:15"
	result := location.String()
	assert.Equal(t, expected, result)
}

func TestCloneLocation_LineCount(t *testing.T) {
	tests := []struct {
		name      string
		startLine int
		endLine   int
		expected  int
	}{
		{"single line", 10, 10, 1},
		{"multiple lines", 10, 15, 6},
		{"zero-based edge case", 0, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			location := &CloneLocation{
				StartLine: tt.startLine,
				EndLine:   tt.endLine,
			}
			result := location.LineCount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClone_String(t *testing.T) {
	location := &CloneLocation{
		FilePath:  "/test.py",
		StartLine: 1,
		EndLine:   10,
	}

	clone := &Clone{
		ID:       42,
		Type:     Type1Clone,
		Location: location,
		Size:     25,
	}

	result := clone.String()
	expected := "Clone{ID: 42, Type: Type-1, Location: /test.py:1:0-10:0, Size: 25}"
	assert.Equal(t, expected, result)
}

func TestClonePair_String(t *testing.T) {
	location1 := &CloneLocation{
		FilePath:  "/test1.py",
		StartLine: 1,
		EndLine:   5,
	}

	location2 := &CloneLocation{
		FilePath:  "/test2.py",
		StartLine: 10,
		EndLine:   14,
	}

	clone1 := &Clone{Location: location1}
	clone2 := &Clone{Location: location2}

	pair := &ClonePair{
		Clone1:     clone1,
		Clone2:     clone2,
		Similarity: 0.85,
		Type:       Type2Clone,
	}

	result := pair.String()
	expected := "Type-2 clone: /test1.py:1:0-5:0 <-> /test2.py:10:0-14:0 (similarity: 0.850)"
	assert.Equal(t, expected, result)
}

func TestCloneGroup_String(t *testing.T) {
	group := &CloneGroup{
		ID:         1,
		Type:       Type3Clone,
		Size:       3,
		Similarity: 0.75,
	}

	result := group.String()
	expected := "CloneGroup{ID: 1, Type: Type-3, Size: 3, Similarity: 0.750}"
	assert.Equal(t, expected, result)
}

func TestCloneGroup_AddClone(t *testing.T) {
	group := &CloneGroup{ID: 1}
	assert.Equal(t, 0, group.Size, "Initial size should be 0")
	assert.Empty(t, group.Clones, "Initial clones should be empty")

	clone := &Clone{
		ID:       1,
		Location: &CloneLocation{FilePath: "/test.py"},
	}

	group.AddClone(clone)

	assert.Equal(t, 1, group.Size, "Size should be updated")
	assert.Len(t, group.Clones, 1, "Clones slice should contain one clone")
	assert.Equal(t, clone, group.Clones[0], "Clone should be stored correctly")

	// Add another clone
	clone2 := &Clone{
		ID:       2,
		Location: &CloneLocation{FilePath: "/test2.py"},
	}

	group.AddClone(clone2)

	assert.Equal(t, 2, group.Size, "Size should be updated again")
	assert.Len(t, group.Clones, 2, "Clones slice should contain two clones")
}

func TestCloneRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		request   *CloneRequest
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid request",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     50.0,
				Type1Threshold:      constants.DefaultType1CloneThreshold,
				Type2Threshold:      constants.DefaultType2CloneThreshold,
				Type3Threshold:      constants.DefaultType3CloneThreshold,
				Type4Threshold:      constants.DefaultType4CloneThreshold,
			},
			expectErr: false,
		},
		{
			name: "empty paths",
			request: &CloneRequest{
				Paths: []string{},
			},
			expectErr: true,
			errMsg:    "paths cannot be empty",
		},
		{
			name: "invalid min lines",
			request: &CloneRequest{
				Paths:    []string{"/test"},
				MinLines: 0,
			},
			expectErr: true,
			errMsg:    "min_lines must be >= 1",
		},
		{
			name: "invalid min nodes",
			request: &CloneRequest{
				Paths:    []string{"/test"},
				MinLines: 5,
				MinNodes: 0,
			},
			expectErr: true,
			errMsg:    "min_nodes must be >= 1",
		},
		{
			name: "invalid similarity threshold - too low",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: -0.1,
			},
			expectErr: true,
			errMsg:    "similarity_threshold must be between 0.0 and 1.0",
		},
		{
			name: "invalid similarity threshold - too high",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 1.1,
			},
			expectErr: true,
			errMsg:    "similarity_threshold must be between 0.0 and 1.0",
		},
		{
			name: "invalid max edit distance",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     -1.0,
			},
			expectErr: true,
			errMsg:    "max_edit_distance must be >= 0.0",
		},
		{
			name: "invalid type1 threshold",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     50.0,
				Type1Threshold:      1.1,
			},
			expectErr: true,
			errMsg:    "type1_threshold must be between 0.0 and 1.0",
		},
		{
			name: "invalid threshold ordering - type1 <= type2",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     50.0,
				Type1Threshold:      0.85,
				Type2Threshold:      0.85,
			},
			expectErr: true,
			errMsg:    "type1_threshold should be > type2_threshold",
		},
		{
			name: "invalid threshold ordering - type2 <= type3",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     50.0,
				Type1Threshold:      0.95,
				Type2Threshold:      0.70,
				Type3Threshold:      0.70,
			},
			expectErr: true,
			errMsg:    "type2_threshold should be > type3_threshold",
		},
		{
			name: "invalid threshold ordering - type3 <= type4",
			request: &CloneRequest{
				Paths:               []string{"/test"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				MaxEditDistance:     50.0,
				Type1Threshold:      0.95,
				Type2Threshold:      0.85,
				Type3Threshold:      0.60,
				Type4Threshold:      0.60,
			},
			expectErr: true,
			errMsg:    "type3_threshold should be > type4_threshold",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectErr {
				assert.Error(t, err, "Expected validation error")
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no validation error")
			}
		})
	}
}

func TestCloneRequest_HasValidOutputWriter(t *testing.T) {
	request := &CloneRequest{}
	assert.False(t, request.HasValidOutputWriter(), "Should return false when OutputWriter is nil")

	// This would require a mock writer in a real test
	// request.OutputWriter = &bytes.Buffer{}
	// assert.True(t, request.HasValidOutputWriter(), "Should return true when OutputWriter is set")
}

func TestCloneRequest_ShouldShowContent(t *testing.T) {
	request := &CloneRequest{ShowContent: false}
	assert.False(t, request.ShouldShowContent(), "Should return false when ShowContent is false")

	request.ShowContent = true
	assert.True(t, request.ShouldShowContent(), "Should return true when ShowContent is true")
}

func TestCloneRequest_ShouldGroupClones(t *testing.T) {
	request := &CloneRequest{GroupClones: false}
	assert.False(t, request.ShouldGroupClones(), "Should return false when GroupClones is false")

	request.GroupClones = true
	assert.True(t, request.ShouldGroupClones(), "Should return true when GroupClones is true")
}

func TestDefaultCloneRequest(t *testing.T) {
	request := DefaultCloneRequest()

	assert.NotNil(t, request, "Default request should not be nil")
	assert.Equal(t, []string{"."}, request.Paths, "Default paths should be current directory")
	assert.True(t, request.Recursive, "Default recursive should be true")
	assert.Equal(t, []string{"*.py"}, request.IncludePatterns, "Default include patterns should be *.py")
	assert.Contains(t, request.ExcludePatterns, "test_*.py", "Default exclude patterns should contain test files")
	assert.Contains(t, request.ExcludePatterns, "*_test.py", "Default exclude patterns should contain test files")
	assert.Equal(t, 5, request.MinLines, "Default min lines should be 5")
	assert.Equal(t, 10, request.MinNodes, "Default min nodes should be 10")
	assert.Equal(t, 0.8, request.SimilarityThreshold, "Default similarity threshold should be 0.8")
	assert.Equal(t, 50.0, request.MaxEditDistance, "Default max edit distance should be 50.0")
	assert.False(t, request.IgnoreLiterals, "Default ignore literals should be false")
	assert.False(t, request.IgnoreIdentifiers, "Default ignore identifiers should be false")
	assert.Equal(t, constants.DefaultType1CloneThreshold, request.Type1Threshold, "Default Type-1 threshold should match constant")
	assert.Equal(t, constants.DefaultType2CloneThreshold, request.Type2Threshold, "Default Type-2 threshold should match constant")
	assert.Equal(t, constants.DefaultType3CloneThreshold, request.Type3Threshold, "Default Type-3 threshold should match constant")
	assert.Equal(t, constants.DefaultType4CloneThreshold, request.Type4Threshold, "Default Type-4 threshold should match constant")
	assert.Equal(t, OutputFormatText, request.OutputFormat, "Default output format should be text")
	assert.False(t, request.ShowDetails, "Default show details should be false")
	assert.False(t, request.ShowContent, "Default show content should be false")
	assert.Equal(t, SortBySimilarity, request.SortBy, "Default sort by should be similarity")
	assert.True(t, request.GroupClones, "Default group clones should be true")
	assert.Equal(t, 0.0, request.MinSimilarity, "Default min similarity should be 0.0")
	assert.Equal(t, 1.0, request.MaxSimilarity, "Default max similarity should be 1.0")
	assert.Len(t, request.CloneTypes, 4, "Default clone types should include all 4 types")
	assert.Contains(t, request.CloneTypes, Type1Clone, "Default should include Type-1")
	assert.Contains(t, request.CloneTypes, Type2Clone, "Default should include Type-2")
	assert.Contains(t, request.CloneTypes, Type3Clone, "Default should include Type-3")
	assert.Contains(t, request.CloneTypes, Type4Clone, "Default should include Type-4")

	// Validate that the default request passes validation
	err := request.Validate()
	assert.NoError(t, err, "Default request should pass validation")
}

func TestNewCloneStatistics(t *testing.T) {
	stats := NewCloneStatistics()

	assert.NotNil(t, stats, "Statistics should not be nil")
	assert.Equal(t, 0, stats.TotalClones, "Initial total clones should be 0")
	assert.Equal(t, 0, stats.TotalClonePairs, "Initial total clone pairs should be 0")
	assert.Equal(t, 0, stats.TotalCloneGroups, "Initial total clone groups should be 0")
	assert.NotNil(t, stats.ClonesByType, "ClonesByType map should be initialized")
	assert.Empty(t, stats.ClonesByType, "ClonesByType map should be empty initially")
	assert.Equal(t, 0.0, stats.AverageSimilarity, "Initial average similarity should be 0.0")
	assert.Equal(t, 0, stats.LinesAnalyzed, "Initial lines analyzed should be 0")
	assert.Equal(t, 0, stats.FilesAnalyzed, "Initial files analyzed should be 0")
}

func TestCloneSortCriteria_Constants(t *testing.T) {
	// Test that constants are defined correctly
	assert.Equal(t, CloneSortCriteria("location"), SortClonesByLocation)
	assert.Equal(t, CloneSortCriteria("similarity"), SortClonesBySimilarity)
	assert.Equal(t, CloneSortCriteria("size"), SortClonesBySize)
	assert.Equal(t, CloneSortCriteria("type"), SortClonesByType)
	assert.Equal(t, CloneSortCriteria("confidence"), SortClonesByConfidence)
}

// Integration test for complete clone detection workflow
func TestCloneDetectionWorkflow(t *testing.T) {
	// Create a sample clone request
	request := DefaultCloneRequest()
	request.Paths = []string{"/test/project"}
	request.MinLines = 3
	request.MinNodes = 5
	request.SimilarityThreshold = 0.7

	// Validate the request
	err := request.Validate()
	require.NoError(t, err, "Request should be valid")

	// Create sample clone response
	location1 := &CloneLocation{
		FilePath:  "/test/project/module1.py",
		StartLine: 10,
		EndLine:   20,
	}

	location2 := &CloneLocation{
		FilePath:  "/test/project/module2.py",
		StartLine: 35,
		EndLine:   45,
	}

	clone1 := &Clone{
		ID:        1,
		Type:      Type1Clone,
		Location:  location1,
		Size:      15,
		LineCount: 11,
	}

	clone2 := &Clone{
		ID:        2,
		Type:      Type1Clone,
		Location:  location2,
		Size:      14,
		LineCount: 11,
	}

	clonePair := &ClonePair{
		ID:         1,
		Clone1:     clone1,
		Clone2:     clone2,
		Similarity: 0.92,
		Distance:   2.0,
		Type:       Type1Clone,
		Confidence: 0.89,
	}

	cloneGroup := &CloneGroup{
		ID:         1,
		Type:       Type1Clone,
		Similarity: 0.92,
		Size:       2,
	}
	cloneGroup.AddClone(clone1)
	cloneGroup.AddClone(clone2)

	statistics := &CloneStatistics{
		TotalClones:       2,
		TotalClonePairs:   1,
		TotalCloneGroups:  1,
		ClonesByType:      map[string]int{"Type-1": 1},
		AverageSimilarity: 0.92,
		LinesAnalyzed:     1000,
		FilesAnalyzed:     5,
	}

	response := &CloneResponse{
		Clones:      []*Clone{clone1, clone2},
		ClonePairs:  []*ClonePair{clonePair},
		CloneGroups: []*CloneGroup{cloneGroup},
		Statistics:  statistics,
		Request:     request,
		Duration:    1500,
		Success:     true,
	}

	// Verify the complete response structure
	assert.NotNil(t, response, "Response should not be nil")
	assert.True(t, response.Success, "Response should indicate success")
	assert.Len(t, response.Clones, 2, "Should have 2 clones")
	assert.Len(t, response.ClonePairs, 1, "Should have 1 clone pair")
	assert.Len(t, response.CloneGroups, 1, "Should have 1 clone group")
	assert.NotNil(t, response.Statistics, "Statistics should be present")
	assert.Equal(t, request, response.Request, "Request should be preserved in response")
	assert.Greater(t, response.Duration, int64(0), "Duration should be positive")
	assert.Empty(t, response.Error, "Error should be empty for successful response")

	// Verify statistics
	assert.Equal(t, 2, response.Statistics.TotalClones)
	assert.Equal(t, 1, response.Statistics.TotalClonePairs)
	assert.Equal(t, 1, response.Statistics.TotalCloneGroups)
	assert.Equal(t, 0.92, response.Statistics.AverageSimilarity)
	assert.Equal(t, 1000, response.Statistics.LinesAnalyzed)
	assert.Equal(t, 5, response.Statistics.FilesAnalyzed)

	// Verify clone pair details
	pair := response.ClonePairs[0]
	assert.Equal(t, Type1Clone, pair.Type)
	assert.Equal(t, 0.92, pair.Similarity)
	assert.Equal(t, 0.89, pair.Confidence)
	assert.NotNil(t, pair.Clone1)
	assert.NotNil(t, pair.Clone2)

	// Verify clone group details
	group := response.CloneGroups[0]
	assert.Equal(t, Type1Clone, group.Type)
	assert.Equal(t, 2, group.Size)
	assert.Len(t, group.Clones, 2)
	assert.Equal(t, 0.92, group.Similarity)
}

func TestCloneResponse_ErrorHandling(t *testing.T) {
	// Test error response
	request := DefaultCloneRequest()
	errorResponse := &CloneResponse{
		Clones:      []*Clone{},
		ClonePairs:  []*ClonePair{},
		CloneGroups: []*CloneGroup{},
		Statistics:  NewCloneStatistics(),
		Request:     request,
		Duration:    100,
		Success:     false,
		Error:       "Failed to parse Python files",
	}

	assert.False(t, errorResponse.Success, "Error response should indicate failure")
	assert.NotEmpty(t, errorResponse.Error, "Error response should have error message")
	assert.Empty(t, errorResponse.Clones, "Error response should have empty clones")
	assert.Empty(t, errorResponse.ClonePairs, "Error response should have empty clone pairs")
	assert.Empty(t, errorResponse.CloneGroups, "Error response should have empty clone groups")
}
