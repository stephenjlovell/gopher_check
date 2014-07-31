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

func split_moves(brd *Board, in_check bool) ([]Move, []Move) { // determine which moves should be searchd sequentially,
	var best_moves, other_moves []Move // and which will be searched in parallel.
	// moves := generate_moves(brd, in_check)

	// apply some heuristic to

	return best_moves, other_moves
}

func generate_moves(brd *Board, in_check bool) ([]Move, int) { // generate and sort all pseudolegal moves
	moves := make([]Move, 0)
	confidence := 0
	if in_check {

	} else {

	}

	return moves, confidence
}

func generate_tactical_moves(brd *Board, in_check bool) []Move { // generate and sort non-quiet pseudolegal moves
	moves := make([]Move, 0)

	if in_check {

	} else {

	}

	return moves
}

func get_non_captures(brd *Board) {
	var from, to int
	var single_advances, double_advances BB
	c, e := brd.c, brd.Enemy()
	occ := brd.Occupied()
	empty := ^occ

	// Castles
	castle := brd.castle
	if castle > uint8(0) { // get_non_captures is only called when not in check.
		if c > 0 {
			if (castle&C_WQ > uint8(0)) && !(castle_queenside_intervening[1]&occ > 0) &&
				!is_attacked_by(brd, D1, e, c) && !is_attacked_by(brd, C1, e, c) {

				// build_castle(INT2NUM(0x1b), E1, C1, INT2NUM(0x17), A1, D1, moves);
			}
			if (castle&C_WK > uint8(0)) && !(castle_kingside_intervening[1]&occ > 0) &&
				!is_attacked_by(brd, F1, e, c) && !is_attacked_by(brd, G1, e, c) {

				// build_castle(INT2NUM(0x1b), E1, G1, INT2NUM(0x17), H1, F1, moves);
			}
		} else {
			if (castle&C_BQ > uint8(0)) && !(castle_queenside_intervening[0]&occ > 0) &&
				!is_attacked_by(brd, D8, e, c) && !is_attacked_by(brd, C8, e, c) {

				// build_castle(INT2NUM(0x1a), E8, C8, INT2NUM(0x16), A8, D8, moves);
			}
			if (castle&C_BK > uint8(0)) && !(castle_kingside_intervening[0]&occ > 0) &&
				!is_attacked_by(brd, F8, e, c) && !is_attacked_by(brd, G8, e, c) {

				// build_castle(INT2NUM(0x1a), E8, G8, INT2NUM(0x16), H8, F8, moves);
			}
		}
	}

	// Pawns
	//  Pawns behave differently than other pieces. They:
	//  1. can move only in one direction;
	//  2. can attack diagonally but can only advance on file (forward);
	//  3. can move an extra space from the starting square;
	//  4. can capture other pawns via the En-Passant Rule;
	//  5. are promoted to another piece type if they reach the enemy's back rank.
	if c > 0 { // white to move
		single_advances = (brd.pieces[WHITE][PAWN] << 8) & empty & (^row_masks[7]) // promotions generated in get_captures
		double_advances = ((single_advances & row_masks[2]) << 8) & empty
	} else { // black to move
		single_advances = (brd.pieces[BLACK][PAWN] >> 8) & empty & (^row_masks[0])
		double_advances = ((single_advances & row_masks[5]) >> 8) & empty
	}

	for ; double_advances > 0; double_advances.Clear(to) {
		to = furthest_forward(c, double_advances)

		// build_move(piece_id, to+pawn_from_offsets[c][1], to, cls_enp_advance, moves);
	}
	for ; single_advances > 0; single_advances.Clear(to) {
		to = furthest_forward(c, single_advances)

		// build_move(piece_id, to+pawn_from_offsets[c][0], to, cls_pawn_move, moves);
	}

	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)                               // Locate each knight for the side to move.
		for t := (knight_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_move(piece_id, from, to, cls_regular_move, moves)
		}
	}

	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_move(piece_id, from, to, cls_regular_move, moves)
		}
	}

	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_move(piece_id, from, to, cls_regular_move, moves);
		}
	}

	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_move(piece_id, from, to, cls_regular_move, moves);
		}
	}

	// Kings
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = furthest_forward(c, brd.pieces[c][KING])
		for t := (king_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_move(piece_id, from, to, cls_regular_move, moves);
		}
	}

}

// Pawn promotions are also generated during get_captures routine.

