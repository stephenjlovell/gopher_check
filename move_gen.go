//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

func get_non_captures(brd *Board, htable *HistoryTable, remaining_moves *MoveList) {
	var from, to int
	var single_advances, double_advances BB
	c := brd.c
	occ := brd.AllOccupied()
	empty := ^occ
	var m Move

	// Castles
	castle := brd.castle
	if castle > uint8(0) { // get_non_captures is only called when not in check.
		e := brd.Enemy()
		if c == WHITE {
			if (castle&C_WQ > uint8(0)) && castle_queenside_intervening[WHITE]&occ == 0 &&
				!is_attacked_by(brd, occ, C1, e, c) && !is_attacked_by(brd, occ, D1, e, c) {
				m = NewRegularMove(E1, C1, KING)
				remaining_moves.Push(SortItem{htable.Probe(KING, c, C1) | 1, m})
			}
			if (castle&C_WK > uint8(0)) && castle_kingside_intervening[WHITE]&occ == 0 &&
				!is_attacked_by(brd, occ, F1, e, c) && !is_attacked_by(brd, occ, G1, e, c) {
				m = NewRegularMove(E1, G1, KING)
				remaining_moves.Push(SortItem{htable.Probe(KING, c, G1) | 1, m})
			}
		} else {
			if (castle&C_BQ > uint8(0)) && castle_queenside_intervening[BLACK]&occ == 0 &&
				!is_attacked_by(brd, occ, C8, e, c) && !is_attacked_by(brd, occ, D8, e, c) {
				m = NewRegularMove(E8, C8, KING)
				remaining_moves.Push(SortItem{htable.Probe(KING, c, C8) | 1, m})
			}
			if (castle&C_BK > uint8(0)) && castle_kingside_intervening[BLACK]&occ == 0 &&
				!is_attacked_by(brd, occ, F8, e, c) && !is_attacked_by(brd, occ, G8, e, c) {
				m = NewRegularMove(E8, G8, KING)
				remaining_moves.Push(SortItem{htable.Probe(KING, c, G8) | 1, m})
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
		from = to + pawn_from_offsets[c][OFF_DOUBLE]
		m = NewRegularMove(from, to, PAWN)
		remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	for ; single_advances > 0; single_advances.Clear(to) {
		to = furthest_forward(c, single_advances)
		from = to + pawn_from_offsets[c][OFF_SINGLE]
		m = NewRegularMove(from, to, PAWN)
		remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)                               // Locate each knight for the side to move.
		for t := (knight_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, KNIGHT)
			remaining_moves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, BISHOP)
			remaining_moves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, ROOK)
			remaining_moves.Push(SortItem{htable.Probe(ROOK, c, to), m})
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, QUEEN)
			remaining_moves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
		}
	}
	// Kings
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (king_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, KING)
			remaining_moves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}
}

