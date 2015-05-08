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
// "fmt"
)

const (
	ENDGAME_COUNT    = 18
	DOUBLED_PENALTY	 = 20
	ISOLATED_PENALTY = 12
	BACKWARD_PENALTY = 4
)

var main_pst = [2][8][64]int{ // Black. White PST will be set in setup_eval.
	{ // Pawn
		{   0,  0, 0,   0,   0, 0,  0,   0,
			-11,  1, 1,   1,   1, 1,  1, -11,
			-12,  0, 1,   2,   2, 1,  0, -12,
			-13, -1, 2,  10,  10, 2, -1, -13,
			-14, -2, 4,  14,  14, 4, -2, -14,
			-15, -3, 0,   9,   9, 0, -3, -15,
			-16, -4, 0, -20, -20, 0, -4, -16,
			  0,  0, 0,   0,   0, 0,  0,   0},
		// Knight
		{  -8,  -8, -6, -6, -6, -6, -8,  -8,
			 -8,   0,  0,  0,  0,  0,  0,  -8,
			 -6,   0,  4,  4,  4,  4,  0,  -6,
			 -6,   0,  4,  8,  8,  4,  0,  -6,
			 -6,   0,  4,  8,  8,  4,  0,  -6,
			 -6,   0,  4,  4,  4,  4,  0,  -6,
			 -8,   0,  1,  2,  2,  1,  0,  -8,
			-10, -12, -6, -6, -6, -6, -12, -10},
		// Bishop
		{ -3, -3,  -3, -3, -3,  -3, -3, -3,
			-3,  0,   0,  0,  0,   0,  0, -3,
			-3,  0,   2,  4,  4,   2,  0, -3,
			-3,  0,   4,  5,  5,   4,  0, -3,
			-3,  0,   4,  5,  5,   4,  0, -3,
			-3,  1,   2,  4,  4,   2,  1, -3,
			-3,  2,   1,  1,  1,   1,  2, -3,
			-3, -3, -10, -3, -3, -10, -3, -3},
		// Rook
		{  4,  4,  4,  4,  4,  4,  4,  4,
			16, 16, 16, 16, 16, 16, 16, 16,
			-4,  0,  0,  0,  0,  0,  0, -4,
			-4,  0,  0,  0,  0,  0,  0, -4,
			-4,  0,  0,  0,  0,  0,  0, -4,
			-4,  0,  0,  0,  0,  0,  0, -4,
			-4,  0,  0,  0,  0,  0,  0, -4,
			 0,  0,  0,  2,  2,  0,  0,  0 },
		// Queen
		{  0,  0,  0,  1,  1,  0,  0,  0,
			 0,  0,  1,  2,  2,  1,  0,  0,
			 0,  1,  2,  2,  2,  2,  1,  0,
			 0,  1,  2,  3,  3,  2,  1,  0,
			 0,  1,  2,  3,  3,  2,  1,  0,
			 0,  1,  1,  2,  2,  1,  1,  0,
			 0,  0,  1,  1,  1,  1,  0,  0,
			-6, -6, -6, -6, -6, -6, -6, -6},
	},
}

var king_pst = [2][2][64]int{ // Black 
	{ // Early game
		{
			-52, -50, -50, -50, -50, -50, -50, -52, // In early game, encourage the king to stay on back
			-50, -48, -48, -48, -48, -48, -48, -50, // row defended by friendly pieces.
			-48, -46, -46, -46, -46, -46, -46, -48,
			-46, -44, -44, -44, -44, -44, -44, -46,
			-44, -42, -42, -42, -42, -42, -42, -44,
			-42, -40, -40, -40, -40, -40, -40, -42,
			-16, -15, -20, -20, -20, -20, -15, -16,
			  0,  20,  30, -30,  0,  -20,  30,  20,
		},
		{ // Endgame
			-30, -20, -10,  0,  0, -10, -20, -30, // In end game (when few friendly pieces are available
			-20, -10,   0, 10, 10,   0, -10, -20, // to protect king), the king should move toward the center
			-10,   0,  10, 20, 20,  10,   0, -10, // and avoid getting trapped in corners.
			  0,  10,  20, 30, 30,  20,  10,   0,
			  0,  10,  20, 30, 30,  20,  10,   0,
			-10,   0,  10, 20, 20,  10,   0, -10,
			-20, -10,   0, 10, 10,   0, -10, -20,
			-30, -20, -10,  0,  0, -10, -20, -30,
		},
	},
}

