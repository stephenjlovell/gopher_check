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

// // determine which moves should be searchd sequentially, and which will be searched in parallel.
// func split_moves(brd *Board, in_check bool, node_type int) ([]Move, int) {

// 	// apply some heuristic to slice off n best moves based on node type.
// 	var best_moves, other_moves []MoveList

// 	if in_check {
// 		get_evasions(brd, best_moves, other_moves)
// 	} else {
// 		get_captures(brd, best_moves, other_moves)
// 		get_castles(brd, other_moves)
// 		get_non_captures(brd, best_moves, other_moves)
// 	}

// 	// switch node_type {
// 	// case Y_PV, Y_ALL:
// 	// 	// only the best available legal move is searched iteratively.  The rest are searched in parallel.
// 	// case Y_CUT:
// 	// 	// hash move, promotions, winning captures, and killers are searched iteratively.
// 	// 	// all remaining moves are searched concurrently.
// 	// }
// 	return best_moves, other_moves
// }

// Ordering: PV/hash, promotions, winning captures, killers, losing captures, quiet moves
// const ( // ordering type
//   S_HASH = iota
//   S_PROMOTION
//   S_WINNING
//   S_KILLER  // /\ 'promising' moves
//   S_CASTLE  // \/ other moves
//   S_LOSING
//   S_QUIET
// )

// // ...Search
// func example(brd *Board) {
// 	next := make(chan bool)
// 	best_moves := make(chan Move) // sequential (unbuffered) channels
// 	other_moes := make(chan Move)

// 	next <- true
// 	go get_next_move(brd, moves, next)
// 	receive_best_moves:
// 	for {
// 		select {
// 		case move, ok := <-best_moves:
// 			if ok {
// 				// search iteratively
// 			} else {
// 				break receive_best_moves  // no more moves to receive.
// 			}
// 		default:
// 			next <- true
// 		}
// 	}
// 	receive_moves:
// 	for {
// 		select {
// 		case move, ok := <-other_moves:
// 			if ok {
// 				// search concurrently
// 			} else {
// 				break receive_moves  // no more moves to receive.
// 			}
// 		default:
// 			next <- true
// 		}
// 	}

// }

// // Moves are generated in batches.  move_gen keeps track of move generation batch phase.
// // Search sends a message to NextMove requesting a move.  If an untried move is available from the current
// // batch, NextMove sends that move to be searched.  If no moves are left in the current batch, NextMove
// // generates and sorts the next batch of moves, and sends the first one.  If every batch has been generated
// // and sent, NextMove returns nil/flag.

// func get_next_move(brd *Board, moves chan Move, next chan bool) {
// 	batch := 0
// 	var best_moves, other_moves []MoveList

// 	for {
// 		select {
// 		case next_batch, ok := <-next:  // if next is closed
// 			if ok {
// 				switch batch {
// 				case S_HASH:
// 					get_hash_move()
// 			  case S_PROMOTION:

// 			  case S_WINNING:

// 			  case S_KILLER:

// 			  case S_CASTLE:

// 			  case S_LOSING:

// 			  case S_QUIET:

// 			  default:
// 			  	close(moves)
// 					return
// 				}
// 				batch++
// 			} else {
// 		  	close(moves)
// 				return
// 			}
// 		}
// 	}
// }

import (
	"container/heap"
)

func get_all_moves(brd *Board, in_check bool, hash_move Move) (*MoveList, *MoveList) {
	var best_moves, remaining_moves MoveList
	if in_check {
		get_evasions(brd, &best_moves, &remaining_moves, hash_move)
	} else {
		get_captures(brd, &best_moves, &remaining_moves, hash_move)
		get_non_captures(brd, hash_move, &remaining_moves)
	}
	return &best_moves, &remaining_moves
}

func get_best_moves(brd *Board, in_check bool, hash_move Move) (*MoveList, *MoveList) {
	var best_moves, remaining_moves MoveList
	if in_check {
		get_evasions(brd, &best_moves, &remaining_moves, hash_move)
	} else {
		get_captures(brd, &best_moves, &remaining_moves, hash_move)
	}
	return &best_moves, &remaining_moves
}

