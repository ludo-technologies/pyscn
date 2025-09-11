package analyzer

import (
    "hash/fnv"
    "math"
    "math/rand"
)

// MinHashSignature holds the signature vector
type MinHashSignature struct {
    signatures []uint64
    numHashes  int
}

// HashFunc maps a 64-bit base hash to another 64-bit value
type HashFunc func(uint64) uint64

// MinHasher computes MinHash signatures for feature sets
type MinHasher struct {
    numHashes     int
    hashFunctions []HashFunc
}

// NewMinHasher creates a MinHasher with numHashes functions (default 128 if invalid)
func NewMinHasher(numHashes int) *MinHasher {
    if numHashes <= 0 {
        numHashes = 128
    }
    mh := &MinHasher{numHashes: numHashes}
    mh.generateHashFunctions()
    return mh
}

func (m *MinHasher) generateHashFunctions() {
    // Use simple 64-bit universal hashing: h_i(x) = (a_i * x) ^ b_i, with overflow
    // Deterministic seed for reproducibility
    rng := rand.New(rand.NewSource(0x5eed_1234_cafe_babe))
    a := make([]uint64, m.numHashes)
    b := make([]uint64, m.numHashes)
    for i := 0; i < m.numHashes; i++ {
        // Choose odd a to avoid trivial cycles
        ai := rng.Uint64() | 1
        bi := rng.Uint64()
        a[i], b[i] = ai, bi
    }
    m.hashFunctions = make([]HashFunc, m.numHashes)
    for i := 0; i < m.numHashes; i++ {
        ai, bi := a[i], b[i]
        m.hashFunctions[i] = func(x uint64) uint64 {
            return (ai*x)^bi + ai + bi
        }
    }
}

// ComputeSignature computes the MinHash signature for a set of features
func (m *MinHasher) ComputeSignature(features []string) *MinHashSignature {
    if len(features) == 0 {
        return &MinHashSignature{signatures: make([]uint64, m.numHashes), numHashes: m.numHashes}
    }
    // Deduplicate features to treat as a set
    set := make(map[string]struct{}, len(features))
    for _, f := range features {
        set[f] = struct{}{}
    }
    // Precompute base 64-bit hashes of features
    base := make([]uint64, 0, len(set))
    for f := range set {
        base = append(base, hash64(f))
    }
    sig := make([]uint64, m.numHashes)
    // Initialize with MaxUint64
    for i := 0; i < m.numHashes; i++ {
        sig[i] = math.MaxUint64
    }
    // Compute minima per hash function
    for i := 0; i < m.numHashes; i++ {
        hi := m.hashFunctions[i]
        minv := uint64(math.MaxUint64)
        for _, x := range base {
            v := hi(x)
            if v < minv {
                minv = v
            }
        }
        sig[i] = minv
    }
    return &MinHashSignature{signatures: sig, numHashes: m.numHashes}
}

// EstimateJaccardSimilarity estimates Jaccard similarity via signature agreement ratio
func (m *MinHasher) EstimateJaccardSimilarity(sig1, sig2 *MinHashSignature) float64 {
    if sig1 == nil || sig2 == nil || len(sig1.signatures) == 0 || len(sig2.signatures) == 0 {
        return 0.0
    }
    n := minInt(len(sig1.signatures), len(sig2.signatures))
    if n == 0 {
        return 0.0
    }
    match := 0
    for i := 0; i < n; i++ {
        if sig1.signatures[i] == sig2.signatures[i] {
            match++
        }
    }
    return float64(match) / float64(n)
}

func (m *MinHasher) NumHashes() int { return m.numHashes }

// Utilities

func hash64(s string) uint64 {
    h := fnv.New64a()
    _, _ = h.Write([]byte(s))
    return h.Sum64()
}

func minInt(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// (no extra helpers: keep only used symbols to satisfy linters)