var square_mirror = [64]int{
	H1, H2, H3, H4, H5, H6, H7, H8,
	G1, G2, G3, G4, G5, G6, G7, G8,
	F1, F2, F3, F4, F5, F6, F7, F8,
	E1, E2, E3, E4, E5, E6, E7, E8,
	D1, D2, D3, D4, D5, D6, D7, D8,
	C1, C2, C3, C4, C5, C6, C7, C8,
	B1, B2, B3, B4, B5, B6, B7, B8,
	A1, A2, A3, A4, A5, A6, A7, A8,
}

var king_threat_bonus = [64]int{
	0, 		 2, 	3, 	 5, 	9, 	15,  24,  37,
	55, 	79, 111, 150, 195, 244, 293, 337,
	370, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
	389, 389, 389, 389, 389, 389, 389, 389,
}

var king_saftey_base = [2][2][64]int{
	{ // Black
		{ // Early-game
			4, 4, 4, 4, 4, 4, 4, 4,
			4, 4, 4, 4, 4, 4, 4, 4,
			4, 4, 4, 4, 4, 4, 4, 4,
			4, 4, 4, 4, 4, 4, 4, 4,
			4, 4, 4, 4, 4, 4, 4, 4,
			4, 3, 3, 3, 3, 3, 3, 4,
			3, 1, 1, 1, 1, 1, 1, 3,
			2, 0, 0, 0, 0, 0, 0, 2,
		},
	},
}

// adjusts value of knights and rooks based on number of own pawns in play.
var knight_pawns = [16]int{-20, -16, -12, -8, -4, 0, 4, 8, 12}
var rook_pawns = [16]int{16, 12, 8, 4, 2, 0, -2, -4, -8}

// adjusts the value of bishop pairs based on number of enemy pawns in play.
var bishop_pair_pawns = [16]int{10, 10, 9, 8, 6, 4, 2, 0, -2}

var knight_mobility = [16]int{-16, -12, -6, -3, 0, 1, 3, 5, 6, 0, 0, 0, 0, 0, 0}

var bishop_mobility = [16]int{-24, -16, -8, -4, -2, 0, 2, 4, 6, 7, 8, 9, 10, 11, 12, 13}

