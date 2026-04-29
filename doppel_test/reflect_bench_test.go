// Package doppel_test. doppel_reflection_bench_test.go compares doppel clone performance against reflection-based deep cloning.
package doppel_test

import (
	"strconv"
	"testing"
)

var benchmarkSink any

func BenchmarkDoppelVsReflect_Score(b *testing.B) {
	score := Score{
		Label: "quality",
		Value: 98.75,
	}

	b.Run("Doppel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned, err := cloneScore(score)
			if err != nil {
				b.Fatal(err)
			}

			benchmarkSink = cloned
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned := reflectDeepClone[Score](score)

			benchmarkSink = cloned
		}
	})
}

func BenchmarkDoppelVsReflect_User(b *testing.B) {
	user := newUser()

	b.Run("Doppel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned, err := user.Clone()
			if err != nil {
				b.Fatal(err)
			}

			benchmarkSink = cloned
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned := reflectDeepClone[*User](user)

			benchmarkSink = cloned
		}
	})
}

func BenchmarkDoppelVsReflect_Order(b *testing.B) {
	order := newOrder(newUser())

	b.Run("Doppel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned, err := order.Clone()
			if err != nil {
				b.Fatal(err)
			}

			benchmarkSink = cloned
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned := reflectDeepClone[*Order](order)

			benchmarkSink = cloned
		}
	})
}

func BenchmarkDoppelVsReflect_UserLargeSlice(b *testing.B) {
	user := newUser()
	user.Tags = make([]string, 1_000)
	user.Aliases = make([]string, 1_000)

	for i := range user.Tags {
		user.Tags[i] = "tag"
		user.Aliases[i] = "alias"
	}

	b.Run("Doppel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned, err := user.Clone()
			if err != nil {
				b.Fatal(err)
			}

			benchmarkSink = cloned
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned := reflectDeepClone[*User](user)

			benchmarkSink = cloned
		}
	})
}

func BenchmarkDoppelVsReflect_UserLargeMap(b *testing.B) {
	user := newUser()
	user.Scores = make(map[string]int, 1_000)

	for i := range 1_000 {
		// Generate repeated-ish keys; it only creates up to 26 unique keys
		//user.Scores[string(rune('a'+i%26))] = i

		user.Scores["score-"+strconv.Itoa(i)] = i
	}

	b.Run("Doppel", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned, err := user.Clone()
			if err != nil {
				b.Fatal(err)
			}

			benchmarkSink = cloned
		}
	})

	b.Run("Reflect", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			cloned := reflectDeepClone[*User](user)

			benchmarkSink = cloned
		}
	})
}