// Pawn promotions are also generated during get_captures routine.
func get_captures(brd *Board, htable *HistoryTable, winning, losing *MoveList) {
	var from, to int
	var m Move

	c, e := brd.c, brd.Enemy()
	occ := brd.AllOccupied()
	enemy := brd.Placement(e)

	// Pawns
	var left_temp, right_temp BB
	var left_attacks, right_attacks BB
	var promotion_captures_left, promotion_captures_right, promotion_advances BB

	if c == WHITE { // white to move
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
		from = to + pawn_from_offsets[c][OFF_LEFT]
		get_promotion_captures(brd, winning, from, to, brd.squares[to])
	}

	for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
		to = furthest_forward(c, promotion_captures_right)
		from = to + pawn_from_offsets[c][OFF_RIGHT]
		get_promotion_captures(brd, winning, from, to, brd.squares[to])
	}

	// promotion advances
	for ; promotion_advances > 0; promotion_advances.Clear(to) {
		to = furthest_forward(c, promotion_advances)
		from = to + pawn_from_offsets[c][OFF_SINGLE]
		get_promotion_advances(brd, winning, losing, from, to)
	}

	// regular pawn attacks
	for ; left_attacks > 0; left_attacks.Clear(to) {
		to = furthest_forward(c, left_attacks)
		from = to + pawn_from_offsets[c][OFF_LEFT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
	}
	for ; right_attacks > 0; right_attacks.Clear(to) {
		to = furthest_forward(c, right_attacks)
		from = to + pawn_from_offsets[c][OFF_RIGHT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
	}
	// en-passant captures
	if brd.enp_target != SQ_INVALID {
		enp_target := brd.enp_target
		for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			if c == WHITE {
				to = int(enp_target) + 8
			} else {
				to = int(enp_target) - 8
			}
			m = NewCapture(from, to, PAWN, PAWN)
			winning.Push(SortItem{mvv_lva(PAWN, PAWN), m})
		}
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (knight_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KNIGHT, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], KNIGHT), m})
			} else {
				losing.Push(SortItem{mvv_lva(brd.squares[to], KNIGHT), m})
			}
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, BISHOP, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], BISHOP), m})
			} else {
				losing.Push(SortItem{mvv_lva(brd.squares[to], BISHOP), m})
			}
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, ROOK, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], ROOK), m})
			} else {
				losing.Push(SortItem{mvv_lva(brd.squares[to], ROOK), m})
			}
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, QUEEN, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], QUEEN), m})
			} else {
				losing.Push(SortItem{mvv_lva(brd.squares[to], QUEEN), m})
			}
		}
	}
	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (king_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KING, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 { // Cannot move into check
				winning.Push(SortItem{mvv_lva(brd.squares[to], KING), m})
			}
		}
	}
}

func get_winning_captures(brd *Board, htable *HistoryTable, winning *MoveList) {
	var from, to int
	var m Move

	c, e := brd.c, brd.Enemy()
	occ := brd.AllOccupied()
	enemy := brd.Placement(e)

	// Pawns
	var left_temp, right_temp, left_attacks, right_attacks BB
	var promotion_captures_left, promotion_captures_right BB
	var promotion_advances BB

	if c == WHITE { // white to move
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
		from = to + pawn_from_offsets[c][OFF_LEFT]
		get_promotion_captures(brd, winning, from, to, brd.squares[to])
	}

	for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
		to = furthest_forward(c, promotion_captures_right)
		from = to + pawn_from_offsets[c][OFF_RIGHT]
		get_promotion_captures(brd, winning, from, to, brd.squares[to])
	}

	// promotion advances
	for ; promotion_advances > 0; promotion_advances.Clear(to) {
		to = furthest_forward(c, promotion_advances)
		from = to + pawn_from_offsets[c][OFF_SINGLE]
		get_promotion_advances(brd, winning, winning, from, to)
	}

	// regular pawn attacks
	for ; left_attacks > 0; left_attacks.Clear(to) {
		to = furthest_forward(c, left_attacks)
		from = to + pawn_from_offsets[c][OFF_LEFT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
	}
	for ; right_attacks > 0; right_attacks.Clear(to) {
		to = furthest_forward(c, right_attacks)
		from = to + pawn_from_offsets[c][OFF_RIGHT]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
	}
	// en-passant captures
	if brd.enp_target != SQ_INVALID {
		enp_target := brd.enp_target
		for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			if c == WHITE {
				to = int(enp_target) + 8
			} else {
				to = int(enp_target) - 8
			}
			m = NewCapture(from, to, PAWN, PAWN)
			winning.Push(SortItem{mvv_lva(PAWN, PAWN), m})
		}
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (knight_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KNIGHT, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], KNIGHT), m})
			}
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, BISHOP, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], BISHOP), m})
			}
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, ROOK, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], ROOK), m})
			}
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, QUEEN, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], QUEEN), m})
			}
		}
	}
	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		for t := (king_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KING, brd.squares[to])
			if get_see(brd, from, to, brd.squares[to]) >= 0 {
				winning.Push(SortItem{mvv_lva(brd.squares[to], KING), m})
			}
		}
	}
}

