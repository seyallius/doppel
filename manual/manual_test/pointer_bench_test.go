package manual_test

import (
	"testing"

	"github.com/seyallius/doppel/manual"
)

func BenchmarkClonePointer_Int(b *testing.B) {
	src := intPointer(42)
	b.ResetTimer()
	for b.Loop() {
		_, _ = manual.ClonePointer(src, manual.Identity[int])
	}
}

func BenchmarkShallowPointerCopy_Int(b *testing.B) {
	src := intPointer(42)
	b.ResetTimer()
	for b.Loop() {
		dst := src // shallow: same address
		_ = dst
	}
}
