package analyzer

import "container/heap"

type completeLinkageCandidate struct {
	leftID       int
	rightID      int
	score        float64
	leftVersion  int
	rightVersion int
}

type completeLinkageCandidateHeap []completeLinkageCandidate

func (h completeLinkageCandidateHeap) Len() int { return len(h) }

func (h completeLinkageCandidateHeap) Less(i, j int) bool {
	if !almostEqual(h[i].score, h[j].score) {
		return h[i].score > h[j].score
	}
	if h[i].leftID != h[j].leftID {
		return h[i].leftID < h[j].leftID
	}
	return h[i].rightID < h[j].rightID
}

func (h completeLinkageCandidateHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *completeLinkageCandidateHeap) Push(x any) {
	*h = append(*h, x.(completeLinkageCandidate))
}

func (h *completeLinkageCandidateHeap) Pop() any {
	old := *h
	last := len(old) - 1
	item := old[last]
	*h = old[:last]
	return item
}

func (h *completeLinkageCandidateHeap) push(candidate completeLinkageCandidate) {
	heap.Push(h, candidate)
}

func (h *completeLinkageCandidateHeap) pop() completeLinkageCandidate {
	return heap.Pop(h).(completeLinkageCandidate)
}

func orderClusterIDs(firstID, secondID int) (int, int) {
	if firstID < secondID {
		return firstID, secondID
	}
	return secondID, firstID
}
