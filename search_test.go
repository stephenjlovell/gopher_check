//-----------------------------------------------------------------------------------
// Copyright (c) 2014 Stephen J. Lovell
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
	"testing"
)

var test_positions = [...]string{
	"3r2k1/ppp2ppp/6q1/b4n2/3nQB2/2p5/P4PPP/RN3RK1 b - -",
	"r4rk1/1bR1bppp/4pn2/1p2N3/1P6/P3P3/4BPPP/3R2K1 b - -",
	"kr2R3/p4r2/2pq4/2N2p1p/3P2p1/Q5P1/5P1P/5BK1 w - -",
	"1r3b1k/p4rpp/4pp2/3q4/2ppbPPQ/6RK/PP5P/2B1NR2 b - -",
	"r1b1k1nr/pp3pQp/4pq2/3pn3/8/P1P5/2P2PPP/R1B1KBNR w KQkq -",
	"2rr3k/pp3pp1/1nnqbN1p/3pN3/2pP4/2P3Q1/PPB4P/R4RK1 w - -",
	"3R1rk1/8/5Qpp/2p5/2P1p1q1/P3P3/1P2PK2/8 b - -",
	"2b3k1/4rrpp/p2p4/2pP2RQ/1pP1Pp1N/1P3P1P/1q6/6RK w - -",
	"r1q3rk/1ppbb1p1/4Np1p/p3pP2/P3P3/2N4R/1PP1Q1PP/3R2K1 w - -",
	"6k1/5p1p/2bP2pb/4p3/2P5/1p1pNPPP/1P1Q1BK1/1q6 b - -",
}

func TestSearch(t *testing.T) {
	setup()
	sum := 0
	for _, pos := range test_positions {
		ResetAll() // reset all shared data structures and prepare to start a new game.
		current_board = ParseFENString(pos)
		// current_board.Print()
		move, count := Search(current_board, make([]Move, 0), MAX_DEPTH, MAX_TIME)
		fmt.Printf("bestmove %s\n\n", move.ToString())
		sum += count
	}
	fmt.Printf("Total nodes searched: %.4f m\n", float64(sum)/1000000.0)
}

// func TestBoardCopy(t *testing.T) {
//   setup()
//   brd := StartPos()
//   result := make(chan *Board)
//   counter := 0
//   for i := 8; i < 16; i++ {
//     copy := brd.Copy()
//     move := NewMove(i, i+8, PAWN, EMPTY, EMPTY)
//     counter++
//     go func(i int) {
//       make_move(copy, move)
//       fmt.Printf("%d ", i)
//       result <- copy
//     }(i)
//   }
//   fmt.Printf("\n")
//   for counter > 0 {
//     select {
//     case r := <-result:
//       r.Print()
//       counter--
//     }
//   }
//   brd.Print()
//   fmt.Println("TestBoardCopy complete.")
// }
