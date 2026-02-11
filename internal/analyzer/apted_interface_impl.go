package analyzer

// ComputeSimilarity (Interface Implementation)
func (a *APTEDAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment, calc *TFIDFCalculator) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}
	return a.ComputeSimilarityTrees(f1.TreeNode, f2.TreeNode, calc)
}

// ComputeDistance (Interface Implementation)
func (a *APTEDAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}
	return a.ComputeDistanceTrees(f1.TreeNode, f2.TreeNode)
}

// GetName returns the name of this analyzer
func (a *APTEDAnalyzer) GetName() string {
	return "apted"
}