func get_evasions(brd *Board, htable *HistoryTable, winning, losing, remaining_moves *MoveList) {
	c, e := brd.c, brd.Enemy()

	var defense_map BB
	var from, to, threat_sq_1, threat_sq_2 int

	threat_dir_1, threat_dir_2 := DIR_INVALID, DIR_INVALID
	occ := brd.AllOccupied()
	empty := ^occ
	enemy := brd.Placement(e)

	king_sq := brd.KingSq(c)
	threats := color_attack_map(brd, occ, king_sq, e, c) // find any enemy pieces that attack the king.
	threat_count := pop_count(threats)

	// Get direction of the attacker(s) and any intervening squares between the attacker and the king.
	if threat_count == 1 {
		threat_sq_1 = lsb(threats)
		if brd.TypeAt(threat_sq_1) != PAWN {
			threat_dir_1 = directions[threat_sq_1][king_sq]
		}
		defense_map |= (intervening[threat_sq_1][king_sq] | threats)
	} else {
		threat_sq_1 = lsb(threats)
		if brd.TypeAt(threat_sq_1) != PAWN {
			threat_dir_1 = directions[threat_sq_1][king_sq]
		}
		threat_sq_2 = msb(threats)
		if brd.TypeAt(threat_sq_2) != PAWN {
			threat_dir_2 = directions[threat_sq_2][king_sq]
		}
	}

	var m Move
	if threat_count == 1 { // Attempt to capture or block the attack with any piece if there's only one attacker.
		// Pawns
		var single_advances, double_advances, left_temp, right_temp, left_attacks, right_attacks BB
		var promotion_captures_left, promotion_captures_right BB
		var promotion_advances BB

		if c > 0 { // white to move
			single_advances = (brd.pieces[WHITE][PAWN] << 8) & empty & (^row_masks[7])
			double_advances = ((single_advances & row_masks[2]) << 8) & empty & defense_map
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
			promotion_advances = ((brd.pieces[BLACK][PAWN] >> 8) & row_masks[0]) & empty & defense_map

			left_temp = ((brd.pieces[BLACK][PAWN] & (^column_masks[0])) >> 9) & enemy & defense_map
			left_attacks = left_temp & (^row_masks[0])
			promotion_captures_left = left_temp & (row_masks[0])

			right_temp = ((brd.pieces[BLACK][PAWN] & (^column_masks[7])) >> 7) & enemy & defense_map
			right_attacks = right_temp & (^row_masks[0])
			promotion_captures_right = right_temp & (row_masks[0])
		}
		single_advances &= defense_map

		// promotion captures
		for ; promotion_captures_left > 0; promotion_captures_left.Clear(to) {
			to = furthest_forward(c, promotion_captures_left)
			from = to + pawn_from_offsets[c][OFF_LEFT]
			if pinned_can_move(brd, from, to, c, e) {
				get_promotion_captures(brd, winning, from, to, brd.squares[to])
			}
		}
		for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
			to = furthest_forward(c, promotion_captures_right)
			from = to + pawn_from_offsets[c][OFF_RIGHT]
			if pinned_can_move(brd, from, to, c, e) {
				get_promotion_captures(brd, winning, from, to, brd.squares[to])
			}
		}
		// promotion advances
		for ; promotion_advances > 0; promotion_advances.Clear(to) {
			to = furthest_forward(c, promotion_advances)
			from = to + pawn_from_offsets[c][OFF_SINGLE]
			if pinned_can_move(brd, from, to, c, e) {
				get_promotion_advances(brd, winning, remaining_moves, from, to)
			}
		}
		// regular pawn attacks
		for ; left_attacks > 0; left_attacks.Clear(to) {
			to = furthest_forward(c, left_attacks)
			from = to + pawn_from_offsets[c][OFF_LEFT]
			if pinned_can_move(brd, from, to, c, e) {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
			}
		}
		for ; right_attacks > 0; right_attacks.Clear(to) {
			to = furthest_forward(c, right_attacks)
			from = to + pawn_from_offsets[c][OFF_RIGHT]
			if pinned_can_move(brd, from, to, c, e) {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				winning.Push(SortItem{mvv_lva(brd.squares[to], PAWN), m})
			}
		}
		// en-passant captures
		if brd.enp_target != SQ_INVALID {
			enp_target := brd.enp_target
			for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
				from = furthest_forward(c, f)
				if c == WHITE {
					to = int(enp_target) + 8
				} else {
					to = int(enp_target) - 8
				}
				// In addition to making sure this capture will get the king out of check and that
				// the piece is not pinned, verify that removing the enemy pawn does not leave the
				// king in check.
				if (sq_mask_on[to]&defense_map) > 0 && pinned_can_move(brd, from, to, c, e) &&
					is_pinned(brd, int(enp_target), c, e)&sq_mask_on[to] > 0 {

					m = NewCapture(from, to, PAWN, PAWN)
					winning.Push(SortItem{mvv_lva(PAWN, PAWN), m})
				}
			}
		}
		// double advances
		for ; double_advances > 0; double_advances.Clear(to) {
			to = furthest_forward(c, double_advances)
			from = to + pawn_from_offsets[c][OFF_DOUBLE]
			if pinned_can_move(brd, from, to, c, e) {
				m = NewRegularMove(from, to, PAWN)
				remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
			}
		}
		// single advances
		for ; single_advances > 0; single_advances.Clear(to) {
			to = furthest_forward(c, single_advances)
			from = to + pawn_from_offsets[c][OFF_SINGLE]
			if pinned_can_move(brd, from, to, c, e) {
				m = NewRegularMove(from, to, PAWN)
				remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
			}
		}
		// Knights
		for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f) // Locate each knight for the side to move.
			// Knights cannot move if pinned by a sliding piece, since they can't move along the ray between
			// the threat piece and their own king.
			if is_pinned(brd, from, c, e) == BB(ANY_SQUARE_MASK) {
				for t := (knight_masks[from] & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {
						m = NewCapture(from, to, KNIGHT, brd.squares[to])
						if get_see(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvv_lva(brd.squares[to], KNIGHT), m})
						} else {
							losing.Push(SortItem{mvv_lva(brd.squares[to], KNIGHT), m})
						}
					} else {
						m = NewRegularMove(from, to, KNIGHT)
						remaining_moves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
					}
				}
			}
		}
		// Bishops
		for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			for t := (bishop_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
				to = furthest_forward(c, t)
				if pinned_can_move(brd, from, to, c, e) {
					if sq_mask_on[to]&enemy > 0 {
						m = NewCapture(from, to, BISHOP, brd.squares[to])
						if get_see(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvv_lva(brd.squares[to], BISHOP), m})
						} else {
							losing.Push(SortItem{mvv_lva(brd.squares[to], BISHOP), m})
						}
					} else {
						m = NewRegularMove(from, to, BISHOP)
						remaining_moves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
					}
				}
			}
		}
		// Rooks
		for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			for t := (rook_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
				to = furthest_forward(c, t)
				if pinned_can_move(brd, from, to, c, e) {
					if sq_mask_on[to]&enemy > 0 {
						m = NewCapture(from, to, ROOK, brd.squares[to])
						if get_see(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvv_lva(brd.squares[to], ROOK), m})
						} else {
							losing.Push(SortItem{mvv_lva(brd.squares[to], ROOK), m})
						}
					} else {
						m = NewRegularMove(from, to, ROOK)
						remaining_moves.Push(SortItem{htable.Probe(ROOK, c, to), m})
					}
				}
			}
		}
		// Queens
		for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f)
			for t := (queen_attacks(occ, from) & defense_map); t > 0; t.Clear(to) { // generate to squares
				to = furthest_forward(c, t)
				if pinned_can_move(brd, from, to, c, e) {
					if sq_mask_on[to]&enemy > 0 {
						m = NewCapture(from, to, QUEEN, brd.squares[to])
						if get_see(brd, from, to, brd.squares[to]) >= 0 {
							winning.Push(SortItem{mvv_lva(brd.squares[to], QUEEN), m})
						} else {
							losing.Push(SortItem{mvv_lva(brd.squares[to], QUEEN), m})
						}
					} else {
						m = NewRegularMove(from, to, QUEEN)
						remaining_moves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
					}
				}
			}
		}
	}
	// If there's more than one attacking piece, the only way out is to move the king.
	// King captures
	for t := (king_masks[king_sq] & enemy); t > 0; t.Clear(to) { // generate to squares
		to = furthest_forward(c, t)
		if !is_attacked_by(brd, occ, to, e, c) && threat_dir_1 != directions[king_sq][to] &&
			threat_dir_2 != directions[king_sq][to] {
			m = NewCapture(king_sq, to, KING, brd.squares[to])
			winning.Push(SortItem{mvv_lva(brd.squares[to], KING), m})
		}
	}
	// King moves
	for t := (king_masks[king_sq] & empty); t > 0; t.Clear(to) { // generate to squares
		to = furthest_forward(c, t)
		if !is_attacked_by(brd, occ, to, e, c) && threat_dir_1 != directions[king_sq][to] &&
			threat_dir_2 != directions[king_sq][to] {
			m = NewRegularMove(king_sq, to, KING)
			remaining_moves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}
}

