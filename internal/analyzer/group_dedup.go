package analyzer

// dedupeStrictSubsetGroupMembers removes clone-group members whose source
// range is a strict subset of (or identical to) another member's range in the
// same file. Groups reduced to fewer than two members are dropped.
//
// Why this exists: tryCreateClonePair / the LSH path already reject *direct*
// pairs between overlapping same-file fragments (see isOverlappingLocation),
// so cd.clonePairs cannot contain a same-file `(A, B)` where one strictly
// covers the other. Union-Find grouping, however, still merges such fragments
// into one group via a shared distinct-file neighbor — e.g., pairs
// `(A=x.py:512-542, C=y.py:1-30)` and `(B=x.py:515-542, C=y.py:1-30)` are both
// legal yet transitively connect A and B. This post-pass collapses those
// overlapping windows back to the maximal one per file.
//
// For exactly-equal ranges (which UF can produce in the same way), the first
// occurrence is kept; later duplicates are suppressed for deterministic output.
func dedupeStrictSubsetGroupMembers(groups []*CloneGroup) []*CloneGroup {
	if len(groups) == 0 {
		return groups
	}
	out := make([]*CloneGroup, 0, len(groups))
	for _, g := range groups {
		if g == nil {
			continue
		}
		kept := filterMaximalPerFile(g.Fragments)
		if len(kept) < 2 {
			continue
		}
		g.Fragments = kept
		g.Size = len(kept)
		out = append(out, g)
	}
	return out
}

// filterMaximalPerFile returns the subset of fragments that are maximal under
// the same-file containment order. A fragment is suppressed if any other
// kept fragment in the same file strictly covers it, or if it duplicates an
// earlier fragment's range exactly.
func filterMaximalPerFile(frags []*CodeFragment) []*CodeFragment {
	n := len(frags)
	if n <= 1 {
		return frags
	}
	suppressed := make([]bool, n)
	for i := 0; i < n; i++ {
		if suppressed[i] || frags[i] == nil || frags[i].Location == nil {
			continue
		}
		for j := 0; j < n; j++ {
			if i == j || suppressed[j] || frags[j] == nil || frags[j].Location == nil {
				continue
			}
			if covers(frags[i].Location, frags[j].Location, i, j) {
				suppressed[j] = true
			}
		}
	}
	out := make([]*CodeFragment, 0, n)
	for i, f := range frags {
		if !suppressed[i] {
			out = append(out, f)
		}
	}
	return out
}

// covers reports whether outer (at index iOuter) covers inner (at index
// iInner) in the same file. Strict coverage suppresses inner outright;
// identical ranges suppress only the later index so that exactly one survives.
func covers(outer, inner *CodeLocation, iOuter, iInner int) bool {
	if outer.FilePath != inner.FilePath {
		return false
	}
	if outer.StartLine > inner.StartLine || outer.EndLine < inner.EndLine {
		return false
	}
	if outer.StartLine == inner.StartLine && outer.EndLine == inner.EndLine {
		return iOuter < iInner
	}
	return true
}
