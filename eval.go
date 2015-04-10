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

var passed_pawn_bonus = [2][8]int{
	{0, 128, 64, 32, 16, 8, 4, 0},
	{0, 4, 8, 16, 32, 64, 128, 0},
}

var promote_row = [2][2]int{
	{1, 2},
	{6, 5},
}

const (
	ENDGAME_COUNT    = 18
	DUO_BONUS        = 2
	ISOLATED_PENALTY = -4
)

var double_pawn_penalty = [8]int{0, 0, -15, -30, -45, -60, -75, -90}

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

// {   0,  2,  3,  6, 12, 18, 25, 37, 50, 75,
//    100,125,150,175,200,225,250,275,300,325,
//    350,375,400,425,450,475,500,525,550,575
//    600,600,600,600,600 }

var king_threat_bonus = [64]int{
	0, 2, 3, 5, 9, 15, 24, 37,
	55, 79, 111, 150, 195, 244, 293, 337,
	370, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
}

var pawn_shield_bonus = [4]int{-9, -3, 3, 9}

// adjusts value of knights and rooks based on number of pawns in play.
var knight_pawns = [16]int{-20, -16, -12, -8, -4, 0, 4, 8, 12}
var rook_pawns = [16]int{16, 12, 8, 4, 2, 0, -2, -4, -8}

// adjusts the value of bishop pairs based on number of enemy pawns in play.
// var bishop_pair_pawns = [16]int{ 50, 50, 50, 50, 37, 25, 12, 6, 3 }
// var bishop_pair_pawns = [16]int{ 24, 24, 24, 24, 18, 12, 6, 3, 0 }
var bishop_pair_pawns = [16]int{10, 10, 8, 8, 6, 4, 2, 1, 0}

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
	setup_borders()
}

var highest_placement, lowest_placement int

func is_passed_pawn(brd *Board, m Move) bool {
	if m.Piece() != PAWN {
		return false
	} else {
		c, e := brd.c, brd.Enemy()
		return pawn_passed_masks[c][m.From()]&brd.pieces[e][PAWN] == 0
	}
}

var placement_test int

func evaluate(brd *Board, alpha, beta int) int {
	c, e := brd.c, brd.Enemy()
	material := int(brd.material[c] - brd.material[e])
	if brd.pieces[c][KING] == 0 {
		fmt.Println("The king is dead. Long live the king.")
		return -MATE
	}
	// lazy evaluation: if material balance is already outside the search window by an amount that outweighs
	// the largest likely placement evaluation, return the material as an approximate evaluation.
	// This prevents the engine from wasting a lot of time evaluating unrealistic positions.

	if material+piece_values[BISHOP] < alpha || material-piece_values[BISHOP] > beta {
		return material + tempo_bonus(c)
	}
	placement := adjusted_placement(brd, c, e) - adjusted_placement(brd, e, c) + tempo_bonus(c)
	// if placement > placement_test {
	// 	placement_test = placement
	// 	fmt.Printf("%d ", placement_test)
	// }
	return material + placement
}

func lazy_eval(brd *Board) int {
	return int(brd.material[brd.c]-brd.material[brd.Enemy()]) + tempo_bonus(brd.c)
}

func tempo_bonus(c uint8) int {
	if c == side_to_move {
		return 5
	} else {
		return -5
	}
}

// current ranges (approximate):
// PST bonus: {-152, 82}
// Mobility: { -39, 55}
// Pawn structure: {  }

// overall +- 550

var queen_tropism_bonus = [8]int{0, 12, 9, 6, 3, 0, -3, -6}

