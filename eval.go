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
)

var non_king_value, endgame_value int

var passed_pawn_bonus = [2][8]int{
	{0, 49, 28, 16, 9, 5, 3, 0},
	{0, 3, 5, 9, 16, 16, 28, 49},
}

var promote_row = [2][2]int{
	{1, 2},
	{6, 5},
}

const isolated_pawn_penalty int = -5
const double_pawn_penalty int = -10
const pawn_duo_bonus int = 3

var main_pst = [2][6][64]int{
	{ // Black
		// Pawn
		{0, 0, 0, 0, 0, 0, 0, 0,
			-1, 1, 1, 1, 1, 1, 1, -1,
			-2, 0, 1, 2, 2, 1, 0, -2,
			-3, -1, 2, 10, 10, 2, -1, -3,
			-4, -2, 4, 14, 14, 4, -2, -4,
			-5, -3, 0, 9, 9, 0, -3, -5,
			-6, -4, 0, -20, -20, 0, -4, -6,
			0, 0, 0, 0, 0, 0, 0, 0},
		// Knight
		{-8, -8, -6, -6, -6, -6, -8, -8,
			-8, 0, 0, 0, 0, 0, 0, -8,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-8, 0, 1, 2, 2, 1, 0, -8,
			-10, -12, -6, -6, -6, -6, -12, -10},
		// Bishop
		{-3, -3, -3, -3, -3, -3, -3, -3,
			-3, 0, 0, 0, 0, 0, 0, -3,
			-3, 0, 2, 4, 4, 2, 0, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 1, 2, 4, 4, 2, 1, -3,
			-3, 2, 1, 1, 1, 1, 2, -3,
			-3, -3, -10, -3, -3, -10, -3, -3},
		// Rook
		{4, 4, 4, 4, 4, 4, 4, 4,
			16, 16, 16, 16, 16, 16, 16, 16,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			0, 0, 0, 2, 2, 0, 0, 0},
		// Queen
		{0, 0, 0, 1, 1, 0, 0, 0,
			0, 0, 1, 2, 2, 1, 0, 0,
			0, 1, 2, 2, 2, 2, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 1, 2, 2, 1, 1, 0,
			0, 0, 1, 1, 1, 1, 0, 0,
			-6, -6, -6, -6, -6, -6, -6, -6},
	}, // White
	{ // Pawn
		{0, 0, 0, 0, 0, 0, 0, 0,
			-6, -4, 0, -20, -20, 0, -4, -6,
			-5, -3, 0, 9, 9, 0, -3, -5,
			-4, -2, 4, 14, 14, 4, -2, -4,
			-3, -1, 2, 10, 10, 2, -1, -3,
			-2, 0, 1, 2, 2, 1, 0, -2,
			-1, 1, 1, 1, 1, 1, 1, -1,
			0, 0, 0, 0, 0, 0, 0, 0},
		// Knight
		{-10, -12, -6, -6, -6, -6, -12, -10,
			-8, 0, 1, 2, 2, 1, 0, -8,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 8, 8, 4, 0, -6,
			-6, 0, 4, 4, 4, 4, 0, -6,
			-8, 0, 0, 0, 0, 0, 0, -8,
			-8, -8, -6, -6, -6, -6, -8, -8},
		// Bishop
		{-3, -3, -10, -3, -3, -10, -3, -3,
			-3, 2, 1, 1, 1, 1, 2, -3,
			-3, 1, 2, 4, 4, 2, 1, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 0, 4, 5, 5, 4, 0, -3,
			-3, 0, 2, 4, 4, 2, 0, -3,
			-3, 0, 0, 0, 0, 0, 0, -3,
			-3, -3, -3, -3, -3, -3, -3, -3},
		// Rook
		{0, 0, 0, 2, 2, 0, 0, 0,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			-4, 0, 0, 0, 0, 0, 0, -4,
			16, 16, 16, 16, 16, 16, 16, 16,
			4, 4, 4, 4, 4, 4, 4, 4},
		// Queen
		{-6, -6, -6, -6, -6, -6, -6, -6,
			0, 0, 1, 1, 1, 1, 0, 0,
			0, 1, 1, 2, 2, 1, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 2, 3, 3, 2, 1, 0,
			0, 1, 2, 2, 2, 2, 1, 0,
			0, 0, 1, 2, 2, 1, 0, 0,
			0, 0, 0, 1, 1, 0, 0, 0},
	},
}

