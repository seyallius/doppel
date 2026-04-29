package manual_test

import (
	"testing"

	"github.com/seyallius/doppel/manual"
)

func BenchmarkClonePointer_Int(b *testing.B) {
	src := intPointer(42)
	b.ResetTimer()
	for range b.N {
		_, _ = manual.ClonePointer(src, manual.Identity[int])
	}
}

func BenchmarkClonePointerOf_Int(b *testing.B) {
	src := intPointer(42)
	b.ResetTimer()
	for range b.N {
		_ = manual.ClonePointerOf(src, manual.IdentityValue[int])
	}
}

func BenchmarkShallowPointerCopy_Int(b *testing.B) {
	src := intPointer(42)
	b.ResetTimer()
	for range b.N {
		dst := src // shallow: same address
		_ = dst
	}
}