var rook_mobility = [16]int{-12, -8, -4, -2, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

var queen_mobility = [32]int{-24, -18, -12, -6, -3, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 24, 24, 24}

var queen_tropism_bonus = [8]int{0, 12, 9, 6, 3, 0, -3, -6}

func is_passed_pawn(brd *Board, m Move) bool {
	if m.Piece() != PAWN {
		return false
	} else {
		return pawn_passed_masks[brd.c][m.From()]&brd.pieces[brd.Enemy()][PAWN] == 0
	}
}

func evaluate(brd *Board, alpha, beta int) int {
	score := int(brd.material[WHITE]-brd.material[BLACK]) + tempo_bonus()
	// lazy evaluation: if material balance is already outside the search window by an amount that outweighs
	// the largest likely placement evaluation, return the material as an approximate evaluation.
	// This prevents the engine from wasting a lot of time evaluating unrealistic positions.
	if score+piece_values[BISHOP] < alpha || score-piece_values[BISHOP] > beta {
		if brd.c == BLACK {
			return -score
		}
		return score
	}

	// pentry := &PawnEntry{}
	pentry := brd.worker.ptt.Probe(brd.pawn_hash_key)
	if pentry.key != brd.pawn_hash_key {
		set_pawn_structure(brd, pentry)
	}

	score += pentry.value 
	score += net_passed_pawns(brd, pentry)

	score += net_major_placement(brd, pentry)

	if brd.c == BLACK { // score is calculated relative to white to move
		return -score
	}
	return score
}

func lazy_eval(brd *Board) int {
	return int(brd.material[brd.c] - brd.material[brd.Enemy()])
}

func tempo_bonus() int {
	if side_to_move == WHITE {
		return 5
	} else {
		return -5
	}
}

func net_major_placement(brd *Board, pentry *PawnEntry) int {
	return major_placement(brd, pentry, WHITE, BLACK) - major_placement(brd, pentry, BLACK, WHITE)
}

func major_placement(brd *Board, pentry *PawnEntry, c, e uint8) int {
	friendly := brd.Placement(c)
	occ := brd.AllOccupied()

	unguarded := ^(pentry.all_attacks[e])
	available := (^friendly) & unguarded

	var sq, mobility, placement, king_threats int
	var b, attacks BB
	enemy_king_sq := furthest_forward(e, brd.pieces[e][KING])
	enemy_king_zone := king_zone_masks[e][enemy_king_sq]
	endgame := brd.InEndgame()

	pawn_count := pentry.count[c]

	for b = brd.pieces[c][KNIGHT]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += knight_pawns[pawn_count]
		attacks = knight_masks[sq] & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += knight_mobility[pop_count(attacks)]
	}

	for b = brd.pieces[c][BISHOP]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = bishop_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += bishop_mobility[pop_count(attacks)]
	}
	// bishop pairs
	if pop_count(brd.pieces[c][BISHOP]) > 1 {
		placement += 40 + bishop_pair_pawns[pentry.count[e]]
	}

	for b = brd.pieces[c][ROOK]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		placement += rook_pawns[pawn_count]
		attacks = rook_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += rook_mobility[pop_count(attacks)]
	}
	for b = brd.pieces[c][QUEEN]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = queen_attacks(occ, sq) & available
		king_threats += pop_count(attacks & enemy_king_zone)
		mobility += queen_mobility[pop_count(attacks)]
		// queen tropism: encourage queen to move toward enemy king.
		placement += queen_tropism_bonus[chebyshev_distance(sq, enemy_king_sq)]
	}

	for b = brd.pieces[c][KING]; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		attacks = king_masks[sq] & available
		if endgame == 0 {
			placement += pawn_shield_bonus[pop_count(brd.pieces[c][PAWN]&king_shield_masks[c][sq])]
		}
	}

	placement += king_threat_bonus[king_threats+king_saftey_base[e][endgame][enemy_king_sq]]

	return placement + mobility
}

// PAWN EVALUATION
// Good structures:
//   -Passed pawns - Bonus for pawns unblocked by an enemy pawn on the same or adjacent file.
//                   May eventually get promoted.
//   -Pawn duos - Pawns side by side to another friendly pawn receive a small bonus
// Bad structures:
//   -Isolated pawns - Penalty for any pawn without friendly pawns on adjacent files.
//   -Double/tripled pawns - Penalty for having multiple pawns on the same file.

var passed_pawn_bonus = [2][8]int{
	{0, 192, 96, 48, 24, 12,   6, 0},
	{0,   6, 12, 24, 48, 96, 192, 0},
}
var tarrasch_bonus = [2][8]int{
	{0,  12,  8,  4,  2,  0,   0,  0},
	{0,   0,  0,  2,  4,  8,  12,  0},
}
var defense_bonus = [2][8]int{
	{0,  12,  8,  6,  5,  4,   3,  0},
	{0,   3,  4,  5,  6,  8,  12,  0},
}
var duo_bonus = [2][8]int{
	{0,   0,  2,  1,  1,  1,   0,  0},
	{0,   0,  1,  1,  1,  2,   0,  0},
}

var pawn_shield_bonus = [4]int{-9, -3, 3, 9}

var promote_row = [2][2]int{
	{1, 2},
	{6, 5},
}


func set_pawn_structure(brd *Board, pentry *PawnEntry) {
	pentry.left_attacks[WHITE], pentry.right_attacks[WHITE] = pawn_attacks(brd, WHITE)
	pentry.left_attacks[BLACK], pentry.right_attacks[BLACK] = pawn_attacks(brd, BLACK)

	pentry.all_attacks[WHITE] = pentry.left_attacks[WHITE]|pentry.right_attacks[WHITE]
	pentry.all_attacks[BLACK] = pentry.left_attacks[BLACK]|pentry.right_attacks[BLACK]

	pentry.count[WHITE] = uint8(pop_count(brd.pieces[WHITE][PAWN]))
	pentry.count[BLACK] = uint8(pop_count(brd.pieces[BLACK][PAWN]))

	pentry.passed_pawns[WHITE], pentry.passed_pawns[BLACK] = 0, 0

	pentry.key = brd.pawn_hash_key
	pentry.value = (pawn_structure(brd, pentry, WHITE, BLACK)-pawn_structure(brd, pentry, BLACK, WHITE))
}

