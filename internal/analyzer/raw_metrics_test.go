package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateRawMetrics(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		result := CalculateRawMetrics([]byte(""), "empty.py")

		require.NotNil(t, result)
		assert.Equal(t, "empty.py", result.FilePath)
		assert.Equal(t, 0, result.TotalLines)
		assert.Equal(t, 0, result.SLOC)
		assert.Equal(t, 0, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 0, result.BlankLines)
		assert.Equal(t, 0.0, result.CommentRatio)
	})

	t.Run("comments and blanks only", func(t *testing.T) {
		content := []byte("# first comment\n\n    # second comment\n")
		result := CalculateRawMetrics(content, "comments.py")

		require.NotNil(t, result)
		assert.Equal(t, 3, result.TotalLines)
		assert.Equal(t, 0, result.SLOC)
		assert.Equal(t, 0, result.LLOC)
		assert.Equal(t, 2, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 1, result.BlankLines)
		assert.InDelta(t, 1.0, result.CommentRatio, 0.0001)
	})

	t.Run("module and block docstrings", func(t *testing.T) {
		content := []byte(`"""Module docstring
second line
"""

# leading comment
class Greeter:
    """Class docstring"""
    def greet(self):
        """Function docstring"""
        message = "hello"  # inline comment
        if message:
            return message
`)

		result := CalculateRawMetrics(content, "greeter.py")

		require.NotNil(t, result)
		assert.Equal(t, 12, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 5, result.LLOC)
		assert.Equal(t, 1, result.CommentLines)
		assert.Equal(t, 5, result.DocstringLines)
		assert.Equal(t, 1, result.BlankLines)
		assert.InDelta(t, float64(1)/float64(6), result.CommentRatio, 0.0001)
	})

	t.Run("multiline statements and semicolons", func(t *testing.T) {
		content := []byte(`value = (
    1 +
    2
)
a = 1; b = 2
`)

		result := CalculateRawMetrics(content, "statements.py")

		require.NotNil(t, result)
		assert.Equal(t, 5, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 3, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 0, result.BlankLines)
	})

	t.Run("triple quoted assignment is not treated as docstring", func(t *testing.T) {
		content := []byte(`TEXT = """
# not a comment

hello
"""
`)

		result := CalculateRawMetrics(content, "assignment.py")

		require.NotNil(t, result)
		assert.Equal(t, 5, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 1, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 0, result.BlankLines)
	})
}

func TestCalculateAggregateRawMetrics(t *testing.T) {
	first := CalculateRawMetrics([]byte("a = 1\n# comment\n"), "first.py")
	second := CalculateRawMetrics([]byte(`"""module"""

b = 2
`), "second.py")

	aggregate := CalculateAggregateRawMetrics([]*RawMetricsResult{first, second})

	require.NotNil(t, aggregate)
	assert.Equal(t, 2, aggregate.FilesAnalyzed)
	assert.Equal(t, 2, aggregate.SLOC)
	assert.Equal(t, 2, aggregate.LLOC)
	assert.Equal(t, 1, aggregate.CommentLines)
	assert.Equal(t, 1, aggregate.DocstringLines)
	assert.Equal(t, 1, aggregate.BlankLines)
	assert.Equal(t, 5, aggregate.TotalLines)
	assert.InDelta(t, float64(1)/float64(3), aggregate.CommentRatio, 0.0001)
}
