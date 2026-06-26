package analyzer

import (
	"fmt"
	"sort"
)

// CommunityPartitionMetrics holds community-level metrics derived from a Leiden
// partition over a module dependency graph.
type CommunityPartitionMetrics struct {
	Communities      []CommunityPartition
	BridgeModules    []BridgeModuleMetrics
	TotalCommunities int
	Modularity       float64
}

// CommunityPartition describes one detected community.
type CommunityPartition struct {
	ID                          string
	Modules                     []string
	Packages                    []string
	InternalEdges               int
	ExternalEdges               int
	ExternalDependencyRatio     float64
	IncomingCrossCommunityEdges int
	OutgoingCrossCommunityEdges int
	Size                        int

	DominantPackage  string
	PackageCount     int
	PackageAlignment float64
}

// PackageMismatchMetrics summarizes how well detected communities align with
// declared package boundaries.
type PackageMismatchMetrics struct {
	PackageAlignmentScore float64
	SplitPackages         []string
	MixedCommunities      []string
}

// BridgeModuleMetrics describes a module that couples multiple communities.
type BridgeModuleMetrics struct {
	Module              string
	CommunityID         string
	CrossCommunityEdges int
	TargetCommunities   []string
}

// ComputeCommunityMetrics derives per-community and bridge-module metrics from
// a Leiden partition. graph supplies package metadata; cg supplies directed edges.
func ComputeCommunityMetrics(graph *DependencyGraph, cg *CommunityGraph, leiden *LeidenResult) *CommunityPartitionMetrics {
	if cg == nil || leiden == nil || cg.NodeCount == 0 || len(leiden.Membership) == 0 {
		return &CommunityPartitionMetrics{}
	}

	communityIDs := buildStableCommunityIDs(leiden.Membership)
	nodeCommunity := make([]string, cg.NodeCount)
	for i, commIdx := range leiden.Membership {
		if i < len(nodeCommunity) && commIdx >= 0 && commIdx < len(communityIDs) {
			nodeCommunity[i] = communityIDs[commIdx]
		}
	}

	commIndex := make(map[string]int, len(communityIDs))
	partitions := make([]CommunityPartition, len(communityIDs))
	for i, id := range communityIDs {
		commIndex[id] = i
		partitions[i].ID = id
	}

	for i, name := range cg.NodeNames {
		if i >= len(nodeCommunity) || nodeCommunity[i] == "" {
			continue
		}
		idx := commIndex[nodeCommunity[i]]
		partitions[idx].Modules = append(partitions[idx].Modules, name)
		if graph != nil {
			if node, ok := graph.Nodes[name]; ok && node.Package != "" {
				partitions[idx].Packages = appendUniqueString(partitions[idx].Packages, node.Package)
			}
		}
	}

	for i := range partitions {
		sort.Strings(partitions[i].Modules)
		sort.Strings(partitions[i].Packages)
		partitions[i].Size = len(partitions[i].Modules)
	}

	for _, edge := range cg.DirectedEdges {
		if edge.FromIndex >= len(nodeCommunity) || edge.ToIndex >= len(nodeCommunity) {
			continue
		}
		fromComm := nodeCommunity[edge.FromIndex]
		toComm := nodeCommunity[edge.ToIndex]
		if fromComm == "" || toComm == "" {
			continue
		}

		if fromComm == toComm {
			fromIdx := commIndex[fromComm]
			partitions[fromIdx].InternalEdges++
			continue
		}

		fromIdx := commIndex[fromComm]
		toIdx := commIndex[toComm]
		partitions[fromIdx].ExternalEdges++
		partitions[fromIdx].OutgoingCrossCommunityEdges++
		partitions[toIdx].ExternalEdges++
		partitions[toIdx].IncomingCrossCommunityEdges++
	}

	for i := range partitions {
		total := partitions[i].InternalEdges + partitions[i].ExternalEdges
		if total > 0 {
			partitions[i].ExternalDependencyRatio = float64(partitions[i].ExternalEdges) / float64(total)
		}
	}

	bridgeModules := computeBridgeModules(cg, nodeCommunity)
	applyPackageMismatchMetrics(graph, cg, nodeCommunity, partitions)

	return &CommunityPartitionMetrics{
		Communities:      partitions,
		BridgeModules:    bridgeModules,
		TotalCommunities: leiden.NumCommunities,
		Modularity:       leiden.Modularity,
	}
}

