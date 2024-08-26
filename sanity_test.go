package metrics

import (
	"testing"
)

var float64Result float64

func BenchmarkSanityUnitConversionMultiplyByOne(b *testing.B) {
	t := float64(b.N)
	m := float64(b.N) / t // i.e. 1.0 but probably not a compile-time constant
	var r float64
	for n := 0; n < b.N; n++ {
		r = float64(n) * m
	}
	float64Result = r
}

func BenchmarkSanityUnitConversionSkipIfOne(b *testing.B) {
	// about twice as slow as just multiplying by 1.0
	t := float64(b.N)
	m := float64(b.N) / t // i.e. 1.0 but probably not a compile-time constant
	var r float64
	for n := 0; n < b.N; n++ {
		if m == 1.0 {
			r = float64(n)
		} else {
			b.Fatalf("m is %f, wanted 1.0", m)
		}
	}
	float64Result = r
}

func BenchmarkSanityUnitConversionMultiplyByOneOverOneThousand(b *testing.B) {
	m := float64(1.0 / 1000.0) // i.e 0.001
	var r float64
	for n := 0; n < b.N; n++ {
		r = float64(n) * m
	}
	float64Result = r
}

func BenchmarkSanityUnitConversionDivideByOneThousand(b *testing.B) {
	// about thrice as slow as multiplying by 0.001
	d := float64(1000.0)
	var r float64
	for n := 0; n < b.N; n++ {
		r = float64(n) / d
	}
	float64Result = r
}