func get_remaining_moves(brd *Board, in_check bool, hash_move Move, remaining_moves *MoveList) {
	// when in check, all moves are generated by the get_evasions() routine (including captures and non-captures).
	if !in_check {
		get_non_captures(brd, hash_move, remaining_moves)
	}
}

// Pawn promotions are also generated during get_captures routine.

func get_captures(brd *Board, best_moves, remaining_moves *MoveList, hash_move Move) {
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

	var m Move
	// promotion captures
	for ; promotion_captures_left > 0; promotion_captures_left.Clear(to) {
		to = furthest_forward(c, promotion_captures_left)
		from = to + pawn_from_offsets[c][2]
		m = NewPromotionCapture(from, to, brd.squares[to], QUEEN)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF + 1})
		}
		m = NewPromotionCapture(from, to, brd.squares[to], KNIGHT)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF + 1})
		}
	}

	for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
		to = furthest_forward(c, promotion_captures_right)
		from = to + pawn_from_offsets[c][2]
		m = NewPromotionCapture(from, to, brd.squares[to], QUEEN)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF + 1})
		}
		m = NewPromotionCapture(from, to, brd.squares[to], KNIGHT)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF + 1})
		}
	}

	// promotion advances
	for ; promotion_advances > 0; promotion_advances.Clear(to) {
		to = furthest_forward(c, promotion_advances)
		from = to + pawn_from_offsets[c][0]
		m = NewPromotion(from, to, QUEEN)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF})
		}
		best_moves.Push(&SortItem{m, INF})
		m = NewPromotion(from, to, KNIGHT)
		if m != hash_move {
			heap.Push(best_moves, &SortItem{m, INF})
		}
	}
	var see int
	// regular pawn attacks
	for ; left_attacks > 0; left_attacks.Clear(to) {
		to = furthest_forward(c, left_attacks)
		from = to + pawn_from_offsets[c][2]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		see = get_see(brd, from, to, c)
		if see >= 0 {
			heap.Push(best_moves, &SortItem{m, see})
		} else {
			heap.Push(remaining_moves, &SortItem{m, INF - see})
		}
	}
	for ; right_attacks > 0; right_attacks.Clear(to) {
		to = furthest_forward(c, right_attacks)
		from = to + pawn_from_offsets[c][3]
		m = NewCapture(from, to, PAWN, brd.squares[to])
		see = get_see(brd, from, to, c)
		if see >= 0 {
			heap.Push(best_moves, &SortItem{m, see})
		} else {
			heap.Push(remaining_moves, &SortItem{m, INF - see})
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
			m = NewCapture(from, to, PAWN, PAWN)

			see = get_see(brd, from, to, c)

			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (knight_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KNIGHT, brd.squares[to])
			see = get_see(brd, from, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, BISHOP, brd.squares[to])
			see = get_see(brd, from, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, ROOK, brd.squares[to])
			see = get_see(brd, from, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, QUEEN, brd.squares[to])
			see = get_see(brd, from, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// King
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = furthest_forward(c, brd.pieces[c][KING])
		for t := (king_masks[from] & enemy); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewCapture(from, to, KING, brd.squares[to])
			see = get_see(brd, from, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
}

func get_non_captures(brd *Board, hash_move Move, remaining_moves *MoveList) {
	var from, to int
	var single_advances, double_advances BB
	c, e := brd.c, brd.Enemy()
	occ := brd.Occupied()
	empty := ^occ
	var m Move
	// Castles
	castle := brd.castle
	if castle > uint8(0) { // get_non_captures is only called when not in check.
		if c == WHITE {
			if (castle&C_WQ > uint8(0)) && !(castle_queenside_intervening[1]&occ > 0) &&
				!is_attacked_by(brd, D1, e, c) && !is_attacked_by(brd, C1, e, c) {
				m = NewMove(KING, E1, C1)
				heap.Push(remaining_moves, &SortItem{m, 10})
			}
			if (castle&C_WK > uint8(0)) && !(castle_kingside_intervening[1]&occ > 0) &&
				!is_attacked_by(brd, F1, e, c) && !is_attacked_by(brd, G1, e, c) {
				m = NewMove(KING, E1, G1)
				heap.Push(remaining_moves, &SortItem{m, 10})
			}
		} else {
			if (castle&C_BQ > uint8(0)) && !(castle_queenside_intervening[0]&occ > 0) &&
				!is_attacked_by(brd, D8, e, c) && !is_attacked_by(brd, C8, e, c) {
				m = NewMove(KING, E8, C8)
				heap.Push(remaining_moves, &SortItem{m, 10})
			}
			if (castle&C_BK > uint8(0)) && !(castle_kingside_intervening[0]&occ > 0) &&
				!is_attacked_by(brd, F8, e, c) && !is_attacked_by(brd, G8, e, c) {
				m = NewMove(KING, E8, G8)
				heap.Push(remaining_moves, &SortItem{m, 10})
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
		from = to + pawn_from_offsets[c][1]
		m = NewMove(from, to, PAWN)
		heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(PAWN, c, to)})
	}
	for ; single_advances > 0; single_advances.Clear(to) {
		to = furthest_forward(c, single_advances)
		from = to + pawn_from_offsets[c][0]
		m = NewMove(from, to, PAWN)
		heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(PAWN, c, to)})
	}
	// Knights
	for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)                               // Locate each knight for the side to move.
		for t := (knight_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewMove(from, to, KNIGHT)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(KNIGHT, c, to)})
		}
	}
	// Bishops
	for f := brd.pieces[c][BISHOP]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (bishop_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewMove(from, to, BISHOP)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(BISHOP, c, to)})
		}
	}
	// Rooks
	for f := brd.pieces[c][ROOK]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (rook_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewMove(from, to, ROOK)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(ROOK, c, to)})
		}
	}
	// Queens
	for f := brd.pieces[c][QUEEN]; f > 0; f.Clear(from) {
		from = furthest_forward(c, f)
		for t := (queen_attacks(occ, from) & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewMove(from, to, QUEEN)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(QUEEN, c, to)})
		}
	}
	// Kings
	for f := brd.pieces[c][KING]; f > 0; f.Clear(from) {
		from = furthest_forward(c, brd.pieces[c][KING])
		for t := (king_masks[from] & empty); t > 0; t.Clear(to) { // generate to squares
			to = furthest_forward(c, t)
			m = NewMove(from, to, KING)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(KING, c, to)})
		}
	}
}

