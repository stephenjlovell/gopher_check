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
// "fmt"
)

// This does not work for discovered checks.
func gives_check(brd *Board, m Move) bool {
	e := brd.Enemy()
	return true
	switch m.Piece() {
	case PAWN:
		return pawn_attack_masks[brd.c][m.To()] & brd.pieces[e][KING] > 0
	case KNIGHT:
		return knight_masks[m.To()] & brd.pieces[e][KING] > 0
	case BISHOP:
		return bishop_attacks(brd.Occupied(), m.To()) & brd.pieces[e][KING] > 0
	case ROOK:
		return rook_attacks(brd.Occupied(), m.To()) & brd.pieces[e][KING] > 0
	case QUEEN:
		return queen_attacks(brd.Occupied(), m.To()) & brd.pieces[e][KING] > 0
	default:
		return false
	}
	// return false
}

func attack_map(brd *Board, sq int) BB {
	var attacks, b_attackers, r_attackers BB
	occ := brd.Occupied()
	attacks |= (pawn_attack_masks[BLACK][sq] & brd.pieces[WHITE][PAWN]) | // Pawns
		(pawn_attack_masks[WHITE][sq] & brd.pieces[BLACK][PAWN])
	attacks |= (knight_masks[sq] & (brd.pieces[WHITE][KNIGHT] | brd.pieces[BLACK][KNIGHT])) // Knights
	b_attackers = brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] |                   // Bishops and Queens
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	attacks |= bishop_attacks(occ, sq) & b_attackers
	r_attackers = brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] | // Rooks and Queens
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	attacks |= rook_attacks(occ, sq) & r_attackers
	attacks |= king_masks[sq] & (brd.pieces[WHITE][KING] | brd.pieces[BLACK][KING]) // Kings
	return attacks
}

func color_attack_map(brd *Board, sq int, c, e uint8) BB {
	var attacks, b_attackers, r_attackers BB
	occ := brd.Occupied()
	attacks |= pawn_attack_masks[e][sq] & brd.pieces[c][PAWN]  // Pawns
	attacks |= knight_masks[sq] & brd.pieces[c][KNIGHT]        // Knights
	b_attackers = brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN] // Bishops and Queens
	attacks |= bishop_attacks(occ, sq) & b_attackers
	r_attackers = brd.pieces[c][ROOK] | brd.pieces[c][QUEEN] // Rooks and Queens
	attacks |= rook_attacks(occ, sq) & r_attackers
	attacks |= king_masks[sq] & brd.pieces[c][KING] // Kings
	return attacks
}

func is_attacked_by(brd *Board, sq int, attacker, defender uint8) bool {
	occ := brd.Occupied()
	if pawn_attack_masks[defender][sq]&brd.pieces[attacker][PAWN] > 0 { // Pawns
		return true
	}
	if knight_masks[sq]&(brd.pieces[attacker][KNIGHT]) > 0 { // Knights
		return true
	}
	if bishop_attacks(occ, sq)&(brd.pieces[attacker][BISHOP]|brd.pieces[attacker][QUEEN]) > 0 { // Bishops and Queens
		return true
	}
	if rook_attacks(occ, sq)&(brd.pieces[attacker][ROOK]|brd.pieces[attacker][QUEEN]) > 0 { // Rooks and Queens
		return true
	}
	if king_masks[sq]&(brd.pieces[attacker][KING]) > 0 { // Kings
		return true
	}
	return false
}

// Determines if a piece is blocking a ray attack to its king, and cannot move off this ray
// without placing its king in check.
// 1. Find the displacement vector between the piece at sq and its own king and determine if it
//    lies along a valid ray attack.  If the vector is a valid ray attack:
// 2. Scan toward the king to see if there are any other pieces blocking this route to the king.
// 3. Scan in the opposite direction to see detect any potential threats along this ray.
func is_pinned(brd *Board, sq int, c, e uint8) BB {
	occ := brd.Occupied()
	var threat, guarded_king BB
	dir := directions[sq][furthest_forward(c, brd.pieces[c][KING])] // get direction toward king
	switch dir {
	case NW, NE:
		threat = scan_down(occ, dir+2, sq) & (brd.pieces[e][BISHOP] | brd.pieces[e][QUEEN])
		guarded_king = scan_up(occ, dir, sq) & (brd.pieces[c][KING])
	case SE, SW:
		threat = scan_up(occ, dir-2, sq) & (brd.pieces[e][BISHOP] | brd.pieces[e][QUEEN])
		guarded_king = scan_down(occ, dir, sq) & (brd.pieces[c][KING])
	case NORTH, EAST:
		threat = scan_down(occ, dir+2, sq) & (brd.pieces[e][ROOK] | brd.pieces[e][QUEEN])
		guarded_king = scan_up(occ, dir, sq) & (brd.pieces[c][KING])
	case SOUTH, WEST:
		threat = scan_up(occ, dir-2, sq) & (brd.pieces[e][ROOK] | brd.pieces[e][QUEEN])
		guarded_king = scan_down(occ, dir, sq) & (brd.pieces[c][KING])
	case DIR_INVALID:
		return 0
	}
	return (threat & guarded_king)
}

