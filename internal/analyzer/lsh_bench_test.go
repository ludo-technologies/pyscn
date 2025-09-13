package analyzer

import (
	"strconv"
	"testing"
)

func BenchmarkMinHasher_Signature(b *testing.B) {
	mh := NewMinHasher(128)
	feats := make([]string, 0, 200)
	for i := 0; i < 200; i++ {
		feats = append(feats, string(rune('a')+rune(i%26))+"_feat_"+string(rune('A')+rune(i%26)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mh.ComputeSignature(feats)
	}
}

func BenchmarkLSHIndex_Candidates(b *testing.B) {
	mh := NewMinHasher(128)
	idx := NewLSHIndex(32, 4)
	// build 1000 signatures with small random overlap
	for i := 0; i < 1000; i++ {
		feats := []string{}
		for j := 0; j < 30; j++ {
			feats = append(feats, string(rune('a')+rune((i+j)%26))+"/"+string(rune('A')+rune((i+2*j)%26)))
		}
		_ = idx.AddFragment("id-"+strconv.Itoa(i), mh.ComputeSignature(feats))
	}
	_ = idx.BuildIndex()
	// Query signature similar to mid entries
	qfeats := []string{}
	for j := 0; j < 30; j++ {
		qfeats = append(qfeats, string(rune('a')+rune((500+j)%26))+"/"+string(rune('A')+rune((500+2*j)%26)))
	}
	qsig := mh.ComputeSignature(qfeats)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = idx.FindCandidates(qsig)
	}
}