var mirror = [64]int{
	56, 57, 58, 59, 60, 61, 62, 63, // Used to create a mirror image of the base PST
	48, 49, 50, 51, 52, 53, 54, 55, // during initialization.
	40, 41, 42, 43, 44, 45, 46, 47,
	32, 33, 34, 35, 36, 37, 38, 39,
	24, 25, 26, 27, 28, 29, 30, 31,
	16, 17, 18, 19, 20, 21, 22, 23,
	8, 9, 10, 11, 12, 13, 14, 15,
	0, 1, 2, 3, 4, 5, 6, 7}

var king_pst = [2][2][64]int{{ // Black // False
	{-52, -50, -50, -50, -50, -50, -50, -52, // In early game, encourage the king to stay on back
		-50, -48, -48, -48, -48, -48, -48, -50, // row defended by friendly pieces.
		-48, -46, -46, -46, -46, -46, -46, -48,
		-46, -44, -44, -44, -44, -44, -44, -46,
		-44, -42, -42, -42, -42, -42, -42, -44,
		-42, -40, -40, -40, -40, -40, -40, -42,
		-16, -15, -20, -20, -20, -20, -15, -16,
		0, 20, 30, -30, 0, -20, 30, 20},
	{ // True
		-30, -20, -10, 0, 0, -10, -20, -30, // In end game (when few friendly pieces are available
		-20, -10, 0, 10, 10, 0, -10, -20, // to protect king), the king should move toward the center
		-10, 0, 10, 20, 20, 10, 0, -10, // and avoid getting trapped in corners.
		0, 10, 20, 30, 30, 20, 10, 0,
		0, 10, 20, 30, 30, 20, 10, 0,
		-10, 0, 10, 20, 20, 10, 0, -10,
		-20, -10, 0, 10, 10, 0, -10, -20,
		-30, -20, -10, 0, 0, -10, -20, -30},
}, { // White // False
	{0, 20, 30, -30, 0, -20, 30, 20,
		-16, -15, -20, -20, -20, -20, -15, -16,
		-42, -40, -40, -40, -40, -40, -40, -42,
		-44, -42, -42, -42, -42, -42, -42, -44,
		-46, -44, -44, -44, -44, -44, -44, -46,
		-48, -46, -46, -46, -46, -46, -46, -48,
		-50, -48, -48, -48, -48, -48, -48, -50,
		-52, -50, -50, -50, -50, -50, -50, -52},
	{ // True
		-30, -20, -10, 0, 0, -10, -20, -30,
		-20, -10, 0, 10, 10, 0, -10, -20,
		-10, 0, 10, 20, 20, 10, 0, -10,
		0, 10, 20, 30, 30, 20, 10, 0,
		0, 10, 20, 30, 30, 20, 10, 0,
		-10, 0, 10, 20, 20, 10, 0, -10,
		-20, -10, 0, 10, 10, 0, -10, -20,
		-30, -20, -10, 0, 0, -10, -20, -30}}}

// adjust value of knights and rooks based on number of pawns in play.
var knight_pawns = [16]int{-20, -16, -12, -8, -4, 0, 4, 8, 12}
var rook_pawns = [16]int{16, 12, 8, 4, 2, 0, -2, -4, -8}

// max mobility bonus/penalty should be 2.5% of piece value:
// 8.0, 8.325000000000001, 12.75, 22.0
// max knight mobility = 8, avg 2
// max bishop/rook mobility = 14, avg 3
// max queen mobility = 28, avg 4
var knight_mobility = [16]int{-6, -3, 0, 1, 2, 3, 4, 5, 8, 0, 0, 0, 0, 0, 0}
var bishop_mobility = [16]int{-8, -4, -2, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}
var rook_mobility = [16]int{-3, -2, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 8, 8, 8, 8}
var queen_mobility = [32]int{-10, -6, -3, -1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16, 16}

func setup_eval() {
	setup_eval_constants()
}

func setup_eval_constants() {
	non_king_value = piece_values[PAWN]*8 + piece_values[KNIGHT]*2 + piece_values[BISHOP]*2 +
		piece_values[ROOK]*2 + piece_values[QUEEN]
	endgame_value = piece_values[KING] - (non_king_value / 4)
}

var highest_placement, lowest_placement int

func evaluate(brd *Board, alpha, beta int) int {
	c, e := brd.c, brd.Enemy()
	material := int(brd.material[c] - brd.material[e])
	// lazy evaluation: if material balance is already outside the search window by an amount that outweighs
	// the largest possible placement evaluation, return the material as an approximate evaluation.
	// This prevents the engine from wasting a lot of time evaluating unrealistic positions.
	if material+piece_values[ROOK] < alpha || material-piece_values[ROOK] > beta {
		return material
	}
	return material + adjusted_placement(brd, c, e) - adjusted_placement(brd, e, c)
}

