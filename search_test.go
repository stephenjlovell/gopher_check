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
	"fmt"
	"math"
	"testing"
)

// // Do a fixed-depth search on a small number of positions.
// func TestSearch(t *testing.T) {
// 	setup()
// 	sum := 0
// 	depth := MAX_DEPTH - 4
// 	for i, pos := range test_positions {
// 		ResetAll() // reset all shared data structures and prepare to start a new game.
// 		current_board = ParseFENString(pos)
// 		fmt.Printf("%d. ", i+1)
// 		_, count := Search(current_board, &RepList{}, depth, MAX_TIME)
// 		sum += count
// 	}
// 	fmt.Printf("Total nodes searched: %.4f m\n", float64(sum)/1000000.0)
// 	fmt.Printf("Average Branching factor by iteration:\n")
// 	var branching float64
// 	fmt.Printf("D: 0, N: %d\n", nodes_per_iteration[0])
// 	fmt.Printf("D: 1, N: %d\n", nodes_per_iteration[1])
// 	for d := 2; d <= depth; d++ {
// 		branching = math.Pow(float64(nodes_per_iteration[d])/float64(nodes_per_iteration[1]), float64(1)/float64(d-1))
// 		fmt.Printf("D: %d, N: %d, EBF: %.4f\n", d, nodes_per_iteration[d], branching)
// 	}
// }

func TestPlayingStrength(t *testing.T) {
	setup()
	print_info = false
	depth := MAX_DEPTH-6
	test := load_epd_file("test_suites/wac_75.epd")
	var move_str string
	sum, score := 0, 0
	for _, epd := range test {
		ResetAll()
		move, count := Search(epd.brd, &RepList{}, depth, 30000)
		move_str = ToSAN(epd.brd, move)
		if correct_move(epd, move_str) {
			score += 1
			fmt.Printf("-")
		} else {
			fmt.Printf("X")
		}
		sum += count
	}
	fmt.Printf("\nTotal nodes searched: %.4f m\n", float64(sum)/1000000.0)
	fmt.Printf("Total score: %d/%d\n", score, len(test))
	fmt.Printf("Average Branching factor by iteration:\n")
	var branching float64
	for d := 2; d <= depth; d++ {
		branching = math.Pow(float64(nodes_per_iteration[d])/float64(nodes_per_iteration[1]), float64(1)/float64(d-1))
		fmt.Printf("%d ABF: %.4f\n", d, branching)
	}
}

func correct_move(epd *EPD, move_str string) bool {
	for _, a := range epd.avoid_moves {
		if move_str == a {
			return false
		}
	}
	for _, b := range epd.best_moves {
		if move_str == b {
			return true
		}
	}
	return false
}

var test_positions = [...]string{
	"3qrbk1/ppp1r2n/3pP2p/3P4/2P4P/1P3Q2/PB6/R4R1K w - -",
	"2k5/pppr4/4R3/4Q3/2pp2q1/8/PPP2PPP/6K1 w - -",
	"3r2k1/ppp2ppp/6q1/b4n2/3nQB2/2p5/P4PPP/RN3RK1 b - -",
	"kr2R3/p4r2/2pq4/2N2p1p/3P2p1/Q5P1/5P1P/5BK1 w - -",
	"r1b1k1nr/pp3pQp/4pq2/3pn3/8/P1P5/2P2PPP/R1B1KBNR w KQkq -",
	"5rk1/p5pp/8/8/2Pbp3/1P4P1/7P/4RN1K b - -",
	"rn2k1nr/pbp2ppp/3q4/1p2N3/2p5/QP6/PB1PPPPP/R3KB1R b KQkq -",
	"3r1rk1/q4ppp/p1Rnp3/8/1p6/1N3P2/PP3QPP/3R2K1 b - -",
	"r1bqr1k1/pp3ppp/1bp5/3n4/3B4/2N2P1P/PPP1B1P1/R2Q1RK1 b - -",
	"r1b1k2r/1pp1q2p/p1n3p1/3QPp2/8/1BP3B1/P5PP/3R1RK1 w kq -",
	"4r1k1/p1qr1p2/2pb1Bp1/1p5p/3P1n1R/1B3P2/PP3PK1/2Q4R w - -",
	"6k1/6p1/2p4p/4Pp2/4b1qP/2Br4/1P2RQPK/8 b - -",
	"5rk1/p4ppp/2p1b3/3Nq3/4P1n1/1p1B2QP/1PPr2P1/1K2R2R w - -",
	"1r5k/p1p3pp/8/8/4p3/P1P1R3/1P1Q1qr1/2KR4 w - -",
	"2r1b3/1pp1qrk1/p1n1P1p1/7R/2B1p3/4Q1P1/PP3PP1/3R2K1 w - -",
	"r3k2r/2p2p2/p2p1n2/1p2p3/4P2p/1PPPPp1q/1P5P/R1N2QRK b kq -",
	"3r1k2/1ppPR1n1/p2p1rP1/3P3p/4Rp1N/5K2/P1P2P2/8 w - -",
	"7k/1p4p1/7p/3P1n2/4Q3/2P2P2/PP3qRP/7K b - -",
	"r3rnk1/1pq2bb1/p4p2/3p1Pp1/3B2P1/1NP4R/P1PQB3/2K4R w - -",
	"3r1r1k/1b4pp/ppn1p3/4Pp1R/Pn5P/3P4/4QP2/1qB1NKR1 w - -",
	"r3k3/P5bp/2N1bp2/4p3/2p5/6NP/1PP2PP1/3R2K1 w q -",
	"8/8/8/1p5r/p1p1k1pN/P2pBpP1/1P1K1P2/8 b - -",
	"r5k1/pQp2qpp/8/4pbN1/3P4/6P1/PPr4P/1K1R3R b - -",
	"4rrn1/ppq3bk/3pPnpp/2p5/2PB4/2NQ1RPB/PP5P/5R1K w - -",
	"k5r1/p4b2/2P5/5p2/3P1P2/4QBrq/P5P1/4R1K1 w - -",
	"r5k1/1bp3pp/p2p4/1p6/5p2/1PBP1nqP/1PP3Q1/R4R1K b - -",
	"2kr2nr/pp1n1ppp/2p1p3/q7/1b1P1B2/P1N2Q1P/1PP1BPP1/R3K2R w KQ -",
	"1r4r1/p2kb2p/bq2p3/3p1p2/5P2/2BB3Q/PP4PP/3RKR2 b - -",
	"2rr3k/1b2bppP/p2p1n2/R7/3P4/1qB2P2/1P4Q1/1K5R w - -",
	"1nbq1r1k/3rbp1p/p1p1pp1Q/1p6/P1pPN3/5NP1/1P2PPBP/R4RK1 w - -",
}