// ComputePackageMismatchMetrics derives system-level package alignment metrics
// from per-community partitions that already include package mismatch fields.
func ComputePackageMismatchMetrics(partitions []CommunityPartition) *PackageMismatchMetrics {
	if len(partitions) == 0 {
		return &PackageMismatchMetrics{}
	}

	packageCommunities := make(map[string]map[string]struct{})
	mixedCommunities := make([]string, 0)
	var alignedPackages int
	var totalPackages int

	for _, partition := range partitions {
		if partition.PackageCount >= 2 {
			mixedCommunities = append(mixedCommunities, partition.ID)
		}

		for _, pkg := range partition.Packages {
			if pkg == "" {
				continue
			}
			if packageCommunities[pkg] == nil {
				packageCommunities[pkg] = make(map[string]struct{})
			}
			packageCommunities[pkg][partition.ID] = struct{}{}
		}
	}

	splitPackages := make([]string, 0)
	for pkg, communities := range packageCommunities {
		totalPackages++
		if len(communities) >= 2 {
			splitPackages = append(splitPackages, pkg)
			continue
		}
		alignedPackages++
	}
	sort.Strings(splitPackages)
	sort.Strings(mixedCommunities)

	score := 0.0
	if totalPackages > 0 {
		score = float64(alignedPackages) / float64(totalPackages)
	}

	return &PackageMismatchMetrics{
		PackageAlignmentScore: score,
		SplitPackages:         splitPackages,
		MixedCommunities:      mixedCommunities,
	}
}

func applyPackageMismatchMetrics(
	graph *DependencyGraph,
	cg *CommunityGraph,
	nodeCommunity []string,
	partitions []CommunityPartition,
) {
	if graph == nil || cg == nil || len(partitions) == 0 {
		return
	}

	modulePackage := make(map[string]string, len(cg.NodeNames))
	for i, name := range cg.NodeNames {
		if i >= len(nodeCommunity) || nodeCommunity[i] == "" {
			continue
		}
		if node, ok := graph.Nodes[name]; ok && node.Package != "" {
			modulePackage[name] = node.Package
		}
	}
	if len(modulePackage) == 0 {
		return
	}

	commIndex := make(map[string]int, len(partitions))
	for i := range partitions {
		commIndex[partitions[i].ID] = i
	}

	type edgeCohesion struct {
		samePackage int
		total       int
	}
	cohesionByCommunity := make(map[string]*edgeCohesion, len(partitions))

	for i := range partitions {
		packageCounts := make(map[string]int)
		modulesWithPackage := 0
		for _, module := range partitions[i].Modules {
			pkg, ok := modulePackage[module]
			if !ok || pkg == "" {
				continue
			}
			modulesWithPackage++
			packageCounts[pkg]++
		}
		if modulesWithPackage == 0 {
			continue
		}

		partitions[i].PackageCount = len(packageCounts)

		dominantPackage := ""
		dominantCount := 0
		for pkg, count := range packageCounts {
			if count > dominantCount || (count == dominantCount && (dominantPackage == "" || pkg < dominantPackage)) {
				dominantPackage = pkg
				dominantCount = count
			}
		}
		partitions[i].DominantPackage = dominantPackage
		partitions[i].PackageAlignment = float64(dominantCount) / float64(modulesWithPackage)
	}

	for _, edge := range cg.DirectedEdges {
		if edge.FromIndex >= len(nodeCommunity) || edge.ToIndex >= len(nodeCommunity) {
			continue
		}
		fromComm := nodeCommunity[edge.FromIndex]
		toComm := nodeCommunity[edge.ToIndex]
		if fromComm == "" || toComm == "" || fromComm != toComm {
			continue
		}

		fromName := cg.NodeNames[edge.FromIndex]
		toName := cg.NodeNames[edge.ToIndex]
		fromPkg, fromOK := modulePackage[fromName]
		toPkg, toOK := modulePackage[toName]
		if !fromOK || !toOK {
			continue
		}

		acc, ok := cohesionByCommunity[fromComm]
		if !ok {
			acc = &edgeCohesion{}
			cohesionByCommunity[fromComm] = acc
		}
		acc.total++
		if fromPkg == toPkg {
			acc.samePackage++
		}
	}

	for commID, acc := range cohesionByCommunity {
		if acc.total == 0 {
			continue
		}
		idx, ok := commIndex[commID]
		if !ok {
			continue
		}
		partitions[idx].PackageAlignment = float64(acc.samePackage) / float64(acc.total)
	}
}

