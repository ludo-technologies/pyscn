package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// DefUseKind classifies how a variable is referenced
type DefUseKind int

const (
	// Definition kinds
	DefKindAssign       DefUseKind = iota // x = ...
	DefKindParameter                      // def f(x):
	DefKindForTarget                      // for x in ...
	DefKindImport                         // import x / from m import x
	DefKindWithTarget                     // with ... as x:
	DefKindExceptTarget                   // except E as x:
	DefKindAugmented                      // x += 1 (both def and use)

	// Use kinds
	UseKindRead      // ... = x (reading variable)
	UseKindCall      // f(x) (as argument)
	UseKindAttribute // x.attr (base object)
	UseKindSubscript // x[i] (base object)
)

// String returns the string representation of DefUseKind
func (k DefUseKind) String() string {
	switch k {
	case DefKindAssign:
		return "assign"
	case DefKindParameter:
		return "parameter"
	case DefKindForTarget:
		return "for_target"
	case DefKindImport:
		return "import"
	case DefKindWithTarget:
		return "with_target"
	case DefKindExceptTarget:
		return "except_target"
	case DefKindAugmented:
		return "augmented"
	case UseKindRead:
		return "read"
	case UseKindCall:
		return "call_arg"
	case UseKindAttribute:
		return "attribute"
	case UseKindSubscript:
		return "subscript"
	default:
		return "unknown"
	}
}

// IsDef returns true if this kind represents a definition
func (k DefUseKind) IsDef() bool {
	return k <= DefKindAugmented
}

// IsUse returns true if this kind represents a use
func (k DefUseKind) IsUse() bool {
	return k >= UseKindRead || k == DefKindAugmented
}

// VarReference represents a single definition or use of a variable
type VarReference struct {
	Name      string       // Variable name
	Kind      DefUseKind   // Type of reference
	Block     *BasicBlock  // Which block contains this reference
	Statement *parser.Node // The AST statement containing the reference
	Position  int          // Position within block's Statements slice
}

// NewVarReference creates a new variable reference
func NewVarReference(name string, kind DefUseKind, block *BasicBlock, stmt *parser.Node, pos int) *VarReference {
	return &VarReference{
		Name:      name,
		Kind:      kind,
		Block:     block,
		Statement: stmt,
		Position:  pos,
	}
}

// DefUsePair links a definition to its use
type DefUsePair struct {
	Def *VarReference
	Use *VarReference
}

// NewDefUsePair creates a new def-use pair
func NewDefUsePair(def, use *VarReference) *DefUsePair {
	return &DefUsePair{
		Def: def,
		Use: use,
	}
}

// IsCrossBlock returns true if the def and use are in different blocks
func (p *DefUsePair) IsCrossBlock() bool {
	if p.Def == nil || p.Use == nil {
		return false
	}
	if p.Def.Block == nil || p.Use.Block == nil {
		return false
	}
	return p.Def.Block.ID != p.Use.Block.ID
}

// DefUseChain represents all def-use relationships for a variable
type DefUseChain struct {
	Variable string
	Defs     []*VarReference
	Uses     []*VarReference
	Pairs    []*DefUsePair
}

// NewDefUseChain creates a new def-use chain for a variable
func NewDefUseChain(variable string) *DefUseChain {
	return &DefUseChain{
		Variable: variable,
		Defs:     []*VarReference{},
		Uses:     []*VarReference{},
		Pairs:    []*DefUsePair{},
	}
}

// AddDef adds a definition to the chain
func (c *DefUseChain) AddDef(ref *VarReference) {
	if ref != nil {
		c.Defs = append(c.Defs, ref)
	}
}

// AddUse adds a use to the chain
func (c *DefUseChain) AddUse(ref *VarReference) {
	if ref != nil {
		c.Uses = append(c.Uses, ref)
	}
}

// AddPair adds a def-use pair to the chain
func (c *DefUseChain) AddPair(pair *DefUsePair) {
	if pair != nil {
		c.Pairs = append(c.Pairs, pair)
	}
}

// DFAInfo holds complete data flow information for a CFG
type DFAInfo struct {
	CFG       *CFG
	Chains    map[string]*DefUseChain    // Variable name -> chain
	BlockDefs map[string][]*VarReference // Block ID -> definitions in block
	BlockUses map[string][]*VarReference // Block ID -> uses in block
}

