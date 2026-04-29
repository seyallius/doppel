package manual_test

import (
	"testing"

	"github.com/seyallius/doppel/manual"
)

func BenchmarkCloneMap_StringInt_50(b *testing.B) {
	src := makeStringIntMap(50)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkCloneMap_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneMap(src, manual.Identity[int])
	}
}

func BenchmarkCloneMapOf_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		_ = manual.CloneMapOf(src, manual.IdentityValue[int])
	}
}

func BenchmarkShallowCopy_StringInt_500(b *testing.B) {
	src := makeStringIntMap(500)
	b.ResetTimer()
	for range b.N {
		dst := make(map[string]int, len(src))
		for k, v := range src {
			dst[k] = v
		}
		_ = dst
	}
}
