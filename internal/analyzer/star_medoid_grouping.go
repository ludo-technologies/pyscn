package analyzer

import (
	"fmt"
	"sort"
)

// StarMedoidGrouping groups fragments by iteratively selecting medoids and reassigning members.
// It uses provided pair similarities only and does not recompute.
type StarMedoidGrouping struct {
	threshold     float64
	maxIterations int
	noChangeLimit int
}

type fragmentSimilarity struct {
	fragment   *CodeFragment
	similarity float64
}

// NewStarMedoidGrouping creates a new Star/Medoid grouping with a similarity threshold.
// Default: maxIterations=10, early-stop after 3 consecutive no-change iterations.
func NewStarMedoidGrouping(threshold float64) *StarMedoidGrouping {
	return &StarMedoidGrouping{
		threshold:     threshold,
		maxIterations: 10,
		noChangeLimit: 3,
	}
}

func (s *StarMedoidGrouping) GetName() string { return "Star/Medoid" }

// GroupClones groups clone pairs using a star/medoid strategy.
func (s *StarMedoidGrouping) GroupClones(pairs []*ClonePair) []*CloneGroup {
	if len(pairs) == 0 {
		return []*CloneGroup{}
	}

	// 1. Collect unique fragments and cache similarities
	fragments := s.collectFragments(pairs)
	if len(fragments) == 0 {
		return []*CloneGroup{}
	}
	simMap := s.buildSimilarityMap(pairs)
	graph := s.buildSimilarityGraph(pairs)

	// 2. Initialize clusters: each fragment alone, with union-find to merge via medoids
	parent := make(map[*CodeFragment]*CodeFragment, len(fragments))
	rank := make(map[*CodeFragment]int, len(fragments))
	var ufFind func(*CodeFragment) *CodeFragment
	ufFind = func(x *CodeFragment) *CodeFragment {
		if parent[x] != x {
			parent[x] = ufFind(parent[x])
		}
		return parent[x]
	}
	ufUnion := func(a, b *CodeFragment) bool {
		ra := ufFind(a)
		rb := ufFind(b)
		if ra == rb {
			return false
		}
		if rank[ra] < rank[rb] {
			parent[ra] = rb
		} else if rank[ra] > rank[rb] {
			parent[rb] = ra
		} else {
			parent[rb] = ra
			rank[ra]++
		}
		return true
	}
	for _, f := range fragments {
		parent[f] = f
		rank[f] = 0
	}

	// helper to rebuild clusters from union-find parents
	buildClusters := func() [][]*CodeFragment {
		groups := make(map[*CodeFragment][]*CodeFragment)
		for _, f := range fragments {
			r := ufFind(f)
			groups[r] = append(groups[r], f)
		}
		out := make([][]*CodeFragment, 0, len(groups))
		for _, members := range groups {
			out = append(out, members)
		}
		return out
	}

	clusters := buildClusters()

	// 3. Iteratively improve: medoid selection and reassignment
	noChangeStreak := 0
	for iter := 0; iter < s.maxIterations; iter++ {
		// a) compute medoids for each cluster
		medoids := make([]*CodeFragment, len(clusters))
		for i, members := range clusters {
			medoids[i] = s.findMedoid(members, graph)
		}

		// b) connect each non-medoid fragment to the most similar medoid (union-find)
		isMedoid := make(map[*CodeFragment]bool, len(medoids))
		for _, m := range medoids {
			if m != nil {
				isMedoid[m] = true
			}
		}
		changed := false
		for _, f := range fragments {
			// Keep medoids anchored after the first iteration to avoid oscillation
			if iter > 0 && isMedoid[f] {
				continue
			}
			bestMedoid, bestSim := s.mostSimilarMedoid(f, isMedoid, graph)
			if bestMedoid != nil && bestSim > 0.0 {
				if ufUnion(f, bestMedoid) {
					changed = true
				}
			}
		}

		// c) rebuild clusters and check convergence
		if !changed {
			noChangeStreak++
		} else {
			noChangeStreak = 0
		}
		clusters = buildClusters()
		if noChangeStreak >= s.noChangeLimit {
			break
		}
	}

	// 4. Filter members below threshold relative to cluster medoid
	// and 5. Build CloneGroup objects (exclude size-1 groups)
	result := make([]*CloneGroup, 0)
	groupID := 0
	for _, members := range clusters {
		if len(members) < 2 {
			continue // skip singletons
		}
		medoid := s.findMedoid(members, graph)
		filtered := make([]*CodeFragment, 0, len(members))
		for _, f := range members {
			if f == medoid {
				filtered = append(filtered, f)
				continue
			}
			if similarity(simMap, f, medoid) >= s.threshold {
				filtered = append(filtered, f)
			}
		}
		if len(filtered) < 2 {
			continue
		}

		// Sort fragments deterministically for stable output
		sort.Slice(filtered, func(i, j int) bool { return fragmentLess(filtered[i], filtered[j]) })

		group := NewCloneGroup(groupID)
		groupID++
		for _, f := range filtered {
			group.AddFragment(f)
		}
		// Compute average similarity from cache among group members
		group.Similarity = averageGroupSimilarity(simMap, filtered)
		result = append(result, group)
	}

	// Stable sort groups by decreasing similarity, then size, then first fragment location
	sort.Slice(result, func(i, j int) bool {
		if !almostEqual(result[i].Similarity, result[j].Similarity) {
			return result[i].Similarity > result[j].Similarity
		}
		if result[i].Size != result[j].Size {
			return result[i].Size > result[j].Size
		}
		// tie-breaker by first fragment location
		if len(result[i].Fragments) == 0 || len(result[j].Fragments) == 0 {
			return false
		}
		return fragmentLess(result[i].Fragments[0], result[j].Fragments[0])
	})
	for i, group := range result {
		group.ID = i
	}

	return result
}

