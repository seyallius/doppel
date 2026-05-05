package manual_test

import (
	"testing"

	"github.com/seyallius/doppel/manual"
)

func BenchmarkCloneMap_StringInt_50(b *testing.B) {
	src := makeStringIntMap(50)
	b.ResetTimer()
	for b.Loop() {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkCloneMap_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for b.Loop() {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkShallowCopy_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for b.Loop() {
		dst := make(map[string]int, len(src))
		for k, v := range src {
			dst[k] = v
		}
		_ = dst
	}
}