func get_checks(brd *Board, htable *HistoryTable, remaining_moves *MoveList) {
	c, e := brd.c, brd.Enemy()
	king_sq := brd.KingSq(e)
	var f, t, single_advances, target, queen_target BB
	var from, to int
	var m Move
	occ := brd.AllOccupied()
	empty := ^occ
	// Pawn direct checks
	if c > 0 { // white to move
		single_advances = (brd.pieces[WHITE][PAWN] << 8) & empty
	} else { // black to move
		single_advances = (brd.pieces[BLACK][PAWN] >> 8) & empty
	}
	target = pawn_attack_masks[e][king_sq]
	for t = single_advances & target; t > 0; t.Clear(to) {
		to = furthest_forward(c, t)
		from = to + pawn_from_offsets[c][OFF_SINGLE]
		if get_see(brd, from, to, EMPTY) >= 0 { // make sure the checking piece won't be immediately recaptured
			m = NewRegularMove(from, to, PAWN)
			remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
		}
	}
	// Knight direct checks
	target = knight_masks[king_sq] & empty
	for f = brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f) // Locate each knight for the side to move.
		for t = (knight_masks[from] & target); t > 0; t.Clear(to) {
			to = furthest_forward(c, t)
			if get_see(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, KNIGHT)
				remaining_moves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
			}
		}
	}
	// Bishop direct checks
	target = bishop_attacks(occ, king_sq) & empty
	queen_target = target
	for f = brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t = (bishop_attacks(occ, from) & target); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			if get_see(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, BISHOP)
				remaining_moves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
			}
		}
	}
	// Rook direct checks
	target = rook_attacks(occ, king_sq) & empty
	queen_target |= target
	for f = brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t = (rook_attacks(occ, from) & target); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			if get_see(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, ROOK)
				remaining_moves.Push(SortItem{htable.Probe(ROOK, c, to), m})
			}
		}
	}
	// Queen direct checks
	for f = brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t = (queen_attacks(occ, from) & queen_target); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			if get_see(brd, from, to, EMPTY) >= 0 {
				m = NewRegularMove(from, to, QUEEN)
				remaining_moves.Push(SortItem{htable.Probe(QUEEN, c, to), m})
			}
		}
	}

	// indirect (discovered) checks
	var rook_blockers, bishop_blockers BB

	rook_blockers = rook_attacks(occ, king_sq) & (brd.pieces[c][BISHOP] |
		brd.pieces[c][KNIGHT] | brd.pieces[c][PAWN])
	bishop_blockers = bishop_attacks(occ, king_sq) & (brd.pieces[c][ROOK] |
		brd.pieces[c][KNIGHT] | brd.pieces[c][PAWN])
	if rook_blockers > 0 {
		rook_attackers := rook_attacks(occ^rook_blockers, king_sq) & (brd.pieces[c][ROOK] | brd.pieces[c][QUEEN])
		for dir := NORTH; dir <= WEST; dir++ {
			if ray_masks[dir][king_sq]&rook_attackers == 0 {
				rook_blockers &= (^ray_masks[dir][king_sq])
			}
		}
	}
	if bishop_blockers > 0 {
		bishop_attackers := bishop_attacks(occ^bishop_blockers, king_sq) & (brd.pieces[c][BISHOP] | brd.pieces[c][QUEEN])
		for dir := NW; dir <= SW; dir++ {
			if ray_masks[dir][king_sq]&bishop_attackers == 0 {
				bishop_blockers &= (^ray_masks[dir][king_sq])
			}
		}
	}

	var unblock_path BB // blockers must move off the path of attack.

	// don't bother with double advances.
	for t = single_advances & (bishop_blockers | rook_blockers); t > 0; t.Clear(to) {
		to = furthest_forward(c, t)
		from = to + pawn_from_offsets[c][OFF_SINGLE]
		m = NewRegularMove(from, to, PAWN)
		remaining_moves.Push(SortItem{htable.Probe(PAWN, c, to), m})
	}
	// Knights
	for f = brd.pieces[c][KNIGHT] & (bishop_blockers | rook_blockers); f > 0; f.Clear(from) {
		from = furthest_forward(c, f) // Locate each knight for the side to move.
		for t = (knight_masks[from] & empty); t > 0; t.Clear(to) {
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, KNIGHT)
			remaining_moves.Push(SortItem{htable.Probe(KNIGHT, c, to), m})
		}
	}
	// Bishops
	for f = brd.pieces[c][BISHOP] & rook_blockers; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		unblock_path = (^intervening[king_sq][from]) & empty
		for t = (bishop_attacks(occ, from) & unblock_path); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, BISHOP)
			remaining_moves.Push(SortItem{htable.Probe(BISHOP, c, to), m})
		}
	}
	// Rooks
	for f = brd.pieces[c][ROOK] & bishop_blockers; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		unblock_path = (^intervening[king_sq][from]) & empty
		for t = (rook_attacks(occ, from) & unblock_path); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, ROOK)
			remaining_moves.Push(SortItem{htable.Probe(ROOK, c, to), m})
		}
	}
	// Queens cannot give discovered check, since the enemy king would already be in check.

	// Kings
	for f := brd.pieces[c][KING] & (bishop_blockers | rook_blockers); f > 0; f.Clear(from) {
		from = brd.KingSq(c)
		unblock_path = (^intervening[king_sq][from]) & empty
		for t := (king_masks[from] & unblock_path); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewRegularMove(from, to, KING)
			remaining_moves.Push(SortItem{htable.Probe(KING, c, to), m})
		}
	}

}

