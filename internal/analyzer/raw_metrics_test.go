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

func TestFunctionSLOC(t *testing.T) {
	slocOf := func(source string, startLine, endLine int) int {
		return CalculateRawMetrics([]byte(source), "test.py").FunctionSLOC(startLine, endLine)
	}

	t.Run("nil receiver and empty content", func(t *testing.T) {
		var nilResult *RawMetricsResult
		assert.Equal(t, 0, nilResult.FunctionSLOC(1, 10))
		assert.Equal(t, 0, slocOf("", 1, 10))
	})

	t.Run("invalid line range", func(t *testing.T) {
		const source = "a = 1\nb = 2\n"
		assert.Equal(t, 0, slocOf(source, 0, 2))
		assert.Equal(t, 0, slocOf(source, 2, 1))
		assert.Equal(t, 0, slocOf(source, -1, 2))
	})

	t.Run("start line beyond content", func(t *testing.T) {
		assert.Equal(t, 0, slocOf("a = 1\nb = 2\n", 100, 200))
	})

	t.Run("end line clamped to content length", func(t *testing.T) {
		assert.Equal(t, 2, slocOf("a = 1\nb = 2\nc = 3\n", 2, 100))
	})

	t.Run("counts only source lines, excluding comments and blanks", func(t *testing.T) {
		const source = `"""module docstring"""
# comment
a = 1

def f():
    # inner comment

    return 42
`
		assert.Equal(t, 2, slocOf(source, 5, 8))
	})

	t.Run("long flat function returns correct SLOC", func(t *testing.T) {
		const source = `def build_table():
    rows = []
    rows.append(1)
    rows.append(2)
    rows.append(3)
    rows.append(4)
    rows.append(5)
    rows.append(6)
    rows.append(7)
    rows.append(8)
    rows.append(9)
    rows.append(10)
    return rows
`
		assert.Equal(t, 13, slocOf(source, 1, 13))
	})

	t.Run("function with comment lines excluded", func(t *testing.T) {
		const source = `def greet(name):
    # Say hello
    # This is a comment
    message = "Hello, " + name
    return message
`
		assert.Equal(t, 3, slocOf(source, 1, 5))
	})

	t.Run("function docstring excluded", func(t *testing.T) {
		const source = `def greet(name):
    """Greet someone.

    Multi-line docstrings never count as source lines.
    """
    return name
`
		assert.Equal(t, 2, slocOf(source, 1, 6))
	})

	t.Run("multi-line string assignment counts as source", func(t *testing.T) {
		const source = `def template():
    text = """
    line one
    line two
    """
    return text
`
		assert.Equal(t, 6, slocOf(source, 1, 6))
	})

	t.Run("nested definition counts toward the enclosing range", func(t *testing.T) {
		const source = `def outer():
    def inner():
        return 1

    return inner()
`
		assert.Equal(t, 4, slocOf(source, 1, 5))
		assert.Equal(t, 2, slocOf(source, 2, 3))
	})

	t.Run("decorated method inside a class", func(t *testing.T) {
		const source = `class Service:
    """Class docstring."""

    @property
    def name(self):
        # comment
        return self._name
`
		assert.Equal(t, 3, slocOf(source, 4, 7))
	})

	t.Run("CRLF line endings", func(t *testing.T) {
		const source = "def f():\r\n    # comment\r\n    return 1\r\n"
		assert.Equal(t, 2, slocOf(source, 1, 3))
	})

	t.Run("ranges are consistent with the file-level metric", func(t *testing.T) {
		const source = `"""Module docstring."""
import os

def f():
    return os


def g():
    # comment
    return 2
`
		result := CalculateRawMetrics([]byte(source), "test.py")
		assert.Equal(t, result.SLOC, result.FunctionSLOC(1, result.TotalLines))
	})
}
