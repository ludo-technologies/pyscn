package analyzer

// completeLinkageClusterer stores only threshold-qualified inter-cluster edges.
// That keeps sparse workloads sparse while still supporting exact complete-linkage
// merges, since a merged cluster can stay adjacent to C only if both source
// clusters already had qualifying edges to C.
type completeLinkageClusterer struct {
	clusters      []completeLinkageCluster
	bestNeighbors *completeLinkageBestNeighborHeap
}

type completeLinkageCluster struct {
	members   []*CodeFragment
	neighbors map[int]float64
	active    bool
}

func newCompleteLinkageClusterer(fragments []*CodeFragment, edges []completeLinkageEdge) *completeLinkageClusterer {
	clusterer := &completeLinkageClusterer{
		clusters:      make([]completeLinkageCluster, len(fragments)),
		bestNeighbors: newCompleteLinkageBestNeighborHeap(len(fragments)),
	}

	for clusterID, fragment := range fragments {
		clusterer.clusters[clusterID] = completeLinkageCluster{
			members:   []*CodeFragment{fragment},
			neighbors: make(map[int]float64),
			active:    true,
		}
	}

	for _, edge := range edges {
		clusterer.clusters[edge.leftID].neighbors[edge.rightID] = edge.score
		clusterer.clusters[edge.rightID].neighbors[edge.leftID] = edge.score
	}

	for clusterID := range clusterer.clusters {
		clusterer.recomputeBestNeighbor(clusterID)
	}

	return clusterer
}

func (c *completeLinkageClusterer) mergeUntilStable() {
	for {
		bestNeighbor, ok := c.bestNeighbors.popBest()
		if !ok {
			return
		}

		targetID, sourceID := orderClusterIDs(bestNeighbor.clusterID, bestNeighbor.neighborID)
		if !c.clusters[targetID].active || !c.clusters[sourceID].active {
			continue
		}

		c.mergeClusters(targetID, sourceID)
	}
}

func (c *completeLinkageClusterer) mergeClusters(targetID, sourceID int) {
	target := &c.clusters[targetID]
	source := &c.clusters[sourceID]
	target.members = append(target.members, source.members...)

	affected := make(map[int]struct{}, len(target.neighbors)+len(source.neighbors))
	for neighborID := range target.neighbors {
		affected[neighborID] = struct{}{}
	}
	for neighborID := range source.neighbors {
		affected[neighborID] = struct{}{}
	}
	delete(affected, targetID)
	delete(affected, sourceID)

	newTargetNeighbors := make(map[int]float64)
	for neighborID := range affected {
		if !c.clusters[neighborID].active {
			continue
		}

		neighbor := &c.clusters[neighborID]
		delete(neighbor.neighbors, sourceID)

		targetScore, targetOK := target.neighbors[neighborID]
		sourceScore, sourceOK := source.neighbors[neighborID]
		if !targetOK || !sourceOK {
			delete(neighbor.neighbors, targetID)
			continue
		}

		mergedScore := targetScore
		if sourceScore < mergedScore {
			mergedScore = sourceScore
		}
		newTargetNeighbors[neighborID] = mergedScore
		neighbor.neighbors[targetID] = mergedScore
	}

	target.neighbors = newTargetNeighbors
	source.active = false
	source.members = nil
	source.neighbors = nil

	c.bestNeighbors.remove(sourceID)
	for neighborID := range affected {
		if c.clusters[neighborID].active {
			c.recomputeBestNeighbor(neighborID)
		}
	}
	c.recomputeBestNeighbor(targetID)
}

func (c *completeLinkageClusterer) recomputeBestNeighbor(clusterID int) {
	cluster := &c.clusters[clusterID]
	if !cluster.active {
		c.bestNeighbors.remove(clusterID)
		return
	}

	bestNeighborID, bestScore, ok := c.findBestNeighbor(clusterID)
	if !ok {
		c.bestNeighbors.remove(clusterID)
		return
	}

	c.bestNeighbors.set(clusterID, bestNeighborID, bestScore)
}

func (c *completeLinkageClusterer) findBestNeighbor(clusterID int) (int, float64, bool) {
	cluster := &c.clusters[clusterID]
	bestNeighborID := -1
	bestScore := 0.0
	for neighborID, score := range cluster.neighbors {
		if !c.clusters[neighborID].active {
			continue
		}
		if bestNeighborID == -1 || betterCompleteLinkageNeighbor(clusterID, neighborID, score, bestNeighborID, bestScore) {
			bestNeighborID = neighborID
			bestScore = score
		}
	}
	if bestNeighborID == -1 {
		return 0, 0.0, false
	}

	return bestNeighborID, bestScore, true
}

func betterCompleteLinkageNeighbor(clusterID, candidateNeighborID int, candidateScore float64, bestNeighborID int, bestScore float64) bool {
	if !almostEqual(candidateScore, bestScore) {
		return candidateScore > bestScore
	}

	candidateLeft, candidateRight := orderClusterIDs(clusterID, candidateNeighborID)
	bestLeft, bestRight := orderClusterIDs(clusterID, bestNeighborID)
	if candidateLeft != bestLeft {
		return candidateLeft < bestLeft
	}
	return candidateRight < bestRight
}

func (c *completeLinkageClusterer) activeClusters() []*completeLinkageCluster {
	activeClusters := make([]*completeLinkageCluster, 0)
	for clusterID := range c.clusters {
		if c.clusters[clusterID].active {
			activeClusters = append(activeClusters, &c.clusters[clusterID])
		}
	}
	return activeClusters
}