// Evaluation features that depend ONLY on the position of pawns go here.
// Only pawn position is used for the pawn hash key.
func pawn_structure(brd *Board, pentry *PawnEntry, c, e uint8) int {
	var value, sq, sq_row int
	own_pawns, enemy_pawns := brd.pieces[c][PAWN], brd.pieces[e][PAWN]

	for b := own_pawns; b > 0; b.Clear(sq) {
		sq = furthest_forward(c, b)
		sq_row = row(sq)
		
		if (pawn_attack_masks[c][sq]|pawn_attack_masks[e][sq])&own_pawns > 0 { // defended pawns
			value += defense_bonus[c][sq_row]
		} else if (pawn_side_masks[sq]&own_pawns) > 0 { // pawn duos
			value += duo_bonus[c][sq_row]
		}

		if pawn_doubled_masks[sq]&own_pawns > 0 { // doubled or tripled pawns
			value -= DOUBLED_PENALTY
		}

		if pawn_passed_masks[c][sq]&enemy_pawns == 0 {	// passed pawns
			value += passed_pawn_bonus[c][sq_row]
			pentry.passed_pawns[c].Add(sq)		// note the passed pawn location in the pawn hash entry.
		} else {  // don't penalize passed pawns for being isolated.
			if pawn_isolated_masks[sq]&own_pawns == 0 { 
				value -= ISOLATED_PENALTY  // isolated pawns
			}
		}

		if pawn_attack_spans[e][pawn_stop_sq[c][sq]]&own_pawns == 0 && // backward pawns
			pawn_stop_masks[c][sq] & (enemy_pawns|pentry.all_attacks[e]) > 0 {
			value -= BACKWARD_PENALTY
		}
	}
	
	return value
}



func net_passed_pawns(brd *Board, pentry *PawnEntry) int {
	return eval_passed_pawns(brd, WHITE, BLACK, pentry.passed_pawns[WHITE]) -
				 eval_passed_pawns(brd, BLACK, WHITE, pentry.passed_pawns[BLACK])		
}

func eval_passed_pawns(brd *Board, c, e uint8, passed_pawns BB) int {
	var value, sq int
	enemy_king_sq := brd.KingSq(e)

	for ; passed_pawns > 0; passed_pawns.Clear(sq) {
		sq = furthest_forward(c, passed_pawns)
		// Tarrasch rule: assign small bonus for friendly rook behind the passed pawn
		if pawn_front_spans[e][sq] & brd.pieces[c][ROOK] > 0 {
			value += tarrasch_bonus[c][row(sq)]
		}
		// pawn race: Assign a bonus if the pawn is closer to its promote square than the enemy king.
		promote_square := pawn_promote_sq[c][sq]
		if side_to_move == c {
			if chebyshev_distance(sq, promote_square) < (chebyshev_distance(enemy_king_sq, promote_square) ) {
				value += passed_pawn_bonus[c][row(sq)]				
			}			
		} else {
			if chebyshev_distance(sq, promote_square) < (chebyshev_distance(enemy_king_sq, promote_square)-1) {
				value += passed_pawn_bonus[c][row(sq)]				
			}		
		}
	}
	return value
}


func setup_eval() {
	// Main PST
	for piece := PAWN; piece < KING; piece++ {
		for sq := 0; sq < 64; sq++ {
			main_pst[WHITE][piece][sq] = main_pst[BLACK][piece][square_mirror[sq]]
		}
	}
	// King PST
	for endgame := 0; endgame < 2; endgame++ {
		for sq := 0; sq < 64; sq++ {
			king_pst[WHITE][endgame][sq] = king_pst[BLACK][endgame][square_mirror[sq]]
		}
	}
	// King saftey counters
	for sq := 0; sq < 64; sq++ {
		king_saftey_base[WHITE][0][sq] = king_saftey_base[BLACK][0][square_mirror[sq]]
	}
}
