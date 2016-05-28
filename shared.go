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
	"flag"
	"fmt"
	// "math/rand"
	"runtime"
	// "sync/atomic"
)

const (
	INF      = 10000            // an arbitrarily large score used for initial bounds
	NO_SCORE = INF - 1          // sentinal value indicating a meaningless score.
	MATE     = NO_SCORE - 1     // maximum checkmate score (i.e. mate in 0)
	MIN_MATE = MATE - MAX_STACK // minimum possible checkmate score (mate in MAX_STACK)
) // total value of all starting pieces for one side: 9006

const ( // color
	BLACK = iota
	WHITE
)
const ( // type
	PAWN = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
	EMPTY // no piece located at this square
)
const ( // direction codes (0...8)
	NW = iota
	NE
	SE
	SW
	NORTH // 4
	EAST
	SOUTH
	WEST // 7
	DIR_INVALID
)

var opposite_dir = [16]int{SE, SW, NW, NE, SOUTH, WEST, NORTH, EAST, DIR_INVALID}

// type SafeCounter int64
//
// func (c *SafeCounter) Add(i int64) int64 {
// 	return atomic.AddInt64((*int64)(c), i)
// }
//
// func (c *SafeCounter) Get() int64 {
// 	return atomic.LoadInt64((*int64)(c))
// }

var middle_rows BB

var mask_of_length [65]uint64

var row_masks, column_masks [8]BB

var ray_masks [8][64]BB

var pawn_isolated_masks, pawn_side_masks, pawn_doubled_masks, knight_masks, bishop_masks, rook_masks,
	queen_masks, king_masks, sq_mask_on, sq_mask_off [64]BB

var intervening, line_masks [64][64]BB

var castle_queenside_intervening, castle_kingside_intervening [2]BB

var pawn_attack_masks, pawn_passed_masks, pawn_attack_spans, pawn_front_spans,
	pawn_stop_masks, king_zone_masks, king_shield_masks [2][64]BB

var pawn_stop_sq, pawn_promote_sq [2][64]int

const (
	OFF_SINGLE = iota
	OFF_DOUBLE
	OFF_LEFT
	OFF_RIGHT
)

// var piece_values = [8]int{100, 325, 325, 500, 975}
var piece_values = [8]int{100, 320, 333, 510, 880, 5000} // default piece values

// var promote_values = [8]int{0, 225, 225, 400, 875}
var promote_values = [8]int{0, 220, 233, 410, 780, 0, 0, 0} // piece_values[pc] - PAWN
var pawn_from_offsets = [2][4]int{{8, 16, 9, 7}, {-8, -16, -7, -9}}
var knight_offsets = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
var bishop_offsets = [4]int{7, 9, -7, -9}
var rook_offsets = [4]int{8, 1, -8, -1}
var king_offsets = [8]int{-9, -7, 7, 9, -8, -1, 1, 8}
var pawn_attack_offsets = [4]int{9, 7, -9, -7}
var pawn_advance_offsets = [4]int{8, 16, -8, -16}

var directions [64][64]int

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}
func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

func on_board(sq int) bool       { return 0 <= sq && sq <= 63 }
func row(sq int) int             { return sq >> 3 }
func column(sq int) int          { return sq & 7 }
func Square(row, column int) int { return (row * 8) + column }

func manhattan_distance(from, to int) int {
	return abs(row(from)-row(to)) + abs(column(from)-column(to))
}

var chebyshev_distance_table [64][64]int

func chebyshev_distance(from, to int) int {
	return chebyshev_distance_table[from][to]
}

func setup_chebyshev_distance() {
	for from := 0; from < 64; from++ {
		for to := 0; to < 64; to++ {
			chebyshev_distance_table[from][to] = max(abs(row(from)-row(to)), abs(column(from)-column(to)))
		}
	}
}

func assert(statement bool, failure_message string) {
	if !statement {
		panic("\nassertion failed: " + failure_message + "\n")
	}
}

func setup() {
	num_cpu := runtime.NumCPU()
	runtime.GOMAXPROCS(num_cpu)
	setup_chebyshev_distance()
	setup_masks()
	setup_magic_move_gen()
	setup_eval()
	setup_rand()
	setup_zobrist()
	reset_main_tt()
	setup_load_balancer(num_cpu)
}

func print_name() {
	fmt.Println("------------------------------------------------------------------")
	fmt.Println("\u265B GopherCheck \u265B\nCopyright \u00A9 2014 Stephen J. Lovell")
	fmt.Println("------------------------------------------------------------------\n")
}

var profile_flag = flag.Bool("profile", false, "Set profile=true to run profiler on test suite.")

func main() {
	print_name()
	setup()
	flag.Parse()

	if *profile_flag {
		RunProfiledTestSuite("test_suites/wac_300.epd", 9, 6000)
	} else {
		ReadUCICommand()
	}
}
