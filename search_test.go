//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

import "testing"

// func BenchmarkSearch(b *testing.B) {
// 	setup()
// 	verbose = false
// 	depth := MAX_DEPTH
// 	for n := 0; n < b.N; n++ {
// 		brd := ParseFENString("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
// 		Start(brd, depth, 4000)
// 		fmt.Printf(".")
// 	}
// }

func TestPlayingStrength(t *testing.T) {
	printName()
	depth := MAX_DEPTH
	timeout := 2000
	RunTestSuite("test_suites/wac_300.epd", depth, timeout)
}