// current ranges (approximate):
// PST bonus: {-152, 82}
// Mobility: { -39, 55}
// Pawn structure: { -65, 168 }
// { -256, 305} => +/- 561

func adjusted_placement(brd *Board, c, e uint8) int {

	friendly := brd.Placement(c)
	available := ^friendly
	occ := brd.Occupied()

	var unguarded BB // a bitmap of squares undefended by enemy pawns
	if c > 0 {       // white to move
		unguarded = ^(((brd.pieces[e][PAWN] & (^column_masks[0])) >> 9) | ((brd.pieces[e][PAWN] & (^column_masks[7])) >> 7))
	} else { // black to move
		unguarded = ^(((brd.pieces[e][PAWN] & (^column_masks[0])) << 7) | ((brd.pieces[e][PAWN] & (^column_masks[7])) << 9))
	}
	var sq, mobility, placement int
	var b BB
	enemy_king_sq := furthest_forward(e, brd.pieces[e][KING])

	if enemy_king_sq > 63 || enemy_king_sq < 0 {
		brd.Print()
		fmt.Printf("%d, %d", brd.material[c], brd.material[e])
		fmt.Printf("Invalid King Square: %d\n", enemy_king_sq)
	}

	pawn_count := pop_count(brd.pieces[c][PAWN])

	for b = brd.pieces[c][KNIGHT]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += tropism_bonus[sq][enemy_king_sq][KNIGHT] + knight_pawns[pawn_count]
		mobility += knight_mobility[pop_count(knight_masks[sq]&available&unguarded)]
	}
	for b = brd.pieces[c][BISHOP]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += tropism_bonus[sq][enemy_king_sq][BISHOP]
		mobility += bishop_mobility[pop_count(bishop_attacks(occ, sq)&available&unguarded)]
	}
	for b = brd.pieces[c][ROOK]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += tropism_bonus[sq][enemy_king_sq][ROOK] + rook_pawns[pawn_count]
		mobility += rook_mobility[pop_count(rook_attacks(occ, sq)&available&unguarded)]
	}
	for b = brd.pieces[c][QUEEN]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += tropism_bonus[sq][enemy_king_sq][QUEEN]
		mobility += queen_mobility[pop_count(queen_attacks(occ, sq)&available&unguarded)]
	}
	for b = brd.pieces[c][KING]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += king_pst[c][in_endgame(brd, c)][sq]
	}
	// placement += pawn_structure(brd, c, e)

	return placement + mobility
}

// PAWN EVALUATION
//
// Good structures:
//   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file.
//                   May eventually get promoted.
//   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
//
// Bad structures:
//   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.
//   -Double/tripled pawns - Penalty for having multiple pawns on the same file.
func pawn_structure(brd *Board, c, e uint8) int {
	var structure, sq int
	own_pawns := brd.pieces[c][PAWN]
	enemy_pawns := brd.pieces[e][PAWN]

	for b := own_pawns; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		// passed pawns
		if !(pawn_passed_masks[c][sq]&enemy_pawns > 0) {
			structure += passed_pawn_bonus[c][row(sq)]
			if row(sq) == promote_row[c][0] {
				if !is_attacked_by(brd, get_offset(c, sq, 8), e, c) {
					structure += passed_pawn_bonus[c][row(sq)] // double the value of the bonus if path to promotion is undefended.
				}
			} else if row(sq) == promote_row[c][1] {
				if !is_attacked_by(brd, get_offset(c, sq, 8), e, c) &&
					!is_attacked_by(brd, get_offset(c, sq, 16), e, c) {
					structure += passed_pawn_bonus[c][row(sq)] // double the value of the bonus if path to promotion is undefended.
				}
			}
		}
		// isolated pawns
		if pawn_isolated_masks[sq]&own_pawns == 0 {
			structure += isolated_pawn_penalty
		}
		// pawn duos
		if pawn_side_masks[sq]&own_pawns > 0 {
			structure += pawn_duo_bonus
		}
	}
	var column_count int
	for i := 0; i < 8; i++ {
		// doubled/tripled pawns
		column_count = pop_count(column_masks[i] & own_pawns)
		if column_count > 1 {
			structure += double_pawn_penalty << (uint(column_count - 2))
		}
	}
	return structure
}

func get_offset(c uint8, sq, off int) int {
	if c > 0 {
		return sq + off
	} else {
		return sq - off
	}
}

func in_endgame(brd *Board, c uint8) int {
	if int(brd.material[c]) <= endgame_value {
		return 1
	} else {
		return 0
	}
}

func pawns_only(brd *Board, c uint8) bool {
	return brd.occupied[c] == brd.pieces[c][PAWN]|brd.pieces[c][KING]
}
