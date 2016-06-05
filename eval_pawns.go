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

// "fmt"

const (
	DOUBLED_PENALTY  = 20
	ISOLATED_PENALTY = 12
	BACKWARD_PENALTY = 4
)

var passed_pawn_bonus = [2][8]int{
	{0, 192, 96, 48, 24, 12, 6, 0},
	{0, 6, 12, 24, 48, 96, 192, 0},
}
var tarrasch_bonus = [2][8]int{
	{0, 12, 8, 4, 2, 0, 0, 0},
	{0, 0, 0, 2, 4, 8, 12, 0},
}
var defense_bonus = [2][8]int{
	{0, 12, 8, 6, 5, 4, 3, 0},
	{0, 3, 4, 5, 6, 8, 12, 0},
}
var duo_bonus = [2][8]int{
	{0, 0, 2, 1, 1, 1, 0, 0},
	{0, 0, 1, 1, 1, 2, 0, 0},
}

var promote_row = [2][2]int{
	{1, 2},
	{6, 5},
}

// PAWN EVALUATION
// Good structures:
//   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file.
//                   May eventually get promoted.
//   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
// Bad structures:
//   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.
//   -Double/tripled pawns - Penalty for having multiple pawns on the same file.
//   -Backward pawns

func set_pawn_structure(brd *Board, pentry *PawnEntry) {
	pentry.key = brd.pawn_hash_key
	set_pawn_maps(brd, pentry, WHITE)
	set_pawn_maps(brd, pentry, BLACK)
	pentry.value[WHITE] = pawn_structure(brd, pentry, WHITE, BLACK) -
		pawn_structure(brd, pentry, BLACK, WHITE)
	pentry.value[BLACK] = -pentry.value[WHITE]
}

func set_pawn_maps(brd *Board, pentry *PawnEntry, c uint8) {
	pentry.left_attacks[c], pentry.right_attacks[c] = pawn_attacks(brd, c)
	pentry.all_attacks[c] = pentry.left_attacks[c] | pentry.right_attacks[c]
	pentry.count[c] = uint8(pop_count(brd.pieces[c][PAWN]))
	pentry.passed_pawns[c] = 0
}

// pawn_structure() sets the remaining pentry attributes for side c
func pawn_structure(brd *Board, pentry *PawnEntry, c, e uint8) int {

	var value, sq, sq_row int
	own_pawns, enemy_pawns := brd.pieces[c][PAWN], brd.pieces[e][PAWN]
	for b := own_pawns; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		sq_row = row(sq)

		if (pawn_attack_masks[e][sq])&own_pawns > 0 { // defended pawns
			value += defense_bonus[c][sq_row]
		}
		if (pawn_side_masks[sq] & own_pawns) > 0 { // pawn duos
			value += duo_bonus[c][sq_row]
		}

		if pawn_doubled_masks[sq]&own_pawns > 0 { // doubled or tripled pawns
			value -= DOUBLED_PENALTY
		}

		if pawn_passed_masks[c][sq]&enemy_pawns == 0 { // passed pawns
			value += passed_pawn_bonus[c][sq_row]
			pentry.passed_pawns[c].Add(sq) // note the passed pawn location in the pawn hash entry.
		} else { // don't penalize passed pawns for being isolated.
			if pawn_isolated_masks[sq]&own_pawns == 0 {
				value -= ISOLATED_PENALTY // isolated pawns
			}
		}

		// https://chessprogramming.wikispaces.com/Backward+Pawn
		// backward pawns:
		// 1. cannot be defended by friendly pawns,
		// 2. their stop square is defended by an enemy sentry pawn,
		// 3. their stop square is not defended by a friendly pawn
		if (pawn_backward_spans[c][sq]&own_pawns == 0) &&
			(pentry.all_attacks[e]&pawn_stop_masks[c][sq] > 0) {
			value -= BACKWARD_PENALTY
		}
	}
	return value
}

func net_pawn_placement(brd *Board, pentry *PawnEntry, c, e uint8) int {
	return pentry.value[c] + net_passed_pawns(brd, pentry, c, e)
}

func net_passed_pawns(brd *Board, pentry *PawnEntry, c, e uint8) int {
	return eval_passed_pawns(brd, c, e, pentry.passed_pawns[c]) -
		eval_passed_pawns(brd, e, c, pentry.passed_pawns[e])
}

func eval_passed_pawns(brd *Board, c, e uint8, passed_pawns BB) int {
	var value, sq int
	enemy_king_sq := brd.KingSq(e)
	for ; passed_pawns > 0; passed_pawns.Clear(sq) {
		sq = furthest_forward(c, passed_pawns)
		// Tarrasch rule: assign small bonus for friendly rook behind the passed pawn
		if pawn_front_spans[e][sq]&brd.pieces[c][ROOK] > 0 {
			value += tarrasch_bonus[c][row(sq)]
		}
		// pawn race: Assign a bonus if the pawn is closer to its promote square than the enemy king.
		promote_square := pawn_promote_sq[c][sq]
		if brd.c == c {
			if chebyshev_distance(sq, promote_square) < (chebyshev_distance(enemy_king_sq, promote_square)) {
				value += passed_pawn_bonus[c][row(sq)]
			}
		} else {
			if chebyshev_distance(sq, promote_square) < (chebyshev_distance(enemy_king_sq, promote_square) - 1) {
				value += passed_pawn_bonus[c][row(sq)]
			}
		}
	}
	return value
}
