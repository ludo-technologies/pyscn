package analyzer

import "container/heap"

type completeLinkageBestNeighbor struct {
	clusterID  int
	neighborID int
	score      float64
}

type completeLinkageBestNeighborHeap struct {
	entries   []completeLinkageBestNeighbor
	positions []int
}

func newCompleteLinkageBestNeighborHeap(clusterCount int) *completeLinkageBestNeighborHeap {
	positions := make([]int, clusterCount)
	for i := range positions {
		positions[i] = -1
	}
	return &completeLinkageBestNeighborHeap{positions: positions}
}

func (h *completeLinkageBestNeighborHeap) Len() int { return len(h.entries) }

func (h *completeLinkageBestNeighborHeap) Less(i, j int) bool {
	if !almostEqual(h.entries[i].score, h.entries[j].score) {
		return h.entries[i].score > h.entries[j].score
	}

	leftI, rightI := orderClusterIDs(h.entries[i].clusterID, h.entries[i].neighborID)
	leftJ, rightJ := orderClusterIDs(h.entries[j].clusterID, h.entries[j].neighborID)
	if leftI != leftJ {
		return leftI < leftJ
	}
	if rightI != rightJ {
		return rightI < rightJ
	}
	return h.entries[i].clusterID < h.entries[j].clusterID
}

func (h *completeLinkageBestNeighborHeap) Swap(i, j int) {
	h.entries[i], h.entries[j] = h.entries[j], h.entries[i]
	h.positions[h.entries[i].clusterID] = i
	h.positions[h.entries[j].clusterID] = j
}

func (h *completeLinkageBestNeighborHeap) Push(x any) {
	entry := x.(completeLinkageBestNeighbor)
	h.positions[entry.clusterID] = len(h.entries)
	h.entries = append(h.entries, entry)
}

func (h *completeLinkageBestNeighborHeap) Pop() any {
	last := len(h.entries) - 1
	entry := h.entries[last]
	h.entries = h.entries[:last]
	h.positions[entry.clusterID] = -1
	return entry
}

func (h *completeLinkageBestNeighborHeap) set(clusterID, neighborID int, score float64) {
	if position := h.positions[clusterID]; position >= 0 {
		h.entries[position].neighborID = neighborID
		h.entries[position].score = score
		heap.Fix(h, position)
		return
	}

	heap.Push(h, completeLinkageBestNeighbor{
		clusterID:  clusterID,
		neighborID: neighborID,
		score:      score,
	})
}

func (h *completeLinkageBestNeighborHeap) remove(clusterID int) {
	position := h.positions[clusterID]
	if position < 0 {
		return
	}
	heap.Remove(h, position)
}

func (h *completeLinkageBestNeighborHeap) popBest() (completeLinkageBestNeighbor, bool) {
	if h.Len() == 0 {
		return completeLinkageBestNeighbor{}, false
	}
	return heap.Pop(h).(completeLinkageBestNeighbor), true
}

func orderClusterIDs(firstID, secondID int) (int, int) {
	if firstID < secondID {
		return firstID, secondID
	}
	return secondID, firstID
}
