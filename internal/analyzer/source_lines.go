package analyzer

import "bytes"

// CountSourceLines counts physical source lines. A terminal newline ends the
// preceding line; it does not create another one.
func CountSourceLines(content []byte) int {
	if len(content) == 0 {
		return 0
	}
	lines := bytes.Count(content, []byte{'\n'})
	if content[len(content)-1] != '\n' {
		lines++
	}
	return lines
}