// findMedoid selects the member with the maximum average similarity to others.
// Ties are broken by smaller location order.
func (s *StarMedoidGrouping) findMedoid(fragments []*CodeFragment, graph map[*CodeFragment][]fragmentSimilarity) *CodeFragment {
	if len(fragments) == 0 {
		return nil
	}
	if len(fragments) == 1 {
		return fragments[0]
	}

	var best *CodeFragment
	bestAvg := -1.0
	memberSet := make(map[*CodeFragment]struct{}, len(fragments))
	for _, fragment := range fragments {
		memberSet[fragment] = struct{}{}
	}
	for _, cand := range fragments {
		sum := 0.0
		for _, edge := range graph[cand] {
			if edge.fragment == cand {
				continue
			}
			if _, ok := memberSet[edge.fragment]; ok {
				sum += edge.similarity
			}
		}
		avg := sum / float64(len(fragments)-1)
		if avg > bestAvg || (almostEqual(avg, bestAvg) && best != nil && fragmentLess(cand, best)) {
			bestAvg = avg
			best = cand
		} else if best == nil { // initialize
			best = cand
			bestAvg = avg
		}
	}
	return best
}

func (s *StarMedoidGrouping) mostSimilarMedoid(
	fragment *CodeFragment,
	medoids map[*CodeFragment]bool,
	graph map[*CodeFragment][]fragmentSimilarity,
) (*CodeFragment, float64) {
	var best *CodeFragment
	bestSim := -1.0
	for _, edge := range graph[fragment] {
		candidate := edge.fragment
		if candidate == nil || candidate == fragment || !medoids[candidate] {
			continue
		}
		if edge.similarity > bestSim || (almostEqual(edge.similarity, bestSim) && best != nil && fragmentLess(candidate, best)) {
			bestSim = edge.similarity
			best = candidate
		} else if best == nil {
			bestSim = edge.similarity
			best = candidate
		}
	}
	return best, bestSim
}

// Helper: collect unique fragments from pairs
func (s *StarMedoidGrouping) collectFragments(pairs []*ClonePair) []*CodeFragment {
	seen := make(map[*CodeFragment]struct{})
	order := make([]*CodeFragment, 0)
	for _, p := range pairs {
		if p.Fragment1 != nil {
			if _, ok := seen[p.Fragment1]; !ok {
				seen[p.Fragment1] = struct{}{}
				order = append(order, p.Fragment1)
			}
		}
		if p.Fragment2 != nil {
			if _, ok := seen[p.Fragment2]; !ok {
				seen[p.Fragment2] = struct{}{}
				order = append(order, p.Fragment2)
			}
		}
	}
	return order
}

// Helper: build similarity cache from given pairs
func (s *StarMedoidGrouping) buildSimilarityMap(pairs []*ClonePair) map[string]float64 {
	sims := make(map[string]float64, len(pairs)*2)
	for _, p := range pairs {
		if p.Fragment1 == nil || p.Fragment2 == nil {
			continue
		}
		k := pairKey(p.Fragment1, p.Fragment2)
		if old, ok := sims[k]; !ok || p.Similarity > old {
			sims[k] = p.Similarity // keep the highest if duplicates exist
		}
	}
	return sims
}

