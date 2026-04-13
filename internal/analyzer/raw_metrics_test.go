package analyzer

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calculateRawMetricsForParsedSource(t *testing.T, content []byte, filePath string) *RawMetricsResult {
	t.Helper()

	result := CalculateRawMetrics(content, filePath)
	require.NotNil(t, result)

	parsed, err := parser.New().Parse(context.Background(), content)
	require.NoError(t, err)

	PopulateLogicalLines(result, parsed.AST)
	return result
}

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
		result := calculateRawMetricsForParsedSource(t, content, "comments.py")

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

		result := calculateRawMetricsForParsedSource(t, content, "greeter.py")

		require.NotNil(t, result)
		assert.Equal(t, 12, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 5, result.LLOC)
		assert.Equal(t, 1, result.CommentLines)
		assert.Equal(t, 5, result.DocstringLines)
		assert.Equal(t, 1, result.BlankLines)
		assert.InDelta(t, float64(1)/float64(6), result.CommentRatio, 0.0001)
	})

	t.Run("raw and unicode prefixed docstrings are treated as docstrings", func(t *testing.T) {
		content := []byte(`r"""Module docstring"""

def greet():
    u"""Function docstring"""
    return "hello"
`)

		result := calculateRawMetricsForParsedSource(t, content, "prefixed_docstrings.py")

		require.NotNil(t, result)
		assert.Equal(t, 5, result.TotalLines)
		assert.Equal(t, 2, result.SLOC)
		assert.Equal(t, 2, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 2, result.DocstringLines)
		assert.Equal(t, 1, result.BlankLines)
	})

	t.Run("multiline statements and semicolons", func(t *testing.T) {
		content := []byte(`value = (
    1 +
    2
)
a = 1; b = 2
`)

		result := calculateRawMetricsForParsedSource(t, content, "statements.py")

		require.NotNil(t, result)
		assert.Equal(t, 5, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 3, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 0, result.BlankLines)
	})

	t.Run("formatted and byte triple quoted strings are not treated as docstrings", func(t *testing.T) {
		content := []byte(`f"""not a docstring"""
b"""also not a docstring"""
`)

		result := calculateRawMetricsForParsedSource(t, content, "non_docstring_prefixes.py")

		require.NotNil(t, result)
		assert.Equal(t, 2, result.TotalLines)
		assert.Equal(t, 2, result.SLOC)
		assert.Equal(t, 2, result.LLOC)
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

		result := calculateRawMetricsForParsedSource(t, content, "assignment.py")

		require.NotNil(t, result)
		assert.Equal(t, 5, result.TotalLines)
		assert.Equal(t, 5, result.SLOC)
		assert.Equal(t, 1, result.LLOC)
		assert.Equal(t, 0, result.CommentLines)
		assert.Equal(t, 0, result.DocstringLines)
		assert.Equal(t, 0, result.BlankLines)
	})

	t.Run("lloc stays zero when source cannot be parsed", func(t *testing.T) {
		result := CalculateRawMetrics([]byte("def broken(:\n    pass\n"), "invalid.py")

		require.NotNil(t, result)
		assert.Equal(t, 0, result.LLOC)
	})
}

func TestCalculateAggregateRawMetrics(t *testing.T) {
	first := calculateRawMetricsForParsedSource(t, []byte("a = 1\n# comment\n"), "first.py")
	second := calculateRawMetricsForParsedSource(t, []byte(`"""module"""

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

	t.Run("ignores nil entries when counting files", func(t *testing.T) {
		aggregate := CalculateAggregateRawMetrics([]*RawMetricsResult{first, nil, second})

		require.NotNil(t, aggregate)
		assert.Equal(t, 2, aggregate.FilesAnalyzed)
		assert.Equal(t, 2, aggregate.SLOC)
		assert.Equal(t, 2, aggregate.LLOC)
		assert.Equal(t, 1, aggregate.CommentLines)
		assert.Equal(t, 1, aggregate.DocstringLines)
		assert.Equal(t, 1, aggregate.BlankLines)
		assert.Equal(t, 5, aggregate.TotalLines)
		assert.InDelta(t, float64(1)/float64(3), aggregate.CommentRatio, 0.0001)
	})
}