// // uncomment for movegen testing.

// func get_promotion_advances(brd *Board, winning, losing *MoveList, from, to int) {
// 	var m Move
// 	var sort uint64
// 	for pc := Piece(QUEEN); pc >= KNIGHT; pc-- {
// 		m = NewMove(from, to, PAWN, EMPTY, pc)
// 		sort = sort_promotion_advances(brd, from, to, pc)
// 		if sort >= SORT_WINNING_PROMOTION {
// 			winning.Push(SortItem{sort})
// 		} else {
// 			losing.Push(SortItem{sort})
// 		}
// 	}
// }
//
// func get_promotion_captures(brd *Board, winning *MoveList, from, to int, captured_piece Piece) {
// 	var m Move
// 	for pc := Piece(QUEEN); pc >= KNIGHT; pc-- {
// 		m = NewMove(from, to, PAWN, captured_piece, pc)
// 		winning.Push(SortItem{sort_promotion_captures(brd, from, to, captured_piece, pc)})
// 	}
// }

func get_promotion_advances(brd *Board, winning, losing *MoveList, from, to int) {
	var m Move
	var sort uint64
	m = NewMove(from, to, PAWN, EMPTY, QUEEN)
	sort = sort_promotion_advances(brd, from, to, QUEEN)
	if sort >= SORT_WINNING_PROMOTION {
		winning.Push(SortItem{sort, m})
	} else {
		losing.Push(SortItem{sort, m})
	}
	m = NewMove(from, to, PAWN, EMPTY, KNIGHT)
	sort = sort_promotion_advances(brd, from, to, KNIGHT)
	if sort >= SORT_WINNING_PROMOTION {
		winning.Push(SortItem{sort, m})
	} else {
		losing.Push(SortItem{sort, m})
	}
}

func get_promotion_captures(brd *Board, winning *MoveList, from, to int, captured_piece Piece) {
	winning.Push(SortItem{sort_promotion_captures(brd, from, to, captured_piece, QUEEN),
		NewMove(from, to, PAWN, captured_piece, QUEEN)})
	winning.Push(SortItem{sort_promotion_captures(brd, from, to, captured_piece, KNIGHT),
		NewMove(from, to, PAWN, captured_piece, KNIGHT)})
}