// The Static Exchange Evaluation (SEE) heuristic provides a way to determine if a capture
// is a 'winning' or 'losing' capture.
// 1. When a capture results in an exchange of pieces by both sides, SEE is used to determine the
//    net gain/loss in material for the side initiating the exchange.
// 2. SEE scoring of moves is used for move ordering of captures at critical nodes.
// 3. During quiescence search, SEE is used to prune losing captures. This provides a very low-risk
//    way of reducing the size of the q-search without impacting playing strength.
const (
	SEE_MIN = -780 // worst possible outcome (trading a queen for a pawn)
	SEE_MAX = 880  // best outcome (capturing an undefended queen)
)

func get_see(brd *Board, from, to int, captured_piece Piece) int {
	var next_victim int
	var t Piece
	// var t, last_t Piece
	temp_color := brd.Enemy()
	// get initial map of all squares directly attacking this square (does not include 'discovered'/hidden attacks)
	b_attackers := brd.pieces[WHITE][BISHOP] | brd.pieces[BLACK][BISHOP] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]
	r_attackers := brd.pieces[WHITE][ROOK] | brd.pieces[BLACK][ROOK] |
		brd.pieces[WHITE][QUEEN] | brd.pieces[BLACK][QUEEN]

	temp_map := attack_map(brd, to)
	temp_occ := brd.Occupied()
	var temp_pieces BB

	var piece_list [20]int
	count := 1

	if captured_piece == KING {
		// this move is illegal and will be discarded by search.  return the lowest possible
		// SEE value so that this move will be put at end of list.  If cutoff occurs before then,
		// the cost of detecting the illegal move will be saved.
		return SEE_MIN
	}
	t = brd.TypeAt(from)
	if t == KING { // Only commit to the attack if target piece is undefended.
		if temp_map&brd.occupied[temp_color] > 0 {
			return SEE_MIN
		} else {
			return piece_values[captured_piece]
		}
	}
	// before entering the main loop, perform each step once for the initial attacking piece.
	// This ensures that the moved piece is the first to capture.
	piece_list[0] = piece_values[captured_piece]
	next_victim = brd.ValueAt(from)

	temp_occ.Clear(from)
	if t != KNIGHT && t != KING { // if the attacker was a pawn, bishop, rook, or queen, re-scan for hidden attacks:
		if t == PAWN || t == BISHOP || t == QUEEN {
			temp_map |= bishop_attacks(temp_occ, to) & b_attackers
		}
		if t == PAWN || t == ROOK || t == QUEEN {
			temp_map |= rook_attacks(temp_occ, to) & r_attackers
		}
	}

	for temp_map &= temp_occ; temp_map > 0; temp_map &= temp_occ {
		for t = PAWN; t <= KING; t++ { // loop over piece ts in order of value.
			temp_pieces = brd.pieces[temp_color][t] & temp_map
			if temp_pieces > 0 {
				break
			} // stop as soon as a match is found.
		}
		if t >= KING {
			if t == KING {
				if temp_map&brd.occupied[temp_color^1] > 0 {
					break // only commit a king to the attack if the other side has no defenders left.
				}
			}
			break
		}

		piece_list[count] = next_victim - piece_list[count-1]
		next_victim = piece_values[t]

		count++

		if (piece_list[count-1] - next_victim) > 0 { // validate this.
			break
		}

		temp_occ ^= (temp_pieces & -temp_pieces) // merge the first set bit of temp_pieces into temp_occ
		if t != KNIGHT && t != KING {
			if t == PAWN || t == BISHOP || t == QUEEN {
				temp_map |= (bishop_attacks(temp_occ, to) & b_attackers)
			}
			if t == ROOK || t == QUEEN {
				temp_map |= (rook_attacks(temp_occ, to) & r_attackers)
			}
		}
		temp_color ^= 1
	}

	for count-1 > 0 {
		count--
		piece_list[count-1] = -max(-piece_list[count-1], piece_list[count])
	}
	// fmt.Printf(" %d ", piece_list[0])
	return piece_list[0]
}

// make these methods of Board type.

func side_in_check(brd *Board, c, e uint8) bool { // determines if specified side is in check
	if brd.pieces[c][KING] == 0 {
		return true
	} else {
		return is_attacked_by(brd, furthest_forward(c, brd.pieces[c][KING]), e, c)
	}
}

func is_in_check(brd *Board) bool { // determines if side to move is in check
	return side_in_check(brd, brd.c, brd.Enemy())
}

func enemy_in_check(brd *Board) bool { // determines if other side is in check
	return side_in_check(brd, brd.Enemy(), brd.c)
}

func avoids_check(brd *Board, m Move, in_check bool) bool {
	return in_check || pseudolegal_avoids_check(brd, m)
}

func pseudolegal_avoids_check(brd *Board, m Move) bool {
	if m.Piece() == KING {
		return !is_attacked_by(brd, m.To(), brd.Enemy(), brd.c)
	} else {
		pinned := is_pinned(brd, m.From(), brd.c, brd.Enemy())
		return !((pinned > 0) && ((^pinned)&sq_mask_on[m.To()]) > 0)
	}
}

