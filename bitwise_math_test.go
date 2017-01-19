//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import (
	"testing"
)

var result int // hacks to make sure compiler doesn't eliminate func under test.
var bbResult BB

//
// func BenchmarkScanDown(b *testing.B) {
// 	setup()
// 	test := loadEpdFile("test_suites/wac300.epd")
//
// 	b.ResetTimer()
// 	var sq int
// 	for _, epd := range test {
// 		occ := epd.brd.AllOccupied()
// 		sq = epd.brd.KingSq(epd.brd.c)
// 		for i := 0; i < b.N; i++ {
// 			bbResult = scanDown(occ, SW, sq)
// 			bbResult = scanDown(occ, SOUTH, sq)
// 			bbResult = scanDown(occ, SE, sq)
// 		}
// 	}
// }
//
// func BenchmarkScanUp(b *testing.B) {
// 	setup()
// 	test := loadEpdFile("test_suites/wac300.epd")
//
// 	b.ResetTimer()
// 	var sq int
// 	for _, epd := range test {
// 		occ := epd.brd.AllOccupied()
// 		sq = epd.brd.KingSq(epd.brd.c)
// 		for i := 0; i < b.N; i++ {
// 			bbResult = scanUp(occ, NW, sq)
// 			bbResult = scanUp(occ, NORTH, sq)
// 			bbResult = scanUp(occ, NE, sq)
// 		}
// 	}
// }
//

func BenchmarkPopCount(b *testing.B) {
	var bb BB
	test := loadEpdFile("test_suites/wac300.epd")
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
	test := loadEpdFile("test_suites/wac300.epd")
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
	bb := rng.RandomBB(BB(MAX_RAND))
	for i := 0; i < b.N; i++ {
		result = lsb(bb)
	}
}