func get_captures(brd *Board) {
	var from, to int
	c, e := brd.c, brd.Enemy()
	occ := brd.Occupied()
	enemy := brd.Placement(e)

	// Pawns
	var left_temp, right_temp, left_attacks, right_attacks BB
	var promotion_captures_left, promotion_captures_right, promotion_advances BB

	if c > 0 { // white to move
		left_temp = ((brd.pieces[c][PAWN] & (^column_masks[0])) << 7) & enemy
		left_attacks = left_temp & (^row_masks[7])
		promotion_captures_left = left_temp & (row_masks[7])

		right_temp = ((brd.pieces[c][PAWN] & (^column_masks[7])) << 9) & enemy
		right_attacks = right_temp & (^row_masks[7])
		promotion_captures_right = right_temp & (row_masks[7])

		promotion_advances = ((brd.pieces[c][PAWN] << 8) & row_masks[7]) & (^occ)
	} else { // black to move
		left_temp = ((brd.pieces[c][PAWN] & (^column_masks[0])) >> 9) & enemy
		left_attacks = left_temp & (^row_masks[0])
		promotion_captures_left = left_temp & (row_masks[0])

		right_temp = ((brd.pieces[c][PAWN] & (^column_masks[7])) >> 7) & enemy
		right_attacks = right_temp & (^row_masks[0])
		promotion_captures_right = right_temp & (row_masks[0])

		promotion_advances = ((brd.pieces[c][PAWN] >> 8) & row_masks[0]) & (^occ)
	}

	// promotion captures
	for ; promotion_captures_left > 0; promotion_captures_left.Clear(to) {
		to = furthest_forward(c, promotion_captures_left)

		// build_promotion_captures(piece_id, to+pawn_from_offsets[c][2], to, color, cls_promotion_capture, sq_board, promotions);
	}

	for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
		to = furthest_forward(c, promotion_captures_right)

		// build_promotion_captures(piece_id, to+pawn_from_offsets[c][3], to, color, cls_promotion_capture, sq_board, promotions);
	}

	// promotion advances
	for ; promotion_advances > 0; promotion_advances.Clear(to) {
		to = furthest_forward(c, promotion_advances)

		// build_promotions(piece_id, to+pawn_from_offsets[c][0], to, color, cls_promotion, promotions);
	}

	// regular pawn attacks
	for ; left_attacks > 0; left_attacks.Clear(to) {
		to = furthest_forward(c, left_attacks)

		// build_capture(piece_id, to+pawn_from_offsets[c][2], to, cls_regular_capture, sq_board, moves)
	}

	for ; right_attacks > 0; right_attacks.Clear(to) {
		to = furthest_forward(c, right_attacks)

		// build_capture(piece_id, to+pawn_from_offsets[c][3], to, cls_regular_capture, sq_board, moves);
	}

	// en-passant captures
	if brd.enp_target != SQ_INVALID {
		enp_target := brd.enp_target
		for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
			from = furthest_forward(c, f)

			// build_enp_capture(piece_id, from, (c?(enp_target+8):(enp_target-8)), cls_enp_capture, enp_target, sq_board, moves);
		}
	}

	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (knight_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_capture(piece_id, from, to, cls_regular_capture, sq_board, moves);
		}
	}

	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_capture(piece_id, from, to, cls_regular_capture, sq_board, moves);
		}
	}

	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_capture(piece_id, from, to, cls_regular_capture, sq_board, moves)
		}
	}

	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_capture(piece_id, from, to, cls_regular_capture, sq_board, moves);
		}
	}

	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = furthest_forward(c, brd.pieces[c][KING])
		for t := (king_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)

			// build_capture(piece_id, from, to, cls_regular_capture, sq_board, moves);
		}
	}

}

