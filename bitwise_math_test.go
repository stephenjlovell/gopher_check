//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//-----------------------------------------------------------------------------------

package main

import (
	"testing"
)

var result int // hacks to make sure compiler doesn't eliminate func under test.
var bb_result BB

//
// func BenchmarkScanDown(b *testing.B) {
// 	setup()
// 	test := load_epd_file("test_suites/wac_300.epd")
//
// 	b.ResetTimer()
// 	var sq int
// 	for _, epd := range test {
// 		occ := epd.brd.AllOccupied()
// 		sq = epd.brd.KingSq(epd.brd.c)
// 		for i := 0; i < b.N; i++ {
// 			bb_result = scan_down(occ, SW, sq)
// 			bb_result = scan_down(occ, SOUTH, sq)
// 			bb_result = scan_down(occ, SE, sq)
// 		}
// 	}
// }
//
// func BenchmarkScanUp(b *testing.B) {
// 	setup()
// 	test := load_epd_file("test_suites/wac_300.epd")
//
// 	b.ResetTimer()
// 	var sq int
// 	for _, epd := range test {
// 		occ := epd.brd.AllOccupied()
// 		sq = epd.brd.KingSq(epd.brd.c)
// 		for i := 0; i < b.N; i++ {
// 			bb_result = scan_up(occ, NW, sq)
// 			bb_result = scan_up(occ, NORTH, sq)
// 			bb_result = scan_up(occ, NE, sq)
// 		}
// 	}
// }
//

func BenchmarkPopCount(b *testing.B) {
	var bb BB
	test := load_epd_file("test_suites/wac_300.epd")
	b.ResetTimer()

	for _, epd := range test {
		bb = epd.brd.occupied[WHITE]
		for i := 0; i < b.N; i++ {
			result = pop_count(bb)
		}
	}

}

func BenchmarkLSB(b *testing.B) {
	var bb BB
	test := load_epd_file("test_suites/wac_300.epd")
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