func adjusted_placement(brd *Board, c, e uint8) int {

	friendly := brd.Placement(c)
	available := ^friendly
	occ := brd.AllOccupied()

	var unguarded BB // a bitmap of squares undefended by enemy pawns
	if c > 0 {       // white to move
		unguarded = ^(((brd.pieces[e][PAWN] & (^column_masks[0])) >> 9) | ((brd.pieces[e][PAWN] & (^column_masks[7])) >> 7))
	} else { // black to move
		unguarded = ^(((brd.pieces[e][PAWN] & (^column_masks[0])) << 7) | ((brd.pieces[e][PAWN] & (^column_masks[7])) << 9))
	}
	var sq, mobility, placement, king_threats int
	var b, attacks BB
	enemy_king_sq := furthest_forward(e, brd.pieces[e][KING])

	enemy_king_zone := king_zone_masks[e][enemy_king_sq]

	pawn_count := pop_count(brd.pieces[c][PAWN])

	for b = brd.pieces[c][KNIGHT]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += knight_pawns[pawn_count]
		attacks = knight_masks[sq] & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += knight_mobility[pop_count(attacks&unguarded)]
	}

	b = brd.pieces[c][BISHOP]
	// if pop_count(b) > 1 {
	// 	// A pair of bishops is more useful the fewer enemy pawns are in play.
	// 	placement += bishop_pair_pawns[pop_count(brd.pieces[e][PAWN])]
	// }
	for ; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = bishop_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += bishop_mobility[pop_count(attacks&unguarded)]
	}
	for b = brd.pieces[c][ROOK]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += rook_pawns[pawn_count]
		attacks = rook_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += rook_mobility[pop_count(attacks&unguarded)]
	}
	for b = brd.pieces[c][QUEEN]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = queen_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += queen_mobility[pop_count(attacks&unguarded)]

		// add minor queen tropism bonus
		placement += queen_tropism_bonus[chebyshev_distance(sq, enemy_king_sq)]

	}

	endgame := in_endgame(brd)
	for b = brd.pieces[c][KING]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = king_masks[sq] & available
		king_threats += pop_count(attacks & enemy_king_zone)

		if endgame == 0 {
			placement += pawn_shield_bonus[pop_count(brd.pieces[c][PAWN]&king_shield_masks[c][sq])]
		}
		// placement += king_pst[c][endgame][sq]
	}

	// Squares along the edges of the board receive higher king threat bonuses, since there are fewer
	// places where the king can escape.
	border_bonus := borders[enemy_king_sq]
	placement += king_threat_bonus[king_threats+border_bonus]

	placement += pawn_structure(brd, c, e, endgame, enemy_king_sq)

	return placement + mobility
}

var borders [64]int

func setup_borders() {
	for i := 0; i < 64; i++ {
		borders[i] = 8 - pop_count(king_masks[i]) - 1
		if borders[i] < 0 {
			borders[i] = 0
		}
	}
}

var pawn_tropism_factor = [8]int{0, 3, 2, 2, 1, 0, 0, 0}

// PAWN EVALUATION
// Good structures:
//   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file.
//                   May eventually get promoted.
//   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
// Bad structures:
//   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.
//   -Double/tripled pawns - Penalty for having multiple pawns on the same file.
func pawn_structure(brd *Board, c, e uint8, endgame, enemy_king_sq int) int {
	var structure, sq int
	own_pawns := brd.pieces[c][PAWN]
	enemy_pawns := brd.pieces[e][PAWN]

	for b := own_pawns; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)

		// passed pawns
		if pawn_passed_masks[c][sq]&enemy_pawns == 0 {

			base_bonus := passed_pawn_bonus[c][row(sq)]

			if endgame == 1 {
				base_bonus += base_bonus >> 1 // increase the base bonus by half during endgame.
			}
			structure += base_bonus

			minor_bonus := base_bonus >> 2
			// passed pawns that are blocked by a friendly rook can't promote and can immobilize the
			// friendly rook if the pawn is attacked.
			if pawn_blocked_masks[c][sq]&brd.pieces[c][ROOK] > 0 {
				if is_attacked_by(brd, sq, e, c) {
					structure -= base_bonus >> 1
				} else {
					structure -= minor_bonus
				}
			}

			if (pawn_blocked_masks[c][sq] & brd.occupied[e]) > 0 {
				structure -= base_bonus >> 1 // pawn is directly blocked by an enemy piece.
			}

			structure -= pawn_tropism_factor[chebyshev_distance(sq, enemy_king_sq)] * minor_bonus

			if pawn_isolated_masks[sq]&own_pawns > 0 {
				structure += minor_bonus
			}

			next := get_offset(c, sq, 8)
			if brd.squares[next] == EMPTY {
				if !is_attacked_by(brd, next, e, c) { // next square undefended.
					structure += minor_bonus
					next = get_offset(c, next, 8)
					if next >= 0 && next < 64 && brd.squares[next] == EMPTY {
						if !is_attacked_by(brd, next, e, c) { // next square undefended.
							structure += minor_bonus
							next = get_offset(c, next, 8)
							if next >= 0 && next < 64 && brd.squares[next] == EMPTY {
								if !is_attacked_by(brd, next, e, c) { // next square undefended.
									structure += minor_bonus
								}
							}
						}
					}
				}
			}
		}

		// isolated pawns
		if pawn_isolated_masks[sq]&own_pawns == 0 {
			structure += ISOLATED_PENALTY
		}

		// pawn duos
		if pawn_side_masks[sq]&own_pawns > 0 {
			structure += DUO_BONUS
		}

	}

	for i := 0; i < 8; i++ {
		structure += double_pawn_penalty[pop_count(column_masks[i]&own_pawns)]
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

func in_endgame(brd *Board) int {
	if brd.endgame_counter < ENDGAME_COUNT {
		return 1
	} else {
		return 0
	}
}