func get_evasions(brd *Board) {
	c, e := brd.c, brd.Enemy()

	if brd.pieces[c][KING] == 0 {
		return
	}

	var defense_map BB
	var from, to, threat_sq_1, threat_sq_2 int

	threat_dir_1, threat_dir_2 := DIR_INVALID, DIR_INVALID
	occ := brd.Occupied()
	empty := ^occ
	enemy := brd.Placement(e)

	king_sq := furthest_forward(c, brd.pieces[c][KING])
	threats := color_attack_map(brd, king_sq, e, c) // find any enemy pieces that attack the king.
	threat_count := pop_count(threats)

	// Get direction of the attacker(s) and any intervening squares between the attacker and the king.
	if threat_count == 1 {
		threat_sq_1 = lsb(threats)
		if brd.TypeAt(threat_sq_1) != PAWN {
			threat_dir_1 = directions[threat_sq_1][king_sq]
		}
		defense_map |= (intervening[threat_sq_1][king_sq] | threats)
		// // allow capturing of enemy king to detect illegal checking move by king capture.
		defense_map |= brd.pieces[e][KING]
	} else {
		threat_sq_1 = lsb(threats)
		if brd.TypeAt(threat_sq_1) != PAWN {
			threat_dir_1 = directions[threat_sq_1][king_sq]
		}
		threat_sq_2 = msb(threats)
		if brd.TypeAt(threat_sq_2) != PAWN {
			threat_dir_2 = directions[threat_sq_2][king_sq]
		}
		// // allow capturing of enemy king to detect illegal checking move by king capture.
		defense_map |= brd.pieces[e][KING]
	}

	if threat_count == 1 { // Attempt to capture or block the attack with any piece if there's only one attacker.
		// Pawns
		var single_advances, double_advances, left_temp, right_temp, left_attacks, right_attacks BB
		var promotion_captures_left, promotion_captures_right, promotion_advances BB

		if c > 0 { // white to move
			single_advances = (brd.pieces[WHITE][PAWN] << 8) & empty & (^row_masks[7])
			double_advances = ((single_advances & row_masks[2]) << 8) & empty & defense_map
			single_advances = single_advances & defense_map
			promotion_advances = ((brd.pieces[c][PAWN] << 8) & row_masks[7]) & empty & defense_map

			left_temp = ((brd.pieces[WHITE][PAWN] & (^column_masks[0])) << 7) & enemy & defense_map
			left_attacks = left_temp & (^row_masks[7])
			promotion_captures_left = left_temp & (row_masks[7])

			right_temp = ((brd.pieces[c][PAWN] & (^column_masks[7])) << 9) & enemy & defense_map
			right_attacks = right_temp & (^row_masks[7])
			promotion_captures_right = right_temp & (row_masks[7])
		} else { // black to move
			single_advances = (brd.pieces[BLACK][PAWN] >> 8) & empty & (^row_masks[0])
			double_advances = ((single_advances & row_masks[5]) >> 8) & empty & defense_map
			single_advances = single_advances & defense_map
			promotion_advances = ((brd.pieces[BLACK][PAWN] >> 8) & row_masks[0]) & empty & defense_map

			left_temp = ((brd.pieces[BLACK][PAWN] & (^column_masks[0])) >> 9) & enemy & defense_map
			left_attacks = left_temp & (^row_masks[0])
			promotion_captures_left = left_temp & (row_masks[0])

			right_temp = ((brd.pieces[BLACK][PAWN] & (^column_masks[7])) >> 7) & enemy & defense_map
			right_attacks = right_temp & (^row_masks[0])
			promotion_captures_right = right_temp & (row_masks[0])
		}

		// promotion captures
		for ; promotion_captures_left > 0; promotion_captures_left.Clear(to) {
			to = furthest_forward(c, promotion_captures_left)
			from = to + pawn_from_offsets[c][2]
			if is_pinned(brd, from, c, e) == 0 {
				// build_promotion_captures(piece_id, from, to, color, cls_promotion_capture, sq_board, promotions)

			}
		}
		for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
			to = furthest_forward(c, promotion_captures_right)
			from = to + pawn_from_offsets[c][3]
			if is_pinned(brd, from, c, e) == 0 {

				// build_promotion_captures(piece_id, from, to, color, cls_promotion_capture, sq_board, promotions);
			}
		}

		// promotion advances
		for ; promotion_advances > 0; promotion_advances.Clear(to) {
			to = furthest_forward(c, promotion_advances)
			from = to + pawn_from_offsets[c][0]
			if is_pinned(brd, from, c, e) == 0 {

				// build_promotions(piece_id, from, to, color, cls_promotion, promotions)
			}
		}

		// regular pawn attacks
		for ; left_attacks > 0; left_attacks.Clear(to) {
			to = furthest_forward(c, left_attacks)
			from = to + pawn_from_offsets[c][2]
			if is_pinned(brd, from, c, e) == 0 {

				// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures)
			}
		}

		for ; right_attacks > 0; right_attacks.Clear(to) {
			to = furthest_forward(c, right_attacks)
			from = to + pawn_from_offsets[c][3]
			if is_pinned(brd, from, c, e) == 0 {

				// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures)
			}
		}

		// en-passant captures
		if brd.enp_target != SQ_INVALID {
			enp_target := brd.enp_target
			for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
				from = furthest_forward(c, f)
				if c > 0 {
					to = int(enp_target) + 8
				} else {
					to = int(enp_target) - 8
				}
				if is_pinned(brd, from, c, e) == 0 {

					// build_enp_capture(piece_id, from, to, cls_enp_capture, enp_target, sq_board, captures)
				}
			}
		}

		// double advances
		for ; double_advances > 0; double_advances.Clear(to) {
			to = furthest_forward(c, double_advances)
			from = to + pawn_from_offsets[c][1]
			if is_pinned(brd, from, c, e) == 0 {

				// build_move(piece_id, from, to, cls_enp_advance, moves)
			}
		}
		// single advances
		for ; single_advances > 0; single_advances.Clear(to) {
			to = furthest_forward(c, single_advances)
			from = to + pawn_from_offsets[c][0]
			if is_pinned(brd, from, c, e) == 0 {

				// build_move(piece_id, from, to, cls_pawn_move, moves)
			}
		}

		// Knights
		for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f) // Locate each knight for the side to move.
			if is_pinned(brd, from, c, e) == 0 {
				for t := (knight_masks[from] & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {

						// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures);

					} else {

						// build_move(piece_id, from, to, cls_regular_move, moves);

					}
				}
			}
		}

		// Bishops
		for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			if is_pinned(brd, from, c, e) == 0 {
				for t := (bishop_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {

						// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures);
					} else {

						// build_move(piece_id, from, to, cls_regular_move, moves);
					}
				}
			}
		}

		// Rooks
		for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			if is_pinned(brd, from, c, e) == 0 {
				for t := (rook_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {

						// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures);
					} else {

						// build_move(piece_id, from, to, cls_regular_move, moves);
					}
				}
			}
		}

		// Queens
		for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			if is_pinned(brd, from, c, e) == 0 {
				for t := (queen_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {

						// build_capture(piece_id, from, to, cls_regular_capture, sq_board, captures);
					} else {

						// build_move(piece_id, from, to, cls_regular_move, moves);
					}
				}
			}
		}

	}

	// If there's more than one attacking piece, the only way out is to move the king.
	// King captures
	for t := (king_masks[king_sq] & enemy); t > 0; t.Clear(to) { // generate to squares
		to = furthest_forward(c, t)
		if !is_attacked_by(brd, to, e, c) && threat_dir_1 != directions[king_sq][to] &&
			threat_dir_2 != directions[king_sq][to] {

			// build_capture(piece_id, king_sq, to, cls_regular_capture, sq_board, captures)
		}
	}
	// King moves
	for t := (king_masks[king_sq] & empty); t > 0; t.Clear(to) { // generate to squares
		to = furthest_forward(c, t)
		if !is_attacked_by(brd, to, e, c) && threat_dir_1 != directions[king_sq][to] &&
			threat_dir_2 != directions[king_sq][to] {

			// build_move(piece_id, king_sq, to, cls_regular_move, moves)
		}
	}

}

// Need to store

func encode_move(from, to int, piece, captured, promoted Piece) Move {
	return Move(from) |
		(Move(to) << 6) |
		(Move(piece) << 12) |
		(Move(captured) << 15) |
		(Move(promoted) << 18)
}

func scan_down(occ BB, dir, sq int) BB {
	ray := ray_masks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (ray_masks[dir][msb(blockers)])
	}
	return ray
}

func scan_up(occ BB, dir, sq int) BB {
	ray := ray_masks[dir][sq]
	blockers := (ray & occ)
	if blockers > 0 {
		ray ^= (ray_masks[dir][lsb(blockers)])
	}
	return ray
}

func rook_attacks(occ BB, sq int) BB {
	return scan_up(occ, NORTH, sq) | scan_up(occ, EAST, sq) | scan_down(occ, SOUTH, sq) | scan_down(occ, WEST, sq)
}

func bishop_attacks(occ BB, sq int) BB {
	return scan_up(occ, NW, sq) | scan_up(occ, NE, sq) | scan_down(occ, SE, sq) | scan_down(occ, SW, sq)
}

func queen_attacks(occ BB, sq int) BB {
	return (bishop_attacks(occ, sq) | rook_attacks(occ, sq))
}
