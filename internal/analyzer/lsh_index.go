package analyzer

import (
	"fmt"
	"sort"
	"strconv"

	corelsh "github.com/ludo-technologies/polyscan/core/lsh"
)

const defaultLSHMaxCandidates = 1024

// lshCandidateIndex adapts core/lsh's string IDs to fragment indexes.
type lshCandidateIndex struct {
	index         *corelsh.LSHIndex
	maxCandidates int
}

func newLSHCandidateIndex(bands, rows, maxCandidates int) *lshCandidateIndex {
	if maxCandidates <= 0 {
		maxCandidates = defaultLSHMaxCandidates
	}
	return &lshCandidateIndex{
		index:         corelsh.NewLSHIndex(bands, rows),
		maxCandidates: maxCandidates,
	}
}

func (idx *lshCandidateIndex) AddFragment(id int, signature *corelsh.MinHashSignature) error {
	if id < 0 {
		return fmt.Errorf("negative fragment id: %d", id)
	}
	return idx.index.AddFragment(strconv.Itoa(id), signature)
}

func (idx *lshCandidateIndex) FindCandidates(signature *corelsh.MinHashSignature) []int {
	candidates := idx.index.FindCandidatesLimit(signature, idx.maxCandidates)
	ids := make([]int, 0, len(candidates))
	for _, candidate := range candidates {
		id, err := strconv.Atoi(candidate)
		if err == nil {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	return ids
}