// NewDFAInfo creates a new DFA info for a CFG
func NewDFAInfo(cfg *CFG) *DFAInfo {
	return &DFAInfo{
		CFG:       cfg,
		Chains:    make(map[string]*DefUseChain),
		BlockDefs: make(map[string][]*VarReference),
		BlockUses: make(map[string][]*VarReference),
	}
}

// GetChain returns the def-use chain for a variable, creating one if needed
func (info *DFAInfo) GetChain(variable string) *DefUseChain {
	if chain, ok := info.Chains[variable]; ok {
		return chain
	}
	chain := NewDefUseChain(variable)
	info.Chains[variable] = chain
	return chain
}

// AddDef adds a definition to the DFA info
func (info *DFAInfo) AddDef(ref *VarReference) {
	if ref == nil {
		return
	}
	chain := info.GetChain(ref.Name)
	chain.AddDef(ref)

	if ref.Block != nil {
		info.BlockDefs[ref.Block.ID] = append(info.BlockDefs[ref.Block.ID], ref)
	}
}

// AddUse adds a use to the DFA info
func (info *DFAInfo) AddUse(ref *VarReference) {
	if ref == nil {
		return
	}
	chain := info.GetChain(ref.Name)
	chain.AddUse(ref)

	if ref.Block != nil {
		info.BlockUses[ref.Block.ID] = append(info.BlockUses[ref.Block.ID], ref)
	}
}

// TotalDefs returns the total number of definitions
func (info *DFAInfo) TotalDefs() int {
	total := 0
	for _, chain := range info.Chains {
		total += len(chain.Defs)
	}
	return total
}

// TotalUses returns the total number of uses
func (info *DFAInfo) TotalUses() int {
	total := 0
	for _, chain := range info.Chains {
		total += len(chain.Uses)
	}
	return total
}

// TotalPairs returns the total number of def-use pairs
func (info *DFAInfo) TotalPairs() int {
	total := 0
	for _, chain := range info.Chains {
		total += len(chain.Pairs)
	}
	return total
}

// UniqueVariables returns the number of unique variables
func (info *DFAInfo) UniqueVariables() int {
	return len(info.Chains)
}

// DFAFeatures captures data flow characteristics for clone comparison
type DFAFeatures struct {
	TotalDefs       int // Total number of definitions
	TotalUses       int // Total number of uses
	TotalPairs      int // Total number of def-use pairs
	UniqueVariables int // Number of unique variables

	AvgChainLength  float64 // Average uses per definition
	MaxChainLength  int     // Maximum def-use chain length
	CrossBlockPairs int     // Def-use pairs spanning blocks
	IntraBlockPairs int     // Def-use pairs within same block

	DefKindCounts map[DefUseKind]int // Distribution of definition kinds
	UseKindCounts map[DefUseKind]int // Distribution of use kinds
}

// NewDFAFeatures creates a new DFA features instance
func NewDFAFeatures() *DFAFeatures {
	return &DFAFeatures{
		DefKindCounts: make(map[DefUseKind]int),
		UseKindCounts: make(map[DefUseKind]int),
	}
}

// ExtractDFAFeatures extracts DFA features from DFAInfo
func ExtractDFAFeatures(info *DFAInfo) *DFAFeatures {
	if info == nil {
		return NewDFAFeatures()
	}

	features := NewDFAFeatures()
	features.TotalDefs = info.TotalDefs()
	features.TotalUses = info.TotalUses()
	features.TotalPairs = info.TotalPairs()
	features.UniqueVariables = info.UniqueVariables()

	// Calculate chain metrics and count pairs by type
	totalChainLength := 0
	for _, chain := range info.Chains {
		chainLength := len(chain.Pairs)
		totalChainLength += chainLength
		if chainLength > features.MaxChainLength {
			features.MaxChainLength = chainLength
		}

		// Count cross-block vs intra-block pairs
		for _, pair := range chain.Pairs {
			if pair.IsCrossBlock() {
				features.CrossBlockPairs++
			} else {
				features.IntraBlockPairs++
			}
		}

		// Count definition kinds
		for _, def := range chain.Defs {
			features.DefKindCounts[def.Kind]++
		}

		// Count use kinds
		for _, use := range chain.Uses {
			features.UseKindCounts[use.Kind]++
		}
	}

	// Calculate average chain length
	if features.TotalDefs > 0 {
		features.AvgChainLength = float64(totalChainLength) / float64(features.TotalDefs)
	}

	return features
}
