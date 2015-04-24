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
	"math/rand"
	"runtime"
)

const (
	INF  = 10000 // an arbitrarily large score used to signal checkmate.
	NO_SCORE = INF - 1 // score used to signal
	MATE = NO_SCORE - 1 // maximum checkmate score (i.e. mate in 0)
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

var mask_of_length [65]uint64

var row_masks [8]BB
var column_masks [8]BB
var ray_masks [8][64]BB

var pawn_isolated_masks, pawn_side_masks [64]BB

var intervening [64][64]BB
var castle_queenside_intervening, castle_kingside_intervening [2]BB

var knight_masks, bishop_masks, rook_masks, queen_masks, king_masks, sq_mask_on, sq_mask_off [64]BB
var pawn_attack_masks, pawn_blocked_masks, pawn_passed_masks, king_zone_masks, king_shield_masks [2][64]BB

const (
	OFF_SINGLE = iota
	OFF_DOUBLE
	OFF_LEFT
	OFF_RIGHT
)

var piece_values = [8]int{100, 320, 333, 510, 880, 5000} // default piece values
var endgame_count_values = [8]uint8{1, 3, 3, 5, 9}       // piece values used to determine endgame status

var promote_values = [8]int{0, 220, 233, 410, 780, 0, 0, 0}

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
func round(x float64) int {
	if x >= 0 {
		return int(x + 0.5)
	} else {
		return int(x - 0.5)
	}
}

func on_board(sq int) bool       { return 0 <= sq && sq <= 63 }
func row(sq int) int             { return sq >> 3 }
func column(sq int) int          { return sq & 7 }
func Square(row, column int) int { return (row * 8) + column }

func manhattan_distance(from, to int) int {
	return abs(row(from)-row(to)) + abs(column(from)-column(to))
}
func chebyshev_distance(from, to int) int {
	return max(abs(row(from)-row(to)), abs(column(from)-column(to)))
}

func assert(statement bool, failure_message string) {
	if !statement {
		panic("\nassertion failed: " + failure_message + "\n")
	}
}

func setup() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(4129246945) // keep the same seed each time for debugging purposes.

	setup_bitwise_ops()
	setup_masks()

	setup_eval()
	setup_zobrist()
	setup_main_tt()

	fmt.Println("\n------------------------------------------------------------------")
	fmt.Println("\u265B GopherCheck \u265B\nCopyright \u00A9 2014 Stephen J. Lovell")
	fmt.Println("------------------------------------------------------------------\n")
}

func main() {
	setup()
	ReadUCICommand()
}