func buildStableCommunityIDs(membership []int) []string {
	if len(membership) == 0 {
		return nil
	}

	seen := make(map[int]struct{})
	ids := make([]int, 0)
	for _, comm := range membership {
		if comm < 0 {
			continue
		}
		if _, ok := seen[comm]; ok {
			continue
		}
		seen[comm] = struct{}{}
		ids = append(ids, comm)
	}
	sort.Ints(ids)

	maxID := -1
	for _, id := range ids {
		if id > maxID {
			maxID = id
		}
	}
	if maxID < 0 {
		return nil
	}

	// Index by raw community id so membership lookups stay O(1).
	byID := make([]string, maxID+1)
	for rank, id := range ids {
		byID[id] = fmt.Sprintf("community_%d", rank+1)
	}
	return byID
}

func computeBridgeModules(cg *CommunityGraph, nodeCommunity []string) []BridgeModuleMetrics {
	type bridgeAccumulator struct {
		homeCommunity     string
		crossEdges        int
		targetCommunities map[string]struct{}
	}

	bridges := make(map[int]*bridgeAccumulator)

	for _, edge := range cg.DirectedEdges {
		if edge.FromIndex >= len(nodeCommunity) || edge.ToIndex >= len(nodeCommunity) {
			continue
		}
		fromComm := nodeCommunity[edge.FromIndex]
		toComm := nodeCommunity[edge.ToIndex]
		if fromComm == "" || toComm == "" || fromComm == toComm {
			continue
		}

		updateBridge := func(nodeIdx int, homeComm, targetComm string) {
			acc, ok := bridges[nodeIdx]
			if !ok {
				acc = &bridgeAccumulator{
					homeCommunity:     homeComm,
					targetCommunities: make(map[string]struct{}),
				}
				bridges[nodeIdx] = acc
			}
			acc.crossEdges++
			acc.targetCommunities[targetComm] = struct{}{}
		}

		updateBridge(edge.FromIndex, fromComm, toComm)
		updateBridge(edge.ToIndex, toComm, fromComm)
	}

	results := make([]BridgeModuleMetrics, 0, len(bridges))
	for nodeIdx, acc := range bridges {
		if len(acc.targetCommunities) == 0 {
			continue
		}

		targets := make([]string, 0, len(acc.targetCommunities))
		for target := range acc.targetCommunities {
			targets = append(targets, target)
		}
		sort.Strings(targets)

		moduleName := ""
		if nodeIdx >= 0 && nodeIdx < len(cg.NodeNames) {
			moduleName = cg.NodeNames[nodeIdx]
		}

		results = append(results, BridgeModuleMetrics{
			Module:              moduleName,
			CommunityID:         acc.homeCommunity,
			CrossCommunityEdges: acc.crossEdges,
			TargetCommunities:   targets,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].CrossCommunityEdges != results[j].CrossCommunityEdges {
			return results[i].CrossCommunityEdges > results[j].CrossCommunityEdges
		}
		return results[i].Module < results[j].Module
	})

	return results
}

func appendUniqueString(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
