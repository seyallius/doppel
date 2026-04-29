package manual_test

import (
	"testing"

	"github.com/seyallius/doppel/manual"
)

func BenchmarkCloneSlice_Strings_10(b *testing.B) {
	src := makeStringSlice(10)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSlice_Strings_100(b *testing.B) {
	src := makeStringSlice(100)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSlice_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[string])
	}
}

func BenchmarkCloneSliceOf_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		_ = manual.CloneSliceOf(src, manual.IdentityValue[string])
	}
}

func BenchmarkShallowCopy_Strings_1000(b *testing.B) {
	src := makeStringSlice(1000)
	b.ResetTimer()
	for range b.N {
		dst := make([]string, len(src))
		copy(dst, src)
		_ = dst
	}
}

func BenchmarkCloneSlice_Ints_1000(b *testing.B) {
	src := makeIntSlice(1000)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.CloneSlice(src, manual.Identity[int])
	}
}

func BenchmarkShallowCopy_Ints_1000(b *testing.B) {
	src := makeIntSlice(1000)
	b.ResetTimer()
	for range b.N {
		dst := make([]int, len(src))
		copy(dst, src)
		_ = dst
	}
}
