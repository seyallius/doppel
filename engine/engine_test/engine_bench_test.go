package engine_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/seyallius/doppel/engine"
)

/*

BenchmarkEngine_PlainStruct            2272365	     520.8 ns/op	      88 B/op	       5 allocs/op
BenchmarkEngine_NestedStruct         	736201	      1586 ns/op	     360 B/op	      14 allocs/op
BenchmarkEngine_LargeSlice           	 12734	     93976 ns/op	   16240 B/op	    1003 allocs/op
BenchmarkEngine_LargeMap                  8905	    126708 ns/op	   63640 B/op	    2005 allocs/op
BenchmarkEngine_WithTypeLookup_Hit    14146173	     83.50 ns/op	      48 B/op	       1 allocs/op
BenchmarkEngine_SelfClonable           1671700	     715.6 ns/op	     232 B/op	       8 allocs/op
BenchmarkEngine_ShallowBaseline      975549486	     1.229 ns/op	       0 B/op	       0 allocs/op

*/

func BenchmarkEngine_PlainStruct(b *testing.B) {
	src := plainStruct{Name: "bench", Value: 42, Score: 3.14, Active: true}
	eng := engine.New(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_NestedStruct(b *testing.B) {
	src := nestedStruct{
		Meta:  plainStruct{Name: "parent", Value: 1},
		Child: &plainStruct{Name: "child", Value: 2},
		Count: 10,
	}
	eng := engine.New(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_LargeSlice(b *testing.B) {
	src := make([]int, 1000)
	for idx := range src {
		src[idx] = idx
	}
	eng := engine.New(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_LargeMap(b *testing.B) {
	src := make(map[string]int, 500)
	for idx := range 500 {
		src[fmt.Sprintf("key_%d", idx)] = idx
	}
	eng := engine.New(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_WithTypeLookup_Hit(b *testing.B) {
	lookup := newStubbedLookup()
	lookup.register(
		reflect.TypeOf(plainStruct{}),
		func(src reflect.Value) (reflect.Value, error) {
			s := src.Interface().(plainStruct)
			return reflect.ValueOf(plainStruct{Name: s.Name, Value: s.Value}), nil
		},
	)

	src := plainStruct{Name: "bench", Value: 42}
	eng := engine.New(lookup)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_SelfClonable(b *testing.B) {
	src := &selfClonable{Data: "bench"}
	eng := engine.New(nil)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, _ = eng.Clone(reflect.ValueOf(src))
	}
}

func BenchmarkEngine_ShallowBaseline(b *testing.B) {
	src := plainStruct{Name: "bench", Value: 42, Score: 3.14, Active: true}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		dst := src
		_ = dst
	}
}
