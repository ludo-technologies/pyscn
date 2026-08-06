package parser

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// The tree-sitter Python grammar still accepts a number of legacy Python 2
// constructs, so a file using them parses into a clean tree with no ERROR or
// MISSING node. Relying on tree-sitter alone therefore lets a file CPython
// refuses to compile be analyzed as if it were valid, and reported as healthy.
//
// pyscn has no target Python version, so these checks only reject constructs
// that every Python 3 release rejects. Anything a current release accepts is
// left alone, because flagging valid code as broken is the worse failure. In
// particular `except A, B:` is deliberately NOT rejected: PEP 758 made the
// unparenthesized exception list valid in Python 3.14, and only the `as` form
// (`except A, B as e:`) remains a syntax error everywhere.

// invalidInPython3 reports why every Python 3 release rejects the construct at
// tsNode, or an empty string when the node is accepted.
func invalidInPython3(tsNode *sitter.Node, source []byte) string {
	switch tsNode.Type() {
	case "print_statement":
		// `print >>f, x` is the Python 2 redirect form, but Python 3 still
		// parses it as `print >> f` followed by a tuple, so it stays valid.
		if firstChildOfType(tsNode, "chevron") != nil {
			return ""
		}
		return "the print statement (Python 3 requires print(...))"
	case "exec_statement":
		return "the exec statement (Python 3 requires exec(...))"
	case "<>":
		return "the <> operator (Python 3 requires !=)"
	case "except_clause":
		// An unparenthesized exception list is allowed since PEP 758, but never
		// together with `as`, which the grammar nests as an as_pattern.
		if firstChildOfType(tsNode, ",") != nil && firstChildOfType(tsNode, "as_pattern") != nil {
			return "an unparenthesized exception list bound with `as` (multiple exception types must be parenthesized when using `as`)"
		}
	case "raise_statement":
		// `raise E, v` / `raise E, v, tb`. A valid Python 3 raise carries a
		// single expression, optionally followed by `from`.
		if firstChildOfType(tsNode, "expression_list") != nil {
			return "raise with a comma-separated argument list (Python 3 requires raise Exc(msg))"
		}
	case "string":
		if start := tsNode.Child(0); start != nil && start.Content(source) == "`" {
			return "backtick repr (Python 3 requires repr(...))"
		}
	case "integer":
		return invalidPython3Integer(tsNode.Content(source))
	}
	return ""
}

// invalidPython3Integer reports why Python 3 rejects the integer literal text,
// or an empty string when the literal is accepted.
func invalidPython3Integer(text string) string {
	if strings.HasSuffix(text, "L") || strings.HasSuffix(text, "l") {
		return "an L-suffixed long literal (Python 3 has a single int type)"
	}
	if isLegacyOctal(text) {
		return "a bare leading-zero octal literal (Python 3 requires the 0o prefix)"
	}
	return ""
}

// isLegacyOctal reports whether text is a Python 2 octal literal such as 0777.
// Python 3 tolerates a leading zero only when every remaining digit is also zero
// (0, 00, 0_0); any other value needs the explicit 0o prefix.
func isLegacyOctal(text string) bool {
	if !strings.HasPrefix(text, "0") {
		return false
	}
	for _, r := range text[1:] {
		switch {
		case r == '0' || r == '_':
			continue
		case r >= '1' && r <= '9':
			return true
		default:
			// 0x/0b/0o prefixes and the j complex suffix are valid Python 3.
			return false
		}
	}
	return false
}

// firstChildOfType returns the first direct child of tsNode with the given type,
// or nil when there is none.
func firstChildOfType(tsNode *sitter.Node, childType string) *sitter.Node {
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		if child := tsNode.Child(i); child != nil && child.Type() == childType {
			return child
		}
	}
	return nil
}