func get_evasions(brd *Board, best_moves, remaining_moves *MoveList, hash_move Move) {
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

	var m Move
	var see int

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
				m = NewPromotionCapture(from, to, brd.squares[to], QUEEN)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF + 1})
				}
				m = NewPromotionCapture(from, to, brd.squares[to], KNIGHT)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF + 1})
				}
			}
		}
		for ; promotion_captures_right > 0; promotion_captures_right.Clear(to) {
			to = furthest_forward(c, promotion_captures_right)
			from = to + pawn_from_offsets[c][3]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewPromotionCapture(from, to, brd.squares[to], QUEEN)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF + 1})
				}
				m = NewPromotionCapture(from, to, brd.squares[to], KNIGHT)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF + 1})
				}
			}
		}
		// promotion advances
		for ; promotion_advances > 0; promotion_advances.Clear(to) {
			to = furthest_forward(c, promotion_advances)
			from = to + pawn_from_offsets[c][0]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewPromotion(from, to, QUEEN)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF})
				}
				best_moves.Push(&SortItem{m, INF})
				m = NewPromotion(from, to, KNIGHT)
				if m != hash_move {
					heap.Push(best_moves, &SortItem{m, INF})
				}
			}
		}
		// regular pawn attacks
		for ; left_attacks > 0; left_attacks.Clear(to) {
			to = furthest_forward(c, left_attacks)
			from = to + pawn_from_offsets[c][2]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				see = get_see(brd, from, to, c)
				if see >= 0 {
					heap.Push(best_moves, &SortItem{m, see})
				} else {
					heap.Push(remaining_moves, &SortItem{m, INF - see})
				}
			}
		}
		for ; right_attacks > 0; right_attacks.Clear(to) {
			to = furthest_forward(c, right_attacks)
			from = to + pawn_from_offsets[c][3]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewCapture(from, to, PAWN, brd.squares[to])
				see = get_see(brd, from, to, c)
				if see >= 0 {
					heap.Push(best_moves, &SortItem{m, see})
				} else {
					heap.Push(remaining_moves, &SortItem{m, INF - see})
				}
			}
		}
		// en-passant captures
		if brd.enp_target != SQ_INVALID {
			enp_target := brd.enp_target
			for f := (brd.pieces[c][PAWN] & pawn_side_masks[enp_target]); f > 0; f.Clear(from) {
				from = furthest_forward(c, f)
				if is_pinned(brd, from, c, e) == 0 {
					if c == WHITE {
						to = int(enp_target) + 8
					} else {
						to = int(enp_target) - 8
					}
					m = NewCapture(from, to, PAWN, PAWN)
					see = get_see(brd, from, to, c)
					if see >= 0 {
						heap.Push(best_moves, &SortItem{m, see})
					} else {
						heap.Push(remaining_moves, &SortItem{m, INF - see})
					}
				}
			}
		}
		// double advances
		for ; double_advances > 0; double_advances.Clear(to) {
			to = furthest_forward(c, double_advances)
			from = to + pawn_from_offsets[c][1]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewMove(from, to, PAWN)
				heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(PAWN, c, to)})
			}
		}
		// single advances
		for ; single_advances > 0; single_advances.Clear(to) {
			to = furthest_forward(c, single_advances)
			from = to + pawn_from_offsets[c][0]
			if is_pinned(brd, from, c, e) == 0 {
				m = NewMove(from, to, PAWN)
				heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(PAWN, c, to)})
			}
		}
		// Knights
		for f := brd.pieces[c][KNIGHT]; f > 0; f.Clear(from) {
			from = furthest_forward(c, f) // Locate each knight for the side to move.
			if is_pinned(brd, from, c, e) == 0 {
				for t := (knight_masks[from] & defense_map); t > 0; t.Clear(to) { // generate to squares
					to = furthest_forward(c, t)
					if sq_mask_on[to]&enemy > 0 {
						m = NewCapture(from, to, KNIGHT, brd.squares[to])
						see = get_see(brd, from, to, c)
						if see >= 0 {
							heap.Push(best_moves, &SortItem{m, see})
						} else {
							heap.Push(remaining_moves, &SortItem{m, INF - see})
						}
					} else {
						m = NewMove(from, to, KNIGHT)
						heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(KNIGHT, c, to)})
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
						m = NewCapture(from, to, BISHOP, brd.squares[to])
						see = get_see(brd, from, to, c)
						if see >= 0 {
							heap.Push(best_moves, &SortItem{m, see})
						} else {
							heap.Push(remaining_moves, &SortItem{m, INF - see})
						}
					} else {
						m = NewMove(from, to, BISHOP)
						heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(BISHOP, c, to)})
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
						m = NewCapture(from, to, ROOK, brd.squares[to])
						see = get_see(brd, from, to, c)
						if see >= 0 {
							heap.Push(best_moves, &SortItem{m, see})
						} else {
							heap.Push(remaining_moves, &SortItem{m, INF - see})
						}
					} else {
						m = NewMove(from, to, BISHOP)
						heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(ROOK, c, to)})
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
						m = NewCapture(from, to, QUEEN, brd.squares[to])
						see = get_see(brd, from, to, c)
						if see >= 0 {
							heap.Push(best_moves, &SortItem{m, see})
						} else {
							heap.Push(remaining_moves, &SortItem{m, INF - see})
						}
					} else {
						m = NewMove(from, to, BISHOP)
						heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(QUEEN, c, to)})
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
			m = NewCapture(king_sq, to, KING, brd.squares[to])
			see = get_see(brd, king_sq, to, c)
			if see >= 0 {
				heap.Push(best_moves, &SortItem{m, see})
			} else {
				heap.Push(remaining_moves, &SortItem{m, INF - see})
			}
		}
	}
	// King moves
	for t := (king_masks[king_sq] & empty); t > 0; t.Clear(to) { // generate to squares
		to = furthest_forward(c, t)
		if !is_attacked_by(brd, to, e, c) && threat_dir_1 != directions[king_sq][to] &&
			threat_dir_2 != directions[king_sq][to] {
			m = NewMove(king_sq, to, KING)
			heap.Push(remaining_moves, &SortItem{m, main_htable.Probe(KING, c, to)})
		}
	}
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