func (s *StarMedoidGrouping) buildSimilarityGraph(pairs []*ClonePair) map[*CodeFragment][]fragmentSimilarity {
	best := make(map[*CodeFragment]map[*CodeFragment]float64)
	for _, p := range pairs {
		if p.Fragment1 == nil || p.Fragment2 == nil || p.Similarity <= 0 {
			continue
		}
		addBestSimilarity(best, p.Fragment1, p.Fragment2, p.Similarity)
		addBestSimilarity(best, p.Fragment2, p.Fragment1, p.Similarity)
	}

	graph := make(map[*CodeFragment][]fragmentSimilarity, len(best))
	for fragment, neighbors := range best {
		edges := make([]fragmentSimilarity, 0, len(neighbors))
		for neighbor, sim := range neighbors {
			edges = append(edges, fragmentSimilarity{fragment: neighbor, similarity: sim})
		}
		sort.Slice(edges, func(i, j int) bool {
			if !almostEqual(edges[i].similarity, edges[j].similarity) {
				return edges[i].similarity > edges[j].similarity
			}
			return fragmentLess(edges[i].fragment, edges[j].fragment)
		})
		graph[fragment] = edges
	}
	return graph
}

func addBestSimilarity(graph map[*CodeFragment]map[*CodeFragment]float64, from, to *CodeFragment, sim float64) {
	neighbors := graph[from]
	if neighbors == nil {
		neighbors = make(map[*CodeFragment]float64)
		graph[from] = neighbors
	}
	if old, ok := neighbors[to]; !ok || sim > old {
		neighbors[to] = sim
	}
}

// similarity returns cached similarity, or 0 if not present.
func similarity(sims map[string]float64, a, b *CodeFragment) float64 {
	if a == nil || b == nil {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	if v, ok := sims[pairKey(a, b)]; ok {
		return v
	}
	return 0.0
}

// averageGroupSimilarity computes average pairwise similarity among members using cache.
// Only pairs that exist in the similarity map are counted (missing pairs are skipped, not treated as 0).
func averageGroupSimilarity(sims map[string]float64, members []*CodeFragment) float64 {
	if len(members) < 2 {
		return 1.0
	}
	sum := 0.0
	cnt := 0
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			key := pairKey(members[i], members[j])
			if sim, ok := sims[key]; ok {
				sum += sim
				cnt++
			}
		}
	}
	if cnt == 0 {
		return 0.0
	}
	return sum / float64(cnt)
}

// pairKey creates a canonical key for a pair of fragments based on stable location ordering.
func pairKey(a, b *CodeFragment) string {
	ka := fragmentID(a)
	kb := fragmentID(b)
	if ka <= kb {
		return ka + "||" + kb
	}
	return kb + "||" + ka
}

// fragmentID returns a stable identifier for a fragment based on its location.
func fragmentID(f *CodeFragment) string {
	if f == nil || f.Location == nil {
		return fmt.Sprintf("%p", f)
	}
	loc := f.Location
	return fmt.Sprintf("%s|%d|%d|%d|%d", loc.FilePath, loc.StartLine, loc.EndLine, loc.StartCol, loc.EndCol)
}

// fragmentLess provides deterministic ordering between two fragments by location.
func fragmentLess(a, b *CodeFragment) bool {
	if a == b {
		return false
	}
	if a == nil {
		return true
	}
	if b == nil {
		return false
	}
	al, bl := a.Location, b.Location
	if al == nil && bl == nil {
		return fmt.Sprintf("%p", a) < fmt.Sprintf("%p", b)
	}
	if al == nil {
		return true
	}
	if bl == nil {
		return false
	}
	if al.FilePath != bl.FilePath {
		return al.FilePath < bl.FilePath
	}
	if al.StartLine != bl.StartLine {
		return al.StartLine < bl.StartLine
	}
	if al.StartCol != bl.StartCol {
		return al.StartCol < bl.StartCol
	}
	if al.EndLine != bl.EndLine {
		return al.EndLine < bl.EndLine
	}
	return al.EndCol < bl.EndCol
}

func almostEqual(a, b float64) bool {
	const eps = 1e-9
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}
