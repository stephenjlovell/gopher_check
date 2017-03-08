//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"fmt"
	"testing"
)

var result int // hacks to make sure compiler doesn't eliminate func under test.

func BenchmarkPopCount(b *testing.B) {
	var bb BB
	test, err := LoadEpdFile("test_suites/wac_300.epd")
	if err != nil {
		fmt.Print(err)
		return
	}
	b.ResetTimer()
	for _, epd := range test {
		bb = epd.brd.occupied[WHITE]
		for i := 0; i < b.N; i++ {
			result = popCount(bb)
		}
	}

}

func BenchmarkLSB(b *testing.B) {
	var bb BB
	test, err := LoadEpdFile("test_suites/wac_300.epd")
	if err != nil {
		fmt.Print(err)
		return
	}
	b.ResetTimer()
	for _, epd := range test {
		bb = epd.brd.occupied[WHITE]
		for i := 0; i < b.N; i++ {
			result = lsb(bb)
		}
	}
}

func BenchmarkLSBRand(b *testing.B) {
	rng := NewRngKiss(74)
	bb := rng.RandomBB(BB((1 << 32) - 1))
	for i := 0; i < b.N; i++ {
		result = lsb(bb)
	}
}
