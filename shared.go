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
	"math/rand"
	"runtime"
)

const (
	INF = 100000 // an arbitrarily large score used to signal checkmate.
)

const (
	MAX_DEPTH = 10
	EXT_MAX   = 5
	MAX_PLY   = MAX_DEPTH + EXT_MAX
)

// Transposition Table, Killer moves, History Table should be shared by all goroutines.
type SHARED struct {
	killer_moves KTable
	history      HTable
	// add transposition table
	side_to_move [MAX_PLY]uint8
	enemy        [MAX_PLY]uint8
}

// Each goroutine gets its own explicit stack to store the local pv and information needed to unmake moves.

type StackItem struct {
	pv             []MV
	castle         uint8
	enp_target     int
	halfmove_clock uint8
}

type Stack []StackItem

// type
const (
	PAWN = iota
	KNIGHT
	BISHOP
	ROOK
	QUEEN
	KING
)

// color
const (
	BLACK = iota
	WHITE
)

// square/ID codes (0...12)
const (
	B_PAWN = iota
	W_PAWN
	B_KNIGHT
	W_KNIGHT
	B_BISHOP
	W_BISHOP
	B_ROOK
	W_ROOK
	B_QUEEN
	W_QUEEN
	B_KING
	W_KING
	PC_EMPTY
)

const ( // direction codes (0...8)
	NW = iota
	NE
	SE
	SW
	NORTH
	EAST
	SOUTH
	WEST
	DIR_INVALID
)

var mask_of_length [64]BB

var row_masks [8]BB
var column_masks [8]BB
var ray_masks [8][64]BB

var pawn_attack_masks, pawn_passed_masks [2][64]BB

var pawn_isolated_masks, pawn_side_masks [64]BB

var intervening [64][64]BB
var castle_queenside_intervening, castle_kingside_intervening [2]BB

var knight_masks, bishop_masks, rook_masks, queen_masks, king_masks, sq_mask_on, sq_mask_off [64]BB

var pawn_from_offsets = [2][4]int{{8, 16, 9, 7}, {-8, -16, -7, -9}}
var knight_offsets = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
var bishop_offsets = [4]int{7, 9, -7, -9}
var rook_offsets = [4]int{8, 1, -8, -1}
var king_offsets = [8]int{-9, -7, 7, 9, -8, -1, 1, 8}
var pawn_attack_offsets = [4]int{9, 7, -9, -7}
var pawn_advance_offsets = [4]int{8, 16, -8, -16}
var pawn_enpassant_offsets = [2]int{1, -1}

var piece_values = [6]int{100, 320, 333, 510, 880, 10000} // default piece values

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

func on_board(sq int) bool { return 0 <= sq && sq <= 63 }
func row(sq int) int       { return sq >> 3 }
func column(sq int) int    { return sq & 7 }

func manhattan_distance(from, to int) int {
	return abs(row(from)-row(to)) + abs(column(from)-column(to))
}
func chebyshev_distance(from, to int) int {
	return max(abs(row(from)-row(to)), abs(column(from)-column(to)))
}

func lsb(b BB) int                     { return 0 }
func msb(b BB) int                     { return 0 }
func furthest_forward(c int, b BB) int { return 0 }
func pop_count(b BB) int               { return 0 }

// #define lsb(bitboard) (__builtin_ctzl(bitboard))
// #define msb(bitboard) (63-__builtin_clzl(bitboard))
// #define furthest_forward(color, bitboard) (color ? lsb(bitboard) : msb(bitboard))
// #define pop_count(bitboard) (__builtin_popcountl(bitboard))

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(9) // keep the same seed each time for debugging purposes.
	setup_zobrist()

	setup_masks()
	setup_bonus_table()

	fmt.Println("Hello Chess World")

}
